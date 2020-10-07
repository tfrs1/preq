package update

import (
	"fmt"
	"preq/internal/configutils"
	"preq/internal/errcodes"
	"preq/internal/gitutils"
	"strings"

	"github.com/spf13/cobra"
)

type updateCmdParams struct {
	Provider                 string
	Repository               string
	EnableDraft              bool
	DisableDraft             bool
	EnableCloseSourceBranch  bool
	DisableCloseSourceBranch bool
}

func fillDefaultUpdateCmdParams(params *updateCmdParams) {
	defaultRepo, err := gitutils.GetRemoteInfo()
	if err == nil {
		params.Repository = fmt.Sprintf("%s/%s", defaultRepo.Owner, defaultRepo.Name)
		params.Provider = string(defaultRepo.Provider)
	}
}

func fillFlagUpdateCmdParams(cmd *cobra.Command, params *updateCmdParams) error {
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

		params.Provider = provider
		params.Repository = repo
	}

	return nil
}
