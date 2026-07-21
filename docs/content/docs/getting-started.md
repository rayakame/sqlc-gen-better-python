---
title: Getting Started
weight: 2
prev: /docs
next: converters
---

## Prerequisites

You need [`sqlc`](https://docs.sqlc.dev/en/latest/overview/install.html) on your
`PATH`. The generated code targets **Python 3.12 or newer** (it uses PEP 695 type
aliases and generics, and `enum.StrEnum`).

## Configure the plugin

The plugin is a WASM binary that `sqlc generate` downloads and runs. Point your
`sqlc.yaml` at a release and select a driver and model type:

```yaml
# filename: sqlc.yaml
version: "2"
plugins:
  - name: python
    wasm:
      url: https://github.com/rayakame/sqlc-gen-better-python/releases/download/v0.5.1/sqlc-gen-better-python.wasm
      sha256: c7cc470df7625ae3232c2b042060b948180ae784ce3d81c32e8a2c040fe04fa7
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

{{< callout type="warning" >}}
  Always pin the `sha256` of the release you use &mdash; `sqlc` refuses to run a
  plugin whose hash does not match. Each release lists its hash.
{{< /callout >}}

## Generate

Run `sqlc` from the directory containing your `sqlc.yaml`:

```bash
sqlc generate
```

This writes a Python package to `out` containing `models.py`, one query module
per query file, and (when your schema has enums) an `enums.py`.

## Next steps

- Full option reference and driver support: see the
  [project README](https://github.com/rayakame/sqlc-gen-better-python#configuration-options)
  while the remaining pages are ported here.
- Plug in your own serialization with [Converters](converters).
