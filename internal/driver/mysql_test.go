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

func mysqlTestConfig(async bool) *config.Config {
	sqlDriver := config.SQLDriverPymysql
	if async {
		sqlDriver = config.SQLDriverAsyncmy
	}

	return &config.Config{
		SqlDriver:           sqlDriver,
		EmitDocstrings:      config.DocstringConventionNone,
		EmitDocstringsSQL:   utils.ToPtr(true),
		IndentChar:          " ",
		CharsPerIndentLevel: 4,
		OmitKwargsLimit:     8,
	}
}

func newMysql(t *testing.T, async bool) driver.Driver {
	t.Helper()
	d, err := driver.New(mysqlTestConfig(async))
	if err != nil {
		t.Fatalf("driver.New() error = %v", err)
	}

	return d
}

func mysqlUserReturn() model.QueryValue {
	return model.QueryValue{
		Table: &model.Table{
			Name: "User",
			Columns: []model.Column{
				{Name: "id_", Type: model.PyType{Type: "int", SQLType: "int"}},
				{Name: "avatar", Type: model.PyType{Type: "memoryview", SQLType: "blob"}},
				{Name: "active", Type: model.PyType{Type: "bool", SQLType: "tinyint"}},
				{Name: "mood", Type: model.PyType{Type: "enums.Mood", SQLType: "enum", IsEnum: true}},
				{Name: "name", Type: model.PyType{Type: "str", SQLType: "varchar"}},
			},
		},
		Type: model.PyType{Type: "models.User"},
	}
}

// mysqlUserDecode is the row-decoding expression for mysqlUserReturn: the
// binary family becomes memoryview, tinyint (the tinyint(1) bool spelling)
// wraps in bool, enums wrap in their class, everything else passes through.
const mysqlUserDecode = "models.User(id_=row[0], avatar=memoryview(row[1]), active=bool(row[2]), mood=enums.Mood(row[3]), name=row[4])"

func TestMysqlDriverMetadata(t *testing.T) {
	t.Parallel()
	sync := newMysql(t, false)
	if got := sync.Name(); got != "pymysql" {
		t.Errorf("sync Name() = %q, want %q", got, "pymysql")
	}
	if got := sync.ConnType(); got != "pymysql.Connection" {
		t.Errorf("sync ConnType() = %q, want %q", got, "pymysql.Connection")
	}
	if sync.IsAsync() {
		t.Error("sync IsAsync() = true, want false")
	}
	if got := sync.TypeCheckingHook(); got != nil {
		t.Errorf("sync TypeCheckingHook() = %v, want nil", got)
	}

	async := newMysql(t, true)
	if got := async.Name(); got != "asyncmy" {
		t.Errorf("async Name() = %q, want %q", got, "asyncmy")
	}
	if got := async.ConnType(); got != "asyncmy.Connection" {
		t.Errorf("async ConnType() = %q, want %q", got, "asyncmy.Connection")
	}
	if !async.IsAsync() {
		t.Error("async IsAsync() = false, want true")
	}
	if got := async.TypeCheckingHook(); got != nil {
		t.Errorf("async TypeCheckingHook() = %v, want nil", got)
	}
}

func TestMysqlSupportsCommand(t *testing.T) {
	t.Parallel()
	for _, async := range []bool{false, true} {
		d := newMysql(t, async)
		for cmd, want := range map[string]bool{
			metadata.CmdExec:       true,
			metadata.CmdExecResult: true,
			metadata.CmdExecLastId: true,
			metadata.CmdExecRows:   true,
			metadata.CmdOne:        true,
			metadata.CmdMany:       true,
			metadata.CmdCopyFrom:   false,
			metadata.CmdBatchExec:  false,
			metadata.CmdBatchMany:  false,
			metadata.CmdBatchOne:   false,
		} {
			if got := d.SupportsCommand(cmd); got != want {
				t.Errorf("async=%v SupportsCommand(%q) = %v, want %v", async, cmd, got, want)
			}
		}
	}
}

