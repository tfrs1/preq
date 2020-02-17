package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"prctl/internal/configutil"
	"prctl/internal/gitutil"
	"prctl/internal/systemcode"
	client "prctl/pkg/bitbucket"
	"strings"

	"github.com/gosuri/uilive"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
)

type listCmdParams struct {
	Provider   string
	Repository string
}

func fillDefaultListCmdParams(params *listCmdParams) {
	defaultRepo, err := gitutil.GetRemoteInfo()
	if err == nil {
		params.Repository = fmt.Sprintf("%s/%s", defaultRepo.Owner, defaultRepo.Name)
		params.Provider = string(defaultRepo.Provider)
	}
}

func fillFlagListCmdParams(cmd *cobra.Command, params *listCmdParams) error {
	var (
		repo     = configutil.GetStringFlagOrDefault(cmd.Flags(), "repository", "")
		provider = configutil.GetStringFlagOrDefault(cmd.Flags(), "provider", "")
	)

	if (repo == "" && provider != "") || (repo != "" && provider == "") {
		return ErrSomeRepoParamsMissing
	}

	if repo != "" && provider != "" {
		v := strings.Split(repo, "/")
		if len(v) != 2 || v[0] == "" || v[1] == "" {
			return ErrRepositoryMustBeInFormOwnerRepo
		}

		params.Provider = provider
		params.Repository = repo
	}

	return nil
}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List pull requests",
	Long:    `Lists all pull requests on the web service hosting your origin respository`,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := client.DefaultClient()
		if err != nil {
			fmt.Println(err)
			os.Exit(systemcode.ErrorCodeGeneric)
		}

		params := &listCmdParams{}
		fillDefaultListCmdParams(params)
		err = fillFlagListCmdParams(cmd, params)
		if err != nil {
			fmt.Println(err)
			os.Exit(3)
		}

		nextURL := ""
		reader := bufio.NewReader(os.Stdin)

		writer := uilive.New()
		defer writer.Stop()
		writer.Start()

		table := uitable.New()
		table.AddRow("#", "TITLE", "SRC/DEST", "URL")
		table.AddRow("-", "-----", "--------", "---")

		r, err := client.NewRepositoryFromOptions(&client.RepositoryOptions{
			Provider:           params.Provider,
			FullRepositoryName: params.Repository,
		})
		if err != nil {
			fmt.Println(err)
			os.Exit(systemcode.ErrorCodeGeneric)
		}

		for {
			prs, err := c.GetPullRequests(&client.GetPullRequestsOptions{
				Repository: r,
				State:      client.PullRequestState_OPEN,
				Next:       nextURL,
			})

			if err != nil {
				fmt.Println(err)
				os.Exit(systemcode.ErrorCodeGeneric)
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
	},
}

func clearLine(out io.Writer) {
	var clear = fmt.Sprintf("%c[%dA%c[2K", 27, 1, 27)
	_, _ = fmt.Fprint(out, clear)
}

func init() {
	listCmd.Flags().StringP("repository", "r", "", "repository in form of owner/repo")
	listCmd.Flags().StringP("provider", "p", "", "repository host, values - (bitbucket-cloud)")
	rootCmd.AddCommand(listCmd)
}
