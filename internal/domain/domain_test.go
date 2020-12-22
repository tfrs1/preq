package domain

import (
	"preq/internal/domain/pullrequest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockPresenter struct {
	PresentCalls int
	NotifyCalls  int
}

func NewMockPresenter() *MockPresenter {
	return &MockPresenter{0, 0}
}

type MockClient struct {
	GetPullRequestsCalls int
}

func (mc *MockClient) Get(*pullrequest.GetOptions) (pullrequest.EntityPageList, error) {
	mc.GetPullRequestsCalls++
	return nil, nil
}

func (mc *MockClient) Create(o *pullrequest.CreateOptions) (*pullrequest.Entity, error) {
	return nil, nil
}

func (mc *MockClient) Approve(o *pullrequest.ApproveOptions) (*pullrequest.Entity, error) {
	return nil, nil
}

func (mc *MockClient) Decline(o *pullrequest.DeclineOptions) (*pullrequest.Entity, error) {
	return nil, nil
}

type MockPullRequestUpdateListener struct {
	UpdateCalls int
}

func (mprul *MockPullRequestUpdateListener) Update(prList pullrequest.EntityPageList) {
	mprul.UpdateCalls++
}

func (mprul *MockPullRequestUpdateListener) UpdateFailed(e error) {
	// mprul.UpdateCalls++
}

func (mp *MockPresenter) Start() {
	mp.PresentCalls++
}

func (mp *MockPresenter) Notify(e *Event) {
	mp.NotifyCalls++
}

func Test_NewDomain(t *testing.T) {
	t.Run("Can instantiate domain", func(t *testing.T) {
		// presenter := NewMockPresenter()
		domain := NewDomain()

		assert.NotEqual(t, domain, nil)
	})

	t.Run("Can start presentation", func(t *testing.T) {
		presenter := NewMockPresenter()
		domain := &Domain{Presenter: presenter}

		domain.Present()

		assert.Equal(t, presenter.PresentCalls, 1)
		assert.NotEqual(t, domain, nil)
	})
}

func Test_LoadPullRequests(t *testing.T) {
	t.Run("fails", func(t *testing.T) {
		mprul := &MockPullRequestUpdateListener{}
		mc := &MockClient{}
		LoadPullRequests(mc, mprul)
		assert.Equal(t, mprul.UpdateCalls, 1)
		assert.Equal(t, mc.GetPullRequestsCalls, 1)
	})
}
