package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/tidwall/gjson"
)

var (
	ErrUnknownRepositoryProvider = errors.New(strings.TrimSpace(`
		unknown repository provider, expected (bitbucket-cloud)
	`))
	ErrMissingBitbucketUsername = errors.New("bitbucket username is missing")
	ErrMissingBitbucketPassword = errors.New("bitbucket password is missing")
)

type client struct {
	username string
	password string
	uuid     string
}

type ClientOptions struct {
	Username string
	Password string
}

func New(o *ClientOptions) *client {
	return &client{
		username: o.Username,
		password: o.Password,
	}
}

type clientConfiguration struct {
	username string
	password string
	uuid     string
}

func getDefaultConfiguration() (*clientConfiguration, error) {
	username := viper.GetString("bitbucket.username")
	if username == "" {
		return nil, ErrMissingBitbucketUsername
	}
	password := viper.GetString("bitbucket.password")
	if password == "" {
		return nil, ErrMissingBitbucketPassword
	}
	uuid := viper.GetString("bitbucket.uuid")

	return &clientConfiguration{
		username: username,
		password: password,
		uuid:     uuid,
	}, nil
}

func DefaultClient() (*client, error) {
	config, err := getDefaultConfiguration()
	if err != nil {
		return nil, err
	}

	return &client{
		username: config.username,
		password: config.password,
		uuid:     config.uuid,
	}, nil
}

type bbPRSourceBranchOptions struct {
	Name string `json:"name,omitempty"`
}

type bbPRSourceOptions struct {
	Branch bbPRSourceBranchOptions `json:"branch,omitempty"`
}

type bbPROptionsReviewer struct {
	UUID     string `json:"uuid"`
	Username string `json:"username"`
}

type bbPROptions struct {
	Title             string                `json:"title,omitempty"`
	Source            bbPRSourceOptions     `json:"source,omitempty"`
	Destination       bbPRSourceOptions     `json:"destination,omitempty"`
	CloseSourceBranch bool                  `json:"close_source_branch,omitempty"`
	Reviewers         []bbPROptionsReviewer `json:"reviewers"`
}

type bbError struct {
	Error   interface{}
	Message string
}

type RepositoryProvider string

const (
	RepositoryProvider_BITBUCKET_CLOUD = RepositoryProvider("bitbucket-cloud")
)

type bitbucketPullRequest struct {
	ID          int
	Title       string
	Description string
	CreatedOn   time.Time `json:"created_on"`
	UpdatedOn   time.Time `json:"update_on"`
	State       PullRequestState
	Author      struct {
		DisplayName string `json:"display_name"`
		UUID        string
		Nickname    string
	}
	Links struct {
		HTML struct {
			Href string
		}
	}
	Destination struct {
		Branch struct {
			Name string
		}
	}
	Source struct {
		Branch struct {
			Name string
		}
	}
	CloseSourceBranch bool `json:"close_source_branch"`
}

func ParseRepositoryProvider(s string) (RepositoryProvider, error) {
	switch s {
	case "bitbucket.org", "bitbucket-cloud":
		return RepositoryProvider_BITBUCKET_CLOUD, nil
	}

	return "", ErrUnknownRepositoryProvider
}

type Repository struct {
	Provider RepositoryProvider
	Owner    string
	Name     string
}

type RepositoryOptions struct {
	Provider           string
	FullRepositoryName string
}

func NewRepositoryFromOptions(options *RepositoryOptions) (*Repository, error) {
	r := strings.Split(options.FullRepositoryName, "/")
	if len(r) != 2 {
		return nil, errors.New("invalid repo name")
	}

	return &Repository{
		Provider: RepositoryProvider(options.Provider),
		Owner:    r[0],
		Name:     r[1],
	}, nil
}

type PullRequestState string

const (
	PullRequestState_DECLINED   = "DECLINED"
	PullRequestState_OPEN       = "OPEN"
	PullRequestState_MERGED     = "MERGED"
	PullRequestState_SUPERSEDED = "SUPERSEDED"
)

type GetPullRequestsOptions struct {
	Repository *Repository
	State      PullRequestState
	Next       string
}

type CreatePullRequestOptions struct {
	Repository  *Repository
	Title       string
	Source      string
	Destination string
	CloseBranch bool
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

type Reviewer struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	UUID        string
	Nickname    string
	Type        string
	AccountID   string `json:"account_id"`
}

type defaultReviewersResponse struct {
	Values []*Reviewer
}

type PullRequestList struct {
	PageLenght uint           `json:"pagelen"`
	Page       uint           `json:"page"`
	Size       uint           `json:"size"`
	NextURL    string         `json:"next"`
	Values     []*PullRequest `json:"values"`
}

