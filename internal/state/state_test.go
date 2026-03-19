package state

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestSaveAndLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	s := &State{
		Users: map[string]*UserState{
			"hyf": {
				Priority: "P1",
				GPUs:     []int{0, 1},
				Expires:  timePtr(time.Now().Add(2 * time.Hour)),
			},
		},
	}
	if err := Save(path, s); err != nil {
		t.Fatalf("Save: %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	u := loaded.Users["hyf"]
	if u.Priority != "P1" || len(u.GPUs) != 2 {
		t.Errorf("unexpected state: %+v", u)
	}
}

func TestLoadMissing(t *testing.T) {
	s, err := Load("/nonexistent/state.json")
	if err != nil {
		t.Fatalf("should return default: %v", err)
	}
	if len(s.Users) != 0 {
		t.Errorf("expected empty users")
	}
}

func timePtr(t time.Time) *time.Time { return &t }

func TestConcurrentLoadAndModify(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	const n = 10
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		user := fmt.Sprintf("user%d", i)
		go func(u string) {
			defer wg.Done()
			err := LoadAndModify(path, func(s *State) error {
				s.Users[u] = &UserState{Priority: "P1"}
				return nil
			})
			if err != nil {
				t.Errorf("LoadAndModify(%s): %v", u, err)
			}
		}(user)
	}
	wg.Wait()

	s, err := Load(path)
	if err != nil {
		t.Fatalf("Load after concurrent writes: %v", err)
	}
	if len(s.Users) != n {
		t.Errorf("expected %d users, got %d", n, len(s.Users))
	}
	for i := 0; i < n; i++ {
		user := fmt.Sprintf("user%d", i)
		if _, ok := s.Users[user]; !ok {
			t.Errorf("missing user %s", user)
		}
	}
}
