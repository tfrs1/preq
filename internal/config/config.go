package config

import (
	"fmt"
	"preq/internal/clientutils"
	"preq/internal/domain"
	"preq/internal/gitutils"
	"preq/internal/pkg/client"

	"github.com/spf13/viper"
)

const (
	DEFAULT_CONFIG_PATH = "~/.config/preq/config.toml"
)

type paramsFiller interface {
	Fill(params *RepositoryParams)
}

type RepositoryParams struct {
	Provider client.RepositoryProvider
	Name     string
}

func (p *RepositoryParams) SetDefault() {
	p.Name = "owner/repo-name"
}

type localRepositoryParamsFiller struct{}

func (pf *localRepositoryParamsFiller) Fill(params *RepositoryParams) {
	defaultRepo, err := gitutils.GetRemoteInfo()
	if err == nil {
		params.Name = fmt.Sprintf("%s/%s", defaultRepo.Owner, defaultRepo.Name)
		params.Provider = defaultRepo.Provider
	}
}

type viperConfigParamsFiller struct{}

func (pf *viperConfigParamsFiller) Fill(params *RepositoryParams) {
	if dp := viper.GetString("default.provider"); dp != "" {
		provider, err := client.ParseRepositoryProvider(dp)
		if err == nil {
			params.Provider = provider
		}
	}

	if dr := viper.GetString("default.repository"); dr != "" {
		params.Name = dr
	}
}

func FillDefaultRepositoryParams(params *RepositoryParams) {
	paramsFillers := []paramsFiller{
		&localRepositoryParamsFiller{},
		&viperConfigParamsFiller{},
	}

	for _, pf := range paramsFillers {
		pf.Fill(params)
	}
}

func LoadLocal() (domain.PullRequestRepository, *client.Repository, error) {
	params := &RepositoryParams{}
	FillDefaultRepositoryParams(params)

	r, err := client.NewRepositoryFromOptions(&client.RepositoryOptions{
		Provider:           params.Provider,
		FullRepositoryName: params.Name,
	})
	if err != nil {
		return nil, nil, err
	}

	c, err := clientutils.ClientFactory{}.DefaultClient(params.Provider)
	if err != nil {
		return nil, nil, err
	}

	return c, r, nil
}
