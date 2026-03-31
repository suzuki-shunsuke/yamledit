package controller

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

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
	pattern := filepath.Join(dir, ".yamledit", "*", "ruleset.yaml")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("glob ruleset files: %w", err)
	}
	names := make([]string, 0, len(matches))
	for _, m := range matches {
		name := filepath.Base(filepath.Dir(m))
		names = append(names, name)
	}
	return names, nil
}

func testMigration(ctx context.Context, logger *slogutil.Logger, ghClient *gh.Client, c *cache.Cache, dir, name string) (bool, error) { //nolint:cyclop
	cfgPath := filepath.Join(dir, ".yamledit", name, "ruleset.yaml")
	cfg, err := config.ReadRuleset(cfgPath)
	if err != nil {
		return false, fmt.Errorf("read config: %w", err)
	}
	if err := config.ResolveImports(ctx, logger.Logger, ghClient, c, cfg); err != nil {
		return false, fmt.Errorf("resolve imports: %w", err)
	}
	actions, err := buildAllActions(logger, cfg)
	if err != nil {
		return false, fmt.Errorf("build actions: %w", err)
	}

	rulesetDir := filepath.Join(dir, ".yamledit", name)
	entries, err := os.ReadDir(rulesetDir)
	if err != nil {
		return false, fmt.Errorf("read ruleset directory: %w", err)
	}

	var failed bool
	var hasTests bool
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		hasTests = true
		f, err := runTestCase(logger, rulesetDir, entry.Name(), name, actions)
		if err != nil {
			return false, err
		}
		if f {
			failed = true
		}
	}
	if !hasTests {
		logger.Warn("no test directories, skipping", "migration", name)
	} else if !failed {
		logger.Info("all tests passed", "migration", name)
	}
	return failed, nil
}

func runTestCase(logger *slogutil.Logger, rulesetDir, testName, migrationName string, actions []yamledit.Action) (bool, error) {
	testCaseDir := filepath.Join(rulesetDir, testName)
	resultPath := filepath.Join(testCaseDir, "result.yaml")
	if _, err := os.Stat(resultPath); err != nil {
		if os.IsNotExist(err) {
			logger.Warn("no result file, skipping", "migration", migrationName, "test", testName)
			return false, nil
		}
		return false, fmt.Errorf("stat result file: %w", err)
	}

	inputPath := filepath.Join(testCaseDir, "test.yaml")
	input, err := os.ReadFile(inputPath)
	if err != nil {
		return false, fmt.Errorf("read test input %s: %w", inputPath, err)
	}
	expected, err := os.ReadFile(resultPath)
	if err != nil {
		return false, fmt.Errorf("read expected result %s: %w", resultPath, err)
	}

	got, err := yamledit.EditBytes(inputPath, input, actions...)
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

func buildAllActions(logger *slogutil.Logger, cfg *config.Ruleset) ([]yamledit.Action, error) {
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
