package driver_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/driver"
	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/utils"
	"github.com/rayakame/sqlc-gen-better-python/internal/writer"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func asyncpgTestConfig() *config.Config {
	return &config.Config{
		SqlDriver:           config.SQLDriverAsyncpg,
		EmitDocstrings:      config.DocstringConventionNone,
		EmitDocstringsSQL:   utils.ToPtr(true),
		IndentChar:          " ",
		CharsPerIndentLevel: 4,
		OmitKwargsLimit:     8,
	}
}

func newAsyncpg(t *testing.T) driver.Driver {
	t.Helper()
	d, err := driver.New(asyncpgTestConfig())
	if err != nil {
		t.Fatalf("driver.New() error = %v", err)
	}

	return d
}

func TestAsyncpgDriverMetadata(t *testing.T) {
	t.Parallel()
	d := newAsyncpg(t)
	if got := d.Name(); got != "asyncpg" {
		t.Errorf("Name() = %q, want %q", got, "asyncpg")
	}
	if got := d.ConnType(); got != "ConnectionLike" {
		t.Errorf("ConnType() = %q, want %q", got, "ConnectionLike")
	}
	if !d.IsAsync() {
		t.Error("IsAsync() = false, want true")
	}
	wantHook := []string{
		"type ConnectionLike = asyncpg.Connection[asyncpg.Record] | asyncpg.pool.PoolConnectionProxy[asyncpg.Record]",
	}
	if got := d.TypeCheckingHook(); !slices.Equal(got, wantHook) {
		t.Errorf("TypeCheckingHook() = %q, want %q", got, wantHook)
	}
	w := writer.NewCodeWriter(asyncpgTestConfig())
	if d.WriteConversionSetup(w, asyncpgTestConfig(), nil) {
		t.Error("WriteConversionSetup() = true, want false")
	}
	if got := w.String(); got != "" {
		t.Errorf("WriteConversionSetup() wrote %q, want nothing", got)
	}
}

func TestAsyncpgNeedsConversion(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		sqlType string
		want    bool
	}{
		{name: "bytea", sqlType: "bytea", want: true},
		{name: "blob", sqlType: "blob", want: true},
		{name: "pg_catalog bytea", sqlType: "pg_catalog.bytea", want: true},
		{name: "inet", sqlType: "inet", want: true},
		{name: "cidr", sqlType: "cidr", want: true},
		{name: "text needs none", sqlType: "text", want: false},
		{name: "empty needs none", sqlType: "", want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			d := newAsyncpg(t)
			if got := d.NeedsConversion(tc.sqlType); got != tc.want {
				t.Errorf("NeedsConversion(%q) = %v, want %v", tc.sqlType, got, tc.want)
			}
			// asyncpg converts everything inline, so both checks must agree.
			if got := d.ConvertsInline(tc.sqlType); got != tc.want {
				t.Errorf("ConvertsInline(%q) = %v, want %v", tc.sqlType, got, tc.want)
			}
		})
	}
}

func TestAsyncpgSupportsCommand(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		cmd  string
		want bool
	}{
		{name: "exec", cmd: metadata.CmdExec, want: true},
		{name: "execresult", cmd: metadata.CmdExecResult, want: true},
		{name: "execrows", cmd: metadata.CmdExecRows, want: true},
		{name: "one", cmd: metadata.CmdOne, want: true},
		{name: "many", cmd: metadata.CmdMany, want: true},
		{name: "copyfrom", cmd: metadata.CmdCopyFrom, want: true},
		{name: "execlastid unsupported", cmd: metadata.CmdExecLastId, want: false},
		{name: "batchexec unsupported", cmd: metadata.CmdBatchExec, want: false},
		{name: "batchmany unsupported", cmd: metadata.CmdBatchMany, want: false},
		{name: "batchone unsupported", cmd: metadata.CmdBatchOne, want: false},
		{name: "empty unsupported", cmd: "", want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			d := newAsyncpg(t)
			if got := d.SupportsCommand(tc.cmd); got != tc.want {
				t.Errorf("SupportsCommand(%q) = %v, want %v", tc.cmd, got, tc.want)
			}
		})
	}
}

