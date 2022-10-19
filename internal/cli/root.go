package cli

import (
	"fmt"

	approvecmd "preq/internal/cli/approve"
	createcmd "preq/internal/cli/create"
	declinecmd "preq/internal/cli/decline"
	listcmd "preq/internal/cli/list"
	opencmd "preq/internal/cli/open"
	"preq/internal/cli/paramutils"
	updatecmd "preq/internal/cli/update"
	"preq/internal/tui"

	"github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func init() {
	logFile, err := homedir.Expand("~/.local/state/preq/full.log")
	if err != nil {
		fileLogger := &lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    20, // megabytes
			MaxBackups: 3,
			MaxAge:     28, //days
			// Compress:   true, // disabled by default
		}
		log.Logger = log.Output(fileLogger)
	}
}

var rootCmd = &cobra.Command{
	Use:     "preq",
	Short:   "preq command-line utility for pull requests",
	Long:    `Command-line utility for all your pull request needs.`,
	Version: fmt.Sprintf("%v, commit %v, built at %v", version, commit, date),
	Run: func(cmd *cobra.Command, args []string) {
		params := &paramutils.RepositoryParams{}
		paramutils.FillDefaultRepositoryParams(params)
		flags := &paramutils.PFlagSetWrapper{Flags: cmd.Flags()}
		paramutils.FillFlagRepositoryParams(flags, params)

		tui.Run(params)
	},
}

func Execute() {
	rootCmd.AddCommand(createcmd.New())
	rootCmd.AddCommand(approvecmd.New())
	rootCmd.AddCommand(declinecmd.New())
	rootCmd.AddCommand(listcmd.New())
	rootCmd.AddCommand(opencmd.New())
	rootCmd.AddCommand(updatecmd.New())

	rootCmd.PersistentFlags().
		StringP("repository", "r", "", "repository in form of owner/repo")
	// TODO: Shorthand names for providers?
	rootCmd.PersistentFlags().
		StringP("provider", "p", "", "repository host, values - (bitbucket)")

	rootCmd.Execute()
}
