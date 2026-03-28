package config

import (
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
			err := New(dir, tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			p := filepath.Join(dir, ".yamledit", tt.input+".yaml")
			b, err := os.ReadFile(p)
			if err != nil {
				t.Fatalf("failed to read created file: %v", err)
			}
			if string(b) != string(defaultConfig) {
				t.Errorf("file content mismatch:\ngot:\n%s\nwant:\n%s", string(b), string(defaultConfig))
			}
		})
	}
}

func TestNew_idempotent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	name := "my-migration"

	if err := New(dir, name); err != nil {
		t.Fatalf("first call: %v", err)
	}

	p := filepath.Join(dir, ".yamledit", name+".yaml")
	info1, err := os.Stat(p)
	if err != nil {
		t.Fatalf("stat after first call: %v", err)
	}

	if err := New(dir, name); err != nil {
		t.Fatalf("second call: %v", err)
	}

	info2, err := os.Stat(p)
	if err != nil {
		t.Fatalf("stat after second call: %v", err)
	}

	if info1.ModTime() != info2.ModTime() {
		t.Error("file was modified on second call, expected idempotent behavior")
	}
}
