package cmdcreate

import (
	"preq/internal/cli/paramutils"
	"preq/internal/config"
	"preq/internal/errcodes"
	"preq/internal/gitutils"
)

type createCmdParams struct {
	Repository  config.RepositoryParams
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

func populateParams(params *createCmdParams, flags paramutils.FlagSet) {
	config.FillDefaultRepositoryParams(&params.Repository)
	fillDefaultParams(params)
	fillFlagParams(flags, params)
}

func fillFlagParams(flags paramutils.FlagSet, params *createCmdParams) {
	config.FillFlagRepositoryParams(flags, &params.Repository)

	var (
		source      = flags.GetStringOrDefault("source", params.Source)
		destination = flags.GetStringOrDefault("destination", params.Destination)
		title       = flags.GetStringOrDefault("title", params.Title)
		close       = flags.GetBoolOrDefault("close", params.CloseBranch)
		draft       = flags.GetBoolOrDefault("draft", params.Draft)
	)

	params.Title = title
	params.Source = source
	params.Destination = destination
	params.CloseBranch = close
	params.Draft = draft
}

func fillDefaultParams(p *createCmdParams) {
	defaultSourceBranch, err := gitutils.GetCurrentBranch()
	if err == nil {
		p.Source = defaultSourceBranch
	}

	// TODO: Make closest branch list configurable
	destination, err := gitutils.GetClosestBranch([]string{"develop", "master"})
	if err == nil {
		p.Destination = destination
	}

	title, err := gitutils.GetCurrentCommitMessage()
	if err == nil {
		p.Title = title
	}
}

func validateParams(params *createCmdParams) error {
	return paramutils.ValidateFlagRepositoryParams(&params.Repository)
}
