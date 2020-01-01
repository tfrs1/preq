package configutil

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type mockFlagSet struct {
	value string
	err   error
}

func (m *mockFlagSet) GetString(f string) (string, error) {
	return m.value, m.err
}

func TestGetStringFlagOrDefault(t *testing.T) {
	// Returns flag value when defined
	v := GetStringFlagOrDefault(
		&mockFlagSet{"value", nil},
		"flag",
		"",
	)
	assert.Equal(t, "value", v)

	// Returns default value on error
	v = GetStringFlagOrDefault(
		&mockFlagSet{"", errors.New("error")},
		"flag",
		"default",
	)
	assert.Equal(t, "default", v)

	// Returns default value on empty string
	v = GetStringFlagOrDefault(
		&mockFlagSet{"", nil},
		"flag",
		"default",
	)
	assert.Equal(t, "default", v)
}
