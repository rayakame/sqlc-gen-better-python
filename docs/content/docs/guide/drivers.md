---
title: Drivers
weight: 20
prev: /docs/guide/configuration
next: /docs/guide/model-types
---

The `sql_driver` option picks which database library the generated code targets.
It must match your `engine`. Seven drivers are supported:

| Driver | Engine | Style |
|---|---|---|
| `asyncpg` | `postgresql` | async |
| `psycopg_async` | `postgresql` | async |
| `psycopg_sync` | `postgresql` | sync |
| `aiosqlite` | `sqlite` | async |
| `sqlite3` | `sqlite` | sync |
| `turso_async` | `sqlite` | async (experimental) |
| `turso_sync` | `sqlite` | sync (experimental) |

Every generated query function takes the connection as its first argument, so you
open and manage the connection yourself and pass it in.

All PostgreSQL drivers produce the same models and type contract, so choosing
between them is about the driver itself: pick `asyncpg` when raw driver
throughput is the priority, and one of the psycopg drivers to stay in the
psycopg ecosystem (libpq, pipeline mode, PgBouncer friendliness) at comparable
speed - `psycopg_async` for asyncio code, `psycopg_sync` for plain synchronous
code.

## asyncpg (PostgreSQL)

```python
import asyncio

import asyncpg

from app.db import queries


async def main() -> None:
    conn = await asyncpg.connect("postgresql://user:pass@localhost/db")
    user = await queries.get_field_naming(conn, id_=1)


asyncio.run(main())
```

asyncpg supports `:copyfrom` (bulk insert via `copy_records_to_table`).

{{< callout type="info" >}}
  asyncpg itself ships without strict type annotations - install
  [asyncpg-stubs](https://pypi.org/project/asyncpg-stubs/) so that pyright or
  mypy understand generated annotations like
  `asyncpg.Connection[asyncpg.Record]`. This only affects type checking and
  is never evaluated at runtime.
{{< /callout >}}

## psycopg_async (PostgreSQL)

```python
import asyncio

import psycopg

from app.db import queries


async def main() -> None:
    conn = await psycopg.AsyncConnection.connect("postgresql://user:pass@localhost/db")
    user = await queries.get_field_naming(conn, id_=1)


asyncio.run(main())
```

The generated code targets [Psycopg 3](https://www.psycopg.org/psycopg3/)
(3.2 or newer) with its default tuple rows - the connection annotation is
`psycopg.AsyncConnection[psycopg.rows.TupleRow]`, so a connection configured
with another row factory is rejected by pyright. `:copyfrom` streams rows
through `cursor.copy()`.

{{< callout type="info" >}}
  Modules returning `json`/`jsonb` columns register a raw-text loader on
  psycopg's process-global adapters map at import time, so those columns stay
  `str` exactly like on asyncpg - including for
  [converters](/docs/guide/converters). On Windows, psycopg's async support
  requires the `SelectorEventLoop`; the default `ProactorEventLoop` is
  rejected.
{{< /callout >}}

## psycopg_sync (PostgreSQL)

```python
import psycopg

from app.db import queries

with psycopg.connect("postgresql://user:pass@localhost/db") as conn:
    user = queries.get_field_naming(conn, id_=1)
```

The synchronous flavor of the psycopg driver (Psycopg 3.2 or newer, like
`psycopg_async`): identical models, placeholders, and type contract, emitted
as plain functions with no `async`/`await`. The connection annotation is
`psycopg.Connection[psycopg.rows.TupleRow]`, and `:many` queries return the
same `QueryResults` helper - call it (`queries.list_x(conn)()`) to fetch every
row at once, or iterate it directly with a plain `for` loop. The json/jsonb
raw-text loader registration works exactly as on `psycopg_async`; the Windows
event-loop caveat does not apply.

## aiosqlite (async SQLite)

```python
import asyncio
import sqlite3

import aiosqlite

from app.db import queries


async def main() -> None:
    async with aiosqlite.connect("app.db", detect_types=sqlite3.PARSE_DECLTYPES) as conn:
        user = await queries.get_field_naming(conn, id_=1)


asyncio.run(main())
```

## sqlite3 (sync SQLite)

```python
import sqlite3

from app.db import queries

conn = sqlite3.connect("app.db", detect_types=sqlite3.PARSE_DECLTYPES)
user = queries.get_field_naming(conn, id_=1)
```

{{< callout type="warning" >}}
  When a query *returns* a `date`, `datetime`/`timestamp`, `decimal`, `bool`, or
  `blob` column, the generated code registers a converter for it - and converters
  only run if the connection was opened with
  `detect_types=sqlite3.PARSE_DECLTYPES`. Adapters, which send those types as
  parameters, work without it. See
  [SQLite type conversion](/docs/guide/sqlite-type-conversion).
{{< /callout >}}

## turso_sync / turso_async (Turso)

[Turso](https://github.com/tursodatabase/turso) is an SQLite-compatible
database engine; its [pyturso](https://pypi.org/project/pyturso/) package
(`pip install pyturso`) mirrors the `sqlite3` module's API and adds a native
asyncio variant.

- `turso_sync` targets the `turso` module, `turso_async` the `turso.aio` module.
- Both use `engine: "sqlite"` and the same `?` placeholders and queries as the
  SQLite drivers.
- Unlike the `sqlite3` module, no connection flags are needed - the generated
  code converts values inline, and the observable Python types match the
  SQLite drivers exactly.

```python
import turso

from app.db import queries

conn = turso.connect("app.db")
user = queries.get_field_naming(conn, id_=1)
```

```python
import asyncio

import turso.aio

from app.db import queries


async def main() -> None:
    conn = await turso.aio.connect("app.db")
    user = await queries.get_field_naming(conn, id_=1)


asyncio.run(main())
```

{{< callout type="warning" >}}
  The turso drivers are **experimental** - Turso and pyturso are pre-1.0, so
  the API may still change. Known differences from the `sqlite3` module:

  - `:execlastid` returns `None` for `UPDATE`/`DELETE` statements - turso's
    `lastrowid` only reflects the cursor's own `INSERT`.
  - `:execrows` reports `0` instead of `-1` for statements like `CREATE TABLE`.
  - Overrides relying on a user-registered sqlite3 adapter have no turso
    equivalent - the override type must bind natively (None, numbers,
    strings, or bytes).
{{< /callout >}}

## Command support

Not every [query command](/docs/guide/writing-queries) works on every driver -
for example `:copyfrom` is PostgreSQL-only and `:execlastid` is limited to the
SQLite-engine drivers. The full matrix is in the
[feature support reference](/docs/reference/feature-support).
