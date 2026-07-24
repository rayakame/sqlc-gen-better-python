---
title: Feature support
weight: 30
prev: /docs/reference/type-mappings
---

Which sqlc features and query commands the plugin supports.

## Macros

Every [sqlc macro](https://docs.sqlc.dev/en/latest/reference/macros.html) is
supported (`sqlc.arg`, `sqlc.narg`, `sqlc.embed`, `sqlc.slice`).

{{< callout type="info" >}}
  `sqlc.slice` is for the SQLite drivers, where a list cannot be passed to the
  `IN` operator: the generated function expands the placeholder at call time,
  one `?` per element, and an empty sequence matches no rows. Because the SQL
  is built per call, it cannot be used with prepared statements. On PostgreSQL
  the macro is not needed - use `= ANY($1::type[])`, which accepts the sequence
  directly.
{{< /callout >}}

## Query commands

The supported [query annotations](https://docs.sqlc.dev/en/latest/reference/query-annotations.html)
depend on the driver:

| Command | aiosqlite | sqlite3 | asyncpg | psycopg_async | psycopg_sync |
|---|---|---|---|---|---|
| `:one` | yes | yes | yes | yes | yes |
| `:many` | yes | yes | yes | yes | yes |
| `:exec` | yes | yes | yes | yes | yes |
| `:execresult` | yes | yes | yes | yes | yes |
| `:execrows` | yes | yes | yes | yes | yes |
| `:execlastid` | yes | yes | no | no | no |
| `:copyfrom` | no | no | yes | yes | yes |

See [Writing queries](/docs/guide/writing-queries) for what each command
generates.

{{< callout type="info" >}}
  `:execlastid` relies on a last-inserted-row id, which PostgreSQL does not
  provide; use a `RETURNING` clause with `:one` instead. `:copyfrom` maps to
  PostgreSQL's bulk `COPY` protocol (`copy_records_to_table` on asyncpg,
  `cursor.copy()` on psycopg), which the SQLite drivers have no equivalent for.
{{< /callout >}}

### Prepared queries

Coming from sqlc's Go workflow you might look for an
[`emit_prepared_queries`](https://docs.sqlc.dev/en/latest/howto/prepared_query.html)
equivalent. There is none, on purpose: every supported Python driver already
prepares statements automatically, so the generated code gets prepared-query
performance without any extra codegen. What differs per driver is *when* a
query gets prepared and which knob controls it:

- **asyncpg** prepares every query it runs and keeps it in a per-connection
  LRU statement cache (100 entries by default). Tune it at connect time:

  ```python
  conn = await asyncpg.connect(
      dsn,
      statement_cache_size=200,  # default 100; 0 disables the cache
  )
  ```

- **psycopg** (both flavors) prepares a query server-side once it has been
  executed more than `prepare_threshold` times on the connection - with the
  default of 5, the sixth execution is the first prepared one. Set it to `0`
  to prepare from the first execution, or `None` to never prepare:

  ```python
  conn = await psycopg.AsyncConnection.connect(dsn, prepare_threshold=0)  # psycopg_async
  conn = psycopg.connect(dsn, prepare_threshold=0)  # psycopg_sync
  ```

- **sqlite3 / aiosqlite** expose no explicit prepare API, but the `sqlite3`
  module compiles each statement once and reuses it through an internal
  per-connection cache (128 entries by default). Raise it with the
  `cached_statements` argument of `connect()` if you have more distinct
  queries than that.

{{< callout type="warning" >}}
  Behind PgBouncer in transaction-pooling mode, server-side prepared
  statements belong to a connection you do not control. Disable them there:
  `statement_cache_size=0` for asyncpg, `prepare_threshold=None` for psycopg.
{{< /callout >}}

## Not supported

- **`:batch*` commands** (`:batchexec`, `:batchmany`, `:batchone`) are not
  supported and likely never will be.
- **`psycopg2` and `mysql`** drivers are not currently supported; Psycopg 3
  is, via the `psycopg_async` (asyncio) and `psycopg_sync` (synchronous)
  drivers.
