package cmdcreate

import (
	"fmt"
	"os"
	"preq/internal/cli/paramutils"
	"preq/internal/cli/utils"
	"preq/internal/config"
	"preq/internal/domain/pullrequest"

	"github.com/spf13/cobra"
)

func setUpFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("destination", "d", "", "destination branch of your pull request")
	cmd.Flags().StringP("source", "s", "", "destination branch of your pull request (default checked out branch)")
	cmd.Flags().StringP("title", "t", "", "the title of the pull request (default last commit message)")
	// TODO: Open default editor for description?
	cmd.Flags().String("description", "", "the description of the pull request")
	// TODO: Deep link interactive in tui?
	// cmd.Flags().BoolP("interactive", "i", false, "the description of the pull request")
	cmd.Flags().Bool("close", true, "do not close source branch")
	cmd.Flags().Bool("draft", false, "mark the pull request as draft")
}

func runCmd(cmd *cobra.Command, args []string) error {
	flags := &paramutils.PFlagSetWrapper{Flags: cmd.Flags()}
	c, err := config.LoadLocal(flags)
	if err != nil {
		fmt.Println("unknown error")
		os.Exit(123)
	}

	params := &createCmdParams{}

	populateParams(params, flags)
	config.FillFlagRepositoryParams(flags, &params.Repository)

	err = params.Validate()
	if err != nil {
		return err
	}

	co := &pullrequest.CreateOptions{
		Title:       params.Title,
		Source:      params.Source,
		Destination: params.Destination,
		CloseBranch: params.CloseBranch,
		Draft:       params.Draft,
	}

	return execute(c, co)
}

func execute(c pullrequest.Creator, params *pullrequest.CreateOptions) error {
	service := pullrequest.NewCreateService(c)
	pr, err := service.Create(params)

	// TODO: Consider implementation like the following
	// commandInvoker := getCommandInvoker()
	// cmd := pullrequest.CreateCommand(c, params)
	// commandInvoker.invoke(cmd)

	if err != nil {
		return err
	}

	fmt.Println(fmt.Sprintf("Created a pull request: %s -> %s", pr.Source, pr.Destination))
	fmt.Println("  ", pr.URL)

	return nil
}

// New creates an instance of create command
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
