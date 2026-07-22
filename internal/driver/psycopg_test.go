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

func psycopgTestConfig() *config.Config {
	return &config.Config{
		SqlDriver:           config.SQLDriverPsycopgAsync,
		EmitDocstrings:      config.DocstringConventionNone,
		EmitDocstringsSQL:   utils.ToPtr(true),
		IndentChar:          " ",
		CharsPerIndentLevel: 4,
		OmitKwargsLimit:     8,
	}
}

func newPsycopg(t *testing.T) driver.Driver {
	t.Helper()
	d, err := driver.New(psycopgTestConfig())
	if err != nil {
		t.Fatalf("driver.New() error = %v", err)
	}

	return d
}

func psycopgAuthorReturn() model.QueryValue {
	return model.QueryValue{
		Table: &model.Table{
			Name: "Author",
			Columns: []model.Column{
				{Name: "id", Type: model.PyType{Type: "int", SQLType: "bigint"}},
				{Name: "name", Type: model.PyType{Type: "str", SQLType: "text"}},
			},
		},
		Type: model.PyType{Type: "models.Author"},
	}
}

func TestPsycopgDriverMetadata(t *testing.T) {
	t.Parallel()
	d := newPsycopg(t)
	if got := d.Name(); got != "psycopg" {
		t.Errorf("Name() = %q, want %q", got, "psycopg")
	}
	if got := d.ConnType(); got != "ConnectionLike" {
		t.Errorf("ConnType() = %q, want %q", got, "ConnectionLike")
	}
	if !d.IsAsync() {
		t.Error("IsAsync() = false, want true")
	}
	wantHook := []string{"type ConnectionLike = psycopg.AsyncConnection[psycopg.rows.TupleRow]"}
	if got := d.TypeCheckingHook(); !slices.Equal(got, wantHook) {
		t.Errorf("TypeCheckingHook() = %q, want %q", got, wantHook)
	}
}

func TestPsycopgConversions(t *testing.T) {
	t.Parallel()
	d := newPsycopg(t)
	// Identical to asyncpg: bytea/inet/cidr convert inline, json does not -
	// registered loaders keep it a str before decode code ever sees it.
	for sqlType, want := range map[string]bool{
		"bytea": true, "inet": true, "cidr": true,
		"json": false, "jsonb": false, "text": false,
	} {
		if got := d.NeedsConversion(sqlType); got != want {
			t.Errorf("NeedsConversion(%q) = %v, want %v", sqlType, got, want)
		}
		if got := d.ConvertsInline(sqlType); got != want {
			t.Errorf("ConvertsInline(%q) = %v, want %v", sqlType, got, want)
		}
	}
}

func TestPsycopgSupportsCommand(t *testing.T) {
	t.Parallel()
	d := newPsycopg(t)
	for cmd, want := range map[string]bool{
		metadata.CmdExec:       true,
		metadata.CmdExecResult: true,
		metadata.CmdExecRows:   true,
		metadata.CmdOne:        true,
		metadata.CmdMany:       true,
		metadata.CmdCopyFrom:   true,
		metadata.CmdExecLastId: false,
		":batchexec":           false,
	} {
		if got := d.SupportsCommand(cmd); got != want {
			t.Errorf("SupportsCommand(%q) = %v, want %v", cmd, got, want)
		}
	}
}

