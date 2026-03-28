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
	Type              string            `yaml:"type"`
	Keys              []string          `yaml:"keys"`                  // for remove_keys
	Key               string            `yaml:"key"`                   // for rename_key, set_key
	NewKey            string            `yaml:"new_key"`               // for rename_key
	WhenDuplicate     string            `yaml:"when_duplicate"`        // for rename_key
	Value             any               `yaml:"value"`                 // for set_key
	SkipIfKeyNotFound bool              `yaml:"skip_if_key_not_found"` // for set_key
	SkipIfKeyFound    bool              `yaml:"skip_if_key_found"`     // for set_key
	ClearComment      bool              `yaml:"clear_comment"`         // for set_key
	InsertAt          []*InsertLocation `yaml:"insert_at"`             // for set_key
}

type InsertLocation struct {
	AfterKey  string `yaml:"after_key"`
	BeforeKey string `yaml:"before_key"`
	First     bool   `yaml:"first"`
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
