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

func DefaultClient() (*client, error) {
	username := viper.GetString("bitbucket.username")
	if username == "" {
		return nil, ErrMissingBitbucketUsername
	}
	password := viper.GetString("bitbucket.password")
	if password == "" {
		return nil, ErrMissingBitbucketPassword
	}

	return New(&ClientOptions{
		Username: username,
		Password: password,
	}), nil
}

type bbPRSourceBranchOptions struct {
	Name string `json:"name,omitempty"`
}

type bbPRSourceOptions struct {
	Branch bbPRSourceBranchOptions `json:"branch,omitempty"`
}

type bbPROptions struct {
	Title       string            `json:"title,omitempty"`
	Source      bbPRSourceOptions `json:"source,omitempty"`
	Destination bbPRSourceOptions `json:"destination,omitempty"`
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
}

func (c *client) GetPullRequests(o *GetPullRequestsOptions) *[]*PullRequest {
	rc := resty.New()
	r, err := rc.R().
		SetBasicAuth(c.username, c.password).
		SetQueryParam("state", string(o.State)).
		SetError(bbError{}).
		Get(fmt.Sprintf(
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

	prs := new([]*PullRequest)
	return prs
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

func (c *client) CreatePullRequest(o *CreatePullRequestOptions) (*PullRequest, error) {
	err := verifyCreatePullRequestOptions(o)
	if err != nil {
		return nil, err
	}

	r, err := resty.New().R().
		SetBasicAuth(c.username, c.password).
		SetHeader("content-type", "application/json").
		SetBody(bbPROptions{
			Title: o.Title,
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
