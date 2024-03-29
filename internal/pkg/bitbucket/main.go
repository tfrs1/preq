package bitbucket

import (
	"encoding/json"
	"errors"
	"fmt"
	"preq/internal/pkg/client"
	"regexp"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
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
	Username   string
	Password   string
	Uuid       string
	Repository string
}

func NewClient(options *ClientOptions) (client.Client, error) {
	return &BitbucketCloudClient{
		username:   options.Username,
		password:   options.Password,
		uuid:       options.Uuid,
		repository: options.Repository,
	}, nil
}

type clientConfiguration struct {
	username   string
	password   string
	uuid       string
	repository string
}

type GetPullRequestActivityOptions struct {
	ID         string
	Repository *client.Repository
}

type PullRequestActivityApprovalEvent struct {
	Created time.Time
	User    string
}

type PullRequestActivityChangesRequestEvent struct {
	Created time.Time
	User    string
}

type PullRequestActivityCommentEvent struct {
	ID               string
	ParentID         string
	User             string
	Deleted          bool
	Content          string
	Created          time.Time
	Updated          time.Time
	BeforeLineNumber uint32
	AfterLineNumber  uint32
	// Check which name it when the file is renamed
	FilePath string
}

type PullRequestActivityLists struct {
	Approvals       *[]*PullRequestActivityApprovalEvent
	ChangesRequests *[]*PullRequestActivityChangesRequestEvent
}

func (c *BitbucketCloudClient) FillMiscInfoAsync(
	repo *client.Repository,
	pr *client.PullRequest,
) error {
	url := fmt.Sprintf(
		"https://api.bitbucket.org/2.0/repositories/%s/pullrequests/%s",
		repo.Name,
		pr.ID,
	)

	rc := resty.New()
	r, err := rc.R().
		SetBasicAuth(c.username, c.password).
		SetHeader("Content-Type", "application/json").
		SetError(bbError{}).
		Get(url)
	if err != nil {
		return err
	}
	if r.IsError() {
		return errors.New(string(r.Body()))
	}

	parsed := gjson.ParseBytes(r.Body())
	parsed.Get("participants").ForEach(func(key, value gjson.Result) bool {
		role := value.Get("role").String()
		if role == "REVIEWER" {
		} else if role == "PARTICIPANT" {
			// Role "PARTICIPANT" doesn't really count? When a user
			// approves his/her own pull request they end up in the category
			// of a "PARTICIPANT" instead of a "REVIEWER"
			return true
		}
		state := value.Get("state").String()
		if state == "approved" {
			pr.Approvals = append(pr.Approvals, &client.PullRequestApproval{
				Created: value.Get("participated_on").Time(),
				User:    value.Get("user.display_name").String(),
			})
		} else if state == "changes_requested" {
			pr.ChangesRequests = append(pr.ChangesRequests, &client.PullRequestChangesRequest{
				Created: value.Get("participated_on").Time(),
				User:    value.Get("user.display_name").String(),
			})
		}

		// FIXME: Does state == "approved" and aprroved == true mean the same thing?
		_ = value.Get("approved").Bool()
		return true
	})

	return nil
}

func buildCommentBody(options *client.CreateCommentOptions) string {
	extra := ""
	if options.ParentRef != nil {
		extra = fmt.Sprintf(`"parent": { "id": %s }`, options.ParentRef.ID)
	} else if options.LineRef != nil {
		t := "to"
		if options.LineRef.Type == client.OriginalLineNumber {
			t = "from"
		}
		extra = fmt.Sprintf(`"inline": {"%s": %d, "path": "%s"}`, t, options.LineRef.LineNumber, options.FilePath)
	}

	return fmt.Sprintf(
		`{ "content": { "raw": "%s" }, %s }`,
		options.Content,
		extra,
	)
}

func (c *BitbucketCloudClient) CreateComment(
	options *client.CreateCommentOptions,
) (*client.PullRequestComment, error) {
	url := fmt.Sprintf(
		"https://api.bitbucket.org/2.0/repositories/%s/pullrequests/%s/comments",
		options.Repository.Name,
		options.ID,
	)

	rc := resty.New()
	r, err := rc.R().
		SetBasicAuth(c.username, c.password).
		SetHeader("Content-Type", "application/json").
		SetBody(buildCommentBody(options)).
		SetError(bbError{}).
		Post(url)
	if err != nil {
		return nil, err
	}
	if r.IsError() {
		return nil, errors.New(string(r.Body()))
	}

	parsed := gjson.ParseBytes(r.Body())
	// FIXME: This is duplicated in GetComments()
	return &client.PullRequestComment{
		ID:       parsed.Get("id").String(),
		ParentID: parsed.Get("parent.id").String(),
		Deleted:  parsed.Get("deleted").Bool(),
		// TODO: Outdated: value.Get("comment.inline.outdated").Bool(),
		Content:          parsed.Get("content.raw").String(),
		Created:          parsed.Get("created_on").Time(),
		Updated:          parsed.Get("updated_on").Time(),
		User:             parsed.Get("user.display_name").String(),
		BeforeLineNumber: uint(parsed.Get("inline.from").Uint()),
		AfterLineNumber:  uint(parsed.Get("inline.to").Uint()),
		// Check which name it when the file is renamed
		FilePath:       parsed.Get("inline.path").String(),
		IsBeingStored:  false,
		IsBeingDeleted: false,
	}, nil
}

func (c *BitbucketCloudClient) DeleteComment(
	options *client.DeleteCommentOptions,
) error {
	url := fmt.Sprintf(
		"https://api.bitbucket.org/2.0/repositories/%s/pullrequests/%s/comments/%s",
		options.Repository.Name,
		options.ID,
		options.CommentID,
	)

	rc := resty.New()
	r, err := rc.R().
		SetBasicAuth(c.username, c.password).
		SetHeader("Content-Type", "application/json").
		SetError(bbError{}).
		Delete(url)
	if err != nil {
		return err
	}
	if r.IsError() {
		return errors.New(string(r.Body()))
	}

	return nil
}

