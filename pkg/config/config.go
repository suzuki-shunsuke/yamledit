package config

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/goccy/go-yaml"
	gh "github.com/suzuki-shunsuke/yamledit/pkg/github"
)

type Config struct {
	Rules []*Rule `json:"rules" yaml:"rules" jsonschema_description:"List of migration rules. Rules are evaluated in order."`
}

type Rule struct {
	Path    string    `json:"path" yaml:"path" jsonschema_description:"YAML path to the target node (e.g. \"$\", \"$.foo\")"`
	Actions []*Action `json:"actions" yaml:"actions" jsonschema_description:"List of actions to apply. Actions are evaluated in order."`
	Files   []string  `json:"files,omitempty" yaml:"files" jsonschema_description:"File path patterns to apply this rule to. Glob patterns with ** supported. Paths starting with ! are excluded."`
	Import  string    `json:"import,omitempty" yaml:"import" jsonschema_description:"URL to import rules from a remote migration file."`
}

type Action struct {
	Type              string            `json:"type" yaml:"type" jsonschema_description:"Action type: remove_keys, rename_key, set_key, add_values, sort_key, remove_values, or sort_list"`
	Keys              []string          `json:"keys,omitempty" yaml:"keys" jsonschema_description:"Keys to remove (for remove_keys)"`
	Key               string            `json:"key,omitempty" yaml:"key" jsonschema_description:"Target key name (for rename_key, set_key)"`
	NewKey            string            `json:"new_key,omitempty" yaml:"new_key" jsonschema_description:"New key name (for rename_key)"`
	WhenDuplicate     string            `json:"when_duplicate,omitempty" yaml:"when_duplicate" jsonschema_description:"Behavior when new_key already exists: skip (default), ignore_existing_key, remove_old_key, fail (for rename_key)"`
	Value             any               `json:"value,omitempty" yaml:"value" jsonschema_description:"Value to set (for set_key)"`
	SkipIfKeyNotFound bool              `json:"skip_if_key_not_found,omitempty" yaml:"skip_if_key_not_found" jsonschema_description:"If true, do nothing when the key does not exist (for set_key)"`
	SkipIfKeyFound    bool              `json:"skip_if_key_found,omitempty" yaml:"skip_if_key_found" jsonschema_description:"If true, do nothing when the key already exists (for set_key)"`
	ClearComment      bool              `json:"clear_comment,omitempty" yaml:"clear_comment" jsonschema_description:"If true, remove the comment on the existing key (for set_key)"`
	InsertAt          []*InsertLocation `json:"insert_at,omitempty" yaml:"insert_at" jsonschema_description:"Where to insert a new key. The first matching condition is used. If none match, the key is appended at the end (for set_key)"`
	Values            []any             `json:"values,omitempty" yaml:"values" jsonschema_description:"Values to add to the list (for add_values)"`
	Index             *int              `json:"index,omitempty" yaml:"index" jsonschema_description:"Index to insert values at. 0 for beginning, negative values count from end. Default -1 (for add_values)"`
	Expr              string            `json:"expr,omitempty" yaml:"expr" jsonschema_description:"Expression for sorting keys. Variables a and b represent key-value pairs with fields: key, value, comment, index (for sort_key)"`
}

type InsertLocation struct {
	AfterKey  string `json:"after_key,omitempty" yaml:"after_key" jsonschema_description:"Insert after this key"`
	BeforeKey string `json:"before_key,omitempty" yaml:"before_key" jsonschema_description:"Insert before this key"`
	First     bool   `json:"first,omitempty" yaml:"first" jsonschema_description:"Insert at the beginning"`
}

func ResolveImports(ctx context.Context, ghClient *gh.Client, cfg *Config) error {
	var resolved []*Rule
	for _, rule := range cfg.Rules {
		if rule.Import == "" {
			resolved = append(resolved, rule)
			continue
		}
		imported, err := resolveImport(ctx, ghClient, rule.Import)
		if err != nil {
			return fmt.Errorf("import %s: %w", rule.Import, err)
		}
		if err := ResolveImports(ctx, ghClient, imported); err != nil {
			return err
		}
		resolved = append(resolved, imported.Rules...)
	}
	cfg.Rules = resolved
	return nil
}

func resolveImport(ctx context.Context, ghClient *gh.Client, s string) (*Config, error) {
	if owner, repo, path, ref, ok := parseGitHubImport(s); ok {
		return fetchAndParseGitHub(ctx, ghClient, owner, repo, path, ref)
	}
	return fetchAndParse(ctx, s)
}

func parseGitHubImport(s string) (owner, repo, path, ref string, ok bool) {
	const prefix = "github.com/"
	if !strings.HasPrefix(s, prefix) {
		return "", "", "", "", false
	}
	rest := s[len(prefix):]
	// Split ref: "owner/repo/path:ref"
	if i := strings.LastIndex(rest, ":"); i >= 0 {
		ref = rest[i+1:]
		rest = rest[:i]
	}
	// Split: owner/repo/path...
	parts := strings.SplitN(rest, "/", 3) //nolint:mnd
	if len(parts) < 3 {                   //nolint:mnd
		return "", "", "", "", false
	}
	return parts[0], parts[1], parts[2], ref, true
}

func fetchAndParseGitHub(ctx context.Context, ghClient *gh.Client, owner, repo, path, ref string) (*Config, error) {
	content, err := ghClient.GetContent(ctx, owner, repo, path, ref)
	if err != nil {
		return nil, fmt.Errorf("get GitHub content: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal YAML: %w", err)
	}
	return &cfg, nil
}

func fetchAndParse(ctx context.Context, url string) (*Config, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch URL: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch URL: status %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal YAML: %w", err)
	}
	return &cfg, nil
}

func ReadConfigs(ctx context.Context, ghClient *gh.Client, dir string) ([]*Config, error) {
	pattern := filepath.Join(dir, ".yamledit", "*.yaml")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("glob migration files: %w", err)
	}
	configs := make([]*Config, 0, len(matches))
	for _, p := range matches {
		cfg, err := ReadConfig(p)
		if err != nil {
			return nil, fmt.Errorf("read migration file %s: %w", p, err)
		}
		if err := ResolveImports(ctx, ghClient, cfg); err != nil {
			return nil, fmt.Errorf("resolve imports in %s: %w", p, err)
		}
		configs = append(configs, cfg)
	}
	return configs, nil
}

func ReadConfigsByPaths(ctx context.Context, ghClient *gh.Client, dir string, paths []string) ([]*Config, error) {
	configs := make([]*Config, 0, len(paths))
	for _, p := range paths {
		if !filepath.IsAbs(p) && !yamlSuffixPattern.MatchString(p) {
			p = filepath.Join(dir, ".yamledit", p+".yaml")
		}
		cfg, err := ReadConfig(p)
		if err != nil {
			return nil, fmt.Errorf("read migration file %s: %w", p, err)
		}
		if err := ResolveImports(ctx, ghClient, cfg); err != nil {
			return nil, fmt.Errorf("resolve imports in %s: %w", p, err)
		}
		configs = append(configs, cfg)
	}
	return configs, nil
}

var yamlSuffixPattern = regexp.MustCompile(`\.ya?ml$`)

func ReadConfig(p string) (*Config, error) {
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal YAML: %w", err)
	}
	return &cfg, nil
}
