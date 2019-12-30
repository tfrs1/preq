package cmd

import (
	"fmt"
	"os"
	"prctl/internal/configutil"
	"prctl/internal/gitutil"
	client "prctl/pkg/bitbucket"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	createCmd.Flags().StringP("destination", "d", "", "destination branch of your pull request")
	createCmd.Flags().StringP("source", "s", "", "destination branch of your pull request (default checked out branch)")
	createCmd.Flags().StringP("repository", "r", "", "repository in form of owner/repo")
	// TODO: Shorthand names for providers?
	createCmd.Flags().StringP("provider", "p", "", "repository host, values - (bitbucket-cloud)")
	// TODO: Lookup last commit message
	createCmd.Flags().StringP("title", "t", "Created with prctl", "the title of the pull request")
	// TODO: Open default editor for description?
	createCmd.Flags().String("description", "", "the description of the pull request")
	rootCmd.AddCommand(createCmd)
}

var (
	ErrMissingRepository        = errors.New("repository is missing")
	ErrMissingDestination       = errors.New("destination is missing")
	ErrMissingBitbucketUsername = errors.New("butbucket username is missing")
	ErrMissingBitbucketPassword = errors.New("butbucket password is missing")
)

var createCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"cr"},
	Short:   "Create pull request",
	Long:    `Creates a pull request on the web service hosting your origin respository`,
	Run: func(cmd *cobra.Command, args []string) {
		c := client.New(&client.ClientOptions{
			Username: configutil.GetStringOrDie(
				viper.GetString("bitbucket.username"),
				ErrMissingBitbucketUsername,
			),
			Password: configutil.GetStringOrDie(
				viper.GetString("bitbucket.password"),
				ErrMissingBitbucketPassword,
			),
		})

		var (
			repo        = configutil.GetStringFlagOrDie(cmd, "repository", ErrMissingRepository)
			destination = configutil.GetStringFlagOrDie(cmd, "destination", ErrMissingDestination)
			owner       string
			repoName    string
		)

		v := strings.Split(repo, "/")
		if len(v) != 2 {
			fmt.Println("repository must be in the form of 'owner/repo'")
			os.Exit(3)
		}
		owner, repoName = v[0], v[1]

		branch, err := gitutil.GetBranch()
		if err != nil {
			fmt.Println(err)
			os.Exit(3)
		}
		source := configutil.GetStringFlagOrDefault(cmd, "source", branch)

		_, err = c.CreatePullRequest(&client.CreatePullRequestOptions{
			Repository: &client.Repository{
				Owner: owner,
				Name:  repoName,
			},
			Title:       cmd.Flags().Lookup("title").Value.String(),
			Source:      source,
			Destination: destination,
		})

		if err != nil {
			fmt.Println(err)
			os.Exit(3)
		}
	},
}