func (c *BitbucketCloudClient) GetComments(
	options *client.GetCommentsOptions,
) ([]*client.PullRequestComment, error) {
	iter := newBitbucketIterator(
		&newBitbucketIteratorOptions[*client.PullRequestComment]{
			Client: c,
			RequestURL: fmt.Sprintf(
				"https://api.bitbucket.org/2.0/repositories/%s/pullrequests/%s/comments",
				options.Repository.Name,
				options.ID,
			),
			Parse: func(key, value gjson.Result) (*client.PullRequestComment, error) {
				var typ client.CommentType = client.CommentTypeInline

				if value.Get("parent").Exists() {
					typ = client.CommentTypeReply
				} else if !value.Get("inline").Exists() {
					typ = client.CommentTypeGlobal
				} else {
					from := value.Get("inline.from").Value()
					to := value.Get("inline.to").Value()
					if from == nil && to == nil {
						typ = client.CommentTypeFile
					}
				}

				// "links.code.href" is in
				// "https://api.bitbucket.org/2.0/repositories/{workspace}/{repo}/diff/{workspace}/{repo}:{sourceHash}..{destHash}?path={filename}"
				// format. We want to extract `sourceHash` as that is the commit the comment has been
				// written on. This value can used to determine whether a comment is outdated or not.
				commitHash := ""
				codeHref := value.Get("links.code.href").String()
				if codeHref != "" {
					matches := regexp.
						MustCompile(`.*?:([a-fA-F0-9]+)\.\..*`).
						FindStringSubmatch(
							value.Get("links.code.href").String(),
						)

					if len(matches) != 2 {
						return nil, errors.New("unable to the comments commit hash location")
					}

					commitHash = matches[1]
				}

				return &client.PullRequestComment{
					ID:               value.Get("id").String(),
					Type:             typ,
					ParentID:         value.Get("parent.id").String(),
					Deleted:          value.Get("deleted").Bool(),
					Content:          value.Get("content.raw").String(),
					Created:          value.Get("created_on").Time(),
					Updated:          value.Get("updated_on").Time(),
					User:             value.Get("user.display_name").String(),
					BeforeLineNumber: uint(value.Get("inline.from").Uint()),
					AfterLineNumber:  uint(value.Get("inline.to").Uint()),
					// Check which name it when the file is renamed
					FilePath:       value.Get("inline.path").String(),
					CommitHash:     commitHash,
					IsBeingStored:  false,
					IsBeingDeleted: false,
				}, nil
			},
		},
	)

	return iter.GetAll()
}

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
			Description:  value.Get("description").String(),
			ID:           value.Get("id").String(),
			CommentCount: int(value.Get("comment_count").Float()),
			Title:        value.Get("title").String(),
			User:         value.Get("author.nickname").String(),
			URL:          value.Get("links.html.href").String(),
			State:        client.PullRequestState(value.Get("state").String()),
			Source: client.PullRequestBranch{
				Name: value.Get("source.branch.name").String(),
				Hash: value.Get("source.commit.hash").String(),
			},
			Destination: client.PullRequestBranch{
				Name: value.Get("destination.branch.name").String(),
				Hash: value.Get("destination.commit.hash").String(),
			},
			Created: value.Get("created_on").Time(),
			Updated: value.Get("updated_on").Time(),
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
		ID:    fmt.Sprint(pr.ID),
		Title: pr.Title,
		URL:   pr.Links.HTML.Href,
		State: pr.State,
		Source: client.PullRequestBranch{
			Name: pr.Source.Branch.Name,
			Hash: pr.Source.Commit.Hash,
		},
		Destination: client.PullRequestBranch{
			Name: pr.Destination.Branch.Name,
			Hash: pr.Destination.Commit.Hash,
		},
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

func (c *BitbucketCloudClient) delete(url string) (*resty.Response, error) {
	rc := resty.New()
	r, err := rc.R().
		SetBasicAuth(c.username, c.password).
		SetError(bbError{}).
		Delete(url)
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

func (c *BitbucketCloudClient) Unapprove(
	o *client.UnapproveOptions,
) (*client.PullRequest, error) {
	url := fmt.Sprintf(
		"https://api.bitbucket.org/2.0/repositories/%s/pullrequests/%s/approve",
		o.Repository.Name,
		o.ID,
	)

	_, err := c.delete(url)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (c *BitbucketCloudClient) Approve(
	o *client.ApproveOptions,
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
	o *client.ApproveOptions,
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

func (c *BitbucketCloudClient) GetDefaultReviewer(
	username string,
) (*client.User, error) {
	panic("not implemented")
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
			o.Repository.Name,
		))
	if err != nil {
		return nil, err
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
			return nil, fmt.Errorf(errorMessage)
		}
		return nil, fmt.Errorf(string(r.Body()))
	}
	pr := &bitbucketPullRequest{}
	err = json.Unmarshal(r.Body(), pr)
	if err != nil {
		return nil, err
	}

	return &client.PullRequest{
		ID:    fmt.Sprint(pr.ID),
		Title: pr.Title,
		URL:   pr.Links.HTML.Href,
		State: pr.State,
		Source: client.PullRequestBranch{
			Name: pr.Source.Branch.Name,
			Hash: pr.Source.Commit.Hash,
		},
		Destination: client.PullRequestBranch{
			Name: pr.Destination.Branch.Name,
			Hash: pr.Destination.Commit.Hash,
		},
	}, nil
}
