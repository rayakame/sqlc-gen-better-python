package render

import (
	"maps"
	"slices"
	"testing"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/driver"
	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/types"
	"github.com/rayakame/sqlc-gen-better-python/internal/utils"
	"github.com/rayakame/sqlc-gen-better-python/internal/writer"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
)

// Expected QueryResultsArgsType alias fragments, mirroring buildQueryResult.
const (
	argsTypeBase     = "type QueryResultsArgsType = int | float | str | memoryview"
	argsTypeDatetime = " | datetime.date | datetime.time | datetime.datetime | datetime.timedelta"
	argsTypeTail     = " | collections.abc.Sequence[QueryResultsArgsType] | None"
)

// Python type names repeated across fixtures.
const (
	typeDate     = "datetime.date"
	typeTime     = "datetime.time"
	typeDatetime = "datetime.datetime"
	typeUUID     = "uuid.UUID"
	typeMoney    = "mymod.Money"
)

// Converter fixture: the module holding the user functions and their dotted paths.
const (
	converterModule = "myconv"
	converterToDB   = "myconv.to_db"
	converterFromDB = "myconv.from_db"
)

// impConverter builds a converter-backed override type for the given SQL type.
func impConverter(sqlType, defaultType string) model.PyType {
	return model.PyType{
		SQLType:       sqlType,
		Type:          typeMoney,
		IsOverride:    true,
		DefaultType:   defaultType,
		ConverterTo:   converterToDB,
		ConverterFrom: converterFromDB,
	}
}

// impConverterConf configures a converter plus the override adopting its py_type.
// Modules is normally derived by config parsing, so it is set explicitly here.
func impConverterConf(c *config.Config) {
	pyType := config.OverridePyType{Import: "mymod", Type: typeMoney}
	c.Converters = []config.Converter{
		{Name: "money", PyType: pyType, ToDB: converterToDB, FromDB: converterFromDB},
	}
	c.Overrides = []config.Override{{PyType: pyType}}
}

func newImportsConfig(sqlDriver config.SQLDriver, mods ...func(*config.Config)) *config.Config {
	conf := &config.Config{Package: "db", SqlDriver: sqlDriver, ModelType: config.ModelTypeDataclass}
	for _, mod := range mods {
		mod(conf)
	}

	return conf
}

func newImportsResolver(t *testing.T, conf *config.Config) *ImportResolver {
	t.Helper()
	drv, err := driver.New(conf)
	if err != nil {
		t.Fatalf("driver.New() error: %v", err)
	}

	return NewImportResolver(conf, drv)
}

func newImportsWriter() *writer.CodeWriter {
	return writer.NewCodeWriter(&config.Config{
		IndentChar:          " ",
		CharsPerIndentLevel: 4,
		EmitDocstringsSQL:   utils.ToPtr(true),
	})
}

func checkImportResult(t *testing.T, got, want ImportResult) {
	t.Helper()
	if !slices.Equal(got.Std, want.Std) {
		t.Errorf("Std = %q, want %q", got.Std, want.Std)
	}
	if !slices.Equal(got.TypeChecking, want.TypeChecking) {
		t.Errorf("TypeChecking = %q, want %q", got.TypeChecking, want.TypeChecking)
	}
	if !slices.Equal(got.Package, want.Package) {
		t.Errorf("Package = %q, want %q", got.Package, want.Package)
	}
}

func impCol(name string, typ model.PyType) model.Column {
	return model.Column{Name: name, Type: typ}
}

func impScalar(typ model.PyType) model.QueryValue {
	return model.QueryValue{Name: "v", Type: typ}
}

func impStruct(emit bool, cols ...model.Column) model.QueryValue {
	return model.QueryValue{EmitTable: emit, Name: "row", Table: &model.Table{Name: "Row", Columns: cols}}
}

func impTable(cols ...model.Column) model.Table {
	return model.Table{Name: "Author", Columns: cols}
}

func TestImportSpecString(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		spec importSpec
		want string
	}{
		{name: "module import", spec: importSpec{Module: "sys"}, want: "import sys"},
		{name: "from import", spec: importSpec{Module: "db", Name: "helpers"}, want: "from db import helpers"},
		{name: "module alias", spec: importSpec{Module: "numpy", Alias: "np"}, want: "import numpy as np"},
		{
			name: "from import alias",
			spec: importSpec{Module: "db", Name: "helpers", Alias: "h"},
			want: "from db import helpers as h",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.spec.String(); got != tc.want {
				t.Errorf("String() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestImportResultWrite(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		res     ImportResult
		omit    bool
		tcLines []string
		want    string
	}{
		{
			name: "empty result writes nothing",
			res:  ImportResult{},
			want: "",
		},
		{
			name: "std only",
			res:  ImportResult{Std: []string{"import typing"}},
			want: "import typing\n",
		},
		{
			name: "std and typechecking separated by blank line",
			res:  ImportResult{Std: []string{"import typing"}, TypeChecking: []string{"import collections.abc"}},
			want: "import typing\n\nif typing.TYPE_CHECKING:\n    import collections.abc\n",
		},
		{
			name: "typechecking only has no leading blank line",
			res:  ImportResult{TypeChecking: []string{"import collections.abc"}},
			want: "if typing.TYPE_CHECKING:\n    import collections.abc\n",
		},
		{
			name:    "hook lines alone open the guard",
			res:     ImportResult{},
			tcLines: []string{"ConnectionLike = object"},
			want:    "if typing.TYPE_CHECKING:\n    ConnectionLike = object\n",
		},
		{
			name:    "hook lines after imports get a blank line",
			res:     ImportResult{TypeChecking: []string{"import asyncpg"}},
			tcLines: []string{"ConnectionLike = object"},
			want:    "if typing.TYPE_CHECKING:\n    import asyncpg\n\n    ConnectionLike = object\n",
		},
		{
			name: "package only",
			res:  ImportResult{Package: []string{"from db import models"}},
			want: "\nfrom db import models\n",
		},
		{
			name: "std and package without typechecking",
			res:  ImportResult{Std: []string{"import typing"}, Package: []string{"from db import models"}},
			want: "import typing\n\nfrom db import models\n",
		},
		{
			name: "omit moves statements and hooks after all imports",
			res: ImportResult{
				Std:          []string{"import typing"},
				TypeChecking: []string{"import asyncpg", "type QueryResultsArgsType = int | None"},
				Package:      []string{"from db import models"},
			},
			omit:    true,
			tcLines: []string{"ConnectionLike = object"},
			want: "import typing\nimport asyncpg\n\nfrom db import models\n\n" +
				"type QueryResultsArgsType = int | None\n\nConnectionLike = object\n",
		},
		{
			name: "omit keeps from imports and skips blank lines",
			res:  ImportResult{Std: []string{"import typing", ""}, TypeChecking: []string{"from x import y"}},
			omit: true,
			want: "import typing\nfrom x import y\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			body := newImportsWriter()
			tc.res.Write(body, tc.omit, tc.tcLines)
			if got := body.String(); got != tc.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tc.want)
			}
		})
	}
}

