package persistance

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/mitchellh/go-homedir"
	"golang.org/x/exp/slices"
)

type PersistanceRepoInfo struct {
	Name        string    `json:"name"`
	Provider    string    `json:"provider"`
	LastVisited time.Time `json:"lastVisited"`
	Path        string    `json:"path,omitempty"`
}

type state struct {
	Visited []*PersistanceRepoInfo `json:"visited,omitempty"`
}

type PersistanceRepo interface {
	AddVisited(name string, provider string, path string) error
	GetVisited() ([]*PersistanceRepoInfo, error)
}

type XDGPersistanceRepo struct {
	s *state
}

func (repo *XDGPersistanceRepo) createConfigDirIfNotExist() error {
	dirPath, err := homedir.Expand("~/.config/preq")
	if err != nil {
		return err
	}

	_, err = os.Stat(dirPath)
	if os.IsNotExist(err) {
		os.MkdirAll(dirPath, 0700)
	}

	return nil
}

func (repo *XDGPersistanceRepo) load() error {
	err := repo.createConfigDirIfNotExist()
	if err != nil {
		return err
	}

	path, _ := homedir.Expand("~/.config/preq/state")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, repo.s)
	if err != nil {
		return fmt.Errorf("cannot load state file: %v", err)
	}

	return nil
}

func (repo *XDGPersistanceRepo) save() error {
	err := repo.createConfigDirIfNotExist()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(repo.s, "", "  ")
	if err != nil {
		return err
	}

	path, err := homedir.Expand("~/.config/preq/state")
	if err != nil {
		return err
	}

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}

	return nil
}
func (repo *XDGPersistanceRepo) GetVisited() ([]*PersistanceRepoInfo, error) {
	err := repo.load()
	if err != nil {
		return nil, err
	}

	return repo.s.Visited, nil
}

func (repo *XDGPersistanceRepo) AddVisited(
	name string,
	provider string,
	path string,
) error {
	err := repo.load()
	if err != nil {
		return err
	}

	index := slices.IndexFunc(
		repo.s.Visited,
		func(v *PersistanceRepoInfo) bool {
			return v.Name == name && v.Provider == provider
		},
	)

	if index != -1 {
		slices.Replace(repo.s.Visited, index, index+1,
			&PersistanceRepoInfo{
				Name:        name,
				Provider:    provider,
				LastVisited: time.Now(),
				Path:        path,
			},
		)
	} else {
		repo.s.Visited = append(
			repo.s.Visited,
			&PersistanceRepoInfo{
				Name:        name,
				Provider:    provider,
				LastVisited: time.Now(),
				Path:        path,
			},
		)
	}

	err = repo.save()
	return err
}

var persistanceRepo PersistanceRepo = &XDGPersistanceRepo{
	s: &state{},
}

func GetRepo() PersistanceRepo {
	return persistanceRepo
}
