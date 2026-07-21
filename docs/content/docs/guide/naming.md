---
title: Naming and identifiers
weight: 100
prev: /docs/guide/sqlite-type-conversion
---

SQL identifiers are not always valid Python names, and SQL naming conventions are
not Python's. This page covers how table, column, and enum names are turned into
Python names, and how invalid ones are sanitized.

## Table names to class names

Table names are converted to `CamelCase` and singularized, so a table of many
rows yields a class describing one row:

| Table | Model |
|---|---|
| `test_field_namings` | `TestFieldNaming` |
| `test_invalid_identifiers` | `TestInvalidIdentifier` |

Two options change this:

- **`emit_exact_table_names: true`** - skip singularization; model names mirror
  table names exactly.
- **`inflection_exclude_table_names`** - a list of tables to leave
  un-singularized while everything else is still singularized. Entries match both
  bare and schema-qualified names.

## Initialisms

Segments listed in `initialisms` are fully upper-cased when building class names,
so you get `AppID` rather than `AppId`. It defaults to `["id"]`:

```yaml
options:
  initialisms: ["id", "url", "api"]
```

## Column names to field names

Column names stay `snake_case`. Names that collide with Python keywords or
builtins get a trailing underscore - which is why an `id` column becomes `id_`:

```python
class TestFieldNaming(msgspec.Struct):
    id_: int
    outputs: str
```

## Sanitization

Identifiers that are not valid Python names are rewritten. Invalid characters
become underscores, and prefixes are added where a name would otherwise be
illegal:

```sql
CREATE TABLE IF NOT EXISTS test_invalid_identifiers
(
    id          bigint PRIMARY KEY NOT NULL,
    "3p%"       text,
    "new notes" text NOT NULL,
    "%pct"      text
);
```

```python
class TestInvalidIdentifier(msgspec.Struct):
    id_: int
    column_3p_: str | None
    new_notes: str
    column__pct: str | None
```

The rules at work:

| Situation | Rule | Example |
|---|---|---|
| Invalid characters | Replaced with `_` | `new notes` -> `new_notes` |
| Column starts with a digit | Prefixed with `column_` | `3p%` -> `column_3p_` |
| Field starts with `_` | Prefixed with `column_` | `%pct` -> `column__pct` |
| Table yields a digit-leading, keyword, or empty class name | Prefixed with `Model` | `3rd_party_stats` -> `Model3RdPartyStat` |
| Enum value starts with a digit or `_` | Prefixed with `VALUE_` | `24h` -> `VALUE_24H` |
| Two results collide | Numeric suffix | `outputs`, `outputs` -> `outputs`, `outputs_2` |

A leading underscore is prefixed rather than kept because `attrs` and `pydantic`
treat underscore-leading fields as private.

The collision rule also covers queries that select the same column name twice -
for example a self-join selecting `a.outputs, b.outputs` produces a row class with
`outputs` and `outputs_2`.

{{< callout type="warning" >}}
  Sanitization checks characters with Go's Unicode tables, which differ from
  Python's identifier rules in exotic cases - characters that may not *start* an
  identifier in Python, or two names Python normalizes to the same identifier via
  NFKC. Such schemas are not detected. Stick to ASCII identifiers if in doubt.
{{< /callout >}}
