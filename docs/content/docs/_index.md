---
title: Documentation
weight: 1
prev: /
next: getting-started
sidebar:
  open: true
---

`sqlc-gen-better-python` is a [sqlc](https://sqlc.dev) plugin that turns your SQL
schema and queries into modern, fully typed Python database code: models, typed
query functions, and enums. You keep writing SQL; the Python stays in sync with
it.

{{< cards >}}
  {{< card link="getting-started" title="Getting Started" icon="play" subtitle="Install the plugin and go from a schema to your first typed queries." >}}
  {{< card link="guide" title="Guide" icon="book-open" subtitle="Every feature, explained with real schema, queries, and generated output." >}}
  {{< card link="reference" title="Reference" icon="clipboard-list" subtitle="Configuration options, type mappings, and per-driver support." >}}
{{< /cards >}}

## What it looks like

You write a query:

```sql
-- name: GetUser :one
SELECT * FROM users WHERE id = $1;
```

and get a typed function, with a model built from your schema:

```python
async def get_user(conn: ConnectionLike, *, id_: int) -> models.User | None:
    row = await conn.fetchrow(GET_USER, id_)
    if row is None:
        return None
    return models.User(id_=row[0], name=row[1])
```

No ORM, no hand-written row unpacking, and pyright checks every field access.

## Explore by feature

{{< cards >}}
  {{< card link="guide/drivers" title="Drivers" icon="database" subtitle="asyncpg, aiosqlite, and sqlite3." >}}
  {{< card link="guide/model-types" title="Model types" icon="cube" subtitle="dataclass, attrs, msgspec, or pydantic." >}}
  {{< card link="guide/writing-queries" title="Writing queries" icon="code" subtitle="How query annotations become typed functions." >}}
  {{< card link="guide/enums" title="Enums" icon="collection" subtitle="PostgreSQL enums as enum.StrEnum classes." >}}
  {{< card link="guide/type-overrides" title="Type overrides" icon="adjustments" subtitle="Swap the Python type of a column." >}}
  {{< card link="guide/converters" title="Converters" icon="puzzle" subtitle="Your own encode/decode functions for a column." >}}
  {{< card link="guide/working-with-json" title="Working with JSON" icon="document-text" subtitle="Typed JSON columns with msgspec structs." >}}
  {{< card link="guide/docstrings" title="Docstrings" icon="document" subtitle="google, numpy, or pep257 docstrings." >}}
{{< /cards >}}

## At a glance

| | |
|---|---|
| **Python** | 3.12 or newer |
| **Engines** | PostgreSQL, SQLite |
| **Drivers** | `asyncpg`, `aiosqlite`, `sqlite3` |
| **Model types** | `dataclass`, `attrs`, `msgspec`, `pydantic` |
| **Docstrings** | `google`, `numpy`, `pep257`, or none |
| **Checked with** | pyright (strict) and ruff |

Full details in the [feature support reference](/docs/reference/feature-support).

## Getting help

{{< cards >}}
  {{< card link="https://github.com/rayakame/sqlc-gen-better-python/issues" title="Issues" icon="github" subtitle="Report a bug or request a feature." >}}
  {{< card link="https://discord.gg/hikari" title="Discord" icon="chat" subtitle="Ask a question or share what you built." >}}
  {{< card link="https://github.com/rayakame/sqlc-gen-better-python/blob/main/CHANGELOG.md" title="Changelog" icon="clipboard-check" subtitle="What changed in each release." >}}
{{< /cards >}}
