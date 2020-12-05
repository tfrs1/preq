package close

import (
	"fmt"
	"preq/internal/cli/paramutils"
	"preq/internal/config"
	"preq/internal/errcodes"
	"preq/internal/gitutils"
	"preq/internal/pkg/client"
	"strings"
)

var getRemoteInfo = gitutils.GetRemoteInfo

type cmdParams struct {
	Repository config.RepositoryParams
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

func fillDefaultCloseCmdParams(params *cmdParams) {
	defaultRepo, err := getRemoteInfo()
	if err == nil {
		params.Repository.Name = fmt.Sprintf("%s/%s", defaultRepo.Owner, defaultRepo.Name)
		params.Repository.Provider = defaultRepo.Provider
	}
}

func fillFlagCloseCmdParams(flags paramutils.FlagSet, params *cmdParams) {
	var (
		repo     = flags.GetStringOrDefault("repository", params.Repository.Name)
		provider = flags.GetStringOrDefault("provider", string(params.Repository.Provider))
	)

	params.Repository.Name = repo
	params.Repository.Provider = client.RepositoryProvider(provider)
}

var validateFlagCloseCmdParams = func(params *cmdParams) error {
	if (params.Repository.Name == "" && params.Repository.Provider != "") || (params.Repository.Name != "" && params.Repository.Provider == "") {
		return errcodes.ErrSomeRepoParamsMissing
	}

	if params.Repository.Name != "" && params.Repository.Provider != "" {
		v := strings.Split(params.Repository.Name, "/")
		if len(v) != 2 || v[0] == "" || v[1] == "" {
			return errcodes.ErrRepositoryMustBeInFormOwnerRepo
		}

		if !params.Repository.Provider.IsValid() {
			return errcodes.ErrorRepositoryProviderUnknown
		}
	}

	return nil
}
