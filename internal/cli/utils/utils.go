package utils

import (
	"fmt"
	"os"

	"preq/internal/domain"
	"preq/internal/pkg/client"
	"preq/internal/systemcodes"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

type PromptPullRequest struct {
	ID    string
	Title string
}

func PromptPullRequestMultiSelect(prList *domain.PullRequestList) map[string]*PromptPullRequest {
	prs := getPromptPullRequestSilce(prList)

	answers := []string{}
	options := make([]string, 0, len(prs))
	for _, v := range prs {
		options = append(options, v.Title)
	}
	prompt := &survey.MultiSelect{
		Message:  "Decline pull requests",
		Options:  options,
		PageSize: 10,
	}
	survey.AskOne(prompt, &answers)

	selectedPRs := make(map[string]*PromptPullRequest, len(answers))
	for _, a := range answers {
		for _, v := range prs {
			if v.Title == a {
				selectedPRs[v.ID] = v
			}
		}
	}

	return selectedPRs
}

func ProcessPullRequestMap(
	selectedPRs map[string]*PromptPullRequest,
	cl domain.Client,
	r *client.Repository,
	processFn func(cl domain.Client, r *client.Repository, id string, c chan interface{}),
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

func maxPRDescriptionLength(prs []*domain.PullRequest, limit int) int {
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

func getPromptPullRequestSilce(prs *domain.PullRequestList) []*PromptPullRequest {
	maxLen := maxPRDescriptionLength(prs.Values, 30)
	prFormat := fmt.Sprintf("#%%s: %%-%ds %%s %%s", maxLen)
	options := make([]*PromptPullRequest, 0, len(prs.Values))
	for _, pr := range prs.Values {
		prDesc := fmt.Sprintf(
			prFormat,
			pr.ID,
			fmt.Sprintf("[%s->%s]", pr.Source, pr.Destination),
			pr.Updated.Format("(2006-01-02 15:04)"),
			pr.Title,
		)
		options = append(options, &PromptPullRequest{
			ID:    pr.ID,
			Title: prDesc,
		})
	}

	return options
}

func PromptPullRequestSelect(prList *domain.PullRequestList) *PromptPullRequest {
	prs := getPromptPullRequestSilce(prList)

	var answer string
	options := make([]string, 0, len(prs))
	for _, v := range prs {
		options = append(options, v.Title)
	}
	prompt := &survey.Select{
		Message:  "Open pull request page",
		Options:  options,
		PageSize: 10,
	}
	survey.AskOne(prompt, &answer)

	for _, v := range prs {
		if v.Title == answer {
			return v
		}
	}

	return nil
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
