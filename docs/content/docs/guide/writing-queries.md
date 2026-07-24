---
title: Writing queries
weight: 40
prev: /docs/guide/model-types
next: /docs/guide/enums
---

Each query `.sql` file becomes one Python module, and each annotated query in it
becomes one typed function. This page shows how the `-- name:` annotation maps to
a function and what each query command generates.

## Anatomy of a query

A query is a `-- name: <Name> :<command>` comment followed by SQL:

```sql
-- name: GetFieldNaming :one
SELECT id, outputs
FROM test_field_namings
WHERE id = $1 LIMIT 1;
```

From that the plugin generates:

```python
GET_FIELD_NAMING: typing.Final[str] = """-- name: GetFieldNaming :one
SELECT id, outputs
FROM test_field_namings
WHERE id = $1 LIMIT 1
"""


async def get_field_naming(conn: ConnectionLike, *, id_: int) -> models.TestFieldNaming | None:
    row = await conn.fetchrow(GET_FIELD_NAMING, id_)
    if row is None:
        return None
    return models.TestFieldNaming(id_=row[0], outputs=row[1])
```

- The SQL is stored once as a module-level constant.
- `GetFieldNaming` becomes the function `get_field_naming`.
- The connection is the first positional argument (`conn`); its type
  `ConnectionLike` is a generated alias for the driver's connection (or pool).
- Query parameters (`$1`) become keyword-only arguments (`id_`).
- The `:one` command sets the return type and body.

## Query commands

### `:one`

Returns the row or `None`. The row is a `models.*` class when the query's columns
match a table, a generated `<Name>Row` class when they do not, or a bare scalar
when the query selects a single column. Shown above.

When a query's columns do not match one table exactly (a join, a partial select,
an aggregate), the plugin generates a dedicated `<Name>Row` class instead of
reusing a table model:

```python
class GetJoinedFieldNamingsRow(msgspec.Struct):
    outputs: str
    outputs_2: str


async def get_joined_field_namings(conn: ConnectionLike, *, id_: int) -> GetJoinedFieldNamingsRow | None:
    row = await conn.fetchrow(GET_JOINED_FIELD_NAMINGS, id_)
    if row is None:
        return None
    return GetJoinedFieldNamingsRow(outputs=row[0], outputs_2=row[1])
```

### `:many`

Returns a `QueryResults[T]` - a helper that supports both async iteration and
one-shot fetching, so you do not pay for materializing the whole result set
unless you want it:

```python
def get_many_test_timestamp_postgres_type(conn: ConnectionLike, *, id_: int) -> QueryResults[datetime.datetime]:
    return QueryResults(conn, GET_MANY_TEST_TIMESTAMP_POSTGRES_TYPE, operator.itemgetter(0), id_)
```

Here the query selects one column, so `T` is a scalar (`datetime.datetime`). A
`:many` that selects full rows returns `QueryResults[models.Foo]` (or a
`<Name>Row`).

{{< callout type="info" >}}
  Note that a `:many` function is **not** a coroutine, even on the async drivers -
  it returns the `QueryResults` helper synchronously. You await (or iterate) the
  helper, not the call.
{{< /callout >}}

### `:exec`

Runs the statement and returns `None`:

```python
async def set_field_naming_outputs(conn: ConnectionLike, *, id_: int, outputs: str, outputs_2: str) -> None:
    await conn.execute(SET_FIELD_NAMING_OUTPUTS, id_, outputs, outputs_2)
```

### `:execresult`, `:execrows`, `:execlastid`

Variants of `:exec` that return something about the write:

- **`:execrows`** - the number of affected rows (`int`). For statements that
  affect no rows, such as `CREATE TABLE`, asyncpg reports `0` while psycopg and
  the SQLite drivers report `-1`.
- **`:execlastid`** - the cursor's `lastrowid`, typed `int | None` - it is `None`
  when no row was affected. SQLite drivers only, and note it is the last
  *affected* row, not strictly the last inserted one.
- **`:execresult`** - the driver's raw result, which differs per driver: a `str`
  status tag on asyncpg, a `psycopg.AsyncCursor` / `psycopg.Cursor` on the
  psycopg drivers, and a `sqlite3.Cursor` / `aiosqlite.Cursor` on the SQLite
  drivers.

See the [feature support matrix](/docs/reference/feature-support) for which
driver supports which.

### `:copyfrom`

PostgreSQL drivers only. Bulk-inserts rows, taking a sequence of generated
`<Name>Params` objects and returning the affected row count. asyncpg goes
through `copy_records_to_table`:

```python
async def test_copy_from(conn: ConnectionLike, *, params: collections.abc.Sequence[TestCopyFromParams]) -> int:
    records = [(param.id_, param.float_test, param.int_test) for param in params]
    r = await conn.copy_records_to_table("test_copy_from", columns=["id", "float_test", "int_test"], records=records)
    return int(n) if (p := r.split()) and (n := p[-1]).isdigit() else 0
```

psycopg streams the rows through `cursor.copy()` instead (`psycopg_sync`
emits the same body without `async`/`await`):

```python
async def test_copy_from(conn: ConnectionLike, *, params: collections.abc.Sequence[TestCopyFromParams]) -> int:
    async with conn.cursor() as cur:
        async with cur.copy('COPY "test_copy_from" ("id", "float_test", "int_test") FROM STDIN') as copy:
            for param in params:
                await copy.write_row((param.id_, param.float_test, param.int_test))
        return cur.rowcount
```

Looking for prepared queries? Every supported driver prepares statements
automatically - see
[prepared queries](/docs/reference/feature-support#prepared-queries) in the
feature support reference for the per-driver details.

## Parameters

By default every parameter is **keyword-only** (note the `*,` in the signatures
above), which keeps call sites readable and order-independent. Two options tune
this:

- **`omit_kwargs_limit`** - queries with this many parameters or fewer allow
  positional arguments. Defaults to `0` (always keyword-only).
- **`query_parameter_limit`** - when set to a non-negative value, queries with
  more parameters than the limit bundle them into a single generated
  `<Name>Params` object instead of expanding them into the signature. Left unset
  or negative, parameters are never bundled. `:copyfrom` always uses a params
  class regardless.

## Grouping into a class

With `emit_classes: true`, the standalone functions of each query file become
methods on a class named after that file - `queries_field_namings.sql` yields
`QueriesFieldNamings` - so you get one class per query module rather than one
class overall.

The connection is passed once to the constructor and is also exposed as a
read-only `conn` property. The bodies are otherwise unchanged, except that `conn`
becomes `self._conn`:

```python
class QueriesFieldNamings:
    __slots__ = ("_conn",)

    def __init__(self, conn: ConnectionLike) -> None:
        self._conn = conn

    @property
    def conn(self) -> ConnectionLike:
        return self._conn

    async def get_field_naming(self, *, id_: int) -> models.TestFieldNaming | None:
        row = await self._conn.fetchrow(GET_FIELD_NAMING, id_)
        ...
```

so the call becomes
`user = await QueriesFieldNamings(conn).get_field_naming(id_=1)` - without the
`await` on the synchronous `sqlite3` driver, where the method is a plain
function.
