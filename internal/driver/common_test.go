package driver

import (
	"slices"
	"strings"
	"testing"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/types"
	"github.com/rayakame/sqlc-gen-better-python/internal/utils"
	"github.com/rayakame/sqlc-gen-better-python/internal/writer"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
)

// commonConnExpr is the connection expression writeFuncSignature returns in
// functions mode.
const commonConnExpr = "conn"

// paramConverterTo is the fully qualified to_db function a converter override carries.
const paramConverterTo = "myconv.to_db"

func commonTestConfig() *config.Config {
	return &config.Config{
		SqlDriver:           config.SQLDriverAsyncpg,
		EmitDocstrings:      config.DocstringConventionNone,
		EmitDocstringsSQL:   utils.ToPtr(true),
		IndentChar:          " ",
		CharsPerIndentLevel: 4,
	}
}

func TestWriteFuncSignature(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		drv         Driver
		emitClasses bool
		kwargsLimit int
		query       model.Query
		annotation  string
		want        string
		wantConn    string
	}{
		{
			name: "async functions mode without params",
			drv:  newAsyncpgDriver(),
			query: model.Query{
				Cmd:      metadata.CmdOne,
				FuncName: "get_author",
			},
			annotation: "models.Author | None",
			want:       "async def get_author(conn: ConnectionLike) -> models.Author | None:\n",
			wantConn:   commonConnExpr,
		},
		{
			name:        "classes mode adds star over kwargs limit",
			drv:         newAsyncpgDriver(),
			emitClasses: true,
			query: model.Query{
				Cmd:      metadata.CmdExec,
				FuncName: "create_author",
				Params: []model.QueryValue{
					{Name: "name", Type: model.PyType{Type: "str", SQLType: "text"}},
				},
			},
			annotation: "None",
			want:       "async def create_author(self, *, name: str) -> None:\n",
			wantConn:   "self._conn",
		},
		{
			name:        "many command skips async prefix",
			drv:         newAsyncpgDriver(),
			kwargsLimit: 8,
			query: model.Query{
				Cmd:      metadata.CmdMany,
				FuncName: "list_ids",
				Params: []model.QueryValue{
					{Name: "author_id", Type: model.PyType{Type: "int", SQLType: "bigint"}},
				},
			},
			annotation: "QueryResults[int]",
			want:       "def list_ids(conn: ConnectionLike, author_id: int) -> QueryResults[int]:\n",
			wantConn:   commonConnExpr,
		},
		{
			name: "repeated mysql param appears once in the signature",
			drv:  newMysqlDriver("pymysql", false),
			query: model.Query{
				Cmd:      metadata.CmdExec,
				FuncName: "rename_author",
				Params: []model.QueryValue{
					{Name: "n", Type: model.PyType{Type: "str", SQLType: "text"}},
					{Name: "n", Type: model.PyType{Type: "str", SQLType: "text"}, Repeated: true},
				},
			},
			annotation: "None",
			want:       "def rename_author(conn: pymysql.Connection, *, n: str) -> None:\n",
			wantConn:   commonConnExpr,
		},
		{
			name: "sync driver has no async prefix",
			drv:  newSqliteDriver("sqlite3", false),
			query: model.Query{
				Cmd:      metadata.CmdOne,
				FuncName: "get_count",
			},
			annotation: "int | None",
			want:       "def get_count(conn: sqlite3.Connection) -> int | None:\n",
			wantConn:   commonConnExpr,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := commonTestConfig()
			cfg.EmitClasses = tc.emitClasses
			cfg.OmitKwargsLimit = tc.kwargsLimit
			w := writer.NewCodeWriter(cfg)
			conn := writeFuncSignature(w, tc.drv, cfg, 0, tc.query, tc.annotation)
			if conn != tc.wantConn {
				t.Errorf("writeFuncSignature() = %q, want %q", conn, tc.wantConn)
			}
			if got := w.String(); got != tc.want {
				t.Errorf("writeFuncSignature() wrote %q, want %q", got, tc.want)
			}
		})
	}
}

