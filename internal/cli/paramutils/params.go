package paramutils

import (
	"preq/internal/errcodes"
	"preq/internal/gitutils"
	"preq/internal/persistance"
	"preq/internal/pkg/client"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type RepositoryParams struct {
	Provider client.RepositoryProvider
	Name     string
}

func (p *RepositoryParams) SetDefault() {
	p.Name = "owner/repo-name"
}

type FlagSet interface {
	GetStringOrDefault(flag, d string) string
	GetBoolOrDefault(flag string, d bool) bool
}

type PFlagSetWrapper struct {
	Flags *pflag.FlagSet
}

func (fs *PFlagSetWrapper) GetStringOrDefault(flag, d string) string {
	s, err := fs.Flags.GetString(flag)
	if err != nil || s == "" {
		return d
	}

	return s
}

func (fs *PFlagSetWrapper) GetBoolOrDefault(flag string, d bool) bool {
	s, err := fs.Flags.GetBool(flag)
	if err != nil {
		return d
	}

	return s
}

type paramsFiller interface {
	Fill(params *RepositoryParams)
}

type localRepositoryParamsFiller struct{}

func (pf *localRepositoryParamsFiller) Fill(params *RepositoryParams) {
	git, err := gitutils.GetWorkingDirectoryRepo()
	if err != nil {
		return
	}

	defaultRepo, err := git.GetRemoteInfo()
	if err != nil {
		return
	}

	params.Name = defaultRepo.Name
	params.Provider = defaultRepo.Provider
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

// type flagsParamsFiller struct {}
// func (pf *flagsParamsFiller) Fill(params *RepositoryParams) {
// 	var (
// 		repo     = flags.GetStringOrDefault("repository", params.Name)
// 		provider = flags.GetStringOrDefault("provider", string(params.Provider))
// 	)

// 	params.Name = repo
// 	params.Provider = client.RepositoryProvider(provider)
// }

func GetRepoAndFillRepoParams(
	flags FlagSet,
	repoParams *RepositoryParams,
) (gitutils.GitUtilsClient, error) {
	repoName := flags.GetStringOrDefault("repository", "")
	providerName := flags.GetStringOrDefault("provider", "")

	if repoName != "" && providerName != "" {
		info, err := persistance.GetDefault().GetInfo(repoName, providerName)
		if err != nil {
			return nil, err
		}

		repoParams.Provider = client.RepositoryProvider(providerName)
		repoParams.Name = repoName

		return gitutils.GetRepo(info.Path)
	}

	git, err := gitutils.GetWorkingDirectoryRepo()
	if err == nil {
		info, err := git.GetRemoteInfo()
		if err == nil {
			repoParams.Provider = info.Provider
			repoParams.Name = info.Name

			return git, nil
		}
	}

	// TODO: Read the default repo from global config?
	// config, err := configutils.DefaultConfig()
	// if err != nil {
	// 	log.Panic().Msg(err.Error())
	// }
	// r := config.GetString("default.repository")
	// p := config.GetString("default.provider")

	return nil, err
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

func FillFlagRepositoryParams(flags FlagSet, params *RepositoryParams) {
	var (
		repo     = flags.GetStringOrDefault("repository", params.Name)
		provider = flags.GetStringOrDefault("provider", string(params.Provider))
	)

	params.Name = repo
	params.Provider = client.RepositoryProvider(provider)
}

func ValidateFlagRepositoryParams(params *RepositoryParams) error {
	if params.Name != "" && params.Provider != "" {
		v := strings.Split(params.Name, "/")
		if len(v) != 2 || v[0] == "" || v[1] == "" {
			return errcodes.ErrRepositoryMustBeInFormOwnerRepo
		}

		if !params.Provider.IsValid() {
			return errcodes.ErrorRepositoryProviderUnknown
		}
	}

	return nil
}

func ParseIDArg(args []string) string {
	id := ""
	if len(args) > 0 {
		id = args[0]
	}

	return id
}
