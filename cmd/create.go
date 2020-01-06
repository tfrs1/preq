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
	ErrMissingRepository               = errors.New("repository is missing")
	ErrMissingDestination              = errors.New("destination is missing")
	ErrMissingBitbucketUsername        = errors.New("butbucket username is missing")
	ErrMissingBitbucketPassword        = errors.New("butbucket password is missing")
	ErrSomeRepoParamsMissing           = errors.New("must specify both provider and repository, or none")
	ErrRepositoryMustBeInFormOwnerRepo = errors.New("repository must be in the form of 'owner/repo'")
)

func getRepo(cmd *cobra.Command) (*client.Repository, error) {
	var (
		repo     = configutil.GetStringFlagOrDefault(cmd.Flags(), "repository", "")
		provider = configutil.GetStringFlagOrDefault(cmd.Flags(), "provider", "")
	)

	if repo != "" && provider != "" {
		v := strings.Split(repo, "/")
		if len(v) != 2 {
			return nil, ErrRepositoryMustBeInFormOwnerRepo
		}

		p, err := client.ParseRepositoryProvider(provider)
		if err != nil {
			return nil, err
		}

		return &client.Repository{
			Provider: p,
			Owner:    v[0],
			Name:     v[1],
		}, nil
	} else if repo == "" && provider == "" {
		defaultRepo, err := gitutil.GetRemoteInfo()
		if err != nil {
			return nil, err
		}

		return defaultRepo, nil
	} else {
		return nil, ErrSomeRepoParamsMissing
	}
}

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

		repo, err := getRepo(cmd)
		if err != nil {
			fmt.Println(err)
			os.Exit(3)
		}

		defaultSourceBranch, err := gitutil.GetCurrentBranch()
		if err != nil {
			fmt.Println(err)
			os.Exit(3)
		}

		source := configutil.GetStringFlagOrDefault(cmd.Flags(), "source", defaultSourceBranch)
		destination := configutil.GetStringFlagOrDefault(cmd.Flags(), "destination", "")
		if destination == "" {
			destination, err = gitutil.GetClosestBranch([]string{"master", "develop"})
			if err != nil {
				fmt.Println("destinatin branch and specified and cannot be automatically resolved")
				os.Exit(3)
			}
		}

		pr, err := c.CreatePullRequest(&client.CreatePullRequestOptions{
			Repository:  repo,
			Title:       cmd.Flags().Lookup("title").Value.String(),
			Source:      source,
			Destination: destination,
		})

		if err != nil {
			fmt.Println(err)
			os.Exit(3)
		}

		fmt.Println(fmt.Sprintf("Created a pull request: %s -> %s", pr.Source, pr.Destination))
		fmt.Println("  ", pr.URL)
	},
}
