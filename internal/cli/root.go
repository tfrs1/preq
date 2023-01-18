package cli

import (
	"fmt"
	"os"
	"path/filepath"
	approvecmd "preq/internal/cli/approve"
	createcmd "preq/internal/cli/create"
	declinecmd "preq/internal/cli/decline"
	listcmd "preq/internal/cli/list"
	opencmd "preq/internal/cli/open"
	"preq/internal/cli/paramutils"
	updatecmd "preq/internal/cli/update"
	"preq/internal/persistance"
	"preq/internal/pkg/client"
	"preq/internal/tui"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog"
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
	if err == nil {
		log.Logger = log.Output(
			zerolog.ConsoleWriter{
				Out: &lumberjack.Logger{
					Filename:   logFile,
					MaxSize:    20, // megabytes
					MaxBackups: 3,
					MaxAge:     28, //days
					Compress:   false,
				},
				TimeFormat: time.RFC3339,
			})
	}
}

var rootCmd = &cobra.Command{
	Use:     "preq",
	Short:   "Pull request manager",
	Long:    "TUI utility for managing pull requests.",
	Version: fmt.Sprintf("%v, commit %v, built at %v", version, commit, date),
	Run: func(cmd *cobra.Command, args []string) {
		params := &paramutils.RepositoryParams{}
		paramutils.FillDefaultRepositoryParams(params)
		flags := &paramutils.PFlagSetWrapper{Flags: cmd.Flags()}
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
		// TODO: remove update function?

		// Store working directory repo in visited state if startup config
		// is not global or repo is not explicit with -r and -p flags
		repoFlag, _ := cmd.Flags().GetString("repository")
		providerFlag, _ := cmd.Flags().GetString("provider")
		isExplicitRepo := repoFlag != "" && providerFlag != ""
		wd := ""
		if !global && !isExplicitRepo {
			wd, _ = os.Getwd()
		}

		isWdGitRepo := false
		if wd != "" {
			info, err := os.Stat(filepath.Join(wd, ".git"))
			if err == nil && info.IsDir() {
				isWdGitRepo = true
			}
		}

		if isWdGitRepo {
			err := persistance.GetRepo().AddVisited(
				fmt.Sprintf("%s/%s", repo.Owner, repo.Name),
				string(repo.Provider),
				wd,
			)
			if err != nil {
				log.Error().Msg(err.Error())
				return
			}
		} else {
			global = true
		}

		// Filter repos before sending them to TUI
		repos, err := persistance.GetRepo().GetVisited()
		if err != nil {
			log.Error().Msg(err.Error())
			return
		}

		if !global {
			filtered := make([]*persistance.PersistanceRepoInfo, 0)
			for _, v := range repos {
				if v.Provider == string(repo.Provider) &&
					v.Name == fmt.Sprintf("%s/%s", repo.Owner, repo.Name) {
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
	rootCmd.AddCommand(updatecmd.New())

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
