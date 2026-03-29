package cache

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestResolveDir(t *testing.T) {
	// Cannot use t.Parallel() because subtests use t.Setenv
	tests := []struct {
		name string
		env  map[string]string
		want string
	}{
		{
			name: "YAMLEDIT_CACHE_HOME takes priority",
			env:  map[string]string{"YAMLEDIT_CACHE_HOME": "/custom/cache"},
			want: "/custom/cache",
		},
		{
			name: "XDG_CACHE_HOME fallback",
			env:  map[string]string{"XDG_CACHE_HOME": "/xdg/cache"},
			want: "/xdg/cache/yamledit",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Cannot use t.Parallel() with t.Setenv
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			// Clear higher-priority vars if not set in this test
			if _, ok := tt.env["YAMLEDIT_CACHE_HOME"]; !ok {
				t.Setenv("YAMLEDIT_CACHE_HOME", "")
			}
			got := resolveDir()
			if got != tt.want {
				t.Errorf("resolveDir() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetPutURL(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	c := &Cache{Dir: dir}

	url := "https://example.com/migration.yaml"
	content := []byte("rules: []\n")

	// Cache miss initially
	if _, ok := c.GetURL(slog.Default(), url); ok {
		t.Fatal("expected cache miss")
	}

	// Put then Get
	if err := c.PutURL(url, content); err != nil {
		t.Fatalf("PutURL: %v", err)
	}
	got, ok := c.GetURL(slog.Default(), url)
	if !ok {
		t.Fatal("expected cache hit")
	}
	if string(got) != string(content) {
		t.Errorf("got %q, want %q", got, content)
	}
}

func TestGetPutGitHub(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	c := &Cache{Dir: dir}

	content := []byte("rules: []\n")

	// Put then Get
	if err := c.PutGitHub("owner", "repo", "path/to/file.yaml", "v1.0.0", content); err != nil {
		t.Fatalf("PutGitHub: %v", err)
	}
	got, ok := c.GetGitHub(slog.Default(), "owner", "repo", "path/to/file.yaml", "v1.0.0")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if string(got) != string(content) {
		t.Errorf("got %q, want %q", got, content)
	}
}

func TestExpiration_Semver(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	c := &Cache{Dir: dir}

	content := []byte("rules: []\n")
	if err := c.PutGitHub("o", "r", "p", "v1.2.3", content); err != nil {
		t.Fatal(err)
	}

	// Backdate metadata to 30 days ago
	backdateCache(t, c.githubDir("o", "r", "p", "v1.2.3"), 30*24*time.Hour)

	// Semver should never expire
	if _, ok := c.GetGitHub(slog.Default(), "o", "r", "p", "v1.2.3"); !ok {
		t.Fatal("semver ref should not expire")
	}
}

func TestExpiration_SHA(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	c := &Cache{Dir: dir}

	sha := "abc123def456abc123def456abc123def456abc1"
	content := []byte("rules: []\n")
	if err := c.PutGitHub("o", "r", "p", sha, content); err != nil {
		t.Fatal(err)
	}

	backdateCache(t, c.githubDir("o", "r", "p", sha), 30*24*time.Hour)

	if _, ok := c.GetGitHub(slog.Default(), "o", "r", "p", sha); !ok {
		t.Fatal("SHA ref should not expire")
	}
}

func TestExpiration_Branch(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	c := &Cache{Dir: dir}

	content := []byte("rules: []\n")
	if err := c.PutGitHub("o", "r", "p", "main", content); err != nil {
		t.Fatal(err)
	}

	// Within 3 days: should hit
	backdateCache(t, c.githubDir("o", "r", "p", "main"), 2*24*time.Hour)
	if _, ok := c.GetGitHub(slog.Default(), "o", "r", "p", "main"); !ok {
		t.Fatal("expected cache hit within 3 days")
	}

	// After 3 days: should miss
	backdateCache(t, c.githubDir("o", "r", "p", "main"), 4*24*time.Hour)
	if _, ok := c.GetGitHub(slog.Default(), "o", "r", "p", "main"); ok {
		t.Fatal("expected cache miss after 3 days")
	}
}

func TestExpiration_URL(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	c := &Cache{Dir: dir}

	url := "https://example.com/rules.yaml"
	content := []byte("rules: []\n")
	if err := c.PutURL(url, content); err != nil {
		t.Fatal(err)
	}

	// After 3 days: should miss
	backdateCache(t, c.urlDir(url), 4*24*time.Hour)
	if _, ok := c.GetURL(slog.Default(), url); ok {
		t.Fatal("expected cache miss after 3 days for URL")
	}
}

func TestCorruptedMigration(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	c := &Cache{Dir: dir}

	content := []byte("rules: []\n")
	if err := c.PutURL("https://example.com/b.yaml", content); err != nil {
		t.Fatal(err)
	}

	// Corrupt migration.yaml
	migPath := filepath.Join(c.urlDir("https://example.com/b.yaml"), "migration.yaml")
	if err := os.WriteFile(migPath, []byte(":\t:\n\tinvalid yaml"), 0o644); err != nil { //nolint:gosec
		t.Fatal(err)
	}

	if _, ok := c.GetURL(slog.Default(), "https://example.com/b.yaml"); ok {
		t.Fatal("expected cache miss with corrupted migration")
	}
}

func TestNoCache(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	c := &Cache{Dir: dir, NoCache: true}

	content := []byte("rules: []\n")
	// Put should still write
	if err := c.PutURL("https://example.com/c.yaml", content); err != nil {
		t.Fatal(err)
	}

	// Get should always miss
	if _, ok := c.GetURL(slog.Default(), "https://example.com/c.yaml"); ok {
		t.Fatal("expected cache miss with NoCache")
	}

	// But the file should exist on disk
	c2 := &Cache{Dir: dir, NoCache: false}
	if _, ok := c2.GetURL(slog.Default(), "https://example.com/c.yaml"); !ok {
		t.Fatal("expected cache hit with NoCache disabled")
	}
}

func TestNilCache(t *testing.T) {
	t.Parallel()
	var c *Cache
	if _, ok := c.GetURL(slog.Default(), "https://example.com"); ok {
		t.Fatal("expected miss on nil cache")
	}
	if _, ok := c.GetGitHub(slog.Default(), "o", "r", "p", "ref"); ok {
		t.Fatal("expected miss on nil cache")
	}
	if err := c.PutURL("https://example.com", []byte("rules: []\n")); err != nil {
		t.Fatalf("PutURL on nil should be no-op: %v", err)
	}
	if err := c.PutGitHub("o", "r", "p", "ref", []byte("rules: []\n")); err != nil {
		t.Fatalf("PutGitHub on nil should be no-op: %v", err)
	}
}

func backdateCache(t *testing.T, dir string, age time.Duration) {
	t.Helper()
	migrationPath := filepath.Join(dir, "migration.yaml")
	past := time.Now().Add(-age)
	if err := os.Chtimes(migrationPath, past, past); err != nil {
		t.Fatalf("chtimes migration.yaml: %v", err)
	}
}
