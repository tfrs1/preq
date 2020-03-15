package mocks

import client "preq/pkg/bitbucket"

type Client struct {
	ErrorValue error
}

func (c *Client) GetPullRequests(o *client.GetPullRequestsOptions) (*client.PullRequestList, error) {
	return nil, c.ErrorValue
}

func (c *Client) CreatePullRequest(o *client.CreatePullRequestOptions) (*client.PullRequest, error) {
	return nil, c.ErrorValue
}

func (c *Client) ApprovePullRequest(o *client.ApprovePullRequestOptions) (*client.PullRequest, error) {
	return nil, c.ErrorValue
}

func (c *Client) DeclinePullRequest(o *client.DeclinePullRequestOptions) (*client.PullRequest, error) {
	return nil, c.ErrorValue
}