//nolint:funlen // Table test enumerating every ModelImports branch.
func TestModelImports(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		conf   *config.Config
		tables []model.Table
		want   ImportResult
	}{
		{
			name:   "no tables skips model library import",
			conf:   newImportsConfig(config.SQLDriverAsyncpg),
			tables: nil,
			want: ImportResult{
				Std:          []string{"import typing"},
				TypeChecking: []string{"import collections.abc"},
			},
		},
		{
			name:   "dataclass",
			conf:   newImportsConfig(config.SQLDriverAsyncpg),
			tables: []model.Table{impTable(impCol("id", model.PyType{SQLType: "int", Type: "int"}))},
			want: ImportResult{
				Std:          []string{"import dataclasses", "import typing"},
				TypeChecking: []string{"import collections.abc"},
			},
		},
		{
			name: "attrs",
			conf: newImportsConfig(config.SQLDriverAsyncpg, func(c *config.Config) { c.ModelType = config.ModelTypeAttrs }),
			tables: []model.Table{
				impTable(impCol("id", model.PyType{SQLType: "int", Type: "int"})),
			},
			want: ImportResult{
				Std:          []string{"import attrs", "import typing"},
				TypeChecking: []string{"import collections.abc"},
			},
		},
		{
			name: "msgspec",
			conf: newImportsConfig(config.SQLDriverAsyncpg, func(c *config.Config) { c.ModelType = config.ModelTypeMsgspec }),
			tables: []model.Table{
				impTable(impCol("id", model.PyType{SQLType: "int", Type: "int"})),
			},
			want: ImportResult{
				Std:          []string{"import msgspec", "import typing"},
				TypeChecking: []string{"import collections.abc"},
			},
		},
		{
			name: "pydantic keeps typing while typechecking block exists",
			conf: newImportsConfig(config.SQLDriverAsyncpg, func(c *config.Config) { c.ModelType = config.ModelTypePydantic }),
			tables: []model.Table{
				impTable(impCol("id", model.PyType{SQLType: "int", Type: "int"})),
			},
			want: ImportResult{
				Std:          []string{"import pydantic", "import typing"},
				TypeChecking: []string{"import collections.abc"},
			},
		},
		{
			name: "decimal column",
			conf: newImportsConfig(config.SQLDriverAsyncpg),
			tables: []model.Table{
				impTable(impCol("price", model.PyType{SQLType: "numeric", Type: "decimal.Decimal"})),
			},
			want: ImportResult{
				Std:          []string{"import dataclasses", "import typing"},
				TypeChecking: []string{"import collections.abc", "import decimal"},
			},
		},
		{
			name: "date column",
			conf: newImportsConfig(config.SQLDriverAsyncpg),
			tables: []model.Table{
				impTable(impCol("d", model.PyType{SQLType: "date", Type: typeDate})),
			},
			want: ImportResult{
				Std:          []string{"import dataclasses", "import typing"},
				TypeChecking: []string{"import collections.abc", "import datetime"},
			},
		},
		{
			name: "time column",
			conf: newImportsConfig(config.SQLDriverAsyncpg),
			tables: []model.Table{
				impTable(impCol("tm", model.PyType{SQLType: "time", Type: typeTime})),
			},
			want: ImportResult{
				Std:          []string{"import dataclasses", "import typing"},
				TypeChecking: []string{"import collections.abc", "import datetime"},
			},
		},
		{
			name: "datetime column",
			conf: newImportsConfig(config.SQLDriverAsyncpg),
			tables: []model.Table{
				impTable(impCol("ts", model.PyType{SQLType: "timestamp", Type: typeDatetime})),
			},
			want: ImportResult{
				Std:          []string{"import dataclasses", "import typing"},
				TypeChecking: []string{"import collections.abc", "import datetime"},
			},
		},
		{
			name: "timedelta column",
			conf: newImportsConfig(config.SQLDriverAsyncpg),
			tables: []model.Table{
				impTable(impCol("iv", model.PyType{SQLType: "interval", Type: "datetime.timedelta"})),
			},
			want: ImportResult{
				Std:          []string{"import dataclasses", "import typing"},
				TypeChecking: []string{"import collections.abc", "import datetime"},
			},
		},
		{
			name: "uuid column",
			conf: newImportsConfig(config.SQLDriverAsyncpg),
			tables: []model.Table{
				impTable(impCol("uid", model.PyType{SQLType: "uuid", Type: typeUUID})),
			},
			want: ImportResult{
				Std:          []string{"import dataclasses", "import typing"},
				TypeChecking: []string{"import collections.abc", "import uuid"},
			},
		},
		{
			name: "enum column non-pydantic goes to typechecking",
			conf: newImportsConfig(config.SQLDriverAsyncpg),
			tables: []model.Table{
				impTable(impCol("status", model.PyType{SQLType: "status", Type: "enums.Status", IsEnum: true})),
			},
			want: ImportResult{
				Std:          []string{"import dataclasses", "import typing"},
				TypeChecking: []string{"from db import enums", "import collections.abc"},
			},
		},
		{
			name: "enum column pydantic goes to package imports",
			conf: newImportsConfig(config.SQLDriverAsyncpg, func(c *config.Config) { c.ModelType = config.ModelTypePydantic }),
			tables: []model.Table{
				impTable(impCol("status", model.PyType{SQLType: "status", Type: "enums.Status", IsEnum: true})),
			},
			want: ImportResult{
				Std:          []string{"import pydantic", "import typing"},
				TypeChecking: []string{"import collections.abc"},
				Package:      []string{"from db import enums"},
			},
		},
		{
			name: "pydantic list column forces runtime collections and drops typing",
			conf: newImportsConfig(config.SQLDriverAsyncpg, func(c *config.Config) { c.ModelType = config.ModelTypePydantic }),
			tables: []model.Table{
				impTable(impCol("tags", model.PyType{SQLType: "text", Type: "str", IsList: true})),
			},
			want: ImportResult{
				Std: []string{"import collections.abc", "import pydantic"},
			},
		},
		{
			name: "pydantic datetime column forced to runtime",
			conf: newImportsConfig(config.SQLDriverAsyncpg, func(c *config.Config) { c.ModelType = config.ModelTypePydantic }),
			tables: []model.Table{
				impTable(impCol("d", model.PyType{SQLType: "date", Type: typeDate})),
			},
			want: ImportResult{
				Std:          []string{"import datetime", "import pydantic", "import typing"},
				TypeChecking: []string{"import collections.abc"},
			},
		},
		{
			name: "omit typechecking block drops unused typing",
			conf: newImportsConfig(config.SQLDriverAsyncpg, func(c *config.Config) { c.OmitTypecheckingBlock = true }),
			tables: []model.Table{
				impTable(impCol("id", model.PyType{SQLType: "int", Type: "int"})),
			},
			want: ImportResult{
				Std:          []string{"import dataclasses"},
				TypeChecking: []string{"import collections.abc"},
			},
		},
		{
			name: "omit typechecking block keeps typing for Any columns",
			conf: newImportsConfig(config.SQLDriverAsyncpg, func(c *config.Config) { c.OmitTypecheckingBlock = true }),
			tables: []model.Table{
				impTable(impCol("data", model.PyType{SQLType: "json", Type: "typing.Any"})),
			},
			want: ImportResult{
				Std:          []string{"import dataclasses", "import typing"},
				TypeChecking: []string{"import collections.abc"},
			},
		},
		{
			name: "used module override",
			conf: newImportsConfig(config.SQLDriverAsyncpg, func(c *config.Config) {
				c.Overrides = []config.Override{{PyType: config.OverridePyType{Import: "mymod", Type: "mymod.Money"}}}
			}),
			tables: []model.Table{
				impTable(impCol("price", model.PyType{
					SQLType: "numeric", Type: "mymod.Money", IsOverride: true, DefaultType: "decimal.Decimal",
				})),
			},
			want: ImportResult{
				Std:          []string{"import dataclasses", "import typing"},
				TypeChecking: []string{"import collections.abc", "import mymod"},
			},
		},
		{
			name: "used from-import override",
			conf: newImportsConfig(config.SQLDriverAsyncpg, func(c *config.Config) {
				c.Overrides = []config.Override{
					{PyType: config.OverridePyType{Import: "mycoll", Type: "UserString", Package: "UserString"}},
				}
			}),
			tables: []model.Table{
				impTable(impCol("name", model.PyType{SQLType: "text", Type: "UserString", IsOverride: true, DefaultType: "str"})),
			},
			want: ImportResult{
				Std:          []string{"import dataclasses", "import typing"},
				TypeChecking: []string{"from mycoll import UserString", "import collections.abc"},
			},
		},
		{
			name: "override without import is skipped",
			conf: newImportsConfig(config.SQLDriverAsyncpg, func(c *config.Config) {
				c.Overrides = []config.Override{{PyType: config.OverridePyType{Type: "noimp.T"}}}
			}),
			tables: []model.Table{
				impTable(impCol("x", model.PyType{SQLType: "text", Type: "noimp.T", IsOverride: true, DefaultType: "str"})),
			},
			want: ImportResult{
				Std:          []string{"import dataclasses", "import typing"},
				TypeChecking: []string{"import collections.abc"},
			},
		},
		{
			name: "override without type is skipped",
			conf: newImportsConfig(config.SQLDriverAsyncpg, func(c *config.Config) {
				c.Overrides = []config.Override{{PyType: config.OverridePyType{Import: "mymod"}}}
			}),
			tables: []model.Table{
				impTable(impCol("id", model.PyType{SQLType: "int", Type: "int"})),
			},
			want: ImportResult{
				Std:          []string{"import dataclasses", "import typing"},
				TypeChecking: []string{"import collections.abc"},
			},
		},
		{
			name: "unused override adds nothing",
			conf: newImportsConfig(config.SQLDriverAsyncpg, func(c *config.Config) {
				c.Overrides = []config.Override{{PyType: config.OverridePyType{Import: "mymod", Type: "mymod.Money"}}}
			}),
			tables: []model.Table{
				impTable(impCol("id", model.PyType{SQLType: "int", Type: "int"})),
			},
			want: ImportResult{
				Std:          []string{"import dataclasses", "import typing"},
				TypeChecking: []string{"import collections.abc"},
			},
		},
		{
			name: "override duplicating a std import is compacted",
			conf: newImportsConfig(config.SQLDriverAsyncpg, func(c *config.Config) {
				c.Overrides = []config.Override{{PyType: config.OverridePyType{Import: "datetime", Type: typeDate}}}
			}),
			tables: []model.Table{
				impTable(impCol("d", model.PyType{SQLType: "date", Type: typeDate})),
			},
			want: ImportResult{
				Std:          []string{"import dataclasses", "import typing"},
				TypeChecking: []string{"import collections.abc", "import datetime"},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			resolver := newImportsResolver(t, tc.conf)
			checkImportResult(t, resolver.ModelImports(tc.tables), tc.want)
		})
	}
}

