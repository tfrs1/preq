package errcodes

import "errors"

var (
	ErrMissingRepository               = errors.New("repository is missing")
	ErrMissingProvider                 = errors.New("provider is missing")
	ErrMissingSource                   = errors.New("source is missing")
	ErrMissingDestination              = errors.New("destination is missing")
	ErrMissingTitle                    = errors.New("title is missing")
	ErrSomeRepoParamsMissing           = errors.New("must specify both provider and repository, or none")
	ErrRepositoryMustBeInFormOwnerRepo = errors.New("repository must be in the form of 'owner/repo'")
	ErrorRepositoryProviderUnknown     = errors.New("repository provider is unknown")
)
