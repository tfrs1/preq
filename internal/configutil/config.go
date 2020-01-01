package configutil

import (
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	ErrHomeDirNotFound = errors.New("unable to determine the home directory")
)

func loadConfig(filename string) {
	if info, err := os.Stat(filename); err == nil && !info.IsDir() {
		f, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
		err = viper.MergeConfig(f)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func Load() {
	hdCfgPath, err := homedir.Expand("~/.config/prctl/config.toml")
	if err != nil {
		log.Fatal(ErrHomeDirNotFound)
	}

	viper.SetConfigType("toml")

	// TODO: Create ~/.config/.prctl dir

	// // configPath := filepath.Join(hd, configName)
	// if err := viper.SafeWriteConfigAs(configPath); err != nil {
	// 	log.Fatal(err)
	// }

	loadConfig(hdCfgPath)
	loadConfig(".prctlcfg")
}

func GetStringFlagOrDie(cmd *cobra.Command, flag string, err error) string {
	s, cmdErr := cmd.Flags().GetString(flag)
	if cmdErr != nil || s == "" {
		e := err
		if cmdErr != nil {
			e = errors.Wrap(cmdErr, err.Error())
		}
		fmt.Println(e)
		os.Exit(3)
	}

	return s
}

func GetStringOrDie(s string, err error) string {
	if s == "" {
		fmt.Println(err)
		os.Exit(3)
	}

	return s
}

type FlagSet interface {
	GetString(string) (string, error)
}

func GetStringFlagOrDefault(fs FlagSet, flag, d string) string {
	s, err := fs.GetString(flag)
	if err != nil || s == "" {
		return d
	}

	return s
}
