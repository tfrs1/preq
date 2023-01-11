package persistance

import (
	"encoding/json"
	"os"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog/log"
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
	AddVisited(name string, provider string, path string)
	GetVisited() []*PersistanceRepoInfo
}

type XDGPersistanceRepo struct {
	s *state
}

func (repo *XDGPersistanceRepo) load() {
	path, _ := homedir.Expand("~/.config/preq/state")
	data, _ := os.ReadFile(path)

	err := json.Unmarshal(data, repo.s)
	if err != nil {
		log.Error().Msg("cannot load state file")
		log.Error().Msg(err.Error())
	}
}

func (repo *XDGPersistanceRepo) save() {
	data, err := json.MarshalIndent(repo.s, "", "  ")
	if err != nil {
		log.Error().Msg("cannot marshal state data:" + err.Error())
	}

	path, err := homedir.Expand("~/.config/preq/state")
	if err != nil {
		log.Error().Msg("cannot find state file location:" + err.Error())
	}

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		log.Error().Msg("cannot save the state file: " + err.Error())
	}
}
func (repo *XDGPersistanceRepo) GetVisited() []*PersistanceRepoInfo {
	repo.load()
	return repo.s.Visited
}

func (repo *XDGPersistanceRepo) AddVisited(
	name string,
	provider string,
	path string,
) {
	repo.load()

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

	repo.save()
}

var persistanceRepo PersistanceRepo = &XDGPersistanceRepo{
	s: &state{},
}

func GetRepo() PersistanceRepo {
	return persistanceRepo
}
