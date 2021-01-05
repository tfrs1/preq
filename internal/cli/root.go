package cli

import (
	"fmt"
	"os"

	approvecmd "preq/internal/cli/approve"
	closecmd "preq/internal/cli/close"
	createcmd "preq/internal/cli/create"
	listcmd "preq/internal/cli/list"
	opencmd "preq/internal/cli/open"
	"preq/internal/cli/paramutils"
	updatecmd "preq/internal/cli/update"
	"preq/internal/clientutils"
	"preq/internal/config"
	"preq/internal/configutils"
	"preq/internal/domain"
	"preq/internal/domain/pullrequest"
	"preq/internal/pkg/client"
	"preq/internal/tui"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func loadConfig() (pullrequest.Repository, error) {
	params := &config.RepositoryParams{}
	config.FillDefaultRepositoryParams(params)

	_, err := client.NewRepositoryFromOptions(&client.RepositoryOptions{
		Provider:           params.Provider,
		FullRepositoryName: params.Name,
	})
	if err != nil {
		return nil, err
	}

	c, err := clientutils.ClientFactory{}.DefaultPullRequestRepository(params.Provider)
	if err != nil {
		return nil, err
	}

	return c, nil
}

type MockStorage struct{}

func (ms *MockStorage) Get() string {
	return ""
}

func (ms *MockStorage) RefreshPullRequestData(c pullrequest.Repository) {

}

func NewStorage() domain.Storage {
	return &MockStorage{}
}

var rootCmd = &cobra.Command{
	Use:     "preq",
	Short:   "preq command-line utility for pull requests",
	Long:    `Command-line utility for all your pull request needs.`,
	Version: fmt.Sprintf("%v, commit %v, built at %v", version, commit, date),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		path, err := cmd.Flags().GetString("config")
		if err != nil {
			fmt.Println(err)
			os.Exit(3)
		}

		err = configutils.LoadGlobal(path)
		if err != nil {
			fmt.Println(err)
			os.Exit(3)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		flags := &paramutils.PFlagSetWrapper{Flags: cmd.Flags()}
		c, err := config.LoadLocal(flags)
		if err != nil {
			// TODO: Do something
			os.Exit(123)
		}

		tui := tui.NewTui([]pullrequest.Repository{c})
		tui.Start()
		// storage := NewStorage()
		// domain := &domain.Domain{
		// 	// Client: c,
		// 	// Repository: repo,
		// 	Presenter: tui,
		// 	Storage:   storage,
		// }

		// domain.Present()
		// domain.subscribe(presenter)
		// domain.Present()

		// selectRow0 := selectRowCommand{RowID: 0}

		// domain.StoreAndExecute(selectRow0)
	},
}

func Execute() {
	rootCmd.AddCommand(
		createcmd.New(),
		approvecmd.New(),
		closecmd.New(),
		listcmd.New(),
		opencmd.New(),
		updatecmd.New(),
	)

	rootCmd.PersistentFlags().StringP("repository", "r", "", "repository in form of owner/repo")
	// TODO: Shorthand names for providers?
	rootCmd.PersistentFlags().StringP("provider", "p", "", "repository host, values - (bitbucket)")
	rootCmd.PersistentFlags().String("config", "", "config path")

	rootCmd.Execute()
}
