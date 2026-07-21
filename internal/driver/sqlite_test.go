package driver_test

import (
	"strings"
	"testing"

	"github.com/rayakame/sqlc-gen-better-python/internal/config"
	"github.com/rayakame/sqlc-gen-better-python/internal/driver"
	"github.com/rayakame/sqlc-gen-better-python/internal/model"
	"github.com/rayakame/sqlc-gen-better-python/internal/utils"
	"github.com/rayakame/sqlc-gen-better-python/internal/writer"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
)

func newTestSqliteDriver(t *testing.T, sqlDriver config.SQLDriver) (driver.Driver, *config.Config) {
	t.Helper()
	conf := &config.Config{
		SqlDriver:           sqlDriver,
		EmitDocstrings:      config.DocstringConventionNone,
		EmitDocstringsSQL:   utils.ToPtr(true),
		IndentChar:          " ",
		CharsPerIndentLevel: 4,
	}
	d, err := driver.New(conf)
	if err != nil {
		t.Fatalf("driver.New() error = %v", err)
	}

	return d, conf
}

func sqliteAuthorReturn() model.QueryValue {
	return model.QueryValue{
		Table: &model.Table{
			Name: "Author",
			Columns: []model.Column{
				{Name: "id", Type: model.PyType{Type: "int", SQLType: "integer"}},
				{Name: "name", Type: model.PyType{Type: "str", SQLType: "text"}},
			},
		},
		Type: model.PyType{Type: "models.Author"},
	}
}

func TestSqliteDriverIdentity(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		sqlDriver config.SQLDriver
		wantName  string
		wantConn  string
		wantAsync bool
	}{
		{
			name:      "sqlite3 sync",
			sqlDriver: config.SQLDriverSQLite,
			wantName:  "sqlite3",
			wantConn:  "sqlite3.Connection",
			wantAsync: false,
		},
		{
			name:      "aiosqlite async",
			sqlDriver: config.SQLDriverAioSQLite,
			wantName:  "aiosqlite",
			wantConn:  "aiosqlite.Connection",
			wantAsync: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			d, _ := newTestSqliteDriver(t, tc.sqlDriver)
			if got := d.Name(); got != tc.wantName {
				t.Errorf("Name() = %q, want %q", got, tc.wantName)
			}
			if got := d.ConnType(); got != tc.wantConn {
				t.Errorf("ConnType() = %q, want %q", got, tc.wantConn)
			}
			if got := d.IsAsync(); got != tc.wantAsync {
				t.Errorf("IsAsync() = %v, want %v", got, tc.wantAsync)
			}
			if got := d.TypeCheckingHook(); got != nil {
				t.Errorf("TypeCheckingHook() = %v, want nil", got)
			}
			// sqlite drivers never convert inline, not even for convertible types.
			if got := d.ConvertsInline("date"); got {
				t.Errorf("ConvertsInline(\"date\") = %v, want false", got)
			}
		})
	}
}

func TestSqliteSupportsCommand(t *testing.T) {
	t.Parallel()
	cases := []struct {
		cmd  string
		want bool
	}{
		{cmd: metadata.CmdExec, want: true},
		{cmd: metadata.CmdExecResult, want: true},
		{cmd: metadata.CmdExecLastId, want: true},
		{cmd: metadata.CmdExecRows, want: true},
		{cmd: metadata.CmdOne, want: true},
		{cmd: metadata.CmdMany, want: true},
		{cmd: metadata.CmdCopyFrom, want: false},
		{cmd: ":batchexec", want: false},
	}
	for _, tc := range cases {
		t.Run(tc.cmd, func(t *testing.T) {
			t.Parallel()
			d, _ := newTestSqliteDriver(t, config.SQLDriverSQLite)
			if got := d.SupportsCommand(tc.cmd); got != tc.want {
				t.Errorf("SupportsCommand(%q) = %v, want %v", tc.cmd, got, tc.want)
			}
		})
	}
}

