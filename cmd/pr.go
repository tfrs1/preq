package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	prCmd.PersistentFlags().StringP("dest", "d", "The closes branch in history", "destination branch of your pull request")
	rootCmd.AddCommand(prCmd)
}

var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "Create pull request",
	Long:  `Creates a pull request on the web service hosting your origin respository`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hugo Static Site Generator v0.9 -- HEAD")
	},
}
