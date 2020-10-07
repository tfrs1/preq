package pullrequest

type Entity struct {
	Title       string
	Source      string
	Destination string
	CloseBranch bool
	Draft       bool
	URL         string
}
