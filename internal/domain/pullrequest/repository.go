package pullrequest

type EntityPageList interface {
	GetPage(int) ([]*Entity, error)
	Next() ([]*Entity, error)
	HasNext() bool
}

type Repository interface {
	Get(*GetOptions) (EntityPageList, error)
	Create(*CreateOptions) (*Entity, error)
	Approve(*ApproveOptions) (*Entity, error)
	Close(*CloseOptions) (*Entity, error)
	WebPageList() string
	WebPage(EntityID) string
}

type GetOptions struct {
	State State
	Next  string
}

type CreateOptions struct {
	Title       string
	Source      string
	Destination string
	CloseBranch bool
	Draft       bool
	// ID string
}

type ApproveOptions struct {
	ID string
}

type CloseOptions struct {
	ID string
}
