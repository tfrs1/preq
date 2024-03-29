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
	DeclinePullRequest(o *DeclinePullRequestOptions) (*PullRequest, error)
	Merge(o *MergeOptions) (*PullRequest, error)
	GetPullRequests(o *GetPullRequestsOptions) (*PullRequestList, error)
	CreatePullRequest(o *CreatePullRequestOptions) (*PullRequest, error)
	Approve(o *ApproveOptions) (*PullRequest, error)
	Unapprove(o *UnapproveOptions) (*PullRequest, error)
	GetPullRequestInfo(o *ApproveOptions) (*PullRequest, error)
	FillMiscInfoAsync(repo *Repository, pr *PullRequest) error
	GetComments(o *GetCommentsOptions) ([]*PullRequestComment, error)
	CreateComment(o *CreateCommentOptions) (*PullRequestComment, error)
	DeleteComment(o *DeleteCommentOptions) error
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

func ParseRepositoryProvider(
	s string,
	aliases map[RepositoryProvider][]string,
) (RepositoryProvider, error) {
	switch s {
	case "bitbucket.org", "bitbucket":
		return RepositoryProviderEnum.BITBUCKET, nil
	case "github.com", "github":
		return RepositoryProviderEnum.GITHUB, nil
	default:
		providers := []RepositoryProvider{
			RepositoryProviderEnum.BITBUCKET,
			RepositoryProviderEnum.GITHUB,
		}

		for _, p := range providers {
			for _, alias := range aliases[p] {
				if alias == s {
					return p, nil
				}
			}
		}
	}

	return "", ErrUnknownRepositoryProvider
}

type Repository struct {
	Provider RepositoryProvider
	Name     string
}

type RepositoryOptions struct {
	Provider RepositoryProvider
	Name     string
}

func NewRepositoryFromOptions(options *RepositoryOptions) (*Repository, error) {
	return &Repository{
		Provider: RepositoryProvider(options.Provider),
		Name:     options.Name,
	}, nil
}

type PullRequestReviewState string

const (
	PullRequestReviewState_APPROVED = "APPROVED"
)

type PullRequestState string

const (
	PullRequestState_APPROVING  = "APPROVING"
	PullRequestState_APPROVED   = "APPROVED"
	PullRequestState_DECLINING  = "DECLINING"
	PullRequestState_DECLINED   = "DECLINED"
	PullRequestState_MERGING    = "MERGING"
	PullRequestState_MERGED     = "MERGED"
	PullRequestState_OPEN       = "OPEN"
	PullRequestState_SUPERSEDED = "SUPERSEDED"
)

type GetPullRequestsOptions struct {
	Repository *Repository
	State      PullRequestState
	Next       string
}

type GetCommentsOptions struct {
	Repository *Repository
	ID         string
}

type CommentLineNumberType int

const (
	OriginalLineNumber CommentLineNumberType = iota
	NewLineNumber
)

type CreateCommentOptionsLineRef struct {
	LineNumber int
	Type       CommentLineNumberType
}

type CreateCommentOptionsParentRef struct {
	ID string
}

type DeleteCommentOptions struct {
	Repository *Repository
	ID         string
	CommentID  string
}

type CreateCommentOptions struct {
	Repository *Repository
	ID         string
	Content    string
	FilePath   string
	LineRef    *CreateCommentOptionsLineRef
	ParentRef  *CreateCommentOptionsParentRef
}

type DeclinePullRequestOptions struct {
	Repository *Repository
	ID         string
}

type MergeOptions struct {
	Repository *Repository
	ID         string
}

type ApproveOptions struct {
	Repository *Repository
	ID         string
}

type UnapproveOptions struct {
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

type PullRequestApproval struct {
	Created time.Time
	User    string
}
type PullRequestChangesRequest struct {
	Created time.Time
	User    string
}

type CommentType int

const (
	CommentTypeInline = iota + 1
	CommentTypeGlobal
	CommentTypeFile
	CommentTypeReply
)

type PullRequestComment struct {
	ID               string
	Created          time.Time
	Updated          time.Time
	Deleted          bool
	IsBeingStored    bool
	IsBeingDeleted   bool
	User             string
	Type             CommentType
	Content          string
	ParentID         string
	BeforeLineNumber uint
	AfterLineNumber  uint
	FilePath         string
	CommitHash       string
}

func (prc PullRequestComment) IsOutdated(sourceHash string) bool {
	return prc.CommitHash != sourceHash
}

type PullRequestBranch struct {
	Name string
	Hash string
}

type PullRequest struct {
	Description     string
	ID              string
	CommentCount    int
	Title           string
	User            string
	URL             string
	State           PullRequestState
	Source          PullRequestBranch
	Destination     PullRequestBranch
	Created         time.Time
	Updated         time.Time
	Approvals       []*PullRequestApproval
	Comments        []*PullRequestComment
	ChangesRequests []*PullRequestChangesRequest
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
