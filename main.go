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

	err = cmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}
}
