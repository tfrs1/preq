package list

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"preq/internal/cli/paramutils"
	"preq/internal/cli/utils"
	"preq/internal/clientutils"
	"preq/internal/pkg/client"
	"preq/internal/systemcodes"

	"github.com/gosuri/uilive"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
)

func runCmd(cmd *cobra.Command, args []string) error {
	flags := &paramutils.PFlagSetWrapper{Flags: cmd.Flags()}

	params := &listCmdParams{}
	_, err := paramutils.GetRepoAndFillRepoParams(flags, &params.Repository)
	if err != nil {
		return err
	}

	c, err := clientutils.ClientFactory{}.DefaultClient(
		params.Repository.Provider,
	)
	if err != nil {
		return err
	}

	r, err := client.NewRepositoryFromOptions(&client.RepositoryOptions{
		Provider:           params.Repository.Provider,
		FullRepositoryName: params.Repository.Name,
	})
	if err != nil {
		return err
	}

	err = execute(c, params, r)
	if err != nil {
		return err
	}

	return nil
}

func execute(
	c client.Client,
	params *listCmdParams,
	repo *client.Repository,
) error {
	nextURL := ""
	reader := bufio.NewReader(os.Stdin)

	writer := uilive.New()
	defer writer.Stop()
	writer.Start()

	table := uitable.New()
	table.AddRow("#", "TITLE", "SRC/DEST", "URL")
	table.AddRow("-", "-----", "--------", "---")

	for {
		prs, err := c.GetPullRequests(&client.GetPullRequestsOptions{
			Repository: repo,
			State:      client.PullRequestState_OPEN,
			Next:       nextURL,
		})

		if err != nil {
			fmt.Println(err)
			os.Exit(systemcodes.ErrorCodeGeneric)
		}

		nextURL = prs.NextURL

		for _, v := range prs.Values {
			table.AddRow(
				v.ID,
				v.Title,
				fmt.Sprintf("%s -> %s", v.Source, v.Destination),
				v.URL,
			)
		}

		fmt.Fprintln(writer, table.String())

		if nextURL == "" {
			break
		}

		moreMsg := "Press Enter to show more..."
		fmt.Fprintln(writer.Newline(), moreMsg)

		_, _, err = reader.ReadRune()
		if err != nil {
			fmt.Println(err)
			break
		}

		// Clear the additional line from loading more request (Enter)
		clearLine(writer.Out)

		loadingMsg := "Loading..."
		fmt.Fprintln(writer, table.String())
		fmt.Fprintln(writer.Newline(), loadingMsg)
	}

	return nil
}

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List pull requests",
		Long:    `Lists all pull requests on the web service hosting your origin repository`,
		Run:     utils.RunCommandWrapper(runCmd),
	}

	return cmd
}

func clearLine(out io.Writer) {
	var clear = fmt.Sprintf("%c[%dA%c[2K", 27, 1, 27)
	_, _ = fmt.Fprint(out, clear)
}
