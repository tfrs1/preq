package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:     "prctl",
	Short:   "prctl command-line utility for pull requests",
	Long:    `Command-line utility for all your pull request needs.`,
	Version: fmt.Sprintf("%v, commit %v, built at %v", version, commit, date),
	Run:     func(cmd *cobra.Command, args []string) {},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
