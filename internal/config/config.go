package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

// RestartPolicy controls when a process should be restarted.
type RestartPolicy string

const (
	RestartAlways    RestartPolicy = "always"
	RestartOnFailure RestartPolicy = "on-failure"
	RestartNever     RestartPolicy = "never"
)

// HealthCheck mirrors process.HealthCheck for config parsing.
type HealthCheck struct {
	URL      string        `toml:"url"`
	Interval time.Duration `toml:"interval"`
	Timeout  time.Duration `toml:"timeout"`
	Retries  int           `toml:"retries"`
}

// Process represents a single managed process configuration.
type Process struct {
	Name          string        `toml:"name"`
	Command       string        `toml:"command"`
	Args          []string      `toml:"args"`
	Dir           string        `toml:"dir"`
	Env           []string      `toml:"env"`
	RestartPolicy RestartPolicy `toml:"restart_policy"`
	MaxRetries    int           `toml:"max_retries"`
	HealthCheck   *HealthCheck  `toml:"health_check"`
}

// Config is the top-level configuration structure.
type Config struct {
	Processes []Process `toml:"process"`
}

// Load reads and parses a TOML config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if _, err := toml.Decode(string(data), &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func validate(cfg *Config) error {
	if len(cfg.Processes) == 0 {
		return errors.New("config must define at least one process")
	}

	seen := make(map[string]bool)
	for i, p := range cfg.Processes {
		if p.Name == "" {
			return fmt.Errorf("process[%d]: name is required", i)
		}
		if p.Command == "" {
			return fmt.Errorf("process %q: command is required", p.Name)
		}
		if seen[p.Name] {
			return fmt.Errorf("duplicate process name: %q", p.Name)
		}
		seen[p.Name] = true

		if p.RestartPolicy == "" {
			cfg.Processes[i].RestartPolicy = RestartOnFailure
		} else if err := validateRestartPolicy(p.RestartPolicy); err != nil {
			return fmt.Errorf("process %q: %w", p.Name, err)
		}

		if p.HealthCheck != nil {
			if err := validateHealthCheck(p.HealthCheck, p.Name); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateRestartPolicy(policy RestartPolicy) error {
	switch policy {
	case RestartAlways, RestartOnFailure, RestartNever:
		return nil
	default:
		return fmt.Errorf("invalid restart_policy %q (want always|on-failure|never)", policy)
	}
}

func validateHealthCheck(hc *HealthCheck, processName string) error {
	if hc.URL == "" {
		return fmt.Errorf("process %q: health_check.url is required", processName)
	}
	return nil
}
