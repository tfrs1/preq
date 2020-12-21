package pullrequest

type EntityPageList interface {
	GetPage(int) ([]*Entity, error)
}

type Repository interface {
	Get(*GetOptions) (EntityPageList, error)
	Create(*CreateOptions) (*Entity, error)
	Approve(*ApproveOptions) (*Entity, error)
	Decline(*DeclineOptions) (*Entity, error)
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

type DeclineOptions struct {
	ID string
}