func TestAsyncpgWriteQueryResultsClass(t *testing.T) {
	t.Parallel()
	d := newAsyncpg(t)
	w := writer.NewCodeWriter(asyncpgTestConfig())
	if got := d.WriteQueryResultsClass(w); got != "QueryResults" {
		t.Errorf("WriteQueryResultsClass() = %q, want %q", got, "QueryResults")
	}
	want := strings.Join([]string{
		"class QueryResults[T]:",
		`    __slots__ = ("_args", "_conn", "_cursor", "_decode_hook", "_iterator", "_sql")`,
		"",
		"    def __init__(",
		"        self,",
		"        conn: ConnectionLike,",
		"        sql: str,",
		"        decode_hook: collections.abc.Callable[[asyncpg.Record], T],",
		"        *args: QueryResultsArgsType,",
		"    ) -> None:",
		"        self._conn = conn",
		"        self._sql = sql",
		"        self._decode_hook = decode_hook",
		"        self._args = args",
		"        self._cursor: asyncpg.cursor.CursorFactory[asyncpg.Record] | None = None",
		"        self._iterator: asyncpg.cursor.CursorIterator[asyncpg.Record] | None = None",
		"",
		"    def __aiter__(self) -> QueryResults[T]:",
		"        return self",
		"",
		"    def __await__(",
		"        self,",
		"    ) -> collections.abc.Generator[None, None, collections.abc.Sequence[T]]:",
		"        async def _wrapper() -> collections.abc.Sequence[T]:",
		"            result = await self._conn.fetch(self._sql, *self._args)",
		"            return [self._decode_hook(row) for row in result]",
		"",
		"        return _wrapper().__await__()",
		"",
		"    async def __anext__(self) -> T:",
		"        if self._cursor is None or self._iterator is None:",
		"            self._cursor = self._conn.cursor(self._sql, *self._args)",
		"            self._iterator = self._cursor.__aiter__()",
		"        try:",
		"            record = await self._iterator.__anext__()",
		"        except StopAsyncIteration:",
		"            self._cursor = None",
		"            self._iterator = None",
		"            raise",
		"        return self._decode_hook(record)",
		"",
	}, "\n")
	if got := w.String(); got != want {
		t.Errorf("WriteQueryResultsClass() wrote %q, want %q", got, want)
	}
}

