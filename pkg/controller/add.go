package controller

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/go-yamledit/yamledit"
	"github.com/suzuki-shunsuke/slog-util/slogutil"
	"github.com/suzuki-shunsuke/yamledit/pkg/cache"
	"github.com/suzuki-shunsuke/yamledit/pkg/config"
	gh "github.com/suzuki-shunsuke/yamledit/pkg/github"
)

func Add(ctx context.Context, stderr io.Writer, logger *slogutil.Logger, ghClient *gh.Client, c *cache.Cache, dir, alias, migration string, force, global bool) error {
	if err := validateAddArgs(alias, migration); err != nil {
		return err
	}
	configPath, err := resolveAddConfigPath(dir, global)
	if err != nil {
		return err
	}
	content, isNew, err := readOrCreateConfig(configPath)
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
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil { //nolint:mnd
		return fmt.Errorf("create config directory: %w", err)
	}
	if err := writeReusableRule(configPath, content, alias, migration, force); err != nil {
		return err
	}
	if isNew {
		fmt.Fprintf(stderr, "%s was created.\n", configPath)
	} else {
		fmt.Fprintf(stderr, "%s was updated.\n", configPath)
	}
	return nil
}

func resolveAddConfigPath(dir string, global bool) (string, error) {
	if !global {
		return filepath.Join(dir, ".yamledit", "config.yaml"), nil
	}
	p := config.GlobalConfigPath()
	if p == "" {
		return "", errors.New("cannot determine global config path")
	}
	return p, nil
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
			return []byte("# yaml-language-server: $schema=https://raw.githubusercontent.com/suzuki-shunsuke/yamledit/main/json-schema/config.json\nreusable_rules: []\n"), true, nil
		}
		return nil, false, fmt.Errorf("read config file: %w", err)
	}
	return b, false, nil
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
	file, err := parser.ParseBytes(content, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse config YAML: %w", err)
	}
	if force {
		if err := replaceReusableRule(file, content, name, migration); err == nil {
			return writeConfigFile(configPath, file)
		}
	}
	act := yamledit.ListAction("$.reusable_rules", yamledit.AddValuesToList(-1,
		yamledit.NewBytes(fmt.Appendf(nil, "name: %s\nimport: %s\n", name, migration)),
	))
	if err := act.Run(file.Docs[0].Body); err != nil {
		return fmt.Errorf("add reusable rule: %w", err)
	}
	return writeConfigFile(configPath, file)
}

func replaceReusableRule(file *ast.File, content []byte, name, migration string) error {
	var cfg config.Config
	if err := config.UnmarshalConfig(content, &cfg); err != nil {
		return fmt.Errorf("parse config file: %w", err)
	}
	idx := findReusableRuleIndex(cfg, name)
	if idx < 0 {
		return fmt.Errorf("rule %q not found", name)
	}
	act := yamledit.MapAction(fmt.Sprintf("$.reusable_rules[%d]", idx),
		yamledit.SetKey("import", migration, nil),
	)
	if err := act.Run(file.Docs[0].Body); err != nil {
		return fmt.Errorf("update reusable rule: %w", err)
	}
	return nil
}

func writeConfigFile(configPath string, file *ast.File) error {
	if err := os.WriteFile(configPath, []byte(file.String()), 0o644); err != nil { //nolint:gosec,mnd
		return fmt.Errorf("write config file: %w", err)
	}
	return nil
}

func findReusableRuleIndex(cfg config.Config, name string) int {
	for i, r := range cfg.ReusableRules {
		if r.Name == name {
			return i
		}
	}
	return -1
}
