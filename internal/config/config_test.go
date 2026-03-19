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
