package cli

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseArgs(t *testing.T) { //nolint:funlen
	t.Parallel()
	tests := []struct {
		name           string
		args           []string
		wantMigrations []string
		wantYAMLFiles  []string
		wantErr        bool
	}{
		{
			name:           "migration name",
			args:           []string{"@foo"},
			wantMigrations: []string{"foo"},
		},
		{
			name:           "migration file path",
			args:           []string{"@foo.yaml"},
			wantMigrations: []string{"foo.yaml"},
		},
		{
			name:           "migration file path yml",
			args:           []string{"@foo.yml"},
			wantMigrations: []string{"foo.yml"},
		},
		{
			name:           "migration file path with dir",
			args:           []string{"@foo/bar.yaml"},
			wantMigrations: []string{"foo/bar.yaml"},
		},
		{
			name:    "migration name with slash errors",
			args:    []string{"@foo/bar"},
			wantErr: true,
		},
		{
			name:           "https URL",
			args:           []string{"@https://example.com/migration.yaml"},
			wantMigrations: []string{"https://example.com/migration.yaml"},
		},
		{
			name:           "http URL",
			args:           []string{"@http://example.com/migration.yaml"},
			wantMigrations: []string{"http://example.com/migration.yaml"},
		},
		{
			name:           "GitHub with ref",
			args:           []string{"@github.com/owner/repo/path.yaml:v1.0.0"},
			wantMigrations: []string{"github.com/owner/repo/path.yaml:v1.0.0"},
		},
		{
			name:           "GitHub without ref",
			args:           []string{"@github.com/owner/repo/path.yaml"},
			wantMigrations: []string{"github.com/owner/repo/path.yaml"},
		},
		{
			name:           "local path escape with ./",
			args:           []string{"@./github.com/local/file.yaml"},
			wantMigrations: []string{"./github.com/local/file.yaml"},
		},
		{
			name:           "local path escape generic",
			args:           []string{"@./some/path.yaml"},
			wantMigrations: []string{"./some/path.yaml"},
		},
		{
			name:          "yaml file",
			args:          []string{"input.yaml"},
			wantYAMLFiles: []string{"input.yaml"},
		},
		{
			name:           "mixed args",
			args:           []string{"@mig1", "a.yaml", "@mig2.yaml", "b.yml"},
			wantMigrations: []string{"mig1", "mig2.yaml"},
			wantYAMLFiles:  []string{"a.yaml", "b.yml"},
		},
		{
			name: "no args",
			args: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			migrations, yamlFiles, err := parseArgs(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tt.wantMigrations, migrations); diff != "" {
				t.Errorf("migrations mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantYAMLFiles, yamlFiles); diff != "" {
				t.Errorf("yamlFiles mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
