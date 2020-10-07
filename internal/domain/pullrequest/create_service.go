package pullrequest

type Creator interface {
	Create(o *CreateOptions) (*Entity, error)
}

type CreateService struct {
	creator Creator
}

type CreateOptions struct {
	Title       string
	Source      string
	Destination string
	CloseBranch bool
	Draft       bool
}

func (cs *CreateService) Create(o *CreateOptions) (*Entity, error) {
	return cs.creator.Create(o)
}

func NewCreateService(c Creator) *CreateService {
	return &CreateService{c}
}