func TestMysqlConversions(t *testing.T) {
	t.Parallel()
	d := newMysql(t, false)
	// Everything converts inline: the drivers' converter maps are
	// per-connection, so both checks agree on the same type set.
	for sqlType, want := range map[string]bool{
		"blob": true, "binary": true, "varbinary": true, "tinyblob": true,
		"mediumblob": true, "longblob": true, "bit": true,
		"tinyint": true, "bool": true, "boolean": true,
		"text": false, "int": false, "datetime": false, "json": false,
		"decimal": false, "varchar": false,
	} {
		if got := d.NeedsConversion(sqlType); got != want {
			t.Errorf("NeedsConversion(%q) = %v, want %v", sqlType, got, want)
		}
		if got := d.ConvertsInline(sqlType); got != want {
			t.Errorf("ConvertsInline(%q) = %v, want %v", sqlType, got, want)
		}
	}
}

func TestMysqlWriteConversionSetup(t *testing.T) {
	t.Parallel()
	for _, async := range []bool{false, true} {
		d := newMysql(t, async)
		conf := mysqlTestConfig(async)
		w := writer.NewCodeWriter(conf)
		queries := []model.Query{{Returns: model.QueryValue{Type: model.PyType{Type: "memoryview", SQLType: "blob"}}}}
		if d.WriteConversionSetup(w, conf, queries) {
			t.Errorf("async=%v WriteConversionSetup() = true, want false: mysql converts inline", async)
		}
		if got := w.String(); got != "" {
			t.Errorf("async=%v WriteConversionSetup() wrote %q, want nothing", async, got)
		}
	}
}

func TestMysqlWriteQueryResultsClassSync(t *testing.T) {
	t.Parallel()
	d := newMysql(t, false)
	w := writer.NewCodeWriter(mysqlTestConfig(false))
	if got := d.WriteQueryResultsClass(w); got != "QueryResults" {
		t.Errorf("WriteQueryResultsClass() = %q, want %q", got, "QueryResults")
	}
	want := strings.Join([]string{
		"class QueryResults[T]:",
		`    __slots__ = ("_args", "_conn", "_cursor", "_decode_hook", "_sql")`,
		"",
		"    def __init__(",
		"        self,",
		"        conn: pymysql.Connection,",
		"        sql: str,",
		"        decode_hook: collections.abc.Callable[[tuple[typing.Any, ...]], T],",
		"        *args: QueryResultsArgsType,",
		"    ) -> None:",
		"        self._conn = conn",
		"        self._sql = sql",
		"        self._decode_hook = decode_hook",
		"        self._args = args",
		"        self._cursor: pymysql.cursors.Cursor | None = None",
		"",
		"    def __iter__(self) -> QueryResults[T]:",
		"        return self",
		"",
		"    def __call__(",
		"        self,",
		"    ) -> collections.abc.Sequence[T]:",
		"        cur = self._conn.cursor()",
		"        cur.execute(self._sql, self._args)",
		"        result = cur.fetchall()",
		"        cur.close()",
		"        return [self._decode_hook(row) for row in result]",
		"",
		"    def __next__(self) -> T:",
		"        if self._cursor is None:",
		"            self._cursor = self._conn.cursor()",
		"            self._cursor.execute(self._sql, self._args)",
		"        record = self._cursor.fetchone()",
		"        if record is None:",
		"            self._cursor = None",
		"            raise StopIteration",
		"        return self._decode_hook(record)",
	}, "\n") + "\n"
	if got := w.String(); got != want {
		t.Errorf("WriteQueryResultsClass() wrote %q, want %q", got, want)
	}
}

