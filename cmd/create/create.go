package cmdcreate

import (
	"fmt"
	"preq/cmd/paramutils"
	"preq/cmd/utils"
	"preq/internal/clientutils"
	"preq/pkg/client"
	"strings"

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
	cmd.Flags().Bool("wip", false, "mark the pull request as Work-In-Progress")
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

	c, err := clientutils.ClientFactory{}.DefaultClient()
	if err != nil {
		return err
	}

	err = execute(c, params)
	if err != nil {
		return err
	}

	return nil
}

func execute(c client.Client, params *createCmdParams) error {
	title := params.Title
	if params.WorkInProgress {
		title = fmt.Sprintf("[WIP] %s", title)
	}

	r := strings.Split(params.Repository.Name, "/")
	pr, err := c.CreatePullRequest(&client.CreatePullRequestOptions{
		Repository: &client.Repository{
			Provider: client.RepositoryProvider(params.Repository.Provider),
			Owner:    r[0],
			Name:     r[1],
		},
		CloseBranch: params.CloseBranch,
		Title:       title,
		Source:      params.Source,
		Destination: params.Destination,
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
