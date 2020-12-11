package domain

import (
	"time"
)

type GitRepository struct {
	Name string
}

type Storage interface {
	RefreshPullRequestData(PullRequestRepository)
	Get() string
}

type Domain struct {
	// Repository *client.Repository
	Client    PullRequestRepository
	Storage   Storage
	Presenter Presenter
}

type Presenter interface {
	Start()
	Notify(*Event)
}

type Command interface {
	execute()
}

func NewDomain() *Domain {
	return &Domain{}
}

type CommandBus struct {
	domain *Domain
}

type LoadPullRequestsCommand struct{}

func (c *LoadPullRequestsCommand) execute() {

}

func (cb *CommandBus) execute(c Command) {
	switch c.(type) {
	case *LoadPullRequestsCommand:
		cb.domain.LoadPullRequests()
	}
}

type PullRequestRepository interface {
	Get(*GetPullRequestOptions) (PullRequestPageList, error)
	Create(*CreatePullRequestOptions) (*PullRequest, error)
	Approve(*ApprovePullRequestOptions) (*PullRequest, error)
	Decline(*DeclinePullRequestOptions) (*PullRequest, error)
}

type Event struct {
	eventType string
	data      string
}

func (d *Domain) notify(e *Event) {
	if d.Presenter != nil {
		d.Presenter.Notify(e)
	}
}

var (
	EVENT_PULL_REQUEST_LIST_UPDATED = "domain/EVENT_PULL_REQUEST_LIST_UPDATED"
)

func (d *Domain) LoadPullRequests() {
	d.Storage.RefreshPullRequestData(d.Client)
	d.notify(&Event{
		eventType: EVENT_PULL_REQUEST_LIST_UPDATED,
		data:      d.Storage.Get(),
	})
}

func (d *Domain) Present() {
	if d.Presenter != nil {
		d.Presenter.Start()
	}
}

type PullRequestUpdateListener interface {
	UpdateFailed(error)
	Update(PullRequestPageList)
}

func LoadPullRequests(c PullRequestRepository, l PullRequestUpdateListener) {
	prList, err := c.Get(&GetPullRequestOptions{})
	if err != nil {
		l.UpdateFailed(err)
	}

	l.Update(prList)
}

type PullRequestState string

type PullRequest struct {
	ID          string
	Title       string
	URL         string
	State       PullRequestState
	Source      string
	Destination string
	Created     time.Time
	Updated     time.Time
}

type PullRequestPageList interface {
	GetPage(int) ([]*PullRequest, error)
}

type PullRequestList struct {
	PageLength uint           `json:"pagelen"`
	Page       uint           `json:"page"`
	Size       uint           `json:"size"`
	NextURL    string         `json:"next"`
	Values     []*PullRequest `json:"values"`
}

type GetPullRequestOptions struct {
	State PullRequestState
	Next  string
}

type CreatePullRequestOptions struct {
	Title       string
	Source      string
	Destination string
	CloseBranch bool
	Draft       bool
	// ID string
}

type ApprovePullRequestOptions struct {
	ID string
}

type DeclinePullRequestOptions struct {
	ID string
}
