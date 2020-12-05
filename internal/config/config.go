package config

import (
	"fmt"
	"preq/internal/clientutils"
	"preq/internal/domain/pullrequest"
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

type RepositoryInfo interface {
	GetCurrentBranch() string
	GetClosestBranch([]string) (string, error)
	GetCurrentCommitMessage() (string, error)
}

func FillDefaultRepositoryParams(params *RepositoryParams) {
	paramsFillers := []paramsFiller{
		&viperConfigParamsFiller{},
		&localRepositoryParamsFiller{},
	}

	for _, pf := range paramsFillers {
		pf.Fill(params)
	}
}

type FlagSet interface {
	GetStringOrDefault(flag, d string) string
	GetBoolOrDefault(flag string, d bool) bool
}

func FillFlagRepositoryParams(flags FlagSet, params *RepositoryParams) {
	var (
		repo     = flags.GetStringOrDefault("repository", params.Name)
		provider = flags.GetStringOrDefault("provider", string(params.Provider))
	)

	params.Name = repo
	params.Provider = client.RepositoryProvider(provider)
}

func LoadLocal(flags FlagSet) (pullrequest.Repository, *client.Repository, error) {
	// TODO: Rename and move somewhere appropriate, refactor
	params := &RepositoryParams{}
	FillDefaultRepositoryParams(params)
	if flags != nil {
		FillFlagRepositoryParams(flags, params)
	}

	r, err := client.NewRepositoryFromOptions(&client.RepositoryOptions{
		Provider: params.Provider,
		// TODO: use version control repository in git repo name?
		FullRepositoryName: params.Name,
	})
	if err != nil {
		return nil, nil, err
	}

	c, err := clientutils.ClientFactory{}.DefaultPullRequestRepository1(params.Provider, r)
	if err != nil {
		return nil, nil, err
	}

	return c, r, nil
}
