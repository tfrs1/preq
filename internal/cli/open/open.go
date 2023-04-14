package open

import (
	"fmt"
	"log"
	"os/exec"
	"preq/internal/cli/paramutils"
	"preq/internal/cli/utils"
	"runtime"

	"github.com/spf13/cobra"
)

func runCmd(cmd *cobra.Command, args []string) error {
	cmdArgs := parseArgs(args)

	_, repoParams, err := paramutils.GetRepoUtilsAndParams(cmd.Flags())
	if err != nil {
		return err
	}
	params := &openCmdParams{
		Repository: *repoParams,
	}

	flags := &paramutils.PFlagSetWrapper{Flags: cmd.Flags()}
	fillDefaultOpenCmdParams(params)
	fillFlagOpenCmdParams(flags, params)

	return execute(cmdArgs, params)
}

func execute(args *cmdArgs, params *openCmdParams) error {
	url := fmt.Sprintf(
		"https://bitbucket.org/%s/pull-requests/",
		params.Repository.Name,
	)
	if args.ID != "" {
		url = fmt.Sprintf(
			"https://bitbucket.org/%s/pull-requests/%s",
			params.Repository.Name,
			args.ID,
		)
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
		Short:   "Open pull request's web page",
		Long:    `Opens all pull requests on the web service hosting your origin repository`,
		Run:     utils.RunCommandWrapper(runCmd),
	}

	cmd.Flags().Bool("print", false, "print the pull request URL")

	return cmd
}

func openInBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).
			Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}
