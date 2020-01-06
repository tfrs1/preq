package main

import (
	"fmt"
	"os"
	"prctl/cmd"
	"prctl/internal/configutil"
)

func main() {
	err := configutil.Load()
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}

	// TODO: Execute doesn't really have a return value when using Run command config option?
	err = cmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}
}
