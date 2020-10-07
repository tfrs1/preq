package client

type MockClient struct {
	ErrorValue error
}

func (c *MockClient) GetPullRequests(o *GetPullRequestsOptions) (*PullRequestList, error) {
	return nil, c.ErrorValue
}

func (c *MockClient) CreatePullRequest(o *CreatePullRequestOptions) (*PullRequest, error) {
	return nil, c.ErrorValue
}

func (c *MockClient) ApprovePullRequest(o *ApprovePullRequestOptions) (*PullRequest, error) {
	return nil, c.ErrorValue
}

func (c *MockClient) DeclinePullRequest(o *DeclinePullRequestOptions) (*PullRequest, error) {
	return nil, c.ErrorValue
}
