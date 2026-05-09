package cache

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"time"

	"github.com/goccy/go-yaml"
)

const expirationDays = 3

type Cache struct {
	Dir     string
	NoCache bool
}

func New(noCache bool) *Cache {
	return &Cache{
		Dir:     resolveDir(),
		NoCache: noCache,
	}
}

func resolveDir() string {
	if v := os.Getenv("YAMLEDIT_CACHE_HOME"); v != "" {
		return v
	}
	if v := os.Getenv("XDG_CACHE_HOME"); v != "" {
		return filepath.Join(v, "yamledit")
	}
	if runtime.GOOS == "windows" {
		if v := os.Getenv("LOCALAPPDATA"); v != "" {
			return filepath.Join(v, "cache", "yamledit")
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".cache", "yamledit")
	}
	return filepath.Join(home, ".cache", "yamledit")
}

func (c *Cache) GetURL(logger *slog.Logger, url string) ([]byte, bool) {
	if c == nil || c.NoCache {
		return nil, false
	}
	dir := c.urlDir(url)
	return c.get(logger, dir, "")
}

func (c *Cache) PutURL(url string, content []byte) error {
	if c == nil {
		return nil
	}
	dir := c.urlDir(url)
	return c.put(dir, content)
}

func (c *Cache) GetGitHub(logger *slog.Logger, owner, repo, path, ref string) ([]byte, bool) {
	if c == nil || c.NoCache {
		return nil, false
	}
	dir := c.githubDir(owner, repo, path, ref)
	return c.get(logger, dir, ref)
}

func (c *Cache) PutGitHub(owner, repo, path, ref string, content []byte) error {
	if c == nil {
		return nil
	}
	dir := c.githubDir(owner, repo, path, ref)
	return c.put(dir, content)
}

const topicExpirationDays = 1

func (c *Cache) GetTopicSearch(_ *slog.Logger) ([]byte, bool) {
	if c == nil || c.NoCache {
		return nil, false
	}
	p := filepath.Join(c.Dir, "topics", "yamledit-rule", "result.json")
	info, err := os.Stat(p)
	if err != nil {
		return nil, false
	}
	if time.Since(info.ModTime()) > topicExpirationDays*24*time.Hour {
		return nil, false
	}
	content, err := os.ReadFile(p)
	if err != nil {
		return nil, false
	}
	return content, true
}

func (c *Cache) PutTopicSearch(content []byte) error {
	if c == nil {
		return nil
	}
	dir := filepath.Join(c.Dir, "topics", "yamledit-rule")
	if err := os.MkdirAll(dir, 0o755); err != nil { //nolint:mnd
		return fmt.Errorf("create topic cache directory: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "result.json"), content, 0o644); err != nil { //nolint:gosec,mnd
		return fmt.Errorf("write result.json: %w", err)
	}
	return nil
}

var (
	semverPattern = regexp.MustCompile(`^v?\d+\.\d+\.\d+$`)
	shaPattern    = regexp.MustCompile(`^[0-9a-f]{40}$`)
)

func (c *Cache) urlDir(url string) string {
	encoded := base64.URLEncoding.EncodeToString([]byte(url))
	return filepath.Join(c.Dir, "remotes", "url", encoded)
}

func (c *Cache) githubDir(owner, repo, path, ref string) string {
	encodedRef := base64.URLEncoding.EncodeToString([]byte(ref))
	encodedPath := base64.URLEncoding.EncodeToString([]byte(path))
	return filepath.Join(c.Dir, "remotes", "github_content", "github.com", owner, repo, encodedRef, encodedPath)
}

func (c *Cache) get(logger *slog.Logger, dir, ref string) ([]byte, bool) {
	migrationPath := filepath.Join(dir, "migration.yaml")
	info, err := os.Stat(migrationPath)
	if err != nil {
		return nil, false
	}
	if isExpired(info.ModTime(), ref) {
		return nil, false
	}
	content, err := os.ReadFile(migrationPath)
	if err != nil {
		return nil, false
	}
	// Validate that content is valid YAML.
	var v any
	if err := yaml.Unmarshal(content, &v); err != nil {
		logger.Debug("parse cached migration.yaml", "path", migrationPath, "error", err)
		return nil, false
	}
	return content, true
}

func (c *Cache) put(dir string, content []byte) error {
	if err := os.MkdirAll(dir, 0o755); err != nil { //nolint:mnd
		return fmt.Errorf("create cache directory: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "migration.yaml"), content, 0o644); err != nil { //nolint:gosec,mnd
		return fmt.Errorf("write migration.yaml: %w", err)
	}
	return nil
}

func isExpired(modTime time.Time, ref string) bool {
	if ref != "" {
		if semverPattern.MatchString(ref) || shaPattern.MatchString(ref) {
			return false
		}
	}
	return time.Since(modTime) > expirationDays*24*time.Hour
}
