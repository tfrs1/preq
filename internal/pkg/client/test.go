package client

import "preq/internal/domain"

type MockClient struct {
	ErrorValue error
}

func (c *MockClient) Get(o *domain.GetPullRequestOptions) (domain.PullRequestPageList, error) {
	return nil, c.ErrorValue
}

func (c *MockClient) Create(o *domain.CreatePullRequestOptions) (*domain.PullRequest, error) {
	return nil, c.ErrorValue
}

func (c *MockClient) Approve(o *domain.ApprovePullRequestOptions) (*domain.PullRequest, error) {
	return nil, c.ErrorValue
}

func (c *MockClient) Decline(o *domain.DeclinePullRequestOptions) (*domain.PullRequest, error) {
	return nil, c.ErrorValue
}
