package config

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	"github.com/goccy/go-yaml"
	"github.com/suzuki-shunsuke/yamledit/pkg/cache"
	gh "github.com/suzuki-shunsuke/yamledit/pkg/github"
)

// ReusableRule represents a reusable rule entry in the project config.
type ReusableRule struct {
	Name   string `json:"name" yaml:"name"`
	Import string `json:"import" yaml:"import"`
}

// Config represents the structure of .yamledit/config.yaml.
type Config struct {
	ReusableRules []ReusableRule `json:"reusable_rules" yaml:"reusable_rules"`
}

// FindReusableRule looks up a reusable rule by name and returns its import URL.
func (c *Config) FindReusableRule(name string) (string, bool) {
	for _, r := range c.ReusableRules {
		if r.Name == name {
			return r.Import, true
		}
	}
	return "", false
}

// UnmarshalConfig parses YAML content into a Config.
func UnmarshalConfig(b []byte, cfg *Config) error {
	if err := yaml.Unmarshal(b, cfg); err != nil {
		return fmt.Errorf("unmarshal project config: %w", err)
	}
	return nil
}

// ReadConfig reads .yamledit/config.yaml and returns the Config.
// Returns an empty Config if the file does not exist.
func ReadConfig(dir string) (*Config, error) {
	p := filepath.Join(dir, ".yamledit", "config.yaml")
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("read project config: %w", err)
	}
	var cfg Config
	if err := UnmarshalConfig(b, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// ReadGlobalConfig reads the global config and returns the Config.
// Returns an empty Config if no global config is found.
func ReadGlobalConfig() (*Config, error) {
	p := GlobalConfigPath()
	if p == "" {
		return &Config{}, nil
	}
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("read global config: %w", err)
	}
	var cfg Config
	if err := UnmarshalConfig(b, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// GlobalConfigPath returns the path to the global config file.
// Returns "" if no suitable location is found.
func GlobalConfigPath() string {
	if v := os.Getenv("YAMLEDIT_GLOBAL_CONFIG"); v != "" {
		return v
	}
	if v := os.Getenv("XDG_CONFIG_HOME"); v != "" {
		return filepath.Join(v, "yamledit", "config.yaml")
	}
	if runtime.GOOS == "windows" {
		if v := os.Getenv("LOCALAPPDATA"); v != "" {
			return filepath.Join(v, "yamledit", "config.yaml")
		}
		return ""
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "yamledit", "config.yaml")
}

// DownloadAndCache downloads a remote migration and caches it.
func DownloadAndCache(ctx context.Context, logger *slog.Logger, ghClient *gh.Client, c *cache.Cache, migration string) error {
	if owner, repo, path, ref, ok := parseGitHubImport(migration); ok {
		return downloadAndCacheGitHub(ctx, logger, ghClient, c, owner, repo, path, ref)
	}
	return downloadAndCacheURL(ctx, logger, c, migration)
}

func downloadAndCacheGitHub(ctx context.Context, logger *slog.Logger, ghClient *gh.Client, c *cache.Cache, owner, repo, path, ref string) error {
	if _, ok := c.GetGitHub(logger, owner, repo, path, ref); ok {
		return nil
	}
	content, err := fetchGitHubContent(ctx, ghClient, owner, repo, path, ref)
	if err != nil {
		return err
	}
	if err := c.PutGitHub(owner, repo, path, ref, []byte(content)); err != nil {
		return fmt.Errorf("cache GitHub content: %w", err)
	}
	return nil
}

func downloadAndCacheURL(ctx context.Context, logger *slog.Logger, c *cache.Cache, url string) error {
	if _, ok := c.GetURL(logger, url); ok {
		return nil
	}
	b, err := fetchURLContent(ctx, url)
	if err != nil {
		return err
	}
	if err := c.PutURL(url, b); err != nil {
		return fmt.Errorf("cache URL content: %w", err)
	}
	return nil
}
