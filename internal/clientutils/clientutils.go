package clientutils

import (
	"errors"
	"fmt"
	"preq/internal/pkg/bitbucket"
	"preq/internal/pkg/client"
	"preq/internal/pkg/github"

	"github.com/spf13/viper"
)

type ClientFactory struct{}

func (cf ClientFactory) NewClient(provider client.RepositoryProvider, config *viper.Viper) (client.Client, error) {
	switch provider {
	case client.RepositoryProviderEnum.BITBUCKET:
		username := viper.GetString("bitbucket.username")
		if username == "" {
			return nil, fmt.Errorf("missing username")
		}
		password := viper.GetString("bitbucket.password")
		if password == "" {
			return nil, fmt.Errorf("missing password")
		}
		uuid := viper.GetString("bitbucket.uuid")
		repository := viper.GetString("default.repository")

		return bitbucket.NewClient(&bitbucket.ClientOptions{
			Username:   username,
			Password:   password,
			Uuid:       uuid,
			Repository: repository,
		})
	case client.RepositoryProviderEnum.GITHUB:
		return github.DefaultClient()
	}

	return nil, errors.New("unknown provider")
}

func (cf ClientFactory) DefaultClientCustom(provider client.RepositoryProvider, project string) (client.Client, error) {
	switch provider {
	case client.RepositoryProviderEnum.BITBUCKET:
		username := viper.GetString("bitbucket.username")
		if username == "" {
			return nil, fmt.Errorf("missing username")
		}
		password := viper.GetString("bitbucket.password")
		if password == "" {
			return nil, fmt.Errorf("missing password")
		}
		uuid := viper.GetString("bitbucket.uuid")
		repository := project

		return bitbucket.NewClient(&bitbucket.ClientOptions{
			Username:   username,
			Password:   password,
			Uuid:       uuid,
			Repository: repository,
		})
	case client.RepositoryProviderEnum.GITHUB:
		return github.DefaultClient()
	}

	return nil, errors.New("unknown provider")
}
