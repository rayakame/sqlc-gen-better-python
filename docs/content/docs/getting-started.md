---
title: Getting Started
weight: 1
prev: /docs
next: /docs/guide
tabs:
  sync: true
---

From an empty project to your first typed queries. Pick your driver in the tabs
below and the whole page follows your choice.

## Prerequisites

You need [`sqlc`](https://docs.sqlc.dev/en/latest/overview/install.html) on your
`PATH` and **Python 3.12 or newer** (the generated code uses PEP 695 type aliases
and generics, and `enum.StrEnum`).

{{< callout type="info" >}}
  Besides the official installation methods, sqlc is also pip-installable via
  [sqlc-bin](https://pypi.org/project/sqlc-bin/), which ships the unmodified
  official binaries - no Go toolchain required. `uv add --dev sqlc-bin` (or
  `pip install sqlc-bin`) puts `sqlc` on your PATH, and the package version
  tracks the sqlc version, so you can pin it like any other dependency.
{{< /callout >}}

Then install the database driver you want to use:

{{< tabs >}}

  {{< tab name="asyncpg" >}}

```bash
pip install asyncpg
```

  {{< /tab >}}

  {{< tab name="psycopg_async" >}}

```bash
pip install "psycopg[binary]"
```

  {{< /tab >}}

  {{< tab name="psycopg_sync" >}}

```bash
pip install "psycopg[binary]"
```

  {{< /tab >}}

  {{< tab name="aiosqlite" >}}

```bash
pip install aiosqlite
```

  {{< /tab >}}

  {{< tab name="sqlite3" >}}

```text
Nothing to install - sqlite3 is in the standard library.
```

  {{< /tab >}}

{{< /tabs >}}

## 1. Configure the plugin

The plugin is a WASM binary that `sqlc generate` downloads and runs. Create a
`sqlc.yaml`:

{{< tabs >}}

  {{< tab name="asyncpg" >}}

```yaml
# filename: sqlc.yaml
version: "2"
plugins:
  - name: python
    wasm:
      url: https://github.com/rayakame/sqlc-gen-better-python/releases/download/v0.6.0/sqlc-gen-better-python.wasm
      sha256: 16f5affb502f2ec65ca61f6fc5ddd993449c4a4fc281996c3c9a9bc2e35b1474
sql:
  - engine: "postgresql"
    queries: "query.sql"
    schema: "schema.sql"
    codegen:
      - out: "app/db"
        plugin: python
        options:
          package: "db"
          emit_init_file: true
          sql_driver: "asyncpg"
          model_type: "dataclass"
```

  {{< /tab >}}

  {{< tab name="psycopg_async" >}}

```yaml
# filename: sqlc.yaml
version: "2"
plugins:
  - name: python
    wasm:
      url: https://github.com/rayakame/sqlc-gen-better-python/releases/download/v0.6.0/sqlc-gen-better-python.wasm
      sha256: 16f5affb502f2ec65ca61f6fc5ddd993449c4a4fc281996c3c9a9bc2e35b1474
sql:
  - engine: "postgresql"
    queries: "query.sql"
    schema: "schema.sql"
    codegen:
      - out: "app/db"
        plugin: python
        options:
          package: "db"
          emit_init_file: true
          sql_driver: "psycopg_async"
          model_type: "dataclass"
```

  {{< /tab >}}

  {{< tab name="psycopg_sync" >}}

```yaml
# filename: sqlc.yaml
version: "2"
plugins:
  - name: python
    wasm:
      url: https://github.com/rayakame/sqlc-gen-better-python/releases/download/v0.6.0/sqlc-gen-better-python.wasm
      sha256: 16f5affb502f2ec65ca61f6fc5ddd993449c4a4fc281996c3c9a9bc2e35b1474
sql:
  - engine: "postgresql"
    queries: "query.sql"
    schema: "schema.sql"
    codegen:
      - out: "app/db"
        plugin: python
        options:
          package: "db"
          emit_init_file: true
          sql_driver: "psycopg_sync"
          model_type: "dataclass"
```

  {{< /tab >}}

  {{< tab name="aiosqlite" >}}

```yaml
# filename: sqlc.yaml
version: "2"
plugins:
  - name: python
    wasm:
      url: https://github.com/rayakame/sqlc-gen-better-python/releases/download/v0.6.0/sqlc-gen-better-python.wasm
      sha256: 16f5affb502f2ec65ca61f6fc5ddd993449c4a4fc281996c3c9a9bc2e35b1474
sql:
  - engine: "sqlite"
    queries: "query.sql"
    schema: "schema.sql"
    codegen:
      - out: "app/db"
        plugin: python
        options:
          package: "db"
          emit_init_file: true
          sql_driver: "aiosqlite"
          model_type: "dataclass"
```

  {{< /tab >}}

  {{< tab name="sqlite3" >}}

```yaml
# filename: sqlc.yaml
version: "2"
plugins:
  - name: python
    wasm:
      url: https://github.com/rayakame/sqlc-gen-better-python/releases/download/v0.6.0/sqlc-gen-better-python.wasm
      sha256: 16f5affb502f2ec65ca61f6fc5ddd993449c4a4fc281996c3c9a9bc2e35b1474
sql:
  - engine: "sqlite"
    queries: "query.sql"
    schema: "schema.sql"
    codegen:
      - out: "app/db"
        plugin: python
        options:
          package: "db"
          emit_init_file: true
          sql_driver: "sqlite3"
          model_type: "dataclass"
```

  {{< /tab >}}

{{< /tabs >}}

{{< callout type="warning" >}}
  Always pin the `sha256` of the release you use - `sqlc` refuses to run a plugin
  whose hash does not match. Each release lists its hash.
{{< /callout >}}

`model_type: "dataclass"` is the default and needs no extra dependency. See
[Model types](/docs/guide/model-types) for `attrs`, `msgspec`, and `pydantic`.

## 2. Describe your schema

{{< tabs >}}

  {{< tab name="asyncpg" >}}

```sql
-- filename: schema.sql
CREATE TABLE users
(
    id   bigint PRIMARY KEY NOT NULL,
    name text               NOT NULL
);
```

  {{< /tab >}}

  {{< tab name="psycopg_async" >}}

```sql
-- filename: schema.sql
CREATE TABLE users
(
    id   bigint PRIMARY KEY NOT NULL,
    name text               NOT NULL
);
```

  {{< /tab >}}

  {{< tab name="psycopg_sync" >}}

```sql
-- filename: schema.sql
CREATE TABLE users
(
    id   bigint PRIMARY KEY NOT NULL,
    name text               NOT NULL
);
```

  {{< /tab >}}

  {{< tab name="aiosqlite" >}}

```sql
-- filename: schema.sql
CREATE TABLE users
(
    id   INTEGER PRIMARY KEY NOT NULL,
    name TEXT                NOT NULL
);
```

  {{< /tab >}}

  {{< tab name="sqlite3" >}}

```sql
-- filename: schema.sql
CREATE TABLE users
(
    id   INTEGER PRIMARY KEY NOT NULL,
    name TEXT                NOT NULL
);
```

  {{< /tab >}}

{{< /tabs >}}

## 3. Write your queries

Each query is a `-- name: <Name> :<command>` comment followed by SQL. The name
becomes the Python function, and the command decides what it returns - `:one`
returns a single row or `None`, `:many` returns all matching rows.

{{< tabs >}}

  {{< tab name="asyncpg" >}}

```sql
-- filename: query.sql
-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users ORDER BY name;
```

  {{< /tab >}}

  {{< tab name="psycopg_async" >}}

```sql
-- filename: query.sql
-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users ORDER BY name;
```

  {{< /tab >}}

  {{< tab name="psycopg_sync" >}}

```sql
-- filename: query.sql
-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users ORDER BY name;
```

  {{< /tab >}}

  {{< tab name="aiosqlite" >}}

```sql
-- filename: query.sql
-- name: GetUser :one
SELECT * FROM users WHERE id = ?;

-- name: ListUsers :many
SELECT * FROM users ORDER BY name;
```

  {{< /tab >}}

  {{< tab name="sqlite3" >}}

```sql
-- filename: query.sql
-- name: GetUser :one
SELECT * FROM users WHERE id = ?;

-- name: ListUsers :many
SELECT * FROM users ORDER BY name;
```

  {{< /tab >}}

{{< /tabs >}}

PostgreSQL uses `$1` placeholders, SQLite uses `?`. Everything else is the same.
(You write `$1` for psycopg too - the plugin rewrites the placeholders to
psycopg's format at generation time.)

## 4. Generate

Run `sqlc` from the directory containing your `sqlc.yaml`:

```bash
sqlc generate
```

This writes a Python package to `out` (`app/db` above) containing `models.py`,
one query module per query file (`query.py`), and an `__init__.py`.

## 5. What you got

A model per table, and a typed function per query:

{{< tabs >}}

  {{< tab name="asyncpg" >}}

```python
# models.py
@dataclasses.dataclass()
class User:
    id_: int
    name: str


# query.py
async def get_user(conn: ConnectionLike, *, id_: int) -> models.User | None:
    row = await conn.fetchrow(GET_USER, id_)
    if row is None:
        return None
    return models.User(id_=row[0], name=row[1])


def list_users(conn: ConnectionLike) -> QueryResults[models.User]:
    ...
```

  {{< /tab >}}

  {{< tab name="psycopg_async" >}}

```python
# models.py
@dataclasses.dataclass()
class User:
    id_: int
    name: str


# query.py
async def get_user(conn: ConnectionLike, *, id_: int) -> models.User | None:
    row = await (await conn.execute(GET_USER, {"p1": id_})).fetchone()
    if row is None:
        return None
    return models.User(id_=row[0], name=row[1])


def list_users(conn: ConnectionLike) -> QueryResults[models.User]:
    ...
```

  {{< /tab >}}

  {{< tab name="psycopg_sync" >}}

```python
# models.py
@dataclasses.dataclass()
class User:
    id_: int
    name: str


# query.py
def get_user(conn: ConnectionLike, *, id_: int) -> models.User | None:
    row = conn.execute(GET_USER, {"p1": id_}).fetchone()
    if row is None:
        return None
    return models.User(id_=row[0], name=row[1])


def list_users(conn: ConnectionLike) -> QueryResults[models.User]:
    ...
```

  {{< /tab >}}

  {{< tab name="aiosqlite" >}}

```python
# models.py
@dataclasses.dataclass()
class User:
    id_: int
    name: str


# query.py
async def get_user(conn: aiosqlite.Connection, *, id_: int) -> models.User | None:
    row = await (await conn.execute(GET_USER, (id_,))).fetchone()
    if row is None:
        return None
    return models.User(id_=row[0], name=row[1])


def list_users(conn: aiosqlite.Connection) -> QueryResults[models.User]:
    ...
```

  {{< /tab >}}

  {{< tab name="sqlite3" >}}

```python
# models.py
@dataclasses.dataclass()
class User:
    id_: int
    name: str


# query.py
def get_user(conn: sqlite3.Connection, *, id_: int) -> models.User | None:
    row = conn.execute(GET_USER, (id_,)).fetchone()
    if row is None:
        return None
    return models.User(id_=row[0], name=row[1])


def list_users(conn: sqlite3.Connection) -> QueryResults[models.User]:
    ...
```

  {{< /tab >}}

{{< /tabs >}}

Two things to notice:

- The `id` column became `id_`, because `id` is a Python builtin. See
  [Naming and identifiers](/docs/guide/naming).
- Parameters are keyword-only (`*,`), so you call `get_user(conn, id_=1)`.

## 6. Use it

`:one` gives you a model or `None`. `:many` returns a `QueryResults`, which you
can consume all at once or stream row by row:

{{< tabs >}}

  {{< tab name="asyncpg" >}}

```python
import asyncio

import asyncpg

from app.db import query


async def main() -> None:
    conn = await asyncpg.connect("postgresql://user:pass@localhost/mydb")

    user = await query.get_user(conn, id_=1)
    if user is not None:
        print(user.name)

    # every row at once
    users = await query.list_users(conn)

    # or stream them
    async for user in query.list_users(conn):
        print(user.name)


asyncio.run(main())
```

  {{< /tab >}}

  {{< tab name="psycopg_async" >}}

```python
import asyncio

import psycopg

from app.db import query


async def main() -> None:
    async with await psycopg.AsyncConnection.connect("postgresql://user:pass@localhost/mydb") as conn:
        user = await query.get_user(conn, id_=1)
        if user is not None:
            print(user.name)

        # every row at once
        users = await query.list_users(conn)

        # or iterate
        async for user in query.list_users(conn):
            print(user.name)


asyncio.run(main())
```

  {{< /tab >}}

  {{< tab name="psycopg_sync" >}}

```python
import psycopg

from app.db import query

with psycopg.connect("postgresql://user:pass@localhost/mydb") as conn:
    user = query.get_user(conn, id_=1)
    if user is not None:
        print(user.name)

    # every row at once
    users = query.list_users(conn)()

    # or iterate
    for user in query.list_users(conn):
        print(user.name)
```

  {{< /tab >}}

  {{< tab name="aiosqlite" >}}

```python
import asyncio
import sqlite3

import aiosqlite

from app.db import query


async def main() -> None:
    async with aiosqlite.connect("app.db", detect_types=sqlite3.PARSE_DECLTYPES) as conn:
        user = await query.get_user(conn, id_=1)
        if user is not None:
            print(user.name)

        # every row at once
        users = await query.list_users(conn)

        # or stream them
        async for user in query.list_users(conn):
            print(user.name)


asyncio.run(main())
```

{{< callout type="info" >}}
  The `detect_types=sqlite3.PARSE_DECLTYPES` above is not needed for this schema,
  but it is what makes generated converters work once your tables use dates,
  decimals, booleans, or blobs. See
  [SQLite type conversion](/docs/guide/sqlite-type-conversion).
{{< /callout >}}

  {{< /tab >}}

  {{< tab name="sqlite3" >}}

```python
import sqlite3

from app.db import query

conn = sqlite3.connect("app.db", detect_types=sqlite3.PARSE_DECLTYPES)

user = query.get_user(conn, id_=1)
if user is not None:
    print(user.name)

# every row at once - call the result
users = query.list_users(conn)()

# or iterate
for user in query.list_users(conn):
    print(user.name)
```

{{< callout type="info" >}}
  The `detect_types=sqlite3.PARSE_DECLTYPES` above is not needed for this schema,
  but it is what makes generated converters work once your tables use dates,
  decimals, booleans, or blobs. See
  [SQLite type conversion](/docs/guide/sqlite-type-conversion).
{{< /callout >}}

  {{< /tab >}}

{{< /tabs >}}

## Next steps

- Work through the [Guide](/docs/guide) - configuration, drivers, model types,
  writing queries, and every feature, each with real generated output.
- Look up specifics in the [Reference](/docs/reference): the full option list,
  SQL-to-Python type mappings, and per-driver support.
