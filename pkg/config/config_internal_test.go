package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveGlobalConfigPath(t *testing.T) {
	// Cannot use t.Parallel() because subtests use t.Setenv
	tests := []struct {
		name string
		env  map[string]string
		want string
	}{
		{
			name: "YAMLEDIT_GLOBAL_CONFIG takes priority",
			env:  map[string]string{"YAMLEDIT_GLOBAL_CONFIG": "/custom/config.yaml"},
			want: "/custom/config.yaml",
		},
		{
			name: "XDG_CONFIG_HOME fallback",
			env:  map[string]string{"XDG_CONFIG_HOME": "/xdg/config"},
			want: "/xdg/config/yamledit/config.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Cannot use t.Parallel() with t.Setenv
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			if _, ok := tt.env["YAMLEDIT_GLOBAL_CONFIG"]; !ok {
				t.Setenv("YAMLEDIT_GLOBAL_CONFIG", "")
			}
			if _, ok := tt.env["XDG_CONFIG_HOME"]; !ok {
				t.Setenv("XDG_CONFIG_HOME", "")
			}
			got := resolveGlobalConfigPath()
			if got != tt.want {
				t.Errorf("resolveGlobalConfigPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestReadGlobalConfig(t *testing.T) {
	t.Run("file exists", func(t *testing.T) {
		dir := t.TempDir()
		configPath := filepath.Join(dir, "config.yaml")
		content := []byte("reusable_rules:\n  - name: test-rule\n    import: https://example.com/test\n")
		if err := os.WriteFile(configPath, content, 0o644); err != nil { //nolint:gosec
			t.Fatal(err)
		}
		t.Setenv("YAMLEDIT_GLOBAL_CONFIG", configPath)

		cfg, err := ReadGlobalConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cfg.ReusableRules) != 1 {
			t.Fatalf("expected 1 reusable rule, got %d", len(cfg.ReusableRules))
		}
		if cfg.ReusableRules[0].Name != "test-rule" {
			t.Errorf("expected name 'test-rule', got %q", cfg.ReusableRules[0].Name)
		}
	})

	t.Run("file not found", func(t *testing.T) {
		t.Setenv("YAMLEDIT_GLOBAL_CONFIG", "/nonexistent/config.yaml")
		t.Setenv("XDG_CONFIG_HOME", "")

		cfg, err := ReadGlobalConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cfg.ReusableRules) != 0 {
			t.Fatalf("expected empty config, got %d rules", len(cfg.ReusableRules))
		}
	})
}
