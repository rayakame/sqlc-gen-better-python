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
				// The same enum name inside system schemas proves they are
				// skipped during enum resolution.
				{Name: types.PgCatalog, Enums: []*plugin.Enum{{Name: "test_mood", Vals: []string{"x"}}}},
				{Name: types.InformationSchema, Enums: []*plugin.Enum{{Name: "test_mood", Vals: []string{"x"}}}},
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
		{"builtin serial", &plugin.Identifier{Name: "serial"}, types.Int},
		{"builtin serial4", &plugin.Identifier{Name: "serial4"}, types.Int},
		{"builtin pg_catalog.serial4", &plugin.Identifier{Name: "pg_catalog.serial4"}, types.Int},
		{"builtin bigserial", &plugin.Identifier{Name: "bigserial"}, types.Int},
		{"builtin serial8", &plugin.Identifier{Name: "serial8"}, types.Int},
		{"builtin pg_catalog.serial8", &plugin.Identifier{Name: "pg_catalog.serial8"}, types.Int},
		{"builtin smallserial", &plugin.Identifier{Name: "smallserial"}, types.Int},
		{"builtin serial2", &plugin.Identifier{Name: "serial2"}, types.Int},
		{"builtin pg_catalog.serial2", &plugin.Identifier{Name: "pg_catalog.serial2"}, types.Int},
		{"builtin integer", &plugin.Identifier{Name: "integer"}, types.Int},
		{"builtin int", &plugin.Identifier{Name: "int"}, types.Int},
		{"builtin int4", &plugin.Identifier{Name: "int4"}, types.Int},
		{"builtin pg_catalog.int4", &plugin.Identifier{Name: "pg_catalog.int4"}, types.Int},
		{"builtin bigint", &plugin.Identifier{Name: "bigint"}, types.Int},
		{"builtin int8", &plugin.Identifier{Name: "int8"}, types.Int},
		{"builtin pg_catalog.int8", &plugin.Identifier{Name: "pg_catalog.int8"}, types.Int},
		{"builtin smallint", &plugin.Identifier{Name: "smallint"}, types.Int},
		{"builtin int2", &plugin.Identifier{Name: "int2"}, types.Int},
		{"builtin pg_catalog.int2", &plugin.Identifier{Name: "pg_catalog.int2"}, types.Int},
		{"builtin float", &plugin.Identifier{Name: "float"}, types.Float},
		{"builtin double precision", &plugin.Identifier{Name: "double precision"}, types.Float},
		{"builtin float8", &plugin.Identifier{Name: "float8"}, types.Float},
		{"builtin pg_catalog.float8", &plugin.Identifier{Name: "pg_catalog.float8"}, types.Float},
		{"builtin real", &plugin.Identifier{Name: "real"}, types.Float},
		{"builtin float4", &plugin.Identifier{Name: "float4"}, types.Float},
		{"builtin pg_catalog.float4", &plugin.Identifier{Name: "pg_catalog.float4"}, types.Float},
		{"builtin numeric", &plugin.Identifier{Name: "numeric"}, types.Decimal},
		{"builtin pg_catalog.numeric", &plugin.Identifier{Name: "pg_catalog.numeric"}, types.Decimal},
		{"builtin money", &plugin.Identifier{Name: "money"}, types.Str},
		{"builtin boolean", &plugin.Identifier{Name: "boolean"}, types.Bool},
		{"builtin bool", &plugin.Identifier{Name: "bool"}, types.Bool},
		{"builtin pg_catalog.bool", &plugin.Identifier{Name: "pg_catalog.bool"}, types.Bool},
		{"builtin pg_catalog.json", &plugin.Identifier{Name: "pg_catalog.json"}, types.Str},
		{"builtin json", &plugin.Identifier{Name: "json"}, types.Str},
		{"builtin jsonb", &plugin.Identifier{Name: "jsonb"}, types.Str},
		{"builtin bytea", &plugin.Identifier{Name: "bytea"}, "memoryview"},
		{"builtin blob", &plugin.Identifier{Name: "blob"}, "memoryview"},
		{"builtin pg_catalog.bytea", &plugin.Identifier{Name: "pg_catalog.bytea"}, "memoryview"},
		{"builtin date", &plugin.Identifier{Name: "date"}, "datetime.date"},
		{"builtin pg_catalog.time", &plugin.Identifier{Name: "pg_catalog.time"}, "datetime.time"},
		{"builtin pg_catalog.timetz", &plugin.Identifier{Name: "pg_catalog.timetz"}, "datetime.time"},
		{"builtin timetz", &plugin.Identifier{Name: "timetz"}, "datetime.time"},
		{"builtin pg_catalog.timestamp", &plugin.Identifier{Name: "pg_catalog.timestamp"}, "datetime.datetime"},
		{"builtin pg_catalog.timestamptz", &plugin.Identifier{Name: "pg_catalog.timestamptz"}, "datetime.datetime"},
		{"builtin timestamptz", &plugin.Identifier{Name: "timestamptz"}, "datetime.datetime"},
		{"builtin interval", &plugin.Identifier{Name: "interval"}, "datetime.timedelta"},
		{"builtin pg_catalog.interval", &plugin.Identifier{Name: "pg_catalog.interval"}, "datetime.timedelta"},
		{"builtin text", &plugin.Identifier{Name: "text"}, types.Str},
		{"builtin pg_catalog.varchar", &plugin.Identifier{Name: "pg_catalog.varchar"}, types.Str},
		{"builtin bpchar", &plugin.Identifier{Name: "bpchar"}, types.Str},
		{"builtin pg_catalog.bpchar", &plugin.Identifier{Name: "pg_catalog.bpchar"}, types.Str},
		{"builtin char", &plugin.Identifier{Name: "char"}, types.Str},
		{"builtin string", &plugin.Identifier{Name: "string"}, types.Str},
		{"builtin citext", &plugin.Identifier{Name: "citext"}, types.Str},
		{"builtin uuid", &plugin.Identifier{Name: "uuid"}, "uuid.UUID"},
		{"builtin pg_catalog.uuid", &plugin.Identifier{Name: "pg_catalog.uuid"}, "uuid.UUID"},
		{"builtin inet", &plugin.Identifier{Name: "inet"}, types.Str},
		{"builtin cidr", &plugin.Identifier{Name: "cidr"}, types.Str},
		{"builtin macaddr", &plugin.Identifier{Name: "macaddr"}, types.Str},
		{"builtin macaddr8", &plugin.Identifier{Name: "macaddr8"}, types.Str},
		{"builtin ltree", &plugin.Identifier{Name: "ltree"}, types.Str},
		{"builtin lquery", &plugin.Identifier{Name: "lquery"}, types.Str},
		{"builtin ltxtquery", &plugin.Identifier{Name: "ltxtquery"}, types.Str},
		{"enum in default schema", &plugin.Identifier{Name: "test_mood"}, "enums.TestMood"},
		{"enum in named schema", &plugin.Identifier{Schema: "other", Name: "other_mood"}, "enums.OtherOtherMood"},
		{
			"catalog-qualified enum resolves via schema part",
			&plugin.Identifier{Name: "somecatalog.other.other_mood"},
			"enums.OtherOtherMood",
		},
		{
			"system-schema qualified enum is not resolved",
			&plugin.Identifier{Schema: types.PgCatalog, Name: "test_mood"},
			types.Any,
		},
		{
			"information_schema qualified enum is not resolved",
			&plugin.Identifier{Schema: types.InformationSchema, Name: "test_mood"},
			types.Any,
		},
		{"enum name missing in matching schema", &plugin.Identifier{Schema: "other", Name: "missing"}, types.Any},
		{"invalid four-part identifier", &plugin.Identifier{Name: "a.b.c.d"}, types.Any},
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
