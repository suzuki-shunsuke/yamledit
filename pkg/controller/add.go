package controller

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/go-yamledit/yamledit"
	"github.com/suzuki-shunsuke/slog-util/slogutil"
	"github.com/suzuki-shunsuke/yamledit/pkg/cache"
	"github.com/suzuki-shunsuke/yamledit/pkg/config"
	gh "github.com/suzuki-shunsuke/yamledit/pkg/github"
)

func Add(ctx context.Context, logger *slogutil.Logger, ghClient *gh.Client, c *cache.Cache, dir, alias, migration string) error {
	if err := validateAddArgs(alias, migration); err != nil {
		return err
	}
	configPath := filepath.Join(dir, ".yamledit", "config.yaml")
	content, isNew, err := readOrCreateConfig(configPath)
	if err != nil {
		return err
	}
	if err := checkAliasUniqueness(content, alias); err != nil {
		return err
	}
	if err := downloadMigration(ctx, logger, ghClient, c, migration); err != nil {
		return fmt.Errorf("download migration: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".yamledit"), 0o755); err != nil { //nolint:mnd
		return fmt.Errorf("create .yamledit directory: %w", err)
	}
	return writeAlias(configPath, content, isNew, alias, migration)
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

func readOrCreateConfig(configPath string) ([]byte, bool, error) {
	b, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []byte("aliases: {}\n"), true, nil
		}
		return nil, false, fmt.Errorf("read config file: %w", err)
	}
	return b, false, nil
}

func checkAliasUniqueness(content []byte, alias string) error {
	var cfg config.ProjectConfig
	if err := config.UnmarshalProjectConfig(content, &cfg); err != nil {
		return fmt.Errorf("parse config file: %w", err)
	}
	if _, ok := cfg.Aliases[alias]; ok {
		return fmt.Errorf("alias %q already exists in config", alias)
	}
	return nil
}

func downloadMigration(ctx context.Context, logger *slogutil.Logger, ghClient *gh.Client, c *cache.Cache, migration string) error {
	if err := config.DownloadAndCache(ctx, logger.Logger, ghClient, c, migration); err != nil {
		return fmt.Errorf("download migration %s: %w", migration, err)
	}
	return nil
}

func writeAlias(configPath string, content []byte, isNew bool, alias, migration string) error {
	file, err := parser.ParseBytes(content, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse config YAML: %w", err)
	}
	var act yamledit.Action
	if isNew {
		// For new config, set aliases as a map at root level to get block style
		act = yamledit.MapAction("$", yamledit.SetKey("aliases", map[string]string{alias: migration}, nil))
	} else {
		act = yamledit.MapAction("$.aliases", yamledit.SetKey(alias, migration, nil))
	}
	if err := act.Run(file.Docs[0].Body); err != nil {
		return fmt.Errorf("add alias to config: %w", err)
	}
	if err := os.WriteFile(configPath, []byte(file.String()), 0o644); err != nil { //nolint:gosec,mnd
		return fmt.Errorf("write config file: %w", err)
	}
	return nil
}