func TestPsycopgJSONTypesReturned(t *testing.T) {
	t.Parallel()
	jsonCol := model.Column{Name: "meta", Type: model.PyType{Type: "str", SQLType: "json"}}
	jsonbType := model.PyType{Type: "str", SQLType: "jsonb"}
	cases := []struct {
		name    string
		queries []model.Query
		want    []string
	}{
		{
			name:    "no json returns",
			queries: []model.Query{{Returns: model.QueryValue{Type: model.PyType{Type: "int", SQLType: "bigint"}}}},
			want:    []string{},
		},
		{
			name:    "scalar jsonb return",
			queries: []model.Query{{Returns: model.QueryValue{Type: jsonbType}}},
			want:    []string{"jsonb"},
		},
		{
			name: "struct with json and pg_catalog.json plus embed",
			queries: []model.Query{{
				Returns: model.QueryValue{
					Table: &model.Table{Name: "Row", Columns: []model.Column{
						jsonCol,
						{Name: "legacy", Type: model.PyType{Type: "str", SQLType: "pg_catalog.json"}},
						{Name: "author", Embed: &model.Embed{Columns: []model.Column{
							{Name: "prefs", Type: jsonbType},
						}}},
					}},
					Type: model.PyType{Type: "models.Row"},
				},
			}},
			want: []string{"json", "jsonb"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := driver.PsycopgJSONTypesReturned(tc.queries)
			if len(got) == 0 && len(tc.want) == 0 {
				return
			}
			if !slices.Equal(got, tc.want) {
				t.Errorf("PsycopgJSONTypesReturned() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestPsycopgWriteConversionSetup(t *testing.T) {
	t.Parallel()
	d := newPsycopg(t)
	conf := psycopgTestConfig()

	w := writer.NewCodeWriter(conf)
	queries := []model.Query{{Returns: model.QueryValue{Type: model.PyType{Type: "str", SQLType: "jsonb"}}}}
	if !d.WriteConversionSetup(w, conf, queries) {
		t.Fatal("WriteConversionSetup() = false, want true for a jsonb return")
	}
	want := "psycopg.adapters.register_loader(\"jsonb\", psycopg.types.string.TextLoader)\n"
	if got := w.String(); got != want {
		t.Errorf("WriteConversionSetup() wrote %q, want %q", got, want)
	}

	w = writer.NewCodeWriter(conf)
	queries = []model.Query{{Returns: model.QueryValue{Type: model.PyType{Type: "int", SQLType: "bigint"}}}}
	if d.WriteConversionSetup(w, conf, queries) {
		t.Error("WriteConversionSetup() = true, want false without json returns")
	}
	if got := w.String(); got != "" {
		t.Errorf("WriteConversionSetup() wrote %q, want nothing", got)
	}
}

func TestPsycopgWriteQueryResultsClass(t *testing.T) {
	t.Parallel()
	d := newPsycopg(t)
	w := writer.NewCodeWriter(psycopgTestConfig())
	if got := d.WriteQueryResultsClass(w); got != "QueryResults" {
		t.Errorf("WriteQueryResultsClass() = %q, want %q", got, "QueryResults")
	}
	want := strings.Join([]string{
		"class QueryResults[T]:",
		`    __slots__ = ("_conn", "_cursor", "_decode_hook", "_iterator", "_params", "_sql")`,
		"",
		"    def __init__(",
		"        self,",
		"        conn: ConnectionLike,",
		"        sql: typing.LiteralString,",
		"        decode_hook: collections.abc.Callable[[psycopg.rows.TupleRow], T],",
		"        params: dict[str, QueryResultsArgsType] | None = None,",
		"    ) -> None:",
		"        self._conn = conn",
		"        self._sql: typing.LiteralString = sql",
		"        self._decode_hook = decode_hook",
		"        self._params = params",
		"        self._cursor: psycopg.AsyncCursor[psycopg.rows.TupleRow] | None = None",
		"        self._iterator: collections.abc.AsyncIterator[psycopg.rows.TupleRow] | None = None",
		"",
		"    def __aiter__(self) -> QueryResults[T]:",
		"        return self",
		"",
		"    def __await__(",
		"        self,",
		"    ) -> collections.abc.Generator[None, None, collections.abc.Sequence[T]]:",
		"        async def _wrapper() -> collections.abc.Sequence[T]:",
		"            result = await (await self._conn.execute(self._sql, self._params)).fetchall()",
		"            return [self._decode_hook(row) for row in result]",
		"",
		"        return _wrapper().__await__()",
		"",
		"    async def __anext__(self) -> T:",
		"        if self._cursor is None or self._iterator is None:",
		"            self._cursor = await self._conn.execute(self._sql, self._params)",
		"            self._iterator = self._cursor.__aiter__()",
		"        try:",
		"            record = await self._iterator.__anext__()",
		"        except StopAsyncIteration:",
		"            self._cursor = None",
		"            self._iterator = None",
		"            raise",
		"        return self._decode_hook(record)",
	}, "\n") + "\n"
	if got := w.String(); got != want {
		t.Errorf("WriteQueryResultsClass() wrote %q, want %q", got, want)
	}
}

func TestPsycopgWriteQueryFunc(t *testing.T) {
	t.Parallel()
	longName := strings.Repeat("p", 340)
	cases := []struct {
		name  string
		query model.Query
		want  string
	}{
		{
			name: "exec without params",
			query: model.Query{
				Cmd:          metadata.CmdExec,
				ConstantName: "DELETE_ALL",
				FuncName:     "delete_all",
				Returns:      model.QueryValue{Type: model.PyType{Type: "None"}},
			},
			want: strings.Join([]string{
				"async def delete_all(conn: ConnectionLike) -> None:",
				"    await conn.execute(DELETE_ALL)",
				"",
			}, "\n"),
		},
		{
			name: "exec with params binds by number",
			query: model.Query{
				Cmd:          metadata.CmdExec,
				ConstantName: "DELETE_AUTHOR",
				FuncName:     "delete_author",
				Params: []model.QueryValue{
					{Name: "author_id", Type: model.PyType{Type: "int", SQLType: "bigint"}, Number: 1},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "None"}},
			},
			want: strings.Join([]string{
				"async def delete_author(conn: ConnectionLike, author_id: int) -> None:",
				`    await conn.execute(DELETE_AUTHOR, {"p1": author_id})`,
				"",
			}, "\n"),
		},
		{
			name: "exec with overridden param converts before binding",
			query: model.Query{
				Cmd:          metadata.CmdExec,
				ConstantName: "SET_CREATED_AT",
				FuncName:     "set_created_at",
				Params: []model.QueryValue{
					{Name: "created_at", Type: model.PyType{
						Type:        "float",
						SQLType:     "timestamp",
						IsOverride:  true,
						DefaultType: "datetime.datetime",
					}, Number: 1},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "None"}},
			},
			want: strings.Join([]string{
				"async def set_created_at(conn: ConnectionLike, created_at: float) -> None:",
				`    await conn.execute(SET_CREATED_AT, {"p1": datetime.datetime(created_at)})`,
				"",
			}, "\n"),
		},
		{
			name: "exec long params hoisted into sql_params",
			query: model.Query{
				Cmd:          metadata.CmdExec,
				ConstantName: "HOIST",
				FuncName:     "hoist",
				Params: []model.QueryValue{
					{Name: longName, Type: model.PyType{Type: "int", SQLType: "bigint"}, Number: 1},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "None"}},
			},
			want: strings.Join([]string{
				"async def hoist(",
				"    conn: ConnectionLike,",
				"    " + longName + ": int,",
				") -> None:",
				"    sql_params: dict[str, QueryResultsArgsType] = {",
				`        "p1": ` + longName + ",",
				"    }",
				"    await conn.execute(HOIST, sql_params)",
				"",
			}, "\n"),
		},
		{
			name: "execresult returns the cursor",
			query: model.Query{
				Cmd:          metadata.CmdExecResult,
				ConstantName: "UPDATE_ROWS",
				FuncName:     "update_rows",
				Params: []model.QueryValue{
					{Name: "id_", Type: model.PyType{Type: "int", SQLType: "bigint"}, Number: 1},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "None"}},
			},
			want: strings.Join([]string{
				"async def update_rows(conn: ConnectionLike, id_: int) -> psycopg.AsyncCursor[psycopg.rows.TupleRow]:",
				`    return await conn.execute(UPDATE_ROWS, {"p1": id_})`,
				"",
			}, "\n"),
		},
		{
			name: "execrows returns rowcount",
			query: model.Query{
				Cmd:          metadata.CmdExecRows,
				ConstantName: "UPDATE_ROWS",
				FuncName:     "update_rows",
				Params: []model.QueryValue{
					{Name: "id_", Type: model.PyType{Type: "int", SQLType: "bigint"}, Number: 1},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "int"}},
			},
			want: strings.Join([]string{
				"async def update_rows(conn: ConnectionLike, id_: int) -> int:",
				`    cur = await conn.execute(UPDATE_ROWS, {"p1": id_})`,
				"    return cur.rowcount",
				"",
			}, "\n"),
		},
		{
			name: "one struct return",
			query: model.Query{
				Cmd:          metadata.CmdOne,
				ConstantName: "GET_AUTHOR",
				FuncName:     "get_author",
				Params: []model.QueryValue{
					{Name: "id_", Type: model.PyType{Type: "int", SQLType: "bigint"}, Number: 1},
				},
				Returns: psycopgAuthorReturn(),
			},
			want: strings.Join([]string{
				"async def get_author(conn: ConnectionLike, id_: int) -> models.Author | None:",
				`    row = await (await conn.execute(GET_AUTHOR, {"p1": id_})).fetchone()`,
				"    if row is None:",
				"        return None",
				"    return models.Author(id=row[0], name=row[1])",
				"",
			}, "\n"),
		},
		{
			name: "one scalar return",
			query: model.Query{
				Cmd:          metadata.CmdOne,
				ConstantName: "COUNT_AUTHORS",
				FuncName:     "count_authors",
				Returns:      model.QueryValue{Type: model.PyType{Type: "int", SQLType: "bigint"}},
			},
			want: strings.Join([]string{
				"async def count_authors(conn: ConnectionLike) -> int | None:",
				"    row = await (await conn.execute(COUNT_AUTHORS)).fetchone()",
				"    if row is None:",
				"        return None",
				"    return row[0]",
				"",
			}, "\n"),
		},
		{
			name: "many struct return with decode hook",
			query: model.Query{
				Cmd:          metadata.CmdMany,
				ConstantName: "LIST_AUTHORS",
				FuncName:     "list_authors",
				Params: []model.QueryValue{
					{Name: "name", Type: model.PyType{Type: "str", SQLType: "text"}, Number: 1},
				},
				Returns: psycopgAuthorReturn(),
			},
			want: strings.Join([]string{
				"def list_authors(conn: ConnectionLike, name: str) -> QueryResults[models.Author]:",
				"    def _decode_hook(row: psycopg.rows.TupleRow) -> models.Author:",
				"        return models.Author(id=row[0], name=row[1])",
				"",
				`    return QueryResults(conn, LIST_AUTHORS, _decode_hook, {"p1": name})`,
				"",
			}, "\n"),
		},
		{
			name: "many scalar without params uses itemgetter",
			query: model.Query{
				Cmd:          metadata.CmdMany,
				ConstantName: "LIST_IDS",
				FuncName:     "list_ids",
				Returns:      model.QueryValue{Type: model.PyType{Type: "int", SQLType: "bigint"}},
			},
			want: strings.Join([]string{
				"def list_ids(conn: ConnectionLike) -> QueryResults[int]:",
				"    return QueryResults(conn, LIST_IDS, operator.itemgetter(0))",
				"",
			}, "\n"),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			d := newPsycopg(t)
			conf := psycopgTestConfig()
			body := writer.NewCodeWriter(conf)
			d.WriteQueryFunc(body, conf, tc.query, 0)
			if got := body.String(); got != tc.want {
				t.Errorf("WriteQueryFunc() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestPsycopgWriteQueryFuncCopyFrom(t *testing.T) {
	t.Parallel()
	longName := strings.Repeat("c", 340)
	cases := []struct {
		name  string
		query model.Query
		want  string
	}{
		{
			name: "copyfrom streams rows and returns rowcount",
			query: model.Query{
				Cmd:          metadata.CmdCopyFrom,
				ConstantName: "COPY_AUTHORS",
				FuncName:     "copy_authors",
				Table:        &plugin.Identifier{Name: "authors"},
				Params: []model.QueryValue{{
					EmitTable: true,
					Name:      "params",
					Type:      model.PyType{Type: "CopyAuthorsParams", IsList: true},
					Table: &model.Table{
						Name: "CopyAuthorsParams",
						Columns: []model.Column{
							{Name: "id_", DBName: "id", Type: model.PyType{Type: "int", SQLType: "bigint"}, Number: 1},
							{Name: "name", DBName: "name", Type: model.PyType{Type: "str", SQLType: "text"}, Number: 2},
						},
					},
				}},
				Returns: model.QueryValue{Type: model.PyType{Type: "int"}},
			},
			want: strings.Join([]string{
				"async def copy_authors(conn: ConnectionLike, params: collections.abc.Sequence[CopyAuthorsParams]) -> int:",
				"    async with conn.cursor() as cur:",
				`        async with cur.copy('COPY "authors" ("id", "name") FROM STDIN') as copy:`,
				"            for param in params:",
				"                await copy.write_row((param.id_, param.name))",
				"        return cur.rowcount",
				"",
			}, "\n"),
		},
		{
			name: "copyfrom single overridden column in a schema",
			query: model.Query{
				Cmd:          metadata.CmdCopyFrom,
				ConstantName: "COPY_PRICES",
				FuncName:     "copy_prices",
				Table:        &plugin.Identifier{Schema: "billing", Name: "prices"},
				Params: []model.QueryValue{{
					EmitTable: true,
					Name:      "params",
					Type:      model.PyType{Type: "CopyPricesParams", IsList: true},
					Table: &model.Table{
						Name: "CopyPricesParams",
						Columns: []model.Column{
							{Name: "amount", DBName: "amount", Type: model.PyType{
								Type:        "float",
								SQLType:     "numeric",
								IsOverride:  true,
								DefaultType: "decimal.Decimal",
							}, Number: 1},
						},
					},
				}},
				Returns: model.QueryValue{Type: model.PyType{Type: "int"}},
			},
			want: strings.Join([]string{
				"async def copy_prices(conn: ConnectionLike, params: collections.abc.Sequence[CopyPricesParams]) -> int:",
				"    async with conn.cursor() as cur:",
				`        async with cur.copy('COPY "billing"."prices" ("amount") FROM STDIN') as copy:`,
				"            for param in params:",
				"                await copy.write_row((decimal.Decimal(param.amount),))",
				"        return cur.rowcount",
				"",
			}, "\n"),
		},
		{
			name: "copyfrom long row tuple explodes",
			query: model.Query{
				Cmd:          metadata.CmdCopyFrom,
				ConstantName: "COPY_WIDE",
				FuncName:     "copy_wide",
				Table:        &plugin.Identifier{Name: "wide"},
				Params: []model.QueryValue{{
					EmitTable: true,
					Name:      "params",
					Type:      model.PyType{Type: "CopyWideParams", IsList: true},
					Table: &model.Table{
						Name: "CopyWideParams",
						Columns: []model.Column{
							{Name: longName, DBName: "a", Type: model.PyType{Type: "int", SQLType: "bigint"}, Number: 1},
							{Name: "b", DBName: "b", Type: model.PyType{Type: "str", SQLType: "text"}, Number: 2},
						},
					},
				}},
				Returns: model.QueryValue{Type: model.PyType{Type: "int"}},
			},
			want: strings.Join([]string{
				"async def copy_wide(conn: ConnectionLike, params: collections.abc.Sequence[CopyWideParams]) -> int:",
				"    async with conn.cursor() as cur:",
				`        async with cur.copy('COPY "wide" ("a", "b") FROM STDIN') as copy:`,
				"            for param in params:",
				"                await copy.write_row((",
				"                    param." + longName + ",",
				"                    param.b,",
				"                ))",
				"        return cur.rowcount",
				"",
			}, "\n"),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			d := newPsycopg(t)
			conf := psycopgTestConfig()
			body := writer.NewCodeWriter(conf)
			d.WriteQueryFunc(body, conf, tc.query, 0)
			if got := body.String(); got != tc.want {
				t.Errorf("WriteQueryFunc() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestPsycopgWriteQueryFuncBundledParams(t *testing.T) {
	t.Parallel()
	d := newPsycopg(t)
	conf := psycopgTestConfig()
	body := writer.NewCodeWriter(conf)
	// A query_parameter_limit bundle: fields bind via their sqlc numbers.
	query := model.Query{
		Cmd:          metadata.CmdExec,
		ConstantName: "UPDATE_AUTHOR",
		FuncName:     "update_author",
		Params: []model.QueryValue{
			{
				EmitTable: true,
				Name:      "params",
				Type:      model.PyType{Type: "UpdateAuthorParams"},
				Table: &model.Table{
					Name: "UpdateAuthorParams",
					Columns: []model.Column{
						{Name: "name", DBName: "name", Type: model.PyType{Type: "str", SQLType: "text"}, Number: 1},
						{Name: "id_", DBName: "id", Type: model.PyType{Type: "int", SQLType: "bigint"}, Number: 2},
					},
				},
			},
		},
		Returns: model.QueryValue{Type: model.PyType{Type: "None"}},
	}
	d.WriteQueryFunc(body, conf, query, 0)
	want := strings.Join([]string{
		"async def update_author(conn: ConnectionLike, params: UpdateAuthorParams) -> None:",
		`    await conn.execute(UPDATE_AUTHOR, {"p1": params.name, "p2": params.id_})`,
		"",
	}, "\n")
	if got := body.String(); got != want {
		t.Errorf("WriteQueryFunc() = %q, want %q", got, want)
	}
}
