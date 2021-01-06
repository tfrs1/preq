package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Table(t *testing.T) {
	t.Run("Can return seleted wows", func(t *testing.T) {
		table := newPullRequestTable()
		table.AddRow(&pullRequestTableItem{})
		table.AddRow(&pullRequestTableItem{
			Selected: true,
		})
		rows := table.SelectedItems()

		assert.Equal(t, len(rows), 1)
	})
}
