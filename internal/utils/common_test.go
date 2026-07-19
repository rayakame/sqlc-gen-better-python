package utils_test

import (
	"testing"

	"github.com/rayakame/sqlc-gen-better-python/internal/utils"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

const (
	tableUsers   = "users"
	schemaPublic = "public"
)

func TestToPtr(t *testing.T) {
	t.Parallel()
	value := "hello"
	ptr := utils.ToPtr(value)
	if ptr == nil || *ptr != value {
		t.Fatalf("ToPtr(%q) = %v, want pointer to %q", value, ptr, value)
	}
	if ptr == &value {
		t.Fatal("ToPtr must return a pointer to a copy, not to the argument")
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
				Name: tableUsers,
			},
			defaultSchema: schemaPublic,
			want:          false,
		},
		{
			name: "nil second table",
			table1: &plugin.Identifier{
				Name: tableUsers,
			},
			defaultSchema: schemaPublic,
			want:          false,
		},
		{
			name:          "both nil",
			defaultSchema: schemaPublic,
			want:          false,
		},
		{
			name:          "equal fully qualified",
			table1:        &plugin.Identifier{Catalog: "db", Schema: schemaPublic, Name: tableUsers},
			table2:        &plugin.Identifier{Catalog: "db", Schema: schemaPublic, Name: tableUsers},
			defaultSchema: schemaPublic,
			want:          true,
		},
		{
			name:          "empty schema falls back to default on first",
			table1:        &plugin.Identifier{Name: tableUsers},
			table2:        &plugin.Identifier{Schema: schemaPublic, Name: tableUsers},
			defaultSchema: schemaPublic,
			want:          true,
		},
		{
			name:          "empty schema falls back to default on second",
			table1:        &plugin.Identifier{Schema: schemaPublic, Name: tableUsers},
			table2:        &plugin.Identifier{Name: tableUsers},
			defaultSchema: schemaPublic,
			want:          true,
		},
		{
			name:          "different schemas",
			table1:        &plugin.Identifier{Schema: schemaPublic, Name: tableUsers},
			table2:        &plugin.Identifier{Schema: "audit", Name: tableUsers},
			defaultSchema: schemaPublic,
			want:          false,
		},
		{
			name:          "different catalogs",
			table1:        &plugin.Identifier{Catalog: "db1", Schema: schemaPublic, Name: tableUsers},
			table2:        &plugin.Identifier{Catalog: "db2", Schema: schemaPublic, Name: tableUsers},
			defaultSchema: schemaPublic,
			want:          false,
		},
		{
			name:          "different names",
			table1:        &plugin.Identifier{Schema: schemaPublic, Name: tableUsers},
			table2:        &plugin.Identifier{Schema: schemaPublic, Name: "orders"},
			defaultSchema: schemaPublic,
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
