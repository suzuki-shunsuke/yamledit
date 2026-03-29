package controller

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func setupTestFiles(t *testing.T, dir, migration, testName, input, result string) { //nolint:unparam
	t.Helper()
	testDir := filepath.Join(dir, ".yamledit", migration+"_test")
	if err := os.MkdirAll(testDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(testDir, testName+".yaml"), []byte(input), 0o644); err != nil { //nolint:gosec
		t.Fatal(err)
	}
	if result != "" {
		if err := os.WriteFile(filepath.Join(testDir, testName+"_result.yaml"), []byte(result), 0o644); err != nil { //nolint:gosec
			t.Fatal(err)
		}
	}
}

func TestTest_pass(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	setupMigration(t, dir, "test", `rules:
  - path: "$"
    actions:
      - type: remove_keys
        keys:
          - age
`)
	setupTestFiles(t, dir, "test", "basic", "name: alice\nage: 30\n", "name: alice\n")

	err := Test(context.Background(), newLogger(), nil, nil, dir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTest_fail(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	setupMigration(t, dir, "test", `rules:
  - path: "$"
    actions:
      - type: remove_keys
        keys:
          - age
`)
	setupTestFiles(t, dir, "test", "basic", "name: alice\nage: 30\n", "name: alice\nage: 30\n")

	err := Test(context.Background(), newLogger(), nil, nil, dir, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTest_missingResultFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	setupMigration(t, dir, "test", `rules:
  - path: "$"
    actions:
      - type: remove_keys
        keys:
          - age
`)
	setupTestFiles(t, dir, "test", "basic", "name: alice\nage: 30\n", "")

	err := Test(context.Background(), newLogger(), nil, nil, dir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTest_noTestDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	setupMigration(t, dir, "test", `rules:
  - path: "$"
    actions:
      - type: remove_keys
        keys:
          - age
`)

	err := Test(context.Background(), newLogger(), nil, nil, dir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDiscoverMigrations_skipsConfigYAML(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configDir := filepath.Join(dir, ".yamledit")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("aliases: {}\n"), 0o644); err != nil { //nolint:gosec
		t.Fatal(err)
	}
	setupMigration(t, dir, "foo", `rules:
  - path: "$"
    actions:
      - type: remove_keys
        keys:
          - age
`)
	names, err := discoverMigrations(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 1 {
		t.Fatalf("expected 1 migration (config.yaml should be skipped), got %d: %v", len(names), names)
	}
	if names[0] != "foo" {
		t.Errorf("expected migration name 'foo', got %q", names[0])
	}
}

func TestTest_specificMigration(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	setupMigration(t, dir, "mig1", `rules:
  - path: "$"
    actions:
      - type: remove_keys
        keys:
          - age
`)
	setupTestFiles(t, dir, "mig1", "basic", "name: alice\nage: 30\n", "name: alice\n")

	err := Test(context.Background(), newLogger(), nil, nil, dir, []string{"mig1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
