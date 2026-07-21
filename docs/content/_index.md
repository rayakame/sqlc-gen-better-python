---
title: sqlc-gen-better-python
layout: hextra-home
---

{{< hextra/hero-badge >}}
  <div class="hx:w-2 hx:h-2 hx:rounded-full hx:bg-primary-400"></div>
  <span>Free, open source, self-hosted</span>
{{< /hextra/hero-badge >}}

<div class="hx:mt-6 hx:mb-6">
{{< hextra/hero-headline >}}
  Type-safe Python from your SQL
{{< /hextra/hero-headline >}}
</div>

<div class="hx:mb-12">
{{< hextra/hero-subtitle >}}
  A sqlc plugin that generates modern, fully typed Python database code from plain SQL - models plus async query functions.
{{< /hextra/hero-subtitle >}}
</div>

<div class="hx:mb-6 hx:flex hx:gap-4 hx:flex-wrap hx:items-center">
{{< hextra/hero-button text="Documentation" link="docs" >}}
{{< hextra/hero-button text="Join the Discord" link="https://discord.gg/hikari" style="background: #5865f2;" >}}
</div>

<div class="hx:mt-6"></div>

{{< hextra/feature-grid >}}
  {{< hextra/feature-card
    title="Four model types"
    link="docs/guide/model-types"
    subtitle="Generate dataclass, attrs, msgspec, or pydantic models - pick per codegen block."
  >}}
  {{< hextra/feature-card
    title="Three drivers"
    link="docs/guide/drivers"
    subtitle="asyncpg for PostgreSQL, plus aiosqlite and sqlite3 for SQLite."
  >}}
  {{< hextra/feature-card
    title="Strictly typed output"
    link="docs/guide/writing-queries"
    subtitle="Generated code passes pyright strict and ruff, targeting Python 3.12+."
  >}}
  {{< hextra/feature-card
    title="Enums"
    link="docs/guide/enums"
    subtitle="PostgreSQL enums become enum.StrEnum classes, wired through models and queries."
  >}}
  {{< hextra/feature-card
    title="Type overrides & converters"
    link="docs/guide/type-overrides"
    subtitle="Swap a column's Python type, or plug in your own encode/decode functions."
  >}}
  {{< hextra/feature-card
    title="Docstrings"
    link="docs/guide/docstrings"
    subtitle="Optional google, numpy, or pep257 docstrings on every generated function."
  >}}
{{< /hextra/feature-grid >}}

<div class="hx:mt-16"></div>

{{< hextra/hero-section >}}
  From SQL to Python
{{< /hextra/hero-section >}}

You write a query, annotated with the name and shape you want:

```sql
-- name: GetUser :one
SELECT * FROM users WHERE id = $1;
```

and get a typed function back, with a model built from your schema:

```python
async def get_user(conn: ConnectionLike, *, id_: int) -> models.User | None:
    row = await conn.fetchrow(GET_USER, id_)
    if row is None:
        return None
    return models.User(id_=row[0], name=row[1])
```

No ORM, no hand-written row unpacking, and pyright checks every field access.

<div class="hx:mt-16"></div>

{{< hextra/hero-section >}}
  One schema, four model types
{{< /hextra/hero-section >}}

Set `model_type` and the same table generates whichever flavour your project
already uses - the fields and annotations are identical:

{{< tabs >}}

  {{< tab name="dataclass" >}}

```python
@dataclasses.dataclass()
class User:
    id_: int
    name: str
```

  {{< /tab >}}

  {{< tab name="attrs" >}}

```python
@attrs.define()
class User:
    id_: int
    name: str
```

  {{< /tab >}}

  {{< tab name="msgspec" >}}

```python
class User(msgspec.Struct):
    id_: int
    name: str
```

  {{< /tab >}}

  {{< tab name="pydantic" >}}

```python
class User(pydantic.BaseModel):
    model_config = pydantic.ConfigDict(arbitrary_types_allowed=True)

    id_: int
    name: str
```

  {{< /tab >}}

{{< /tabs >}}

<div class="hx:mt-16"></div>

{{< hextra/hero-section >}}
  Built to be trusted
{{< /hextra/hero-section >}}

Generated code is held to the same standard as hand-written code:

- **Type-checked.** Every generated file passes pyright in strict mode and ruff.
- **Tested against real databases.** The suite runs the generated code against
  live PostgreSQL and SQLite, across every driver and model type.
- **Deterministic.** Output is byte-identical between runs, and CI fails if a
  change would silently alter what you get.

<div class="hx:mt-16"></div>

{{< hextra/hero-section >}}
  Used by
{{< /hextra/hero-section >}}

Projects running their database layer on generated code:

<div class="hx:mt-6 hx:flex hx:flex-wrap hx:gap-4 hx:items-center">
  <a href="https://nmarkov.xyz/" target="_blank" rel="noreferrer" class="hx:flex hx:items-center hx:gap-4">
    <img src="images/used-by/nmarkov.png" alt="nMarkov logo" width="48" height="48" style="border-radius: 8px;" />
    <span><strong>nMarkov</strong> - a Discord chatbot that learns from your server's messages and generates its own.</span>
  </a>
</div>

Using `sqlc-gen-better-python` in your project?
[Open an issue](https://github.com/rayakame/sqlc-gen-better-python/issues) to get
listed here.

<div class="hx:mt-16"></div>

{{< hextra/hero-section >}}
  Set up in three steps
{{< /hextra/hero-section >}}

Point `sqlc` at the plugin, pick a driver and a model type, and generate:

```yaml
# sqlc.yaml
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
```

```bash
sqlc generate
```

<div class="hx:mt-6"></div>

{{< hextra/hero-button text="Read the Getting Started guide" link="docs/getting-started" >}}

<div class="hx:mt-12"></div>
