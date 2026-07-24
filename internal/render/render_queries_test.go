package render_test

import (
	"testing"

	"github.com/sqlc-dev/plugin-sdk-go/metadata"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func TestRenderQueriesModules(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		engine  string
		options string
		queries []*plugin.Query
		want    string
	}{
		{
			name:    "exec in functions mode",
			engine:  "postgresql",
			options: `{"package":"testpkg","sql_driver":"asyncpg","emit_init_file":false}`,
			queries: []*plugin.Query{{
				Name:     "InsertItem",
				Cmd:      metadata.CmdExec,
				Text:     "INSERT INTO test_items (id) VALUES ($1)",
				Filename: "queries.sql",
				Params:   []*plugin.Parameter{pgParam(pgColumn("id", "int4", true))},
			}},
			want: sqlcFileHeader("queries.sql") + `from __future__ import annotations

__all__: collections.abc.Sequence[str] = ("insert_item",)

import typing

if typing.TYPE_CHECKING:
    import asyncpg
    import collections.abc

    type ConnectionLike = asyncpg.Connection[asyncpg.Record] | asyncpg.pool.PoolConnectionProxy[asyncpg.Record]


INSERT_ITEM: typing.Final[str] = """-- name: InsertItem :exec
INSERT INTO test_items (id) VALUES ($1)
"""


async def insert_item(conn: ConnectionLike, *, id_: int) -> None:
    await conn.execute(INSERT_ITEM, id_)
`,
		},
		{
			name:    "psycopg rewrites placeholders and binds by name",
			engine:  "postgresql",
			options: `{"package":"testpkg","sql_driver":"psycopg_async","emit_init_file":false}`,
			queries: []*plugin.Query{{
				Name:     "InsertItem",
				Cmd:      metadata.CmdExec,
				Text:     "INSERT INTO test_items (id) VALUES ($1)",
				Filename: "queries.sql",
				Params:   []*plugin.Parameter{{Number: 1, Column: pgColumn("id", "int4", true)}},
			}},
			want: sqlcFileHeader("queries.sql") + `from __future__ import annotations

__all__: collections.abc.Sequence[str] = ("insert_item",)

import typing

if typing.TYPE_CHECKING:
    import collections.abc
    import psycopg
    import psycopg.rows

    type ConnectionLike = psycopg.AsyncConnection[psycopg.rows.TupleRow]


INSERT_ITEM: typing.Final[typing.LiteralString] = """-- name: InsertItem :exec
INSERT INTO test_items (id) VALUES (%(p1)s)
"""


async def insert_item(conn: ConnectionLike, *, id_: int) -> None:
    await conn.execute(INSERT_ITEM, {"p1": id_})
`,
		},
		{
			name:    "psycopg_sync emits synchronous bodies",
			engine:  "postgresql",
			options: `{"package":"testpkg","sql_driver":"psycopg_sync","emit_init_file":false}`,
			queries: []*plugin.Query{{
				Name:     "InsertItem",
				Cmd:      metadata.CmdExec,
				Text:     "INSERT INTO test_items (id) VALUES ($1)",
				Filename: "queries.sql",
				Params:   []*plugin.Parameter{{Number: 1, Column: pgColumn("id", "int4", true)}},
			}},
			want: sqlcFileHeader("queries.sql") + `from __future__ import annotations

__all__: collections.abc.Sequence[str] = ("insert_item",)

import typing

if typing.TYPE_CHECKING:
    import collections.abc
    import psycopg
    import psycopg.rows

    type ConnectionLike = psycopg.Connection[psycopg.rows.TupleRow]


INSERT_ITEM: typing.Final[typing.LiteralString] = """-- name: InsertItem :exec
INSERT INTO test_items (id) VALUES (%(p1)s)
"""


def insert_item(conn: ConnectionLike, *, id_: int) -> None:
    conn.execute(INSERT_ITEM, {"p1": id_})
`,
		},
		{
			name:    "exec in classes mode",
			engine:  "postgresql",
			options: `{"package":"testpkg","sql_driver":"asyncpg","emit_init_file":false,"emit_classes":true}`,
			queries: []*plugin.Query{{
				Name:     "InsertItem",
				Cmd:      metadata.CmdExec,
				Text:     "INSERT INTO test_items (id) VALUES ($1)",
				Filename: "queries.sql",
				Params:   []*plugin.Parameter{pgParam(pgColumn("id", "int4", true))},
			}},
			want: sqlcFileHeader("queries.sql") + `from __future__ import annotations

__all__: collections.abc.Sequence[str] = ("Queries",)

import typing

if typing.TYPE_CHECKING:
    import asyncpg
    import collections.abc

    type ConnectionLike = asyncpg.Connection[asyncpg.Record] | asyncpg.pool.PoolConnectionProxy[asyncpg.Record]


INSERT_ITEM: typing.Final[str] = """-- name: InsertItem :exec
INSERT INTO test_items (id) VALUES ($1)
"""


class Queries:
    __slots__ = ("_conn",)

    def __init__(self, conn: ConnectionLike) -> None:
        self._conn = conn

    @property
    def conn(self) -> ConnectionLike:
        return self._conn

    async def insert_item(self, *, id_: int) -> None:
        await self._conn.execute(INSERT_ITEM, id_)
`,
		},
		{
			name:    "many emits QueryResults in functions mode",
			engine:  "postgresql",
			options: `{"package":"testpkg","sql_driver":"asyncpg","emit_init_file":false}`,
			queries: []*plugin.Query{{
				Name:     "ListItemIds",
				Cmd:      metadata.CmdMany,
				Text:     "SELECT id FROM test_items",
				Filename: "queries.sql",
				Columns:  []*plugin.Column{pgColumn("id", "int4", true)},
			}},
			want: sqlcFileHeader("queries.sql") + `from __future__ import annotations

__all__: collections.abc.Sequence[str] = (
    "QueryResults",
    "list_item_ids",
)

import operator
import typing

if typing.TYPE_CHECKING:
    import asyncpg
    import asyncpg.cursor
    import collections.abc

    type QueryResultsArgsType = int | float | str | memoryview | collections.abc.Sequence[QueryResultsArgsType] | None

    type ConnectionLike = asyncpg.Connection[asyncpg.Record] | asyncpg.pool.PoolConnectionProxy[asyncpg.Record]


LIST_ITEM_IDS: typing.Final[str] = """-- name: ListItemIds :many
SELECT id FROM test_items
"""


class QueryResults[T]:
    __slots__ = ("_args", "_conn", "_cursor", "_decode_hook", "_iterator", "_sql")

    def __init__(
        self,
        conn: ConnectionLike,
        sql: str,
        decode_hook: collections.abc.Callable[[asyncpg.Record], T],
        *args: QueryResultsArgsType,
    ) -> None:
        self._conn = conn
        self._sql = sql
        self._decode_hook = decode_hook
        self._args = args
        self._cursor: asyncpg.cursor.CursorFactory[asyncpg.Record] | None = None
        self._iterator: asyncpg.cursor.CursorIterator[asyncpg.Record] | None = None

    def __aiter__(self) -> QueryResults[T]:
        return self

    def __await__(
        self,
    ) -> collections.abc.Generator[None, None, collections.abc.Sequence[T]]:
        async def _wrapper() -> collections.abc.Sequence[T]:
            result = await self._conn.fetch(self._sql, *self._args)
            return [self._decode_hook(row) for row in result]

        return _wrapper().__await__()

    async def __anext__(self) -> T:
        if self._cursor is None or self._iterator is None:
            self._cursor = self._conn.cursor(self._sql, *self._args)
            self._iterator = self._cursor.__aiter__()
        try:
            record = await self._iterator.__anext__()
        except StopAsyncIteration:
            self._cursor = None
            self._iterator = None
            raise
        return self._decode_hook(record)


def list_item_ids(conn: ConnectionLike) -> QueryResults[int]:
    return QueryResults(conn, LIST_ITEM_IDS, operator.itemgetter(0))
`,
		},
		{
			name:    "one with unmatched columns emits Row class",
			engine:  "postgresql",
			options: `{"package":"testpkg","sql_driver":"asyncpg","emit_init_file":false}`,
			queries: []*plugin.Query{{
				Name:     "GetItem",
				Cmd:      metadata.CmdOne,
				Text:     "SELECT id, name FROM test_items WHERE id = $1",
				Filename: "queries.sql",
				Columns:  []*plugin.Column{pgColumn("id", "int4", true), pgColumn("name", "text", false)},
				Params:   []*plugin.Parameter{pgParam(pgColumn("id", "int4", true))},
			}},
			want: sqlcFileHeader("queries.sql") + `from __future__ import annotations

__all__: collections.abc.Sequence[str] = (
    "GetItemRow",
    "get_item",
)

import dataclasses
import typing

if typing.TYPE_CHECKING:
    import asyncpg
    import collections.abc

    type ConnectionLike = asyncpg.Connection[asyncpg.Record] | asyncpg.pool.PoolConnectionProxy[asyncpg.Record]


@dataclasses.dataclass()
class GetItemRow:
    id_: int
    name: str | None


GET_ITEM: typing.Final[str] = """-- name: GetItem :one
SELECT id, name FROM test_items WHERE id = $1
"""


async def get_item(conn: ConnectionLike, *, id_: int) -> GetItemRow | None:
    row = await conn.fetchrow(GET_ITEM, id_)
    if row is None:
        return None
    return GetItemRow(id_=row[0], name=row[1])
`,
		},
		{
			name:    "query parameter limit emits Params class",
			engine:  "postgresql",
			options: `{"package":"testpkg","sql_driver":"asyncpg","emit_init_file":false,"query_parameter_limit":0}`,
			queries: []*plugin.Query{{
				Name:     "InsertItem",
				Cmd:      metadata.CmdExec,
				Text:     "INSERT INTO test_items (id, name) VALUES ($1, $2)",
				Filename: "queries.sql",
				Params: []*plugin.Parameter{
					pgParam(pgColumn("id", "int4", true)),
					pgParam(pgColumn("name", "text", true)),
				},
			}},
			want: sqlcFileHeader("queries.sql") + `from __future__ import annotations

__all__: collections.abc.Sequence[str] = (
    "InsertItemParams",
    "insert_item",
)

import dataclasses
import typing

if typing.TYPE_CHECKING:
    import asyncpg
    import collections.abc

    type ConnectionLike = asyncpg.Connection[asyncpg.Record] | asyncpg.pool.PoolConnectionProxy[asyncpg.Record]


@dataclasses.dataclass()
class InsertItemParams:
    id_: int
    name: str


INSERT_ITEM: typing.Final[str] = """-- name: InsertItem :exec
INSERT INTO test_items (id, name) VALUES ($1, $2)
"""


async def insert_item(conn: ConnectionLike, *, params: InsertItemParams) -> None:
    await conn.execute(INSERT_ITEM, params.id_, params.name)
`,
		},
		{
			name:    "sqlite3 emits conversion setup",
			engine:  "sqlite",
			options: `{"package":"testpkg","sql_driver":"sqlite3","emit_init_file":false}`,
			queries: []*plugin.Query{
				{
					Name:     "TouchItem",
					Cmd:      metadata.CmdExec,
					Text:     "UPDATE test_items SET created = ?",
					Filename: "queries.sql",
					Params:   []*plugin.Parameter{pgParam(pgColumn("created", "date", true))},
				},
				{
					Name:     "GetCreated",
					Cmd:      metadata.CmdOne,
					Text:     "SELECT created FROM test_items LIMIT 1",
					Filename: "queries.sql",
					Columns:  []*plugin.Column{pgColumn("created", "date", true)},
				},
			},
			want: sqlcFileHeader("queries.sql") + `from __future__ import annotations

__all__: collections.abc.Sequence[str] = (
    "get_created",
    "touch_item",
)

import datetime
import sqlite3
import typing

if typing.TYPE_CHECKING:
    import collections.abc


def _adapt_date(val: datetime.date) -> str:
    return val.isoformat()


def _convert_date(val: bytes) -> datetime.date:
    return datetime.date.fromisoformat(val.decode())


sqlite3.register_adapter(datetime.date, _adapt_date)

sqlite3.register_converter("date", _convert_date)


TOUCH_ITEM: typing.Final[str] = """-- name: TouchItem :exec
UPDATE test_items SET created = ?
"""

GET_CREATED: typing.Final[str] = """-- name: GetCreated :one
SELECT created FROM test_items LIMIT 1
"""


def touch_item(conn: sqlite3.Connection, *, created: datetime.date) -> None:
    conn.execute(TOUCH_ITEM, (created,))


def get_created(conn: sqlite3.Connection) -> datetime.date | None:
    row = conn.execute(GET_CREATED).fetchone()
    if row is None:
        return None
    return row[0]
`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := newRenderRequest(tc.engine, tc.options, nil, tc.queries)

			got := renderedFile(t, mustRenderFiles(t, req), "queries.py")
			if got != tc.want {
				t.Errorf("queries.py mismatch\ngot:\n%q\nwant:\n%q", got, tc.want)
			}
		})
	}
}

