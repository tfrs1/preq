package clientutils

import (
	"errors"
	"preq/internal/config"
	"preq/internal/domain/pullrequest"
	"preq/internal/pkg/bitbucket"
	"preq/internal/pkg/client"
	"preq/internal/pkg/github"
)

type ClientFactory struct{}

func (cf ClientFactory) DefaultWithFlags(flags config.FlagSet) (pullrequest.Repository, error) {
	params := &config.RepositoryParams{}
	config.FillDefaultRepositoryParams(params)
	if flags != nil {
		config.FillFlagRepositoryParams(flags, params)
	}

	c, err := cf.DefaultPullRequestRepository1(params.Provider, params.Name)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (cf ClientFactory) DefaultPullRequestRepository(provider client.RepositoryProvider) (pullrequest.Repository, error) {
	switch provider {
	case client.RepositoryProviderEnum.BITBUCKET:
		return bitbucket.DefaultClient()
	case client.RepositoryProviderEnum.GITHUB:
		return github.DefaultClient()
	}

	return nil, errors.New("unknown provider")
}

func (cf ClientFactory) DefaultPullRequestRepository1(provider client.RepositoryProvider, repo string) (pullrequest.Repository, error) {
	switch provider {
	case client.RepositoryProviderEnum.BITBUCKET:
		return bitbucket.DefaultClient1(repo)
	case client.RepositoryProviderEnum.GITHUB:
		return github.DefaultClient1(repo)
	}

	return nil, errors.New("unknown provider")
}
