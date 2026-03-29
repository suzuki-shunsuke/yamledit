package config

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/suzuki-shunsuke/yamledit/pkg/cache"
	gh "github.com/suzuki-shunsuke/yamledit/pkg/github"
)

type Ruleset struct {
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

func ResolveImports(ctx context.Context, logger *slog.Logger, ghClient *gh.Client, c *cache.Cache, cfg *Ruleset) error {
	var resolved []*Rule
	for _, rule := range cfg.Rules {
		if rule.Import == "" {
			resolved = append(resolved, rule)
			continue
		}
		imported, err := resolveImport(ctx, logger, ghClient, c, rule.Import)
		if err != nil {
			return fmt.Errorf("import %s: %w", rule.Import, err)
		}
		if err := ResolveImports(ctx, logger, ghClient, c, imported); err != nil {
			return err
		}
		resolved = append(resolved, imported.Rules...)
	}
	cfg.Rules = resolved
	return nil
}

func resolveImport(ctx context.Context, logger *slog.Logger, ghClient *gh.Client, c *cache.Cache, s string) (*Ruleset, error) {
	if owner, repo, path, ref, ok := parseGitHubImport(s); ok {
		return resolveGitHubImport(ctx, logger, ghClient, c, owner, repo, path, ref)
	}
	return resolveURLImport(ctx, logger, c, s)
}

func resolveGitHubImport(ctx context.Context, logger *slog.Logger, ghClient *gh.Client, c *cache.Cache, owner, repo, path, ref string) (*Ruleset, error) {
	if b, ok := c.GetGitHub(logger, owner, repo, path, ref); ok {
		var cfg Ruleset
		if err := yaml.Unmarshal(b, &cfg); err != nil {
			return nil, fmt.Errorf("unmarshal cached YAML: %w", err)
		}
		return &cfg, nil
	}
	content, err := fetchGitHubContent(ctx, ghClient, owner, repo, path, ref)
	if err != nil {
		return nil, err
	}
	if err := c.PutGitHub(owner, repo, path, ref, []byte(content)); err != nil {
		return nil, fmt.Errorf("cache GitHub content: %w", err)
	}
	var cfg Ruleset
	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal YAML: %w", err)
	}
	return &cfg, nil
}

func resolveURLImport(ctx context.Context, logger *slog.Logger, c *cache.Cache, url string) (*Ruleset, error) {
	if b, ok := c.GetURL(logger, url); ok {
		var cfg Ruleset
		if err := yaml.Unmarshal(b, &cfg); err != nil {
			return nil, fmt.Errorf("unmarshal cached YAML: %w", err)
		}
		return &cfg, nil
	}
	b, err := fetchURLContent(ctx, url)
	if err != nil {
		return nil, err
	}
	if err := c.PutURL(url, b); err != nil {
		return nil, fmt.Errorf("cache URL content: %w", err)
	}
	var cfg Ruleset
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal YAML: %w", err)
	}
	return &cfg, nil
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

func fetchGitHubContent(ctx context.Context, ghClient *gh.Client, owner, repo, path, ref string) (string, error) {
	content, err := ghClient.GetContent(ctx, owner, repo, path, ref)
	if err != nil {
		return "", fmt.Errorf("get GitHub content: %w", err)
	}
	return content, nil
}

func fetchURLContent(ctx context.Context, url string) ([]byte, error) {
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
	return b, nil
}

func ReadRulesets(ctx context.Context, logger *slog.Logger, ghClient *gh.Client, c *cache.Cache, dir string) ([]*Ruleset, error) {
	pattern := filepath.Join(dir, ".yamledit", "*.yaml")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("glob migration files: %w", err)
	}
	configs := make([]*Ruleset, 0, len(matches))
	for _, p := range matches {
		if filepath.Base(p) == "config.yaml" {
			continue
		}
		cfg, err := ReadRuleset(p)
		if err != nil {
			return nil, fmt.Errorf("read migration file %s: %w", p, err)
		}
		if err := ResolveImports(ctx, logger, ghClient, c, cfg); err != nil {
			return nil, fmt.Errorf("resolve imports in %s: %w", p, err)
		}
		configs = append(configs, cfg)
	}
	// Also load reusable rules from project config (global config is not included in default run)
	projCfg, err := ReadConfig(dir)
	if err != nil {
		return nil, fmt.Errorf("read project config: %w", err)
	}
	for _, rule := range projCfg.ReusableRules {
		cfg, err := resolveImport(ctx, logger, ghClient, c, rule.Import)
		if err != nil {
			return nil, fmt.Errorf("resolve reusable rule %s (%s): %w", rule.Name, rule.Import, err)
		}
		if err := ResolveImports(ctx, logger, ghClient, c, cfg); err != nil {
			return nil, fmt.Errorf("resolve imports in reusable rule %s: %w", rule.Name, err)
		}
		configs = append(configs, cfg)
	}
	return configs, nil
}

