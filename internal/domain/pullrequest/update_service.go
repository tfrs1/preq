package pullrequest

type Updater interface {
	Update(o *UpdateOptions) (*Entity, error)
}

type UpdateService struct {
	updater Updater
}

type UpdateOptions struct {
	Title       string
	Source      string
	Destination string
	CloseBranch bool
	Draft       bool
}

func (us *UpdateService) Update(o *UpdateOptions) (*Entity, error) {
	return &Entity{}, nil
}

func NewUpdateService(c Updater) *UpdateService {
	return &UpdateService{c}
}
