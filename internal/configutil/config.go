package configutil

import (
	"log"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"

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
	viper.SetDefault("bitbucket.user", "")
	viper.SetDefault("bitbucket.token", "")

	// TODO: Create ~/.config/.prctl dir

	// // configPath := filepath.Join(hd, configName)
	// if err := viper.SafeWriteConfigAs(configPath); err != nil {
	// 	log.Fatal(err)
	// }

	loadConfig(hdCfgPath)
	loadConfig(".prctlcfg")
}
