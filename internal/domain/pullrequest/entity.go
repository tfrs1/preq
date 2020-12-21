package pullrequest

import "time"

type State string

// type PullRequest struct {
// 	ID          string
// 	Title       string
// 	URL         string
// 	State       PullRequestState
// 	Source      string
// 	Destination string
// 	Created     time.Time
// 	Updated     time.Time
// }

type EntityID string

type Entity struct {
	ID          EntityID
	Title       string
	State       State
	Source      string
	Destination string
	CloseBranch bool
	Draft       bool
	URL         string
	Created     time.Time
	Updated     time.Time
}
