package configutil

import (
	"io"
	"prctl/internal/fs"
	"prctl/mocks"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type mockConfigMerger struct {
	err error
}

func (m *mockConfigMerger) MergeConfig(in io.Reader) error {
	return m.err
}

type mockFlagSet struct {
	value     string
	boolValue bool
	err       error
}

func (m *mockFlagSet) GetString(f string) (string, error) {
	return m.value, m.err
}

func (m *mockFlagSet) GetBool(f string) (bool, error) {
	return m.boolValue, m.err
}

func Test_mergeConfig(t *testing.T) {
	t.Run("returns nil when merge succeeds", func(t *testing.T) {
		err := mergeConfig(nil, &mockConfigMerger{nil})
		assert.Equal(t, nil, err)
	})

	t.Run("returns error when merge fails", func(t *testing.T) {
		vErr := errors.New("mergeFailed")
		err := mergeConfig(nil, &mockConfigMerger{vErr})
		assert.EqualError(t, err, vErr.Error())
	})
}

func Test_fileExists(t *testing.T) {
	t.Run("returns nil if file exists", func(t *testing.T) {
		err := fileExists("", mocks.FS{mocks.FileInfo{false}, nil})
		assert.Equal(t, nil, err)
	})

	t.Run("returns error if file does not exists", func(t *testing.T) {
		vErr := errors.New("file does not exist")
		err := fileExists("", mocks.FS{mocks.FileInfo{}, vErr})
		assert.EqualError(t, err, vErr.Error())
	})

	t.Run("returns error if file is a directory", func(t *testing.T) {
		err := fileExists("", mocks.FS{mocks.FileInfo{true}, nil})
		assert.EqualError(t, err, ErrConfigFileIsDir.Error())
	})
}

func Test_loadFile(t *testing.T) {
	oldFileExists := fileExists

	t.Run("fails if file does not exist", func(t *testing.T) {
		vErr := errors.New("file err")
		fileExists = func(string, fs.Filesystem) error { return vErr }
		_, err := loadFile("", nil)
		assert.EqualError(t, err, vErr.Error())
	})

	t.Run("fails if file cannot be opened", func(t *testing.T) {
		vErr := errors.New("file err")
		fileExists = func(string, fs.Filesystem) error { return nil }
		_, err := loadFile("", mocks.FS{mocks.FileInfo{}, vErr})
		assert.EqualError(t, err, vErr.Error())
	})

	t.Run("succeeds if file exists and can be opened", func(t *testing.T) {
		fileExists = func(string, fs.Filesystem) error { return nil }
		_, err := loadFile("", mocks.FS{})
		assert.Equal(t, nil, err)
	})

	fileExists = oldFileExists
}

func Test_loadConfig(t *testing.T) {
	oldLoadFile := loadFile
	oldMergeConfig := mergeConfig

	t.Run("succeeds when file is loaded and merged", func(t *testing.T) {
		loadFile = func(string, fs.Filesystem) (io.Reader, error) { return nil, nil }
		mergeConfig = func(io.Reader, configMerger) error { return nil }
		err := loadConfig("")
		assert.Equal(t, nil, err)
	})

	t.Run("doesn't throw error for missing files", func(t *testing.T) {
		vErr := errors.New("load err")
		loadFile = func(string, fs.Filesystem) (io.Reader, error) { return nil, vErr }
		err := loadConfig("")
		assert.Equal(t, nil, err)
	})

	t.Run("fails when merge fails", func(t *testing.T) {
		vErr := errors.New("load err")
		loadFile = func(string, fs.Filesystem) (io.Reader, error) { return nil, nil }
		mergeConfig = func(io.Reader, configMerger) error { return vErr }
		err := loadConfig("")
		assert.EqualError(t, err, vErr.Error())
	})

	loadFile = oldLoadFile
	mergeConfig = oldMergeConfig
}

func TestLoad(t *testing.T) {
	oldLoadConfig := loadConfig
	oldGetGlobalConfigPath := getGlobalConfigPath

	t.Run("", func(t *testing.T) {
		loadConfig = func(string) error { return nil }
		getGlobalConfigPath = func() (string, error) { return "", nil }
		err := Load()
		assert.Equal(t, nil, err)
	})

	t.Run("", func(t *testing.T) {
		getGlobalConfigPath = func() (string, error) { return "", errors.New("") }
		err := Load()
		assert.EqualError(t, err, ErrHomeDirNotFound.Error())
	})

	t.Run("", func(t *testing.T) {
		vErr := errors.New("load err")
		loadConfig = func(string) error { return vErr }
		getGlobalConfigPath = func() (string, error) { return "", nil }
		err := Load()
		assert.EqualError(t, err, vErr.Error())
	})

	loadConfig = oldLoadConfig
	getGlobalConfigPath = oldGetGlobalConfigPath
}

func TestGetStringFlagOrDefault(t *testing.T) {
	t.Run("returns flag value when defined", func(t *testing.T) {
		v := GetStringFlagOrDefault(
			&mockFlagSet{value: "value", err: nil},
			"flag",
			"",
		)
		assert.Equal(t, "value", v)
	})

	t.Run("returns default value on error", func(t *testing.T) {
		v := GetStringFlagOrDefault(
			&mockFlagSet{value: "", err: errors.New("error")},
			"flag",
			"default",
		)
		assert.Equal(t, "default", v)
	})

	t.Run("returns default value on empty string", func(t *testing.T) {
		v := GetStringFlagOrDefault(
			&mockFlagSet{value: "", err: nil},
			"flag",
			"default",
		)
		assert.Equal(t, "default", v)
	})
}

func TestGetBoolFlagOrDefault(t *testing.T) {
	t.Run("returns flag value when defined", func(t *testing.T) {
		v := GetBoolFlagOrDefault(
			&mockFlagSet{boolValue: false, err: nil},
			"flag",
			true,
		)
		assert.Equal(t, false, v)
	})

	t.Run("returns default value on error", func(t *testing.T) {
		v := GetBoolFlagOrDefault(
			&mockFlagSet{boolValue: true, err: errors.New("error")},
			"flag",
			false,
		)
		assert.Equal(t, false, v)
	})
}
