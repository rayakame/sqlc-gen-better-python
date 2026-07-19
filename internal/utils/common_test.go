package utils_test

import (
	"testing"

	"github.com/rayakame/sqlc-gen-better-python/internal/utils"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func TestToPtr(t *testing.T) {
	t.Parallel()
	value := "hello"
	ptr := utils.ToPtr(value)
	if *ptr != value {
		t.Fatalf("ToPtr(%q) points at %q, want %q", value, *ptr, value)
	}
	*ptr = "changed"
	if value != "hello" {
		t.Fatal("ToPtr must return a pointer to a copy; mutating it changed the original")
	}
}

func TestSameTableName(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name          string
		table1        *plugin.Identifier
		table2        *plugin.Identifier
		defaultSchema string
		want          bool
	}{
		{
			name: "nil first table",
			table2: &plugin.Identifier{
				Name: "users",
			},
			defaultSchema: "public",
			want:          false,
		},
		{
			name: "nil second table",
			table1: &plugin.Identifier{
				Name: "users",
			},
			defaultSchema: "public",
			want:          false,
		},
		{
			name:          "both nil",
			defaultSchema: "public",
			want:          false,
		},
		{
			name:          "equal fully qualified",
			table1:        &plugin.Identifier{Catalog: "db", Schema: "public", Name: "users"},
			table2:        &plugin.Identifier{Catalog: "db", Schema: "public", Name: "users"},
			defaultSchema: "public",
			want:          true,
		},
		{
			name:          "empty schema falls back to default on first",
			table1:        &plugin.Identifier{Name: "users"},
			table2:        &plugin.Identifier{Schema: "public", Name: "users"},
			defaultSchema: "public",
			want:          true,
		},
		{
			name:          "empty schema falls back to default on second",
			table1:        &plugin.Identifier{Schema: "public", Name: "users"},
			table2:        &plugin.Identifier{Name: "users"},
			defaultSchema: "public",
			want:          true,
		},
		{
			name:          "empty default schema with both schemas empty",
			table1:        &plugin.Identifier{Name: "users"},
			table2:        &plugin.Identifier{Name: "users"},
			defaultSchema: "",
			want:          true,
		},
		{
			name:          "catalog has no default fallback",
			table1:        &plugin.Identifier{Catalog: "db", Schema: "public", Name: "users"},
			table2:        &plugin.Identifier{Schema: "public", Name: "users"},
			defaultSchema: "public",
			want:          false,
		},
		{
			name:          "different schemas",
			table1:        &plugin.Identifier{Schema: "public", Name: "users"},
			table2:        &plugin.Identifier{Schema: "audit", Name: "users"},
			defaultSchema: "public",
			want:          false,
		},
		{
			name:          "different catalogs",
			table1:        &plugin.Identifier{Catalog: "db1", Schema: "public", Name: "users"},
			table2:        &plugin.Identifier{Catalog: "db2", Schema: "public", Name: "users"},
			defaultSchema: "public",
			want:          false,
		},
		{
			name:          "different names",
			table1:        &plugin.Identifier{Schema: "public", Name: "users"},
			table2:        &plugin.Identifier{Schema: "public", Name: "orders"},
			defaultSchema: "public",
			want:          false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := utils.SameTableName(tc.table1, tc.table2, tc.defaultSchema); got != tc.want {
				t.Errorf("SameTableName(%v, %v, %q) = %v, want %v", tc.table1, tc.table2, tc.defaultSchema, got, tc.want)
			}
		})
	}
}
