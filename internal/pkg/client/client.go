package client

import (
	"errors"
	"reflect"
	"strings"
	"time"
)

var (
	ErrUnknownRepositoryProvider = errors.New(strings.TrimSpace(`
		unknown repository provider, expected (bitbucket)
	`))
	ErrMissingBitbucketUsername = errors.New("bitbucket username is missing")
	ErrMissingBitbucketPassword = errors.New("bitbucket password is missing")
)

type Client interface {
	Close(o *ClosePullRequestOptions) (*PullRequest, error)
	Get(o *GetPullRequestOptions) (*PullRequestList, error)
	Create(o *CreatePullRequestOptions) (*PullRequest, error)
	Approve(o *ApprovePullRequestOptions) (*PullRequest, error)
}

type RepositoryProvider string

func (rp RepositoryProvider) IsValid() bool {
	v := reflect.ValueOf(*RepositoryProviderEnum)

	for i := 0; i < v.NumField(); i++ {
		if rp == v.Field(i).Interface() {
			return true
		}
	}

	return false
}

type list struct {
	BITBUCKET RepositoryProvider
	GITHUB    RepositoryProvider
}

var RepositoryProviderEnum = &list{
	BITBUCKET: RepositoryProvider("bitbucket"),
	GITHUB:    RepositoryProvider("github"),
}

func ParseRepositoryProvider(s string) (RepositoryProvider, error) {
	switch s {
	case "bitbucket.org", "bitbucket":
		return RepositoryProviderEnum.BITBUCKET, nil
	case "github.com", "github":
		return RepositoryProviderEnum.GITHUB, nil
	}

	return "", ErrUnknownRepositoryProvider
}

type Repository struct {
	Provider RepositoryProvider
	Owner    string
	Name     string
}

type RepositoryOptions struct {
	Provider RepositoryProvider
	Name     string
}

func NewRepositoryFromOptions(options *RepositoryOptions) (*Repository, error) {
	r := strings.Split(options.Name, "/")
	if len(r) != 2 {
		return nil, errors.New("invalid repo name")
	}

	return &Repository{
		Provider: RepositoryProvider(options.Provider),
		Owner:    r[0],
		Name:     r[1],
	}, nil
}

type PullRequestReviewState string

const (
	PullRequestReviewState_APPROVED = "APPROVED"
)

type PullRequestState string

const (
	PullRequestState_CLOSED     = "CLOSED"
	PullRequestState_OPEN       = "OPEN"
	PullRequestState_MERGED     = "MERGED"
	PullRequestState_SUPERSEDED = "SUPERSEDED"
)

type GetPullRequestOptions struct {
	Repository *Repository
	State      PullRequestState
	Next       string
}

type ClosePullRequestOptions struct {
	Repository *Repository
	ID         string
}

type ApprovePullRequestOptions struct {
	Repository *Repository
	ID         string
}

type CreatePullRequestOptions struct {
	Repository  *Repository
	Title       string
	Source      string
	Destination string
	CloseBranch bool
	Draft       bool
}

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

type User struct {
	ID string
}

// type Reviewer struct {
// 	Username    string `json:"username"`
// 	DisplayName string `json:"display_name"`
// 	UUID        string
// 	Nickname    string
// 	Type        string
// 	AccountID   string `json:"account_id"`
// }

type PullRequestList struct {
	PageLength uint           `json:"pagelen"`
	Page       uint           `json:"page"`
	Size       uint           `json:"size"`
	NextURL    string         `json:"next"`
	Values     []*PullRequest `json:"values"`
}

// func verifyCreatePullRequestOptions(o *CreatePullRequestOptions) error {
// 	if o.Source == "" {
// 		return errors.New("missing source branch")
// 	}

// 	if o.Destination == "" {
// 		return errors.New("missing destination branch")
// 	}

// 	return nil
// }

// type User struct {
// 	UUID string `json:"uuid"`
// }
