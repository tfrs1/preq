package cmdcreate

import (
	"fmt"
	"preq/internal/cli/paramutils"
	"preq/internal/cli/utils"
	"preq/internal/domain/pullrequest"
	"preq/internal/gitutils"
	"preq/internal/pkg/client"

	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"cr"},
		Short:   "Create pull request",
		Long:    `Creates a pull request on the web service hosting your origin repository`,
		Run:     utils.RunCommandWrapper(runCmd),
	}

	setUpFlags(cmd)

	return cmd
}

func setUpFlags(cmd *cobra.Command) {
	cmd.Flags().
		StringP("destination", "d", "", "destination branch of your pull request")
	cmd.Flags().
		StringP("source", "s", "", "destination branch of your pull request (default checked out branch)")
	cmd.Flags().
		StringP("title", "t", "", "the title of the pull request (default last commit message)")
	// TODO: Open default editor for description?
	cmd.Flags().String("description", "", "the description of the pull request")
	cmd.Flags().Bool("close", true, "do not close source branch")
	cmd.Flags().Bool("draft", false, "mark the pull request as draft")
}

func runCmd(cmd *cobra.Command, args []string) error {
	flags := paramutils.PFlagSetWrapper{Flags: cmd.Flags()}

	var c gitutils.GitUtilsClient

	params := &createCmdParams{}
	c, repoParams, err := paramutils.GetRepoUtilsAndParams(cmd.Flags())
	if err != nil {
		return err
	}

	params.Repository = *repoParams

	fillInParamsFromFlags(&flags, params)
	fillInParamsFromRepo(c, params)
	// TODO: Also add fillInParamsFromConfig()?
	fillInDefaultParams(params)

	err = params.Validate()
	if err != nil {
		return err
	}

	cl, _, err := paramutils.GetClientAndRepoParams(cmd.Flags())
	if err != nil {
		return err
	}

	utils.SafelyWriteVisitToState(cmd.Flags(), &params.Repository)

	return execute(cl, params)
}

type creatorAdapter struct {
	Client client.Client
}

func (ca *creatorAdapter) Create(
	o *pullrequest.CreateOptions,
) (*pullrequest.Entity, error) {
	cpro := &client.CreatePullRequestOptions{
		Repository:  o.Repository,
		CloseBranch: o.CloseBranch,
		Destination: o.Destination,
		Source:      o.Source,
		Title:       o.Title,
		Draft:       o.Draft,
	}

	pr, err := ca.Client.CreatePullRequest(cpro)
	if err != nil {
		return nil, err
	}

	return &pullrequest.Entity{
		// TODO: Does not make sense that this type has hash (when creating a pull request)
		Destination: pr.Destination.Name,
		Source:      pr.Source.Name,
		Title:       pr.Title,
		URL:         pr.URL,
	}, nil
}

func execute(c client.Client, params *createCmdParams) error {
	ca := &creatorAdapter{Client: c}

	service := pullrequest.NewCreateService(ca)
	pr, err := service.Create(&pullrequest.CreateOptions{
		Repository: &client.Repository{
			Provider: params.Repository.Provider,
			Name:     params.Repository.Name,
		},
		CloseBranch: params.CloseBranch,
		Title:       params.Title,
		Source:      params.Source,
		Destination: params.Destination,
		Draft:       params.Draft,
	})
	if err != nil {
		return err
	}

	fmt.Printf(
		"Created a pull request: %s -> %s\n",
		pr.Source,
		pr.Destination,
	)
	fmt.Println("  ", pr.URL)

	return nil
}
