package github

import (
	"context"
	"encoding/json"

	"github.com/go-resty/resty/v2"
)

type newClientOptions struct {
	Token string
}

type service struct {
	token string
}

type SearchService service
type UserService service

type client struct {
	Search *SearchService
	User   *UserService
}

func newClient(o *newClientOptions) *client {
	return &client{
		Search: &SearchService{
			token: o.Token,
		},
		User: &UserService{
			token: o.Token,
		},
	}
}

type User struct {
	Login                   string `json:"login"`
	ID                      int64  `json:"id"`
	NodeID                  string `json:"node_id"`
	AvatarURL               string `json:"avatar_url"`
	GravatarID              string `json:"gravatar_id"`
	URL                     string `json:"url"`
	HTMLURL                 string `json:"html_url"`
	FollowersURL            string `json:"followers_url"`
	FollowingURL            string `json:"following_url"`
	GistsURL                string `json:"gists_url"`
	StarredURL              string `json:"starred_url"`
	SubscriptionsURL        string `json:"subscriptions_url"`
	OrganizationsURL        string `json:"organizations_url"`
	ReposURL                string `json:"repos_url"`
	EventsURL               string `json:"events_url"`
	ReceivedEventsURL       string `json:"received_events_url"`
	Type                    string `json:"type"`
	SiteAdmin               bool   `json:"site_admin"`
	Name                    string `json:"name"`
	Company                 string `json:"company"`
	Blog                    string `json:"blog"`
	Location                string `json:"location"`
	Email                   string `json:"email"`
	Hireable                bool   `json:"hireable"`
	Bio                     string `json:"bio"`
	TwitterUsername         string `json:"twitter_username"`
	PublicRepos             int64  `json:"public_repos"`
	PublicGists             int64  `json:"public_gists"`
	Followers               int64  `json:"followers"`
	Following               int64  `json:"following"`
	CreatedAt               string `json:"created_at"`
	UpdatedAt               string `json:"updated_at"`
	PrivateGists            int64  `json:"private_gists"`
	TotalPrivateRepos       int64  `json:"total_private_repos"`
	OwnedPrivateRepos       int64  `json:"owned_private_repos"`
	DiskUsage               int64  `json:"disk_usage"`
	Collaborators           int64  `json:"collaborators"`
	TwoFactorAuthentication bool   `json:"two_factor_authentication"`
	Plan                    Plan   `json:"plan"`
}

type Plan struct {
	Name          string `json:"name"`
	Space         int64  `json:"space"`
	PrivateRepos  int64  `json:"private_repos"`
	Collaborators int64  `json:"collaborators"`
}

func (c *UserService) Current(ctx context.Context) (*User, error) {
	r, err := resty.New().R().
		SetAuthToken(c.token).
		SetError(githubError{}).
		Get("https://api.github.com/user")
	if err != nil {
		return nil, err
	}

	var usr *User
	err = json.Unmarshal(r.Body(), &usr)
	if err != nil {
		return nil, err
	}

	return usr, nil
}

type IssuesSearchResult struct {
	TotalCount        int     `json:"total_count"`
	IncompleteResults bool    `json:"incomplete_results"`
	Items             []*Item `json:"items"`
}

type Item struct {
	URL               string        `json:"url"`
	RepositoryURL     string        `json:"repository_url"`
	LabelsURL         string        `json:"labels_url"`
	CommentsURL       string        `json:"comments_url"`
	EventsURL         string        `json:"events_url"`
	HTMLURL           string        `json:"html_url"`
	ID                int64         `json:"id"`
	NodeID            string        `json:"node_id"`
	Number            int64         `json:"number"`
	Title             string        `json:"title"`
	User              User          `json:"user"`
	Labels            []interface{} `json:"labels"`
	State             string        `json:"state"`
	Locked            bool          `json:"locked"`
	Assignee          interface{}   `json:"assignee"`
	Assignees         []interface{} `json:"assignees"`
	Milestone         interface{}   `json:"milestone"`
	Comments          int64         `json:"comments"`
	CreatedAt         string        `json:"created_at"`
	UpdatedAt         string        `json:"updated_at"`
	ClosedAt          interface{}   `json:"closed_at"`
	AuthorAssociation string        `json:"author_association"`
	ActiveLockReason  interface{}   `json:"active_lock_reason"`
	Draft             bool          `json:"draft"`
	PullRequest       PullRequest   `json:"pull_request"`
	Body              string        `json:"body"`
	Score             float64       `json:"score"`
}

// type PullRequest struct {
// 	URL      string `json:"url"`
// 	HTMLURL  string `json:"html_url"`
// 	DiffURL  string `json:"diff_url"`
// 	PatchURL string `json:"patch_url"`
// }

func (c *SearchService) Issues(ctx context.Context, query string) (*IssuesSearchResult, error) {
	r, err := resty.New().R().
		SetAuthToken(c.token).
		SetError(githubError{}).
		SetQueryParam("q", query).
		Get("https://api.github.com/search/issues")
	if err != nil {
		return nil, err
	}

	var isr *IssuesSearchResult
	err = json.Unmarshal(r.Body(), &isr)
	if err != nil {
		return nil, err
	}

	return isr, err
}
