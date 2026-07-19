package types_test

import (
	"testing"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/types"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func TestPostgresTypeToPython(t *testing.T) {
	t.Parallel()
	req := &plugin.GenerateRequest{
		Catalog: &plugin.Catalog{
			DefaultSchema: "public",
			Schemas: []*plugin.Schema{
				// The same enum name inside pg_catalog proves system schemas
				// are skipped during enum resolution.
				{Name: types.PgCatalog, Enums: []*plugin.Enum{{Name: "test_mood", Vals: []string{"x"}}}},
				{Name: "public", Enums: []*plugin.Enum{{Name: "test_mood", Vals: []string{"happy"}}}},
				{Name: "other", Enums: []*plugin.Enum{{Name: "other_mood", Vals: []string{"y"}}}},
			},
		},
	}
	conf := &config.Config{}
	cases := []struct {
		name       string
		pluginType *plugin.Identifier
		want       string
	}{
		{"builtin int4", &plugin.Identifier{Name: "int4"}, types.Int},
		{"enum in default schema", &plugin.Identifier{Name: "test_mood"}, "enums.TestMood"},
		{"enum in named schema", &plugin.Identifier{Schema: "other", Name: "other_mood"}, "enums.OtherOtherMood"},
		{
			"system-schema qualified enum is not resolved",
			&plugin.Identifier{Schema: types.PgCatalog, Name: "test_mood"},
			types.Any,
		},
		{"unknown type", &plugin.Identifier{Name: "geometry"}, types.Any},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := types.PostgresTypeToPython(req, conf, tc.pluginType); got != tc.want {
				t.Errorf("PostgresTypeToPython(%+v) = %q, want %q", tc.pluginType, got, tc.want)
			}
		})
	}
}
