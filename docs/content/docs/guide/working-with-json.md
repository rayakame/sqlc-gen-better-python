---
title: Working with JSON
weight: 75
prev: /docs/guide/converters
next: /docs/guide/docstrings
---

A `json`/`jsonb` column maps to `str` by default - correct, but untyped: you get
a string and parse it yourself everywhere. With a [converter](/docs/guide/converters)
and a `msgspec.Struct` you get a fully typed, validated Python object instead,
and the plugin inserts the encode/decode calls for you.

## Why JSON needs a converter

A [type override](/docs/guide/type-overrides) converts by *calling the type*, so
it would generate `Preferences(row[2])` - handing a JSON string to a struct
constructor, which is meaningless. JSON needs a real parse step, which is exactly
what a converter's `from_db`/`to_db` pair provides.

## The wire type contract

This is the one rule to internalize:

{{< callout type="info" >}}
  A `jsonb` (or `json`) column's wire type is **`str`**. So `to_db` must *return*
  `str`, and `from_db` *receives* `str`.
{{< /callout >}}

That matters for `msgspec`, because `msgspec.json.encode()` returns **`bytes`** -
you must `.decode()` it. `msgspec.json.decode()` happily accepts either.

## A complete example

A `users` table with a nested preferences document:

```sql
CREATE TABLE users
(
    id          bigint PRIMARY KEY NOT NULL,
    name        text               NOT NULL,
    preferences jsonb              NOT NULL
);
```

```sql
-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: CreateUser :exec
INSERT INTO users (id, name, preferences) VALUES ($1, $2, $3);
```

{{< tabs >}}

  {{< tab name="models.py" >}}
```python
import msgspec


class Theme(msgspec.Struct):
    name: str
    dark_mode: bool = False


class Preferences(msgspec.Struct):
    theme: Theme
    languages: list[str] = msgspec.field(default_factory=list)
    email_opt_in: bool | None = None
```

Structs nest freely - `Preferences` holds a `Theme`, plus a list and an optional
field. `msgspec` handles all of it.
  {{< /tab >}}

  {{< tab name="converters.py" >}}
```python
import msgspec

from myapp.models import Preferences


def encode_preferences(value: Preferences) -> str:
    # jsonb's wire type is str, and msgspec encodes to bytes.
    return msgspec.json.encode(value).decode()


def decode_preferences(value: str) -> Preferences:
    return msgspec.json.decode(value, type=Preferences)
```

The `type=` argument is what makes decoding *typed* - msgspec builds real
`Preferences` objects and validates the payload while parsing.
  {{< /tab >}}

  {{< tab name="sqlc.yaml" >}}
```yaml
options:
  model_type: "msgspec"
  converters:
    - name: preferences
      py_type:
        import: myapp.models
        package: Preferences
        type: Preferences
      to_db: myapp.converters.encode_preferences
      from_db: myapp.converters.decode_preferences
  overrides:
    - column: users.preferences
      converter: preferences
```
  {{< /tab >}}

{{< /tabs >}}

### What gets generated

The model field is typed as your struct, and both query functions call your
converter:

```python
class User(msgspec.Struct):
    id_: int
    name: str
    preferences: Preferences


async def get_user(conn: ConnectionLike, *, id_: int) -> models.User | None:
    row = await conn.fetchrow(GET_USER, id_)
    if row is None:
        return None
    return models.User(id_=row[0], name=row[1], preferences=myapp.converters.decode_preferences(row[2]))


async def create_user(conn: ConnectionLike, *, id_: int, name: str, preferences: Preferences) -> None:
    await conn.execute(CREATE_USER, id_, name, myapp.converters.encode_preferences(preferences))
```

### Using it

```python
await create_user(
    conn,
    id_=1,
    name="ada",
    preferences=Preferences(theme=Theme(name="solarized", dark_mode=True), languages=["en"]),
)

user = await get_user(conn, id_=1)
user.preferences.theme.dark_mode  # True - typed all the way down
```

No `json.loads`, no `dict["theme"]["dark_mode"]`, and pyright checks every access.

## Several JSON columns with different shapes

Match **per column**, one converter per struct:

```yaml
options:
  converters:
    - name: preferences
      py_type: { import: myapp.models, package: Preferences, type: Preferences }
      to_db: myapp.converters.encode_preferences
      from_db: myapp.converters.decode_preferences
    - name: audit_meta
      py_type: { import: myapp.models, package: AuditMeta, type: AuditMeta }
      to_db: myapp.converters.encode_audit_meta
      from_db: myapp.converters.decode_audit_meta
  overrides:
    - column: users.preferences
      converter: preferences
    - column: audit_log.meta
      converter: audit_meta
```

{{< callout type="warning" >}}
  Reaching for `db_type: jsonb` applies **one** struct to *every* `jsonb` column
  in your schema. That is only what you want when all of them genuinely share a
  shape - otherwise match by `column`.
{{< /callout >}}

## Nullable and array JSON columns

- **Nullable** (`preferences jsonb` without `NOT NULL`): the field becomes
  `Preferences | None` and the plugin guards the call, so `decode_preferences`
  never receives `None`.
- **Arrays** (`jsonb[]`): the field becomes
  `collections.abc.Sequence[Preferences]` and conversion happens element-wise -
  your functions still deal with a single value at a time. To also cover an
  `ANY($1::jsonb[])` parameter, match with `db_type: jsonb`.

## Validation

`msgspec.json.decode(..., type=Preferences)` validates while parsing and raises
`msgspec.ValidationError` when a stored document does not match the struct (a
missing required field, a wrong type). That is usually what you want - it surfaces
bad data at the boundary. If you would rather tolerate legacy rows, give fields
defaults or widen them to `| None`.

## Other libraries

The converter contract is just "return the wire type, accept the wire type", so
any library works - as long as your model and your converter use the same one.
The examples below assume `Theme`/`Preferences` are defined with that library
rather than as `msgspec.Struct`s.

{{< tabs >}}

  {{< tab name="stdlib json" >}}
With `Theme` and `Preferences` defined as `dataclasses.dataclass`:

```python
import dataclasses
import json

from myapp.models import Preferences, Theme


def encode_preferences(value: Preferences) -> str:
    return json.dumps(dataclasses.asdict(value))


def decode_preferences(value: str) -> Preferences:
    raw = json.loads(value)
    return Preferences(
        theme=Theme(**raw["theme"]),
        languages=raw.get("languages", []),
        email_opt_in=raw.get("email_opt_in"),
    )
```

No dependencies, but no validation - and nested objects must be rebuilt by hand,
since `json.loads` leaves `theme` as a plain `dict`.
  {{< /tab >}}

  {{< tab name="pydantic" >}}
With `Theme` and `Preferences` defined as `pydantic.BaseModel`:

```python
from myapp.models import Preferences


def encode_preferences(value: Preferences) -> str:
    return value.model_dump_json()


def decode_preferences(value: str) -> Preferences:
    return Preferences.model_validate_json(value)
```

Full validation, nested models rebuilt for you, and `model_validate_json` parses
straight from the string.
  {{< /tab >}}

{{< /tabs >}}

The same approach works for the SQLite drivers: a `json` column is `str` there
too, so the identical converter applies.
