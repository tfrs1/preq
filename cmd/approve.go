package cmd

import (
	"fmt"
	"os"
	"prctl/internal/configutil"
	"prctl/internal/gitutil"
	"prctl/internal/systemcode"
	client "prctl/pkg/bitbucket"
	"strings"

	"github.com/spf13/cobra"
)

type approveCmdParams struct {
	Provider   string
	Repository string
}

func fillDefaultApproveCmdParams(params *approveCmdParams) {
	defaultRepo, err := gitutil.GetRemoteInfo()
	if err == nil {
		params.Repository = fmt.Sprintf("%s/%s", defaultRepo.Owner, defaultRepo.Name)
		params.Provider = string(defaultRepo.Provider)
	}
}

func fillFlagApproveCmdParams(cmd *cobra.Command, params *approveCmdParams) error {
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

var approveCmd = &cobra.Command{
	Use:     "approve",
	Aliases: []string{"del", "dec", "d"},
	Short:   "Approve pull request",
	Long:    `Approves a pull requests on the web service hosting your origin respository`,
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

		params := &approveCmdParams{}
		fillDefaultApproveCmdParams(params)
		err = fillFlagApproveCmdParams(cmd, params)
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
			_, err = cl.ApprovePullRequest(&client.ApprovePullRequestOptions{
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
			processPullRequestMap(selectedPRs, cl, r, approvePR, func(msg interface{}) string {
				m := msg.(approveResponse)
				return fmt.Sprintf("Approving #%s... %s\n", m.ID, m.Status)
			})
		}
	},
}

type approveResponse struct {
	ID     string
	Status string
	Error  error
}

func approvePR(cl *client.Client, r *client.Repository, id string, c chan interface{}) {
	_, err := cl.ApprovePullRequest(&client.ApprovePullRequestOptions{
		Repository: r,
		ID:         id,
	})

	res := approveResponse{ID: id, Status: "Done"}
	if err != nil {
		res.Status = "Error"
		res.Error = err
	}

	c <- res
}

func init() {
	approveCmd.Flags().StringP("repository", "r", "", "repository in form of owner/repo")
	approveCmd.Flags().StringP("provider", "p", "", "repository host, values - (bitbucket-cloud)")
	rootCmd.AddCommand(approveCmd)
}
