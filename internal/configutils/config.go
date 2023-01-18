package configutils

import (
	"io"
	"path/filepath"
	"preq/internal/pkg/fs"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"

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
	if err == nil {
		err = mergeConfig(f, v)
		if err != nil {
			return err
		}
	}

	return nil
}

var getGlobalConfigPath = func() (string, error) {
	return homedir.Expand("~/.config/preq/config.toml")
}

func MergeLocalConfig(v *viper.Viper, path string) error {
	// TODO: Extract .preqcfg file name to global
	return loadConfig(filepath.Join(path, ".preqcfg"), v)
}

func DefaultConfig() (*viper.Viper, error) {
	hdCfgPath, err := getGlobalConfigPath()
	if err != nil {
		return nil, ErrHomeDirNotFound
	}

	v := viper.New()
	v.SetConfigType("toml")
	err = loadConfig(hdCfgPath, v)

	return v, err
}

func Load() error {
	hdCfgPath, err := getGlobalConfigPath()
	if err != nil {
		return ErrHomeDirNotFound
	}

	configs := []string{hdCfgPath, ".preqcfg"}
	for _, v := range configs {
		err = loadConfig(v, viper.GetViper())
		if err != nil {
			return err
		}
	}

	return nil
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

func init() {
	viper.SetConfigType("toml")
}
