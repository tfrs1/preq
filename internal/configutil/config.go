package configutil

import (
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	ErrHomeDirNotFound = errors.New("unable to determine the home direcotry")
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

func GetStringFlagOrDie(cmd *cobra.Command, flag string) string {
	s := GetStringFlag(cmd, flag)
	if s == "" {
		log.Fatal("string empty")
	}

	return s
}

func GetStringFlagOrDefault(cmd *cobra.Command, flag, d string) string {
	s := GetStringFlag(cmd, flag)
	if s == "" {
		return d
	}

	return s
}

func GetStringFlag(cmd *cobra.Command, flag string) string {
	return cmd.Flags().Lookup(flag).Value.String()
}
