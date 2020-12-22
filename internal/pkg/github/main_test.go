package github

import "testing"

// func TestASD(t *testing.T) {
// 	c := New()
// 	c.Get()
// 	// assert.IsType(t, client, c)
// }

func TestGithubPullRequestList(t *testing.T) {
	t.Run("implements hasNext()", func(t *testing.T) {
		list := GithubPullRequestPageList{}
		list.hasNext()
	})
}