func TestMysqlWriteQueryResultsClassAsync(t *testing.T) {
	t.Parallel()
	d := newMysql(t, true)
	w := writer.NewCodeWriter(mysqlTestConfig(true))
	if got := d.WriteQueryResultsClass(w); got != "QueryResults" {
		t.Errorf("WriteQueryResultsClass() = %q, want %q", got, "QueryResults")
	}
	want := strings.Join([]string{
		"class QueryResults[T]:",
		`    __slots__ = ("_args", "_conn", "_cursor", "_decode_hook", "_sql")`,
		"",
		"    def __init__(",
		"        self,",
		"        conn: asyncmy.Connection,",
		"        sql: str,",
		"        decode_hook: collections.abc.Callable[[tuple[typing.Any, ...]], T],",
		"        *args: QueryResultsArgsType,",
		"    ) -> None:",
		"        self._conn = conn",
		"        self._sql = sql",
		"        self._decode_hook = decode_hook",
		"        self._args = args",
		"        self._cursor: asyncmy.cursors.Cursor | None = None",
		"",
		"    def __aiter__(self) -> QueryResults[T]:",
		"        return self",
		"",
		"    def __await__(",
		"        self,",
		"    ) -> collections.abc.Generator[None, None, collections.abc.Sequence[T]]:",
		"        async def _wrapper() -> collections.abc.Sequence[T]:",
		"            cur = self._conn.cursor()",
		"            await cur.execute(self._sql, self._args)",
		"            result = await cur.fetchall()",
		"            await cur.close()",
		"            return [self._decode_hook(row) for row in result]",
		"",
		"        return _wrapper().__await__()",
		"",
		"    async def __anext__(self) -> T:",
		"        if self._cursor is None:",
		"            self._cursor = self._conn.cursor()",
		"            await self._cursor.execute(self._sql, self._args)",
		"        record = await self._cursor.fetchone()",
		"        if record is None:",
		"            self._cursor = None",
		"            raise StopAsyncIteration",
		"        return self._decode_hook(record)",
	}, "\n") + "\n"
	if got := w.String(); got != want {
		t.Errorf("WriteQueryResultsClass() wrote %q, want %q", got, want)
	}
}

