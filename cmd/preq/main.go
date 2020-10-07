package main

import (
	"fmt"
	"os"
	"preq/internal/configutils"

	"preq/internal/cli"
)

func main() {
	err := configutils.Load()
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}

	cli.Execute()
}
