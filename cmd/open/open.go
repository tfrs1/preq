package open

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"preq/cmd/utils"
	"preq/internal/clientutils"
	"preq/internal/systemcodes"
	"preq/pkg/client"
	"runtime"

	"github.com/spf13/cobra"
)

func runCmd(cmd *cobra.Command, args []string) error {
	cmdArgs := parseArgs(args)

	params := &openCmdParams{}
	fillDefaultOpenCmdParams(params)
	fillFlagOpenCmdParams(cmd, params)
	err := validateFlagOpenCmdParams(params)
	if err != nil {
		return err
	}

	err = execute(cmdArgs, params)
	if err != nil {
		return err
	}

	return nil
}

func execute(args *cmdArgs, params *openCmdParams) error {
	url := fmt.Sprintf("https://bitbucket.org/%s/pull-requests/", params.Repository)
	if args.ID != "" {
		url = fmt.Sprintf("https://bitbucket.org/%s/pull-requests/%s", params.Repository, args.ID)
	} else if params.Interactive {
		cl, err := clientutils.ClientFactory{}.DefaultClient()
		if err != nil {
			return err
		}
		r, err := client.NewRepositoryFromOptions(&client.RepositoryOptions{
			Provider:           client.RepositoryProvider(params.Repository.Provider),
			FullRepositoryName: params.Repository.Name,
		})
		if err != nil {
			return err
		}
		prList, err := cl.GetPullRequests(&client.GetPullRequestsOptions{
			Repository: r,
			State:      client.PullRequestState_OPEN,
		})
		if err != nil {
			return err
		}

		selectedPR := utils.PromptPullRequestSelect(prList)
		if selectedPR == nil {
			os.Exit(systemcodes.ErrorCodeGeneric)
		}

		url = fmt.Sprintf("https://bitbucket.org/%s/pull-requests/%s", params.Repository, selectedPR.ID)
	}

	if params.PrintOnly {
		fmt.Println(url)
	} else {
		openInBrowser(url)
	}

	return nil
}

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "open [ID]",
		Aliases: []string{"o", "op"},
		Args:    cobra.MaximumNArgs(1),
		Short:   "Open pull requests",
		Long:    `Opens all pull requests on the web service hosting your origin repository`,
		Run:     utils.RunCommandWrapper(runCmd),
	}

	cmd.Flags().BoolP("interactive", "i", false, "interactive mode")
	cmd.Flags().Bool("print", false, "print the pull request URL")

	return cmd
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
