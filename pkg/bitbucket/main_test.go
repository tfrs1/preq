package client

import (
	"testing"
)

func TestASD(t *testing.T) {
	c := New()
	c.GetPullRequests()
	// assert.IsType(t, client, c)
}
