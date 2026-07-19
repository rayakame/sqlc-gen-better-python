package config_test

import (
	"testing"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/sqlc-dev/plugin-sdk-go/pattern"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

// newOverrideRequest wraps a single override JSON fragment in otherwise valid
// plugin options; override parse errors surface before config validation.
func newOverrideRequest(catalog *plugin.Catalog, overrideJSON string) *plugin.GenerateRequest {
	options := `{"package":"db","sql_driver":"asyncpg","emit_init_file":true,"overrides":[` + overrideJSON + `]}`

	return &plugin.GenerateRequest{
		Settings:      &plugin.Settings{Engine: "postgresql"},
		Catalog:       catalog,
		PluginOptions: []byte(options),
	}
}

func mustPattern(t *testing.T, expr string) *pattern.Match {
	t.Helper()
	match, err := pattern.MatchCompile(expr)
	if err != nil {
		t.Fatalf("MatchCompile(%q) returned error: %v", expr, err)
	}

	return match
}

func checkPattern(t *testing.T, field string, match *pattern.Match, want string) {
	t.Helper()
	if want == "" {
		if match != nil {
			t.Errorf("%s = %v, want nil", field, match)
		}

		return
	}
	if match == nil {
		t.Fatalf("%s is nil, want pattern matching %q", field, want)
	}
	if !match.MatchString(want) {
		t.Errorf("%s does not match %q", field, want)
	}
}

func TestOverrideParseErrors(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		override string
		wantErr  string
	}{
		{
			name:     "both column and db_type",
			override: `{"py_type":{"type":"str"},"column":"authors.name","db_type":"text"}`,
			wantErr:  "override specifying both `column` (\"authors.name\") and `db_type` (\"text\") is not valid",
		},
		{
			name:     "neither column nor db_type",
			override: `{"py_type":{"type":"str"}}`,
			wantErr:  "override must specify one of either `column` or `db_type`",
		},
		{
			name:     "db_type without py_type type",
			override: `{"db_type":"text"}`,
			wantErr:  "override must specify a `py_type` with a non-empty `type`",
		},
		{
			name:     "column without py_type type",
			override: `{"py_type":{"import":"collections"},"column":"authors.name"}`,
			wantErr:  "override must specify a `py_type` with a non-empty `type`",
		},
		{
			name:     "column with one part",
			override: `{"py_type":{"type":"str"},"column":"name"}`,
			wantErr:  "override `column` specifier \"name\" is not the proper format, expected '[catalog.][schema.]tablename.colname'",
		},
		{
			name:     "column with five parts",
			override: `{"py_type":{"type":"str"},"column":"a.b.c.d.e"}`,
			wantErr:  "override `column` specifier \"a.b.c.d.e\" is not the proper format, expected '[catalog.][schema.]tablename.colname'",
		},
		{
			name:     "invalid escape in column pattern",
			override: `{"py_type":{"type":"str"},"column":"authors.na\\me"}`,
			wantErr:  "Invalid escaped character 'm'",
		},
		{
			name:     "unterminated escape in column pattern",
			override: `{"py_type":{"type":"str"},"column":"authors.name\\"}`,
			wantErr:  "Unterminated escape at end of pattern",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := newOverrideRequest(&plugin.Catalog{DefaultSchema: "public"}, tc.override)
			cfg, err := config.NewConfig(req)
			if cfg != nil {
				t.Errorf("NewConfig returned non-nil config %v, want nil", cfg)
			}
			if err == nil {
				t.Fatalf("NewConfig returned nil error, want %q", tc.wantErr)
			}
			if err.Error() != tc.wantErr {
				t.Errorf("NewConfig error = %q, want %q", err.Error(), tc.wantErr)
			}
		})
	}
}

func TestOverrideParseValid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		catalog     *plugin.Catalog
		override    string
		colMatch    string
		relMatch    string
		schemaMatch string
		catMatch    string
	}{
		{
			name:     "db_type override compiles no patterns",
			catalog:  &plugin.Catalog{DefaultSchema: "public"},
			override: `{"py_type":{"type":"decimal.Decimal"},"db_type":"pg_catalog.numeric"}`,
		},
		{
			name:        "two part column uses catalog default schema",
			catalog:     &plugin.Catalog{DefaultSchema: "main"},
			override:    `{"py_type":{"type":"str"},"column":"authors.name"}`,
			colMatch:    "name",
			relMatch:    "authors",
			schemaMatch: "main",
		},
		{
			name:        "two part column without catalog defaults to public",
			override:    `{"py_type":{"type":"str"},"column":"authors.name"}`,
			colMatch:    "name",
			relMatch:    "authors",
			schemaMatch: "public",
		},
		{
			name:        "three part column sets schema pattern",
			catalog:     &plugin.Catalog{DefaultSchema: "public"},
			override:    `{"py_type":{"type":"str"},"column":"myschema.authors.name"}`,
			colMatch:    "name",
			relMatch:    "authors",
			schemaMatch: "myschema",
		},
		{
			name:        "four part column sets catalog pattern",
			catalog:     &plugin.Catalog{DefaultSchema: "public"},
			override:    `{"py_type":{"type":"str"},"column":"mycat.myschema.authors.name"}`,
			colMatch:    "name",
			relMatch:    "authors",
			schemaMatch: "myschema",
			catMatch:    "mycat",
		},
		{
			name:        "wildcard table matches any table",
			catalog:     &plugin.Catalog{DefaultSchema: "public"},
			override:    `{"py_type":{"type":"str"},"column":"*.name"}`,
			colMatch:    "name",
			relMatch:    "any_table_at_all",
			schemaMatch: "public",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg, err := config.NewConfig(newOverrideRequest(tc.catalog, tc.override))
			if err != nil {
				t.Fatalf("NewConfig returned error: %v", err)
			}
			if len(cfg.Overrides) != 1 {
				t.Fatalf("len(cfg.Overrides) = %d, want 1", len(cfg.Overrides))
			}
			override := cfg.Overrides[0]
			checkPattern(t, "ColumnName", override.ColumnName, tc.colMatch)
			checkPattern(t, "TableRel", override.TableRel, tc.relMatch)
			checkPattern(t, "TableSchema", override.TableSchema, tc.schemaMatch)
			checkPattern(t, "TableCatalog", override.TableCatalog, tc.catMatch)
		})
	}

	t.Run("py_type fields are preserved", func(t *testing.T) {
		t.Parallel()
		override := `{"py_type":{"import":"collections","type":"UserString","package":"UserString"},"db_type":"text"}`
		cfg, err := config.NewConfig(newOverrideRequest(nil, override))
		if err != nil {
			t.Fatalf("NewConfig returned error: %v", err)
		}
		got := cfg.Overrides[0]
		want := config.OverridePyType{Import: "collections", Type: "UserString", Package: "UserString"}
		if got.PyType != want {
			t.Errorf("PyType = %+v, want %+v", got.PyType, want)
		}
		if got.DBType != "text" {
			t.Errorf("DBType = %q, want %q", got.DBType, "text")
		}
	})

	t.Run("compiled patterns are anchored", func(t *testing.T) {
		t.Parallel()
		override := `{"py_type":{"type":"str"},"column":"authors.name"}`
		cfg, err := config.NewConfig(newOverrideRequest(nil, override))
		if err != nil {
			t.Fatalf("NewConfig returned error: %v", err)
		}
		colName := cfg.Overrides[0].ColumnName
		if colName.MatchString("name2") || colName.MatchString("aname") {
			t.Error("ColumnName pattern must be anchored to the full string")
		}
	})
}

