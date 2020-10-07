package pullrequest

import (
	"testing"
)

func Test_UpdateService_Update(t *testing.T) {
	t.Run("Updates a pull request without error", func(t *testing.T) {
		s := NewUpdateService(&MockPullRequestUpdater{})
		_, err := s.Update(&UpdateOptions{})
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("Updates a pull request and returns entity instance", func(t *testing.T) {
		s := NewUpdateService(&MockPullRequestUpdater{})
		pr, err := s.Update(&UpdateOptions{})

		if err != nil {
			t.Error(err)
		}

		if pr == nil {
			t.Error("Expected entity received nil")
		}
	})
}
