package types_test

import (
	"testing"

	"github.com/rayakame/sqlc-gen-better-python/internal/types"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func TestSqliteTypeToPython(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		pluginType *plugin.Identifier
		want       string
	}{
		{"int", &plugin.Identifier{Name: "int"}, types.Int},
		{"integer uppercase is lowered", &plugin.Identifier{Name: "INTEGER"}, types.Int},
		{"tinyint", &plugin.Identifier{Name: "tinyint"}, types.Int},
		{"smallint", &plugin.Identifier{Name: "smallint"}, types.Int},
		{"mediumint", &plugin.Identifier{Name: "mediumint"}, types.Int},
		{"bigint", &plugin.Identifier{Name: "bigint"}, types.Int},
		{"unsignedbigint", &plugin.Identifier{Name: "unsignedbigint"}, types.Int},
		{"int2", &plugin.Identifier{Name: "int2"}, types.Int},
		{"int8", &plugin.Identifier{Name: "int8"}, types.Int},
		{"bigserial", &plugin.Identifier{Name: "bigserial"}, types.Int},
		{"blob", &plugin.Identifier{Name: "blob"}, "memoryview"},
		{"real", &plugin.Identifier{Name: "real"}, types.Float},
		{"double", &plugin.Identifier{Name: "double"}, types.Float},
		{"double precision", &plugin.Identifier{Name: "double precision"}, types.Float},
		{"doubleprecision", &plugin.Identifier{Name: "doubleprecision"}, types.Float},
		{"float", &plugin.Identifier{Name: "float"}, types.Float},
		{"numeric", &plugin.Identifier{Name: "numeric"}, types.Float},
		{"boolean", &plugin.Identifier{Name: "boolean"}, types.Bool},
		{"bool", &plugin.Identifier{Name: "bool"}, types.Bool},
		{"date", &plugin.Identifier{Name: "date"}, "datetime.date"},
		{"datetime", &plugin.Identifier{Name: "datetime"}, "datetime.datetime"},
		{"timestamp", &plugin.Identifier{Name: "timestamp"}, "datetime.datetime"},
		{"bare decimal", &plugin.Identifier{Name: "decimal"}, types.Decimal},
		{"character with length", &plugin.Identifier{Name: "character(20)"}, types.Str},
		{"varchar with length", &plugin.Identifier{Name: "varchar(255)"}, types.Str},
		{"varyingcharacter with length", &plugin.Identifier{Name: "varyingcharacter(10)"}, types.Str},
		{"nchar with length", &plugin.Identifier{Name: "nchar(55)"}, types.Str},
		{"nativecharacter with length", &plugin.Identifier{Name: "nativecharacter(70)"}, types.Str},
		{"nvarchar with length", &plugin.Identifier{Name: "nvarchar(100)"}, types.Str},
		{"text", &plugin.Identifier{Name: "text"}, types.Str},
		{"clob", &plugin.Identifier{Name: "clob"}, types.Str},
		{"json", &plugin.Identifier{Name: "json"}, types.Str},
		{"decimal with precision", &plugin.Identifier{Name: "decimal(10,5)"}, types.Decimal},
		{"unknown type", &plugin.Identifier{Name: "geometry"}, types.Any},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := types.SqliteTypeToPython(nil, nil, tc.pluginType); got != tc.want {
				t.Errorf("SqliteTypeToPython(%+v) = %q, want %q", tc.pluginType, got, tc.want)
			}
		})
	}
}
