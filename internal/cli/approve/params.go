package approve

import (
	"preq/internal/cli/paramutils"
	"preq/internal/errcodes"
	"preq/internal/gitutils"
	"preq/internal/pkg/client"
	"strings"
)

type approveCmdParams struct {
	Repository paramutils.RepositoryParams
}

var getWorkingDirectoryRepo = gitutils.GetWorkingDirectoryRepo

type cmdParams struct {
	Provider   client.RepositoryProvider
	Repository string
}

type cmdArgs struct {
	ID string
}

func parseArgs(args []string) *cmdArgs {
	id := ""
	if len(args) > 0 {
		id = args[0]
	}

	return &cmdArgs{ID: id}
}

func fillFlagDeclineCmdParams(flags paramutils.FlagRepo, params *cmdParams) {
	var (
		repo     = flags.GetStringOrDefault("repository", "")
		provider = flags.GetStringOrDefault("provider", "")
	)

	params.Repository = repo
	params.Provider = client.RepositoryProvider(provider)
}

var validateFlagDeclineCmdParams = func(params *cmdParams) error {
	if params.Repository != "" && params.Provider != "" {
		v := strings.Split(params.Repository, "/")
		if len(v) != 2 || v[0] == "" || v[1] == "" {
			return errcodes.ErrRepositoryMustBeInFormOwnerRepo
		}

		if !params.Provider.IsValid() {
			return errcodes.ErrorRepositoryProviderUnknown
		}
	}

	return nil
}
