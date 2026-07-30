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

func tursoTestConfig(async bool) *config.Config {
	sqlDriver := config.SQLDriverTursoSync
	if async {
		sqlDriver = config.SQLDriverTursoAsync
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

func newTurso(t *testing.T, async bool) driver.Driver {
	t.Helper()
	d, err := driver.New(tursoTestConfig(async))
	if err != nil {
		t.Fatalf("driver.New() error = %v", err)
	}

	return d
}

func tursoItemReturn() model.QueryValue {
	return model.QueryValue{
		Table: &model.Table{
			Name: "Item",
			Columns: []model.Column{
				{Name: "id_", Type: model.PyType{Type: "int", SQLType: "integer"}},
				{Name: "day", Type: model.PyType{Type: "datetime.date", SQLType: "date"}},
				{Name: "ts", Type: model.PyType{Type: "datetime.datetime", SQLType: "datetime"}},
				{Name: "amount", Type: model.PyType{Type: "decimal.Decimal", SQLType: "decimal(10,5)"}},
				{Name: "flag", Type: model.PyType{Type: "bool", SQLType: "boolean", IsNullable: true}},
				{Name: "data", Type: model.PyType{Type: "memoryview", SQLType: "blob"}},
			},
		},
		Type: model.PyType{Type: "models.Item"},
	}
}

func TestTursoDriverMetadata(t *testing.T) {
	t.Parallel()
	sync := newTurso(t, false)
	if got := sync.Name(); got != "turso" {
		t.Errorf("sync Name() = %q, want %q", got, "turso")
	}
	if got := sync.ConnType(); got != "turso.Connection" {
		t.Errorf("sync ConnType() = %q, want %q", got, "turso.Connection")
	}
	if sync.IsAsync() {
		t.Error("sync IsAsync() = true, want false")
	}
	if got := sync.TypeCheckingHook(); got != nil {
		t.Errorf("sync TypeCheckingHook() = %v, want nil", got)
	}

	async := newTurso(t, true)
	if got := async.Name(); got != "turso.aio" {
		t.Errorf("async Name() = %q, want %q", got, "turso.aio")
	}
	if got := async.ConnType(); got != "turso.aio.Connection" {
		t.Errorf("async ConnType() = %q, want %q", got, "turso.aio.Connection")
	}
	if !async.IsAsync() {
		t.Error("async IsAsync() = false, want true")
	}
}

func TestTursoConversions(t *testing.T) {
	t.Parallel()
	d := newTurso(t, false)
	// Everything converts inline: the declared-type set matches the sqlite
	// drivers, but with no registration mechanism both checks agree.
	for sqlType, want := range map[string]bool{
		"date": true, "datetime": true, "timestamp": true, "decimal": true,
		"decimal(10,5)": true, "bool": true, "boolean": true, "blob": true,
		"text": false, "integer": false, "real": false,
	} {
		if got := d.NeedsConversion(sqlType); got != want {
			t.Errorf("NeedsConversion(%q) = %v, want %v", sqlType, got, want)
		}
		if got := d.ConvertsInline(sqlType); got != want {
			t.Errorf("ConvertsInline(%q) = %v, want %v", sqlType, got, want)
		}
	}
}

func TestTursoSupportsCommand(t *testing.T) {
	t.Parallel()
	for _, async := range []bool{false, true} {
		d := newTurso(t, async)
		for cmd, want := range map[string]bool{
			metadata.CmdExec:       true,
			metadata.CmdExecResult: true,
			metadata.CmdExecRows:   true,
			metadata.CmdExecLastId: true,
			metadata.CmdOne:        true,
			metadata.CmdMany:       true,
			metadata.CmdCopyFrom:   false,
			":batchexec":           false,
		} {
			if got := d.SupportsCommand(cmd); got != want {
				t.Errorf("async=%v SupportsCommand(%q) = %v, want %v", async, cmd, got, want)
			}
		}
	}
}

func TestTursoWriteConversionSetup(t *testing.T) {
	t.Parallel()
	d := newTurso(t, false)
	conf := tursoTestConfig(false)
	w := writer.NewCodeWriter(conf)
	queries := []model.Query{{Returns: model.QueryValue{Type: model.PyType{Type: "datetime.date", SQLType: "date"}}}}
	if d.WriteConversionSetup(w, conf, queries) {
		t.Error("WriteConversionSetup() = true, want false: turso converts inline")
	}
	if got := w.String(); got != "" {
		t.Errorf("WriteConversionSetup() wrote %q, want nothing", got)
	}
}

func TestTursoWriteQueryResultsClassSync(t *testing.T) {
	t.Parallel()
	d := newTurso(t, false)
	w := writer.NewCodeWriter(tursoTestConfig(false))
	if got := d.WriteQueryResultsClass(w); got != "QueryResults" {
		t.Errorf("WriteQueryResultsClass() = %q, want %q", got, "QueryResults")
	}
	want := strings.Join([]string{
		"class QueryResults[T]:",
		`    __slots__ = ("_args", "_conn", "_cursor", "_decode_hook", "_sql")`,
		"",
		"    def __init__(",
		"        self,",
		"        conn: turso.Connection,",
		"        sql: str,",
		"        decode_hook: collections.abc.Callable[[tuple[typing.Any, ...]], T],",
		"        *args: QueryResultsArgsType,",
		"    ) -> None:",
		"        self._conn = conn",
		"        self._sql = sql",
		"        self._decode_hook = decode_hook",
		"        self._args = args",
		"        self._cursor: turso.Cursor | None = None",
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
		"        if self._cursor is None:",
		"            self._cursor = self._conn.execute(self._sql, self._args)",
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

func TestTursoWriteQueryResultsClassAsync(t *testing.T) {
	t.Parallel()
	d := newTurso(t, true)
	w := writer.NewCodeWriter(tursoTestConfig(true))
	if got := d.WriteQueryResultsClass(w); got != "QueryResults" {
		t.Errorf("WriteQueryResultsClass() = %q, want %q", got, "QueryResults")
	}
	want := strings.Join([]string{
		"class QueryResults[T]:",
		`    __slots__ = ("_args", "_conn", "_cursor", "_decode_hook", "_sql")`,
		"",
		"    def __init__(",
		"        self,",
		"        conn: turso.aio.Connection,",
		"        sql: str,",
		"        decode_hook: collections.abc.Callable[[tuple[typing.Any, ...]], T],",
		"        *args: QueryResultsArgsType,",
		"    ) -> None:",
		"        self._conn = conn",
		"        self._sql = sql",
		"        self._decode_hook = decode_hook",
		"        self._args = args",
		"        self._cursor: turso.aio.Cursor | None = None",
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
		"        if self._cursor is None:",
		"            self._cursor = await self._conn.execute(self._sql, self._args)",
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

func TestTursoWriteQueryFuncSync(t *testing.T) {
	t.Parallel()
	longName := strings.Repeat("p", 340)
	cases := []struct {
		name  string
		query model.Query
		want  string
	}{
		{
			name: "exec converts date and blob params to wire types",
			query: model.Query{
				Cmd:          metadata.CmdExec,
				ConstantName: "TOUCH",
				FuncName:     "touch",
				Params: []model.QueryValue{
					{Name: "day", Type: model.PyType{Type: "datetime.date", SQLType: "date"}},
					{Name: "data", Type: model.PyType{Type: "memoryview", SQLType: "blob"}},
					{Name: "flag", Type: model.PyType{Type: "bool", SQLType: "bool"}},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "None"}},
			},
			want: strings.Join([]string{
				"def touch(conn: turso.Connection, day: datetime.date, data: memoryview, flag: bool) -> None:",
				"    conn.execute(TOUCH, (day.isoformat(), bytes(data), flag))",
				"",
			}, "\n"),
		},
		{
			name: "exec nullable decimal param guards the wire conversion",
			query: model.Query{
				Cmd:          metadata.CmdExec,
				ConstantName: "SET_AMOUNT",
				FuncName:     "set_amount",
				Params: []model.QueryValue{
					{Name: "amount", Type: model.PyType{Type: "decimal.Decimal", SQLType: "decimal(10,5)", IsNullable: true}},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "None"}},
			},
			want: strings.Join([]string{
				"def set_amount(conn: turso.Connection, amount: decimal.Decimal | None) -> None:",
				"    conn.execute(SET_AMOUNT, (str(amount) if amount is not None else None,))",
				"",
			}, "\n"),
		},
		{
			name: "exec overridden date param composes override and wire conversion",
			query: model.Query{
				Cmd:          metadata.CmdExec,
				ConstantName: "SET_DAY",
				FuncName:     "set_day",
				Params: []model.QueryValue{
					{Name: "day", Type: model.PyType{
						Type:        "str",
						SQLType:     "date",
						IsOverride:  true,
						DefaultType: "datetime.date",
					}},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "None"}},
			},
			want: strings.Join([]string{
				"def set_day(conn: turso.Connection, day: str) -> None:",
				"    conn.execute(SET_DAY, (datetime.date(day).isoformat(),))",
				"",
			}, "\n"),
		},
		{
			name: "exec long params hoist into sql_args",
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
				"    conn: turso.Connection,",
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
			name: "execresult returns the cursor",
			query: model.Query{
				Cmd:          metadata.CmdExecResult,
				ConstantName: "RUN",
				FuncName:     "run",
				Returns:      model.QueryValue{Type: model.PyType{Type: "None"}},
			},
			want: strings.Join([]string{
				"def run(conn: turso.Connection) -> turso.Cursor:",
				"    return conn.execute(RUN)",
				"",
			}, "\n"),
		},
		{
			name: "execrows returns rowcount",
			query: model.Query{
				Cmd:          metadata.CmdExecRows,
				ConstantName: "BUMP",
				FuncName:     "bump",
				Returns:      model.QueryValue{Type: model.PyType{Type: "int"}},
			},
			want: strings.Join([]string{
				"def bump(conn: turso.Connection) -> int:",
				"    return conn.execute(BUMP).rowcount",
				"",
			}, "\n"),
		},
		{
			name: "execlastid returns lastrowid",
			query: model.Query{
				Cmd:          metadata.CmdExecLastId,
				ConstantName: "ADD",
				FuncName:     "add",
				Returns:      model.QueryValue{Type: model.PyType{Type: "int", IsNullable: true}},
			},
			want: strings.Join([]string{
				"def add(conn: turso.Connection) -> int | None:",
				"    return conn.execute(ADD).lastrowid",
				"",
			}, "\n"),
		},
		{
			name: "one struct return decodes every convertible column",
			query: model.Query{
				Cmd:          metadata.CmdOne,
				ConstantName: "GET_ITEM",
				FuncName:     "get_item",
				Params: []model.QueryValue{
					{Name: "id_", Type: model.PyType{Type: "int", SQLType: "integer"}},
				},
				Returns: tursoItemReturn(),
			},
			want: strings.Join([]string{
				"def get_item(conn: turso.Connection, id_: int) -> models.Item | None:",
				"    row = conn.execute(GET_ITEM, (id_,)).fetchone()",
				"    if row is None:",
				"        return None",
				"    return models.Item(id_=row[0], day=datetime.date.fromisoformat(row[1]), ts=datetime.datetime.fromisoformat(row[2]), amount=decimal.Decimal(str(row[3])), flag=bool(row[4]) if row[4] is not None else None, data=memoryview(row[5]))",
				"",
			}, "\n"),
		},
		{
			name: "one scalar without conversion returns the raw value",
			query: model.Query{
				Cmd:          metadata.CmdOne,
				ConstantName: "COUNT",
				FuncName:     "count",
				Returns:      model.QueryValue{Type: model.PyType{Type: "int", SQLType: "integer"}},
			},
			want: strings.Join([]string{
				"def count(conn: turso.Connection) -> int | None:",
				"    row = conn.execute(COUNT).fetchone()",
				"    if row is None:",
				"        return None",
				"    return row[0]",
				"",
			}, "\n"),
		},
		{
			name: "one slice query expands the placeholder at runtime",
			query: model.Query{
				Cmd:          metadata.CmdOne,
				ConstantName: "GET_BY_IDS",
				FuncName:     "get_by_ids",
				SQL:          "SELECT count(*) FROM t WHERE id IN (/*SLICE:ids*/?)",
				Params: []model.QueryValue{
					{Name: "ids", Type: model.PyType{Type: "int", SQLType: "integer", IsList: true, SqlcSliceName: "ids"}},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "int", SQLType: "integer"}},
			},
			want: strings.Join([]string{
				"def get_by_ids(conn: turso.Connection, ids: collections.abc.Sequence[int]) -> int | None:",
				`    sql = GET_BY_IDS.replace("/*SLICE:ids*/?", ",".join("?" * len(ids)) or "NULL", 1)`,
				"    row = conn.execute(sql, (*ids,)).fetchone()",
				"    if row is None:",
				"        return None",
				"    return row[0]",
				"",
			}, "\n"),
		},
		{
			name: "one slice of dates wire-converts element-wise before unpacking",
			query: model.Query{
				Cmd:          metadata.CmdOne,
				ConstantName: "GET_BY_DAYS",
				FuncName:     "get_by_days",
				SQL:          "SELECT count(*) FROM t WHERE d IN (/*SLICE:days*/?)",
				Params: []model.QueryValue{
					{
						Name: "days",
						Type: model.PyType{Type: "datetime.date", SQLType: "date", IsList: true, SqlcSliceName: "days"},
					},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "int", SQLType: "integer"}},
			},
			want: strings.Join([]string{
				"def get_by_days(conn: turso.Connection, days: collections.abc.Sequence[datetime.date]) -> int | None:",
				`    sql = GET_BY_DAYS.replace("/*SLICE:days*/?", ",".join("?" * len(days)) or "NULL", 1)`,
				"    row = conn.execute(sql, (*[v.isoformat() for v in days],)).fetchone()",
				"    if row is None:",
				"        return None",
				"    return row[0]",
				"",
			}, "\n"),
		},
		{
			name: "many scalar without conversion uses itemgetter",
			query: model.Query{
				Cmd:          metadata.CmdMany,
				ConstantName: "LIST_IDS",
				FuncName:     "list_ids",
				Returns:      model.QueryValue{Type: model.PyType{Type: "int", SQLType: "integer"}},
			},
			want: strings.Join([]string{
				"def list_ids(conn: turso.Connection) -> QueryResults[int]:",
				"    return QueryResults(conn, LIST_IDS, operator.itemgetter(0))",
				"",
			}, "\n"),
		},
		{
			name: "many scalar date return needs a decode hook",
			query: model.Query{
				Cmd:          metadata.CmdMany,
				ConstantName: "LIST_DAYS",
				FuncName:     "list_days",
				Returns:      model.QueryValue{Type: model.PyType{Type: "datetime.date", SQLType: "date"}},
			},
			want: strings.Join([]string{
				"def list_days(conn: turso.Connection) -> QueryResults[datetime.date]:",
				"    def _decode_hook(row: tuple[typing.Any, ...]) -> datetime.date:",
				"        return datetime.date.fromisoformat(row[0])",
				"",
				"    return QueryResults(conn, LIST_DAYS, _decode_hook)",
				"",
			}, "\n"),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			d := newTurso(t, false)
			conf := tursoTestConfig(false)
			body := writer.NewCodeWriter(conf)
			d.WriteQueryFunc(body, conf, tc.query, 0)
			if got := body.String(); got != tc.want {
				t.Errorf("WriteQueryFunc() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestTursoWriteQueryFuncAsync(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		query model.Query
		want  string
	}{
		{
			name: "exec awaits the execute call",
			query: model.Query{
				Cmd:          metadata.CmdExec,
				ConstantName: "TOUCH",
				FuncName:     "touch",
				Params: []model.QueryValue{
					{Name: "day", Type: model.PyType{Type: "datetime.date", SQLType: "date"}},
				},
				Returns: model.QueryValue{Type: model.PyType{Type: "None"}},
			},
			want: strings.Join([]string{
				"async def touch(conn: turso.aio.Connection, day: datetime.date) -> None:",
				"    await conn.execute(TOUCH, (day.isoformat(),))",
				"",
			}, "\n"),
		},
		{
			name: "execresult returns the awaited cursor",
			query: model.Query{
				Cmd:          metadata.CmdExecResult,
				ConstantName: "RUN",
				FuncName:     "run",
				Returns:      model.QueryValue{Type: model.PyType{Type: "None"}},
			},
			want: strings.Join([]string{
				"async def run(conn: turso.aio.Connection) -> turso.aio.Cursor:",
				"    return await conn.execute(RUN)",
				"",
			}, "\n"),
		},
		{
			name: "execrows parenthesizes the awaited cursor",
			query: model.Query{
				Cmd:          metadata.CmdExecRows,
				ConstantName: "BUMP",
				FuncName:     "bump",
				Returns:      model.QueryValue{Type: model.PyType{Type: "int"}},
			},
			want: strings.Join([]string{
				"async def bump(conn: turso.aio.Connection) -> int:",
				"    return (await conn.execute(BUMP)).rowcount",
				"",
			}, "\n"),
		},
		{
			name: "execlastid parenthesizes the awaited cursor",
			query: model.Query{
				Cmd:          metadata.CmdExecLastId,
				ConstantName: "ADD",
				FuncName:     "add",
				Returns:      model.QueryValue{Type: model.PyType{Type: "int", IsNullable: true}},
			},
			want: strings.Join([]string{
				"async def add(conn: turso.aio.Connection) -> int | None:",
				"    return (await conn.execute(ADD)).lastrowid",
				"",
			}, "\n"),
		},
		{
			name: "one struct return awaits execute and fetchone",
			query: model.Query{
				Cmd:          metadata.CmdOne,
				ConstantName: "GET_ITEM",
				FuncName:     "get_item",
				Params: []model.QueryValue{
					{Name: "id_", Type: model.PyType{Type: "int", SQLType: "integer"}},
				},
				Returns: tursoItemReturn(),
			},
			want: strings.Join([]string{
				"async def get_item(conn: turso.aio.Connection, id_: int) -> models.Item | None:",
				"    row = await (await conn.execute(GET_ITEM, (id_,))).fetchone()",
				"    if row is None:",
				"        return None",
				"    return models.Item(id_=row[0], day=datetime.date.fromisoformat(row[1]), ts=datetime.datetime.fromisoformat(row[2]), amount=decimal.Decimal(str(row[3])), flag=bool(row[4]) if row[4] is not None else None, data=memoryview(row[5]))",
				"",
			}, "\n"),
		},
		{
			name: "many struct return stays a plain def",
			query: model.Query{
				Cmd:          metadata.CmdMany,
				ConstantName: "LIST_ITEMS",
				FuncName:     "list_items",
				Returns:      tursoItemReturn(),
			},
			want: strings.Join([]string{
				"def list_items(conn: turso.aio.Connection) -> QueryResults[models.Item]:",
				"    def _decode_hook(row: tuple[typing.Any, ...]) -> models.Item:",
				"        return models.Item(id_=row[0], day=datetime.date.fromisoformat(row[1]), ts=datetime.datetime.fromisoformat(row[2]), amount=decimal.Decimal(str(row[3])), flag=bool(row[4]) if row[4] is not None else None, data=memoryview(row[5]))",
				"",
				"    return QueryResults(conn, LIST_ITEMS, _decode_hook)",
				"",
			}, "\n"),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			d := newTurso(t, true)
			conf := tursoTestConfig(true)
			body := writer.NewCodeWriter(conf)
			d.WriteQueryFunc(body, conf, tc.query, 0)
			if got := body.String(); got != tc.want {
				t.Errorf("WriteQueryFunc() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestTursoWriteQueryFuncBundledParams(t *testing.T) {
	t.Parallel()
	d := newTurso(t, false)
	conf := tursoTestConfig(false)
	body := writer.NewCodeWriter(conf)
	query := model.Query{
		Cmd:          metadata.CmdExec,
		ConstantName: "UPDATE_ITEM",
		FuncName:     "update_item",
		Params: []model.QueryValue{
			{
				EmitTable: true,
				Name:      "params",
				Type:      model.PyType{Type: "UpdateItemParams"},
				Table: &model.Table{
					Name: "UpdateItemParams",
					Columns: []model.Column{
						{Name: "day", DBName: "day", Type: model.PyType{Type: "datetime.date", SQLType: "date"}},
						{Name: "id_", DBName: "id", Type: model.PyType{Type: "int", SQLType: "integer"}},
					},
				},
			},
		},
		Returns: model.QueryValue{Type: model.PyType{Type: "None"}},
	}
	d.WriteQueryFunc(body, conf, query, 0)
	want := strings.Join([]string{
		"def update_item(conn: turso.Connection, params: UpdateItemParams) -> None:",
		"    conn.execute(UPDATE_ITEM, (params.day.isoformat(), params.id_))",
		"",
	}, "\n")
	if got := body.String(); got != want {
		t.Errorf("WriteQueryFunc() = %q, want %q", got, want)
	}
}

func TestTursoSpeedupsHelpers(t *testing.T) {
	t.Parallel()
	for sqlType, want := range map[string]bool{
		"date": true, "datetime": true, "timestamp": true,
		"decimal": false, "decimal(10,5)": false, "bool": false, "blob": false, "text": false,
	} {
		if got := driver.TursoSpeedupsDecodes(sqlType); got != want {
			t.Errorf("TursoSpeedupsDecodes(%q) = %v, want %v", sqlType, got, want)
		}
	}

	dateReturn := model.QueryValue{Type: model.PyType{Type: "datetime.date", SQLType: "date"}}
	overriddenDate := model.QueryValue{Type: model.PyType{
		Type: "str", SQLType: "date", IsOverride: true, DefaultType: "datetime.date",
	}}
	cases := []struct {
		name    string
		queries []model.Query
		want    bool
	}{
		{name: "date scalar return", queries: []model.Query{{Returns: dateReturn}}, want: true},
		{name: "date only as param", queries: []model.Query{{Params: []model.QueryValue{dateReturn}}}, want: false},
		{name: "overridden date return decodes via the override", queries: []model.Query{{Returns: overriddenDate}}, want: false},
		{name: "decimal return has no speedups variant", queries: []model.Query{{
			Returns: model.QueryValue{Type: model.PyType{Type: "decimal.Decimal", SQLType: "decimal"}},
		}}, want: false},
		{
			name: "struct return with a top-level date column",
			queries: []model.Query{{Returns: model.QueryValue{
				Table: &model.Table{Name: "Row", Columns: []model.Column{
					{Name: "id_", Type: model.PyType{Type: "int", SQLType: "integer"}},
					{Name: "day", Type: model.PyType{Type: "datetime.date", SQLType: "date"}},
				}},
				Type: model.PyType{Type: "models.Row"},
			}}},
			want: true,
		},
		{
			name: "embed without speedups columns",
			queries: []model.Query{{Returns: model.QueryValue{
				Table: &model.Table{Name: "Row", Columns: []model.Column{
					{Name: "inner", Embed: &model.Embed{Columns: []model.Column{
						{Name: "id_", Type: model.PyType{Type: "int", SQLType: "integer"}},
					}}},
					{Name: "name", Type: model.PyType{Type: "str", SQLType: "text"}},
				}},
				Type: model.PyType{Type: "models.Row"},
			}}},
			want: false,
		},
		{
			name: "struct return without speedups columns",
			queries: []model.Query{{Returns: model.QueryValue{
				Table: &model.Table{Name: "Row", Columns: []model.Column{
					{Name: "id_", Type: model.PyType{Type: "int", SQLType: "integer"}},
					{Name: "name", Type: model.PyType{Type: "str", SQLType: "text"}},
				}},
				Type: model.PyType{Type: "models.Row"},
			}}},
			want: false,
		},
		{
			name: "embedded timestamp column",
			queries: []model.Query{{Returns: model.QueryValue{
				Table: &model.Table{Name: "Row", Columns: []model.Column{
					{Name: "outer", Type: model.PyType{Type: "int", SQLType: "integer"}},
					{Name: "inner", Embed: &model.Embed{Columns: []model.Column{
						{Name: "ts", Type: model.PyType{Type: "datetime.datetime", SQLType: "timestamp"}},
					}}},
				}},
				Type: model.PyType{Type: "models.Row"},
			}}},
			want: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := driver.TursoSpeedupsUsed(tc.queries); got != tc.want {
				t.Errorf("TursoSpeedupsUsed() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestTursoSpeedupsWriteQueryFunc(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		query model.Query
		want  string
	}{
		{
			name: "one date return decodes via ciso8601",
			query: model.Query{
				Cmd:          metadata.CmdOne,
				ConstantName: "GET_DAY",
				FuncName:     "get_day",
				Returns:      model.QueryValue{Type: model.PyType{Type: "datetime.date", SQLType: "date"}},
			},
			want: strings.Join([]string{
				"def get_day(conn: turso.Connection) -> datetime.date | None:",
				"    row = conn.execute(GET_DAY).fetchone()",
				"    if row is None:",
				"        return None",
				"    return ciso8601.parse_datetime(row[0]).date()",
				"",
			}, "\n"),
		},
		{
			name: "struct return mixes ciso8601 and unchanged decodes",
			query: model.Query{
				Cmd:          metadata.CmdOne,
				ConstantName: "GET_ITEM",
				FuncName:     "get_item",
				Returns:      tursoItemReturn(),
			},
			want: strings.Join([]string{
				"def get_item(conn: turso.Connection) -> models.Item | None:",
				"    row = conn.execute(GET_ITEM).fetchone()",
				"    if row is None:",
				"        return None",
				"    return models.Item(id_=row[0], day=ciso8601.parse_datetime(row[1]).date(), ts=ciso8601.parse_datetime(row[2]), amount=decimal.Decimal(str(row[3])), flag=bool(row[4]) if row[4] is not None else None, data=memoryview(row[5]))",
				"",
			}, "\n"),
		},
		{
			name: "many timestamp return hooks through ciso8601",
			query: model.Query{
				Cmd:          metadata.CmdMany,
				ConstantName: "LIST_TS",
				FuncName:     "list_ts",
				Returns:      model.QueryValue{Type: model.PyType{Type: "datetime.datetime", SQLType: "timestamp"}},
			},
			want: strings.Join([]string{
				"def list_ts(conn: turso.Connection) -> QueryResults[datetime.datetime]:",
				"    def _decode_hook(row: tuple[typing.Any, ...]) -> datetime.datetime:",
				"        return ciso8601.parse_datetime(row[0])",
				"",
				"    return QueryResults(conn, LIST_TS, _decode_hook)",
				"",
			}, "\n"),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			conf := tursoTestConfig(false)
			conf.Speedups = true
			d, err := driver.New(conf)
			if err != nil {
				t.Fatalf("driver.New() error = %v", err)
			}
			body := writer.NewCodeWriter(conf)
			d.WriteQueryFunc(body, conf, tc.query, 0)
			if got := body.String(); got != tc.want {
				t.Errorf("WriteQueryFunc() = %q, want %q", got, tc.want)
			}
		})
	}
}
