package controller

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/suzuki-shunsuke/go-yamledit/yamledit"
	"github.com/suzuki-shunsuke/slog-util/slogutil"
	"github.com/suzuki-shunsuke/yamledit/pkg/cache"
	"github.com/suzuki-shunsuke/yamledit/pkg/config"
	gh "github.com/suzuki-shunsuke/yamledit/pkg/github"
)

func Test(ctx context.Context, logger *slogutil.Logger, ghClient *gh.Client, c *cache.Cache, dir string, migrations []string) error {
	if len(migrations) == 0 {
		var err error
		migrations, err = discoverMigrations(dir)
		if err != nil {
			return fmt.Errorf("discover migrations: %w", err)
		}
	}
	var failed bool
	for _, name := range migrations {
		f, err := testMigration(ctx, logger, ghClient, c, dir, name)
		if err != nil {
			return fmt.Errorf("test migration %s: %w", name, err)
		}
		if f {
			failed = true
		}
	}
	if failed {
		return errors.New("some tests failed")
	}
	return nil
}

func discoverMigrations(dir string) ([]string, error) {
	pattern := filepath.Join(dir, ".yamledit", "*.yaml")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("glob migration files: %w", err)
	}
	names := make([]string, 0, len(matches))
	for _, m := range matches {
		name := strings.TrimSuffix(filepath.Base(m), ".yaml")
		names = append(names, name)
	}
	return names, nil
}

func testMigration(ctx context.Context, logger *slogutil.Logger, ghClient *gh.Client, c *cache.Cache, dir, name string) (bool, error) { //nolint:cyclop
	cfgPath := filepath.Join(dir, ".yamledit", name+".yaml")
	cfg, err := config.ReadConfig(cfgPath)
	if err != nil {
		return false, fmt.Errorf("read config: %w", err)
	}
	if err := config.ResolveImports(ctx, ghClient, c, cfg); err != nil {
		return false, fmt.Errorf("resolve imports: %w", err)
	}
	actions, err := buildAllActions(logger, cfg)
	if err != nil {
		return false, fmt.Errorf("build actions: %w", err)
	}

	testDir := filepath.Join(dir, ".yamledit", name+"_test")
	entries, err := os.ReadDir(testDir)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Warn("no test directory, skipping", "migration", name)
			return false, nil
		}
		return false, fmt.Errorf("read test directory: %w", err)
	}

	var failed bool
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fname := entry.Name()
		if !strings.HasSuffix(fname, ".yaml") || strings.HasSuffix(fname, "_result.yaml") {
			continue
		}
		f, err := runTestCase(logger, testDir, fname, name, actions)
		if err != nil {
			return false, err
		}
		if f {
			failed = true
		}
	}
	if !failed {
		logger.Info("all tests passed", "migration", name)
	}
	return failed, nil
}

func runTestCase(logger *slogutil.Logger, testDir, fname, migrationName string, actions []yamledit.Action) (bool, error) {
	testName := strings.TrimSuffix(fname, ".yaml")
	resultPath := filepath.Join(testDir, testName+"_result.yaml")
	if _, err := os.Stat(resultPath); err != nil {
		if os.IsNotExist(err) {
			logger.Warn("no result file, skipping", "migration", migrationName, "test", testName)
			return false, nil
		}
		return false, fmt.Errorf("stat result file: %w", err)
	}

	inputPath := filepath.Join(testDir, fname)
	input, err := os.ReadFile(inputPath)
	if err != nil {
		return false, fmt.Errorf("read test input %s: %w", inputPath, err)
	}
	expected, err := os.ReadFile(resultPath)
	if err != nil {
		return false, fmt.Errorf("read expected result %s: %w", resultPath, err)
	}

	got, err := applyActions(input, actions)
	if err != nil {
		return false, fmt.Errorf("apply actions to %s: %w", inputPath, err)
	}

	if got == string(expected) {
		logger.Info("PASS", "migration", migrationName, "test", testName)
		return false, nil
	}

	fmt.Fprintf(os.Stderr, "FAIL: %s/%s\n", migrationName, testName)
	printDiff(resultPath, string(expected), got)
	return true, nil
}

func buildAllActions(logger *slogutil.Logger, cfg *config.Config) ([]yamledit.Action, error) {
	var actions []yamledit.Action
	for _, rule := range cfg.Rules {
		for _, a := range rule.Actions {
			act, err := buildAction(logger, rule.Path, a)
			if err != nil {
				return nil, err
			}
			actions = append(actions, act)
		}
	}
	return actions, nil
}