func TestSqliteNeedsConversion(t *testing.T) {
	t.Parallel()
	cases := []struct {
		sqlType string
		want    bool
	}{
		{sqlType: "date", want: true},
		{sqlType: "datetime", want: true},
		{sqlType: "timestamp", want: true},
		{sqlType: "decimal", want: true},
		{sqlType: "decimal(10,5)", want: true},
		{sqlType: "bool", want: true},
		{sqlType: "boolean", want: true},
		{sqlType: "blob", want: true},
		{sqlType: "text", want: false},
		{sqlType: "integer", want: false},
	}
	for _, tc := range cases {
		t.Run(tc.sqlType, func(t *testing.T) {
			t.Parallel()
			d, _ := newTestSqliteDriver(t, config.SQLDriverAioSQLite)
			if got := d.NeedsConversion(tc.sqlType); got != tc.want {
				t.Errorf("NeedsConversion(%q) = %v, want %v", tc.sqlType, got, tc.want)
			}
		})
	}
}

func TestSqliteWriteQueryResultsClass(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		sqlDriver config.SQLDriver
		want      string
	}{
		{
			name:      "sync sqlite3",
			sqlDriver: config.SQLDriverSQLite,
			want: strings.Join([]string{
				"class QueryResults[T]:",
				`    __slots__ = ("_args", "_conn", "_cursor", "_decode_hook", "_iterator", "_sql")`,
				"",
				"    def __init__(",
				"        self,",
				"        conn: sqlite3.Connection,",
				"        sql: str,",
				"        decode_hook: collections.abc.Callable[[sqlite3.Row], T],",
				"        *args: QueryResultsArgsType,",
				"    ) -> None:",
				"        self._conn = conn",
				"        self._sql = sql",
				"        self._decode_hook = decode_hook",
				"        self._args = args",
				"        self._cursor: sqlite3.Cursor | None = None",
				"        self._iterator: collections.abc.Iterator[sqlite3.Row] | None = None",
				"",
				"    def __iter__(self) -> QueryResults[T]:",
				"        return self",
				"",
				"    def __call__(",
				"        self,",
				"    ) -> collections.abc.Sequence[T]:",
				"        result = self._conn.execute(self._sql, self._args).fetchall()",
				"        return [self._decode_hook(row) for row in result]",
				"",
				"    def __next__(self) -> T:",
				"        if self._cursor is None or self._iterator is None:",
				"            self._cursor: sqlite3.Cursor | None = self._conn.execute(self._sql, self._args)",
				"            self._iterator = self._cursor.__iter__()",
				"        try:",
				"            record = self._iterator.__next__()",
				"        except StopIteration:",
				"            self._cursor = None",
				"            self._iterator = None",
				"            raise",
				"        return self._decode_hook(record)",
				"",
			}, "\n"),
		},
		{
			name:      "async aiosqlite",
			sqlDriver: config.SQLDriverAioSQLite,
			want: strings.Join([]string{
				"class QueryResults[T]:",
				`    __slots__ = ("_args", "_conn", "_cursor", "_decode_hook", "_iterator", "_sql")`,
				"",
				"    def __init__(",
				"        self,",
				"        conn: aiosqlite.Connection,",
				"        sql: str,",
				"        decode_hook: collections.abc.Callable[[sqlite3.Row], T],",
				"        *args: QueryResultsArgsType,",
				"    ) -> None:",
				"        self._conn = conn",
				"        self._sql = sql",
				"        self._decode_hook = decode_hook",
				"        self._args = args",
				"        self._cursor: aiosqlite.Cursor | None = None",
				"        self._iterator: collections.abc.AsyncIterator[sqlite3.Row] | None = None",
				"",
				"    def __aiter__(self) -> QueryResults[T]:",
				"        return self",
				"",
				"    def __await__(",
				"        self,",
				"    ) -> collections.abc.Generator[None, None, collections.abc.Sequence[T]]:",
				"        async def _wrapper() -> collections.abc.Sequence[T]:",
				"            result = await (await self._conn.execute(self._sql, self._args)).fetchall()",
				"            return [self._decode_hook(row) for row in result]",
				"",
				"        return _wrapper().__await__()",
				"",
				"    async def __anext__(self) -> T:",
				"        if self._cursor is None or self._iterator is None:",
				"            self._cursor: aiosqlite.Cursor | None = await self._conn.execute(self._sql, self._args)",
				"            self._iterator = self._cursor.__aiter__()",
				"        try:",
				"            record = await self._iterator.__anext__()",
				"        except StopAsyncIteration:",
				"            self._cursor = None",
				"            self._iterator = None",
				"            raise",
				"        return self._decode_hook(record)",
				"",
			}, "\n"),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			d, conf := newTestSqliteDriver(t, tc.sqlDriver)
			body := writer.NewCodeWriter(conf)
			if got := d.WriteQueryResultsClass(body); got != "QueryResults" {
				t.Errorf("WriteQueryResultsClass() = %q, want %q", got, "QueryResults")
			}
			if got := body.String(); got != tc.want {
				t.Errorf("WriteQueryResultsClass() wrote %q, want %q", got, tc.want)
			}
		})
	}
}

