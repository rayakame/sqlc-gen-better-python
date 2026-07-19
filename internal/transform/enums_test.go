package transform_test

import (
	"testing"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
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
