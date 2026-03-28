package controller

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/expr-lang/expr"
	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/go-yamledit/yamledit"
	"github.com/suzuki-shunsuke/slog-util/slogutil"
	"github.com/suzuki-shunsuke/yamledit/pkg/config"
	gh "github.com/suzuki-shunsuke/yamledit/pkg/github"
)

type ruleActions struct {
	files   []string
	actions []yamledit.Action
}

func Run(ctx context.Context, logger *slogutil.Logger, ghClient *gh.Client, dir string, migrations, yamlFiles []string) error {
	configs, err := loadConfigs(ctx, ghClient, dir, migrations)
	if err != nil {
		return fmt.Errorf("read migration configs: %w", err)
	}
	rules, err := buildRuleActions(logger, configs)
	if err != nil {
		return fmt.Errorf("build actions: %w", err)
	}
	if len(yamlFiles) == 0 {
		yamlFiles, err = discoverYAMLFiles(dir)
		if err != nil {
			return fmt.Errorf("discover YAML files: %w", err)
		}
	}
	for _, f := range yamlFiles {
		for _, rule := range rules {
			matched, err := matchFile(f, rule.files)
			if err != nil {
				return fmt.Errorf("match file %s: %w", f, err)
			}
			if !matched {
				continue
			}
			if err := editFile(logger, f, rule.actions); err != nil {
				return fmt.Errorf("edit file %s: %w", f, err)
			}
		}
	}
	return nil
}

func loadConfigs(ctx context.Context, ghClient *gh.Client, dir string, migrations []string) ([]*config.Config, error) {
	if len(migrations) == 0 {
		configs, err := config.ReadConfigs(ctx, ghClient, dir)
		if err != nil {
			return nil, fmt.Errorf("read all configs: %w", err)
		}
		return configs, nil
	}
	configs, err := config.ReadConfigsByPaths(ctx, ghClient, dir, migrations)
	if err != nil {
		return nil, fmt.Errorf("read configs by paths: %w", err)
	}
	return configs, nil
}

func discoverYAMLFiles(dir string) ([]string, error) {
	var files []string
	for _, pattern := range []string{"**/*.yaml", "**/*.yml"} {
		matches, err := doublestar.Glob(os.DirFS(dir), pattern)
		if err != nil {
			return nil, fmt.Errorf("glob %s: %w", pattern, err)
		}
		files = append(files, matches...)
	}
	return files, nil
}

func buildRuleActions(logger *slogutil.Logger, configs []*config.Config) ([]*ruleActions, error) {
	var rules []*ruleActions
	for _, cfg := range configs {
		for _, rule := range cfg.Rules {
			var actions []yamledit.Action
			for _, a := range rule.Actions {
				act, err := buildAction(logger, rule.Path, a)
				if err != nil {
					return nil, err
				}
				actions = append(actions, act)
			}
			rules = append(rules, &ruleActions{
				files:   rule.Files,
				actions: actions,
			})
		}
	}
	return rules, nil
}

func matchFile(file string, patterns []string) (bool, error) {
	if len(patterns) == 0 {
		return true, nil
	}
	matched := false
	for _, p := range patterns {
		exclude := strings.HasPrefix(p, "!")
		pattern := p
		if exclude {
			pattern = p[1:]
		}
		ok, err := doublestar.Match(pattern, file)
		if err != nil {
			return false, fmt.Errorf("match pattern %s: %w", p, err)
		}
		if ok {
			matched = !exclude
		}
	}
	return matched, nil
}

func buildAction(logger *slogutil.Logger, path string, a *config.Action) (yamledit.Action, error) {
	switch a.Type {
	case "remove_keys":
		keys := make([]any, len(a.Keys))
		for i, k := range a.Keys {
			keys[i] = k
		}
		return yamledit.MapAction(path, yamledit.RemoveKeys(keys...)), nil
	case "rename_key":
		return buildRenameKeyAction(path, a)
	case "set_key":
		return buildSetKeyAction(path, a), nil
	case "add_values":
		return buildAddValuesAction(path, a), nil
	case "sort_key":
		return buildSortKeyAction(logger, path, a)
	case "remove_values":
		return buildRemoveValuesAction(logger, path, a)
	case "sort_list":
		return buildSortListAction(logger, path, a)
	default:
		return nil, fmt.Errorf("unsupported action type: %s", a.Type)
	}
}

func buildRenameKeyAction(path string, a *config.Action) (yamledit.Action, error) {
	wd, err := parseWhenDuplicate(a.WhenDuplicate)
	if err != nil {
		return nil, err
	}
	return yamledit.MapAction(path, yamledit.RenameKey(a.Key, a.NewKey, wd)), nil
}

