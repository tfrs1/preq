package cmdcreate

import (
	"fmt"
	"preq/internal/cli/paramutils"
	"preq/internal/errcodes"
	"preq/internal/gitutils"
)

type createCmdParams struct {
	Repository  paramutils.RepositoryParams
	Source      string
	Destination string
	Title       string
	CloseBranch bool
	Draft       bool
}

func (params *createCmdParams) Validate() error {
	err := paramutils.ValidateFlagRepositoryParams(&params.Repository)
	if err != nil {
		return err
	}

	if params.Source == "" {
		return errcodes.ErrMissingSource
	}
	if params.Destination == "" {
		return errcodes.ErrMissingDestination
	}
	if params.Repository.Name == "" {
		return errcodes.ErrMissingRepository
	}
	if params.Repository.Provider == "" {
		return errcodes.ErrMissingProvider
	}
	if params.Title == "" {
		return errcodes.ErrMissingTitle
	}
	return nil
}

func fillInParamsFromFlags(flags paramutils.FlagRepo, params *createCmdParams) {
	if params.Source == "" {
		params.Source = flags.GetStringOrDefault("source", params.Source)
	}

	if params.Destination == "" {
		params.Destination = flags.GetStringOrDefault(
			"destination",
			params.Destination,
		)
	}

	if params.Title == "" {
		params.Title = flags.GetStringOrDefault("title", params.Title)
	}

	params.CloseBranch = flags.GetBoolOrDefault("close", params.CloseBranch)
	params.Draft = flags.GetBoolOrDefault("draft", params.Draft)
}

func fillInParamsFromRepo(
	git gitutils.GitUtilsClient,
	params *createCmdParams,
) {
	defaultSourceBranch, err := git.GetCurrentBranch()
	if params.Source == "" && err == nil {
		params.Source = defaultSourceBranch
	}

	// TODO: Make closest branch list configurable
	// TODO: From branch needs to be a parameter otherwise -r and -p server no purpose
	destination, err := git.GetClosestBranch([]string{"develop", "master"})
	if params.Destination == "" && err == nil {
		params.Destination = destination
	}

	title, err := git.GetBranchLastCommitMessage(params.Source)
	if params.Title == "" && err == nil {
		params.Title = title
	}
}

func fillInDefaultParams(params *createCmdParams) {
	if params.Source == "" {
		params.Source = "develop"
	}

	if params.Destination == "" {
		params.Destination = "master"
	}

	if params.Title == "" {
		params.Title = fmt.Sprintf(
			"%v -> %v",
			params.Source,
			params.Destination,
		)
	}
}

func validateParams(params *createCmdParams) error {
	return paramutils.ValidateFlagRepositoryParams(&params.Repository)
}
