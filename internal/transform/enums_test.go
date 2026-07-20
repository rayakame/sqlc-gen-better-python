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

func TestBuildEnumsSkipsSystemSchemas(t *testing.T) {
	t.Parallel()
	req := &plugin.GenerateRequest{
		Catalog: &plugin.Catalog{
			DefaultSchema: "public",
			Schemas: []*plugin.Schema{
				{Name: types.PgCatalog, Enums: []*plugin.Enum{{Name: "hidden_pg", Vals: []string{"a"}}}},
				{Name: types.InformationSchema, Enums: []*plugin.Enum{{Name: "hidden_info", Vals: []string{"b"}}}},
				{Name: "public", Enums: []*plugin.Enum{{Name: "test_mood", Vals: []string{"happy", "sad"}}}},
			},
		},
	}
	tf := transform.NewTransformer(&config.Config{}, req, types.PostgresTypeToPython)

	enums := tf.BuildEnums()

	if len(enums) != 1 {
		t.Fatalf("BuildEnums returned %d enums, want 1 (system schemas must be skipped)", len(enums))
	}
	if enums[0].Name != "TestMood" {
		t.Errorf("enum name = %q, want %q", enums[0].Name, "TestMood")
	}
	if len(enums[0].Constants) != 2 {
		t.Errorf("enum has %d constants, want 2", len(enums[0].Constants))
	}
}

func TestBuildEnumsSchemaNamingDedupAndSort(t *testing.T) {
	t.Parallel()
	req := &plugin.GenerateRequest{
		Catalog: &plugin.Catalog{
			DefaultSchema: "public",
			Schemas: []*plugin.Schema{
				{Name: "public", Enums: []*plugin.Enum{
					{Name: "zeta", Vals: []string{"z"}},
					{Name: "foo_bar", Vals: []string{"x"}},
				}},
				// Non-default schema: the class name gets the schema prefix,
				// colliding with public.foo_bar after camel-casing.
				{Name: "foo", Enums: []*plugin.Enum{{Name: "bar", Vals: []string{"y"}}}},
			},
		},
	}
	tf := transform.NewTransformer(&config.Config{}, req, types.PostgresTypeToPython)

	enums := tf.BuildEnums()

	want := []model.Enum{
		{Name: "FooBar", Constants: []model.EnumConstants{{Name: "X", Value: "x"}}},
		{Name: "FooBar2", Constants: []model.EnumConstants{{Name: "Y", Value: "y"}}},
		{Name: "Zeta", Constants: []model.EnumConstants{{Name: "Z", Value: "z"}}},
	}
	if !reflect.DeepEqual(enums, want) {
		t.Errorf("BuildEnums() = %+v, want %+v", enums, want)
	}
}

func TestBuildEnumsConstantNaming(t *testing.T) {
	t.Parallel()
	req := &plugin.GenerateRequest{
		Catalog: &plugin.Catalog{
			DefaultSchema: "public",
			Schemas: []*plugin.Schema{
				{Name: "public", Enums: []*plugin.Enum{
					{Name: "edge", Vals: []string{"happy", "a-b", "a_b", "9lives", "---", "_x", "HAPPY"}},
				}},
			},
		},
	}
	tf := transform.NewTransformer(&config.Config{}, req, types.PostgresTypeToPython)

	enums := tf.BuildEnums()

	want := []model.Enum{{Name: "Edge", Constants: []model.EnumConstants{
		{Name: "HAPPY", Value: "happy"},
		{Name: "A_B", Value: "a-b"},
		{Name: "A_B_2", Value: "a_b"},
		{Name: "VALUE_9LIVES", Value: "9lives"},
		{Name: "VALUE_5", Value: "---"},
		{Name: "VALUE__X", Value: "_x"},
		{Name: "HAPPY_2", Value: "HAPPY"},
	}}}
	if !reflect.DeepEqual(enums, want) {
		t.Errorf("BuildEnums() = %+v, want %+v", enums, want)
	}
}
