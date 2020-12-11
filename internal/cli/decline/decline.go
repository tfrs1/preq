package decline

import (
	"preq/internal/cli/paramutils"
	"preq/internal/cli/utils"
	"preq/internal/clientutils"
	"preq/internal/domain"
	"preq/internal/pkg/client"

	"github.com/spf13/cobra"
)

var promptPullRequestMultiSelect = utils.PromptPullRequestMultiSelect
var processPullRequestMap = utils.ProcessPullRequestMap

func runCmd(cmd *cobra.Command, args []string) error {
	cmdArgs := parseArgs(args)

	params := &cmdParams{}
	fillDefaultDeclineCmdParams(params)
	fillFlagDeclineCmdParams(&paramutils.PFlagSetWrapper{Flags: cmd.Flags()}, params)
	err := validateFlagDeclineCmdParams(params)
	if err != nil {
		return err
	}

	cl, err := clientutils.ClientFactory{}.DefaultClient(params.Provider)
	if err != nil {
		return err
	}

	r, err := client.NewRepositoryFromOptions(&client.RepositoryOptions{
		Provider:           params.Provider,
		FullRepositoryName: params.Repository,
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

func execute(c domain.PullRequestRepository, args *cmdArgs, params *cmdParams, repo *client.Repository) error {
	if args.ID != "" {
		_, err := c.Decline(&domain.DeclinePullRequestOptions{
			// Repository: repo,
			ID: args.ID,
		})
		if err != nil {
			return err
		}
	} else {
		// prList, err := c.Get(&domain.GetPullRequestOptions{
		// 	// Repository: repo,
		// 	State: client.PullRequestState_OPEN,
		// })
		// if err != nil {
		// 	return err
		// }

		// selectedPRs := promptPullRequestMultiSelect(prList)
		// processPullRequestMap(
		// 	selectedPRs,
		// 	c,
		// 	repo,
		// 	declinePR,
		// 	func(msg interface{}) string {
		// 		m := msg.(declineResponse)
		// 		return fmt.Sprintf("Declining #%s... %s\n", m.ID, m.Status)
		// 	},
		// )
	}

	return nil
}

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

type declineResponse struct {
	ID     string
	Status string
	Error  error
}

func declinePR(cl domain.PullRequestRepository, r *client.Repository, id string, c chan interface{}) {
	_, err := cl.Decline(&domain.DeclinePullRequestOptions{
		// Repository: r,
		ID: id,
	})

	res := declineResponse{ID: id, Status: "Done"}
	if err != nil {
		res.Status = "Error"
		res.Error = err
	}

	c <- res
}
