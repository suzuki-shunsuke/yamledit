package config

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
	"github.com/suzuki-shunsuke/yamledit/pkg/cache"
	gh "github.com/suzuki-shunsuke/yamledit/pkg/github"
)

// ProjectConfig represents the structure of .yamledit/config.yaml.
type ProjectConfig struct {
	Aliases map[string]string `json:"aliases" yaml:"aliases"`
}

// UnmarshalProjectConfig parses YAML content into a ProjectConfig.
func UnmarshalProjectConfig(b []byte, cfg *ProjectConfig) error {
	if err := yaml.Unmarshal(b, cfg); err != nil {
		return fmt.Errorf("unmarshal project config: %w", err)
	}
	return nil
}

// ReadProjectConfig reads .yamledit/config.yaml and returns the ProjectConfig.
// Returns an empty ProjectConfig if the file does not exist.
func ReadProjectConfig(dir string) (*ProjectConfig, error) {
	p := filepath.Join(dir, ".yamledit", "config.yaml")
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &ProjectConfig{}, nil
		}
		return nil, fmt.Errorf("read project config: %w", err)
	}
	var cfg ProjectConfig
	if err := UnmarshalProjectConfig(b, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
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