func (c *client) GetPullRequests(o *GetPullRequestsOptions) (*PullRequestList, error) {
	url := fmt.Sprintf(
		"https://api.bitbucket.org/2.0/repositories/%s/%s/pullrequests",
		o.Repository.Owner,
		o.Repository.Name,
	)

	if o.Next != "" {
		url = o.Next
	}

	rc := resty.New()
	r, err := rc.R().
		SetBasicAuth(c.username, c.password).
		SetQueryParam("state", string(o.State)).
		SetError(bbError{}).
		Get(url)

	if err != nil {
		return nil, err
	}
	if r.IsError() {
		return nil, errors.New(string(r.Body()))
	}

	var pr PullRequestList
	parsed := gjson.ParseBytes(r.Body())
	pr.PageLenght = uint(parsed.Get("pagelen").Uint())
	pr.Page = uint(parsed.Get("page").Uint())
	pr.Size = uint(parsed.Get("size").Uint())
	pr.NextURL = parsed.Get("next").String()
	result := parsed.Get("values")
	result.ForEach(func(key, value gjson.Result) bool {
		pr.Values = append(pr.Values, &PullRequest{
			ID:          value.Get("id").String(),
			Title:       value.Get("title").String(),
			URL:         value.Get("links.html.href").String(),
			State:       PullRequestState(value.Get("state").String()),
			Source:      value.Get("source.branch.name").String(),
			Destination: value.Get("destination.branch.name").String(),
			Created:     value.Get("created_on").Time(),
			Updated:     value.Get("updated_on").Time(),
		})

		return true
	})

	return &pr, nil
}

func verifyCreatePullRequestOptions(o *CreatePullRequestOptions) error {
	if o.Source == "" {
		return errors.New("missing source branch")
	}

	if o.Destination == "" {
		return errors.New("missing destination branch")
	}

	return nil
}

type User struct {
	UUID string `json:"uuid"`
}

func (c *client) GetCurrentUser() (*User, error) {
	r, err := resty.New().R().
		SetBasicAuth(c.username, c.password).
		SetError(bbError{}).
		Get("https://api.bitbucket.org/2.0/user")
	if err != nil {
		return nil, err
	}

	user := User{}
	err = json.Unmarshal(r.Body(), &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *client) GetDefaultReviewers(o *CreatePullRequestOptions) ([]*Reviewer, error) {
	r, err := resty.New().R().
		SetBasicAuth(c.username, c.password).
		SetError(bbError{}).
		Get(fmt.Sprintf(
			"https://api.bitbucket.org/2.0/repositories/%s/%s/default-reviewers",
			o.Repository.Owner,
			o.Repository.Name,
		))

	if err != nil {
		return nil, err
	}

	pr := &defaultReviewersResponse{}
	err = json.Unmarshal(r.Body(), pr)
	if err != nil {
		return nil, err
	}

	return pr.Values, nil
}

func (c *client) CreatePullRequest(o *CreatePullRequestOptions) (*PullRequest, error) {
	err := verifyCreatePullRequestOptions(o)
	if err != nil {
		return nil, err
	}

	dr, err := c.GetDefaultReviewers(o)
	if err != nil {
		return nil, err
	}

	uuid := c.uuid
	if uuid == "" {
		u, err := c.GetCurrentUser()
		if err != nil {
			return nil, err
		}
		uuid = u.UUID
	}

	ddr := make([]bbPROptionsReviewer, 0, len(dr))
	for _, v := range dr {
		if v.UUID != uuid {
			ddr = append(ddr, bbPROptionsReviewer{UUID: v.UUID})
		}
	}

	r, err := resty.New().R().
		SetBasicAuth(c.username, c.password).
		SetHeader("content-type", "application/json").
		SetBody(bbPROptions{
			Title:             o.Title,
			CloseSourceBranch: o.CloseBranch,
			Reviewers:         ddr,
			Source: bbPRSourceOptions{
				Branch: bbPRSourceBranchOptions{
					Name: o.Source,
				},
			},
			Destination: bbPRSourceOptions{
				Branch: bbPRSourceBranchOptions{
					Name: o.Destination,
				},
			},
		}).
		SetError(bbError{}).
		Post(fmt.Sprintf(
			"https://api.bitbucket.org/2.0/repositories/%s/%s/pullrequests",
			o.Repository.Owner,
			o.Repository.Name,
		))

	if err != nil {
		log.Fatal(err)
	}
	if r.IsError() {
		log.Fatal(string(r.Body()))
	}
	pr := &bitbucketPullRequest{}
	err = json.Unmarshal(r.Body(), pr)
	if err != nil {
		log.Fatal(err)
	}

	return &PullRequest{
		ID:          string(pr.ID),
		Title:       pr.Title,
		URL:         pr.Links.HTML.Href,
		State:       pr.State,
		Source:      pr.Source.Branch.Name,
		Destination: pr.Destination.Branch.Name,
	}, nil
}
