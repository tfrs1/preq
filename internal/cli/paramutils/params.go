package paramutils

import (
	"preq/internal/config"
	"preq/internal/errcodes"
	"strings"

	"github.com/spf13/pflag"
)

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

// type flagsParamsFiller struct {}
// func (pf *flagsParamsFiller) Fill(params *RepositoryParams) {
// 	var (
// 		repo     = flags.GetStringOrDefault("repository", params.Name)
// 		provider = flags.GetStringOrDefault("provider", string(params.Provider))
// 	)

// 	params.Name = repo
// 	params.Provider = client.RepositoryProvider(provider)
// }

func ValidateFlagRepositoryParams(params *config.RepositoryParams) error {
	if (params.Name == "" && params.Provider != "") || (params.Name != "" && params.Provider == "") {
		return errcodes.ErrSomeRepoParamsMissing
	}

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
