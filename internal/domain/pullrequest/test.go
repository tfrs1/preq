package pullrequest

type MockPullRequestCreator struct{}

func (m *MockPullRequestCreator) Create(o *CreateOptions) (*Entity, error) {
	return &Entity{}, nil
}

type MockPullRequestUpdater struct{}

func (m *MockPullRequestUpdater) Update(o *UpdateOptions) (*Entity, error) {
	return &Entity{}, nil
}
