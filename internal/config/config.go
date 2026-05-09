package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// RestartPolicy defines when a process should be restarted.
type RestartPolicy string

const (
	RestartAlways    RestartPolicy = "always"
	RestartOnFailure RestartPolicy = "on-failure"
	RestartNever     RestartPolicy = "never"
)

// ProcessConfig holds the configuration for a single supervised process.
type ProcessConfig struct {
	Name          string        `yaml:"name"`
	Command       string        `yaml:"command"`
	Args          []string      `yaml:"args"`
	RestartPolicy RestartPolicy `yaml:"restart_policy"`
	MaxRestarts   int           `yaml:"max_restarts"`
	Backoff       time.Duration `yaml:"backoff"`
	Env           []string      `yaml:"env"`
	WorkDir       string        `yaml:"work_dir"`
}

// Config is the top-level procwatch configuration.
type Config struct {
	LogFormat  string          `yaml:"log_format"`
	Processes  []ProcessConfig `yaml:"processes"`
}

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// validate checks that the config is semantically valid.
func (c *Config) validate() error {
	if c.LogFormat == "" {
		c.LogFormat = "json"
	}
	if c.LogFormat != "json" && c.LogFormat != "text" {
		return fmt.Errorf("log_format must be \"json\" or \"text\", got %q", c.LogFormat)
	}

	names := make(map[string]bool)
	for i, p := range c.Processes {
		if p.Name == "" {
			return fmt.Errorf("process[%d]: name is required", i)
		}
		if p.Command == "" {
			return fmt.Errorf("process %q: command is required", p.Name)
		}
		if names[p.Name] {
			return fmt.Errorf("duplicate process name %q", p.Name)
		}
		names[p.Name] = true

		if c.Processes[i].RestartPolicy == "" {
			c.Processes[i].RestartPolicy = RestartOnFailure
		}
		if c.Processes[i].MaxRestarts == 0 {
			c.Processes[i].MaxRestarts = 5
		}
		if c.Processes[i].Backoff == 0 {
			c.Processes[i].Backoff = 2 * time.Second
		}
	}
	return nil
}
