package tui

import (
	"preq/internal/cli/utils"
	"preq/internal/pkg/client"
)

type promptPullRequest struct {
	ID         string
	GlobalID   string
	Title      string
	Client     client.Client
	Repository *client.Repository
}

func processPullRequestMap(
	selectedPRs map[string]*promptPullRequest,
	processor func(
		cl client.Client,
		r *client.Repository,
		id string, globalId string,
		ch chan *utils.ProcessPullRequestResponse,
	),
	callback func(*utils.ProcessPullRequestResponse),
) {
	ch := make(chan *utils.ProcessPullRequestResponse)
	defer close(ch)

	for _, v := range selectedPRs {
		go processor(v.Client, v.Repository, v.ID, v.GlobalID, ch)
	}

	for range selectedPRs {
		msg := <-ch
		callback(msg)
	}
}
