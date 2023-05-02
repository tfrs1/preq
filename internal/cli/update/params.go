package update

import (
	"preq/internal/configutils"
	"preq/internal/errcodes"
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

func fillFlagUpdateCmdParams(
	cmd *cobra.Command,
	params *updateCmdParams,
) error {
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

		params.Provider = provider
		params.Repository = repo
	}

	return nil
}
