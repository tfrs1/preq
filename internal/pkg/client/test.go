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

func (c *MockClient) Close(o *pullrequest.CloseOptions) (*pullrequest.Entity, error) {
	return nil, c.ErrorValue
}

// TODO: There seems a mock client already in domain.go
func (mc *MockClient) WebPage(e pullrequest.EntityID) string {
	return ""
}

func (mc *MockClient) WebPageList() string {
	return ""
}
