---
title: Feature support
weight: 30
prev: /docs/reference/type-mappings
---

Which sqlc features and query commands the plugin supports.

## Macros

Of the [sqlc macros](https://docs.sqlc.dev/en/latest/reference/macros.html),
`sqlc.arg`, `sqlc.narg`, and `sqlc.embed` are supported.

{{< callout type="warning" >}}
  **`sqlc.slice` is not supported.** The parameter is typed as a
  `collections.abc.Sequence[...]`, but the `/*SLICE:name*/` placeholder that sqlc
  emits is never expanded, so the query fails at runtime when the driver is handed
  a list. On PostgreSQL, use `= ANY($1::type[])` instead - that is fully
  supported, including with [converters](/docs/guide/converters).
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
