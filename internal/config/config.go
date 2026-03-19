package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	LLM      LLMConfig               `yaml:"llm"`
	Daemon   DaemonConfig            `yaml:"daemon"`
	Priority map[string]PriorityTier `yaml:"priority"`
	Notify   NotifyConfig            `yaml:"notify"`
	State    StateConfig             `yaml:"state"`
}

type LLMConfig struct {
	BaseURLEnv string `yaml:"base_url_env"`
	APIKeyEnv  string `yaml:"api_key_env"`
	ModelEnv   string `yaml:"model_env"`
	EnvFile    string `yaml:"env_file"`
}

type DaemonConfig struct {
	CheckInterval string `yaml:"check_interval"`
	GracePeriod   string `yaml:"grace_period"`
}

type PriorityTier struct {
	MaxGPUs int `yaml:"max_gpus"`
}

type NotifyConfig struct {
	Wall    bool   `yaml:"wall"`
	LogPath string `yaml:"log_path"`
}

type StateConfig struct {
	Path string `yaml:"path"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}
