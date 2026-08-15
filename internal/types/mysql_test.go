package types_test

import (
	"testing"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/types"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func TestMysqlTypeToPython(t *testing.T) {
	t.Parallel()
	req := &plugin.GenerateRequest{
		Catalog: &plugin.Catalog{
			DefaultSchema: "public",
			Schemas: []*plugin.Schema{
				// The same enum name inside system schemas proves they are
				// skipped during enum resolution.
				{Name: types.PgCatalog, Enums: []*plugin.Enum{{Name: "authors_status", Vals: []string{"x"}}}},
				{Name: types.InformationSchema, Enums: []*plugin.Enum{{Name: "authors_status", Vals: []string{"x"}}}},
				{Name: "public", Enums: []*plugin.Enum{{Name: "authors_status", Vals: []string{"draft"}}}},
				{Name: "other", Enums: []*plugin.Enum{{Name: "other_mood", Vals: []string{"y"}}}},
			},
		},
	}
	conf := &config.Config{}
	cases := []struct {
		name       string
		pluginType *plugin.Identifier
		length     int32
		want       string
	}{
		{"tinyint length 1 is bool", &plugin.Identifier{Name: "tinyint"}, 1, types.Bool},
		{"tinyint length 4", &plugin.Identifier{Name: "tinyint"}, 4, types.Int},
		{"tinyint without length", &plugin.Identifier{Name: "tinyint"}, 0, types.Int},
		{"tinyint uppercase is lowered", &plugin.Identifier{Name: "TINYINT"}, 0, types.Int},
		{"bool", &plugin.Identifier{Name: "bool"}, 0, types.Bool},
		{"boolean", &plugin.Identifier{Name: "boolean"}, 0, types.Bool},
		{"int", &plugin.Identifier{Name: "int"}, 0, types.Int},
		{"integer", &plugin.Identifier{Name: "integer"}, 0, types.Int},
		{"mediumint", &plugin.Identifier{Name: "mediumint"}, 0, types.Int},
		{"smallint", &plugin.Identifier{Name: "smallint"}, 0, types.Int},
		{"bigint", &plugin.Identifier{Name: "bigint"}, 0, types.Int},
		{"year", &plugin.Identifier{Name: "year"}, 0, types.Int},
		{"serial", &plugin.Identifier{Name: "serial"}, 0, types.Int},
		{"bigint unsigned", &plugin.Identifier{Name: "bigint unsigned"}, 0, types.Int},
		{"bigint signed", &plugin.Identifier{Name: "bigint signed"}, 0, types.Int},
		{"float", &plugin.Identifier{Name: "float"}, 0, types.Float},
		{"double", &plugin.Identifier{Name: "double"}, 0, types.Float},
		{"double precision", &plugin.Identifier{Name: "double precision"}, 0, types.Float},
		{"real", &plugin.Identifier{Name: "real"}, 0, types.Float},
		{"decimal", &plugin.Identifier{Name: "decimal"}, 0, types.Decimal},
		{"dec", &plugin.Identifier{Name: "dec"}, 0, types.Decimal},
		{"fixed", &plugin.Identifier{Name: "fixed"}, 0, types.Decimal},
		{"numeric", &plugin.Identifier{Name: "numeric"}, 0, types.Decimal},
		{"varchar", &plugin.Identifier{Name: "varchar"}, 0, types.Str},
		{"char", &plugin.Identifier{Name: "char"}, 0, types.Str},
		{"text", &plugin.Identifier{Name: "text"}, 0, types.Str},
		{"tinytext", &plugin.Identifier{Name: "tinytext"}, 0, types.Str},
		{"mediumtext", &plugin.Identifier{Name: "mediumtext"}, 0, types.Str},
		{"longtext", &plugin.Identifier{Name: "longtext"}, 0, types.Str},
		{"set", &plugin.Identifier{Name: "set"}, 0, types.Str},
		{"bare enum", &plugin.Identifier{Name: "enum"}, 0, types.Str},
		{"blob", &plugin.Identifier{Name: "blob"}, 0, "memoryview"},
		{"binary", &plugin.Identifier{Name: "binary"}, 0, "memoryview"},
		{"varbinary", &plugin.Identifier{Name: "varbinary"}, 0, "memoryview"},
		{"tinyblob", &plugin.Identifier{Name: "tinyblob"}, 0, "memoryview"},
		{"mediumblob", &plugin.Identifier{Name: "mediumblob"}, 0, "memoryview"},
		{"longblob", &plugin.Identifier{Name: "longblob"}, 0, "memoryview"},
		{"bit", &plugin.Identifier{Name: "bit"}, 0, "memoryview"},
		{"date", &plugin.Identifier{Name: "date"}, 0, "datetime.date"},
		{"datetime", &plugin.Identifier{Name: "datetime"}, 0, "datetime.datetime"},
		{"datetime uppercase is lowered", &plugin.Identifier{Name: "DATETIME"}, 0, "datetime.datetime"},
		{"timestamp", &plugin.Identifier{Name: "timestamp"}, 0, "datetime.datetime"},
		{"time", &plugin.Identifier{Name: "time"}, 0, "datetime.timedelta"},
		{"json", &plugin.Identifier{Name: "json"}, 0, types.Str},
		{"any", &plugin.Identifier{Name: "any"}, 0, types.Any},
		{"enum in default schema", &plugin.Identifier{Name: "authors_status"}, 0, "enums.AuthorsStatus"},
		{"enum in named schema", &plugin.Identifier{Schema: "other", Name: "other_mood"}, 0, "enums.OtherOtherMood"},
		{
			"system-schema qualified enum is not resolved",
			&plugin.Identifier{Schema: types.PgCatalog, Name: "authors_status"},
			0,
			types.Any,
		},
		{
			"information_schema qualified enum is not resolved",
			&plugin.Identifier{Schema: types.InformationSchema, Name: "authors_status"},
			0,
			types.Any,
		},
		{"invalid four-part identifier", &plugin.Identifier{Name: "a.b.c.d"}, 0, types.Any},
		{"unknown type", &plugin.Identifier{Name: "geometry"}, 0, types.Any},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			col := &plugin.Column{Type: tc.pluginType, Length: tc.length}
			if got := types.MysqlTypeToPython(req, conf, col); got != tc.want {
				t.Errorf("MysqlTypeToPython(%+v, length=%d) = %q, want %q", tc.pluginType, tc.length, got, tc.want)
			}
		})
	}
}
