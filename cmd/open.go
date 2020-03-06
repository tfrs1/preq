package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"preq/internal/configutil"
	"preq/internal/gitutil"
	"preq/internal/systemcode"
	client "preq/pkg/bitbucket"
	"runtime"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

type openCmdParams struct {
	Provider    string
	Repository  string
	PrintOnly   bool
	Interactive bool
}

func fillDefaultOpenCmdParams(params *openCmdParams) {
	params.PrintOnly = false
	params.Interactive = false
	defaultRepo, err := gitutil.GetRemoteInfo()
	if err == nil {
		params.Repository = fmt.Sprintf("%s/%s", defaultRepo.Owner, defaultRepo.Name)
		params.Provider = string(defaultRepo.Provider)
	}
}

func fillFlagOpenCmdParams(cmd *cobra.Command, params *openCmdParams) error {
	var (
		repo        = configutil.GetStringFlagOrDefault(cmd.Flags(), "repository", "")
		provider    = configutil.GetStringFlagOrDefault(cmd.Flags(), "provider", "")
		printOnly   = configutil.GetBoolFlagOrDefault(cmd.Flags(), "print", false)
		interactive = configutil.GetBoolFlagOrDefault(cmd.Flags(), "interactive", false)
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

	params.PrintOnly = printOnly
	params.Interactive = interactive

	return nil
}

func openInBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

var openCmd = &cobra.Command{
	Use:     "open [ID]",
	Aliases: []string{"op"},
	Args:    cobra.MaximumNArgs(1),
	Short:   "List pull requests",
	Long:    `Lists all pull requests on the web service hosting your origin repository`,
	Run: func(cmd *cobra.Command, args []string) {
		id := ""
		if len(args) > 0 {
			id = args[0]
		}

		params := &openCmdParams{}
		fillDefaultOpenCmdParams(params)
		err := fillFlagOpenCmdParams(cmd, params)
		if err != nil {
			fmt.Println(err)
			os.Exit(3)
		}

		url := fmt.Sprintf("https://bitbucket.org/%s/pull-requests/", params.Repository)
		if id != "" {
			url = fmt.Sprintf("https://bitbucket.org/%s/pull-requests/%s", params.Repository, id)
		} else if params.Interactive {
			cl, err := client.DefaultClient()
			if err != nil {
				fmt.Println(err)
				os.Exit(systemcode.ErrorCodeGeneric)
			}
			r, err := client.NewRepositoryFromOptions(&client.RepositoryOptions{
				Provider:           params.Provider,
				FullRepositoryName: params.Repository,
			})
			if err != nil {
				fmt.Println(err)
				os.Exit(systemcode.ErrorCodeGeneric)
			}
			prList, err := cl.GetPullRequests(&client.GetPullRequestsOptions{
				Repository: r,
				State:      client.PullRequestState_OPEN,
			})
			if err != nil {
				fmt.Println(err)
				os.Exit(systemcode.ErrorCodeGeneric)
			}

			selectedPR := promptPullRequestSelect(prList)
			url = fmt.Sprintf("https://bitbucket.org/%s/pull-requests/%s", params.Repository, selectedPR.ID)
		}

		if params.PrintOnly {
			fmt.Println(url)
		} else {
			openInBrowser(url)
		}
	},
}

func promptPullRequestSelect(prList *client.PullRequestList) *promptPullRequest {
	prs := getPromptPullRequestSilce(prList)

	answer := ""
	options := make([]string, 0, len(prs))
	for _, v := range prs {
		options = append(options, v.Title)
	}
	prompt := &survey.Select{
		Message:  "Decline pull requests",
		Options:  options,
		PageSize: 10,
	}
	survey.AskOne(prompt, &answer)

	for _, v := range prs {
		if v.Title == answer {
			return v
		}
	}

	return nil
}

func init() {
	openCmd.Flags().StringP("repository", "r", "", "repository in form of owner/repo")
	openCmd.Flags().StringP("provider", "p", "", "repository host, values - (bitbucket-cloud)")
	openCmd.Flags().BoolP("interactive", "i", false, "interactive mode")
	openCmd.Flags().Bool("print", false, "print the pull request URL")
	rootCmd.AddCommand(openCmd)
}