//nolint:funlen,maintidx // Table test enumerating every QueryImports branch.
func TestQueryImports(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		conf    *config.Config
		queries []model.Query
		want    ImportResult
	}{
		{
			name: "asyncpg one scalar",
			conf: newImportsConfig(config.SQLDriverAsyncpg),
			queries: []model.Query{
				{Cmd: metadata.CmdOne, Returns: impScalar(model.PyType{SQLType: "int", Type: "int"})},
			},
			want: ImportResult{
				Std:          []string{"import typing"},
				TypeChecking: []string{"import asyncpg", "import collections.abc"},
			},
		},
		{
			name: "psycopg json return forces runtime module for loaders",
			conf: newImportsConfig(config.SQLDriverPsycopgAsync),
			queries: []model.Query{
				{Cmd: metadata.CmdOne, Returns: impScalar(model.PyType{SQLType: "jsonb", Type: "str"})},
			},
			want: ImportResult{
				Std: []string{"import psycopg", "import psycopg.rows", "import psycopg.types.string", "import typing"},
				TypeChecking: []string{
					"import collections.abc",
				},
			},
		},
		{
			name: "psycopg without json returns keeps the module lazy",
			conf: newImportsConfig(config.SQLDriverPsycopgAsync),
			queries: []model.Query{
				{Cmd: metadata.CmdOne, Returns: impScalar(model.PyType{SQLType: "bigint", Type: "int"})},
			},
			want: ImportResult{
				Std: []string{"import typing"},
				TypeChecking: []string{
					"import collections.abc",
					"import psycopg",
					"import psycopg.rows",
				},
			},
		},
		{
			name: "psycopg many simple return imports operator",
			conf: newImportsConfig(config.SQLDriverPsycopgAsync),
			queries: []model.Query{
				{Cmd: metadata.CmdMany, Returns: impScalar(model.PyType{SQLType: "bigint", Type: "int"})},
			},
			want: ImportResult{
				Std: []string{"import operator", "import typing"},
				TypeChecking: []string{
					"import collections.abc",
					"import psycopg",
					"import psycopg.rows\n",
					argsTypeBase + argsTypeTail,
				},
			},
		},
		{
			name: "asyncpg many simple return imports operator",
			conf: newImportsConfig(config.SQLDriverAsyncpg),
			queries: []model.Query{
				{Cmd: metadata.CmdMany, Returns: impScalar(model.PyType{SQLType: "int", Type: "int"})},
			},
			want: ImportResult{
				Std: []string{"import operator", "import typing"},
				TypeChecking: []string{
					"import asyncpg",
					"import asyncpg.cursor",
					"import collections.abc\n",
					argsTypeBase + argsTypeTail,
				},
			},
		},
		{
			name: "asyncpg many emitted struct with decimal uuid datetime",
			conf: newImportsConfig(config.SQLDriverAsyncpg),
			queries: []model.Query{
				{Cmd: metadata.CmdMany, Returns: impStruct(true,
					impCol("a", model.PyType{SQLType: "numeric", Type: "decimal.Decimal"}),
					impCol("b", model.PyType{SQLType: "pg_catalog.uuid", Type: typeUUID}),
					impCol("c", model.PyType{SQLType: "timestamp", Type: typeDatetime}),
				)},
			},
			want: ImportResult{
				Std: []string{"import dataclasses", "import typing"},
				TypeChecking: []string{
					"import asyncpg",
					"import asyncpg.cursor",
					"import collections.abc",
					"import datetime",
					"import decimal",
					"import uuid\n",
					argsTypeBase + " | decimal.Decimal | uuid.UUID" + argsTypeDatetime + argsTypeTail,
				},
			},
		},
		{
			name: "models prefix return imports models package",
			conf: newImportsConfig(config.SQLDriverAsyncpg),
			queries: []model.Query{
				{Cmd: metadata.CmdOne, Returns: model.QueryValue{
					Name: "author",
					Type: model.PyType{Type: "models.Author"},
					Table: &model.Table{Name: "models.Author", Columns: []model.Column{
						impCol("id", model.PyType{SQLType: "int", Type: "int"}),
					}},
				}},
			},
			want: ImportResult{
				Std:          []string{"import typing"},
				TypeChecking: []string{"import asyncpg", "import collections.abc"},
				Package:      []string{"from db import models"},
			},
		},
		{
			name: "enum param imports enums package",
			conf: newImportsConfig(config.SQLDriverAsyncpg),
			queries: []model.Query{
				{Cmd: metadata.CmdExec, Params: []model.QueryValue{
					impScalar(model.PyType{SQLType: "status", Type: "enums.Status", IsEnum: true}),
				}},
			},
			want: ImportResult{
				Std:          []string{"import typing"},
				TypeChecking: []string{"import asyncpg", "import collections.abc"},
				Package:      []string{"from db import enums"},
			},
		},
		{
			name: "enums prefix without enum flag imports enums package",
			conf: newImportsConfig(config.SQLDriverAsyncpg),
			queries: []model.Query{
				{Cmd: metadata.CmdOne, Returns: impScalar(model.PyType{SQLType: "status", Type: "enums.Status"})},
			},
			want: ImportResult{
				Std:          []string{"import typing"},
				TypeChecking: []string{"import asyncpg", "import collections.abc"},
				Package:      []string{"from db import enums"},
			},
		},
		{
			name: "overridden enum param still imports enums package",
			conf: newImportsConfig(config.SQLDriverAsyncpg),
			queries: []model.Query{
				{Cmd: metadata.CmdExec, Params: []model.QueryValue{
					impScalar(model.PyType{SQLType: "status", Type: "str", IsOverride: true, DefaultType: "enums.Status"}),
				}},
			},
			want: ImportResult{
				Std:          []string{"import typing"},
				TypeChecking: []string{"import asyncpg", "import collections.abc"},
				Package:      []string{"from db import enums"},
			},
		},
		{
			name: "override param forces default type module to runtime",
			conf: newImportsConfig(config.SQLDriverAsyncpg, func(c *config.Config) {
				c.Overrides = []config.Override{{PyType: config.OverridePyType{Import: "mymod", Type: "mytype"}}}
			}),
			queries: []model.Query{
				{
					Cmd:     metadata.CmdOne,
					Returns: impScalar(model.PyType{SQLType: "date", Type: typeDate}),
					Params: []model.QueryValue{
						impScalar(model.PyType{SQLType: "date", Type: "mytype", IsOverride: true, DefaultType: typeDate}),
					},
				},
			},
			want: ImportResult{
				// datetime is runtime (the param converts via DefaultType);
				// mymod only annotates, so it stays lazy.
				Std:          []string{"import datetime", "import typing"},
				TypeChecking: []string{"import asyncpg", "import collections.abc", "import mymod"},
			},
		},
		{
			name: "runtime datetime import keeps priority over later annotation use",
			conf: newImportsConfig(config.SQLDriverAsyncpg),
			queries: []model.Query{
				{Cmd: metadata.CmdExec, Params: []model.QueryValue{
					impScalar(model.PyType{SQLType: "date", Type: "mytype", IsOverride: true, DefaultType: typeDate}),
				}},
				{Cmd: metadata.CmdOne, Returns: impScalar(model.PyType{SQLType: "date", Type: typeDate})},
			},
			want: ImportResult{
				Std:          []string{"import datetime", "import typing"},
				TypeChecking: []string{"import asyncpg", "import collections.abc"},
			},
		},
		{
			name: "runtime module wins across datetime type variants",
			conf: newImportsConfig(config.SQLDriverAsyncpg),
			queries: []model.Query{
				{
					Cmd:     metadata.CmdOne,
					Returns: impScalar(model.PyType{SQLType: "time", Type: typeTime}),
					Params: []model.QueryValue{
						impScalar(model.PyType{SQLType: "date", Type: "mytype", IsOverride: true, DefaultType: typeDate}),
					},
				},
			},
			want: ImportResult{
				Std:          []string{"import datetime", "import typing"},
				TypeChecking: []string{"import asyncpg", "import collections.abc"},
			},
		},
		{
			name: "inline converted return forces runtime override import",
			conf: newImportsConfig(config.SQLDriverAsyncpg, func(c *config.Config) {
				c.Overrides = []config.Override{
					{PyType: config.OverridePyType{Import: "ipaddress", Type: "ipaddress.IPv4Address"}},
				}
			}),
			queries: []model.Query{
				{Cmd: metadata.CmdOne, Returns: impScalar(model.PyType{SQLType: "inet", Type: "ipaddress.IPv4Address"})},
			},
			want: ImportResult{
				Std:          []string{"import ipaddress", "import typing"},
				TypeChecking: []string{"import asyncpg", "import collections.abc"},
			},
		},
		{
			name: "inline converted struct column forces runtime override import",
			conf: newImportsConfig(config.SQLDriverAsyncpg, func(c *config.Config) {
				c.Overrides = []config.Override{
					{PyType: config.OverridePyType{Import: "ipaddress", Type: "ipaddress.IPv4Address"}},
				}
			}),
			queries: []model.Query{
				{Cmd: metadata.CmdOne, Returns: impStruct(true,
					impCol("ip", model.PyType{SQLType: "inet", Type: "ipaddress.IPv4Address"}),
				)},
			},
			want: ImportResult{
				Std:          []string{"import dataclasses", "import ipaddress", "import typing"},
				TypeChecking: []string{"import asyncpg", "import collections.abc"},
			},
		},
		{
			name: "asyncpg many overridden return skips operator",
			conf: newImportsConfig(config.SQLDriverAsyncpg, func(c *config.Config) {
				c.Overrides = []config.Override{{PyType: config.OverridePyType{Import: "mymod", Type: "mytype"}}}
			}),
			queries: []model.Query{
				{Cmd: metadata.CmdMany, Returns: impScalar(
					model.PyType{SQLType: "text", Type: "mytype", IsOverride: true, DefaultType: "str"},
				)},
			},
			want: ImportResult{
				Std: []string{"import mymod", "import typing"},
				TypeChecking: []string{
					"import asyncpg",
					"import asyncpg.cursor",
					"import collections.abc\n",
					argsTypeBase + argsTypeTail,
				},
			},
		},
		{
			name: "asyncpg many inline converted return skips operator",
			conf: newImportsConfig(config.SQLDriverAsyncpg),
			queries: []model.Query{
				{Cmd: metadata.CmdMany, Returns: impScalar(model.PyType{SQLType: "bytea", Type: "memoryview"})},
			},
			want: ImportResult{
				Std: []string{"import typing"},
				TypeChecking: []string{
					"import asyncpg",
					"import asyncpg.cursor",
					"import collections.abc\n",
					argsTypeBase + argsTypeTail,
				},
			},
		},
		{
			name: "asyncpg copyfrom imports model library",
			conf: newImportsConfig(config.SQLDriverAsyncpg),
			queries: []model.Query{
				{Cmd: metadata.CmdCopyFrom, Params: []model.QueryValue{
					impScalar(model.PyType{SQLType: "int", Type: "int"}),
				}},
			},
			want: ImportResult{
				Std:          []string{"import dataclasses", "import typing"},
				TypeChecking: []string{"import asyncpg", "import collections.abc"},
			},
		},
		{
			name: "aiosqlite adapter param registers module at runtime",
			conf: newImportsConfig(config.SQLDriverAioSQLite),
			queries: []model.Query{
				{Cmd: metadata.CmdExec, Params: []model.QueryValue{
					impScalar(model.PyType{SQLType: "date", Type: typeDate}),
				}},
			},
			want: ImportResult{
				Std:          []string{"import aiosqlite", "import datetime", "import typing"},
				TypeChecking: []string{"import collections.abc"},
			},
		},
		{
			name: "aiosqlite without conversions stays typechecking",
			conf: newImportsConfig(config.SQLDriverAioSQLite),
			queries: []model.Query{
				{Cmd: metadata.CmdOne, Returns: impScalar(model.PyType{SQLType: "text", Type: "str"})},
			},
			want: ImportResult{
				Std:          []string{"import typing"},
				TypeChecking: []string{"import aiosqlite", "import collections.abc"},
			},
		},
		{
			name: "aiosqlite many simple return imports operator and sqlite3",
			conf: newImportsConfig(config.SQLDriverAioSQLite),
			queries: []model.Query{
				{Cmd: metadata.CmdMany, Returns: impScalar(model.PyType{SQLType: "int", Type: "int"})},
			},
			want: ImportResult{
				Std: []string{"import operator", "import typing"},
				TypeChecking: []string{
					"import aiosqlite",
					"import collections.abc",
					"import sqlite3\n",
					argsTypeBase + argsTypeTail,
				},
			},
		},
		{
			name: "aiosqlite many struct return skips operator",
			conf: newImportsConfig(config.SQLDriverAioSQLite),
			queries: []model.Query{
				{Cmd: metadata.CmdMany, Returns: impStruct(true,
					impCol("id", model.PyType{SQLType: "int", Type: "int"}),
				)},
			},
			want: ImportResult{
				Std: []string{"import dataclasses", "import typing"},
				TypeChecking: []string{
					"import aiosqlite",
					"import collections.abc",
					"import sqlite3\n",
					argsTypeBase + argsTypeTail,
				},
			},
		},
		{
			name: "aiosqlite adapter for unannotated module adds plain import",
			conf: newImportsConfig(config.SQLDriverAioSQLite),
			queries: []model.Query{
				{Cmd: metadata.CmdExec, Params: []model.QueryValue{
					impScalar(model.PyType{SQLType: "date", Type: "str"}),
				}},
			},
			want: ImportResult{
				Std:          []string{"import aiosqlite", "import datetime", "import typing"},
				TypeChecking: []string{"import collections.abc"},
			},
		},
		{
			name: "aiosqlite adapter for overridden param already at runtime",
			conf: newImportsConfig(config.SQLDriverAioSQLite),
			queries: []model.Query{
				{Cmd: metadata.CmdExec, Params: []model.QueryValue{
					impScalar(model.PyType{SQLType: "date", Type: "mytype", IsOverride: true, DefaultType: typeDate}),
				}},
			},
			want: ImportResult{
				Std:          []string{"import aiosqlite", "import datetime", "import typing"},
				TypeChecking: []string{"import collections.abc"},
			},
		},
		{
			name: "sqlite3 converter return registers module at runtime",
			conf: newImportsConfig(config.SQLDriverSQLite),
			queries: []model.Query{
				{Cmd: metadata.CmdOne, Returns: impScalar(model.PyType{SQLType: "date", Type: typeDate})},
			},
			want: ImportResult{
				Std:          []string{"import datetime", "import sqlite3", "import typing"},
				TypeChecking: []string{"import collections.abc"},
			},
		},
		{
			name: "sqlite3 without conversions stays typechecking",
			conf: newImportsConfig(config.SQLDriverSQLite),
			queries: []model.Query{
				{Cmd: metadata.CmdOne, Returns: impScalar(model.PyType{SQLType: "text", Type: "str"})},
			},
			want: ImportResult{
				Std:          []string{"import typing"},
				TypeChecking: []string{"import collections.abc", "import sqlite3"},
			},
		},
		{
			name: "sqlite3 many simple return imports operator",
			conf: newImportsConfig(config.SQLDriverSQLite),
			queries: []model.Query{
				{Cmd: metadata.CmdMany, Returns: impScalar(model.PyType{SQLType: "int", Type: "int"})},
			},
			want: ImportResult{
				Std: []string{"import operator", "import typing"},
				TypeChecking: []string{
					"import collections.abc",
					"import sqlite3\n",
					argsTypeBase + argsTypeTail,
				},
			},
		},
		{
			name: "sqlite3 many enum return skips operator",
			conf: newImportsConfig(config.SQLDriverSQLite),
			queries: []model.Query{
				{Cmd: metadata.CmdMany, Returns: impScalar(model.PyType{SQLType: "status", Type: "enums.Status", IsEnum: true})},
			},
			want: ImportResult{
				Std: []string{"import typing"},
				TypeChecking: []string{
					"import collections.abc",
					"import sqlite3\n",
					argsTypeBase + argsTypeTail,
				},
				Package: []string{"from db import enums"},
			},
		},
		{
			name: "sqlite3 speedups datetime converter keeps datetime lazy",
			conf: newImportsConfig(config.SQLDriverSQLite, func(c *config.Config) { c.Speedups = true }),
			queries: []model.Query{
				{Cmd: metadata.CmdOne, Returns: impScalar(model.PyType{SQLType: "timestamp", Type: typeDatetime})},
			},
			want: ImportResult{
				Std:          []string{"import ciso8601", "import sqlite3", "import typing"},
				TypeChecking: []string{"import collections.abc", "import datetime"},
			},
		},
		{
			name: "sqlite3 speedups decimal converter needs no ciso8601",
			conf: newImportsConfig(config.SQLDriverSQLite, func(c *config.Config) { c.Speedups = true }),
			queries: []model.Query{
				{Cmd: metadata.CmdOne, Returns: impScalar(model.PyType{SQLType: "decimal", Type: "decimal.Decimal"})},
			},
			want: ImportResult{
				Std:          []string{"import decimal", "import sqlite3", "import typing"},
				TypeChecking: []string{"import collections.abc"},
			},
		},
		{
			name: "pydantic emitted struct with list field empties typechecking",
			conf: newImportsConfig(config.SQLDriverSQLite, func(c *config.Config) { c.ModelType = config.ModelTypePydantic }),
			queries: []model.Query{
				{Cmd: metadata.CmdMany, Returns: impStruct(true,
					impCol("tags", model.PyType{SQLType: "text", Type: "str", IsList: true}),
					impCol("d", model.PyType{SQLType: "date", Type: typeDate}),
				)},
			},
			want: ImportResult{
				Std: []string{
					"import collections.abc",
					"import datetime",
					"import pydantic",
					"import sqlite3",
					"import typing",
				},
				TypeChecking: []string{argsTypeBase + argsTypeDatetime + argsTypeTail},
			},
		},
		{
			name: "pydantic emitted struct without list keeps collections lazy",
			conf: newImportsConfig(config.SQLDriverAsyncpg, func(c *config.Config) { c.ModelType = config.ModelTypePydantic }),
			queries: []model.Query{
				{Cmd: metadata.CmdOne, Returns: impStruct(true,
					impCol("d", model.PyType{SQLType: "date", Type: typeDate}),
				)},
			},
			want: ImportResult{
				Std:          []string{"import datetime", "import pydantic", "import typing"},
				TypeChecking: []string{"import asyncpg", "import collections.abc"},
			},
		},
		{
			name: "pydantic without emitted models keeps lazy imports",
			conf: newImportsConfig(config.SQLDriverAsyncpg, func(c *config.Config) { c.ModelType = config.ModelTypePydantic }),
			queries: []model.Query{
				{Cmd: metadata.CmdOne, Returns: impScalar(model.PyType{SQLType: "date", Type: typeDate})},
			},
			want: ImportResult{
				Std:          []string{"import typing"},
				TypeChecking: []string{"import asyncpg", "import collections.abc", "import datetime"},
			},
		},
		{
			name: "embed columns drive conversion imports and enums",
			conf: newImportsConfig(config.SQLDriverAsyncpg),
			queries: []model.Query{
				{Cmd: metadata.CmdOne, Returns: impStruct(true,
					impCol("id", model.PyType{SQLType: "int", Type: "int"}),
					model.Column{Name: "author", Embed: &model.Embed{ModelName: "Author", Columns: []model.Column{
						impCol("d", model.PyType{SQLType: "date", Type: typeDate, IsOverride: true, DefaultType: "str"}),
						impCol("tags", model.PyType{SQLType: "text", Type: "str", IsList: true}),
						impCol("status", model.PyType{SQLType: "status", Type: "enums.Status", IsEnum: true}),
					}}},
				)},
			},
			want: ImportResult{
				Std:          []string{"import dataclasses", "import datetime", "import typing"},
				TypeChecking: []string{"import asyncpg", "import collections.abc"},
				Package:      []string{"from db import enums"},
			},
		},
		{
			name: "param struct with overridden column imports default type at runtime",
			conf: newImportsConfig(config.SQLDriverAsyncpg),
			queries: []model.Query{
				{Cmd: metadata.CmdExec, Params: []model.QueryValue{
					impStruct(true,
						impCol("d", model.PyType{SQLType: "date", Type: typeDate}),
						impCol("amount", model.PyType{
							SQLType: "numeric", Type: "mymoney", IsOverride: true, DefaultType: "decimal.Decimal",
						}),
					),
				}},
			},
			want: ImportResult{
				Std:          []string{"import dataclasses", "import decimal", "import typing"},
				TypeChecking: []string{"import asyncpg", "import collections.abc", "import datetime"},
			},
		},
		{
			name:    "converter return imports the converter module and keeps the type lazy",
			conf:    newImportsConfig(config.SQLDriverAsyncpg, impConverterConf),
			queries: []model.Query{{Cmd: metadata.CmdOne, Returns: impScalar(impConverter("numeric", "decimal.Decimal"))}},
			want: ImportResult{
				// decimal never appears: the converter replaces the DefaultType call.
				Std:          []string{"import myconv", "import typing"},
				TypeChecking: []string{"import asyncpg", "import collections.abc", "import mymod"},
			},
		},
		{
			name: "converter param imports the converter module instead of the default type",
			conf: newImportsConfig(config.SQLDriverAsyncpg, impConverterConf),
			queries: []model.Query{
				{Cmd: metadata.CmdExec, Params: []model.QueryValue{impScalar(impConverter("date", typeDate))}},
			},
			want: ImportResult{
				Std:          []string{"import myconv", "import typing"},
				TypeChecking: []string{"import asyncpg", "import collections.abc", "import mymod"},
			},
		},
		{
			name: "converter struct column return imports the converter module",
			conf: newImportsConfig(config.SQLDriverAsyncpg, impConverterConf),
			queries: []model.Query{
				{Cmd: metadata.CmdOne, Returns: impStruct(true, impCol("amount", impConverter("inet", "str")))},
			},
			want: ImportResult{
				Std:          []string{"import dataclasses", "import myconv", "import typing"},
				TypeChecking: []string{"import asyncpg", "import collections.abc", "import mymod"},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			resolver := newImportsResolver(t, tc.conf)
			checkImportResult(t, resolver.QueryImports(tc.queries), tc.want)
		})
	}
}

