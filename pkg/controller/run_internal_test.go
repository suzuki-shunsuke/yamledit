package controller

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/suzuki-shunsuke/slog-util/slogutil"
)

func newLogger() *slogutil.Logger {
	return slogutil.New(&slogutil.InputNew{
		Out: os.Stderr,
	})
}

func setupMigration(t *testing.T, dir, name, content string) {
	t.Helper()
	configDir := filepath.Join(dir, ".yamledit")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, name+".yaml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func setupYAMLFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestRun(t *testing.T) { //nolint:funlen
	t.Parallel()
	tests := []struct {
		name      string
		migration string
		input     string
		want      string
		wantErr   bool
	}{
		{
			name: "remove key from root",
			migration: `rules:
  - path: "$"
    actions:
      - type: remove_keys
        keys:
          - age
`,
			input: "name: alice\nage: 30\n",
			want:  "name: alice\n",
		},
		{
			name: "no changes needed",
			migration: `rules:
  - path: "$"
    actions:
      - type: remove_keys
        keys:
          - age
`,
			input: "name: alice\n",
			want:  "name: alice\n",
		},
		{
			name: "nested path",
			migration: `rules:
  - path: "$.foo"
    actions:
      - type: remove_keys
        keys:
          - bar
`,
			input: "foo:\n  bar: 1\n  baz: 2\n",
			want:  "foo:\n  baz: 2\n",
		},
		{
			name: "comments preserved",
			migration: `rules:
  - path: "$"
    actions:
      - type: remove_keys
        keys:
          - age
`,
			input: "name: alice # keep this\nage: 30\n",
			want:  "name: alice # keep this\n",
		},
		{
			name: "unsupported action type",
			migration: `rules:
  - path: "$"
    actions:
      - type: unknown
`,
			input:   "name: alice\n",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			setupMigration(t, dir, "test", tt.migration)
			yamlFile := setupYAMLFile(t, dir, "input.yaml", tt.input)

			err := Run(context.Background(), newLogger(), dir, []string{yamlFile})
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got, err := os.ReadFile(yamlFile)
			if err != nil {
				t.Fatalf("failed to read output file: %v", err)
			}
			if diff := cmp.Diff(tt.want, string(got)); diff != "" {
				t.Errorf("file content mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRun_multipleFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	setupMigration(t, dir, "test", `rules:
  - path: "$"
    actions:
      - type: remove_keys
        keys:
          - age
`)
	file1 := setupYAMLFile(t, dir, "a.yaml", "name: alice\nage: 30\n")
	file2 := setupYAMLFile(t, dir, "b.yaml", "name: bob\nage: 25\n")

	if err := Run(context.Background(), newLogger(), dir, []string{file1, file2}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, tc := range []struct {
		path string
		want string
	}{
		{file1, "name: alice\n"},
		{file2, "name: bob\n"},
	} {
		got, err := os.ReadFile(tc.path)
		if err != nil {
			t.Fatalf("failed to read %s: %v", tc.path, err)
		}
		if diff := cmp.Diff(tc.want, string(got)); diff != "" {
			t.Errorf("%s content mismatch (-want +got):\n%s", tc.path, diff)
		}
	}
}

func TestRun_nonexistentFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	setupMigration(t, dir, "test", `rules:
  - path: "$"
    actions:
      - type: remove_keys
        keys:
          - age
`)

	err := Run(context.Background(), newLogger(), dir, []string{filepath.Join(dir, "nonexistent.yaml")})
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}
