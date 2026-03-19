package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

type State struct {
	Users map[string]*UserState `json:"users"`
}

type UserState struct {
	Priority string     `json:"priority"`
	GPUs     []int      `json:"gpus"`
	Expires  *time.Time `json:"expires"`
}

func Load(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &State{Users: make(map[string]*UserState)}, nil
		}
		return nil, fmt.Errorf("read state: %w", err)
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return &State{Users: make(map[string]*UserState)}, nil
	}
	if s.Users == nil {
		s.Users = make(map[string]*UserState)
	}
	return &s, nil
}

func Save(path string, s *State) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func (s *State) GetPriority(user string) string {
	u, ok := s.Users[user]
	if !ok {
		return "P0"
	}
	if u.Expires != nil && time.Now().After(*u.Expires) {
		u.Priority = "P0"
		u.GPUs = nil
		u.Expires = nil
	}
	return u.Priority
}
