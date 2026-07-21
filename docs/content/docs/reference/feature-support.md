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

| Command | aiosqlite | sqlite3 | asyncpg |
|---|---|---|---|
| `:one` | yes | yes | yes |
| `:many` | yes | yes | yes |
| `:exec` | yes | yes | yes |
| `:execresult` | yes | yes | yes |
| `:execrows` | yes | yes | yes |
| `:execlastid` | yes | yes | no |
| `:copyfrom` | no | no | yes |

See [Writing queries](/docs/guide/writing-queries) for what each command
generates.

{{< callout type="info" >}}
  `:execlastid` relies on a last-inserted-row id, which asyncpg/PostgreSQL do not
  provide; use a `RETURNING` clause with `:one` instead. `:copyfrom` maps to
  asyncpg's bulk `copy_records_to_table`, which the SQLite drivers have no
  equivalent for.
{{< /callout >}}

## Not supported

- **`:batch*` commands** (`:batchexec`, `:batchmany`, `:batchone`) are not
  supported and likely never will be.
- **Prepared queries** are not planned for the near future.
- **`psycopg2` and `mysql`** drivers are not currently supported.
