package config

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

//go:embed init.yaml
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
	p := filepath.Join(configDir, name+".yaml")
	if _, err := os.Stat(p); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat config file: %w", err)
	}
	if err := os.MkdirAll(configDir, dirPermission); err != nil {
		return fmt.Errorf("create config directory .yamledit: %w", err)
	}
	if err := os.WriteFile(p, defaultConfig, filePermission); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}
	return nil
}
