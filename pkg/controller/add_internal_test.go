package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdd(t *testing.T) { //nolint:funlen,gocognit,cyclop
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`rules: []`)) //nolint:errcheck
	}))
	t.Cleanup(srv.Close)

	tests := []struct {
		name            string
		initialConfig   string // "" = no file, "no-dir" = don't create .yamledit
		alias           string
		migration       string // "$URL" replaced with srv.URL
		force           bool
		wantErr         bool
		wantContains    []string // "$URL" replaced with srv.URL
		wantNotContains []string
	}{
		{
			name:         "create new config",
			alias:        "my-rule",
			migration:    "$URL/migration.yaml",
			wantContains: []string{"name: my-rule", "import: $URL/migration.yaml"},
		},
		{
			name:          "add to existing config",
			initialConfig: "reusable_rules:\n  - name: existing\n    import: https://example.com/existing\n",
			alias:         "new-rule",
			migration:     "$URL/migration.yaml",
			wantContains:  []string{"name: existing", "name: new-rule"},
		},
		{
			name:          "duplicate name without force",
			initialConfig: "reusable_rules:\n  - name: my-rule\n    import: https://example.com/old\n",
			alias:         "my-rule",
			migration:     "https://example.com/new",
			wantErr:       true,
		},
		{
			name:            "force overwrite preserves order",
			initialConfig:   "reusable_rules:\n  - name: first\n    import: https://example.com/first\n  - name: my-rule\n    import: https://example.com/old\n  - name: last\n    import: https://example.com/last\n",
			alias:           "my-rule",
			migration:       "$URL/new.yaml",
			force:           true,
			wantContains:    []string{"$URL/new.yaml", "name: first", "name: last"},
			wantNotContains: []string{"https://example.com/old"},
		},
		{
			name:          "preserves comments on add",
			initialConfig: "# my config\nreusable_rules:\n  - name: existing # keep this\n    import: https://example.com/existing\n",
			alias:         "new-rule",
			migration:     "$URL/migration.yaml",
			wantContains:  []string{"# my config", "# keep this", "name: new-rule"},
		},
		{
			name:            "force preserves comments",
			initialConfig:   "# my config\nreusable_rules:\n  - name: my-rule # keep this\n    import: https://example.com/old\n",
			alias:           "my-rule",
			migration:       "$URL/new.yaml",
			force:           true,
			wantContains:    []string{"# my config", "# keep this", "$URL/new.yaml"},
			wantNotContains: []string{"https://example.com/old"},
		},
		{
			name:      "invalid migration prefix",
			alias:     "my-rule",
			migration: "ftp://example.com/migration.yaml",
			wantErr:   true,
		},
		{
			name:      "invalid alias",
			alias:     "My-Rule",
			migration: "https://example.com/migration.yaml",
			wantErr:   true,
		},
		{
			name:          "creates .yamledit dir",
			initialConfig: "no-dir",
			alias:         "my-rule",
			migration:     "$URL/migration.yaml",
			wantContains:  []string{"name: my-rule"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			migration := strings.ReplaceAll(tt.migration, "$URL", srv.URL)

			if tt.initialConfig != "no-dir" {
				configDir := filepath.Join(dir, ".yamledit")
				if err := os.MkdirAll(configDir, 0o755); err != nil {
					t.Fatal(err)
				}
				if tt.initialConfig != "" {
					if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(tt.initialConfig), 0o644); err != nil { //nolint:gosec
						t.Fatal(err)
					}
				}
			}

			err := Add(context.Background(), newLogger(), nil, nil, dir, tt.alias, migration, tt.force)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			b, err := os.ReadFile(filepath.Join(dir, ".yamledit", "config.yaml"))
			if err != nil {
				t.Fatalf("failed to read config: %v", err)
			}
			got := string(b)
			for _, want := range tt.wantContains {
				want = strings.ReplaceAll(want, "$URL", srv.URL)
				if !strings.Contains(got, want) {
					t.Errorf("expected %q in output, got:\n%s", want, got)
				}
			}
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(got, notWant) {
					t.Errorf("unexpected %q in output, got:\n%s", notWant, got)
				}
			}
		})
	}
}
