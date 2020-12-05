package client

import "preq/internal/domain"

type MockClient struct {
	ErrorValue error
}

func (c *MockClient) GetPullRequests(o *domain.GetPullRequestOptions) (*domain.PullRequestList, error) {
	return nil, c.ErrorValue
}

func (c *MockClient) CreatePullRequest(o *domain.CreatePullRequestOptions) (*domain.PullRequest, error) {
	return nil, c.ErrorValue
}

func (c *MockClient) ApprovePullRequest(o *domain.ApprovePullRequestOptions) (*domain.PullRequest, error) {
	return nil, c.ErrorValue
}

func (c *MockClient) DeclinePullRequest(o *domain.DeclinePullRequestOptions) (*domain.PullRequest, error) {
	return nil, c.ErrorValue
}
