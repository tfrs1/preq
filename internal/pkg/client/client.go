package client

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
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

func ParseRepositoryProvider(s string) (RepositoryProvider, error) {
	switch s {
	case "bitbucket.org", "bitbucket":
		return RepositoryProviderEnum.BITBUCKET, nil
	case "github.com", "github":
		return RepositoryProviderEnum.GITHUB, nil
	default:
		aliases := viper.GetStringSlice("bitbucket.aliases")
		if aliases == nil {
			log.Warn().
				Msg(fmt.Sprintf("Parsing unknown provider: %v. Add repository info to local preq configuration (.preqcfg)", s))
			break
		}

		for _, a := range aliases {
			if a == s {
				return RepositoryProviderEnum.BITBUCKET, nil
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

type PullRequestComment struct {
	ID               string
	Created          time.Time
	Updated          time.Time
	Deleted          bool
	IsBeingStored    bool
	IsBeingDeleted   bool
	User             string
	Content          string
	ParentID         string
	BeforeLineNumber uint
	AfterLineNumber  uint
	FilePath         string
}

type PullRequestBranch struct {
	Name string
	Hash string
}

type PullRequest struct {
	ID              string
	Title           string
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
