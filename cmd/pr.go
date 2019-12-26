package cmd

import (
	"log"
	"prctl/internal/configutil"
	"prctl/internal/gitutil"
	client "prctl/pkg/bitbucket"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	prCmd.Flags().StringP("destination", "d", "", "destination branch of your pull request")
	prCmd.Flags().StringP("source", "s", "", "destination branch of your pull request (default checked out branch)")
	prCmd.Flags().StringP("repository", "r", "", "repository in form of owner/repo")
	// TODO: Shorthand names for providers?
	prCmd.Flags().StringP("provider", "p", "", "repository host, e.g. bitbucket")
	// TODO: Lookup last commit message
	prCmd.Flags().StringP("title", "t", "Title missing", "repository host, e.g. bitbucket")
	// TODO: Open default editor for description?
	prCmd.Flags().String("description", "", "repository host, e.g. bitbucket")
	rootCmd.AddCommand(prCmd)
}

var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "Create pull request",
	Long:  `Creates a pull request on the web service hosting your origin respository`,
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
