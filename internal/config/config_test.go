package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "config.yaml")
	data := []byte(`
llm:
  base_url_env: "BASE_URL"
  api_key_env: "API_KEY"
  model_env: "MODEL_NAME"
  env_file: ".env"
daemon:
  check_interval: "10m"
  grace_period: "30m"
priority:
  P0:
    max_gpus: 1
  P1:
    max_gpus: 2
  P2:
    max_gpus: 3
notify:
  wall: true
  log_path: "/tmp/vigilon-test/"
state:
  path: "/tmp/vigilon-test/state.json"
`)
	os.WriteFile(tmp, data, 0644)

	cfg, err := Load(tmp)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Daemon.CheckInterval != "10m" {
		t.Errorf("expected 10m, got %s", cfg.Daemon.CheckInterval)
	}
	if cfg.Priority["P2"].MaxGPUs != 3 {
		t.Errorf("expected P2 max_gpus=3, got %d", cfg.Priority["P2"].MaxGPUs)
	}
}

func TestValidate_MissingStatePath(t *testing.T) {
	cfg := &Config{Notify: NotifyConfig{LogPath: "/tmp"}, Daemon: DaemonConfig{IdleThreshold: 3, DurationBuffer: 1.5}}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for missing state.path")
	}
}

func TestValidate_InvalidDuration(t *testing.T) {
	cfg := &Config{
		State:  StateConfig{Path: "/tmp/state.json"},
		Notify: NotifyConfig{LogPath: "/tmp"},
		Daemon: DaemonConfig{CheckInterval: "not-a-duration", IdleThreshold: 3, DurationBuffer: 1.5},
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for invalid check_interval")
	}
}

func TestValidate_InvalidMaxGPUs(t *testing.T) {
	cfg := &Config{
		State:    StateConfig{Path: "/tmp/state.json"},
		Notify:   NotifyConfig{LogPath: "/tmp"},
		Daemon:   DaemonConfig{IdleThreshold: 3, DurationBuffer: 1.5},
		Priority: map[string]PriorityTier{"P0": {MaxGPUs: 0}},
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for max_gpus <= 0")
	}
}

func TestLoad_AppliesDefaults(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "config.yaml")
	os.WriteFile(tmp, []byte("state:\n  path: /tmp/s.json\nnotify:\n  log_path: /tmp\n"), 0644)
	cfg, err := Load(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Daemon.IdleThreshold != 3 {
		t.Errorf("expected idle_threshold=3, got %d", cfg.Daemon.IdleThreshold)
	}
	if cfg.Daemon.DurationBuffer != 1.5 {
		t.Errorf("expected duration_buffer=1.5, got %f", cfg.Daemon.DurationBuffer)
	}
}

