package update

import (
	"fmt"

	"github.com/spf13/cobra"
)

func runCmd(cmd *cobra.Command, args []string) {
	id := ""
	if len(args) > 0 {
		id = args[0]
	}
	fmt.Println(id)
	fmt.Println(cmd.Flags().Changed("draft"))

	// TODO: Implement update command
	// cl, err := clientutils.ClientFactory{}.DefaultPullRequestRepository()
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(systemcodes.ErrorCodeGeneric)
	// }

	// params := &updateCmdParams{}
	// fillDefaultUpdateCmdParams(params)
	// err = fillFlagUpdateCmdParams(cmd, params)
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(3)
	// }

	// r, err := client.NewRepositoryFromOptions(&client.RepositoryOptions{
	// 	Provider:           params.Provider,
	// 	FullRepositoryName: params.Repository,
	// })
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(systemcodes.ErrorCodeGeneric)
	// }

	// if id != "" {
	// 	_, err = cl.Approve(&client.ApprovePullRequestOptions{
	// 		Repository: r,
	// 		ID:         args[0],
	// 	})
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		os.Exit(systemcodes.ErrorCodeGeneric)
	// 	}
	// } else {
	// 	prList, err := cl.Get(&client.GetPullRequestOptions{
	// 		Repository: r,
	// 		State:      client.PullRequestState_OPEN,
	// 	})
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		os.Exit(systemcodes.ErrorCodeGeneric)
	// 	}

	// 	selectedPRs := promptPullRequestMultiSelect(prList)
	// 	processPullRequestMap(selectedPRs, cl, r, approvePR, func(msg interface{}) string {
	// 		m := msg.(approveResponse)
	// 		return fmt.Sprintf("Approving #%s... %s\n", m.ID, m.Status)
	// 	})
	// }
}

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update [ID]",
		Aliases: []string{"up"},
		Short:   "Update pull request",
		Long:    `Updates a pull requests on the web service hosting your origin repository`,
		Args:    cobra.MaximumNArgs(1),
		Run:     runCmd,
	}

	cmd.Flags().Bool("draft", false, "repository host, values - (bitbucket)")
	cmd.Flags().Bool("close", false, "repository host, values - (bitbucket)")

	return cmd
}
