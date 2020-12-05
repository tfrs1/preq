package open

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"preq/internal/cli/paramutils"
	"preq/internal/cli/utils"
	"preq/internal/config"
	"preq/internal/domain/pullrequest"
	"runtime"

	"github.com/spf13/cobra"
)

func runCmd(cmd *cobra.Command, args []string) error {
	flags := &paramutils.PFlagSetWrapper{Flags: cmd.Flags()}
	c, _, err := config.LoadLocal(flags)
	if err != nil {
		fmt.Println("unknown error")
		os.Exit(123)
	}

	params := &openCmdParams{}
	cmdArgs := parseArgs(args)

	fillDefaultOpenCmdParams(params)
	fillFlagOpenCmdParams(cmd, params)
	err = validateFlagOpenCmdParams(params)
	if err != nil {
		return err
	}

	err = execute(c, cmdArgs, params)
	if err != nil {
		return err
	}

	return nil
}

func execute(c pullrequest.Repository, args *cmdArgs, params *openCmdParams) error {
	url := c.WebPageList()
	if args.ID != "" {
		url = c.WebPage(pullrequest.EntityID(args.ID))
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
