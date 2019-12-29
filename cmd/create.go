package cmd

import (
	"prctl/internal/configutil"
	"prctl/internal/gitutil"
	client "prctl/pkg/bitbucket"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	createCmd.Flags().StringP("destination", "d", "", "destination branch of your pull request")
	createCmd.Flags().StringP("source", "s", "", "destination branch of your pull request (default checked out branch)")
	createCmd.Flags().StringP("repository", "r", "", "repository in form of owner/repo")
	// TODO: Shorthand names for providers?
	createCmd.Flags().StringP("provider", "p", "", "repository host, e.g. bitbucket")
	// TODO: Lookup last commit message
	createCmd.Flags().StringP("title", "t", "Title missing", "repository host, e.g. bitbucket")
	// TODO: Open default editor for description?
	createCmd.Flags().String("description", "", "repository host, e.g. bitbucket")
	rootCmd.AddCommand(createCmd)
}

var createCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"cr"},
	Short:   "Create pull request",
	Long:    `Creates a pull request on the web service hosting your origin respository`,
	Run: func(cmd *cobra.Command, args []string) {
		c := client.New(&client.ClientOptions{
			Username: viper.GetString("bitbucket.username"),
			Password: viper.GetString("bitbucket.password"),
		})

		var (
			repo        = configutil.GetStringFlagOrDie(cmd, "repository")
			destination = configutil.GetStringFlagOrDie(cmd, "destination")
			owner       string
			repoName    string
		)

		v := strings.Split(repo, "/")
		if len(v) != 2 {
			log.Fatal("repo must be in form owner/repo")
		}
		owner, repoName = v[0], v[1]

		branch, err := gitutil.GetBranch()
		if err != nil {
			log.Fatal(err)
		}
		source := configutil.GetStringFlagOrDefault(cmd, "source", branch)

		c.CreatePullRequest(&client.CreatePullRequestOptions{
			Repository: &client.Repository{
				Owner: owner,
				Name:  repoName,
			},
			Title:       cmd.Flags().Lookup("title").Value.String(),
			Source:      source,
			Destination: destination,
		})
	},
}
