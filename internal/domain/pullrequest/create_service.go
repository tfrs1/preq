package pullrequest

type Creator interface {
	Create(o *CreateOptions) (*Entity, error)
}

type CreateService struct {
	creator Creator
}

func (cs *CreateService) Create(o *CreateOptions) (*Entity, error) {
	return cs.creator.Create(o)
}

func NewCreateService(c Creator) *CreateService {
	return &CreateService{c}
}
