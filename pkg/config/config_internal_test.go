package config

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const testRemoteConfig = `rules:
  - path: "$"
    actions:
      - type: remove_keys
        keys:
          - age
`

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

func TestReadConfigs(t *testing.T) { //nolint:funlen,maintidx
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
			name: "valid sort_list migration",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				setupMigration(t, dir, "sortlist", `rules:
  - path: "$"
    actions:
      - type: sort_list
        expr: "a.value < b.value ? -1 : (a.value > b.value ? 1 : 0)"
`)
			},
			want: []*Config{
				{
					Rules: []*Rule{
						{
							Path: "$",
							Actions: []*Action{
								{
									Type: "sort_list",
									Expr: `a.value < b.value ? -1 : (a.value > b.value ? 1 : 0)`,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "rule with files",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				setupMigration(t, dir, "withfiles", `rules:
  - path: "$"
    files:
      - "**/*.yaml"
      - "!vendor/**"
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
							Path:  "$",
							Files: []string{"**/*.yaml", "!vendor/**"},
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
			got, err := ReadConfigs(context.Background(), slog.Default(), nil, nil, dir)
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

func TestResolveImports(t *testing.T) {
	t.Parallel()
	remoteConfig := testRemoteConfig
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(remoteConfig)) //nolint:errcheck
	}))
	t.Cleanup(srv.Close)

	cfg := &Config{
		Rules: []*Rule{
			{
				Path: "$",
				Actions: []*Action{
					{Type: "remove_keys", Keys: []string{"name"}},
				},
			},
			{
				Import: srv.URL + "/migration.yaml",
			},
		},
	}
	if err := ResolveImports(context.Background(), slog.Default(), nil, nil, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []*Rule{
		{
			Path: "$",
			Actions: []*Action{
				{Type: "remove_keys", Keys: []string{"name"}},
			},
		},
		{
			Path: "$",
			Actions: []*Action{
				{Type: "remove_keys", Keys: []string{"age"}},
			},
		},
	}
	if diff := cmp.Diff(want, cfg.Rules); diff != "" {
		t.Errorf("ResolveImports() mismatch (-want +got):\n%s", diff)
	}
}

func TestReadConfigsByPaths_remoteURL(t *testing.T) {
	t.Parallel()
	remoteConfig := testRemoteConfig
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(remoteConfig)) //nolint:errcheck
	}))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	configs, err := ReadConfigsByPaths(context.Background(), slog.Default(), nil, nil, dir, []string{srv.URL + "/migration.yaml"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(configs) != 1 {
		t.Fatalf("expected 1 config, got %d", len(configs))
	}
	want := []*Rule{
		{
			Path: "$",
			Actions: []*Action{
				{Type: "remove_keys", Keys: []string{"age"}},
			},
		},
	}
	if diff := cmp.Diff(want, configs[0].Rules); diff != "" {
		t.Errorf("ReadConfigsByPaths() mismatch (-want +got):\n%s", diff)
	}
}

func TestReadConfigsByPaths_localPathEscape(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	setupMigration(t, dir, "test", `rules:
  - path: "$"
    actions:
      - type: remove_keys
        keys:
          - age
`)
	configs, err := ReadConfigsByPaths(context.Background(), slog.Default(), nil, nil, dir, []string{"./" + filepath.Join(dir, ".yamledit", "test.yaml")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(configs) != 1 {
		t.Fatalf("expected 1 config, got %d", len(configs))
	}
	want := []*Rule{
		{
			Path: "$",
			Actions: []*Action{
				{Type: "remove_keys", Keys: []string{"age"}},
			},
		},
	}
	if diff := cmp.Diff(want, configs[0].Rules); diff != "" {
		t.Errorf("ReadConfigsByPaths() mismatch (-want +got):\n%s", diff)
	}
}

func TestReadConfigs_skipsConfigYAML(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// config.yaml with no reusable rules — should not be loaded as a migration
	setupMigration(t, dir, "config", "reusable_rules: []\n")
	setupMigration(t, dir, "foo", `rules:
  - path: "$"
    actions:
      - type: remove_keys
        keys:
          - age
`)
	got, err := ReadConfigs(context.Background(), slog.Default(), nil, nil, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 config (config.yaml should be skipped), got %d", len(got))
	}
}

func TestReadConfigs_withReusableRules(t *testing.T) {
	t.Parallel()
	remoteConfig := `rules:
  - path: "$"
    actions:
      - type: remove_keys
        keys:
          - name
`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(remoteConfig)) //nolint:errcheck
	}))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	// Create a local migration
	setupMigration(t, dir, "local", `rules:
  - path: "$"
    actions:
      - type: remove_keys
        keys:
          - age
`)
	// Create config.yaml with a reusable rule
	setupMigration(t, dir, "config", "reusable_rules:\n  - name: remote-rule\n    import: "+srv.URL+"/migration.yaml\n")

	got, err := ReadConfigs(context.Background(), slog.Default(), nil, nil, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 configs (1 local + 1 reusable rule), got %d", len(got))
	}
}

func TestReadConfigsByPaths_reusableRuleFallback(t *testing.T) {
	t.Parallel()
	remoteConfig := testRemoteConfig
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(remoteConfig)) //nolint:errcheck
	}))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	// Create config.yaml with a reusable rule but no local migration file
	setupMigration(t, dir, "config", "reusable_rules:\n  - name: my-rule\n    import: "+srv.URL+"/migration.yaml\n")

	configs, err := ReadConfigsByPaths(context.Background(), slog.Default(), nil, nil, dir, []string{"my-rule"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(configs) != 1 {
		t.Fatalf("expected 1 config, got %d", len(configs))
	}
	want := []*Rule{
		{
			Path: "$",
			Actions: []*Action{
				{Type: "remove_keys", Keys: []string{"age"}},
			},
		},
	}
	if diff := cmp.Diff(want, configs[0].Rules); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
}

func TestReadConfigsByPaths_localOverReusableRule(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// Create both a local migration and a reusable rule with the same name
	setupMigration(t, dir, "my-rule", `rules:
  - path: "$"
    actions:
      - type: remove_keys
        keys:
          - age
`)
	setupMigration(t, dir, "config", "reusable_rules:\n  - name: my-rule\n    import: https://example.com/should-not-be-used\n")

	configs, err := ReadConfigsByPaths(context.Background(), slog.Default(), nil, nil, dir, []string{"my-rule"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(configs) != 1 {
		t.Fatalf("expected 1 config, got %d", len(configs))
	}
}

func TestReadConfigsByPaths_reusableRuleNotFound(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// Create config.yaml without the requested reusable rule
	setupMigration(t, dir, "config", "reusable_rules:\n  - name: other\n    import: https://example.com/other\n")

	_, err := ReadConfigsByPaths(context.Background(), slog.Default(), nil, nil, dir, []string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error for nonexistent migration/reusable rule, got nil")
	}
}

func TestResolveImports_noImport(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		Rules: []*Rule{
			{Path: "$", Actions: []*Action{{Type: "remove_keys", Keys: []string{"age"}}}},
		},
	}
	if err := ResolveImports(context.Background(), slog.Default(), nil, nil, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(cfg.Rules))
	}
}
