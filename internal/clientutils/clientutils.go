package clientutils

import (
	"errors"
	"preq/internal/domain/pullrequest"
	"preq/internal/pkg/bitbucket"
	"preq/internal/pkg/client"
	"preq/internal/pkg/github"
)

type ClientFactory struct{}

func (cf ClientFactory) DefaultPullRequestRepository(provider client.RepositoryProvider) (pullrequest.Repository, error) {
	switch provider {
	case client.RepositoryProviderEnum.BITBUCKET:
		return bitbucket.DefaultClient()
	case client.RepositoryProviderEnum.GITHUB:
		return github.DefaultClient()
	}

	return nil, errors.New("unknown provider")
}

func (cf ClientFactory) DefaultPullRequestRepository1(provider client.RepositoryProvider, repo *client.Repository) (pullrequest.Repository, error) {
	switch provider {
	case client.RepositoryProviderEnum.BITBUCKET:
		return bitbucket.DefaultClient()
	case client.RepositoryProviderEnum.GITHUB:
		return github.DefaultClient1(repo)
	}

	return nil, errors.New("unknown provider")
}