func TestAsyncpgWriteQueryFunc(t *testing.T) {
	t.Parallel()
	authorTable := &model.Table{
		Name: "Author",
		Columns: []model.Column{
			{Name: "id", DBName: "id", Type: model.PyType{Type: "int", SQLType: "bigint"}},
			{Name: "name", DBName: "name", Type: model.PyType{Type: "str", SQLType: "text"}},
		},
	}
	cases := []struct {
		name  string
		query model.Query
		want  string
	}{
		{
			name: "exec without params",
			query: model.Query{
				Cmd:          metadata.CmdExec,
				ConstantName: "TRUNCATE_AUTHORS",
				FuncName:     "truncate_authors",
				QueryName:    "TruncateAuthors",
				Returns:      model.QueryValue{Type: model.PyType{Type: "None"}},
			},
			want: strings.Join([]string{
				"async def truncate_authors(conn: ConnectionLike) -> None:",
				"    await conn.execute(TRUNCATE_AUTHORS)",
				"",
			}, "\n"),
		},
		{
			name: "exec with overridden param",
			query: model.Query{
				Cmd:          metadata.CmdExec,
				ConstantName: "TOUCH_HOST",
				FuncName:     "touch_host",
				QueryName:    "TouchHost",
				Params: []model.QueryValue{
					{
						Name: "addr",
						Type: model.PyType{Type: "IPv4Address", SQLType: "inet", IsOverride: true, DefaultType: "str"},
					},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "None"}},
			},
			want: strings.Join([]string{
				"async def touch_host(conn: ConnectionLike, addr: IPv4Address) -> None:",
				"    await conn.execute(TOUCH_HOST, str(addr))",
				"",
			}, "\n"),
		},
		{
			name: "execresult returns status string",
			query: model.Query{
				Cmd:          metadata.CmdExecResult,
				ConstantName: "UPDATE_AUTHOR_NAME",
				FuncName:     "update_author_name",
				QueryName:    "UpdateAuthorName",
				Params: []model.QueryValue{
					{Name: "author_id", Type: model.PyType{Type: "int", SQLType: "bigint"}},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "None"}},
			},
			want: strings.Join([]string{
				"async def update_author_name(conn: ConnectionLike, author_id: int) -> str:",
				"    return await conn.execute(UPDATE_AUTHOR_NAME, author_id)",
				"",
			}, "\n"),
		},
		{
			name: "execrows parses status string",
			query: model.Query{
				Cmd:          metadata.CmdExecRows,
				ConstantName: "DELETE_AUTHORS",
				FuncName:     "delete_authors",
				QueryName:    "DeleteAuthors",
				Returns:      model.QueryValue{Type: model.PyType{Type: "int"}},
			},
			want: strings.Join([]string{
				"async def delete_authors(conn: ConnectionLike) -> int:",
				"    r = await conn.execute(DELETE_AUTHORS)",
				"    return int(n) if (p := r.split()) and (n := p[-1]).isdigit() else 0",
				"",
			}, "\n"),
		},
		{
			name: "one with struct return",
			query: model.Query{
				Cmd:          metadata.CmdOne,
				ConstantName: "GET_AUTHOR",
				FuncName:     "get_author",
				QueryName:    "GetAuthor",
				Params: []model.QueryValue{
					{Name: "author_id", Type: model.PyType{Type: "int", SQLType: "bigint"}},
				},
				Returns: model.QueryValue{Table: authorTable, Type: model.PyType{Type: "models.Author"}},
			},
			want: strings.Join([]string{
				"async def get_author(conn: ConnectionLike, author_id: int) -> models.Author | None:",
				"    row = await conn.fetchrow(GET_AUTHOR, author_id)",
				"    if row is None:",
				"        return None",
				"    return models.Author(id=row[0], name=row[1])",
				"",
			}, "\n"),
		},
		{
			name: "one with scalar return",
			query: model.Query{
				Cmd:          metadata.CmdOne,
				ConstantName: "GET_AUTHOR_ID",
				FuncName:     "get_author_id",
				QueryName:    "GetAuthorID",
				Params: []model.QueryValue{
					{Name: "name", Type: model.PyType{Type: "str", SQLType: "text"}},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "int", SQLType: "bigint"}},
			},
			want: strings.Join([]string{
				"async def get_author_id(conn: ConnectionLike, name: str) -> int | None:",
				"    row = await conn.fetchrow(GET_AUTHOR_ID, name)",
				"    if row is None:",
				"        return None",
				"    return row[0]",
				"",
			}, "\n"),
		},
		{
			name: "many scalar uses itemgetter",
			query: model.Query{
				Cmd:          metadata.CmdMany,
				ConstantName: "LIST_AUTHOR_IDS",
				FuncName:     "list_author_ids",
				QueryName:    "ListAuthorIDs",
				Returns:      model.QueryValue{Type: model.PyType{Type: "int", SQLType: "bigint"}},
			},
			want: strings.Join([]string{
				"def list_author_ids(conn: ConnectionLike) -> QueryResults[int]:",
				"    return QueryResults(conn, LIST_AUTHOR_IDS, operator.itemgetter(0))",
				"",
			}, "\n"),
		},
		{
			name: "many struct uses decode hook",
			query: model.Query{
				Cmd:          metadata.CmdMany,
				ConstantName: "LIST_AUTHORS_BY_NAME",
				FuncName:     "list_authors_by_name",
				QueryName:    "ListAuthorsByName",
				Params: []model.QueryValue{
					{Name: "name", Type: model.PyType{Type: "str", SQLType: "text"}},
				},
				Returns: model.QueryValue{Table: authorTable, Type: model.PyType{Type: "models.Author"}},
			},
			want: strings.Join([]string{
				"def list_authors_by_name(conn: ConnectionLike, name: str) -> QueryResults[models.Author]:",
				"    def _decode_hook(row: asyncpg.Record) -> models.Author:",
				"        return models.Author(id=row[0], name=row[1])",
				"",
				"    return QueryResults(conn, LIST_AUTHORS_BY_NAME, _decode_hook, name)",
				"",
			}, "\n"),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			d := newAsyncpg(t)
			cfg := asyncpgTestConfig()
			w := writer.NewCodeWriter(cfg)
			d.WriteQueryFunc(w, cfg, tc.query, 0)
			if got := w.String(); got != tc.want {
				t.Errorf("WriteQueryFunc() wrote %q, want %q", got, tc.want)
			}
		})
	}
}

