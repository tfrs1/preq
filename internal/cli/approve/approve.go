package approve

import (
	"preq/internal/cli/paramutils"
	"preq/internal/cli/utils"
	"preq/internal/clientutils"
	"preq/internal/config"
	"preq/internal/domain/pullrequest"
	"preq/internal/pkg/client"

	"github.com/spf13/cobra"
)

func runCmd(cmd *cobra.Command, args []string) error {
	cmdArgs := parseArgs(args)
	flags := &paramutils.PFlagSetWrapper{Flags: cmd.Flags()}

	params := &approveCmdParams{}
	config.FillDefaultRepositoryParams(&params.Repository)
	paramutils.FillFlagRepositoryParams(flags, &params.Repository)
	err := paramutils.ValidateFlagRepositoryParams(&params.Repository)
	if err != nil {
		return err
	}

	cl, err := clientutils.ClientFactory{}.DefaultPullRequestRepository(params.Repository.Provider)
	if err != nil {
		return err
	}

	r, err := client.NewRepositoryFromOptions(&client.RepositoryOptions{
		Provider:           client.RepositoryProvider(params.Repository.Provider),
		FullRepositoryName: params.Repository.Name,
	})
	if err != nil {
		return err
	}

	err = execute(cl, cmdArgs, params, r)
	if err != nil {
		return err
	}

	return nil
}

func execute(c pullrequest.Repository, args *cmdArgs, params *approveCmdParams, repo *client.Repository) error {
	if args.ID != "" {
		_, err := c.Approve(&pullrequest.ApproveOptions{
			// Repository: repo,
			ID: args.ID,
		})
		if err != nil {
			return err
		}
	} else {
		// prList, err := c.Get(&pullrequest.GetOptions{
		// 	// Repository: repo,
		// 	State: client.PullRequestState_OPEN,
		// })
		// if err != nil {
		// 	return err
		// }

		// selectedPRs := utils.PromptPullRequestMultiSelect(prList)
		// utils.ProcessPullRequestMap(selectedPRs, c, repo, approvePR, func(msg interface{}) string {
		// 	m := msg.(approveResponse)
		// 	return fmt.Sprintf("Approving #%s... %s\n", m.ID, m.Status)
		// })
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

func approvePR(cl pullrequest.Repository, r *client.Repository, id string, c chan interface{}) {
	_, err := cl.Approve(&pullrequest.ApproveOptions{
		// Repository: r,
		ID: id,
	})

	res := approveResponse{ID: id, Status: "Done"}
	if err != nil {
		res.Status = "Error"
		res.Error = err
	}

	c <- res
}