func TestRenderQueriesManyClassesMode(t *testing.T) {
	t.Parallel()
	req := newRenderRequest(
		"postgresql",
		`{"package":"testpkg","sql_driver":"asyncpg","emit_init_file":false,"emit_classes":true}`,
		nil,
		[]*plugin.Query{{
			Name:     "ListItemIds",
			Cmd:      metadata.CmdMany,
			Text:     "SELECT id FROM test_items",
			Filename: "queries.sql",
			Columns:  []*plugin.Column{pgColumn("id", "int4", true)},
		}},
	)

	got := renderedFile(t, mustRenderFiles(t, req), "queries.py")
	// QueryResults is emitted once before the Querier class and the query
	// method lives on the class, with no module-level function duplication.
	want := sqlcFileHeader("queries.sql") + `from __future__ import annotations

__all__: collections.abc.Sequence[str] = (
    "Queries",
    "QueryResults",
)

import operator
import typing

if typing.TYPE_CHECKING:
    import asyncpg
    import asyncpg.cursor
    import collections.abc

    type QueryResultsArgsType = int | float | str | memoryview | collections.abc.Sequence[QueryResultsArgsType] | None

    type ConnectionLike = asyncpg.Connection[asyncpg.Record] | asyncpg.pool.PoolConnectionProxy[asyncpg.Record]


LIST_ITEM_IDS: typing.Final[str] = """-- name: ListItemIds :many
SELECT id FROM test_items
"""


class QueryResults[T]:
    __slots__ = ("_args", "_conn", "_cursor", "_decode_hook", "_iterator", "_sql")

    def __init__(
        self,
        conn: ConnectionLike,
        sql: str,
        decode_hook: collections.abc.Callable[[asyncpg.Record], T],
        *args: QueryResultsArgsType,
    ) -> None:
        self._conn = conn
        self._sql = sql
        self._decode_hook = decode_hook
        self._args = args
        self._cursor: asyncpg.cursor.CursorFactory[asyncpg.Record] | None = None
        self._iterator: asyncpg.cursor.CursorIterator[asyncpg.Record] | None = None

    def __aiter__(self) -> QueryResults[T]:
        return self

    def __await__(
        self,
    ) -> collections.abc.Generator[None, None, collections.abc.Sequence[T]]:
        async def _wrapper() -> collections.abc.Sequence[T]:
            result = await self._conn.fetch(self._sql, *self._args)
            return [self._decode_hook(row) for row in result]

        return _wrapper().__await__()

    async def __anext__(self) -> T:
        if self._cursor is None or self._iterator is None:
            self._cursor = self._conn.cursor(self._sql, *self._args)
            self._iterator = self._cursor.__aiter__()
        try:
            record = await self._iterator.__anext__()
        except StopAsyncIteration:
            self._cursor = None
            self._iterator = None
            raise
        return self._decode_hook(record)


class Queries:
    __slots__ = ("_conn",)

    def __init__(self, conn: ConnectionLike) -> None:
        self._conn = conn

    @property
    def conn(self) -> ConnectionLike:
        return self._conn

    def list_item_ids(self) -> QueryResults[int]:
        return QueryResults(self._conn, LIST_ITEM_IDS, operator.itemgetter(0))
`
	if got != want {
		t.Errorf("queries.py mismatch\ngot:\n%q\nwant:\n%q", got, want)
	}
}
