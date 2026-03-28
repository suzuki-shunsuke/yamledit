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
	if err := os.WriteFile(filepath.Join(configDir, name+".yaml"), []byte(content), 0o644); err != nil { //nolint:gosec
		t.Fatal(err)
	}
}

func setupYAMLFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil { //nolint:gosec
		t.Fatal(err)
	}
	return p
}

func TestRun(t *testing.T) { //nolint:funlen,maintidx
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
			name: "rename key at root",
			migration: `rules:
  - path: "$"
    actions:
      - type: rename_key
        key: name
        new_key: first_name
`,
			input: "name: alice\nage: 30\n",
			want:  "first_name: alice\nage: 30\n",
		},
		{
			name: "rename key preserves comment",
			migration: `rules:
  - path: "$"
    actions:
      - type: rename_key
        key: name
        new_key: first_name
`,
			input: "name: alice # important\nage: 30\n",
			want:  "first_name: alice # important\nage: 30\n",
		},
		{
			name: "rename key not found",
			migration: `rules:
  - path: "$"
    actions:
      - type: rename_key
        key: missing
        new_key: new_key
`,
			input: "name: alice\n",
			want:  "name: alice\n",
		},
		{
			name: "rename key nested path",
			migration: `rules:
  - path: "$.foo"
    actions:
      - type: rename_key
        key: bar
        new_key: baz
`,
			input: "foo:\n  bar: 1\n  qux: 2\n",
			want:  "foo:\n  baz: 1\n  qux: 2\n",
		},
		{
			name: "rename key ignore_existing_key",
			migration: `rules:
  - path: "$"
    actions:
      - type: rename_key
        key: name
        new_key: first_name
        when_duplicate: ignore_existing_key
`,
			input: "name: foo\nfirst_name: bar\n",
			want:  "first_name: foo\n",
		},
		{
			name: "rename key remove_old_key",
			migration: `rules:
  - path: "$"
    actions:
      - type: rename_key
        key: name
        new_key: first_name
        when_duplicate: remove_old_key
`,
			input: "name: foo\nfirst_name: bar\n",
			want:  "first_name: bar\n",
		},
		{
			name: "rename key fail on duplicate",
			migration: `rules:
  - path: "$"
    actions:
      - type: rename_key
        key: name
        new_key: first_name
        when_duplicate: fail
`,
			input:   "name: foo\nfirst_name: bar\n",
			wantErr: true,
		},
		{
			name: "rename key skip on duplicate",
			migration: `rules:
  - path: "$"
    actions:
      - type: rename_key
        key: name
        new_key: first_name
        when_duplicate: skip
`,
			input: "name: foo\nfirst_name: bar\n",
			want:  "name: foo\nfirst_name: bar\n",
		},
		{
			name: "set_key existing key",
			migration: `rules:
  - path: "$"
    actions:
      - type: set_key
        key: name
        value: bob
`,
			input: "name: alice\nage: 30\n",
			want:  "name: bob\nage: 30\n",
		},
		{
			name: "set_key add new key",
			migration: `rules:
  - path: "$"
    actions:
      - type: set_key
        key: age
        value: 25
`,
			input: "name: alice\n",
			want:  "name: alice\nage: 25\n",
		},
		{
			name: "set_key skip if key not found",
			migration: `rules:
  - path: "$"
    actions:
      - type: set_key
        key: age
        value: 25
        skip_if_key_not_found: true
`,
			input: "name: alice\n",
			want:  "name: alice\n",
		},
		{
			name: "set_key skip if key found",
			migration: `rules:
  - path: "$"
    actions:
      - type: set_key
        key: name
        value: bob
        skip_if_key_found: true
`,
			input: "name: alice\n",
			want:  "name: alice\n",
		},
		{
			name: "set_key insert before key",
			migration: `rules:
  - path: "$"
    actions:
      - type: set_key
        key: gender
        value: male
        insert_at:
          - before_key: age
`,
			input: "name: alice\nage: 30\n",
			want:  "name: alice\ngender: male\nage: 30\n",
		},
		{
			name: "set_key insert after key",
			migration: `rules:
  - path: "$"
    actions:
      - type: set_key
        key: gender
        value: male
        insert_at:
          - after_key: name
`,
			input: "name: alice\nage: 30\n",
			want:  "name: alice\ngender: male\nage: 30\n",
		},
		{
			name: "set_key insert first",
			migration: `rules:
  - path: "$"
    actions:
      - type: set_key
        key: id
        value: 1
        insert_at:
          - first: true
`,
			input: "name: alice\nage: 30\n",
			want:  "id: 1\nname: alice\nage: 30\n",
		},
		{
			name: "set_key clear comment",
			migration: `rules:
  - path: "$"
    actions:
      - type: set_key
        key: name
        value: bob
        clear_comment: true
`,
			input: "name: alice # important\nage: 30\n",
			want:  "name: bob\nage: 30\n",
		},
		{
			name: "set_key nested path",
			migration: `rules:
  - path: "$.foo"
    actions:
      - type: set_key
        key: bar
        value: 99
`,
			input: "foo:\n  bar: 1\n  baz: 2\n",
			want:  "foo:\n  bar: 99\n  baz: 2\n",
		},
		{
			name: "add_values append to end",
			migration: `rules:
  - path: "$"
    actions:
      - type: add_values
        values:
          - baz
`,
			input: "- foo\n- bar\n",
			want:  "- foo\n- bar\n- baz\n",
		},
		{
			name: "add_values insert at beginning",
			migration: `rules:
  - path: "$"
    actions:
      - type: add_values
        values:
          - first
        index: 0
`,
			input: "- foo\n- bar\n",
			want:  "- first\n- foo\n- bar\n",
		},
		{
			name: "add_values nested path",
			migration: `rules:
  - path: "$.items"
    actions:
      - type: add_values
        values:
          - c
`,
			input: "items:\n  - a\n  - b\n",
			want:  "items:\n  - a\n  - b\n  - c\n",
		},
		{
			name: "sort_key alphabetical",
			migration: `rules:
  - path: "$"
    actions:
      - type: sort_key
        expr: "a.key < b.key ? -1 : (a.key > b.key ? 1 : 0)"
`,
			input: "c: 3\na: 1\nb: 2\n",
			want:  "a: 1\nb: 2\nc: 3\n",
		},
		{
			name: "sort_key already sorted",
			migration: `rules:
  - path: "$"
    actions:
      - type: sort_key
        expr: "a.key < b.key ? -1 : (a.key > b.key ? 1 : 0)"
`,
			input: "a: 1\nb: 2\nc: 3\n",
			want:  "a: 1\nb: 2\nc: 3\n",
		},
		{
			name: "sort_key nested path",
			migration: `rules:
  - path: "$.foo"
    actions:
      - type: sort_key
        expr: "a.key < b.key ? -1 : (a.key > b.key ? 1 : 0)"
`,
			input: "foo:\n  c: 3\n  a: 1\n  b: 2\n",
			want:  "foo:\n  a: 1\n  b: 2\n  c: 3\n",
		},
		{
			name: "remove_values matching elements",
			migration: `rules:
  - path: "$"
    actions:
      - type: remove_values
        expr: 'value.value == "bar"'
`,
			input: "- foo\n- bar\n- baz\n",
			want:  "- foo\n- baz\n",
		},
		{
			name: "remove_values no matches",
			migration: `rules:
  - path: "$"
    actions:
      - type: remove_values
        expr: 'value.value == "missing"'
`,
			input: "- foo\n- bar\n",
			want:  "- foo\n- bar\n",
		},
		{
			name: "remove_values nested path",
			migration: `rules:
  - path: "$.items"
    actions:
      - type: remove_values
        expr: 'value.value == "b"'
`,
			input: "items:\n  - a\n  - b\n  - c\n",
			want:  "items:\n  - a\n  - c\n",
		},
		{
			name: "sort_list alphabetical",
			migration: `rules:
  - path: "$"
    actions:
      - type: sort_list
        expr: "a.value < b.value ? -1 : (a.value > b.value ? 1 : 0)"
`,
			input: "- cherry\n- apple\n- banana\n",
			want:  "- apple\n- banana\n- cherry\n",
		},
		{
			name: "sort_list already sorted",
			migration: `rules:
  - path: "$"
    actions:
      - type: sort_list
        expr: "a.value < b.value ? -1 : (a.value > b.value ? 1 : 0)"
`,
			input: "- apple\n- banana\n- cherry\n",
			want:  "- apple\n- banana\n- cherry\n",
		},
		{
			name: "sort_list nested path",
			migration: `rules:
  - path: "$.items"
    actions:
      - type: sort_list
        expr: "a.value < b.value ? -1 : (a.value > b.value ? 1 : 0)"
`,
			input: "items:\n  - c\n  - a\n  - b\n",
			want:  "items:\n  - a\n  - b\n  - c\n",
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
