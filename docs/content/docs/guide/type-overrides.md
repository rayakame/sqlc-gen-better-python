---
title: Type overrides
weight: 60
prev: /docs/guide/enums
next: /docs/guide/converters
---

An override replaces the Python type the plugin would pick for a column with one
of your own. You can target a specific column or every column of a given SQL
type.

## Specification

Each entry in `overrides` matches columns in one of two ways (exactly one is
required):

- **`column`** - a specific column, as `[schema.]table.column` (wildcards
  allowed).
- **`db_type`** - every column of a SQL type, e.g. `text`.

and supplies a `py_type` describing the replacement type and how to import it
(`import`, `package`, `type`). Full schema in the
[reference](/docs/reference/configuration-options#override).

## Example

Override one `text` column with `collections.UserString`:

```yaml
options:
  overrides:
    - column: test_type_override.text_test
      py_type:
        import: collections
        package: UserString
        type: UserString
```

The model field now uses your type:

```python
class TestTypeOverride(msgspec.Struct):
    id_: int
    text_test: UserString | None
```

and query functions convert on the way in and out:

```python
# read: construct your type from the column value
return models.TestTypeOverride(id_=row[0], text_test=UserString(row[1]) if row[1] is not None else None)

# write: convert back to the column's default type (str for text)
await conn.execute(INSERT_TYPE_OVERRIDE, id_, str(text_test) if text_test is not None else None)
```

## How conversion works

An override converts by **calling the type**. On read the plugin builds the value
with `YourType(column_value)`; on write it calls the column's default type
(`str(...)` here) to turn it back. Nullable columns are guarded, so a SQL `NULL`
stays `None` and the type is never called with `None`.

This is why an override works for anything constructible from its wire value
(`UserString(str)`, `PurePosixPath(str)`, ...) but *not* for things like JSON,
where `MyModel(json_string)` is meaningless. For those, use a
[converter](/docs/guide/converters), which names your own encode/decode functions
instead.

## Matching by SQL type

Use `db_type` to override every column of a type at once:

```yaml
overrides:
  - db_type: text
    py_type:
      import: collections
      package: UserString
      type: UserString
```

When both a `column` and a `db_type` override could match a column, the `column`
override wins.

{{< callout type="warning" >}}
  Column overrides do not attach to `ANY($1::type[])` array parameters - sqlc does
  not link a column override to those. Only `db_type` overrides (and converters
  matched by `db_type`) reach them.
{{< /callout >}}
