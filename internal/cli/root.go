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
	"preq/internal/configutils"
	"preq/internal/domain"
	"preq/internal/pkg/client"
	"preq/internal/pkg/github"
	"preq/internal/tui"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func loadConfig() (domain.PullRequestRepository, *client.Repository, error) {
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

func (ms *MockStorage) Get() string {
	return ""
}

func (ms *MockStorage) RefreshPullRequestData(c domain.PullRequestRepository) {

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
		path, err := cmd.Flags().GetString("config")
		if err != nil {
			fmt.Println(err)
			os.Exit(3)
		}

		err = configutils.LoadDefault(path)
		if err != nil {
			fmt.Println(err)
			os.Exit(3)
		}

		c, r, err := config.LoadLocal()
		if err != nil {
			// TODO: Do something
		}

		fmt.Println(c, r)
		c, _, err = loadConfig()
		if err != nil {
			os.Exit(123)
		}

		ghc := &github.GithubCloudClient{
			Repository: domain.GitRepository{
				Name: "tfrs1/preq",
			},
			Username: "tfrs1",
			Token:    "a99fd333ea529996de5cc8374473a5436ece2ef0",
		}

		tui := tui.NewTui([]domain.PullRequestRepository{ghc})
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
	rootCmd.AddCommand(createcmd.New())
	rootCmd.AddCommand(approvecmd.New())
	rootCmd.AddCommand(declinecmd.New())
	rootCmd.AddCommand(listcmd.New())
	rootCmd.AddCommand(opencmd.New())
	rootCmd.AddCommand(updatecmd.New())

	rootCmd.PersistentFlags().StringP("repository", "r", "", "repository in form of owner/repo")
	// TODO: Shorthand names for providers?
	rootCmd.PersistentFlags().StringP("provider", "p", "", "repository host, values - (bitbucket)")
	rootCmd.PersistentFlags().String("config", config.DEFAULT_CONFIG_PATH, "default config path")

	rootCmd.Execute()
}
