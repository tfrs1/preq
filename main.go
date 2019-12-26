package main

import (
	"prctl/cmd"
	"prctl/internal/configutil"
)

func main() {
	configutil.Load()
	cmd.Execute()
}
