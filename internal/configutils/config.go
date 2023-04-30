package configutils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"preq/internal/pkg/fs"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type FlagSet interface {
	GetString(string) (string, error)
	GetBool(string) (bool, error)
}

type configMerger interface {
	MergeConfig(io.Reader) error
}

var (
	ErrHomeDirNotFound = errors.New("unable to determine the home directory")
	ErrConfigFileIsDir = errors.New("configuration file is a directory")
)

var mergeConfig = func(in io.Reader, cm configMerger) error {
	err := cm.MergeConfig(in)
	if err != nil {
		return err
	}

	return nil
}

var fileExists = func(filename string, fs fs.Filesystem) error {
	info, err := fs.Stat(filename)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return ErrConfigFileIsDir
	}

	return nil
}

var loadFile = func(filename string, fs fs.Filesystem) (io.Reader, error) {
	err := fileExists(filename, fs)
	if err != nil {
		return nil, err
	}

	f, err := fs.Open(filename)
	if err != nil {
		return nil, err
	}

	return f, nil
}

var loadConfig = func(filename string, v *viper.Viper) error {
	f, err := loadFile(filename, fs.OS{})
	if err != nil {
		return err
	}

	return mergeConfig(f, v)
}

var getGlobalConfigPath = func() (string, error) {
	return homedir.Expand("~/.config/preq/config.toml")
}

func MergeLocalConfig(v *viper.Viper, path string) error {
	// TODO: Extract .preqcfg file name to global
	f := filepath.Join(path, ".preqcfg")
	if _, err := os.Stat(f); errors.Is(err, os.ErrNotExist) {
		return nil
	}

	// Try for every supported file type
	filetypes := []string{"yaml", "json", "toml"}
	var err error = nil
	for _, ft := range filetypes {
		v.SetConfigType(ft)
		err := loadConfig(f, v)
		if err == nil {
			return nil
		}
	}

	return err
}

func LoadConfigForPath(path string) (*viper.Viper, error) {
	v, err := DefaultConfig()
	if err != nil {
		return nil, err
	}
	if v == nil {
		v = viper.New()
	}

	err = MergeLocalConfig(v, path)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func DefaultConfig() (*viper.Viper, error) {
	cfgDir, err := homedir.Expand("~/.config/preq")
	if err != nil {
		return nil, ErrHomeDirNotFound
	}

	v := viper.New()
	filetypes := []string{"yaml", "json", "toml"}
	for _, ft := range filetypes {
		f := filepath.Join(cfgDir, fmt.Sprintf("config.%s", ft))
		v.SetConfigType(ft)
		err = loadConfig(f, v)
		if err == nil {
			return v, nil
		}
		log.Debug().
			Msgf("config loading failed for type %s, skipping to next filetype", ft)
	}

	return nil, errors.Wrap(err, "could not load config")
}

func GetBoolFlagOrDefault(fs FlagSet, flag string, d bool) bool {
	v, err := fs.GetBool(flag)
	if err != nil {
		return d
	}

	return v
}

func GetStringFlagOrDefault(fs FlagSet, flag, d string) string {
	s, err := fs.GetString(flag)
	if err != nil || s == "" {
		return d
	}

	return s
}
