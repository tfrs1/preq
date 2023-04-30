package cli

import (
	"fmt"
	"io"
	"os"
	"path"
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

	"github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:     "preq",
	Short:   "Pull request manager",
	Long:    "TUI utility for managing pull requests.",
	Version: fmt.Sprintf("%v, commit %v, built at %v", version, commit, date),
	Run: func(cmd *cobra.Command, args []string) {
		params := &paramutils.RepositoryParams{}
		flags := &paramutils.PFlagSetWrapper{Flags: cmd.Flags()}
		paramutils.FillFlagRepositoryParams(flags, params)

		global, err := cmd.Flags().GetBool("global")
		if err != nil {
			global = false
			// TODO: Log error
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
			utils.SafelyWriteVisitToState(cmd.Flags(), params)
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
			// TODO: Add params only utils method? Without flags param?
			_, params, err := paramutils.GetRepoUtilsAndParams(&pflag.FlagSet{})
			if err != nil {
				log.Error().Msg(err.Error())
				os.Exit(5)
			}

			repo, err := client.NewRepositoryFromOptions(
				&client.RepositoryOptions{
					Provider: client.RepositoryProvider(
						params.Provider,
					),
					Name: params.Name,
				},
			)

			if repo == nil && !global && err != nil {
				log.Error().Msg(err.Error())
				os.Exit(123)
			}

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
	dir, err := homedir.Expand(`~/.local/state/preq`)
	if err != nil {
		panic(err)
	}

	os.MkdirAll(dir, 0o700)
	file, err := os.OpenFile(
		path.Join(dir, "app.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0o0600,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to open log file")
	}
	defer file.Close()

	mw := io.MultiWriter(
		// TODO: Write to a in memory file writer? To show errors (logs) in the app, or maybe just use a hook or something?
		zerolog.ConsoleWriter{
			Out:        file,
			TimeFormat: time.RFC3339,
		},
	)

	log.Logger = zerolog.New(mw).With().Timestamp().Logger()

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
