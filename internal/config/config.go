package config

import (
	"fmt"
	"os"
	"time"

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
	CheckInterval  string  `yaml:"check_interval"`
	GracePeriod    string  `yaml:"grace_period"`
	IdleThreshold  int     `yaml:"idle_threshold"`
	DurationBuffer float64 `yaml:"duration_buffer"`
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
	if cfg.Daemon.IdleThreshold <= 0 {
		cfg.Daemon.IdleThreshold = 3
	}
	if cfg.Daemon.DurationBuffer <= 0 {
		cfg.Daemon.DurationBuffer = 1.5
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.State.Path == "" {
		return fmt.Errorf("state.path is required")
	}
	if c.Notify.LogPath == "" {
		return fmt.Errorf("notify.log_path is required")
	}
	if c.Daemon.CheckInterval != "" {
		if _, err := time.ParseDuration(c.Daemon.CheckInterval); err != nil {
			return fmt.Errorf("invalid daemon.check_interval: %w", err)
		}
	}
	if c.Daemon.GracePeriod != "" {
		if _, err := time.ParseDuration(c.Daemon.GracePeriod); err != nil {
			return fmt.Errorf("invalid daemon.grace_period: %w", err)
		}
	}
	if c.Daemon.IdleThreshold < 0 {
		return fmt.Errorf("daemon.idle_threshold must be >= 0")
	}
	if c.Daemon.DurationBuffer < 0 {
		return fmt.Errorf("daemon.duration_buffer must be >= 0")
	}
	for name, tier := range c.Priority {
		if tier.MaxGPUs <= 0 {
			return fmt.Errorf("priority.%s.max_gpus must be > 0", name)
		}
	}
	return nil
}
