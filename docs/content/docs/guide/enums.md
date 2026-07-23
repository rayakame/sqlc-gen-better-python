---
title: Enums
weight: 50
prev: /docs/guide/writing-queries
next: /docs/guide/type-overrides
---

PostgreSQL enum types become `enum.StrEnum` classes in a generated `enums.py`
module. Columns of that type are annotated with the class, and values read from
the database are coerced into it.

## Example

```sql
CREATE TYPE test_mood AS ENUM ('sad', 'ok', 'happy', '24h', '_hidden');
```

generates:

```python
class TestMood(enum.StrEnum):
    SAD = "sad"
    OK = "ok"
    HAPPY = "happy"
    VALUE_24H = "24h"
    VALUE__HIDDEN = "_hidden"
```

Notice the member names are uppercased, and values that are not valid Python
identifiers are sanitized: the digit-leading `24h` becomes `VALUE_24H` and the
underscore-leading `_hidden` becomes `VALUE__HIDDEN`. The string *values* are
untouched, so round-tripping to the database is exact. See
[Naming and identifiers](/docs/guide/naming) for the full sanitization rules.

## Using enums

A column of an enum type is annotated with the generated class, and nullable
columns get `| None`:

```sql
CREATE TABLE test_enum_types
(
    id   int PRIMARY KEY NOT NULL,
    mood test_mood       NOT NULL,
    maybe_mood test_mood
);
```

```python
class TestEnumType(msgspec.Struct):
    id_: int
    mood: enums.TestMood
    maybe_mood: enums.TestMood | None
```

Query functions coerce database values into these classes automatically, so you
get real `TestMood` members back, not bare strings.

## Enums in other schemas

Enums in a non-default schema get schema-qualified class names so same-named
enums never collide - for example `custom.mood` becomes `CustomMood`, distinct
from a `public.mood` that would become `Mood`.

{{< callout type="info" >}}
  Enum classes are a PostgreSQL feature - SQLite has no native enum type, so this
  applies to the PostgreSQL drivers (`asyncpg` and `psycopg_async`).
{{< /callout >}}
