package tui

import "preq/internal/pkg/client"

type PullRequest struct {
	PullRequest              *client.PullRequest
	Selected                 bool
	Visible                  bool
	Client                   client.Client
	Repository               *client.Repository
	IsApprovalsLoading       bool
	IsCommentsLoading        bool
	IsChangesRequestsLoading bool
}

type RepositoryData struct {
	Name         string
	IsLoading    bool
	PullRequests map[string]*PullRequest
}

type tuiState struct {
	RepositoryData map[string]*RepositoryData
}

var state = &tuiState{}
