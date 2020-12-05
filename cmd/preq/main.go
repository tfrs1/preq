package main

import (
	"fmt"
	"os"
	"preq/internal/config"
	"preq/internal/configutils"

	"preq/internal/cli"

	"github.com/spf13/viper"
)

func main() {
	err := configutils.Load()
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}

	c, r, err := config.Load()
	if err != nil {
		// TODO: Do something
	}
	fmt.Println(c, r)

	t := viper.GetString("github.username")
	fmt.Println(t)
	cli.Execute()
}
