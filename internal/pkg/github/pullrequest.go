package github

type PullRequest struct {
	URL                 string          `json:"url"`
	ID                  int64           `json:"id"`
	NodeID              string          `json:"node_id"`
	HTMLURL             string          `json:"html_url"`
	DiffURL             string          `json:"diff_url"`
	PatchURL            string          `json:"patch_url"`
	IssueURL            string          `json:"issue_url"`
	CommitsURL          string          `json:"commits_url"`
	ReviewCommentsURL   string          `json:"review_comments_url"`
	ReviewCommentURL    string          `json:"review_comment_url"`
	CommentsURL         string          `json:"comments_url"`
	StatusesURL         string          `json:"statuses_url"`
	Number              int64           `json:"number"`
	State               string          `json:"state"`
	Locked              bool            `json:"locked"`
	Title               string          `json:"title"`
	User                Assignee        `json:"user"`
	Body                string          `json:"body"`
	Labels              []Label         `json:"labels"`
	Milestone           Milestone       `json:"milestone"`
	ActiveLockReason    string          `json:"active_lock_reason"`
	CreatedAt           string          `json:"created_at"`
	UpdatedAt           string          `json:"updated_at"`
	ClosedAt            string          `json:"closed_at"`
	MergedAt            string          `json:"merged_at"`
	MergeCommitSHA      string          `json:"merge_commit_sha"`
	Assignee            Assignee        `json:"assignee"`
	Assignees           []Assignee      `json:"assignees"`
	RequestedReviewers  []Assignee      `json:"requested_reviewers"`
	RequestedTeams      []RequestedTeam `json:"requested_teams"`
	Head                Base            `json:"head"`
	Base                Base            `json:"base"`
	Links               Links           `json:"_links"`
	AuthorAssociation   string          `json:"author_association"`
	Draft               bool            `json:"draft"`
	Merged              bool            `json:"merged"`
	Mergeable           bool            `json:"mergeable"`
	Rebaseable          bool            `json:"rebaseable"`
	MergeableState      string          `json:"mergeable_state"`
	MergedBy            Assignee        `json:"merged_by"`
	Comments            int64           `json:"comments"`
	ReviewComments      int64           `json:"review_comments"`
	MaintainerCanModify bool            `json:"maintainer_can_modify"`
	Commits             int64           `json:"commits"`
	Additions           int64           `json:"additions"`
	Deletions           int64           `json:"deletions"`
	ChangedFiles        int64           `json:"changed_files"`
}

