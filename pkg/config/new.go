package config

import (
	_ "embed"
	"fmt"
	"io"
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

func New(stderr io.Writer, dir, name string) error {
	pattern := regexp.MustCompile(`^[a-z0-9_-]+$`)
	if !pattern.MatchString(name) {
		return fmt.Errorf(`invalid migration name: '%s'. Must match pattern '%s'`, name, pattern.String())
	}
	configDir := filepath.Join(dir, ".yamledit")
	if err := os.MkdirAll(configDir, dirPermission); err != nil {
		return fmt.Errorf("create config directory .yamledit: %w", err)
	}
	files := []struct {
		path    string
		content []byte
	}{
		{filepath.Join(configDir, name+".yaml"), defaultConfig},
		{filepath.Join(configDir, name+"_test", "normal.yaml"), []byte("age: 10\n")},
		{filepath.Join(configDir, name+"_test", "normal_result.yaml"), []byte("age: 10\n")},
	}
	testDir := filepath.Join(configDir, name+"_test")
	if err := os.MkdirAll(testDir, dirPermission); err != nil {
		return fmt.Errorf("create test directory: %w", err)
	}
	for _, f := range files {
		created, err := writeFileIfNotExist(f.path, f.content)
		if err != nil {
			return fmt.Errorf("write file %s: %w", f.path, err)
		}
		if created {
			fmt.Fprintf(stderr, "%s was created.\n", f.path)
		}
	}
	return nil
}

func writeFileIfNotExist(p string, content []byte) (bool, error) {
	if _, err := os.Stat(p); err == nil {
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, fmt.Errorf("stat file: %w", err)
	}
	if err := os.WriteFile(p, content, filePermission); err != nil {
		return false, fmt.Errorf("write file: %w", err)
	}
	return true, nil
}
