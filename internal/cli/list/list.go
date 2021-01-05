package list

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"preq/internal/cli/paramutils"
	"preq/internal/cli/utils"
	"preq/internal/clientutils"
	"preq/internal/domain/pullrequest"
	"preq/internal/pkg/client"

	"github.com/gosuri/uilive"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
)

func runCmd(cmd *cobra.Command, args []string) error {
	flags := &paramutils.PFlagSetWrapper{Flags: cmd.Flags()}
	c, err := clientutils.ClientFactory{}.DefaultWithFlags(flags)
	if err != nil {
		fmt.Println("unknown error")
		os.Exit(123)
	}

	return execute(c)
}

func execute(c pullrequest.Repository) error {
	list, err := c.Get(&pullrequest.GetOptions{
		State: client.PullRequestState_OPEN,
	})
	if err != nil {
		return err
	}

	reader := bufio.NewReader(os.Stdin)
	writer := uilive.New()
	defer writer.Stop()
	writer.Start()

	table := uitable.New()
	table.AddRow("#", "TITLE", "SRC/DEST", "URL")
	table.AddRow("-", "-----", "--------", "---")

	for {
		prs, err := list.Next()
		if err != nil {
			return err
		}

		for _, v := range prs {
			table.AddRow(
				v.ID,
				v.Title,
				fmt.Sprintf("%s -> %s", v.Source, v.Destination),
				v.URL,
			)
		}

		fmt.Fprintln(writer, table.String())

		if !list.HasNext() {
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

// TODO: Add attribution
func clearLine(out io.Writer) {
	var clear = fmt.Sprintf("%c[%dA%c[2K", 27, 1, 27)
	_, _ = fmt.Fprint(out, clear)
}
