package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Rules []*Rule `yaml:"rules"`
}

type Rule struct {
	Path    string    `yaml:"path"`
	Actions []*Action `yaml:"actions"`
}

type Action struct {
	Type   string   `yaml:"type"`
	Keys   []string `yaml:"keys"`    // for remove_keys
	Key    string   `yaml:"key"`     // for rename_key
	NewKey string   `yaml:"new_key"` // for rename_key
}

func ReadConfigs(dir string) ([]*Config, error) {
	pattern := filepath.Join(dir, ".yamledit", "*.yaml")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("glob migration files: %w", err)
	}
	configs := make([]*Config, 0, len(matches))
	for _, p := range matches {
		cfg, err := readConfig(p)
		if err != nil {
			return nil, fmt.Errorf("read migration file %s: %w", p, err)
		}
		configs = append(configs, cfg)
	}
	return configs, nil
}

func readConfig(p string) (*Config, error) {
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal YAML: %w", err)
	}
	return &cfg, nil
}
