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
  Known issue in `v0.5.1`: a query using `sqlc.slice` is generated with sqlc's
  `/*SLICE:name*/` placeholder left in place, so it fails at runtime. A fix is
  landing before the next release - see
  [issue #210](https://github.com/rayakame/sqlc-gen-better-python/issues/210).
  On PostgreSQL, `= ANY($1::type[])` is the idiomatic form anyway and is
  unaffected.
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
