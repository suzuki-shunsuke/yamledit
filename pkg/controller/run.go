package controller

import (
	"context"
	"fmt"
	"os"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/go-yamledit/yamledit"
	"github.com/suzuki-shunsuke/slog-util/slogutil"
	"github.com/suzuki-shunsuke/yamledit/pkg/config"
)

func Run(_ context.Context, logger *slogutil.Logger, dir string, yamlFiles []string) error {
	configs, err := config.ReadConfigs(dir)
	if err != nil {
		return fmt.Errorf("read migration configs: %w", err)
	}
	actions, err := buildActions(configs)
	if err != nil {
		return fmt.Errorf("build actions: %w", err)
	}
	for _, f := range yamlFiles {
		if err := editFile(logger, f, actions); err != nil {
			return fmt.Errorf("edit file %s: %w", f, err)
		}
	}
	return nil
}

func buildActions(configs []*config.Config) ([]yamledit.Action, error) {
	var actions []yamledit.Action
	for _, cfg := range configs {
		for _, rule := range cfg.Rules {
			for _, a := range rule.Actions {
				act, err := buildAction(rule.Path, a)
				if err != nil {
					return nil, err
				}
				actions = append(actions, act)
			}
		}
	}
	return actions, nil
}

func buildAction(path string, a *config.Action) (yamledit.Action, error) {
	switch a.Type {
	case "remove_keys":
		keys := make([]any, len(a.Keys))
		for i, k := range a.Keys {
			keys[i] = k
		}
		return yamledit.MapAction(path, yamledit.RemoveKeys(keys...)), nil
	default:
		return nil, fmt.Errorf("unsupported action type: %s", a.Type)
	}
}

func editFile(logger *slogutil.Logger, path string, actions []yamledit.Action) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}
	file, err := parser.ParseBytes(b, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse YAML: %w", err)
	}
	for _, act := range actions {
		if err := act.Run(file.Docs[0].Body); err != nil {
			return fmt.Errorf("run action: %w", err)
		}
	}
	result := file.String()
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
