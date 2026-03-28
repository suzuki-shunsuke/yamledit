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

func New(name string) error {
	pattern := regexp.MustCompile(`^[a-z0-9_-]+$`)
	if !pattern.MatchString(name) {
		return fmt.Errorf(`invalid name: "%s". Must match pattern "%s"`, name, pattern.String())
	}
	p := filepath.Join(".yamledit", name+".yaml")
	if _, err := os.Stat(p); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat config file: %w", err)
	}
	if err := os.MkdirAll(".yamledit", dirPermission); err != nil {
		return err
	}
	if err := os.WriteFile(p, defaultConfig, filePermission); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}
	return nil
}
