---
title: Drivers
weight: 20
prev: /docs/guide/configuration
next: /docs/guide/model-types
---

The `sql_driver` option picks which database library the generated code targets.
It must match your `engine`. Four drivers are supported:

| Driver | Engine | Style |
|---|---|---|
| `asyncpg` | `postgresql` | async |
| `psycopg_async` | `postgresql` | async |
| `aiosqlite` | `sqlite` | async |
| `sqlite3` | `sqlite` | sync |

Every generated query function takes the connection as its first argument, so you
open and manage the connection yourself and pass it in.

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

The generated code targets [Psycopg 3](https://www.psycopg.org/psycopg3/) with
its default tuple rows - the connection annotation is
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

## Command support

Not every [query command](/docs/guide/writing-queries) works on every driver -
for example `:copyfrom` is asyncpg-only and `:execlastid` is SQLite-only. The
full matrix is in the [feature support reference](/docs/reference/feature-support).