func TestSqliteWriteQueryFunc(t *testing.T) {
	t.Parallel()
	longName := strings.Repeat("p", 340)
	cases := []struct {
		name      string
		sqlDriver config.SQLDriver
		query     model.Query
		want      string
	}{
		{
			name:      "exec sync no params",
			sqlDriver: config.SQLDriverSQLite,
			query: model.Query{
				Cmd:          metadata.CmdExec,
				ConstantName: "DELETE_ALL",
				FuncName:     "delete_all",
				Returns:      model.QueryValue{Type: model.PyType{Type: "None"}},
			},
			want: strings.Join([]string{
				"def delete_all(conn: sqlite3.Connection) -> None:",
				"    conn.execute(DELETE_ALL)",
				"",
			}, "\n"),
		},
		{
			name:      "exec async single param",
			sqlDriver: config.SQLDriverAioSQLite,
			query: model.Query{
				Cmd:          metadata.CmdExec,
				ConstantName: "DELETE_AUTHOR",
				FuncName:     "delete_author",
				Params: []model.QueryValue{
					{Name: "author_id", Type: model.PyType{Type: "int", SQLType: "integer"}},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "None"}},
			},
			want: strings.Join([]string{
				"async def delete_author(conn: aiosqlite.Connection, *, author_id: int) -> None:",
				"    await conn.execute(DELETE_AUTHOR, (author_id,))",
				"",
			}, "\n"),
		},
		{
			name:      "exec sync override param converted",
			sqlDriver: config.SQLDriverSQLite,
			query: model.Query{
				Cmd:          metadata.CmdExec,
				ConstantName: "SET_CREATED_AT",
				FuncName:     "set_created_at",
				Params: []model.QueryValue{
					{Name: "created_at", Type: model.PyType{
						Type:        "float",
						SQLType:     "date",
						IsOverride:  true,
						DefaultType: "datetime.date",
					}},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "None"}},
			},
			want: strings.Join([]string{
				"def set_created_at(conn: sqlite3.Connection, *, created_at: float) -> None:",
				"    conn.execute(SET_CREATED_AT, (datetime.date(created_at),))",
				"",
			}, "\n"),
		},
		{
			name:      "exec sync long params hoisted",
			sqlDriver: config.SQLDriverSQLite,
			query: model.Query{
				Cmd:          metadata.CmdExec,
				ConstantName: "HOIST",
				FuncName:     "hoist",
				Params: []model.QueryValue{
					{Name: longName, Type: model.PyType{Type: "int", SQLType: "integer"}},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "None"}},
			},
			want: strings.Join([]string{
				"def hoist(",
				"    conn: sqlite3.Connection,",
				"    *,",
				"    " + longName + ": int,",
				") -> None:",
				"    sql_args = (",
				"        " + longName + ",",
				"    )",
				"    conn.execute(HOIST, sql_args)",
				"",
			}, "\n"),
		},
		{
			name:      "execresult sync no params",
			sqlDriver: config.SQLDriverSQLite,
			query: model.Query{
				Cmd:          metadata.CmdExecResult,
				ConstantName: "CREATE_TABLE",
				FuncName:     "create_table",
				Returns:      model.QueryValue{Type: model.PyType{Type: "None"}},
			},
			want: strings.Join([]string{
				"def create_table(conn: sqlite3.Connection) -> sqlite3.Cursor:",
				"    return conn.execute(CREATE_TABLE)",
				"",
			}, "\n"),
		},
		{
			name:      "execresult async no params",
			sqlDriver: config.SQLDriverAioSQLite,
			query: model.Query{
				Cmd:          metadata.CmdExecResult,
				ConstantName: "CREATE_TABLE",
				FuncName:     "create_table",
				Returns:      model.QueryValue{Type: model.PyType{Type: "None"}},
			},
			want: strings.Join([]string{
				"async def create_table(conn: aiosqlite.Connection) -> aiosqlite.Cursor:",
				"    return await conn.execute(CREATE_TABLE)",
				"",
			}, "\n"),
		},
		{
			name:      "execrows sync two params",
			sqlDriver: config.SQLDriverSQLite,
			query: model.Query{
				Cmd:          metadata.CmdExecRows,
				ConstantName: "UPDATE_NAMES",
				FuncName:     "update_names",
				Params: []model.QueryValue{
					{Name: "name", Type: model.PyType{Type: "str", SQLType: "text"}},
					{Name: "author_id", Type: model.PyType{Type: "int", SQLType: "integer"}},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "int"}},
			},
			want: strings.Join([]string{
				"def update_names(conn: sqlite3.Connection, *, name: str, author_id: int) -> int:",
				"    return conn.execute(UPDATE_NAMES, (name, author_id)).rowcount",
				"",
			}, "\n"),
		},
		{
			name:      "execrows async no params",
			sqlDriver: config.SQLDriverAioSQLite,
			query: model.Query{
				Cmd:          metadata.CmdExecRows,
				ConstantName: "CLEAR_NAMES",
				FuncName:     "clear_names",
				Returns:      model.QueryValue{Type: model.PyType{Type: "int"}},
			},
			want: strings.Join([]string{
				"async def clear_names(conn: aiosqlite.Connection) -> int:",
				"    return (await conn.execute(CLEAR_NAMES)).rowcount",
				"",
			}, "\n"),
		},
		{
			name:      "execlastid sync single param",
			sqlDriver: config.SQLDriverSQLite,
			query: model.Query{
				Cmd:          metadata.CmdExecLastId,
				ConstantName: "INSERT_AUTHOR",
				FuncName:     "insert_author",
				Params: []model.QueryValue{
					{Name: "name", Type: model.PyType{Type: "str", SQLType: "text"}},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "int", IsNullable: true}},
			},
			want: strings.Join([]string{
				"def insert_author(conn: sqlite3.Connection, *, name: str) -> int | None:",
				"    return conn.execute(INSERT_AUTHOR, (name,)).lastrowid",
				"",
			}, "\n"),
		},
		{
			name:      "execlastid async single param",
			sqlDriver: config.SQLDriverAioSQLite,
			query: model.Query{
				Cmd:          metadata.CmdExecLastId,
				ConstantName: "INSERT_AUTHOR",
				FuncName:     "insert_author",
				Params: []model.QueryValue{
					{Name: "name", Type: model.PyType{Type: "str", SQLType: "text"}},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "int", IsNullable: true}},
			},
			want: strings.Join([]string{
				"async def insert_author(conn: aiosqlite.Connection, *, name: str) -> int | None:",
				"    return (await conn.execute(INSERT_AUTHOR, (name,))).lastrowid",
				"",
			}, "\n"),
		},
		{
			name:      "one scalar sync",
			sqlDriver: config.SQLDriverSQLite,
			query: model.Query{
				Cmd:          metadata.CmdOne,
				ConstantName: "COUNT_AUTHORS",
				FuncName:     "count_authors",
				Returns:      model.QueryValue{Type: model.PyType{Type: "int", SQLType: "integer"}},
			},
			want: strings.Join([]string{
				"def count_authors(conn: sqlite3.Connection) -> int | None:",
				"    row = conn.execute(COUNT_AUTHORS).fetchone()",
				"    if row is None:",
				"        return None",
				"    return row[0]",
				"",
			}, "\n"),
		},
		{
			name:      "one scalar enum sync converted",
			sqlDriver: config.SQLDriverSQLite,
			query: model.Query{
				Cmd:          metadata.CmdOne,
				ConstantName: "GET_STATUS",
				FuncName:     "get_status",
				Params: []model.QueryValue{
					{Name: "author_id", Type: model.PyType{Type: "int", SQLType: "integer"}},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "models.Status", SQLType: "text", IsEnum: true}},
			},
			want: strings.Join([]string{
				"def get_status(conn: sqlite3.Connection, *, author_id: int) -> models.Status | None:",
				"    row = conn.execute(GET_STATUS, (author_id,)).fetchone()",
				"    if row is None:",
				"        return None",
				"    return models.Status(row[0])",
				"",
			}, "\n"),
		},
		{
			name:      "one struct async",
			sqlDriver: config.SQLDriverAioSQLite,
			query: model.Query{
				Cmd:          metadata.CmdOne,
				ConstantName: "GET_AUTHOR",
				FuncName:     "get_author",
				Params: []model.QueryValue{
					{Name: "author_id", Type: model.PyType{Type: "int", SQLType: "integer"}},
				},
				Returns: sqliteAuthorReturn(),
			},
			want: strings.Join([]string{
				"async def get_author(conn: aiosqlite.Connection, *, author_id: int) -> models.Author | None:",
				"    row = await (await conn.execute(GET_AUTHOR, (author_id,))).fetchone()",
				"    if row is None:",
				"        return None",
				"    return models.Author(id=row[0], name=row[1])",
				"",
			}, "\n"),
		},
		{
			name:      "many struct sync decode hook",
			sqlDriver: config.SQLDriverSQLite,
			query: model.Query{
				Cmd:          metadata.CmdMany,
				ConstantName: "LIST_AUTHORS",
				FuncName:     "list_authors",
				Params: []model.QueryValue{
					{Name: "name", Type: model.PyType{Type: "str", SQLType: "text"}},
				},
				Returns: sqliteAuthorReturn(),
			},
			want: strings.Join([]string{
				"def list_authors(conn: sqlite3.Connection, *, name: str) -> QueryResults[models.Author]:",
				"    def _decode_hook(row: sqlite3.Row) -> models.Author:",
				"        return models.Author(id=row[0], name=row[1])",
				"",
				"    return QueryResults(conn, LIST_AUTHORS, _decode_hook, name)",
				"",
			}, "\n"),
		},
		{
			name:      "many scalar async itemgetter",
			sqlDriver: config.SQLDriverAioSQLite,
			query: model.Query{
				Cmd:          metadata.CmdMany,
				ConstantName: "LIST_IDS",
				FuncName:     "list_ids",
				Returns:      model.QueryValue{Type: model.PyType{Type: "int", SQLType: "integer"}},
			},
			want: strings.Join([]string{
				"def list_ids(conn: aiosqlite.Connection) -> QueryResults[int]:",
				"    return QueryResults(conn, LIST_IDS, operator.itemgetter(0))",
				"",
			}, "\n"),
		},
		{
			name:      "many struct sync slice expanded after decode hook",
			sqlDriver: config.SQLDriverSQLite,
			query: model.Query{
				Cmd:          metadata.CmdMany,
				ConstantName: "GET_ROWS",
				FuncName:     "get_rows",
				Params: []model.QueryValue{
					{Name: "ids", Type: model.PyType{Type: "int", SQLType: "integer", IsList: true, SqlcSliceName: "ids"}},
				},
				Returns: sqliteAuthorReturn(),
			},
			want: strings.Join([]string{
				"def get_rows(conn: sqlite3.Connection, *, ids: collections.abc.Sequence[int]) -> QueryResults[models.Author]:",
				"    def _decode_hook(row: sqlite3.Row) -> models.Author:",
				"        return models.Author(id=row[0], name=row[1])",
				"",
				`    sql = GET_ROWS.replace("/*SLICE:ids*/?", ",".join("?" * len(ids)) or "NULL", 1)`,
				"    return QueryResults(conn, sql, _decode_hook, *ids)",
				"",
			}, "\n"),
		},
		{
			name:      "one async slice between plain params",
			sqlDriver: config.SQLDriverAioSQLite,
			query: model.Query{
				Cmd:          metadata.CmdOne,
				ConstantName: "GET_ROW",
				FuncName:     "get_row",
				Params: []model.QueryValue{
					{Name: "name", Type: model.PyType{Type: "str", SQLType: "text"}},
					{Name: "ids", Type: model.PyType{Type: "int", SQLType: "integer", IsList: true, SqlcSliceName: "ids"}},
					{Name: "note", Type: model.PyType{Type: "str", SQLType: "text", IsNullable: true}},
				},
				Returns: sqliteAuthorReturn(),
			},
			want: strings.Join([]string{
				"async def get_row(conn: aiosqlite.Connection, *, name: str, ids: collections.abc.Sequence[int], note: str | None) -> models.Author | None:",
				`    sql = GET_ROW.replace("/*SLICE:ids*/?", ",".join("?" * len(ids)) or "NULL", 1)`,
				"    row = await (await conn.execute(sql, (name, *ids, note))).fetchone()",
				"    if row is None:",
				"        return None",
				"    return models.Author(id=row[0], name=row[1])",
				"",
			}, "\n"),
		},
		{
			name:      "exec sync two slices replaced sequentially",
			sqlDriver: config.SQLDriverSQLite,
			query: model.Query{
				Cmd:          metadata.CmdExec,
				ConstantName: "DELETE_ROWS",
				FuncName:     "delete_rows",
				Params: []model.QueryValue{
					{Name: "ids", Type: model.PyType{Type: "int", SQLType: "integer", IsList: true, SqlcSliceName: "ids"}},
					{Name: "names", Type: model.PyType{Type: "str", SQLType: "text", IsList: true, SqlcSliceName: "names"}},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "None"}},
			},
			want: strings.Join([]string{
				"def delete_rows(conn: sqlite3.Connection, *, ids: collections.abc.Sequence[int], names: collections.abc.Sequence[str]) -> None:",
				`    sql = DELETE_ROWS.replace("/*SLICE:ids*/?", ",".join("?" * len(ids)) or "NULL", 1)`,
				`    sql = sql.replace("/*SLICE:names*/?", ",".join("?" * len(names)) or "NULL", 1)`,
				"    conn.execute(sql, (*ids, *names))",
				"",
			}, "\n"),
		},
		{
			// sqlc merges same-named sqlc.slice uses into one parameter but
			// keeps a marker per use site: all of them are replaced and the
			// arguments are repeated once per occurrence.
			name:      "exec sync reused slice replaces all markers and repeats args",
			sqlDriver: config.SQLDriverSQLite,
			query: model.Query{
				Cmd:          metadata.CmdExec,
				SQL:          "DELETE FROM t WHERE id IN (/*SLICE:ids*/?) OR ref_id IN (/*SLICE:ids*/?)",
				ConstantName: "DELETE_LINKED",
				FuncName:     "delete_linked",
				Params: []model.QueryValue{
					{Name: "ids", Type: model.PyType{Type: "int", SQLType: "integer", IsList: true, SqlcSliceName: "ids"}},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "None"}},
			},
			want: strings.Join([]string{
				"def delete_linked(conn: sqlite3.Connection, *, ids: collections.abc.Sequence[int]) -> None:",
				`    sql = DELETE_LINKED.replace("/*SLICE:ids*/?", ",".join("?" * len(ids)) or "NULL")`,
				"    conn.execute(sql, (*ids, *ids))",
				"",
			}, "\n"),
		},
		{
			// A plain placeholder between the reuse sites: arguments follow
			// the SQL text order, not the parameter order.
			name:      "exec sync reused slice keeps text order around plain params",
			sqlDriver: config.SQLDriverSQLite,
			query: model.Query{
				Cmd:          metadata.CmdExec,
				SQL:          "DELETE FROM t WHERE id IN (/*SLICE:ids*/?) AND name = ? AND ref_id IN (/*SLICE:ids*/?)",
				ConstantName: "DELETE_BETWEEN",
				FuncName:     "delete_between",
				Params: []model.QueryValue{
					{Name: "ids", Type: model.PyType{Type: "int", SQLType: "integer", IsList: true, SqlcSliceName: "ids"}},
					{Name: "name", Type: model.PyType{Type: "str", SQLType: "text"}},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "None"}},
			},
			want: strings.Join([]string{
				"def delete_between(conn: sqlite3.Connection, *, ids: collections.abc.Sequence[int], name: str) -> None:",
				`    sql = DELETE_BETWEEN.replace("/*SLICE:ids*/?", ",".join("?" * len(ids)) or "NULL")`,
				"    conn.execute(sql, (*ids, name, *ids))",
				"",
			}, "\n"),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			d, conf := newTestSqliteDriver(t, tc.sqlDriver)
			body := writer.NewCodeWriter(conf)
			d.WriteQueryFunc(body, conf, tc.query, 0)
			if got := body.String(); got != tc.want {
				t.Errorf("WriteQueryFunc() = %q, want %q", got, tc.want)
			}
		})
	}
}
