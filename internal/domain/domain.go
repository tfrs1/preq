package domain

import "preq/internal/domain/pullrequest"

// TODO: VCRepository?
type GitRepository struct {
	Name string
}

type Storage interface {
	RefreshPullRequestData(pullrequest.Repository)
	Get() string
}

type Domain struct {
	// Repository *client.Repository
	Client    pullrequest.Repository
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
	Update(pullrequest.EntityPageList)
}

func LoadPullRequests(c pullrequest.Repository, l PullRequestUpdateListener) {
	prList, err := c.Get(&pullrequest.GetOptions{})
	if err != nil {
		l.UpdateFailed(err)
	}

	l.Update(prList)
}

type PullRequestList struct {
	PageLength uint                  `json:"pagelen"`
	Page       uint                  `json:"page"`
	Size       uint                  `json:"size"`
	NextURL    string                `json:"next"`
	Values     []*pullrequest.Entity `json:"values"`
}
