package main

import (
	"fmt"
	"os"
	"preq/cmd"
	"preq/internal/configutils"
)

func main() {
	err := configutils.Load()
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}

	cmd.Execute()
}
