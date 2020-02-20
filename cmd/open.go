package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"prctl/internal/configutil"
	"prctl/internal/gitutil"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

type openCmdParams struct {
	Provider   string
	Repository string
	PrintOnly  bool
}

func fillDefaultOpenCmdParams(params *openCmdParams) {
	params.PrintOnly = false
	defaultRepo, err := gitutil.GetRemoteInfo()
	if err == nil {
		params.Repository = fmt.Sprintf("%s/%s", defaultRepo.Owner, defaultRepo.Name)
		params.Provider = string(defaultRepo.Provider)
	}
}

func fillFlagOpenCmdParams(cmd *cobra.Command, params *openCmdParams) error {
	var (
		repo      = configutil.GetStringFlagOrDefault(cmd.Flags(), "repository", "")
		provider  = configutil.GetStringFlagOrDefault(cmd.Flags(), "provider", "")
		printOnly = configutil.GetBoolFlagOrDefault(cmd.Flags(), "print", false)
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

	return nil
}

func openBrowser(url string) {
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
	Use:   "open",
	Short: "List pull requests",
	Long:  `Lists all pull requests on the web service hosting your origin respository`,
	Run: func(cmd *cobra.Command, args []string) {
		params := &openCmdParams{}
		fillDefaultOpenCmdParams(params)
		err := fillFlagOpenCmdParams(cmd, params)
		if err != nil {
			fmt.Println(err)
			os.Exit(3)
		}

		url := fmt.Sprintf("https://bitbucket.org/%s/pull-requests/", params.Repository)
		if params.PrintOnly {
			fmt.Println(url)
		} else {
			openBrowser(url)
		}
	},
}

func init() {
	openCmd.Flags().StringP("repository", "r", "", "repository in form of owner/repo")
	openCmd.Flags().StringP("provider", "p", "", "repository host, values - (bitbucket-cloud)")
	openCmd.Flags().Bool("print", false, "print the pull request URL")
	rootCmd.AddCommand(openCmd)
}
