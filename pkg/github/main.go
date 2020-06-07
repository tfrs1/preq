package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"preq/pkg/client"
	"strings"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/tidwall/gjson"
)

var (
	ErrUnknownRepositoryProvider = errors.New(strings.TrimSpace(`
		unknown repository provider, expected (bitbucket-cloud, github)
	`))
	ErrMissingGithubUsername = errors.New("github username is missing")
	ErrMissingGithubPassword = errors.New("github password is missing")
)

type GithubCloudClient struct {
	username string
	token    string
	uuid     string
}

type ClientOptions struct {
	Username string
	Password string
	Token    string
}

func New(o *ClientOptions) client.Client {
	return &GithubCloudClient{
		username: o.Username,
		token:    o.Token,
	}
}

type clientConfiguration struct {
	username string
	token    string
	uuid     string
}

func getDefaultConfiguration() (*clientConfiguration, error) {
	username := viper.GetString("github.username")
	if username == "" {
		return nil, ErrMissingGithubUsername
	}
	token := viper.GetString("github.token")
	if token == "" {
		return nil, ErrMissingGithubPassword
	}
	uuid := viper.GetString("bitbucket.uuid")

	return &clientConfiguration{
		username: username,
		token:    token,
		uuid:     uuid,
	}, nil
}

func DefaultClient() (client.Client, error) {
	config, err := getDefaultConfiguration()
	if err != nil {
		return nil, err
	}

	return &GithubCloudClient{
		username: config.username,
		token:    config.token,
		uuid:     config.uuid,
	}, nil
}

type ghPRSourceBranchOptions struct {
	Name string `json:"name,omitempty"`
}

type ghPRSourceOptions struct {
	Branch ghPRSourceBranchOptions `json:"branch,omitempty"`
}

type ghPROptionsReviewer struct {
	UUID     string `json:"uuid"`
	Username string `json:"username"`
}

type ghPROptions struct {
	Title               string `json:"title,omitempty"`
	Head                string `json:"head,omitempty"`
	Base                string `json:"base,omitempty"`
	Body                string `json:"body,omitempty"`
	MaintainerCanModify bool   `json:"maintainer_can_modify,omitempty`
	Draft               bool   `json:"draft,omitempty"`
}

type bbError struct {
	Error   interface{}
	Message string
}

// type PullRequestList struct {
// 	PageLength uint                  `json:"pagelen"`
// 	Page       uint                  `json:"page"`
// 	Size       uint                  `json:"size"`
// 	NextURL    string                `json:"next"`
// 	Values     []*client.PullRequest `json:"values"`
// }

func (c *GithubCloudClient) GetPullRequests(o *client.GetPullRequestsOptions) (*client.PullRequestList, error) {
	url := fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/pulls",
		o.Repository.Owner,
		o.Repository.Name,
	)

	if o.Next != "" {
		url = o.Next
	}

	rc := resty.New()
	r, err := rc.R().
		SetAuthToken(c.token).
		SetQueryParam("state", string(o.State)).
		SetError(bbError{}).
		Get(url)

	if err != nil {
		return nil, err
	}
	if r.IsError() {
		return nil, errors.New(string(r.Body()))
	}

	var pr client.PullRequestList
	parsed := gjson.ParseBytes(r.Body())
	parsed.ForEach(func(key, value gjson.Result) bool {
		pr.Values = append(pr.Values, &client.PullRequest{
			ID:          value.Get("number").String(),
			Title:       value.Get("title").String(),
			URL:         value.Get("html_url").String(),
			State:       client.PullRequestState(value.Get("state").String()),
			Source:      value.Get("head.ref").String(),
			Destination: value.Get("base.ref").String(),
			Created:     value.Get("created_at").Time(),
			Updated:     value.Get("updated_at").Time(),
		})

		return true
	})

	// pr.PageLength = uint(parsed.Get("pagelen").Uint())
	// pr.Page = uint(parsed.Get("page").Uint())
	// pr.Size = uint(parsed.Get("size").Uint())

	// pr.NextURL = r.Header().Get("Link") // Split on , -> split on ; check for rel="next"

	return &pr, nil
}

func unmarshalPR(data []byte) (*client.PullRequest, error) {
	pr := &PullRequest{}
	err := json.Unmarshal(data, pr)
	if err != nil {
		return nil, err
	}

	return &client.PullRequest{
		ID:          string(pr.Number),
		Title:       pr.Title,
		URL:         pr.Links.HTML.Href,
		State:       client.PullRequestState(pr.State),
		Source:      pr.Head.Ref,
		Destination: pr.Base.Ref,
	}, nil
}

func (c *GithubCloudClient) post(url string) (*resty.Response, error) {
	rc := resty.New()
	r, err := rc.R().
		SetAuthToken(c.token).
		SetError(bbError{}).
		Post(url)

	if err != nil {
		return nil, err
	}

	if r.IsError() {
		return nil, errors.New(string(r.Body()))
	}

	return r, nil
}

func (c *GithubCloudClient) DeclinePullRequest(o *client.DeclinePullRequestOptions) (*client.PullRequest, error) {
	url := fmt.Sprintf(
		"https://api.bitbucket.org/2.0/repositories/%s/%s/pullrequests/%s/decline",
		o.Repository.Owner,
		o.Repository.Name,
		o.ID,
	)

	r, err := c.post(url)
	if err != nil {
		return nil, err
	}

	return unmarshalPR(r.Body())
}

func (c *GithubCloudClient) ApprovePullRequest(o *client.ApprovePullRequestOptions) (*client.PullRequest, error) {
	url := fmt.Sprintf(
		"https://api.bitbucket.org/2.0/repositories/%s/%s/pullrequests/%s/approve",
		o.Repository.Owner,
		o.Repository.Name,
		o.ID,
	)

	r, err := c.post(url)
	if err != nil {
		return nil, err
	}

	return unmarshalPR(r.Body())
}

func verifyCreatePullRequestOptions(o *client.CreatePullRequestOptions) error {
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

func (c *GithubCloudClient) GetCurrentUser() (*User, error) {
	r, err := resty.New().R().
		SetAuthToken(c.token).
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

func (c *GithubCloudClient) CreatePullRequest(o *client.CreatePullRequestOptions) (*client.PullRequest, error) {
	err := verifyCreatePullRequestOptions(o)
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

	r, err := resty.New().R().
		SetAuthToken(c.token).
		// SetHeader("content-type", "application/json").
		SetBody(ghPROptions{
			Title: o.Title,
			// CloseSourceBranch: o.CloseBranch,
			Head: o.Source,
			Base: o.Destination,
		}).
		SetError(bbError{}).
		Post(fmt.Sprintf(
			"https://api.github.com/repos/%s/%s/pulls",
			o.Repository.Owner,
			o.Repository.Name,
		))

	if err != nil {
		log.Fatal(err)
	}
	if r.IsError() {
		log.Fatal(string(r.Body()))
	}
	pr := &PullRequest{}
	err = json.Unmarshal(r.Body(), pr)
	if err != nil {
		log.Fatal(err)
	}

	return &client.PullRequest{
		ID:          string(pr.Number),
		Title:       pr.Title,
		URL:         pr.Links.HTML.Href,
		State:       client.PullRequestState(pr.State),
		Source:      pr.Head.Ref,
		Destination: pr.Base.Ref,
	}, nil
}
