package controller

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/suzuki-shunsuke/slog-util/slogutil"
	"github.com/suzuki-shunsuke/yamledit/pkg/cache"
	"github.com/suzuki-shunsuke/yamledit/pkg/config"
	gh "github.com/suzuki-shunsuke/yamledit/pkg/github"
)

func Add(ctx context.Context, logger *slogutil.Logger, ghClient *gh.Client, c *cache.Cache, dir, alias, migration string, force bool) error {
	if err := validateAddArgs(alias, migration); err != nil {
		return err
	}
	configPath := filepath.Join(dir, ".yamledit", "config.yaml")
	content, err := readOrCreateConfig(configPath)
	if err != nil {
		return err
	}
	if !force {
		if err := checkNameUniqueness(content, alias); err != nil {
			return err
		}
	}
	if err := downloadMigration(ctx, logger, ghClient, c, migration); err != nil {
		return fmt.Errorf("download migration: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".yamledit"), 0o755); err != nil { //nolint:mnd
		return fmt.Errorf("create .yamledit directory: %w", err)
	}
	return writeReusableRule(configPath, content, alias, migration, force)
}

func validateAddArgs(alias, migration string) error {
	if !strings.HasPrefix(migration, "https://") && !strings.HasPrefix(migration, "http://") && !strings.HasPrefix(migration, "github.com/") {
		return fmt.Errorf("invalid migration %q: must start with https://, http://, or github.com/", migration)
	}
	pattern := regexp.MustCompile(`^[a-z0-9_-]+$`)
	if !pattern.MatchString(alias) {
		return fmt.Errorf("invalid alias %q: must match %s", alias, pattern.String())
	}
	return nil
}

func readOrCreateConfig(configPath string) ([]byte, error) {
	b, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []byte("reusable_rules: []\n"), nil
		}
		return nil, fmt.Errorf("read config file: %w", err)
	}
	return b, nil
}

func checkNameUniqueness(content []byte, name string) error {
	var cfg config.Config
	if err := config.UnmarshalConfig(content, &cfg); err != nil {
		return fmt.Errorf("parse config file: %w", err)
	}
	if _, ok := cfg.FindReusableRule(name); ok {
		return fmt.Errorf("reusable rule %q already exists in config", name)
	}
	return nil
}

func downloadMigration(ctx context.Context, logger *slogutil.Logger, ghClient *gh.Client, c *cache.Cache, migration string) error {
	if err := config.DownloadAndCache(ctx, logger.Logger, ghClient, c, migration); err != nil {
		return fmt.Errorf("download migration %s: %w", migration, err)
	}
	return nil
}

func writeReusableRule(configPath string, content []byte, name, migration string, force bool) error {
	var cfg config.Config
	if err := config.UnmarshalConfig(content, &cfg); err != nil {
		return fmt.Errorf("parse config file: %w", err)
	}
	newRule := config.ReusableRule{Name: name, Import: migration}
	replaced := false
	if force {
		for i, r := range cfg.ReusableRules {
			if r.Name == name {
				cfg.ReusableRules[i] = newRule
				replaced = true
				break
			}
		}
	}
	if !replaced {
		cfg.ReusableRules = append(cfg.ReusableRules, newRule)
	}
	b, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(configPath, b, 0o644); err != nil { //nolint:gosec,mnd
		return fmt.Errorf("write config file: %w", err)
	}
	return nil
}
