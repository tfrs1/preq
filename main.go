package main

import (
	"fmt"
	"prctl/cmd"
	"prctl/internal/configutil"

	"github.com/spf13/viper"
)

func main() {
	configutil.Load()
	cmd.Execute()

	fmt.Println(viper.GetBool("test"))
	fmt.Println(viper.Get("bitbucket"))
	fmt.Println(viper.GetString("bitbucket.token"))
}
