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