func TestExpandParams(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		query model.Query
		want  []string
	}{
		{
			name:  "no params",
			query: model.Query{},
			want:  []string{},
		},
		{
			name: "empty value skipped",
			query: model.Query{
				Params: []model.QueryValue{{}},
			},
			want: []string{},
		},
		{
			name: "plain params pass through",
			query: model.Query{
				Params: []model.QueryValue{
					{Name: "author_id", Type: model.PyType{Type: "int", SQLType: "bigint"}},
					{Name: "name", Type: model.PyType{Type: "str", SQLType: "text"}},
				},
			},
			want: []string{"author_id", "name"},
		},
		{
			name: "bundled table expands columns with conversion",
			query: model.Query{
				Params: []model.QueryValue{
					{
						EmitTable: true,
						Name:      "params",
						Type:      model.PyType{Type: "GetAuthorParams"},
						Table: &model.Table{
							Name: "GetAuthorParams",
							Columns: []model.Column{
								{Name: "id", DBName: "id", Type: model.PyType{Type: "int", SQLType: "bigint"}},
								{
									Name:   "addr",
									DBName: "addr",
									Type: model.PyType{
										Type:        "IPv4Address",
										SQLType:     "inet",
										IsOverride:  true,
										DefaultType: "str",
									},
								},
							},
						},
					},
				},
			},
			want: []string{"params.id", "str(params.addr)"},
		},
		{
			name: "emit table without table falls through",
			query: model.Query{
				Params: []model.QueryValue{
					{EmitTable: true, Name: "params", Type: model.PyType{Type: "GetAuthorParams"}},
				},
			},
			want: []string{"params"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := expandParams(tc.query); !slices.Equal(got, tc.want) {
				t.Errorf("expandParams() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestExpandParamsFlattenSlices(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		query model.Query
		want  []string
	}{
		{
			name: "slice params are star-unpacked between plain params",
			query: model.Query{
				Params: []model.QueryValue{
					{Name: "name", Type: model.PyType{Type: "str", SQLType: "text"}},
					{Name: "ids", Type: model.PyType{Type: "int", SQLType: "integer", IsList: true, SqlcSliceName: "ids"}},
					{Name: "note", Type: model.PyType{Type: "str", SQLType: "text", IsNullable: true}},
				},
			},
			want: []string{"name", "*ids", "note"},
		},
		{
			name: "converted slice unpacks the element-wise conversion",
			query: model.Query{
				Params: []model.QueryValue{
					{Name: "days", Type: model.PyType{
						Type:          "float",
						IsList:        true,
						IsOverride:    true,
						DefaultType:   "datetime.date",
						SqlcSliceName: "days",
					}},
				},
			},
			want: []string{"*[datetime.date(v) for v in days]"},
		},
		{
			name: "reused slice repeats the starred copy per marker occurrence",
			query: model.Query{
				SQL: "DELETE FROM t WHERE id IN (/*SLICE:ids*/?) OR ref_id IN (/*SLICE:ids*/?)",
				Params: []model.QueryValue{
					{Name: "ids", Type: model.PyType{Type: "int", SQLType: "integer", IsList: true, SqlcSliceName: "ids"}},
				},
			},
			want: []string{"*ids", "*ids"},
		},
		{
			// The parameter array puts the merged slice first, but the SQL
			// binds name between the two use sites: text order must win.
			name: "reused slice interleaves plain params in SQL text order",
			query: model.Query{
				SQL: "SELECT id FROM t WHERE id IN (/*SLICE:ids*/?) AND name = ? AND ref_id IN (/*SLICE:ids*/?)",
				Params: []model.QueryValue{
					{Name: "ids", Type: model.PyType{Type: "int", SQLType: "integer", IsList: true, SqlcSliceName: "ids"}},
					{Name: "name", Type: model.PyType{Type: "str", SQLType: "text"}},
				},
			},
			want: []string{"*ids", "name", "*ids"},
		},
		{
			name: "reused slice with unaccounted plain placeholder falls back to consecutive copies",
			query: model.Query{
				SQL: "SELECT id FROM t WHERE id IN (/*SLICE:ids*/?) AND name = ? AND ref_id IN (/*SLICE:ids*/?)",
				Params: []model.QueryValue{
					{Name: "ids", Type: model.PyType{Type: "int", SQLType: "integer", IsList: true, SqlcSliceName: "ids"}},
				},
			},
			want: []string{"*ids", "*ids"},
		},
		{
			name: "reused slice with unknown marker falls back to consecutive copies",
			query: model.Query{
				SQL: "SELECT id FROM t WHERE id IN (/*SLICE:ids*/?) OR a IN (/*SLICE:ids*/?) OR b IN (/*SLICE:other*/?)",
				Params: []model.QueryValue{
					{Name: "ids", Type: model.PyType{Type: "int", SQLType: "integer", IsList: true, SqlcSliceName: "ids"}},
				},
			},
			want: []string{"*ids", "*ids"},
		},
		{
			name: "reused slice with leftover plain param falls back to consecutive copies",
			query: model.Query{
				SQL: "SELECT id FROM t WHERE id IN (/*SLICE:ids*/?) OR ref_id IN (/*SLICE:ids*/?)",
				Params: []model.QueryValue{
					{Name: "ids", Type: model.PyType{Type: "int", SQLType: "integer", IsList: true, SqlcSliceName: "ids"}},
					{Name: "ghost", Type: model.PyType{Type: "str", SQLType: "text"}},
				},
			},
			want: []string{"*ids", "*ids", "ghost"},
		},
		{
			name: "bundled table field slice unpacks the attribute",
			query: model.Query{
				Params: []model.QueryValue{
					{
						EmitTable: true,
						Name:      "params",
						Type:      model.PyType{Type: "GetRowsParams"},
						Table: &model.Table{
							Name: "GetRowsParams",
							Columns: []model.Column{
								{Name: "name", DBName: "name", Type: model.PyType{Type: "str", SQLType: "text"}},
								{
									Name:   "ids",
									DBName: "id",
									Type:   model.PyType{Type: "int", SQLType: "integer", IsList: true, SqlcSliceName: "ids"},
								},
							},
						},
					},
				},
			},
			want: []string{"params.name", "*params.ids"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := expandParamsFlattenSlices(tc.query); !slices.Equal(got, tc.want) {
				t.Errorf("expandParamsFlattenSlices() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestSliceParams(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		query model.Query
		want  []sliceParam
	}{
		{
			name: "no slices",
			query: model.Query{
				Params: []model.QueryValue{
					{Name: "name", Type: model.PyType{Type: "str", SQLType: "text"}},
				},
			},
			want: nil,
		},
		{
			name: "empty value skipped",
			query: model.Query{
				Params: []model.QueryValue{{}},
			},
			want: nil,
		},
		{
			name: "plain and escaped slice params keep the raw marker name",
			query: model.Query{
				Params: []model.QueryValue{
					{Name: "for_", Type: model.PyType{Type: "int", SQLType: "integer", IsList: true, SqlcSliceName: "for"}},
					{Name: "names", Type: model.PyType{Type: "str", SQLType: "text", IsList: true, SqlcSliceName: "names"}},
				},
			},
			want: []sliceParam{{marker: "for", expr: "for_"}, {marker: "names", expr: "names"}},
		},
		{
			name: "bundled table fields contribute their slices",
			query: model.Query{
				Params: []model.QueryValue{
					{
						EmitTable: true,
						Name:      "params",
						Type:      model.PyType{Type: "GetRowsParams"},
						Table: &model.Table{
							Name: "GetRowsParams",
							Columns: []model.Column{
								{Name: "name", DBName: "name", Type: model.PyType{Type: "str", SQLType: "text"}},
								{
									Name:   "ids",
									DBName: "id",
									Type:   model.PyType{Type: "int", SQLType: "integer", IsList: true, SqlcSliceName: "ids"},
								},
							},
						},
					},
				},
			},
			want: []sliceParam{{marker: "ids", expr: "params.ids"}},
		},
		{
			name: "emit table without table falls through to the plain path",
			query: model.Query{
				Params: []model.QueryValue{
					{EmitTable: true, Name: "params", Type: model.PyType{Type: "GetRowsParams"}},
				},
			},
			want: nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := sliceParams(tc.query); !slices.Equal(got, tc.want) {
				t.Errorf("sliceParams() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestWriteQueryDocstring(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		convention  config.DocstringConvention
		emitClasses bool
		query       model.Query
		retType     string
		want        string
	}{
		{
			name:       "docstrings disabled writes nothing",
			convention: config.DocstringConventionNone,
			query: model.Query{
				Cmd:       metadata.CmdOne,
				QueryName: "GetAuthor",
				SQL:       "SELECT id FROM authors WHERE id = $1",
				Params: []model.QueryValue{
					{Name: "author_id", Type: model.PyType{Type: "int", SQLType: "bigint"}},
				},
			},
			retType: "models.Author",
			want:    "",
		},
		{
			name:       "repeated mysql param is documented once",
			convention: config.DocstringConventionGoogle,
			query: model.Query{
				Cmd:       metadata.CmdExec,
				QueryName: "RenameAuthor",
				SQL:       "UPDATE authors SET name = %s WHERE name = %s",
				Params: []model.QueryValue{
					{Name: "n", Type: model.PyType{Type: "str", SQLType: "text"}},
					{Name: "n", Type: model.PyType{Type: "str", SQLType: "text"}, Repeated: true},
				},
			},
			want: strings.Join([]string{
				"    \"\"\"Execute SQL query with `name: RenameAuthor :exec`.",
				"",
				"    ```sql",
				"    UPDATE authors SET name = %s WHERE name = %s",
				"    ```",
				"",
				"    Args:",
				"        conn:",
				"            Connection object of type `ConnectionLike` used to execute the query.",
				"        n: str.",
				"    \"\"\"",
				"",
			}, "\n"),
		},
		{
			name:       "google one with conn and sql",
			convention: config.DocstringConventionGoogle,
			query: model.Query{
				Cmd:       metadata.CmdOne,
				QueryName: "GetAuthor",
				SQL:       "SELECT id FROM authors WHERE id = $1",
				Params: []model.QueryValue{
					{Name: "author_id", Type: model.PyType{Type: "int", SQLType: "bigint"}},
				},
			},
			retType: "models.Author",
			want: strings.Join([]string{
				"    \"\"\"Fetch one from the db using the SQL query with `name: GetAuthor :one`.",
				"",
				"    ```sql",
				"    SELECT id FROM authors WHERE id = $1",
				"    ```",
				"",
				"    Args:",
				"        conn:",
				"            Connection object of type `ConnectionLike` used to execute the query.",
				"        author_id: int.",
				"",
				"    Returns:",
				"        Result of type `models.Author` fetched from the db. Will be `None` if not found.",
				"    \"\"\"",
				"",
			}, "\n"),
		},
		{
			name:        "google copyfrom in classes mode adds extra and skips conn",
			convention:  config.DocstringConventionGoogle,
			emitClasses: true,
			query: model.Query{
				Cmd:       metadata.CmdCopyFrom,
				QueryName: "CopyAuthors",
				SQL:       "COPY authors (name) FROM STDIN",
				Params: []model.QueryValue{
					{},
					{Name: "params", Type: model.PyType{Type: "CopyAuthorsParams", IsList: true}},
				},
			},
			retType: "int",
			want: strings.Join([]string{
				"    \"\"\"Execute COPY FROM query to insert rows into a table with `name: CopyAuthors :copyfrom` and return the number of affected rows.",
				"",
				"    Args:",
				"        params: collections.abc.Sequence[CopyAuthorsParams].",
				"            A list of params for rows that should be inserted.",
				"",
				"    Returns:",
				"        The number (`int`) of affected rows.",
				"    \"\"\"",
				"",
			}, "\n"),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := commonTestConfig()
			cfg.EmitDocstrings = tc.convention
			cfg.EmitClasses = tc.emitClasses
			w := writer.NewCodeWriter(cfg)
			writeQueryDocstring(w, newAsyncpgDriver(), cfg, tc.query, 1, tc.retType)
			if got := w.String(); got != tc.want {
				t.Errorf("writeQueryDocstring() wrote %q, want %q", got, tc.want)
			}
		})
	}
}

func TestConvertParamExpr(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		typ  model.PyType
		want string
	}{
		{
			name: "no override passes through",
			typ:  model.PyType{Type: "int", SQLType: "bigint"},
			want: "x",
		},
		{
			name: "override to typing.Any passes through",
			typ:  model.PyType{Type: "Unknown", IsOverride: true, DefaultType: types.Any},
			want: "x",
		},
		{
			name: "override scalar converts",
			typ:  model.PyType{Type: "IPv4Address", IsOverride: true, DefaultType: "str"},
			want: "str(x)",
		},
		{
			name: "override list converts element-wise",
			typ:  model.PyType{Type: "IPv4Address", IsOverride: true, DefaultType: "str", IsList: true},
			want: "[str(v) for v in x]",
		},
		{
			name: "override nullable guards against None",
			typ:  model.PyType{Type: "IPv4Address", IsOverride: true, DefaultType: "str", IsNullable: true},
			want: "str(x) if x is not None else None",
		},
		{
			name: "override nullable list guards comprehension",
			typ:  model.PyType{Type: "IPv4Address", IsOverride: true, DefaultType: "str", IsList: true, IsNullable: true},
			want: "[str(v) for v in x] if x is not None else None",
		},
		{
			name: "converter scalar calls to_db instead of the default type",
			typ: model.PyType{
				Type: "mymod.Money", IsOverride: true, DefaultType: "str", ConverterTo: paramConverterTo,
			},
			want: "myconv.to_db(x)",
		},
		{
			name: "converter list converts element-wise",
			typ: model.PyType{
				Type: "mymod.Money", IsOverride: true, DefaultType: "str", ConverterTo: paramConverterTo, IsList: true,
			},
			want: "[myconv.to_db(v) for v in x]",
		},
		{
			name: "converter nullable guards against None",
			typ: model.PyType{
				Type: "mymod.Money", IsOverride: true, DefaultType: "str", ConverterTo: paramConverterTo, IsNullable: true,
			},
			want: "myconv.to_db(x) if x is not None else None",
		},
		{
			// A converter makes an otherwise non-instantiable typing.Any default convertible.
			name: "converter overrides the typing.Any passthrough",
			typ: model.PyType{
				Type: "mymod.Money", IsOverride: true, DefaultType: types.Any, ConverterTo: paramConverterTo,
			},
			want: "myconv.to_db(x)",
		},
		{
			name: "converter nullable list with typing.Any default guards comprehension",
			typ: model.PyType{
				Type:        "mymod.Money",
				IsOverride:  true,
				DefaultType: types.Any,
				ConverterTo: paramConverterTo,
				IsList:      true,
				IsNullable:  true,
			},
			want: "[myconv.to_db(v) for v in x] if x is not None else None",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := convertParamExpr("x", tc.typ); got != tc.want {
				t.Errorf("convertParamExpr(%q, %+v) = %q, want %q", "x", tc.typ, got, tc.want)
			}
		})
	}
}

func TestWriteExecRowsReturn(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		speedups bool
		want     string
	}{
		{
			name: "default walrus chain",
			want: "    return int(n) if (p := r.split()) and (n := p[-1]).isdigit() else 0\n",
		},
		{
			name:     "speedups skips empty split guard",
			speedups: true,
			want:     "    return int(n) if (n := r.split()[-1]).isdigit() else 0\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := commonTestConfig()
			cfg.Speedups = tc.speedups
			w := writer.NewCodeWriter(cfg)
			writeExecRowsReturn(w, cfg, 1)
			if got := w.String(); got != tc.want {
				t.Errorf("writeExecRowsReturn() wrote %q, want %q", got, tc.want)
			}
		})
	}
}

func TestPsycopgParamEntries(t *testing.T) {
	t.Parallel()
	query := model.Query{
		Params: []model.QueryValue{
			{},
			{Name: "name", Type: model.PyType{Type: "str", SQLType: "text"}, Number: 2},
		},
	}
	want := []string{`"p2": name`}
	if got := psycopgParamEntries(query); !slices.Equal(got, want) {
		t.Errorf("psycopgParamEntries() = %q, want %q", got, want)
	}
}

func TestConvertParamExprWire(t *testing.T) {
	t.Parallel()
	overrideDate := model.PyType{
		Type: "float", SQLType: "date", IsOverride: true, DefaultType: "datetime.date",
	}
	converterDate := model.PyType{
		Type: "float", SQLType: "date", IsOverride: true, DefaultType: "datetime.date",
		ConverterTo: "conv.to_db", ConverterFrom: "conv.from_db",
	}
	unknownOverride := model.PyType{
		Type: "u.U", SQLType: "sometype", IsOverride: true, DefaultType: types.Any,
	}
	cases := []struct {
		name string
		expr string
		typ  model.PyType
		wire wireConvertFunc
		want string
	}{
		{name: "plain without wire", expr: "x", typ: model.PyType{SQLType: "text", Type: "str"}, want: "x"},
		{
			name: "wire without override",
			expr: "x",
			typ:  model.PyType{SQLType: "date", Type: "datetime.date"},
			wire: tursoWire,
			want: "x.isoformat()",
		},
		{
			name: "wire skips natively bindable types",
			expr: "x",
			typ:  model.PyType{SQLType: "bool", Type: "bool"},
			wire: tursoWire,
			want: "x",
		},
		{
			name: "wire composes onto override conversion",
			expr: "x",
			typ:  overrideDate,
			wire: tursoWire,
			want: "datetime.date(x).isoformat()",
		},
		{
			name: "wire composes onto converter to_db",
			expr: "x",
			typ:  converterDate,
			wire: tursoWire,
			want: "conv.to_db(x).isoformat()",
		},
		{
			name: "unknown override passes through even with wire",
			expr: "x",
			typ:  unknownOverride,
			wire: tursoWire,
			want: "x",
		},
		{
			// A convertible SQL type with a typing.Any default cannot occur
			// in real IR (a known type always maps to a real DefaultType),
			// but the passthrough contract must hold defensively: the wire
			// template assumes the default type and must not wrap.
			name: "any default skips the wire even for a convertible type",
			expr: "x",
			typ: model.PyType{
				Type: "u.U", SQLType: "date", IsOverride: true, DefaultType: types.Any,
			},
			wire: tursoWire,
			want: "x",
		},
		{
			name: "wire list converts element-wise",
			expr: "xs",
			typ:  model.PyType{SQLType: "date", Type: "datetime.date", IsList: true},
			wire: tursoWire,
			want: "[v.isoformat() for v in xs]",
		},
		{
			name: "wire nullable guards",
			expr: "x",
			typ:  model.PyType{SQLType: "blob", Type: "memoryview", IsNullable: true},
			wire: tursoWire,
			want: "bytes(x) if x is not None else None",
		},
		{
			name: "wire nullable list guards the comprehension",
			expr: "xs",
			typ:  model.PyType{SQLType: "decimal", Type: "decimal.Decimal", IsList: true, IsNullable: true},
			wire: tursoWire,
			want: "[str(v) for v in xs] if xs is not None else None",
		},
		{
			name: "override without wire keeps legacy shape",
			expr: "x",
			typ:  overrideDate,
			want: "datetime.date(x)",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := convertParamExprWire(tc.expr, tc.typ, tc.wire); got != tc.want {
				t.Errorf("convertParamExprWire(%q) = %q, want %q", tc.expr, got, tc.want)
			}
		})
	}
}

func TestExpandParamsFlattenSlicesWire(t *testing.T) {
	t.Parallel()
	query := model.Query{
		SQL: "SELECT 1 FROM t WHERE d = ? AND id IN (/*SLICE:ids*/?)",
		Params: []model.QueryValue{
			{Name: "d", Type: model.PyType{SQLType: "date", Type: "datetime.date"}},
			{Name: "ids", Type: model.PyType{SQLType: "integer", Type: "int", IsList: true, SqlcSliceName: "ids"}},
		},
	}
	got := expandParamsFlattenSlicesWire(query, tursoWire)
	want := []string{"d.isoformat()", "*ids"}
	if len(got) != len(want) {
		t.Fatalf("expandParamsFlattenSlicesWire() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("part[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestSlotMarkerHelpers(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		sliceName string
		sql       string
		style     placeholderStyle
		wantCount int
		wantText  string
	}{
		{
			name:      "question marker keeps sqlc's raw form",
			sliceName: "ids",
			sql:       "WHERE id IN (/*SLICE:ids*/?)",
			style:     questionStyle,
			wantCount: 1,
			wantText:  "/*SLICE:ids*/?",
		},
		{
			name:      "pyformat marker keeps the rewritten token",
			sliceName: "ids",
			sql:       "WHERE id IN (/*SLICE:ids*/%s)",
			style:     pyformatStyle,
			wantCount: 1,
			wantText:  "/*SLICE:ids*/%s",
		},
		{
			// The rewriter doubles a percent inside the marker, so the text
			// scanned back out has to be doubled too or the generated
			// replace misses.
			name:      "percent in a slice name keeps its doubled marker",
			sliceName: "a%b",
			sql:       "WHERE id IN (/*SLICE:a%%b*/%s)",
			style:     pyformatStyle,
			wantCount: 1,
			wantText:  "/*SLICE:a%%b*/%s",
		},
		{
			name:      "two markers count both",
			sliceName: "ids",
			sql:       "WHERE id IN (/*SLICE:ids*/%s) OR r IN (/*SLICE:ids*/%s)",
			style:     pyformatStyle,
			wantCount: 2,
			wantText:  "/*SLICE:ids*/%s",
		},
		{
			// A marker inside a string literal is not a bind slot, so it
			// neither counts nor supplies text.
			// A marker inside a string literal is not a bind slot, so the
			// text falls back to the reconstructed marker rather than the
			// empty string, which str.replace would prepend.
			name:      "marker inside a string literal is invisible",
			sliceName: "ids",
			sql:       "WHERE note = '/*SLICE:ids*/?'",
			style:     questionStyle,
			wantCount: 1,
			wantText:  "/*SLICE:ids*/?",
		},
		{
			name:      "missing marker clamps to one and rebuilds its text",
			sliceName: "ids",
			sql:       "WHERE id = ?",
			style:     questionStyle,
			wantCount: 1,
			wantText:  "/*SLICE:ids*/?",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			slots := querySlots(model.Query{SQL: tc.sql}, tc.style)
			if got := slotMarkerCount(slots, tc.sliceName); got != tc.wantCount {
				t.Errorf("slotMarkerCount(%q) = %d, want %d", tc.sliceName, got, tc.wantCount)
			}
			if got := slotMarkerText(slots, tc.sliceName, tc.style); got != tc.wantText {
				t.Errorf("slotMarkerText(%q) = %q, want %q", tc.sliceName, got, tc.wantText)
			}
		})
	}
}

// mysqlTestWire is a stand-in wire conversion for the pyformat expansion
// tests: blobs go over the wire as bytes, everything else binds natively.
func mysqlTestWire(sqlType string) (string, bool) {
	if sqlType == "blob" {
		return "bytes(%s)", true
	}

	return "", false
}

func TestExpandParamsPyformat(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		query model.Query
		want  []string
	}{
		{
			name: "wire conversion wraps the blob param",
			query: model.Query{
				Params: []model.QueryValue{
					{Name: "data", Type: model.PyType{Type: "memoryview", SQLType: "blob"}},
					{Name: "name", Type: model.PyType{Type: "str", SQLType: "text"}},
				},
			},
			want: []string{"bytes(data)", "name"},
		},
		{
			name: "slice param with pyformat marker is star-unpacked",
			query: model.Query{
				SQL: "SELECT id FROM t WHERE id IN (/*SLICE:ids*/%s)",
				Params: []model.QueryValue{
					{Name: "ids", Type: model.PyType{Type: "int", SQLType: "integer", IsList: true, SqlcSliceName: "ids"}},
				},
			},
			want: []string{"*ids"},
		},
		{
			name: "reused slice binds a copy per marker between plain params",
			query: model.Query{
				SQL: "SELECT id FROM t WHERE a = %s AND id IN (/*SLICE:ids*/%s) OR id IN (/*SLICE:ids*/%s) AND b = %s",
				Params: []model.QueryValue{
					{Name: "a", Type: model.PyType{Type: "int", SQLType: "integer"}},
					{Name: "ids", Type: model.PyType{Type: "int", SQLType: "integer", IsList: true, SqlcSliceName: "ids"}},
					{Name: "b", Type: model.PyType{Type: "str", SQLType: "text"}},
				},
			},
			want: []string{"a", "*ids", "*ids", "b"},
		},
		{
			// The parameter array puts the merged slice first, but the SQL
			// binds a before the two use sites: text order must win.
			name: "reused slice interleaves plain params in SQL text order",
			query: model.Query{
				SQL: "SELECT id FROM t WHERE a = %s AND id IN (/*SLICE:ids*/%s) OR ref_id IN (/*SLICE:ids*/%s)",
				Params: []model.QueryValue{
					{Name: "ids", Type: model.PyType{Type: "int", SQLType: "integer", IsList: true, SqlcSliceName: "ids"}},
					{Name: "a", Type: model.PyType{Type: "int", SQLType: "integer"}},
				},
			},
			want: []string{"a", "*ids", "*ids"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := expandParamsPyformat(tc.query, mysqlTestWire); !slices.Equal(got, tc.want) {
				t.Errorf("expandParamsPyformat() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFindTursoConversion(t *testing.T) {
	t.Parallel()
	cases := []struct {
		sqlType   string
		needs     bool
		decodeFmt string
		wireFmt   string
		hasWire   bool
	}{
		{sqlType: "date", needs: true, decodeFmt: "datetime.date.fromisoformat(%s)", wireFmt: "%s.isoformat()", hasWire: true},
		{
			sqlType:   "datetime",
			needs:     true,
			decodeFmt: "datetime.datetime.fromisoformat(%s)",
			wireFmt:   "%s.isoformat()",
			hasWire:   true,
		},
		{
			sqlType:   "timestamp",
			needs:     true,
			decodeFmt: "datetime.datetime.fromisoformat(%s)",
			wireFmt:   "%s.isoformat()",
			hasWire:   true,
		},
		{sqlType: "decimal", needs: true, decodeFmt: "decimal.Decimal(str(%s))", wireFmt: "str(%s)", hasWire: true},
		{sqlType: "decimal(10,5)", needs: true, decodeFmt: "decimal.Decimal(str(%s))", wireFmt: "str(%s)", hasWire: true},
		{sqlType: "bool", needs: true, decodeFmt: "bool(%s)", hasWire: false},
		{sqlType: "boolean", needs: true, decodeFmt: "bool(%s)", hasWire: false},
		{sqlType: "blob", needs: true, decodeFmt: "memoryview(%s)", wireFmt: "bytes(%s)", hasWire: true},
		{sqlType: "text", needs: false},
		{sqlType: "integer", needs: false},
	}
	for _, tc := range cases {
		t.Run(tc.sqlType, func(t *testing.T) {
			t.Parallel()
			if got := tursoNeedsConversion(tc.sqlType); got != tc.needs {
				t.Fatalf("tursoNeedsConversion(%q) = %v, want %v", tc.sqlType, got, tc.needs)
			}
			decodeFmt, ok := tursoDecodeExpr(tc.sqlType)
			if ok != tc.needs || decodeFmt != tc.decodeFmt {
				t.Errorf("tursoDecodeExpr(%q) = %q, %v, want %q, %v", tc.sqlType, decodeFmt, ok, tc.decodeFmt, tc.needs)
			}
			wireFmt, ok := tursoWire(tc.sqlType)
			if ok != tc.hasWire || wireFmt != tc.wireFmt {
				t.Errorf("tursoWire(%q) = %q, %v, want %q, %v", tc.sqlType, wireFmt, ok, tc.wireFmt, tc.hasWire)
			}
		})
	}
}

func TestRowBuilderDecodeExpr(t *testing.T) {
	t.Parallel()
	rb := newRowBuilderWithDecode(tursoNeedsConversion, tursoDecodeExpr)
	overrideDate := model.PyType{
		Type: "float", SQLType: "date", IsOverride: true, DefaultType: "datetime.date",
	}
	cases := []struct {
		name string
		typ  model.PyType
		want string
	}{
		{name: "no conversion", typ: model.PyType{SQLType: "text", Type: "str"}, want: "row[0]"},
		{
			name: "date decode",
			typ:  model.PyType{SQLType: "date", Type: "datetime.date"},
			want: "datetime.date.fromisoformat(row[0])",
		},
		{
			name: "nullable decimal decode",
			typ:  model.PyType{SQLType: "decimal(10,5)", Type: "decimal.Decimal", IsNullable: true},
			want: "decimal.Decimal(str(row[0])) if row[0] is not None else None",
		},
		{
			name: "list blob decode is element-wise",
			typ:  model.PyType{SQLType: "blob", Type: "memoryview", IsList: true},
			want: "[memoryview(v) for v in row[0]]",
		},
		{
			name: "override keeps the constructor call",
			typ:  overrideDate,
			want: "float(row[0])",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := rb.convertExpr(tc.typ, "row[0]"); got != tc.want {
				t.Errorf("convertExpr() = %q, want %q", got, tc.want)
			}
		})
	}

	// A decode hook that never matches falls back to the constructor call.
	fallback := newRowBuilderWithDecode(func(string) bool { return true }, func(string) (string, bool) { return "", false })
	if got := fallback.convertExpr(model.PyType{SQLType: "text", Type: "str"}, "row[0]"); got != "str(row[0])" {
		t.Errorf("fallback convertExpr() = %q, want %q", got, "str(row[0])")
	}
}

func TestTursoSpeedupsDecodeExpr(t *testing.T) {
	t.Parallel()
	cases := []struct {
		sqlType string
		want    string
		ok      bool
	}{
		{sqlType: "date", want: "ciso8601.parse_datetime(%s).date()", ok: true},
		{sqlType: "timestamp", want: "ciso8601.parse_datetime(%s)", ok: true},
		// No speedups variant: falls back to the plain decode.
		{sqlType: "decimal(10,5)", want: "decimal.Decimal(str(%s))", ok: true},
		{sqlType: "blob", want: "memoryview(%s)", ok: true},
		// Non-convertible types report false, mirroring tursoDecodeExpr.
		{sqlType: "text", want: "", ok: false},
	}
	for _, tc := range cases {
		t.Run(tc.sqlType, func(t *testing.T) {
			t.Parallel()
			got, ok := tursoSpeedupsDecodeExpr(tc.sqlType)
			if got != tc.want || ok != tc.ok {
				t.Errorf("tursoSpeedupsDecodeExpr(%q) = %q, %v, want %q, %v", tc.sqlType, got, ok, tc.want, tc.ok)
			}
		})
	}
}