func TestAsyncpgWriteQueryFuncCopyFrom(t *testing.T) {
	t.Parallel()
	// Column names sized against writer.MaxLineLength (320): 290 overflows the
	// single-line comprehension but keeps the tuple on one line, 2x200 forces
	// the fully exploded records list and columns list.
	wideCol := strings.Repeat("a", 290)
	widerColA := strings.Repeat("a", 200)
	widerColB := strings.Repeat("b", 200)
	cases := []struct {
		name  string
		query model.Query
		want  string
	}{
		{
			name: "single line records and copy call",
			query: model.Query{
				Cmd:          metadata.CmdCopyFrom,
				ConstantName: "COPY_AUTHORS",
				FuncName:     "copy_authors",
				QueryName:    "CopyAuthors",
				Params: []model.QueryValue{
					{
						EmitTable: true,
						Name:      "params",
						Type:      model.PyType{Type: "CopyAuthorsParams", IsList: true},
						Table: &model.Table{
							Name: "CopyAuthorsParams",
							Columns: []model.Column{
								{Name: "id", DBName: "id", Type: model.PyType{Type: "int", SQLType: "bigint"}},
								{Name: "name", DBName: "name", Type: model.PyType{Type: "str", SQLType: "text"}},
							},
						},
					},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "int"}},
				Table:   &plugin.Identifier{Name: "authors"},
			},
			want: strings.Join([]string{
				"async def copy_authors(conn: ConnectionLike, params: collections.abc.Sequence[CopyAuthorsParams]) -> int:",
				"    records = [(param.id, param.name) for param in params]",
				`    r = await conn.copy_records_to_table("authors", columns=["id", "name"], records=records)`,
				"    return int(n) if (p := r.split()) and (n := p[-1]).isdigit() else 0",
				"",
			}, "\n"),
		},
		{
			name: "single overridden column with schema",
			query: model.Query{
				Cmd:          metadata.CmdCopyFrom,
				ConstantName: "COPY_HOSTS",
				FuncName:     "copy_hosts",
				QueryName:    "CopyHosts",
				Params: []model.QueryValue{
					{
						EmitTable: true,
						Name:      "params",
						Type:      model.PyType{Type: "CopyHostsParams", IsList: true},
						Table: &model.Table{
							Name: "CopyHostsParams",
							Columns: []model.Column{
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
				Returns: model.QueryValue{Type: model.PyType{Type: "int"}},
				Table:   &plugin.Identifier{Schema: "app", Name: "hosts"},
			},
			want: strings.Join([]string{
				"async def copy_hosts(conn: ConnectionLike, params: collections.abc.Sequence[CopyHostsParams]) -> int:",
				"    records = [(str(param.addr),) for param in params]",
				`    r = await conn.copy_records_to_table("hosts", columns=["addr"], records=records, schema_name="app")`,
				"    return int(n) if (p := r.split()) and (n := p[-1]).isdigit() else 0",
				"",
			}, "\n"),
		},
		{
			name: "wide column explodes copy call but keeps tuple",
			query: model.Query{
				Cmd:          metadata.CmdCopyFrom,
				ConstantName: "COPY_WIDE",
				FuncName:     "copy_wide",
				QueryName:    "CopyWide",
				Params: []model.QueryValue{
					{
						EmitTable: true,
						Name:      "params",
						Type:      model.PyType{Type: "CopyWideParams", IsList: true},
						Table: &model.Table{
							Name: "CopyWideParams",
							Columns: []model.Column{
								{Name: wideCol, DBName: wideCol, Type: model.PyType{Type: "str", SQLType: "text"}},
							},
						},
					},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "int"}},
				Table:   &plugin.Identifier{Name: "wide"},
			},
			want: strings.Join([]string{
				"async def copy_wide(conn: ConnectionLike, params: collections.abc.Sequence[CopyWideParams]) -> int:",
				"    records = [",
				"        (param." + wideCol + ",)",
				"        for param in params",
				"    ]",
				"    r = await conn.copy_records_to_table(",
				`        "wide",`,
				`        columns=["` + wideCol + `"],`,
				"        records=records,",
				"    )",
				"    return int(n) if (p := r.split()) and (n := p[-1]).isdigit() else 0",
				"",
			}, "\n"),
		},
		{
			name: "wider columns explode records and columns list",
			query: model.Query{
				Cmd:          metadata.CmdCopyFrom,
				ConstantName: "COPY_WIDER",
				FuncName:     "copy_wider",
				QueryName:    "CopyWider",
				Params: []model.QueryValue{
					{
						EmitTable: true,
						Name:      "params",
						Type:      model.PyType{Type: "CopyWiderParams", IsList: true},
						Table: &model.Table{
							Name: "CopyWiderParams",
							Columns: []model.Column{
								{Name: widerColA, DBName: widerColA, Type: model.PyType{Type: "str", SQLType: "text"}},
								{Name: widerColB, DBName: widerColB, Type: model.PyType{Type: "str", SQLType: "text"}},
							},
						},
					},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "int"}},
				Table:   &plugin.Identifier{Name: "wider"},
			},
			want: strings.Join([]string{
				"async def copy_wider(conn: ConnectionLike, params: collections.abc.Sequence[CopyWiderParams]) -> int:",
				"    records = [",
				"        (",
				"            param." + widerColA + ",",
				"            param." + widerColB + ",",
				"        )",
				"        for param in params",
				"    ]",
				"    r = await conn.copy_records_to_table(",
				`        "wider",`,
				"        columns=[",
				`            "` + widerColA + `",`,
				`            "` + widerColB + `",`,
				"        ],",
				"        records=records,",
				"    )",
				"    return int(n) if (p := r.split()) and (n := p[-1]).isdigit() else 0",
				"",
			}, "\n"),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			d := newAsyncpg(t)
			cfg := asyncpgTestConfig()
			w := writer.NewCodeWriter(cfg)
			d.WriteQueryFunc(w, cfg, tc.query, 0)
			if got := w.String(); got != tc.want {
				t.Errorf("WriteQueryFunc() wrote %q, want %q", got, tc.want)
			}
		})
	}
}
