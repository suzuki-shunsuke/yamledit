package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Rules []*Rule `json:"rules" yaml:"rules" jsonschema_description:"List of migration rules. Rules are evaluated in order."`
}

type Rule struct {
	Path    string    `json:"path" yaml:"path" jsonschema_description:"YAML path to the target node (e.g. \"$\", \"$.foo\")"`
	Actions []*Action `json:"actions" yaml:"actions" jsonschema_description:"List of actions to apply. Actions are evaluated in order."`
}

type Action struct {
	Type              string            `json:"type" yaml:"type" jsonschema_description:"Action type: remove_keys, rename_key, set_key, or add_values"`
	Keys              []string          `json:"keys,omitempty" yaml:"keys" jsonschema_description:"Keys to remove (for remove_keys)"`
	Key               string            `json:"key,omitempty" yaml:"key" jsonschema_description:"Target key name (for rename_key, set_key)"`
	NewKey            string            `json:"new_key,omitempty" yaml:"new_key" jsonschema_description:"New key name (for rename_key)"`
	WhenDuplicate     string            `json:"when_duplicate,omitempty" yaml:"when_duplicate" jsonschema_description:"Behavior when new_key already exists: skip (default), ignore_existing_key, remove_old_key, fail (for rename_key)"`
	Value             any               `json:"value,omitempty" yaml:"value" jsonschema_description:"Value to set (for set_key)"`
	SkipIfKeyNotFound bool              `json:"skip_if_key_not_found,omitempty" yaml:"skip_if_key_not_found" jsonschema_description:"If true, do nothing when the key does not exist (for set_key)"`
	SkipIfKeyFound    bool              `json:"skip_if_key_found,omitempty" yaml:"skip_if_key_found" jsonschema_description:"If true, do nothing when the key already exists (for set_key)"`
	ClearComment      bool              `json:"clear_comment,omitempty" yaml:"clear_comment" jsonschema_description:"If true, remove the comment on the existing key (for set_key)"`
	InsertAt          []*InsertLocation `json:"insert_at,omitempty" yaml:"insert_at" jsonschema_description:"Where to insert a new key. The first matching condition is used. If none match, the key is appended at the end (for set_key)"`
	Values            []any             `json:"values,omitempty" yaml:"values" jsonschema_description:"Values to add to the list (for add_values)"`
	Index             *int              `json:"index,omitempty" yaml:"index" jsonschema_description:"Index to insert values at. 0 for beginning, negative values count from end. Default -1 (for add_values)"`
	Expr              string            `json:"expr,omitempty" yaml:"expr" jsonschema_description:"Expression for sorting keys. Variables a and b represent key-value pairs with fields: key, value, comment, index (for sort_key)"`
}

type InsertLocation struct {
	AfterKey  string `json:"after_key,omitempty" yaml:"after_key" jsonschema_description:"Insert after this key"`
	BeforeKey string `json:"before_key,omitempty" yaml:"before_key" jsonschema_description:"Insert before this key"`
	First     bool   `json:"first,omitempty" yaml:"first" jsonschema_description:"Insert at the beginning"`
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
