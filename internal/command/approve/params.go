package approve

// import (
// 	"preq/internal/command/paramutils"
// )

// package decline

import (
	"fmt"
	"preq/internal/command/paramutils"
	"preq/internal/errcodes"
	"preq/internal/gitutils"
	"preq/pkg/client"
	"strings"
)

type approveCmdParams struct {
	Repository paramutils.RepositoryParams
}

var getRemoteInfo = gitutils.GetRemoteInfo

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
	defaultRepo, err := getRemoteInfo()
	if err == nil {
		params.Repository = fmt.Sprintf("%s/%s", defaultRepo.Owner, defaultRepo.Name)
		params.Provider = defaultRepo.Provider
	}
}

func fillFlagDeclineCmdParams(flags paramutils.FlagSet, params *cmdParams) {
	var (
		repo     = flags.GetStringOrDefault("repository", "")
		provider = flags.GetStringOrDefault("provider", "")
	)

	params.Repository = repo
	params.Provider = client.RepositoryProvider(provider)
}

var validateFlagDeclineCmdParams = func(params *cmdParams) error {
	if (params.Repository == "" && params.Provider != "") || (params.Repository != "" && params.Provider == "") {
		return errcodes.ErrSomeRepoParamsMissing
	}

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
