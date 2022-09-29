package tui

import (
	"preq/internal/pkg/client"
)

type promptPullRequest struct {
	ID    string
	Title string
}

func processPullRequestMap(
	selectedPRs map[string]*promptPullRequest,
	cl client.Client,
	r *client.Repository,
	processFn func(cl client.Client, r *client.Repository, id string, c chan interface{}),
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
		fn(msg)

		if count >= end {
			break
		}
	}
}

func execute(c client.Client, repo *client.Repository, selectedPRs map[string]*promptPullRequest,
	fn func(interface{}) string,
) error {
	processPullRequestMap(
		selectedPRs,
		c,
		repo,
		declinePR,
		fn,
	)

	return nil
}

type declineResponse struct {
	ID     string
	Status string
	Error  error
}

func declinePR(cl client.Client, r *client.Repository, id string, c chan interface{}) {
	_, err := cl.DeclinePullRequest(&client.DeclinePullRequestOptions{
		Repository: r,
		ID:         id,
	})

	res := declineResponse{ID: id, Status: "Done"}
	if err != nil {
		res.Status = "Error"
		res.Error = err
	}

	c <- res
}
