package config

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) { //nolint:funlen
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:  "valid lowercase name",
			input: "hello",
		},
		{
			name:  "valid name with hyphen",
			input: "my-migration",
		},
		{
			name:  "valid name with underscore",
			input: "my_migration",
		},
		{
			name:  "valid name with digits",
			input: "test123",
		},
		{
			name:    "invalid uppercase",
			input:   "Hello",
			wantErr: true,
		},
		{
			name:    "invalid space",
			input:   "hello world",
			wantErr: true,
		},
		{
			name:    "invalid empty",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid dot",
			input:   "hello.world",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			err := New(io.Discard, dir, tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			p := filepath.Join(dir, ".yamledit", tt.input, "ruleset.yaml")
			b, err := os.ReadFile(p)
			if err != nil {
				t.Fatalf("failed to read created file: %v", err)
			}
			if string(b) != string(defaultConfig) {
				t.Errorf("file content mismatch:\ngot:\n%s\nwant:\n%s", string(b), string(defaultConfig))
			}
			// Verify test files
			for _, testFile := range []string{"test.yaml", "result.yaml"} {
				tp := filepath.Join(dir, ".yamledit", tt.input, "normal", testFile)
				tb, err := os.ReadFile(tp)
				if err != nil {
					t.Fatalf("failed to read test file %s: %v", testFile, err)
				}
				if string(tb) != "age: 10\n" {
					t.Errorf("test file %s content mismatch:\ngot:\n%s\nwant:\n%s", testFile, string(tb), "age: 10\n")
				}
			}
		})
	}
}

func TestNew_idempotent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	name := "my-migration"

	if err := New(io.Discard, dir, name); err != nil {
		t.Fatalf("first call: %v", err)
	}

	files := []string{
		filepath.Join(dir, ".yamledit", name, "ruleset.yaml"),
		filepath.Join(dir, ".yamledit", name, "normal", "test.yaml"),
		filepath.Join(dir, ".yamledit", name, "normal", "result.yaml"),
	}
	infos := make([]os.FileInfo, len(files))
	for i, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			t.Fatalf("stat %s after first call: %v", f, err)
		}
		infos[i] = info
	}

	if err := New(io.Discard, dir, name); err != nil {
		t.Fatalf("second call: %v", err)
	}

	for i, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			t.Fatalf("stat %s after second call: %v", f, err)
		}
		if infos[i].ModTime() != info.ModTime() {
			t.Errorf("%s was modified on second call, expected idempotent behavior", f)
		}
	}
}

func TestNew_migrationExistsButTestFilesMissing(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	name := "my-migration"

	// Create only the migration file
	rulesetDir := filepath.Join(dir, ".yamledit", name)
	if err := os.MkdirAll(rulesetDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(rulesetDir, "ruleset.yaml"), defaultConfig, 0o644); err != nil { //nolint:gosec
		t.Fatal(err)
	}

	// Run New — should create test files even though migration exists
	if err := New(io.Discard, dir, name); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, testFile := range []string{"test.yaml", "result.yaml"} {
		tp := filepath.Join(rulesetDir, "normal", testFile)
		b, err := os.ReadFile(tp)
		if err != nil {
			t.Fatalf("failed to read test file %s: %v", testFile, err)
		}
		if string(b) != "age: 10\n" {
			t.Errorf("test file %s content mismatch:\ngot:\n%s\nwant:\n%s", testFile, string(b), "age: 10\n")
		}
	}
}
