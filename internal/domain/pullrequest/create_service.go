package pullrequest

type Creator interface {
	Create(o *CreateOptions) (*Entity, error)
}

type Closer interface {
	Close(o *CloseOptions) (*Entity, error)
}

type CreateService struct {
	creator Creator
}

func (cs *CreateService) Create(o *CreateOptions) (*Entity, error) {
	return cs.creator.Create(o)
}

type CloseService struct {
	closer Closer
}

func (cs *CloseService) Close(o *CloseOptions) (*Entity, error) {
	return cs.closer.Close(o)
}

func NewCreateService(c Creator) *CreateService {
	return &CreateService{c}
}

func NewCloseService(d Closer) *CloseService {
	return &CloseService{d}
}
