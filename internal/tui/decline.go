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
	processFn func(cl client.Client, r *client.Repository, id string, globalId string, c chan utils.ProcessPullRequestResponse),
	fn func(utils.ProcessPullRequestResponse) string,
) {
	c := make(chan utils.ProcessPullRequestResponse)
	for _, v := range selectedPRs {
		go processFn(v.Client, v.Repository, v.ID, v.GlobalID, c)
	}

	end := len(selectedPRs)
	count := 0
	for {
		msg := <-c
		count++
		fn(msg)

		if count >= end {
			break
		}
	}
}

func execute(
	selectedPRs map[string]*promptPullRequest,
	fn func(utils.ProcessPullRequestResponse) string,
) error {
	processPullRequestMap(
		selectedPRs,
		declinePR,
		fn,
	)

	return nil
}

func declinePR(
	cl client.Client,
	r *client.Repository,
	id string,
	globalId string,
	c chan utils.ProcessPullRequestResponse,
) {
	_, err := cl.DeclinePullRequest(&client.DeclinePullRequestOptions{
		Repository: r,
		ID:         id,
	})

	res := utils.ProcessPullRequestResponse{
		ID:       id,
		GlobalID: globalId,
		Status:   "Done",
	}
	if err != nil {
		res.Status = "Error"
		res.Error = err
	}

	c <- res
}

func approvePR(
	cl client.Client,
	r *client.Repository,
	id string,
	globalId string,
	c chan utils.ProcessPullRequestResponse,
) {
	_, err := cl.Approve(&client.ApproveOptions{
		Repository: r,
		ID:         id,
	})

	res := utils.ProcessPullRequestResponse{
		ID:       id,
		GlobalID: globalId,
		Status:   "Done",
	}
	if err != nil {
		res.Status = "Error"
		res.Error = err
	}

	c <- res
}

func mergePR(
	cl client.Client,
	r *client.Repository,
	id string,
	globalId string,
	c chan utils.ProcessPullRequestResponse,
) {
	_, err := cl.Merge(&client.MergeOptions{
		Repository: r,
		ID:         id,
	})

	res := utils.ProcessPullRequestResponse{
		ID:       id,
		GlobalID: globalId,
		Status:   "Done",
	}
	if err != nil {
		res.Status = "Error"
		res.Error = err
	}

	c <- res
}

func unapprovePR(
	cl client.Client,
	r *client.Repository,
	id string,
	globalId string,
	c chan utils.ProcessPullRequestResponse,
) {
	_, err := cl.Unapprove(&client.UnapproveOptions{
		Repository: r,
		ID:         id,
	})

	res := utils.ProcessPullRequestResponse{
		ID:       id,
		GlobalID: globalId,
		Status:   "Done",
	}
	if err != nil {
		res.Status = "Error"
		res.Error = err
	}

	c <- res
}
