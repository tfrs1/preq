package bitbucket

import (
	"encoding/json"
	"errors"
	"fmt"
	"preq/internal/pkg/client"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/tidwall/gjson"
)

var (
	ErrUnknownRepositoryProvider = errors.New(strings.TrimSpace(`
		unknown repository provider, expected (bitbucket)
	`))
	ErrMissingBitbucketUsername = errors.New("bitbucket username is missing")
	ErrMissingBitbucketPassword = errors.New("bitbucket password is missing")
)

type BitbucketCloudClient struct {
	username   string
	password   string
	uuid       string
	repository string
}

type ClientOptions struct {
	Username string
	Password string
}

func New(o *ClientOptions) client.Client {
	return &BitbucketCloudClient{
		username: o.Username,
		password: o.Password,
	}
}

type clientConfiguration struct {
	username   string
	password   string
	uuid       string
	repository string
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
	repository := viper.GetString("default.repository")

	return &clientConfiguration{
		username:   username,
		password:   password,
		uuid:       uuid,
		repository: repository,
	}, nil
}

func DefaultClientCustom(repository string) (client.Client, error) {
	config, err := getDefaultConfiguration()
	if err != nil {
		return nil, err
	}

	return &BitbucketCloudClient{
		username:   config.username,
		password:   config.password,
		uuid:       config.uuid,
		repository: repository,
	}, nil
}

