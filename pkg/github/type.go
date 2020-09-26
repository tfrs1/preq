package github

type reviewRequests struct {
	Users []*user `json:"users"`
	Teams []*team `json:"teams"`
}

type team struct {
	ID              int64       `json:"id"`
	NodeID          string      `json:"node_id"`
	URL             string      `json:"url"`
	HTMLURL         string      `json:"html_url"`
	Name            string      `json:"name"`
	Slug            string      `json:"slug"`
	Description     string      `json:"description"`
	Privacy         string      `json:"privacy"`
	Permission      string      `json:"permission"`
	MembersURL      string      `json:"members_url"`
	RepositoriesURL string      `json:"repositories_url"`
	Parent          interface{} `json:"parent"`
}

type review struct {
	ID             int64  `json:"id"`
	NodeID         string `json:"node_id"`
	User           user   `json:"user"`
	Body           string `json:"body"`
	State          string `json:"state"`
	HTMLURL        string `json:"html_url"`
	PullRequestURL string `json:"pull_request_url"`
	Links          links  `json:"_links"`
	SubmittedAt    string `json:"submitted_at"`
	CommitID       string `json:"commit_id"`
}

type links struct {
	HTML        html `json:"html"`
	PullRequest html `json:"pull_request"`
}

type html struct {
	Href string `json:"href"`
}

type user struct {
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
	Name              string `json:"name"`
	// Company           interface{} `json:"company"`
	Blog string `json:"blog"`
	// Location          interface{} `json:"location"`
	// Email             interface{} `json:"email"`
	// Hireable          interface{} `json:"hireable"`
	// Bio               interface{} `json:"bio"`
	// TwitterUsername   interface{} `json:"twitter_username"`
	PublicRepos int64  `json:"public_repos"`
	PublicGists int64  `json:"public_gists"`
	Followers   int64  `json:"followers"`
	Following   int64  `json:"following"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type searchIssuesResult struct {
	TotalCount        int64               `json:"total_count"`
	IncompleteResults bool                `json:"incomplete_results"`
	Items             []*searchIssuesItem `json:"items"`
}

type searchIssuesItem struct {
	URL               string               `json:"url"`
	RepositoryURL     string               `json:"repository_url"`
	LabelsURL         string               `json:"labels_url"`
	CommentsURL       string               `json:"comments_url"`
	EventsURL         string               `json:"events_url"`
	HTMLURL           string               `json:"html_url"`
	ID                int64                `json:"id"`
	NodeID            string               `json:"node_id"`
	Number            int64                `json:"number"`
	Title             string               `json:"title"`
	User              searchIssuesItemUser `json:"user"`
	Labels            []interface{}        `json:"labels"`
	State             string               `json:"state"`
	Locked            bool                 `json:"locked"`
	Assignee          interface{}          `json:"assignee"`
	Assignees         []interface{}        `json:"assignees"`
	Milestone         interface{}          `json:"milestone"`
	Comments          int64                `json:"comments"`
	CreatedAt         string               `json:"created_at"`
	UpdatedAt         string               `json:"updated_at"`
	ClosedAt          interface{}          `json:"closed_at"`
	AuthorAssociation string               `json:"author_association"`
	ActiveLockReason  interface{}          `json:"active_lock_reason"`
	Draft             bool                 `json:"draft"`
	PullRequest       pullRequest          `json:"pull_request"`
	Body              string               `json:"body"`
	Score             float64              `json:"score"`
}

type pullRequest struct {
	URL      string `json:"url"`
	HTMLURL  string `json:"html_url"`
	DiffURL  string `json:"diff_url"`
	PatchURL string `json:"patch_url"`
}

type searchIssuesItemUser struct {
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
