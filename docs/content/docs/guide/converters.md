---
title: Converters
weight: 70
prev: /docs/guide/type-overrides
next: /docs/guide/working-with-json
---

A [type override](/docs/guide/type-overrides) replaces a column's Python type,
but it converts by *calling that type* (`Preferences(value)`), which cannot
express JSON or model (de)serialization. A **converter** names two of your own
functions instead, used whenever the column is read from or written to the
database.

## Specification

A converter is declared under `converters` and referenced *by an override*:

| Field | Description |
|---|---|
| `name` | Unique name, referenced by an override's `converter`. |
| `py_type` | The Python type the converter produces and accepts (`import`/`package`/`type`). |
| `to_db` | Dotted path to a function turning the Python value into the column's wire type. |
| `from_db` | Dotted path to a function turning the wire value into the Python type. |

```yaml
options:
  converters:
    - name: prefs
      py_type:
        import: myapp.models
        package: Preferences
        type: Preferences
      to_db: myapp.converters.encode_preferences
      from_db: myapp.converters.decode_preferences
  overrides:
    - db_type: jsonb            # every jsonb column
      converter: prefs
    - column: users.preferences # or one specific column
      converter: prefs
```

Because the converter is referenced by an override, it applies to whichever
columns that override matches, and `converter` replaces `py_type` on the
override.

## Generated code

The column is typed with your `py_type`, and the plugin inserts calls to your
functions on both sides:

```python
# read - from_db on each column
return models.User(id_=row[0], preferences=myapp.converters.decode_preferences(row[1]))

# write - to_db on each parameter
await conn.execute(CREATE_USER, id_, myapp.converters.encode_preferences(preferences))
```

Nullable columns are guarded, so your functions never receive `None`:

```python
maybe_prefs=myapp.converters.decode_preferences(row[2]) if row[2] is not None else None
```

and list columns convert element-wise:

```python
[myapp.converters.encode_label(v) for v in labels]
```

## Rules

- **Both directions are required.** `to_db` and `from_db` are dotted paths to
  module-level functions; the module is imported and the function called fully
  qualified, so the names can never collide with generated code.
- **Wire type.** `to_db` must return the type the column would have had *without*
  the override (`jsonb` is a `str`, `bytea` a `memoryview` - see
  [type mappings](/docs/reference/type-mappings)), and `from_db` receives that
  same type. For an unrecognised SQL type there is no such type, so `to_db` must
  return one the driver accepts.
- **Your functions never see `None`.** A SQL `NULL` stays `None` without calling
  the converter.
- **List columns convert element-wise.**

{{< callout type="warning" >}}
  To reach an `ANY($1::type[])` array parameter, match the override with
  `db_type` rather than `column` - sqlc does not link column overrides to those
  parameters.
{{< /callout >}}

## Next

Converters are most often reached for with JSON columns. See
[Working with JSON](/docs/guide/working-with-json) for a complete walkthrough
using `msgspec` structs.