func DefaultClient() (client.Client, error) {
	config, err := getDefaultConfiguration()
	if err != nil {
		return nil, err
	}

	return &BitbucketCloudClient{
		username:   config.username,
		password:   config.password,
		uuid:       config.uuid,
		repository: config.repository,
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

type bbErrorReal struct {
	Error struct {
		Message string
	}
}

type bitbucketPullRequest struct {
	ID          int
	Title       string
	Description string
	CreatedOn   time.Time `json:"created_on"`
	UpdatedOn   time.Time `json:"update_on"`
	State       client.PullRequestState
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

type Reviewer struct {
	DisplayName string `json:"display_name"`
	UUID        string
	Nickname    string
	Type        string
	AccountID   string `json:"account_id"`
}

type defaultReviewersResponse struct {
	Values []struct {
		Reviewer *Reviewer `json:"user"`
	}
}

// type PullRequestList struct {
// 	PageLength uint                  `json:"pagelen"`
// 	Page       uint                  `json:"page"`
// 	Size       uint                  `json:"size"`
// 	NextURL    string                `json:"next"`
// 	Values     []*client.PullRequest `json:"values"`
// }

func (c *BitbucketCloudClient) GetPullRequests(
	o *client.GetPullRequestsOptions,
) (*client.PullRequestList, error) {
	url := fmt.Sprintf(
		"https://api.bitbucket.org/2.0/repositories/%s/pullrequests",
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

	var pr client.PullRequestList
	parsed := gjson.ParseBytes(r.Body())
	pr.PageLength = uint(parsed.Get("pagelen").Uint())
	pr.Page = uint(parsed.Get("page").Uint())
	pr.Size = uint(parsed.Get("size").Uint())
	pr.NextURL = parsed.Get("next").String()
	result := parsed.Get("values")
	result.ForEach(func(key, value gjson.Result) bool {
		pr.Values = append(pr.Values, &client.PullRequest{
			ID:           value.Get("id").String(),
			Title:        value.Get("title").String(),
			URL:          value.Get("links.html.href").String(),
			State:        client.PullRequestState(value.Get("state").String()),
			Source:       value.Get("source.branch.name").String(),
			Destination:  value.Get("destination.branch.name").String(),
			Created:      value.Get("created_on").Time(),
			Updated:      value.Get("updated_on").Time(),
			CommentCount: uint(value.Get("comment_count").Uint()),
		})

		return true
	})

	return &pr, nil
}

func unmarshalPR(data []byte) (*client.PullRequest, error) {
	pr := &bitbucketPullRequest{}
	err := json.Unmarshal(data, pr)
	if err != nil {
		return nil, err
	}

	return &client.PullRequest{
		ID:          fmt.Sprint(pr.ID),
		Title:       pr.Title,
		URL:         pr.Links.HTML.Href,
		State:       pr.State,
		Source:      pr.Source.Branch.Name,
		Destination: pr.Destination.Branch.Name,
	}, nil
}

func (c *BitbucketCloudClient) get(url string) (*resty.Response, error) {
	rc := resty.New()
	r, err := rc.R().
		SetBasicAuth(c.username, c.password).
		SetError(bbError{}).
		Get(url)

	if err != nil {
		return nil, err
	}

	if r.IsError() {
		return nil, errors.New(string(r.Body()))
	}

	return r, nil
}

func (c *BitbucketCloudClient) post(url string) (*resty.Response, error) {
	rc := resty.New()
	r, err := rc.R().
		SetBasicAuth(c.username, c.password).
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

func (c *BitbucketCloudClient) Merge(
	o *client.MergeOptions,
) (*client.PullRequest, error) {
	url := fmt.Sprintf(
		"https://api.bitbucket.org/2.0/repositories/%s/pullrequests/%s/merge",
		o.Repository.Name,
		o.ID,
	)

	r, err := c.post(url)
	if err != nil {
		return nil, err
	}

	return unmarshalPR(r.Body())
}

func (c *BitbucketCloudClient) DeclinePullRequest(
	o *client.DeclinePullRequestOptions,
) (*client.PullRequest, error) {
	url := fmt.Sprintf(
		"https://api.bitbucket.org/2.0/repositories/%s/pullrequests/%s/decline",
		o.Repository.Name,
		o.ID,
	)

	r, err := c.post(url)
	if err != nil {
		return nil, err
	}

	return unmarshalPR(r.Body())
}

func (c *BitbucketCloudClient) ApprovePullRequest(
	o *client.ApprovePullRequestOptions,
) (*client.PullRequest, error) {
	url := fmt.Sprintf(
		"https://api.bitbucket.org/2.0/repositories/%s/pullrequests/%s/approve",
		o.Repository.Name,
		o.ID,
	)

	r, err := c.post(url)
	if err != nil {
		return nil, err
	}

	return unmarshalPR(r.Body())
}

func (c *BitbucketCloudClient) GetPullRequestInfo(
	o *client.ApprovePullRequestOptions,
) (*client.PullRequest, error) {
	return nil, errors.New("not implemented")
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

type user struct {
	UUID string `json:"uuid"`
}

func (c *BitbucketCloudClient) GetDefaultReviewer(
	username string,
) (*client.User, error) {
	r, err := resty.New().R().
		SetBasicAuth(c.username, c.password).
		SetResult(user{}).
		SetError(bbErrorReal{}).
		Get(fmt.Sprintf(
			"https://api.bitbucket.org/2.0/repositories/%s/default-reviewers/%s",
			c.repository,
			username,
		))
	if err != nil {
		return nil, err
	}

	if r.IsError() {
		return nil, errors.New(r.Error().(*bbErrorReal).Error.Message)
	}

	return &client.User{
		ID: r.Result().(*user).UUID,
	}, nil
}

func (c *BitbucketCloudClient) GetCurrentUser() (*client.User, error) {
	r, err := resty.New().R().
		SetBasicAuth(c.username, c.password).
		SetResult(user{}).
		SetError(bbError{}).
		Get("https://api.bitbucket.org/2.0/user")
	if err != nil {
		return nil, err
	}

	if r.IsError() {
		return nil, errors.New(r.Error().(*bbError).Message)
	}

	return &client.User{
		ID: r.Result().(*user).UUID,
	}, nil
}

func (c *BitbucketCloudClient) GetDefaultReviewers(
	o *client.CreatePullRequestOptions,
) ([]*Reviewer, error) {
	r, err := resty.New().R().
		SetBasicAuth(c.username, c.password).
		SetError(bbError{}).
		Get(fmt.Sprintf(
			"https://api.bitbucket.org/2.0/repositories/%s/effective-default-reviewers",
			c.repository,
		))

	if err != nil {
		return nil, err
	}

	pr := &defaultReviewersResponse{}
	err = json.Unmarshal(r.Body(), pr)
	if err != nil {
		return nil, err
	}

	res := make([]*Reviewer, 0, len(pr.Values))
	for _, v := range pr.Values {
		res = append(res, v.Reviewer)
	}

	return res, nil
}

func (c *BitbucketCloudClient) CreatePullRequest(
	o *client.CreatePullRequestOptions,
) (*client.PullRequest, error) {
	err := verifyCreatePullRequestOptions(o)
	if err != nil {
		return nil, err
	}

	dr, err := c.GetDefaultReviewers(o)
	if err != nil {
		return nil, err
	}

	ddr := make([]bbPROptionsReviewer, 0, len(dr))
	for _, v := range dr {
		if v.UUID != c.uuid {
			ddr = append(ddr, bbPROptionsReviewer{UUID: v.UUID})
		}
	}

	if o.Draft {
		o.Title = fmt.Sprintf("[DRAFT] %s", o.Title)
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
		SetError(bbErrorReal{}).
		Post(fmt.Sprintf(
			"https://api.bitbucket.org/2.0/repositories/%s/pullrequests",
			c.repository,
		))

	if err != nil {
		log.Fatal(err)
	}
	if r.IsError() {
		m := r.Error().(*bbErrorReal).Error.Message
		if strings.HasSuffix(
			m,
			"is the author and cannot be included as a reviewer.",
		) {
			errorMessage := "You have to add your UUID to the configuration\n\n"
			for _, v := range dr {
				errorMessage += fmt.Sprintf(
					"%s - %s - %s\n",
					v.DisplayName,
					v.Nickname,
					v.UUID,
				)
			}
			log.Fatal(errorMessage)
		}
		log.Fatal(string(r.Body()))
	}
	pr := &bitbucketPullRequest{}
	err = json.Unmarshal(r.Body(), pr)
	if err != nil {
		log.Fatal(err)
	}

	return &client.PullRequest{
		ID:          fmt.Sprint(pr.ID),
		Title:       pr.Title,
		URL:         pr.Links.HTML.Href,
		State:       pr.State,
		Source:      pr.Source.Branch.Name,
		Destination: pr.Destination.Branch.Name,
	}, nil
}
