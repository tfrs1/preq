package decline

import (
	"fmt"
	"preq/internal/cli/paramutils"
	"preq/internal/errcodes"
	"preq/internal/gitutils"
	"preq/internal/pkg/client"
	"strings"
)

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

func fillDefaultDeclineCmdParams(params *cmdParams) {
	git, err := getWorkingDirectoryRepo()
	if err != nil {
		return
	}

	defaultRepo, err := git.GetRemoteInfo()
	if err != nil {
		return
	}

	params.Repository = fmt.Sprintf(
		"%s/%s",
		defaultRepo.Owner,
		defaultRepo.Name,
	)
	params.Provider = defaultRepo.Provider
}

func fillFlagDeclineCmdParams(flags paramutils.FlagSet, params *cmdParams) {
	var (
		repo     = flags.GetStringOrDefault("repository", params.Repository)
		provider = flags.GetStringOrDefault("provider", string(params.Provider))
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
