package cmdcreate

import (
	"fmt"
	"preq/internal/cli/paramutils"
	"preq/internal/cli/utils"
	"preq/internal/clientutils"
	"preq/internal/domain/pullrequest"
	"preq/internal/pkg/client"

	"github.com/spf13/cobra"
)

func setUpFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("destination", "d", "", "destination branch of your pull request")
	cmd.Flags().StringP("source", "s", "", "destination branch of your pull request (default checked out branch)")
	cmd.Flags().StringP("title", "t", "", "the title of the pull request (default last commit message)")
	// TODO: Open default editor for description?
	cmd.Flags().String("description", "", "the description of the pull request")
	cmd.Flags().BoolP("interactive", "i", false, "the description of the pull request")
	cmd.Flags().Bool("close", true, "do not close source branch")
	cmd.Flags().Bool("draft", false, "mark the pull request as draft")
}

func runCmd(cmd *cobra.Command, args []string) error {
	flags := paramutils.PFlagSetWrapper{Flags: cmd.Flags()}
	params := &createCmdParams{}
	fillDefaultParams(params)
	fillFlagParams(&flags, params)

	interactive := flags.GetBoolOrDefault("interactive", false)
	if interactive {
		err := fillInteractiveParams(params)
		if err != nil {
			return err
		}
	}

	err := params.Validate()
	if err != nil {
		return err
	}

	c, err := clientutils.ClientFactory{}.DefaultClientCustom(
		params.Repository.Provider,
		params.Repository.Name,
	)

	if err != nil {
		return err
	}

	err = execute(c, params)
	if err != nil {
		return err
	}

	return nil
}

type creatorAdatapter struct {
	Client client.Client
}

func (ca *creatorAdatapter) Create(o *pullrequest.CreateOptions) (*pullrequest.Entity, error) {
	cpro := &client.CreatePullRequestOptions{
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
		Destination: pr.Destination,
		Source:      pr.Source,
		Title:       pr.Title,
		URL:         pr.URL,
	}, nil
}

func execute(c client.Client, params *createCmdParams) error {
	ca := &creatorAdatapter{Client: c}

	service := pullrequest.NewCreateService(ca)
	pr, err := service.Create(&pullrequest.CreateOptions{
		CloseBranch: params.CloseBranch,
		Title:       params.Title,
		Source:      params.Source,
		Destination: params.Destination,
		Draft:       params.Draft,
	})

	if err != nil {
		return err
	}

	fmt.Println(fmt.Sprintf("Created a pull request: %s -> %s", pr.Source, pr.Destination))
	fmt.Println("  ", pr.URL)

	return nil
}

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