type Assignee struct {
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

type Base struct {
	Label string   `json:"label"`
	Ref   string   `json:"ref"`
	SHA   string   `json:"sha"`
	User  Assignee `json:"user"`
	Repo  Repo     `json:"repo"`
}

type Repo struct {
	ID                 int64       `json:"id"`
	NodeID             string      `json:"node_id"`
	Name               string      `json:"name"`
	FullName           string      `json:"full_name"`
	Owner              Assignee    `json:"owner"`
	Private            bool        `json:"private"`
	HTMLURL            string      `json:"html_url"`
	Description        string      `json:"description"`
	Fork               bool        `json:"fork"`
	URL                string      `json:"url"`
	ArchiveURL         string      `json:"archive_url"`
	AssigneesURL       string      `json:"assignees_url"`
	BlobsURL           string      `json:"blobs_url"`
	BranchesURL        string      `json:"branches_url"`
	CollaboratorsURL   string      `json:"collaborators_url"`
	CommentsURL        string      `json:"comments_url"`
	CommitsURL         string      `json:"commits_url"`
	CompareURL         string      `json:"compare_url"`
	ContentsURL        string      `json:"contents_url"`
	ContributorsURL    string      `json:"contributors_url"`
	DeploymentsURL     string      `json:"deployments_url"`
	DownloadsURL       string      `json:"downloads_url"`
	EventsURL          string      `json:"events_url"`
	ForksURL           string      `json:"forks_url"`
	GitCommitsURL      string      `json:"git_commits_url"`
	GitRefsURL         string      `json:"git_refs_url"`
	GitTagsURL         string      `json:"git_tags_url"`
	GitURL             string      `json:"git_url"`
	IssueCommentURL    string      `json:"issue_comment_url"`
	IssueEventsURL     string      `json:"issue_events_url"`
	IssuesURL          string      `json:"issues_url"`
	KeysURL            string      `json:"keys_url"`
	LabelsURL          string      `json:"labels_url"`
	LanguagesURL       string      `json:"languages_url"`
	MergesURL          string      `json:"merges_url"`
	MilestonesURL      string      `json:"milestones_url"`
	NotificationsURL   string      `json:"notifications_url"`
	PullsURL           string      `json:"pulls_url"`
	ReleasesURL        string      `json:"releases_url"`
	SSHURL             string      `json:"ssh_url"`
	StargazersURL      string      `json:"stargazers_url"`
	StatusesURL        string      `json:"statuses_url"`
	SubscribersURL     string      `json:"subscribers_url"`
	SubscriptionURL    string      `json:"subscription_url"`
	TagsURL            string      `json:"tags_url"`
	TeamsURL           string      `json:"teams_url"`
	TreesURL           string      `json:"trees_url"`
	CloneURL           string      `json:"clone_url"`
	MirrorURL          string      `json:"mirror_url"`
	HooksURL           string      `json:"hooks_url"`
	SvnURL             string      `json:"svn_url"`
	Homepage           string      `json:"homepage"`
	Language           interface{} `json:"language"`
	ForksCount         int64       `json:"forks_count"`
	StargazersCount    int64       `json:"stargazers_count"`
	WatchersCount      int64       `json:"watchers_count"`
	Size               int64       `json:"size"`
	DefaultBranch      string      `json:"default_branch"`
	OpenIssuesCount    int64       `json:"open_issues_count"`
	IsTemplate         bool        `json:"is_template"`
	Topics             []string    `json:"topics"`
	HasIssues          bool        `json:"has_issues"`
	HasProjects        bool        `json:"has_projects"`
	HasWiki            bool        `json:"has_wiki"`
	HasPages           bool        `json:"has_pages"`
	HasDownloads       bool        `json:"has_downloads"`
	Archived           bool        `json:"archived"`
	Disabled           bool        `json:"disabled"`
	Visibility         string      `json:"visibility"`
	PushedAt           string      `json:"pushed_at"`
	CreatedAt          string      `json:"created_at"`
	UpdatedAt          string      `json:"updated_at"`
	Permissions        Permissions `json:"permissions"`
	AllowRebaseMerge   bool        `json:"allow_rebase_merge"`
	TemplateRepository interface{} `json:"template_repository"`
	TempCloneToken     string      `json:"temp_clone_token"`
	AllowSquashMerge   bool        `json:"allow_squash_merge"`
	AllowMergeCommit   bool        `json:"allow_merge_commit"`
	SubscribersCount   int64       `json:"subscribers_count"`
	NetworkCount       int64       `json:"network_count"`
}

type Permissions struct {
	Admin bool `json:"admin"`
	Push  bool `json:"push"`
	Pull  bool `json:"pull"`
}

type Label struct {
	ID          int64  `json:"id"`
	NodeID      string `json:"node_id"`
	URL         string `json:"url"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Default     bool   `json:"default"`
}

type Links struct {
	Self           Comments `json:"self"`
	HTML           Comments `json:"html"`
	Issue          Comments `json:"issue"`
	Comments       Comments `json:"comments"`
	ReviewComments Comments `json:"review_comments"`
	ReviewComment  Comments `json:"review_comment"`
	Commits        Comments `json:"commits"`
	Statuses       Comments `json:"statuses"`
}

type Comments struct {
	Href string `json:"href"`
}

type Milestone struct {
	URL          string   `json:"url"`
	HTMLURL      string   `json:"html_url"`
	LabelsURL    string   `json:"labels_url"`
	ID           int64    `json:"id"`
	NodeID       string   `json:"node_id"`
	Number       int64    `json:"number"`
	State        string   `json:"state"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Creator      Assignee `json:"creator"`
	OpenIssues   int64    `json:"open_issues"`
	ClosedIssues int64    `json:"closed_issues"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
	ClosedAt     string   `json:"closed_at"`
	DueOn        string   `json:"due_on"`
}

type RequestedTeam struct {
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
