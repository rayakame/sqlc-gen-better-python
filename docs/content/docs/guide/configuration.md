---
title: Configuration
weight: 10
prev: /docs/guide
next: /docs/guide/drivers
---

Everything the plugin does is driven by your `sqlc.yaml`. This page covers how
the plugin is wired in and how its options are structured. For the exhaustive
option list, see the [configuration options reference](/docs/reference/configuration-options).

## Anatomy of `sqlc.yaml`

Two parts matter: a `plugins` entry that declares the WASM plugin, and a
`codegen` entry inside your `sql` block that runs it with a set of `options`.

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
          model_type: "msgspec"
```

- **`plugins`** declares the plugin once, pinned to a release `url` and its
  `sha256`. `sqlc` downloads and runs the WASM binary.
- **`sql`** points `sqlc` at your `engine`, `queries`, and `schema`.
- **`codegen`** runs the plugin: `out` is the output directory, `plugin` refers
  to the name declared above, and `options` configures this plugin.

{{< callout type="warning" >}}
  The `sha256` must match the release you pin - `sqlc` refuses to run a plugin
  whose hash does not match. Each release lists its hash.
{{< /callout >}}

## The options block

`options` is where all plugin configuration lives. Three options are required:

| Option | What it does |
|---|---|
| `package` | The name of the generated package. |
| `sql_driver` | `asyncpg`, `aiosqlite`, or `sqlite3` - must match the `engine`. See [Drivers](/docs/guide/drivers). |
| `emit_init_file` | Whether to emit `__init__.py`. Must be set explicitly. |

Everything else is optional and has a sensible default. The most common ones to
reach for next are [`model_type`](/docs/guide/model-types) and
[`docstrings`](/docs/guide/docstrings). The full list, with types and defaults,
is in the [reference](/docs/reference/configuration-options).

## Multiple outputs

A single `sql` block can have several `codegen` entries, each writing to its own
`out`. This is handy to generate more than one flavor from the same schema and
queries - for example a `msgspec` package and a `dataclass` package:

```yaml
    codegen:
      - out: "app/db_msgspec"
        plugin: python
        options:
          package: "db_msgspec"
          emit_init_file: true
          sql_driver: "asyncpg"
          model_type: "msgspec"
      - out: "app/db_dataclass"
        plugin: python
        options:
          package: "db_dataclass"
          emit_init_file: true
          sql_driver: "asyncpg"
          model_type: "dataclass"
```

## Common pitfalls

- **Driver/engine mismatch.** `sql_driver: asyncpg` requires `engine: "postgresql"`;
  `aiosqlite`/`sqlite3` require `engine: "sqlite"`. A mismatch is an error.
- **Forgetting `emit_init_file`.** It has no default and generation fails if it
  is omitted. Set it to `true` unless the package already has an `__init__.py`.
- **A stale `sha256`.** When you bump the plugin version, update the hash too.
