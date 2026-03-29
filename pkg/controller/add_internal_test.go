package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdd_createNewConfig(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`rules: []`)) //nolint:errcheck
	}))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".yamledit"), 0o755); err != nil {
		t.Fatal(err)
	}
	err := Add(context.Background(), newLogger(), nil, nil, dir, "my-rule", srv.URL+"/migration.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(dir, ".yamledit", "config.yaml"))
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	got := string(b)
	if !strings.Contains(got, "name: my-rule") || !strings.Contains(got, "import: "+srv.URL+"/migration.yaml") {
		t.Errorf("config should contain reusable rule, got:\n%s", got)
	}
}

func TestAdd_addToExistingConfig(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`rules: []`)) //nolint:errcheck
	}))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	configDir := filepath.Join(dir, ".yamledit")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("reusable_rules:\n  - name: existing\n    import: https://example.com/existing\n"), 0o644); err != nil { //nolint:gosec
		t.Fatal(err)
	}

	err := Add(context.Background(), newLogger(), nil, nil, dir, "new-rule", srv.URL+"/migration.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(configDir, "config.yaml"))
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	got := string(b)
	if !strings.Contains(got, "name: existing") || !strings.Contains(got, "name: new-rule") {
		t.Errorf("config should contain both rules, got:\n%s", got)
	}
}

func TestAdd_aliasAlreadyExists(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configDir := filepath.Join(dir, ".yamledit")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("reusable_rules:\n  - name: my-rule\n    import: https://example.com/old\n"), 0o644); err != nil { //nolint:gosec
		t.Fatal(err)
	}

	err := Add(context.Background(), newLogger(), nil, nil, dir, "my-rule", "https://example.com/new")
	if err == nil {
		t.Fatal("expected error for duplicate alias, got nil")
	}
}

func TestAdd_invalidMigrationPrefix(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	err := Add(context.Background(), newLogger(), nil, nil, dir, "my-rule", "ftp://example.com/migration.yaml")
	if err == nil {
		t.Fatal("expected error for invalid migration prefix, got nil")
	}
}

func TestAdd_invalidAlias(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	err := Add(context.Background(), newLogger(), nil, nil, dir, "My-Rule", "https://example.com/migration.yaml")
	if err == nil {
		t.Fatal("expected error for invalid alias, got nil")
	}
}

func TestAdd_createsYamleditDir(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`rules: []`)) //nolint:errcheck
	}))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	// Don't create .yamledit dir — Add should create it
	err := Add(context.Background(), newLogger(), nil, nil, dir, "my-rule", srv.URL+"/migration.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".yamledit", "config.yaml")); err != nil {
		t.Fatalf("config.yaml not created: %v", err)
	}
}
