package close

import (
	"fmt"
	"os"
	"preq/internal/cli/paramutils"
	"preq/internal/cli/utils"
	"preq/internal/clientutils"
	"preq/internal/config"
	"preq/internal/domain/pullrequest"

	"github.com/spf13/cobra"
)

var processPullRequestMap = utils.ProcessPullRequestMap

func runCmd(cmd *cobra.Command, args []string) error {
	flags := &paramutils.PFlagSetWrapper{Flags: cmd.Flags()}
	c, err := clientutils.ClientFactory{}.DefaultWithFlags(flags)
	if err != nil {
		fmt.Println("unknown error")
		os.Exit(123)
	}

	cmdArgs := parseArgs(args)

	params := &cmdParams{}
	fillDefaultCloseCmdParams(params)
	fillFlagCloseCmdParams(&paramutils.PFlagSetWrapper{Flags: cmd.Flags()}, params)
	config.FillFlagRepositoryParams(flags, &params.Repository)
	err = validateFlagCloseCmdParams(params)
	if err != nil {
		return err
	}

	return execute(c, cmdArgs, params)
}

func execute(c pullrequest.Repository, args *cmdArgs, params *cmdParams) error {
	if args.ID != "" {
		service := pullrequest.NewCloseService(c)
		_, err := service.Close(&pullrequest.CloseOptions{
			ID: pullrequest.EntityID(args.ID),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "close [ID]",
		Aliases: []string{"delete", "del", "d", "decline"},
		Short:   "Close pull request",
		Long:    `Closes a pull requests on the web service hosting your origin repository`,
		Args:    cobra.MaximumNArgs(1),
		Run:     utils.RunCommandWrapper(runCmd),
	}

	return cmd
}

type closeResponse struct {
	ID     string
	Status string
	Error  error
}

// TODO: Maybe something like this for TUI close
// func closePR(cl pullrequest.Repository, r *client.Repository, id string, c chan interface{}) {
// 	_, err := cl.Close(&pullrequest.CloseOptions{
// 		ID: id,
// 	})

// 	res := closeResponse{ID: id, Status: "Done"}
// 	if err != nil {
// 		res.Status = "Error"
// 		res.Error = err
// 	}

// 	c <- res
// }
