package cmd

import (
	"fmt"
	"os"
	"prctl/internal/configutil"
	"prctl/internal/gitutil"
	"prctl/internal/systemcode"
	client "prctl/pkg/bitbucket"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

type declineCmdParams struct {
	Provider   string
	Repository string
}

func fillDefaultDeclineCmdParams(params *declineCmdParams) {
	defaultRepo, err := gitutil.GetRemoteInfo()
	if err == nil {
		params.Repository = fmt.Sprintf("%s/%s", defaultRepo.Owner, defaultRepo.Name)
		params.Provider = string(defaultRepo.Provider)
	}
}

func fillFlagDeclineCmdParams(cmd *cobra.Command, params *declineCmdParams) error {
	var (
		repo     = configutil.GetStringFlagOrDefault(cmd.Flags(), "repository", "")
		provider = configutil.GetStringFlagOrDefault(cmd.Flags(), "provider", "")
	)

	if (repo == "" && provider != "") || (repo != "" && provider == "") {
		return ErrSomeRepoParamsMissing
	}

	if repo != "" && provider != "" {
		v := strings.Split(repo, "/")
		if len(v) != 2 || v[0] == "" || v[1] == "" {
			return ErrRepositoryMustBeInFormOwnerRepo
		}

		params.Provider = provider
		params.Repository = repo
	}

	return nil
}

func maxPRDescriptionLength(prs []*client.PullRequest, limit int) int {
	maxLen := 0
	for _, pr := range prs {
		l := len(pr.Source) + len(pr.Destination) + 4
		if l > maxLen {
			maxLen = l
		}
	}

	if limit > 0 && maxLen > limit {
		return limit
	}

	return maxLen
}

type promptPullRequest struct {
	ID    string
	Title string
}

func getPromptPullRequestSilce(prs *client.PullRequestList) []*promptPullRequest {
	maxLen := maxPRDescriptionLength(prs.Values, 30)
	prFormat := fmt.Sprintf("#%%s: %%-%ds %%s %%s", maxLen)
	options := make([]*promptPullRequest, 0, len(prs.Values))
	for _, pr := range prs.Values {
		prDesc := fmt.Sprintf(
			prFormat,
			pr.ID,
			fmt.Sprintf("[%s->%s]", pr.Source, pr.Destination),
			pr.Updated.Format("(2006-01-02 15:04)"),
			pr.Title,
		)
		options = append(options, &promptPullRequest{
			ID:    pr.ID,
			Title: prDesc,
		})
	}

	return options
}

func promptPullRequestMultiSelect(prList *client.PullRequestList) map[string]*promptPullRequest {
	prs := getPromptPullRequestSilce(prList)

	answers := []string{}
	options := make([]string, 0, len(prs))
	for _, v := range prs {
		options = append(options, v.Title)
	}
	prompt := &survey.MultiSelect{
		Message:  "Decline pull requests",
		Options:  options,
		PageSize: 10,
	}
	survey.AskOne(prompt, &answers)

	selectedPRs := make(map[string]*promptPullRequest, len(answers))
	for _, a := range answers {
		for _, v := range prs {
			if v.Title == a {
				selectedPRs[v.ID] = v
			}
		}
	}

	return selectedPRs
}

var declineCmd = &cobra.Command{
	Use:     "decline",
	Aliases: []string{"del", "dec", "d"},
	Short:   "Decline pull request",
	Long:    `Declines a pull requests on the web service hosting your origin respository`,
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := ""
		if len(args) > 0 {
			id = args[0]
		}

		cl, err := client.DefaultClient()
		if err != nil {
			fmt.Println(err)
			os.Exit(systemcode.ErrorCodeGeneric)
		}

		params := &declineCmdParams{}
		fillDefaultDeclineCmdParams(params)
		err = fillFlagDeclineCmdParams(cmd, params)
		if err != nil {
			fmt.Println(err)
			os.Exit(3)
		}

		r, err := client.NewRepositoryFromOptions(&client.RepositoryOptions{
			Provider:           params.Provider,
			FullRepositoryName: params.Repository,
		})
		if err != nil {
			fmt.Println(err)
			os.Exit(systemcode.ErrorCodeGeneric)
		}

		if id != "" {
			_, err = cl.DeclinePullRequest(&client.DeclinePullRequestOptions{
				Repository: r,
				ID:         args[0],
			})
			if err != nil {
				fmt.Println(err)
				os.Exit(systemcode.ErrorCodeGeneric)
			}
		} else {
			prList, err := cl.GetPullRequests(&client.GetPullRequestsOptions{
				Repository: r,
				State:      client.PullRequestState_OPEN,
			})
			if err != nil {
				fmt.Println(err)
				os.Exit(systemcode.ErrorCodeGeneric)
			}

			selectedPRs := promptPullRequestMultiSelect(prList)
			processPullRequestMap(selectedPRs, cl, r, declinePR, func(msg interface{}) string {
				m := msg.(declineResponse)
				return fmt.Sprintf("Declining #%s... %s\n", m.ID, m.Status)
			})
		}
	},
}

func processPullRequestMap(
	selectedPRs map[string]*promptPullRequest,
	cl *client.Client,
	r *client.Repository,
	processFn func(cl *client.Client, r *client.Repository, id string, c chan interface{}),
	fn func(interface{}) string,
) {
	c := make(chan interface{})
	for id := range selectedPRs {
		go processFn(cl, r, id, c)
	}

	end := len(selectedPRs)
	count := 0
	for {
		msg := <-c
		count++
		fmt.Printf(fn(msg))

		if count >= end {
			break
		}
	}
}

type declineResponse struct {
	ID     string
	Status string
	Error  error
}

func declinePR(cl *client.Client, r *client.Repository, id string, c chan interface{}) {
	_, err := cl.DeclinePullRequest(&client.DeclinePullRequestOptions{
		Repository: r,
		ID:         id,
	})

	res := declineResponse{ID: id, Status: "Done"}
	if err != nil {
		res.Status = "Error"
		res.Error = err
	}

	c <- res
}

func init() {
	declineCmd.Flags().StringP("repository", "r", "", "repository in form of owner/repo")
	declineCmd.Flags().StringP("provider", "p", "", "repository host, values - (bitbucket-cloud)")
	rootCmd.AddCommand(declineCmd)
}
