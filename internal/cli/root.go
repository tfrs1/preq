package cli

import (
	"fmt"
	"os"

	approvecmd "preq/internal/cli/approve"
	createcmd "preq/internal/cli/create"
	declinecmd "preq/internal/cli/decline"
	listcmd "preq/internal/cli/list"
	opencmd "preq/internal/cli/open"
	updatecmd "preq/internal/cli/update"
	"preq/internal/clientutils"
	"preq/internal/config"
	"preq/internal/domain"
	"preq/internal/pkg/client"
	"preq/internal/tui"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func loadConfig() (domain.Client, *client.Repository, error) {
	params := &config.RepositoryParams{}
	config.FillDefaultRepositoryParams(params)

	r, err := client.NewRepositoryFromOptions(&client.RepositoryOptions{
		Provider:           params.Provider,
		FullRepositoryName: params.Name,
	})
	if err != nil {
		return nil, nil, err
	}

	c, err := clientutils.ClientFactory{}.DefaultClient(params.Provider)
	if err != nil {
		return nil, nil, err
	}

	return c, r, nil
}

type MockStorage struct{}

func (ms *MockStorage) GetPullRequests() string {
	return ""
}

func (ms *MockStorage) RefreshPullRequestData(c domain.Client) {

}

func NewStorage() domain.Storage {
	return &MockStorage{}
}

var rootCmd = &cobra.Command{
	Use:     "preq",
	Short:   "preq command-line utility for pull requests",
	Long:    `Command-line utility for all your pull request needs.`,
	Version: fmt.Sprintf("%v, commit %v, built at %v", version, commit, date),
	Run: func(cmd *cobra.Command, args []string) {
		c, _, err := loadConfig()
		if err != nil {
			os.Exit(123)
		}

		tui := tui.NewTui(c)
		storage := NewStorage()
		domain := &domain.Domain{
			Client: c,
			// Repository: repo,
			Presenter: tui,
			Storage:   storage,
		}

		domain.Present()
		// domain.subscribe(presenter)
		// domain.Present()

		// selectRow0 := selectRowCommand{RowID: 0}

		// domain.StoreAndExecute(selectRow0)
	},
}

func Execute() {
	rootCmd.AddCommand(createcmd.New())
	rootCmd.AddCommand(approvecmd.New())
	rootCmd.AddCommand(declinecmd.New())
	rootCmd.AddCommand(listcmd.New())
	rootCmd.AddCommand(opencmd.New())
	rootCmd.AddCommand(updatecmd.New())

	rootCmd.PersistentFlags().StringP("repository", "r", "", "repository in form of owner/repo")
	// TODO: Shorthand names for providers?
	rootCmd.PersistentFlags().StringP("provider", "p", "", "repository host, values - (bitbucket)")

	rootCmd.Execute()
}
