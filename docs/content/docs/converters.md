---
title: Converters
weight: 3
prev: getting-started
---

A [type override](https://github.com/rayakame/sqlc-gen-better-python#type-overrides)
replaces a column's Python type, but it converts by *calling that type*
(`Preferences(value)`), which cannot express JSON or model (de)serialization. A
**converter** names two of your own functions instead, used whenever the column
is read from or written to the database.

## Defining a converter

Declare the converter under `converters`, then reference it from an override:

```yaml
options:
  # ...
  converters:
    - name: prefs
      py_type:
        import: myapp.models
        package: Preferences
        type: Preferences
      to_db: myapp.converters.encode_preferences
      from_db: myapp.converters.decode_preferences
  overrides:
    - db_type: jsonb          # every jsonb column
      converter: prefs
    - column: users.preferences   # or one specific column
      converter: prefs
```

The generated code calls your functions on both sides:

```python
# read
preferences = myapp.converters.decode_preferences(row[1])
# write
await conn.execute(CREATE_USER, id_, myapp.converters.encode_preferences(preferences))
```

## Rules

{{< callout type="info" >}}
  A converter is referenced *by an override*, so it applies to whichever columns
  that override matches. `converter` replaces `py_type` on the override.
{{< /callout >}}

- **Both directions are required.** `to_db` and `from_db` are dotted paths to
  module-level functions; the module is imported and the function called fully
  qualified, so the names can never collide with generated code.
- **Wire type.** `to_db` must return the type the column would have had *without*
  the override (`jsonb` is a `str`, `bytea` a `memoryview`), and `from_db`
  receives that same type. For an unrecognised SQL type there is no such type, so
  `to_db` must return one the driver accepts.
- **Your functions never see `None`.** Nullable columns are guarded, so a SQL
  `NULL` stays `None` without calling the converter.
- **List columns convert element-wise.**

## Example: msgspec structs to and from JSON

Because `msgspec.json.decode` needs a `type=` argument and returns `bytes`, you
wrap it in two small functions that match the wire type of a `jsonb` column
(`str`):

{{< tabs >}}

  {{< tab name="converters.py" >}}
```python
import msgspec

from myapp.models import User


def encode_user(value: User) -> str:
    # jsonb's wire type is str, so decode msgspec's bytes back to str.
    return msgspec.json.encode(value).decode()


def decode_user(value: str) -> User:
    return msgspec.json.decode(value, type=User)
```
  {{< /tab >}}

  {{< tab name="models.py" >}}
```python
import msgspec


class User(msgspec.Struct):
    name: str
    groups: set[str]
    email: str | None = None
```
  {{< /tab >}}

  {{< tab name="sqlc.yaml" >}}
```yaml
converters:
  - name: user
    py_type:
      import: myapp.models
      package: User
      type: User
    to_db: myapp.converters.encode_user
    from_db: myapp.converters.decode_user
overrides:
  - db_type: jsonb
    converter: user
```
  {{< /tab >}}

{{< /tabs >}}

A `get_user` query then returns a fully typed `User` (set field and all), and a
`create_user` query accepts one directly &mdash; the plugin inserts the
`decode_user` / `encode_user` calls for you.