func buildSetKeyAction(path string, a *config.Action) yamledit.Action {
	opt := &yamledit.SetKeyOption{
		IgnoreIfKeyNotExist: a.SkipIfKeyNotFound,
		IgnoreIfKeyExist:    a.SkipIfKeyFound,
		ClearComment:        a.ClearComment,
	}
	for _, loc := range a.InsertAt {
		yloc := &yamledit.InsertLocation{First: loc.First}
		if loc.AfterKey != "" {
			yloc.AfterKey = loc.AfterKey
		}
		if loc.BeforeKey != "" {
			yloc.BeforeKey = loc.BeforeKey
		}
		opt.InsertLocations = append(opt.InsertLocations, yloc)
	}
	return yamledit.MapAction(path, yamledit.SetKey(a.Key, a.Value, opt))
}

func buildSortKeyAction(logger *slogutil.Logger, path string, a *config.Action) (yamledit.Action, error) {
	program, err := expr.Compile(a.Expr, expr.Env(map[string]any{
		"a": map[string]any{},
		"b": map[string]any{},
	}))
	if err != nil {
		return nil, fmt.Errorf("compile sort_key expr: %w", err)
	}
	return yamledit.MapAction(path, yamledit.SortKey(func(a, b *yamledit.KeyValue[any]) int {
		env := map[string]any{
			"a": map[string]any{"key": a.Key, "value": a.Value, "comment": a.Comment, "index": a.Index},
			"b": map[string]any{"key": b.Key, "value": b.Value, "comment": b.Comment, "index": b.Index},
		}
		result, err := expr.Run(program, env)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 0
		}
		i, ok := result.(int)
		if !ok {
			logger.Error("sort_key expr must return int", "got_type", fmt.Sprintf("%T", result))
			return 0
		}
		return i
	})), nil
}

func buildRemoveValuesAction(logger *slogutil.Logger, path string, a *config.Action) (yamledit.Action, error) {
	program, err := expr.Compile(a.Expr, expr.Env(map[string]any{
		"value": map[string]any{},
	}))
	if err != nil {
		return nil, fmt.Errorf("compile remove_values expr: %w", err)
	}
	return yamledit.ListAction(path, yamledit.RemoveValuesFromList(func(node *yamledit.Node[any]) (bool, error) {
		env := map[string]any{
			"value": map[string]any{"value": node.Value, "comment": node.Comment},
		}
		result, err := expr.Run(program, env)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return false, nil
		}
		b, ok := result.(bool)
		if !ok {
			logger.Error("remove_values expr must return bool", "got_type", fmt.Sprintf("%T", result))
			return false, nil
		}
		return b, nil
	})), nil
}

func buildSortListAction(logger *slogutil.Logger, path string, a *config.Action) (yamledit.Action, error) {
	program, err := expr.Compile(a.Expr, expr.Env(map[string]any{
		"a": map[string]any{},
		"b": map[string]any{},
	}))
	if err != nil {
		return nil, fmt.Errorf("compile sort_list expr: %w", err)
	}
	return yamledit.ListAction(path, yamledit.SortList(func(a, b *yamledit.Node[any]) int {
		env := map[string]any{
			"a": map[string]any{"value": a.Value, "comment": a.Comment},
			"b": map[string]any{"value": b.Value, "comment": b.Comment},
		}
		result, err := expr.Run(program, env)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 0
		}
		i, ok := result.(int)
		if !ok {
			logger.Error("sort_list expr must return int", "got_type", fmt.Sprintf("%T", result))
			return 0
		}
		return i
	})), nil
}

func buildAddValuesAction(path string, a *config.Action) yamledit.Action {
	idx := -1
	if a.Index != nil {
		idx = *a.Index
	}
	values := make([]any, len(a.Values))
	copy(values, a.Values)
	return yamledit.ListAction(path, yamledit.AddValuesToList(idx, values...))
}

func parseWhenDuplicate(s string) (yamledit.WhenDuplicateKey, error) {
	switch s {
	case "", "skip":
		return yamledit.Skip, nil
	case "ignore_existing_key":
		return yamledit.IgnoreExistingKey, nil
	case "remove_old_key":
		return yamledit.RemoveOldKey, nil
	case "fail":
		return yamledit.RaiseError, nil
	default:
		return 0, fmt.Errorf("unsupported when_duplicate value: %s", s)
	}
}

func applyActions(b []byte, actions []yamledit.Action) (string, error) {
	file, err := parser.ParseBytes(b, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("parse YAML: %w", err)
	}
	for _, act := range actions {
		if err := act.Run(file.Docs[0].Body); err != nil {
			return "", fmt.Errorf("run action: %w", err)
		}
	}
	return file.String(), nil
}

func editFile(logger *slogutil.Logger, path string, actions []yamledit.Action) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}
	result, err := applyActions(b, actions)
	if err != nil {
		return err
	}
	if result == string(b) {
		logger.Info("no changes", "file", path)
		return nil
	}
	if err := os.WriteFile(path, []byte(result), 0o644); err != nil { //nolint:gosec,mnd
		return fmt.Errorf("write file: %w", err)
	}
	logger.Info("updated", "file", path)
	return nil
}
