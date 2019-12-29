package client

import (
	"errors"
	"fmt"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
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

type Repository struct {
	Owner string
	Name  string
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

type PullRequest struct{}

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

	return &PullRequest{}, nil
}
