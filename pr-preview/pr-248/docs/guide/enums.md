# Enums

PostgreSQL enum types and MySQL inline `ENUM(...)` columns become
`enum.StrEnum` classes in a generated `enums.py` module. Columns of that type
are annotated with the class, and values read from the database are coerced
into it.

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

Member names are uppercased and invalid identifiers sanitized: `24h` becomes
`VALUE_24H`, `_hidden` becomes `VALUE__HIDDEN`. The string *values* are
untouched, so round-tripping to the database is exact. Full rules in
[Naming and identifiers](/docs/guide/naming).

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

## MySQL

MySQL has no named enum types - `ENUM` is declared inline on the column, so
the class is named after the table and column:

```sql
CREATE TABLE test_enum_override
(
    id        bigint PRIMARY KEY NOT NULL,
    mood_test enum('sad','ok','happy') NOT NULL
);
```

generates `TestEnumOverrideMoodTest`, used exactly like a PostgreSQL enum
class. `SET(...)` columns generate a class the same way.

{{< callout type="warning" >}}
  A `SET` column can hold several members at once, but the database returns
  them as one comma-joined string - coercing `"alpha,beta"` into the enum
  class raises `ValueError`. Only single-valued sets round-trip; for
  multi-valued sets add a [type override](/docs/guide/type-overrides) to
  `str`.
{{< /callout >}}

{{< callout type="info" >}}
  SQLite has no native enum type, so `enums.py` is generated for the
  PostgreSQL and MySQL drivers only.
{{< /callout >}}

