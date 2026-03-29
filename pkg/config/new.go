package config

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

//go:embed new.yaml
var defaultConfig []byte

const (
	filePermission = 0o644
	dirPermission  = 0o755
)

func New(dir, name string) error {
	pattern := regexp.MustCompile(`^[a-z0-9_-]+$`)
	if !pattern.MatchString(name) {
		return fmt.Errorf(`invalid migration name: '%s'. Must match pattern '%s'`, name, pattern.String())
	}
	configDir := filepath.Join(dir, ".yamledit")
	if err := os.MkdirAll(configDir, dirPermission); err != nil {
		return fmt.Errorf("create config directory .yamledit: %w", err)
	}
	if err := writeFileIfNotExist(filepath.Join(configDir, name+".yaml"), defaultConfig); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}
	testContent := []byte("age: 10\n")
	testDir := filepath.Join(configDir, name+"_test")
	if err := os.MkdirAll(testDir, dirPermission); err != nil {
		return fmt.Errorf("create test directory: %w", err)
	}
	if err := writeFileIfNotExist(filepath.Join(testDir, "normal.yaml"), testContent); err != nil {
		return fmt.Errorf("write test file normal.yaml: %w", err)
	}
	if err := writeFileIfNotExist(filepath.Join(testDir, "normal_result.yaml"), testContent); err != nil {
		return fmt.Errorf("write test file normal_result.yaml: %w", err)
	}
	return nil
}

func writeFileIfNotExist(p string, content []byte) error {
	if _, err := os.Stat(p); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat file: %w", err)
	}
	if err := os.WriteFile(p, content, filePermission); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}
