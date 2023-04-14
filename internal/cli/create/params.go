package cmdcreate

import (
	"fmt"
	"preq/internal/cli/paramutils"
	"preq/internal/errcodes"
	"preq/internal/gitutils"
	"strings"

	"github.com/AlecAivazis/survey/v2"
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
	qs := []*survey.Question{
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