func TestOverrideMatches(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name          string
		override      *config.Override
		identifier    *plugin.Identifier
		defaultSchema string
		want          bool
	}{
		{
			name:          "nil identifier",
			override:      &config.Override{TableSchema: mustPattern(t, "public"), TableRel: mustPattern(t, "authors")},
			identifier:    nil,
			defaultSchema: "public",
			want:          false,
		},
		{
			name: "catalog pattern mismatch",
			override: &config.Override{
				TableCatalog: mustPattern(t, "db1"),
				TableSchema:  mustPattern(t, "public"),
				TableRel:     mustPattern(t, "authors"),
			},
			identifier:    &plugin.Identifier{Catalog: "db2", Schema: "public", Name: "authors"},
			defaultSchema: "public",
			want:          false,
		},
		{
			name: "full match with catalog",
			override: &config.Override{
				TableCatalog: mustPattern(t, "db1"),
				TableSchema:  mustPattern(t, "public"),
				TableRel:     mustPattern(t, "authors"),
			},
			identifier:    &plugin.Identifier{Catalog: "db1", Schema: "public", Name: "authors"},
			defaultSchema: "public",
			want:          true,
		},
		{
			name:          "nil schema pattern rejects explicit schema",
			override:      &config.Override{TableRel: mustPattern(t, "authors")},
			identifier:    &plugin.Identifier{Schema: "public", Name: "authors"},
			defaultSchema: "public",
			want:          false,
		},
		{
			name:          "nil schema pattern rejects default schema",
			override:      &config.Override{TableRel: mustPattern(t, "authors")},
			identifier:    &plugin.Identifier{Name: "authors"},
			defaultSchema: "public",
			want:          false,
		},
		{
			name:          "nil schema pattern with empty schemas matches",
			override:      &config.Override{TableRel: mustPattern(t, "authors")},
			identifier:    &plugin.Identifier{Name: "authors"},
			defaultSchema: "",
			want:          true,
		},
		{
			name:          "all nil patterns match empty identifier",
			override:      &config.Override{},
			identifier:    &plugin.Identifier{},
			defaultSchema: "",
			want:          true,
		},
		{
			name:          "nil rel pattern rejects named table",
			override:      &config.Override{},
			identifier:    &plugin.Identifier{Name: "authors"},
			defaultSchema: "",
			want:          false,
		},
		{
			name:          "schema pattern mismatch",
			override:      &config.Override{TableSchema: mustPattern(t, "public"), TableRel: mustPattern(t, "authors")},
			identifier:    &plugin.Identifier{Schema: "audit", Name: "authors"},
			defaultSchema: "public",
			want:          false,
		},
		{
			name:          "empty schema falls back to default schema",
			override:      &config.Override{TableSchema: mustPattern(t, "public"), TableRel: mustPattern(t, "authors")},
			identifier:    &plugin.Identifier{Name: "authors"},
			defaultSchema: "public",
			want:          true,
		},
		{
			name:          "rel pattern mismatch",
			override:      &config.Override{TableSchema: mustPattern(t, "public"), TableRel: mustPattern(t, "authors")},
			identifier:    &plugin.Identifier{Schema: "public", Name: "orders"},
			defaultSchema: "public",
			want:          false,
		},
		{
			name:          "wildcard rel matches any table",
			override:      &config.Override{TableSchema: mustPattern(t, "public"), TableRel: mustPattern(t, "*")},
			identifier:    &plugin.Identifier{Schema: "public", Name: "whatever"},
			defaultSchema: "public",
			want:          true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.override.Matches(tc.identifier, tc.defaultSchema); got != tc.want {
				t.Errorf("Matches(%v, %q) = %v, want %v", tc.identifier, tc.defaultSchema, got, tc.want)
			}
		})
	}
}
