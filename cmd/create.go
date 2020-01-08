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
	createCmd.Flags().BoolP("interactive", "i", false, "the description of the pull request")
	rootCmd.AddCommand(createCmd)
}

var (
	ErrMissingRepository               = errors.New("repository is missing")
	ErrMissingProvider                 = errors.New("provider is missing")
	ErrMissingSource                   = errors.New("source is missing")
	ErrMissingDestination              = errors.New("destination is missing")
	ErrSomeRepoParamsMissing           = errors.New("must specify both provider and repository, or none")
	ErrRepositoryMustBeInFormOwnerRepo = errors.New("repository must be in the form of 'owner/repo'")
)

func fillFlagParams(cmd *cobra.Command, params *createCmdParams) error {
	var (
		repo        = configutil.GetStringFlagOrDefault(cmd.Flags(), "repository", "")
		provider    = configutil.GetStringFlagOrDefault(cmd.Flags(), "provider", "")
		source      = configutil.GetStringFlagOrDefault(cmd.Flags(), "source", params.Source)
		destination = configutil.GetStringFlagOrDefault(cmd.Flags(), "destination", params.Destination)
	)

	if (repo == "" && provider != "") || (repo != "" && provider == "") {
		return ErrSomeRepoParamsMissing
	}

	if repo != "" && provider != "" {
		v := strings.Split(repo, "/")
		if len(v) != 2 || v[0] == "" || v[1] == "" {
			return ErrRepositoryMustBeInFormOwnerRepo
		}

		params.Provider = provider
		params.Repository = repo
	}

	params.Source = source
	params.Destination = destination

	return nil
}

func fillDefaultParams(params *createCmdParams) {
	defaultSourceBranch, err := gitutil.GetCurrentBranch()
	if err == nil {
		params.Source = defaultSourceBranch
	}

	destination, err := gitutil.GetClosestBranch([]string{"master", "develop"})
	if err == nil {
		params.Destination = destination
	}

	defaultRepo, err := gitutil.GetRemoteInfo()
	if err == nil {
		params.Repository = fmt.Sprintf("%s/%s", defaultRepo.Owner, defaultRepo.Name)
		params.Provider = string(defaultRepo.Provider)
	}
}

type createCmdParams struct {
	Provider    string
	Repository  string
	Source      string
	Destination string
	Title       string
}

func (c *createCmdParams) Validate() error {
	if c.Source == "" {
		return ErrMissingSource
	}
	if c.Destination == "" {
		return ErrMissingDestination
	}
	if c.Repository == "" {
		return ErrMissingRepository
	}
	if c.Provider == "" {
		return ErrMissingProvider
	}
	return nil
}

var createCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"cr"},
	Short:   "Create pull request",
	Long:    `Creates a pull request on the web service hosting your origin respository`,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := client.DefaultClient()
		if err != nil {
			fmt.Println(err)
			os.Exit(3)
		}

		interactive, err := cmd.Flags().GetBool("interactive")
		if err != nil {
			fmt.Println(err)
			os.Exit(3)
		}

		params := &createCmdParams{}
		fillDefaultParams(params)
		err = fillFlagParams(cmd, params)
		if err != nil {
			fmt.Println(err)
			os.Exit(3)
		}

		if interactive {
			err = fillInteractiveParams(params)
			if err != nil {
				fmt.Println(err)
				os.Exit(3)
			}
		}

		err = params.Validate()
		if err != nil {
			fmt.Println(err)
			os.Exit(3)
		}

		r := strings.Split(params.Repository, "/")
		pr, err := c.CreatePullRequest(&client.CreatePullRequestOptions{
			Repository: &client.Repository{
				Provider: client.RepositoryProvider(params.Provider),
				Owner:    r[0],
				Name:     r[1],
			},
			Title:       cmd.Flags().Lookup("title").Value.String(),
			Source:      params.Source,
			Destination: params.Destination,
		})

		if err != nil {
			fmt.Println(err)
			os.Exit(3)
		}

		fmt.Println(fmt.Sprintf("Created a pull request: %s -> %s", pr.Source, pr.Destination))
		fmt.Println("  ", pr.URL)
	},
}
