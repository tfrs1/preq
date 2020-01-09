package cmd

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

func fillInteractiveParams(params *createCmdParams) error {
	// NOTE: Just hitting enter to select the first option
	// does not work when the default value is an empty string
	var defaultProvider interface{}
	if params.Provider != "" {
		defaultProvider = params.Provider
	} else {
		defaultProvider = nil
	}

	// the questions to ask
	var qs = []*survey.Question{
		{
			Name: "provider",
			Prompt: &survey.Select{
				Message: "Provider:",
				Options: []string{"bitbucket-cloud"},
				Default: defaultProvider,
			},
			Validate: survey.Required,
		},
		{
			Name: "repository",
			Prompt: &survey.Input{
				Message: "Repository",
				Default: params.Repository,
			},
			Validate: func(val interface{}) error {
				err := survey.Required(val)
				if err != nil {
					return err
				}

				value := fmt.Sprintf("%v", val)

				v := strings.Split(value, "/")
				if len(v) != 2 || v[0] == "" || v[1] == "" {
					return ErrRepositoryMustBeInFormOwnerRepo
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