func TestEnumImports(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		omit bool
		want ImportResult
	}{
		{
			name: "default",
			want: ImportResult{
				Std:          []string{"import enum", "import typing"},
				TypeChecking: []string{"import collections.abc"},
			},
		},
		{
			name: "omit typechecking block drops typing",
			omit: true,
			want: ImportResult{
				Std:          []string{"import enum"},
				TypeChecking: []string{"import collections.abc"},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			conf := newImportsConfig(config.SQLDriverAsyncpg, func(c *config.Config) { c.OmitTypecheckingBlock = tc.omit })
			resolver := newImportsResolver(t, conf)
			checkImportResult(t, resolver.EnumImports(), tc.want)
		})
	}
}

func TestQueryValueMatches(t *testing.T) {
	t.Parallel()
	pred := func(typ model.PyType) bool { return typ.Type == "target" }
	cases := []struct {
		name string
		qv   model.QueryValue
		want bool
	}{
		{name: "empty value", qv: model.QueryValue{}, want: false},
		{name: "scalar match", qv: impScalar(model.PyType{Type: "target"}), want: true},
		{name: "scalar no match", qv: impScalar(model.PyType{Type: "other"}), want: false},
		{name: "struct column match", qv: impStruct(false, impCol("a", model.PyType{Type: "target"})), want: true},
		{
			name: "struct embed column match",
			qv: impStruct(false, model.Column{Name: "e", Embed: &model.Embed{ModelName: "M", Columns: []model.Column{
				impCol("a", model.PyType{Type: "target"}),
			}}}),
			want: true,
		},
		{name: "struct no match", qv: impStruct(false, impCol("a", model.PyType{Type: "other"})), want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := queryValueMatches(tc.qv, pred); got != tc.want {
				t.Errorf("queryValueMatches() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestAnyQueryTypeAndAnyParamType(t *testing.T) {
	t.Parallel()
	pred := func(typ model.PyType) bool { return typ.Type == "target" }
	returnMatch := []model.Query{{Cmd: metadata.CmdOne, Returns: impScalar(model.PyType{Type: "target"})}}
	paramMatch := []model.Query{{Cmd: metadata.CmdExec, Params: []model.QueryValue{impScalar(model.PyType{Type: "target"})}}}
	noMatch := []model.Query{{
		Cmd:     metadata.CmdOne,
		Returns: impScalar(model.PyType{Type: "other"}),
		Params:  []model.QueryValue{impScalar(model.PyType{Type: "other"})},
	}}
	cases := []struct {
		name      string
		queries   []model.Query
		wantAny   bool
		wantParam bool
	}{
		{name: "return match", queries: returnMatch, wantAny: true, wantParam: false},
		{name: "param match", queries: paramMatch, wantAny: true, wantParam: true},
		{name: "no match", queries: noMatch, wantAny: false, wantParam: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := anyQueryType(tc.queries, pred); got != tc.wantAny {
				t.Errorf("anyQueryType() = %v, want %v", got, tc.wantAny)
			}
			if got := anyParamType(tc.queries, pred); got != tc.wantParam {
				t.Errorf("anyParamType() = %v, want %v", got, tc.wantParam)
			}
		})
	}
}

func TestEmittedModelFields(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		queries   []model.Query
		wantEmits bool
		wantList  bool
	}{
		{name: "no queries", queries: nil, wantEmits: false, wantList: false},
		{
			name:      "scalar values only",
			queries:   []model.Query{{Cmd: metadata.CmdOne, Returns: impScalar(model.PyType{Type: "int"})}},
			wantEmits: false,
			wantList:  false,
		},
		{
			name: "struct without emit flag",
			queries: []model.Query{
				{Cmd: metadata.CmdOne, Returns: impStruct(false, impCol("a", model.PyType{Type: "int", IsList: true}))},
			},
			wantEmits: false,
			wantList:  false,
		},
		{
			name: "emitted return struct without list",
			queries: []model.Query{
				{Cmd: metadata.CmdOne, Returns: impStruct(true, impCol("a", model.PyType{Type: "int"}))},
			},
			wantEmits: true,
			wantList:  false,
		},
		{
			name: "emitted param struct with list column",
			queries: []model.Query{
				{Cmd: metadata.CmdExec, Params: []model.QueryValue{
					impStruct(true, impCol("a", model.PyType{Type: "int", IsList: true})),
				}},
			},
			wantEmits: true,
			wantList:  true,
		},
		{
			name: "emitted struct with list embed column",
			queries: []model.Query{
				{Cmd: metadata.CmdOne, Returns: impStruct(true,
					model.Column{Name: "e", Embed: &model.Embed{ModelName: "M", Columns: []model.Column{
						impCol("a", model.PyType{Type: "int", IsList: true}),
					}}},
				)},
			},
			wantEmits: true,
			wantList:  true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gotEmits, gotList := emittedModelFields(tc.queries)
			if gotEmits != tc.wantEmits || gotList != tc.wantList {
				t.Errorf("emittedModelFields() = (%v, %v), want (%v, %v)", gotEmits, gotList, tc.wantEmits, tc.wantList)
			}
		})
	}
}

func TestOverrideDefaultTypeUses(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		qv   model.QueryValue
		want bool
	}{
		{name: "empty value", qv: model.QueryValue{}, want: false},
		{
			name: "scalar override match",
			qv:   impScalar(model.PyType{Type: "mytype", IsOverride: true, DefaultType: typeDate}),
			want: true,
		},
		{
			name: "scalar override different default",
			qv:   impScalar(model.PyType{Type: "mytype", IsOverride: true, DefaultType: "str"}),
			want: false,
		},
		{name: "scalar without override", qv: impScalar(model.PyType{Type: typeDate}), want: false},
		{
			name: "struct column override match",
			qv:   impStruct(true, impCol("a", model.PyType{Type: "mytype", IsOverride: true, DefaultType: typeDate})),
			want: true,
		},
		{
			name: "struct without override match",
			qv:   impStruct(true, impCol("a", model.PyType{Type: typeDate})),
			want: false,
		},
		{
			// The converter replaces the DefaultType call, so its module is never imported.
			name: "scalar converter override is skipped",
			qv:   impScalar(impConverter("date", typeDate)),
			want: false,
		},
		{
			name: "struct column converter override is skipped",
			qv:   impStruct(true, impCol("a", impConverter("date", typeDate))),
			want: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := overrideDefaultTypeUses(typeDate, tc.qv); got != tc.want {
				t.Errorf("overrideDefaultTypeUses() = %v, want %v", got, tc.want)
			}
		})
	}
}

