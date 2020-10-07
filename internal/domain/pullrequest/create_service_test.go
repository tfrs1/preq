package pullrequest

import (
	"testing"
)

func Test_CreateService_Create(t *testing.T) {
	t.Run("Creates a pull request", func(t *testing.T) {
		s := NewCreateService(&MockPullRequestCreator{})
		_, err := s.Create(&CreateOptions{})
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("Returns a pull request entity", func(t *testing.T) {
		s := NewCreateService(&MockPullRequestCreator{})
		m, err := s.Create(&CreateOptions{})
		if err != nil {
			t.Error(err)
		}

		if m == nil {
			t.Error("Expected an entity and received nil")
		}
	})
}
