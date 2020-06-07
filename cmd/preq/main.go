package main

import (
	"fmt"
	"os"
	"preq/internal/configutils"

	"preq/internal/command"
)

func main() {
	err := configutils.Load()
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}

	command.Execute()
}
