package approve

import (
	"preq/internal/cli/paramutils"
	"preq/internal/cli/utils"
	"preq/internal/pkg/client"

	"github.com/spf13/cobra"
)

func runCmd(cmd *cobra.Command, args []string) error {
	cmdArgs := parseArgs(args)

	cl, repoParams, err := paramutils.GetClientAndRepoParams(cmd.Flags())
	if err != nil {
		return err
	}

	utils.SafelyWriteVisitToState(cmd.Flags(), repoParams)

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
		_, err := c.Approve(&client.ApproveOptions{
			Repository: repo,
			ID:         args.ID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func New() *cobra.Command {
	approveCmd := &cobra.Command{
		Use:     "approve [ID]",
		Aliases: []string{"ap"},
		Short:   "Approve pull request",
		Long:    `Approves a pull requests on the web service hosting your origin repository`,
		Args:    cobra.MaximumNArgs(1),
		Run:     utils.RunCommandWrapper(runCmd),
	}

	return approveCmd
}

type approveResponse struct {
	ID     string
	Status string
	Error  error
}

func approvePR(
	cl client.Client,
	r *client.Repository,
	id string,
	c chan interface{},
) {
	_, err := cl.Approve(&client.ApproveOptions{
		Repository: r,
		ID:         id,
	})

	res := approveResponse{ID: id, Status: "Done"}
	if err != nil {
		res.Status = "Error"
		res.Error = err
	}

	c <- res
}
