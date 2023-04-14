package decline

import (
	"preq/internal/cli/paramutils"
	"preq/internal/cli/utils"
	"preq/internal/pkg/client"

	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "decline [ID]",
		Aliases: []string{"del", "dec", "d"},
		Short:   "Decline pull request",
		Long:    `Declines a pull requests on the web service hosting your origin repository`,
		Args:    cobra.MaximumNArgs(1),
		Run:     utils.RunCommandWrapper(runCmd),
	}

	return cmd
}

func runCmd(cmd *cobra.Command, args []string) error {
	cmdArgs := parseArgs(args)

	cl, repoParams, err := paramutils.GetClientAndRepoParams(cmd.Flags())
	if err != nil {
		return err
	}

	return execute(cl, cmdArgs, &client.Repository{
		Provider: repoParams.Provider,
		Name:     repoParams.Name,
	})
}

func execute(
	c client.Client,
	args *cmdArgs,
	repo *client.Repository,
) error {
	if args.ID != "" {
		_, err := c.DeclinePullRequest(&client.DeclinePullRequestOptions{
			Repository: repo,
			ID:         args.ID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

type ProcessPullRequestResponse struct {
	ID     string
	Status string
	Error  error
}

func declinePR(
	cl client.Client,
	r *client.Repository,
	id string,
	c chan interface{},
) {
	_, err := cl.DeclinePullRequest(&client.DeclinePullRequestOptions{
		Repository: r,
		ID:         id,
	})

	res := ProcessPullRequestResponse{ID: id, Status: "Done"}
	if err != nil {
		res.Status = "Error"
		res.Error = err
	}

	c <- res
}