//nolint:funlen // Table test enumerating every addConverterImports branch.
func TestAddConverterImports(t *testing.T) {
	t.Parallel()
	conv := func(sqlType, toDB, fromDB string) model.PyType {
		return model.PyType{
			SQLType: sqlType, Type: typeMoney, IsOverride: true,
			DefaultType: types.Str, ConverterTo: toDB, ConverterFrom: fromDB,
		}
	}
	runtimeSpec := func(module string) importSpec {
		return importSpec{Module: module, TypeChecking: false}
	}
	cases := []struct {
		name             string
		typeChecking     map[string]importSpec
		queries          []model.Query
		wantStd          map[string]importSpec
		wantTypeChecking map[string]importSpec
	}{
		{
			name: "no converter type used adds nothing",
			queries: []model.Query{
				{Cmd: metadata.CmdOne, Returns: impScalar(model.PyType{SQLType: "int", Type: "int"})},
			},
			wantStd: map[string]importSpec{},
		},
		{
			name: "converter param imports the to_db module at runtime",
			queries: []model.Query{
				{Cmd: metadata.CmdExec, Params: []model.QueryValue{impScalar(conv("numeric", converterToDB, converterFromDB))}},
			},
			wantStd: map[string]importSpec{converterModule: runtimeSpec(converterModule)},
		},
		{
			name: "converter return imports the from_db module at runtime",
			queries: []model.Query{
				{Cmd: metadata.CmdOne, Returns: impScalar(conv("numeric", converterToDB, converterFromDB))},
			},
			wantStd: map[string]importSpec{converterModule: runtimeSpec(converterModule)},
		},
		{
			name: "converter struct column is found through the table walk",
			queries: []model.Query{
				{
					Cmd:     metadata.CmdOne,
					Returns: impStruct(true, impCol("amount", conv("numeric", converterToDB, converterFromDB))),
				},
			},
			wantStd: map[string]importSpec{converterModule: runtimeSpec(converterModule)},
		},
		{
			name: "read-only query imports only the from_db module",
			queries: []model.Query{
				{Cmd: metadata.CmdOne, Returns: impScalar(conv("numeric", "encode.enc", "decode.dec"))},
			},
			wantStd: map[string]importSpec{"decode": runtimeSpec("decode")},
		},
		{
			name: "write-only query imports only the to_db module",
			queries: []model.Query{
				{Cmd: metadata.CmdExec, Params: []model.QueryValue{impScalar(conv("numeric", "encode.enc", "decode.dec"))}},
			},
			wantStd: map[string]importSpec{"encode": runtimeSpec("encode")},
		},
		{
			name: "both directions used imports both modules",
			queries: []model.Query{
				{Cmd: metadata.CmdOne, Returns: impScalar(conv("numeric", "encode.enc", "decode.dec"))},
				{Cmd: metadata.CmdExec, Params: []model.QueryValue{impScalar(conv("numeric", "encode.enc", "decode.dec"))}},
			},
			wantStd: map[string]importSpec{"encode": runtimeSpec("encode"), "decode": runtimeSpec("decode")},
		},
		{
			name:         "runtime converter module drops a colliding annotation-only import",
			typeChecking: map[string]importSpec{typeMoney: {Module: "mymod", TypeChecking: true}},
			queries: []model.Query{
				{Cmd: metadata.CmdOne, Returns: impScalar(conv("numeric", "mymod.enc", "mymod.dec"))},
			},
			wantStd:          map[string]importSpec{"mymod": runtimeSpec("mymod")},
			wantTypeChecking: map[string]importSpec{},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			resolver := newImportsResolver(t, newImportsConfig(config.SQLDriverAsyncpg))
			std := map[string]importSpec{}
			typeChecking := tc.typeChecking
			if typeChecking == nil {
				typeChecking = map[string]importSpec{}
			}
			resolver.addConverterImports(std, typeChecking, tc.queries)
			if !maps.Equal(std, tc.wantStd) {
				t.Errorf("std = %+v, want %+v", std, tc.wantStd)
			}
			if tc.wantTypeChecking != nil && !maps.Equal(typeChecking, tc.wantTypeChecking) {
				t.Errorf("typeChecking = %+v, want %+v", typeChecking, tc.wantTypeChecking)
			}
		})
	}
}

