package cli

import (
	"fmt"
	"os"
	approvecmd "preq/internal/cli/approve"
	createcmd "preq/internal/cli/create"
	declinecmd "preq/internal/cli/decline"
	listcmd "preq/internal/cli/list"
	opencmd "preq/internal/cli/open"
	"preq/internal/cli/paramutils"
	"preq/internal/cli/utils"
	"preq/internal/gitutils"
	"preq/internal/persistance"
	"preq/internal/pkg/client"
	"preq/internal/tui"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func init() {
	log.Logger = log.Output(
		zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
		})

	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	// TODO: Read the log level value from the config
	levelInput := "debug"
	level, err := zerolog.ParseLevel(levelInput)
	if err != nil {
		log.Error().
			Msgf("unknown log level '%v', reverting to default error level", levelInput)
	} else {
		zerolog.SetGlobalLevel(level)
	}

	// TODO: Add full file logging behind a flag?
	// logFile, err := homedir.Expand("~/.local/state/preq/full.log")
	// if err == nil {
	// 	log.Logger = log.Output(
	// 		zerolog.ConsoleWriter{
	// 			Out: &lumberjack.Logger{
	// 				Filename:   logFile,
	// 				MaxSize:    20, // megabytes
	// 				MaxBackups: 3,
	// 				MaxAge:     28, //days
	// 				Compress:   false,
	// 			},
	// 			TimeFormat: time.RFC3339,
	// 		})
	// }
}

var rootCmd = &cobra.Command{
	Use:     "preq",
	Short:   "Pull request manager",
	Long:    "TUI utility for managing pull requests.",
	Version: fmt.Sprintf("%v, commit %v, built at %v", version, commit, date),
	Run: func(cmd *cobra.Command, args []string) {
		params := &paramutils.RepositoryParams{}
		flags := &paramutils.PFlagSetWrapper{Flags: cmd.Flags()}
		paramutils.FillDefaultRepositoryParams(params)
		paramutils.FillFlagRepositoryParams(flags, params)

		global, err := cmd.Flags().GetBool("global")
		if err != nil {
			global = false
			// TODO: Log error
		}

		repo, err := client.NewRepositoryFromOptions(
			&client.RepositoryOptions{
				Provider: client.RepositoryProvider(
					params.Provider,
				),
				FullRepositoryName: params.Name,
			},
		)

		if repo == nil && !global && err != nil {
			log.Error().Msg(err.Error())
			os.Exit(123)
		}

		// TODO: Check other configs (empty, no default, etc)
		// TODO: Do we even want a default?
		// TODO: preq create -d -s doesn't work without -r -p?!?
		// TODO: Remove update command?

		// Store working directory repo in visited state if startup config
		// is not global or repo is not explicit with -r and -p flags
		repoFlag, _ := cmd.Flags().GetString("repository")
		providerFlag, _ := cmd.Flags().GetString("provider")
		isExplicitRepo := repoFlag != "" && providerFlag != ""
		wd := ""
		if !global && !isExplicitRepo {
			wd, _ = os.Getwd()
		}

		isWdGitRepo := gitutils.IsDirGitRepo(wd)
		if isWdGitRepo {
			utils.WriteVisitToState(cmd.Flags(), params)
		} else {
			global = true
		}

		// Filter repos before sending them to TUI
		repos, err := persistance.GetDefault().GetVisited()
		if err != nil {
			log.Error().Msg(err.Error())
			return
		}

		if !global {
			filtered := make([]*persistance.PersistanceRepoInfo, 0)
			for _, v := range repos {
				if v.Provider == string(repo.Provider) &&
					v.Name == repo.Name {
					filtered = append(filtered, v)
				}
			}

			repos = filtered
		}

		tui.Run(params, repos)
	},
}

func Execute() {
	rootCmd.AddCommand(createcmd.New())
	rootCmd.AddCommand(approvecmd.New())
	rootCmd.AddCommand(declinecmd.New())
	rootCmd.AddCommand(listcmd.New())
	rootCmd.AddCommand(opencmd.New())

	//! Update command is not currently implemented
	// rootCmd.AddCommand(updatecmd.New())

	// TODO: Create config command?

	rootCmd.Flags().
		BoolP("global", "g", false, "Show information about all known (previously visited) repositories.")

	rootCmd.PersistentFlags().
		StringP("repository", "r", "", "repository in form of owner/repo")
	// TODO: Shorthand names for providers?
	rootCmd.PersistentFlags().
		StringP("provider", "p", "", "repository host, values - (bitbucket)")
	rootCmd.MarkFlagsRequiredTogether("repository", "provider")

	rootCmd.Execute()
}
