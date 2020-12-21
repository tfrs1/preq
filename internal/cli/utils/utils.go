package utils

import (
	"fmt"
	"os"

	"preq/internal/domain/pullrequest"
	"preq/internal/pkg/client"
	"preq/internal/systemcodes"

	"github.com/spf13/cobra"
)

type PromptPullRequest struct {
	ID    string
	Title string
}

func ProcessPullRequestMap(
	selectedPRs map[string]*PromptPullRequest,
	cl pullrequest.Repository,
	r *client.Repository,
	processFn func(cl pullrequest.Repository, r *client.Repository, id string, c chan interface{}),
	fn func(interface{}) string,
) {
	c := make(chan interface{})
	for id := range selectedPRs {
		go processFn(cl, r, id, c)
	}

	end := len(selectedPRs)
	count := 0
	for {
		msg := <-c
		count++
		fmt.Printf(fn(msg))

		if count >= end {
			break
		}
	}
}

func maxPRDescriptionLength(prs []*pullrequest.Entity, limit int) int {
	maxLen := 0
	for _, pr := range prs {
		l := len(pr.Source) + len(pr.Destination) + 4
		if l > maxLen {
			maxLen = l
		}
	}

	if limit > 0 && maxLen > limit {
		return limit
	}

	return maxLen
}

type runCommandError func(*cobra.Command, []string) error
type runCommandNoError func(*cobra.Command, []string)

func RunCommandWrapper(fn runCommandError) runCommandNoError {
	return func(cmd *cobra.Command, args []string) {
		err := fn(cmd, args)
		if err != nil {
			fmt.Println(err)

			switch err {
			// TODO: Add more specific exit codes
			default:
				os.Exit(systemcodes.ErrorCodeGeneric)
			}
		}
	}
}
