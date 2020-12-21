package client

import (
	"preq/internal/domain/pullrequest"
)

type MockClient struct {
	ErrorValue error
}

func (c *MockClient) Get(o *pullrequest.GetOptions) (pullrequest.EntityPageList, error) {
	return nil, c.ErrorValue
}

func (c *MockClient) Create(o *pullrequest.CreateOptions) (*pullrequest.Entity, error) {
	return nil, c.ErrorValue
}

func (c *MockClient) Approve(o *pullrequest.ApproveOptions) (*pullrequest.Entity, error) {
	return nil, c.ErrorValue
}

func (c *MockClient) Decline(o *pullrequest.DeclineOptions) (*pullrequest.Entity, error) {
	return nil, c.ErrorValue
}
