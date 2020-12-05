package cmdcreate

import (
	"fmt"
	"preq/internal/cli/paramutils"
	"preq/internal/config"
	"preq/internal/errcodes"
	"preq/internal/gitutils"
	"strings"

	"github.com/AlecAivazis/survey/v2"
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

func fillFlagParams(flags paramutils.FlagSet, params *createCmdParams) {
	paramutils.FillFlagRepositoryParams(flags, &params.Repository)

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
	config.FillDefaultRepositoryParams(&p.Repository)

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

func fillInteractiveParams(params *createCmdParams) error {
	// NOTE: Just hitting enter to select the first option
	// does not work when the default value is an empty string
	var defaultProvider interface{}
	if params.Repository.Provider != "" {
		defaultProvider = params.Repository.Provider
	} else {
		defaultProvider = nil
	}

	// the questions to ask
	var qs = []*survey.Question{
		{
			Name: "provider",
			Prompt: &survey.Select{
				Message: "Provider:",
				Options: []string{"bitbucket"},
				Default: defaultProvider,
			},
			Validate: survey.Required,
		},
		{
			Name: "repository",
			Prompt: &survey.Input{
				Message: "Repository",
				Default: params.Repository.Name,
			},
			Validate: func(val interface{}) error {
				err := survey.Required(val)
				if err != nil {
					return err
				}

				value := fmt.Sprintf("%v", val)

				v := strings.Split(value, "/")
				if len(v) != 2 || v[0] == "" || v[1] == "" {
					return errcodes.ErrRepositoryMustBeInFormOwnerRepo
				}

				return nil
			},
		},
		{
			Name: "source",
			Prompt: &survey.Input{
				Message: "Source branch",
				Default: params.Source,
			},
			Validate: survey.Required,
		},
		{
			Name: "destination",
			Prompt: &survey.Input{
				Message: "Destination branch",
				Default: params.Destination,
			},
			Validate: survey.Required,
		},
		{
			Name: "title",
			Prompt: &survey.Input{
				Message: "Title",
				Default: params.Title,
			},
			Validate: survey.Required,
		},
	}

	err := survey.Ask(qs, params)
	if err != nil {
		return err
	}

	return nil
}

func validateParams(params *createCmdParams) error {
	return paramutils.ValidateFlagRepositoryParams(&params.Repository)
}
