package tui

import (
	"encoding/json"
	"os"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog/log"
)

type repoInfo struct {
	Name        string    `json:"name"`
	Provider    string    `json:"provider"`
	LastVisited time.Time `json:"lastVisited"`
}

type state struct {
	Visited []*repoInfo `json:"visited,omitempty"`
}

type PersistanceRepo interface {
	AddVisited(name string, provider string)
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

func (repo *XDGPersistanceRepo) AddVisited(name string, provider string) {
	repo.load()

	found := false
	for _, v := range repo.s.Visited {
		if v.Name == name && v.Provider == provider {
			found = true
			break
		}
	}

	if found {
		return
	}

	repo.s.Visited = append(
		repo.s.Visited,
		&repoInfo{Name: name, Provider: provider, LastVisited: time.Now()},
	)
	repo.save()
}

var persistanceRepo PersistanceRepo = &XDGPersistanceRepo{
	s: &state{},
}
