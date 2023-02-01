package list

import (
	"preq/internal/cli/paramutils"
	"preq/internal/configutils"
	"preq/internal/errcodes"
	"preq/internal/gitutils"
	"preq/internal/pkg/client"
	"strings"

	"github.com/spf13/cobra"
)

type listCmdParams struct {
	Repository paramutils.RepositoryParams
}

var getWorkingDirectoryRepo = gitutils.GetWorkingDirectoryRepo

func fillDefaultListCmdParams(params *listCmdParams) {
	git, err := getWorkingDirectoryRepo()
	if err != nil {
		return
	}

	defaultRepo, err := git.GetRemoteInfo()
	if err != nil {
		return
	}

	params.Repository.Name = defaultRepo.Name
	params.Repository.Provider = defaultRepo.Provider
}

func fillFlagListCmdParams(cmd *cobra.Command, params *listCmdParams) error {
	var (
		repo = configutils.GetStringFlagOrDefault(
			cmd.Flags(),
			"repository",
			"",
		)
		provider = configutils.GetStringFlagOrDefault(
			cmd.Flags(),
			"provider",
			"",
		)
	)

	if repo != "" && provider != "" {
		v := strings.Split(repo, "/")
		if len(v) != 2 || v[0] == "" || v[1] == "" {
			return errcodes.ErrRepositoryMustBeInFormOwnerRepo
		}

		params.Repository.Provider = client.RepositoryProvider(provider)
		params.Repository.Name = repo
	}

	return nil
}
