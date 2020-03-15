package list

import (
	"fmt"
	"preq/cmd/paramutils"
	"preq/internal/configutils"
	"preq/internal/errcodes"
	"preq/internal/gitutils"
	client "preq/pkg/bitbucket"
	"strings"

	"github.com/spf13/cobra"
)

type listCmdParams struct {
	Repository paramutils.RepositoryParams
}

func fillDefaultListCmdParams(params *listCmdParams) {
	defaultRepo, err := gitutils.GetRemoteInfo()
	if err == nil {
		params.Repository.Name = fmt.Sprintf("%s/%s", defaultRepo.Owner, defaultRepo.Name)
		params.Repository.Provider = defaultRepo.Provider
	}
}

func fillFlagListCmdParams(cmd *cobra.Command, params *listCmdParams) error {
	var (
		repo     = configutils.GetStringFlagOrDefault(cmd.Flags(), "repository", "")
		provider = configutils.GetStringFlagOrDefault(cmd.Flags(), "provider", "")
	)

	if (repo == "" && provider != "") || (repo != "" && provider == "") {
		return errcodes.ErrSomeRepoParamsMissing
	}

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