func ReadRulesetsByPaths(ctx context.Context, logger *slog.Logger, ghClient *gh.Client, c *cache.Cache, dir string, paths []string) ([]*Ruleset, error) {
	configs := make([]*Ruleset, 0, len(paths))
	for _, p := range paths {
		cfg, err := readConfigByPath(ctx, logger, ghClient, c, dir, p)
		if err != nil {
			return nil, err
		}
		configs = append(configs, cfg)
	}
	return configs, nil
}

func isRemoteImport(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") || strings.HasPrefix(s, "github.com/")
}

func readConfigByPath(ctx context.Context, logger *slog.Logger, ghClient *gh.Client, c *cache.Cache, dir, p string) (*Ruleset, error) {
	if isRemoteImport(p) {
		cfg, err := resolveImport(ctx, logger, ghClient, c, p)
		if err != nil {
			return nil, fmt.Errorf("fetch remote migration %s: %w", p, err)
		}
		if err := ResolveImports(ctx, logger, ghClient, c, cfg); err != nil {
			return nil, fmt.Errorf("resolve imports in %s: %w", p, err)
		}
		return cfg, nil
	}
	// Local path: strip ./ prefix if present
	origName := p
	p = strings.TrimPrefix(p, "./")
	isMigrationName := !filepath.IsAbs(p) && !yamlSuffixPattern.MatchString(p)
	if isMigrationName {
		p = filepath.Join(dir, ".yamledit", p+".yaml")
	}
	cfg, err := ReadRuleset(p)
	if err != nil {
		if isMigrationName && errors.Is(err, os.ErrNotExist) {
			// Fallback to reusable rule in project config
			return resolveReusableRule(ctx, logger, ghClient, c, dir, origName)
		}
		return nil, fmt.Errorf("read migration file %s: %w", p, err)
	}
	if err := ResolveImports(ctx, logger, ghClient, c, cfg); err != nil {
		return nil, fmt.Errorf("resolve imports in %s: %w", p, err)
	}
	return cfg, nil
}

func resolveReusableRule(ctx context.Context, logger *slog.Logger, ghClient *gh.Client, c *cache.Cache, dir, name string) (*Ruleset, error) {
	projCfg, err := ReadConfig(dir)
	if err != nil {
		return nil, fmt.Errorf("read project config: %w", err)
	}
	if target, ok := projCfg.FindReusableRule(name); ok {
		cfg, err := resolveImport(ctx, logger, ghClient, c, target)
		if err != nil {
			return nil, fmt.Errorf("resolve reusable rule %s (%s): %w", name, target, err)
		}
		if err := ResolveImports(ctx, logger, ghClient, c, cfg); err != nil {
			return nil, fmt.Errorf("resolve imports in reusable rule %s: %w", name, err)
		}
		return cfg, nil
	}
	// Fallback to global config
	globalCfg, err := ReadGlobalConfig()
	if err != nil {
		return nil, fmt.Errorf("read global config: %w", err)
	}
	if target, ok := globalCfg.FindReusableRule(name); ok {
		cfg, err := resolveImport(ctx, logger, ghClient, c, target)
		if err != nil {
			return nil, fmt.Errorf("resolve global reusable rule %s (%s): %w", name, target, err)
		}
		if err := ResolveImports(ctx, logger, ghClient, c, cfg); err != nil {
			return nil, fmt.Errorf("resolve imports in global reusable rule %s: %w", name, err)
		}
		return cfg, nil
	}
	return nil, fmt.Errorf("migration %q not found in .yamledit/, project config, or global config", name)
}

var yamlSuffixPattern = regexp.MustCompile(`\.ya?ml$`)

func ReadRuleset(p string) (*Ruleset, error) {
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	var cfg Ruleset
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal YAML: %w", err)
	}
	return &cfg, nil
}
