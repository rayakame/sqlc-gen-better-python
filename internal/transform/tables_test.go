package transform_test

import (
	"reflect"
	"testing"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/model"
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

func TestBuildTablesClassNameDedupAcrossSchemas(t *testing.T) {
	t.Parallel()
	req := &plugin.GenerateRequest{
		Catalog: &plugin.Catalog{
			DefaultSchema: "public",
			Schemas: []*plugin.Schema{
				{Name: "public", Tables: []*plugin.Table{{Rel: &plugin.Identifier{Name: "foo_bars"}}}},
				// foo.bars singularizes and schema-qualifies to the same
				// class name as public.foo_bars.
				{Name: "foo", Tables: []*plugin.Table{{Rel: &plugin.Identifier{Name: "bars"}}}},
			},
		},
	}
	tf := transform.NewTransformer(&config.Config{}, req, types.PostgresTypeToPython)

	tables := tf.BuildTables()

	want := []model.Table{
		{Name: "FooBar", Columns: []model.Column{}, Identifier: &plugin.Identifier{Schema: "public", Name: "foo_bars"}},
		{Name: "FooBar2", Columns: []model.Column{}, Identifier: &plugin.Identifier{Schema: "foo", Name: "bars"}},
	}
	if !reflect.DeepEqual(tables, want) {
		t.Errorf("BuildTables() = %+v, want %+v", tables, want)
	}
}

func TestBuildTablesColumnNameDedup(t *testing.T) {
	t.Parallel()
	req := &plugin.GenerateRequest{
		Catalog: &plugin.Catalog{
			DefaultSchema: "public",
			Schemas: []*plugin.Schema{
				{Name: "public", Tables: []*plugin.Table{{
					Rel: &plugin.Identifier{Name: "items"},
					Columns: []*plugin.Column{
						// "a b" and "a_b" both sanitize to a_b.
						{Name: "a b", Type: &plugin.Identifier{Name: "text"}, NotNull: true},
						{Name: "a_b", Type: &plugin.Identifier{Name: "text"}, NotNull: true},
						{Name: "", Type: &plugin.Identifier{Name: "int4"}, NotNull: true},
					},
				}}},
			},
		},
	}
	tf := transform.NewTransformer(&config.Config{}, req, types.PostgresTypeToPython)

	tables := tf.BuildTables()

	want := []model.Table{{
		Name: "Item",
		Columns: []model.Column{
			{Name: "a_b", DBName: "a b", Type: model.PyType{SQLType: "text", Type: "str", DefaultType: "str"}},
			{Name: "a_b_2", DBName: "a_b", Type: model.PyType{SQLType: "text", Type: "str", DefaultType: "str"}},
			{Name: "column_3", DBName: "column_3", Type: model.PyType{SQLType: "int4", Type: types.Int, DefaultType: types.Int}},
		},
		Identifier: &plugin.Identifier{Schema: "public", Name: "items"},
	}}
	if !reflect.DeepEqual(tables, want) {
		t.Errorf("BuildTables() = %+v, want %+v", tables, want)
	}
}
