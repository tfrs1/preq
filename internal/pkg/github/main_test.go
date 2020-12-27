package github

import (
	"preq/internal/domain/pullrequest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// func TestASD(t *testing.T) {
// 	c := New()
// 	c.Get()
// 	// assert.IsType(t, client, c)
// }

func TestGithubPullRequestList(t *testing.T) {
	t.Run("has a constructor", func(t *testing.T) {
		list := NewPullRequestPageList(nil, nil)
		assert.Implements(t, (*pullrequest.EntityPageList)(nil), list)
	})

	t.Run("implements hasNext()", func(t *testing.T) {
		list := NewPullRequestPageList(nil, nil)
		assert.True(t, list.HasNext())
	})
}
