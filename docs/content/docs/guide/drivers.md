---
title: Drivers
weight: 20
prev: /docs/guide/configuration
next: /docs/guide/model-types
---

The `sql_driver` option picks which database library the generated code targets.
It must match your `engine`. Three drivers are supported:

| Driver | Engine | Style |
|---|---|---|
| `asyncpg` | `postgresql` | async |
| `aiosqlite` | `sqlite` | async |
| `sqlite3` | `sqlite` | sync |

Every generated query function takes the connection as its first argument, so you
open and manage the connection yourself and pass it in.

## asyncpg (PostgreSQL)

```python
import asyncpg

from app.db import queries

conn = await asyncpg.connect("postgresql://user:pass@localhost/db")
user = await queries.get_field_naming(conn, id_=1)
```

asyncpg is the only driver that supports `:copyfrom` (bulk insert via
`copy_records_to_table`).

## aiosqlite (async SQLite)

```python
import sqlite3

import aiosqlite

from app.db import queries

async with aiosqlite.connect("app.db", detect_types=sqlite3.PARSE_DECLTYPES) as conn:
    user = await queries.get_field_naming(conn, id_=1)
```

## sqlite3 (sync SQLite)

```python
import sqlite3

from app.db import queries

conn = sqlite3.connect("app.db", detect_types=sqlite3.PARSE_DECLTYPES)
user = queries.get_field_naming(conn, id_=1)
```

{{< callout type="warning" >}}
  Whenever the generated code registers type conversions - which it does for
  `date`, `datetime`/`timestamp`, `decimal`, `bool`, and `blob` columns - both
  SQLite drivers need `detect_types=sqlite3.PARSE_DECLTYPES` on the connection,
  because converters only run when declared-type parsing is enabled. See
  [SQLite type conversion](/docs/guide/sqlite-type-conversion).
{{< /callout >}}

## Command support

Not every [query command](/docs/guide/writing-queries) works on every driver -
for example `:copyfrom` is asyncpg-only and `:execlastid` is SQLite-only. The
full matrix is in the [feature support reference](/docs/reference/feature-support).