func TestHasSimpleReturn(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		query model.Query
		want  bool
	}{
		{
			name:  "non-many command",
			query: model.Query{Cmd: metadata.CmdOne, Returns: impScalar(model.PyType{SQLType: "int", Type: "int"})},
			want:  false,
		},
		{
			name: "struct return",
			query: model.Query{
				Cmd:     metadata.CmdMany,
				Returns: impStruct(true, impCol("a", model.PyType{SQLType: "int", Type: "int"})),
			},
			want: false,
		},
		{
			name: "enum return",
			query: model.Query{
				Cmd:     metadata.CmdMany,
				Returns: impScalar(model.PyType{SQLType: "status", Type: "enums.Status", IsEnum: true}),
			},
			want: false,
		},
		{
			name: "overridden return",
			query: model.Query{
				Cmd:     metadata.CmdMany,
				Returns: impScalar(model.PyType{SQLType: "text", Type: "mytype", IsOverride: true, DefaultType: "str"}),
			},
			want: false,
		},
		{
			name:  "inline converted return",
			query: model.Query{Cmd: metadata.CmdMany, Returns: impScalar(model.PyType{SQLType: "bytea", Type: "memoryview"})},
			want:  false,
		},
		{
			name:  "simple scalar return",
			query: model.Query{Cmd: metadata.CmdMany, Returns: impScalar(model.PyType{SQLType: "int", Type: "int"})},
			want:  true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			resolver := newImportsResolver(t, newImportsConfig(config.SQLDriverAsyncpg))
			if got := resolver.hasSimpleReturn([]model.Query{tc.query}); got != tc.want {
				t.Errorf("hasSimpleReturn() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestQueryValueUses(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		lookup   string
		qv       model.QueryValue
		isReturn bool
		wantUsed bool
		wantTC   bool
	}{
		{name: "empty value", lookup: "int", qv: model.QueryValue{}, isReturn: true, wantUsed: false, wantTC: false},
		{
			name:     "scalar annotation only",
			lookup:   typeDate,
			qv:       impScalar(model.PyType{SQLType: "date", Type: typeDate}),
			isReturn: true,
			wantUsed: true,
			wantTC:   true,
		},
		{
			name:     "scalar inline converted",
			lookup:   "ipaddress.IPv4Address",
			qv:       impScalar(model.PyType{SQLType: "inet", Type: "ipaddress.IPv4Address"}),
			isReturn: true,
			wantUsed: true,
			wantTC:   false,
		},
		{
			name:     "scalar overridden",
			lookup:   "mytype",
			qv:       impScalar(model.PyType{SQLType: "text", Type: "mytype", IsOverride: true, DefaultType: "str"}),
			isReturn: true,
			wantUsed: true,
			wantTC:   false,
		},
		{
			name:     "scalar no match",
			lookup:   typeDate,
			qv:       impScalar(model.PyType{SQLType: "int", Type: "int"}),
			isReturn: true,
			wantUsed: false,
			wantTC:   false,
		},
		{
			name:     "struct no match",
			lookup:   typeDate,
			qv:       impStruct(true, impCol("a", model.PyType{SQLType: "int", Type: "int"})),
			isReturn: true,
			wantUsed: false,
			wantTC:   false,
		},
		{
			name:     "struct annotation only column",
			lookup:   typeDate,
			qv:       impStruct(true, impCol("a", model.PyType{SQLType: "date", Type: typeDate})),
			isReturn: true,
			wantUsed: true,
			wantTC:   true,
		},
		{
			name:   "struct embed overridden column forces runtime",
			lookup: typeDate,
			qv: impStruct(true,
				impCol("a", model.PyType{SQLType: "date", Type: typeDate}),
				model.Column{Name: "e", Embed: &model.Embed{ModelName: "M", Columns: []model.Column{
					impCol("d", model.PyType{SQLType: "date", Type: typeDate, IsOverride: true, DefaultType: "str"}),
				}}},
			),
			isReturn: true,
			wantUsed: true,
			wantTC:   false,
		},
		{
			name:     "param overridden stays annotation only",
			lookup:   "mytype",
			qv:       impScalar(model.PyType{SQLType: "text", Type: "mytype", IsOverride: true, DefaultType: "str"}),
			isReturn: false,
			wantUsed: true,
			wantTC:   true,
		},
		{
			name:     "param inline converted stays annotation only",
			lookup:   "ipaddress.IPv4Address",
			qv:       impScalar(model.PyType{SQLType: "inet", Type: "ipaddress.IPv4Address"}),
			isReturn: false,
			wantUsed: true,
			wantTC:   true,
		},
		{
			name:   "param struct overridden column stays annotation only",
			lookup: typeDate,
			qv: impStruct(true,
				impCol("d", model.PyType{SQLType: "date", Type: typeDate, IsOverride: true, DefaultType: "str"}),
			),
			isReturn: false,
			wantUsed: true,
			wantTC:   true,
		},
		{
			// "inet" converts inline for asyncpg, so only the converter keeps this lazy.
			name:     "scalar converter return stays annotation only",
			lookup:   typeMoney,
			qv:       impScalar(impConverter("inet", "str")),
			isReturn: true,
			wantUsed: true,
			wantTC:   true,
		},
		{
			name:     "struct converter column return stays annotation only",
			lookup:   typeMoney,
			qv:       impStruct(true, impCol("amount", impConverter("inet", "str"))),
			isReturn: true,
			wantUsed: true,
			wantTC:   true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			resolver := newImportsResolver(t, newImportsConfig(config.SQLDriverAsyncpg))
			gotUsed, gotTC := resolver.queryValueUses(tc.lookup, tc.qv, tc.isReturn)
			if gotUsed != tc.wantUsed || gotTC != tc.wantTC {
				t.Errorf("queryValueUses() = (%v, %v), want (%v, %v)", gotUsed, gotTC, tc.wantUsed, tc.wantTC)
			}
		})
	}
}

func TestAddWithPriority(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		start map[string]importSpec
		spec  importSpec
		want  importSpec
	}{
		{
			name:  "adds missing key",
			start: map[string]importSpec{},
			spec:  importSpec{Module: "mod", TypeChecking: true},
			want:  importSpec{Module: "mod", TypeChecking: true},
		},
		{
			name:  "runtime replaces typechecking",
			start: map[string]importSpec{"mod": {Module: "mod", TypeChecking: true}},
			spec:  importSpec{Module: "mod", TypeChecking: false},
			want:  importSpec{Module: "mod", TypeChecking: false},
		},
		{
			name:  "existing runtime import is kept",
			start: map[string]importSpec{"mod": {Module: "mod", TypeChecking: false}},
			spec:  importSpec{Module: "mod", TypeChecking: true},
			want:  importSpec{Module: "mod", TypeChecking: false},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			addWithPriority(tc.start, "mod", tc.spec)
			if got := tc.start["mod"]; got != tc.want {
				t.Errorf("addWithPriority() stored %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestSplitTypeChecking(t *testing.T) {
	t.Parallel()
	specs := map[string]importSpec{
		"rt": {Module: "rt", TypeChecking: false},
		"tc": {Module: "tc", TypeChecking: true},
	}
	runtime, typeChecking := splitTypeChecking(specs)
	wantRuntime := map[string]importSpec{"rt": {Module: "rt", TypeChecking: false}}
	wantTC := map[string]importSpec{"tc": {Module: "tc", TypeChecking: true}}
	if !maps.Equal(runtime, wantRuntime) {
		t.Errorf("runtime = %+v, want %+v", runtime, wantRuntime)
	}
	if !maps.Equal(typeChecking, wantTC) {
		t.Errorf("typeChecking = %+v, want %+v", typeChecking, wantTC)
	}
}

func TestMergeMaps(t *testing.T) {
	t.Parallel()
	first := map[string]importSpec{"a": {Module: "a"}, "b": {Module: "b", TypeChecking: true}}
	second := map[string]importSpec{"b": {Module: "b", TypeChecking: false}}
	want := map[string]importSpec{"a": {Module: "a"}, "b": {Module: "b", TypeChecking: false}}
	if got := mergeMaps(first, second); !maps.Equal(got, want) {
		t.Errorf("mergeMaps() = %+v, want %+v", got, want)
	}
}

func TestBuildImportBlock(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		specs map[string]importSpec
		want  []string
	}{
		{name: "empty map returns nil", specs: map[string]importSpec{}, want: nil},
		{
			name: "sorted with duplicates compacted",
			specs: map[string]importSpec{
				"beta.dup": {Module: "beta"},
				"beta":     {Module: "beta"},
				"alpha":    {Module: "alpha"},
			},
			want: []string{"import alpha", "import beta"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := buildImportBlock(tc.specs); !slices.Equal(got, tc.want) {
				t.Errorf("buildImportBlock() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestPassthroughParamTypes(t *testing.T) {
	t.Parallel()
	passthrough := func(name string) model.PyType {
		return model.PyType{SQLType: "weird", Type: name, IsOverride: true, DefaultType: types.Any}
	}
	cases := []struct {
		name    string
		queries []model.Query
		want    []string
	}{
		{name: "no queries", queries: nil, want: []string{}},
		{
			name: "override with a known default type is converted back",
			queries: []model.Query{
				{
					Params: []model.QueryValue{
						impScalar(model.PyType{SQLType: "text", Type: "mytype", IsOverride: true, DefaultType: types.Str}),
					},
				},
			},
			want: []string{},
		},
		{
			name:    "plain param",
			queries: []model.Query{{Params: []model.QueryValue{impScalar(model.PyType{SQLType: "text", Type: types.Str})}}},
			want:    []string{},
		},
		{
			name:    "passthrough override",
			queries: []model.Query{{Params: []model.QueryValue{impScalar(passthrough("pathlib.PurePosixPath"))}}},
			want:    []string{"pathlib.PurePosixPath"},
		},
		{
			name: "deduplicated and sorted across queries",
			queries: []model.Query{
				{Params: []model.QueryValue{impScalar(passthrough("zeta.Type"))}},
				{Params: []model.QueryValue{impScalar(passthrough("alpha.Type")), impScalar(passthrough("zeta.Type"))}},
			},
			want: []string{"alpha.Type", "zeta.Type"},
		},
		{
			name:    "struct param columns are scanned",
			queries: []model.Query{{Params: []model.QueryValue{impStruct(true, impCol("c", passthrough("beta.Type")))}}},
			want:    []string{"beta.Type"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := passthroughParamTypes(tc.queries); !slices.Equal(got, tc.want) {
				t.Errorf("passthroughParamTypes() = %v, want %v", got, tc.want)
			}
		})
	}
}
