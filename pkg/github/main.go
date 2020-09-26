package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	preqClient "preq/pkg/client"
	"strings"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/tidwall/gjson"
)

var (
	ErrUnknownRepositoryProvider = errors.New(strings.TrimSpace(`
		unknown repository provider, expected (bitbucket, github)
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

func New(o *ClientOptions) preqClient.Client {
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

func DefaultClient() (preqClient.Client, error) {
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
	State               string `json:"state,omitempty"`
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

func (c *GithubCloudClient) GetPullRequests(o *preqClient.GetPullRequestsOptions) (*preqClient.PullRequestList, error) {
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

	var pr preqClient.PullRequestList
	parsed := gjson.ParseBytes(r.Body())
	parsed.ForEach(func(key, value gjson.Result) bool {
		pr.Values = append(pr.Values, &preqClient.PullRequest{
			ID:          value.Get("number").String(),
			Title:       value.Get("title").String(),
			URL:         value.Get("html_url").String(),
			State:       preqClient.PullRequestState(value.Get("state").String()),
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

func unmarshalPR(data []byte) (*preqClient.PullRequest, error) {
	pr := &PullRequest{}
	err := json.Unmarshal(data, pr)
	if err != nil {
		return nil, err
	}

	return &preqClient.PullRequest{
		ID:          string(pr.Number),
		Title:       pr.Title,
		URL:         pr.Links.HTML.Href,
		State:       preqClient.PullRequestState(pr.State),
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

func (c *GithubCloudClient) DeclinePullRequest(o *preqClient.DeclinePullRequestOptions) (*preqClient.PullRequest, error) {
	r, err := resty.New().R().
		SetAuthToken(c.token).
		SetBody(ghPROptions{
			State: "closed",
		}).
		SetError(bbError{}).
		Patch(fmt.Sprintf(
			"https://api.github.com/repos/%s/%s/pulls/%s",
			o.Repository.Owner,
			o.Repository.Name,
			o.ID,
		))
	if err != nil {
		return nil, err
	}

	return unmarshalPR(r.Body())
}

type getReviewsOptions struct {
	Repository preqClient.Repository
	ID         string
	User       string
}

type githubError struct {
	Message          string `json:"message"`
	DocumentationURL string `json:"documentation_url"`
}

type reviewRequest struct {
	Login             string `json:"login"`
	ID                int64  `json:"id"`
	NodeID            string `json:"node_id"`
	AvatarURL         string `json:"avatar_url"`
	GravatarID        string `json:"gravatar_id"`
	URL               string `json:"url"`
	HTMLURL           string `json:"html_url"`
	FollowersURL      string `json:"followers_url"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	OrganizationsURL  string `json:"organizations_url"`
	ReposURL          string `json:"repos_url"`
	EventsURL         string `json:"events_url"`
	ReceivedEventsURL string `json:"received_events_url"`
	Type              string `json:"type"`
	SiteAdmin         bool   `json:"site_admin"`
}

func (c *GithubCloudClient) getReviewRequests(o *getReviewsOptions) ([]int64, error) {
	r, err := resty.New().R().
		SetAuthToken(c.token).
		SetError(githubError{}).
		Get(fmt.Sprintf(
			"https://api.github.com/repos/%s/%s/pulls/%s/requested_reviewers",
			o.Repository.Owner,
			o.Repository.Name,
			o.ID,
		))
	if err != nil {
		return nil, err
	}

	res := &reviewRequests{}
	err = json.Unmarshal(r.Body(), res)
	if err != nil {
		return nil, err
	}

	var filteredUsers []int64
	for _, review := range res.Users {
		if o.User == "" || review.Login == o.User {
			filteredUsers = append(filteredUsers, review.ID)
		}
	}

	return filteredUsers, nil
}

func (c *GithubCloudClient) getReviews(o *getReviewsOptions) (*[]review, error) {
	r, err := resty.New().R().
		SetAuthToken(c.token).
		SetError(githubError{}).
		Get(fmt.Sprintf(
			"https://api.github.com/repos/%s/%s/pulls/%s/reviews",
			o.Repository.Owner,
			o.Repository.Name,
			o.ID,
		))

	if err != nil {
		return nil, err
	}

	reviews := &[]review{}
	err = json.Unmarshal(r.Body(), reviews)
	if err != nil {
		return nil, err
	}

	fmt.Println(o.ID, string(r.Body()))

	return reviews, nil
}

func (c *GithubCloudClient) ApprovePullRequest(o *preqClient.ApprovePullRequestOptions) (*preqClient.PullRequest, error) {
	_, err := resty.New().R().
		SetAuthToken(c.token).
		SetHeader("content-type", "application/json").
		SetError(githubError{}).
		SetBody(`{"event": "APPROVE"}`).
		Post(fmt.Sprintf(
			"https://api.github.com/repos/%s/%s/pulls/%s/reviews",
			o.Repository.Owner,
			o.Repository.Name,
			o.ID,
		))
	if err != nil {
		return nil, err
	}

	// TODO: Parse the response

	return &preqClient.PullRequest{
		ID: o.ID,
		// State: preqClient.PullRequestReviewState_APPROVED,
	}, nil
}

func verifyCreatePullRequestOptions(o *preqClient.CreatePullRequestOptions) error {
	if o.Source == "" {
		return errors.New("missing source branch")
	}

	if o.Destination == "" {
		return errors.New("missing destination branch")
	}

	return nil
}

func (c *GithubCloudClient) getReviewRequestsForUser(u *User) ([]*Item, error) {
	client := newClient(&newClientOptions{
		Token: c.token,
	})

	res, err := client.Search.Issues(
		context.Background(),
		// fmt.Sprintf("repo:%s type:pr state:open review-requested:%s", u.Login),
		fmt.Sprintf("type:pr state:open review-requested:%s", u.Login),
	)
	if err != nil {
		return nil, err
	}

	for _, r := range res.Items {
		fmt.Println(r.ID)
	}

	return res.Items, nil
}

func (c *GithubCloudClient) GetCurrentUser() (*preqClient.User, error) {
	client := newClient(&newClientOptions{Token: c.token})

	u, err := client.User.Current(context.Background())
	if err != nil {
		return nil, err
	}

	return &preqClient.User{
		ID: string(u.ID),
	}, nil
}

func (c *GithubCloudClient) CreatePullRequest(o *preqClient.CreatePullRequestOptions) (*preqClient.PullRequest, error) {
	err := verifyCreatePullRequestOptions(o)
	if err != nil {
		return nil, err
	}

	// uuid := c.uuid
	// if uuid == "" {
	// 	u, err := c.GetCurrentUser()
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	uuid = (string)u.ID
	// }

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

	return &preqClient.PullRequest{
		ID:          string(pr.Number),
		Title:       pr.Title,
		URL:         pr.Links.HTML.Href,
		State:       preqClient.PullRequestState(pr.State),
		Source:      pr.Head.Ref,
		Destination: pr.Base.Ref,
	}, nil
}