func TestMysqlWriteQueryFuncSync(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		query model.Query
		want  string
	}{
		{
			name: "exec opens a cursor in a with block",
			query: model.Query{
				Cmd:          metadata.CmdExec,
				ConstantName: "TOUCH",
				FuncName:     "touch",
				Params: []model.QueryValue{
					{Name: "a", Type: model.PyType{Type: "int", SQLType: "int"}},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "None"}},
			},
			want: strings.Join([]string{
				"def touch(conn: pymysql.Connection, a: int) -> None:",
				"    with conn.cursor() as cur:",
				"        cur.execute(TOUCH, (a,))",
				"",
			}, "\n"),
		},
		{
			name: "exec blob param wire-converts to bytes",
			query: model.Query{
				Cmd:          metadata.CmdExec,
				ConstantName: "SET_AVATAR",
				FuncName:     "set_avatar",
				Params: []model.QueryValue{
					{Name: "avatar", Type: model.PyType{Type: "memoryview", SQLType: "blob"}},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "None"}},
			},
			want: strings.Join([]string{
				"def set_avatar(conn: pymysql.Connection, avatar: memoryview) -> None:",
				"    with conn.cursor() as cur:",
				"        cur.execute(SET_AVATAR, (bytes(avatar),))",
				"",
			}, "\n"),
		},
		{
			name: "exec nullable blob param guards the wire conversion",
			query: model.Query{
				Cmd:          metadata.CmdExec,
				ConstantName: "SET_AVATAR",
				FuncName:     "set_avatar",
				Params: []model.QueryValue{
					{Name: "avatar", Type: model.PyType{Type: "memoryview", SQLType: "blob", IsNullable: true}},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "None"}},
			},
			want: strings.Join([]string{
				"def set_avatar(conn: pymysql.Connection, avatar: memoryview | None) -> None:",
				"    with conn.cursor() as cur:",
				"        cur.execute(SET_AVATAR, (bytes(avatar) if avatar is not None else None,))",
				"",
			}, "\n"),
		},
		{
			name: "execresult hands the open cursor to the caller without a with block",
			query: model.Query{
				Cmd:          metadata.CmdExecResult,
				ConstantName: "RUN",
				FuncName:     "run",
				Returns:      model.QueryValue{Type: model.PyType{Type: "None"}},
			},
			want: strings.Join([]string{
				"def run(conn: pymysql.Connection) -> pymysql.cursors.Cursor:",
				"    cur = conn.cursor()",
				"    cur.execute(RUN)",
				"    return cur",
				"",
			}, "\n"),
		},
		{
			name: "execrows returns the execute call inside the with block",
			query: model.Query{
				Cmd:          metadata.CmdExecRows,
				ConstantName: "BUMP",
				FuncName:     "bump",
				Returns:      model.QueryValue{Type: model.PyType{Type: "int"}},
			},
			want: strings.Join([]string{
				"def bump(conn: pymysql.Connection) -> int:",
				"    with conn.cursor() as cur:",
				"        return cur.execute(BUMP)",
				"",
			}, "\n"),
		},
		{
			name: "execlastid executes then returns lastrowid",
			query: model.Query{
				Cmd:          metadata.CmdExecLastId,
				ConstantName: "ADD",
				FuncName:     "add",
				Params: []model.QueryValue{
					{Name: "name", Type: model.PyType{Type: "str", SQLType: "varchar"}},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "int", IsNullable: true}},
			},
			want: strings.Join([]string{
				"def add(conn: pymysql.Connection, name: str) -> int | None:",
				"    with conn.cursor() as cur:",
				"        cur.execute(ADD, (name,))",
				"        return cur.lastrowid",
				"",
			}, "\n"),
		},
		{
			name: "one scalar fetches inside the with block and returns outside",
			query: model.Query{
				Cmd:          metadata.CmdOne,
				ConstantName: "COUNT",
				FuncName:     "count",
				Returns:      model.QueryValue{Type: model.PyType{Type: "int", SQLType: "bigint"}},
			},
			want: strings.Join([]string{
				"def count(conn: pymysql.Connection) -> int | None:",
				"    with conn.cursor() as cur:",
				"        cur.execute(COUNT)",
				"        row = cur.fetchone()",
				"    if row is None:",
				"        return None",
				"    return row[0]",
				"",
			}, "\n"),
		},
		{
			name: "one struct converts blob and tinyint-bool and enum columns",
			query: model.Query{
				Cmd:          metadata.CmdOne,
				ConstantName: "GET_USER",
				FuncName:     "get_user",
				Params: []model.QueryValue{
					{Name: "id_", Type: model.PyType{Type: "int", SQLType: "int"}},
				},
				Returns: mysqlUserReturn(),
			},
			want: strings.Join([]string{
				"def get_user(conn: pymysql.Connection, id_: int) -> models.User | None:",
				"    with conn.cursor() as cur:",
				"        cur.execute(GET_USER, (id_,))",
				"        row = cur.fetchone()",
				"    if row is None:",
				"        return None",
				"    return " + mysqlUserDecode,
				"",
			}, "\n"),
		},
		{
			name: "many scalar without conversion uses itemgetter",
			query: model.Query{
				Cmd:          metadata.CmdMany,
				ConstantName: "LIST_IDS",
				FuncName:     "list_ids",
				Returns:      model.QueryValue{Type: model.PyType{Type: "int", SQLType: "int"}},
			},
			want: strings.Join([]string{
				"def list_ids(conn: pymysql.Connection) -> QueryResults[int]:",
				"    return QueryResults(conn, LIST_IDS, operator.itemgetter(0))",
				"",
			}, "\n"),
		},
		{
			name: "many struct emits a decode hook and forwards params",
			query: model.Query{
				Cmd:          metadata.CmdMany,
				ConstantName: "LIST_USERS",
				FuncName:     "list_users",
				Params: []model.QueryValue{
					{Name: "name", Type: model.PyType{Type: "str", SQLType: "varchar"}},
				},
				Returns: mysqlUserReturn(),
			},
			want: strings.Join([]string{
				"def list_users(conn: pymysql.Connection, name: str) -> QueryResults[models.User]:",
				"    def _decode_hook(row: tuple[typing.Any, ...]) -> models.User:",
				"        return " + mysqlUserDecode,
				"",
				"    return QueryResults(conn, LIST_USERS, _decode_hook, name)",
				"",
			}, "\n"),
		},
		{
			name: "many struct slice expanded after decode hook",
			query: model.Query{
				Cmd:          metadata.CmdMany,
				ConstantName: "LIST_BY_IDS",
				FuncName:     "list_by_ids",
				SQL:          "SELECT id FROM users WHERE id IN (/*SLICE:ids*/%s)",
				Params: []model.QueryValue{
					{Name: "ids", Type: model.PyType{Type: "int", SQLType: "int", IsList: true, SqlcSliceName: "ids"}},
				},
				Returns: mysqlUserReturn(),
			},
			want: strings.Join([]string{
				"def list_by_ids(conn: pymysql.Connection, ids: collections.abc.Sequence[int]) -> QueryResults[models.User]:",
				"    def _decode_hook(row: tuple[typing.Any, ...]) -> models.User:",
				"        return " + mysqlUserDecode,
				"",
				`    sql = LIST_BY_IDS.replace("/*SLICE:ids*/%s", ",".join(("%s",) * len(ids)) or "NULL", 1)`,
				"    return QueryResults(conn, sql, _decode_hook, *ids)",
				"",
			}, "\n"),
		},
		{
			name: "one single slice marker replaces once and star-unpacks",
			query: model.Query{
				Cmd:          metadata.CmdOne,
				ConstantName: "GET_BY_IDS",
				FuncName:     "get_by_ids",
				SQL:          "SELECT count(*) FROM t WHERE id IN (/*SLICE:ids*/%s)",
				Params: []model.QueryValue{
					{Name: "ids", Type: model.PyType{Type: "int", SQLType: "int", IsList: true, SqlcSliceName: "ids"}},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "int", SQLType: "bigint"}},
			},
			want: strings.Join([]string{
				"def get_by_ids(conn: pymysql.Connection, ids: collections.abc.Sequence[int]) -> int | None:",
				`    sql = GET_BY_IDS.replace("/*SLICE:ids*/%s", ",".join(("%s",) * len(ids)) or "NULL", 1)`,
				"    with conn.cursor() as cur:",
				"        cur.execute(sql, (*ids,))",
				"        row = cur.fetchone()",
				"    if row is None:",
				"        return None",
				"    return row[0]",
				"",
			}, "\n"),
		},
		{
			// sqlc merges same-named sqlc.slice uses into one parameter but
			// keeps a marker per use site: all of them are replaced and the
			// arguments are repeated once per occurrence.
			name: "exec reused slice replaces all markers and repeats args",
			query: model.Query{
				Cmd:          metadata.CmdExec,
				ConstantName: "DELETE_LINKED",
				FuncName:     "delete_linked",
				SQL:          "DELETE FROM t WHERE id IN (/*SLICE:ids*/%s) OR ref_id IN (/*SLICE:ids*/%s)",
				Params: []model.QueryValue{
					{Name: "ids", Type: model.PyType{Type: "int", SQLType: "int", IsList: true, SqlcSliceName: "ids"}},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "None"}},
			},
			want: strings.Join([]string{
				"def delete_linked(conn: pymysql.Connection, ids: collections.abc.Sequence[int]) -> None:",
				`    sql = DELETE_LINKED.replace("/*SLICE:ids*/%s", ",".join(("%s",) * len(ids)) or "NULL")`,
				"    with conn.cursor() as cur:",
				"        cur.execute(sql, (*ids, *ids))",
				"",
			}, "\n"),
		},
		{
			// A plain placeholder between the reuse sites: arguments follow
			// the SQL text order, not the parameter order.
			name: "exec reused slice keeps text order around plain params",
			query: model.Query{
				Cmd:          metadata.CmdExec,
				ConstantName: "DELETE_BETWEEN",
				FuncName:     "delete_between",
				SQL:          "DELETE FROM t WHERE id IN (/*SLICE:ids*/%s) AND name = %s AND ref_id IN (/*SLICE:ids*/%s)",
				Params: []model.QueryValue{
					{Name: "ids", Type: model.PyType{Type: "int", SQLType: "int", IsList: true, SqlcSliceName: "ids"}},
					{Name: "name", Type: model.PyType{Type: "str", SQLType: "varchar"}},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "None"}},
			},
			want: strings.Join([]string{
				"def delete_between(conn: pymysql.Connection, ids: collections.abc.Sequence[int], name: str) -> None:",
				`    sql = DELETE_BETWEEN.replace("/*SLICE:ids*/%s", ",".join(("%s",) * len(ids)) or "NULL")`,
				"    with conn.cursor() as cur:",
				"        cur.execute(sql, (*ids, name, *ids))",
				"",
			}, "\n"),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			d := newMysql(t, false)
			conf := mysqlTestConfig(false)
			body := writer.NewCodeWriter(conf)
			d.WriteQueryFunc(body, conf, tc.query, 0)
			if got := body.String(); got != tc.want {
				t.Errorf("WriteQueryFunc() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestMysqlWriteQueryFuncAsync(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		query model.Query
		want  string
	}{
		{
			name: "exec awaits inside an async with block",
			query: model.Query{
				Cmd:          metadata.CmdExec,
				ConstantName: "TOUCH",
				FuncName:     "touch",
				Params: []model.QueryValue{
					{Name: "a", Type: model.PyType{Type: "int", SQLType: "int"}},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "None"}},
			},
			want: strings.Join([]string{
				"async def touch(conn: asyncmy.Connection, a: int) -> None:",
				"    async with conn.cursor() as cur:",
				"        await cur.execute(TOUCH, (a,))",
				"",
			}, "\n"),
		},
		{
			name: "execresult awaits execute and returns the cursor",
			query: model.Query{
				Cmd:          metadata.CmdExecResult,
				ConstantName: "RUN",
				FuncName:     "run",
				Returns:      model.QueryValue{Type: model.PyType{Type: "None"}},
			},
			want: strings.Join([]string{
				"async def run(conn: asyncmy.Connection) -> asyncmy.cursors.Cursor:",
				"    cur = conn.cursor()",
				"    await cur.execute(RUN)",
				"    return cur",
				"",
			}, "\n"),
		},
		{
			name: "execrows returns the awaited execute call",
			query: model.Query{
				Cmd:          metadata.CmdExecRows,
				ConstantName: "BUMP",
				FuncName:     "bump",
				Returns:      model.QueryValue{Type: model.PyType{Type: "int"}},
			},
			want: strings.Join([]string{
				"async def bump(conn: asyncmy.Connection) -> int:",
				"    async with conn.cursor() as cur:",
				"        return await cur.execute(BUMP)",
				"",
			}, "\n"),
		},
		{
			name: "execlastid awaits execute but not lastrowid",
			query: model.Query{
				Cmd:          metadata.CmdExecLastId,
				ConstantName: "ADD",
				FuncName:     "add",
				Params: []model.QueryValue{
					{Name: "name", Type: model.PyType{Type: "str", SQLType: "varchar"}},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "int", IsNullable: true}},
			},
			want: strings.Join([]string{
				"async def add(conn: asyncmy.Connection, name: str) -> int | None:",
				"    async with conn.cursor() as cur:",
				"        await cur.execute(ADD, (name,))",
				"        return cur.lastrowid",
				"",
			}, "\n"),
		},
		{
			name: "one struct awaits execute and fetchone",
			query: model.Query{
				Cmd:          metadata.CmdOne,
				ConstantName: "GET_USER",
				FuncName:     "get_user",
				Params: []model.QueryValue{
					{Name: "id_", Type: model.PyType{Type: "int", SQLType: "int"}},
				},
				Returns: mysqlUserReturn(),
			},
			want: strings.Join([]string{
				"async def get_user(conn: asyncmy.Connection, id_: int) -> models.User | None:",
				"    async with conn.cursor() as cur:",
				"        await cur.execute(GET_USER, (id_,))",
				"        row = await cur.fetchone()",
				"    if row is None:",
				"        return None",
				"    return " + mysqlUserDecode,
				"",
			}, "\n"),
		},
		{
			name: "many struct stays a plain def",
			query: model.Query{
				Cmd:          metadata.CmdMany,
				ConstantName: "LIST_USERS",
				FuncName:     "list_users",
				Returns:      mysqlUserReturn(),
			},
			want: strings.Join([]string{
				"def list_users(conn: asyncmy.Connection) -> QueryResults[models.User]:",
				"    def _decode_hook(row: tuple[typing.Any, ...]) -> models.User:",
				"        return " + mysqlUserDecode,
				"",
				"    return QueryResults(conn, LIST_USERS, _decode_hook)",
				"",
			}, "\n"),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			d := newMysql(t, true)
			conf := mysqlTestConfig(true)
			body := writer.NewCodeWriter(conf)
			d.WriteQueryFunc(body, conf, tc.query, 0)
			if got := body.String(); got != tc.want {
				t.Errorf("WriteQueryFunc() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestMysqlWriteQueryFuncClassesMode(t *testing.T) {
	t.Parallel()
	d := newMysql(t, false)
	conf := mysqlTestConfig(false)
	conf.EmitClasses = true
	body := writer.NewCodeWriter(conf)
	query := model.Query{
		Cmd:          metadata.CmdExec,
		ConstantName: "TOUCH",
		FuncName:     "touch",
		Params: []model.QueryValue{
			{Name: "a", Type: model.PyType{Type: "int", SQLType: "int"}},
		},
		Returns: model.QueryValue{Type: model.PyType{Type: "None"}},
	}
	d.WriteQueryFunc(body, conf, query, 1)
	want := strings.Join([]string{
		"    def touch(self, a: int) -> None:",
		"        with self._conn.cursor() as cur:",
		"            cur.execute(TOUCH, (a,))",
		"",
	}, "\n")
	if got := body.String(); got != want {
		t.Errorf("WriteQueryFunc() = %q, want %q", got, want)
	}
}

func TestMysqlWriteQueryFuncBundledParams(t *testing.T) {
	t.Parallel()
	d := newMysql(t, false)
	conf := mysqlTestConfig(false)
	body := writer.NewCodeWriter(conf)
	query := model.Query{
		Cmd:          metadata.CmdExec,
		ConstantName: "UPDATE_USER",
		FuncName:     "update_user",
		Params: []model.QueryValue{
			{
				EmitTable: true,
				Name:      "params",
				Type:      model.PyType{Type: "UpdateUserParams"},
				Table: &model.Table{
					Name: "UpdateUserParams",
					Columns: []model.Column{
						{Name: "name", DBName: "name", Type: model.PyType{Type: "str", SQLType: "varchar"}},
						{Name: "avatar", DBName: "avatar", Type: model.PyType{Type: "memoryview", SQLType: "blob"}},
						{Name: "id_", DBName: "id", Type: model.PyType{Type: "int", SQLType: "int"}},
					},
				},
			},
		},
		Returns: model.QueryValue{Type: model.PyType{Type: "None"}},
	}
	d.WriteQueryFunc(body, conf, query, 0)
	want := strings.Join([]string{
		"def update_user(conn: pymysql.Connection, params: UpdateUserParams) -> None:",
		"    with conn.cursor() as cur:",
		"        cur.execute(UPDATE_USER, (params.name, bytes(params.avatar), params.id_))",
		"",
	}, "\n")
	if got := body.String(); got != want {
		t.Errorf("WriteQueryFunc() = %q, want %q", got, want)
	}
}
