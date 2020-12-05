package bitbucket

import (
	"encoding/json"
	"errors"
	"fmt"
	"preq/internal/domain"
	"preq/internal/domain/pullrequest"
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
	Repository domain.GitRepository
	username   string
	password   string
	uuid       string
}

type ClientOptions struct {
	Username string
	Password string
}

func New(o *ClientOptions) pullrequest.Repository {
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

// TODO: rename
func DefaultClient1(repo *client.Repository) (pullrequest.Repository, error) {
	config, err := getDefaultConfiguration()
	if err != nil {
		return nil, err
	}

	return &BitbucketCloudClient{
		username: config.username,
		password: config.password,
		uuid:     config.uuid,
		Repository: domain.GitRepository{
			Name: fmt.Sprintf("%s/%s", repo.Owner, repo.Name),
		},
	}, nil
}

func DefaultClient() (pullrequest.Repository, error) {
	config, err := getDefaultConfiguration()
	if err != nil {
		return nil, err
	}

	return &BitbucketCloudClient{
		username: config.username,
		password: config.password,
		uuid:     config.uuid,
		Repository: domain.GitRepository{
			Name: config.repository,
		},
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

type bitbucketPullRequest struct {
	ID          int
	Title       string
	Description string
	CreatedOn   time.Time `json:"created_on"`
	UpdatedOn   time.Time `json:"update_on"`
	State       pullrequest.State
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

// type PullRequestList struct {
// 	PageLength uint                  `json:"pagelen"`
// 	Page       uint                  `json:"page"`
// 	Size       uint                  `json:"size"`
// 	NextURL    string                `json:"next"`
// 	Values     []*client.PullRequest `json:"values"`
// }

type BitbucketCloudPullRequestPageList struct {
	pageSize int
	hasNext  bool
	counter  int
	client   *BitbucketCloudClient
	options  *pullrequest.GetOptions
}

func NewPullRequestPageList(c *BitbucketCloudClient, o *pullrequest.GetOptions) pullrequest.EntityPageList {
	return &BitbucketCloudPullRequestPageList{
		client:   c,
		options:  o,
		pageSize: 20,
		hasNext:  true,
		counter:  0,
	}
}

// TODO: return *[]*pullrequest.Entity instead
func (pl *BitbucketCloudPullRequestPageList) GetPage(page int) ([]*pullrequest.Entity, error) {
	url := fmt.Sprintf(
		"https://api.bitbucket.org/2.0/repositories/%s/pullrequests",
		pl.client.Repository.Name,
	)

	rc := resty.New()
	r, err := rc.R().
		SetBasicAuth(pl.client.username, pl.client.password).
		// SetQueryParam("state", string(o.State)).
		// TODO: get the actual page from page param
		// TODO: user state from the options param
		SetQueryParam("state", "OPEN").
		SetError(bbError{}).
		Get(url)

	if err != nil {
		return nil, err
	}
	if r.IsError() {
		return nil, errors.New(string(r.Body()))
	}

	var pr []*pullrequest.Entity
	parsed := gjson.ParseBytes(r.Body())
	// pr.PageLength = uint(parsed.Get("pagelen").Uint())
	// pr.Page = uint(parsed.Get("page").Uint())
	// pr.Size = uint(parsed.Get("size").Uint())
	nextURL := parsed.Get("next").String()
	pl.hasNext = nextURL != ""

	result := parsed.Get("values")
	result.ForEach(func(key, value gjson.Result) bool {
		pr = append(pr, &pullrequest.Entity{
			ID:          pullrequest.EntityID(value.Get("id").String()),
			Title:       value.Get("title").String(),
			URL:         value.Get("links.html.href").String(),
			State:       pullrequest.State(value.Get("state").String()),
			Source:      value.Get("source.branch.name").String(),
			Destination: value.Get("destination.branch.name").String(),
			Created:     value.Get("created_on").Time(),
			Updated:     value.Get("updated_on").Time(),
		})

		return true
	})

	return pr, nil
}

func (pl *BitbucketCloudPullRequestPageList) Next() ([]*pullrequest.Entity, error) {
	pl.counter++
	// TODO: Actually implement Next() with nextURL
	prs, err := pl.GetPage(pl.counter)
	if err != nil {
		return nil, err
	}

	pl.hasNext = len(prs) == pl.pageSize

	return prs, nil
}

func (pl *BitbucketCloudPullRequestPageList) HasNext() bool {
	return false
}

func (c *BitbucketCloudClient) Get(o *pullrequest.GetOptions) (pullrequest.EntityPageList, error) {
	return NewPullRequestPageList(c, o), nil

	// url := fmt.Sprintf(
	// 	"https://api.bitbucket.org/2.0/repositories/%s/%s/pullrequests",
	// 	c.Repository.Owner,
	// 	c.Repository.Name,
	// )

	// if o.Next != "" {
	// 	url = o.Next
	// }

	// rc := resty.New()
	// r, err := rc.R().
	// 	SetBasicAuth(c.username, c.password).
	// 	SetQueryParam("state", string(o.State)).
	// 	SetError(bbError{}).
	// 	Get(url)

	// if err != nil {
	// 	return nil, err
	// }
	// if r.IsError() {
	// 	return nil, errors.New(string(r.Body()))
	// }

	// var pr pullrequest.EntityPageList
	// parsed := gjson.ParseBytes(r.Body())
	// pr.PageLength = uint(parsed.Get("pagelen").Uint())
	// pr.Page = uint(parsed.Get("page").Uint())
	// pr.Size = uint(parsed.Get("size").Uint())
	// pr.NextURL = parsed.Get("next").String()
	// result := parsed.Get("values")
	// result.ForEach(func(key, value gjson.Result) bool {
	// 	pr.Values = append(pr.Values, &pullrequest.Entity{
	// 		ID:          value.Get("id").String(),
	// 		Title:       value.Get("title").String(),
	// 		URL:         value.Get("links.html.href").String(),
	// 		State:       domain.PullRequestState(value.Get("state").String()),
	// 		Source:      value.Get("source.branch.name").String(),
	// 		Destination: value.Get("destination.branch.name").String(),
	// 		Created:     value.Get("created_on").Time(),
	// 		Updated:     value.Get("updated_on").Time(),
	// 	})

	// 	return true
	// })

	// return &pr, nil
}

func unmarshalPR(data []byte) (*pullrequest.Entity, error) {
	pr := &bitbucketPullRequest{}
	err := json.Unmarshal(data, pr)
	if err != nil {
		return nil, err
	}

	return &pullrequest.Entity{
		ID:          pullrequest.EntityID(fmt.Sprint(pr.ID)),
		Title:       pr.Title,
		URL:         pr.Links.HTML.Href,
		State:       pr.State,
		Source:      pr.Source.Branch.Name,
		Destination: pr.Destination.Branch.Name,
	}, nil
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

func (c *BitbucketCloudClient) Close(o *pullrequest.CloseOptions) (*pullrequest.Entity, error) {
	url := fmt.Sprintf(
		"https://api.bitbucket.org/2.0/repositories/%s/pullrequests/%s/decline",
		c.Repository.Name,
		o.ID,
	)

	r, err := c.post(url)
	if err != nil {
		return nil, err
	}

	return unmarshalPR(r.Body())
}

func (c *BitbucketCloudClient) Approve(o *pullrequest.ApproveOptions) (*pullrequest.Entity, error) {
	url := fmt.Sprintf(
		"https://api.bitbucket.org/2.0/repositories/%s/pullrequests/%s/approve",
		c.Repository.Name,
		o.ID,
	)

	r, err := c.post(url)
	if err != nil {
		return nil, err
	}

	return unmarshalPR(r.Body())
}

func (c *BitbucketCloudClient) WebPageList() string {
	return fmt.Sprintf("https://bitbucket.org/%s/pull-requests/", c.Repository.Name)
}

func (c *BitbucketCloudClient) WebPage(id pullrequest.EntityID) string {
	return fmt.Sprintf("https://bitbucket.org/%s/pull-requests/%s", c.Repository.Name, id)
}

func verifyCreatePullRequestOptions(o *pullrequest.CreateOptions) error {
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

func (c *BitbucketCloudClient) GetCurrentUser() (*client.User, error) {
	r, err := resty.New().R().
		SetBasicAuth(c.username, c.password).
		SetError(bbError{}).
		Get("https://api.bitbucket.org/2.0/user")
	if err != nil {
		return nil, err
	}

	u := user{}
	err = json.Unmarshal(r.Body(), &u)
	if err != nil {
		return nil, err
	}
	return &client.User{
		ID: u.UUID,
	}, nil
}

func (c *BitbucketCloudClient) GetDefaultReviewers(o *pullrequest.CreateOptions) ([]*Reviewer, error) {
	r, err := resty.New().R().
		SetBasicAuth(c.username, c.password).
		SetError(bbError{}).
		Get(fmt.Sprintf(
			"https://api.bitbucket.org/2.0/repositories/%s/default-reviewers",
			c.Repository.Name,
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

func (c *BitbucketCloudClient) Create(o *pullrequest.CreateOptions) (*pullrequest.Entity, error) {
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
		uuid = u.ID
	}

	ddr := make([]bbPROptionsReviewer, 0, len(dr))
	for _, v := range dr {
		if v.UUID != uuid {
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
		SetError(bbError{}).
		Post(fmt.Sprintf(
			"https://api.bitbucket.org/2.0/repositories/%s/pullrequests",
			c.Repository.Name,
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

	return &pullrequest.Entity{
		ID:          pullrequest.EntityID(fmt.Sprint(pr.ID)),
		Title:       pr.Title,
		URL:         pr.Links.HTML.Href,
		State:       pr.State,
		Source:      pr.Source.Branch.Name,
		Destination: pr.Destination.Branch.Name,
	}, nil
}
