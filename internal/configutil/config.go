package configutil

import (
	"fmt"
	"io"
	"os"
	"prctl/internal/fs"

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

var loadConfig = func(filename string) {
	f, err := loadFile(filename, fs.OS{})
	if err == nil {
		mergeConfig(f, viper.GetViper())
	}
}

var getGlobalConfigPath = func() (string, error) {
	return homedir.Expand("~/.config/prctl/config.toml")
}

func Load() error {
	hdCfgPath, err := getGlobalConfigPath()
	if err != nil {
		return ErrHomeDirNotFound
	}

	// TODO: Create ~/.config/.prctl dir

	// // configPath := filepath.Join(hd, configName)
	// if err := viper.SafeWriteConfigAs(configPath); err != nil {
	// 	log.Fatal(err)
	// }

	configs := []string{hdCfgPath, ".prctlcfg"}
	for _, v := range configs {
		loadConfig(v)
	}

	return nil
}

func GetStringOrDie(s string, err error) string {
	if s == "" {
		fmt.Println(err)
		os.Exit(3)
	}

	return s
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
