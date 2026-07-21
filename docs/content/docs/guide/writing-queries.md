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

Returns a single model instance or `None`. Shown above.

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

### `:exec`

Runs the statement and returns `None`:

```python
async def set_field_naming_outputs(conn: ConnectionLike, *, id_: int, outputs: str, outputs_2: str) -> None:
    await conn.execute(SET_FIELD_NAMING_OUTPUTS, id_, outputs, outputs_2)
```

### `:execresult`, `:execrows`, `:execlastid`

Variants of `:exec` that return something about the write:

- **`:execrows`** - the number of affected rows (`int`).
- **`:execlastid`** - the id of the last inserted row (`int`); SQLite drivers only.
- **`:execresult`** - the driver's raw result/status object.

See the [feature support matrix](/docs/reference/feature-support) for which
driver supports which.

### `:copyfrom`

asyncpg only. Bulk-inserts rows via `copy_records_to_table`, taking a sequence of
generated `<Name>Params` objects and returning the affected row count:

```python
async def test_copy_from(conn: ConnectionLike, *, params: collections.abc.Sequence[TestCopyFromParams]) -> int:
    records = [(param.id_, param.float_test, param.int_test) for param in params]
    r = await conn.copy_records_to_table("test_copy_from", columns=["id", "float_test", "int_test"], records=records)
    return int(n) if (p := r.split()) and (n := p[-1]).isdigit() else 0
```

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

With `emit_classes: true`, the standalone functions become methods on a single
`Querier` class instead. The bodies are identical; only the call style changes
(`Querier(conn).get_field_naming(id_=1)`).
