package transform_test

import (
	"testing"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/transform"
	"github.com/rayakame/sqlc-gen-better-python/internal/types"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func TestBuildTablesSkipsSystemSchemas(t *testing.T) {
	t.Parallel()
	req := &plugin.GenerateRequest{
		Catalog: &plugin.Catalog{
			DefaultSchema: "public",
			Schemas: []*plugin.Schema{
				{
					Name:   types.PgCatalog,
					Tables: []*plugin.Table{{Rel: &plugin.Identifier{Name: "pg_class"}}},
				},
				{
					Name: "public",
					Tables: []*plugin.Table{{
						Rel: &plugin.Identifier{Name: "test_items"},
						Columns: []*plugin.Column{
							{Name: "id", Type: &plugin.Identifier{Name: "int4"}, NotNull: true},
						},
					}},
				},
			},
		},
	}
	tf := transform.NewTransformer(&config.Config{}, req, types.PostgresTypeToPython)

	tables := tf.BuildTables()

	if len(tables) != 1 {
		t.Fatalf("BuildTables returned %d tables, want 1 (system schemas must be skipped)", len(tables))
	}
	if tables[0].Name != "TestItem" {
		t.Errorf("table name = %q, want %q (singularized CapWords)", tables[0].Name, "TestItem")
	}
	if len(tables[0].Columns) != 1 || tables[0].Columns[0].DBName != "id" {
		t.Errorf("columns = %+v, want a single id column", tables[0].Columns)
	}
}
