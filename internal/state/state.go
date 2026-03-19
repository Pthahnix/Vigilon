package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"
)

type State struct {
	Users map[string]*UserState `json:"users"`
}

type UserState struct {
	Priority  string     `json:"priority"`
	GPUs      []int      `json:"gpus"`
	Expires   *time.Time `json:"expires"`
	IdleCount int        `json:"idle_count"` // consecutive idle detection cycles
}

func withLock(statePath string, fn func() error) error {
	lockPath := statePath + ".lock"
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("open lock file: %w", err)
	}
	defer f.Close()
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("acquire lock: %w", err)
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	return fn()
}

func Load(path string) (*State, error) {
	// Fast path: if the state file doesn't exist we can return the default
	// without acquiring a lock (the lock file's parent dir may not exist either).
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return &State{Users: make(map[string]*UserState)}, nil
	}
	var result *State
	err := withLock(path, func() error {
		data, err := os.ReadFile(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				result = &State{Users: make(map[string]*UserState)}
				return nil
			}
			return fmt.Errorf("read state: %w", err)
		}
		var s State
		if err := json.Unmarshal(data, &s); err != nil {
			result = &State{Users: make(map[string]*UserState)}
			return nil
		}
		if s.Users == nil {
			s.Users = make(map[string]*UserState)
		}
		result = &s
		return nil
	})
	return result, err
}

func Save(path string, s *State) error {
	return withLock(path, func() error {
		data, err := json.MarshalIndent(s, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal state: %w", err)
		}
		return os.WriteFile(path, data, 0644)
	})
}

func LoadAndModify(path string, fn func(s *State) error) error {
	return withLock(path, func() error {
		data, err := os.ReadFile(path)
		var s State
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("read state: %w", err)
			}
			s = State{Users: make(map[string]*UserState)}
		} else {
			if err := json.Unmarshal(data, &s); err != nil {
				s = State{Users: make(map[string]*UserState)}
			}
			if s.Users == nil {
				s.Users = make(map[string]*UserState)
			}
		}
		if err := fn(&s); err != nil {
			return err
		}
		out, err := json.MarshalIndent(&s, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal state: %w", err)
		}
		return os.WriteFile(path, out, 0644)
	})
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
