package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

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

func TestReadConfigs(t *testing.T) { //nolint:funlen
	t.Parallel()
	tests := []struct {
		name    string
		setup   func(t *testing.T, dir string)
		want    []*Config
		wantErr bool
	}{
		{
			name: "valid single migration",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				setupMigration(t, dir, "foo", `rules:
  - path: "$"
    actions:
      - type: remove_keys
        keys:
          - age
`)
			},
			want: []*Config{
				{
					Rules: []*Rule{
						{
							Path: "$",
							Actions: []*Action{
								{
									Type: "remove_keys",
									Keys: []string{"age"},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "multiple migration files",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				setupMigration(t, dir, "aaa", `rules:
  - path: "$"
    actions:
      - type: remove_keys
        keys:
          - age
`)
				setupMigration(t, dir, "bbb", `rules:
  - path: "$.foo"
    actions:
      - type: remove_keys
        keys:
          - bar
`)
			},
			want: []*Config{
				{
					Rules: []*Rule{
						{
							Path: "$",
							Actions: []*Action{
								{
									Type: "remove_keys",
									Keys: []string{"age"},
								},
							},
						},
					},
				},
				{
					Rules: []*Rule{
						{
							Path: "$.foo",
							Actions: []*Action{
								{
									Type: "remove_keys",
									Keys: []string{"bar"},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "valid rename_key migration",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				setupMigration(t, dir, "rename", `rules:
  - path: "$"
    actions:
      - type: rename_key
        key: name
        new_key: first_name
`)
			},
			want: []*Config{
				{
					Rules: []*Rule{
						{
							Path: "$",
							Actions: []*Action{
								{
									Type:   "rename_key",
									Key:    "name",
									NewKey: "first_name",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "valid set_key migration",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				setupMigration(t, dir, "setkey", `rules:
  - path: "$"
    actions:
      - type: set_key
        key: name
        value: bob
        skip_if_key_not_found: true
        skip_if_key_found: false
        clear_comment: true
        insert_at:
          - after_key: id
          - before_key: age
          - first: true
`)
			},
			want: []*Config{
				{
					Rules: []*Rule{
						{
							Path: "$",
							Actions: []*Action{
								{
									Type:              "set_key",
									Key:               "name",
									Value:             "bob",
									SkipIfKeyNotFound: true,
									ClearComment:      true,
									InsertAt: []*InsertLocation{
										{AfterKey: "id"},
										{BeforeKey: "age"},
										{First: true},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "valid add_values migration",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				setupMigration(t, dir, "addvals", `rules:
  - path: "$"
    actions:
      - type: add_values
        values:
          - foo
          - bar
        index: 0
`)
			},
			want: []*Config{
				{
					Rules: []*Rule{
						{
							Path: "$",
							Actions: []*Action{
								{
									Type:   "add_values",
									Values: []any{"foo", "bar"},
									Index:  new(0),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "valid sort_key migration",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				setupMigration(t, dir, "sortkey", `rules:
  - path: "$"
    actions:
      - type: sort_key
        expr: "a.key < b.key ? -1 : (a.key > b.key ? 1 : 0)"
`)
			},
			want: []*Config{
				{
					Rules: []*Rule{
						{
							Path: "$",
							Actions: []*Action{
								{
									Type: "sort_key",
									Expr: `a.key < b.key ? -1 : (a.key > b.key ? 1 : 0)`,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "valid remove_values migration",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				setupMigration(t, dir, "removevals", `rules:
  - path: "$"
    actions:
      - type: remove_values
        expr: 'value.value == "foo"'
`)
			},
			want: []*Config{
				{
					Rules: []*Rule{
						{
							Path: "$",
							Actions: []*Action{
								{
									Type: "remove_values",
									Expr: `value.value == "foo"`,
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "no yamledit dir",
			setup: func(_ *testing.T, _ string) {},
			want:  []*Config{},
		},
		{
			name: "no migration files",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				if err := os.MkdirAll(filepath.Join(dir, ".yamledit"), 0o755); err != nil {
					t.Fatal(err)
				}
			},
			want: []*Config{},
		},
		{
			name: "invalid YAML",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				setupMigration(t, dir, "bad", `{invalid yaml`)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			tt.setup(t, dir)
			got, err := ReadConfigs(dir)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ReadConfigs() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
