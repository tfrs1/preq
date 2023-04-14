package paramutils

import (
	"os"
	"preq/internal/clientutils"
	"preq/internal/configutils"
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

type FlagRepo interface {
	GetStringOrDefault(flag, d string) string
	GetBoolOrDefault(flag string, d bool) bool
}

func NewFlagRepo(flags *pflag.FlagSet) FlagRepo {
	return &PFlagSetWrapper{Flags: flags}
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

func GetRepoPath(flagSet *pflag.FlagSet) (string, error) {
	flags := NewFlagRepo(flagSet)
	name := flags.GetStringOrDefault("repository", "")
	provider := flags.GetStringOrDefault("provider", "")
	isExplicitRepo := name != "" && provider != ""
	path := ""

	if isExplicitRepo {
		info, err := persistance.GetDefault().GetInfo(name, provider)
		if err != nil {
			return "", err
		}
		path = info.Path
	} else {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}

		path, err = gitutils.GetRepoRootDir(wd)
		if err != nil {
			return "", err
		}
	}

	return path, nil
}

func GetRepoUtilsAndParams(flagSet *pflag.FlagSet) (gitutils.GitUtilsClient, *RepositoryParams, error) {
	path, err := GetRepoPath(flagSet)
	params := &RepositoryParams{}

	git, err := gitutils.GetRepo(path)
	if err != nil {
		return nil, nil, err
	}

	info, err := git.GetRemoteInfo()
	if err != nil {
		return nil, nil, err
	}

	params.Provider = info.Provider
	params.Name = info.Name

	return git, params, nil
}

func GetClientAndRepoParams(flagSet *pflag.FlagSet) (client.Client, *RepositoryParams, error) {
	path, err := GetRepoPath(flagSet)
	params := &RepositoryParams{}

	git, err := gitutils.GetRepo(path)
	if err != nil {
		return nil, nil, err
	}

	info, err := git.GetRemoteInfo()
	if err != nil {
		return nil, nil, err
	}

	params.Provider = info.Provider
	params.Name = info.Name

	config, err := configutils.LoadConfigForPath(path)
	if err != nil {
		return nil, nil, err
	}

	cl, err := clientutils.ClientFactory{}.NewClient(
		params.Provider,
		config,
	)
	if err != nil {
		return nil, nil, err
	}

	return cl, params, nil
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

func FillFlagRepositoryParams(flags FlagRepo, params *RepositoryParams) {
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
