---
title: SQLite type conversion
weight: 90
prev: /docs/guide/docstrings
next: /docs/guide/naming
---

SQLite only stores a handful of native types, so dates, decimals, booleans and
blobs have to be translated in both directions. For the `sqlite3` and `aiosqlite`
drivers the generated code registers **adapters** (Python value -> SQLite) and
**converters** (SQLite value -> Python) for you.

## Enable declared-type parsing

Converters only run when the connection parses declared column types, so you must
open the connection with `PARSE_DECLTYPES`:

{{< tabs >}}

  {{< tab name="sqlite3" >}}
```python
conn = sqlite3.connect("app.db", detect_types=sqlite3.PARSE_DECLTYPES)
```
  {{< /tab >}}

  {{< tab name="aiosqlite" >}}
```python
import sqlite3

import aiosqlite

async with aiosqlite.connect("app.db", detect_types=sqlite3.PARSE_DECLTYPES) as conn:
    ...
```
  {{< /tab >}}

{{< /tabs >}}

{{< callout type="warning" >}}
  Without `detect_types=PARSE_DECLTYPES` the converters never fire and you get raw
  strings and bytes back instead of `datetime`, `Decimal`, `bool`, or `memoryview`
  values. This is the most common SQLite setup mistake.
{{< /callout >}}

## What gets converted

| Declared SQL type | Python type | Sent to SQLite as | Read back with |
|---|---|---|---|
| `date` | `datetime.date` | `str` | `datetime.date.fromisoformat` |
| `datetime`, `timestamp` | `datetime.datetime` | `str` | `datetime.datetime.fromisoformat` |
| `decimal`, `decimal(p,s)` | `decimal.Decimal` | `str` | `decimal.Decimal` |
| `bool`, `boolean` | `bool` | `int` | `bool(int(...))` |
| `blob` | `memoryview` | `bytes` | `memoryview` |

The converter is keyed on the column's **declared type name**, which is why the
declared types in your schema matter for SQLite.

## Only what each module needs

Registration is emitted per generated module, and only for what that module
actually uses:

- A type used as a **parameter** gets its **adapter** registered.
- A type used in a **returned column** gets its **converter** registered.

{{< callout type="info" >}}
  `register_converter` is process-global in Python's `sqlite3`, so a registration
  made by one generated module applies to every connection in the process.
  Registering only what a module needs avoids installing conversions it never
  uses - it does not isolate modules from one another, and two modules
  registering the same declared type will share that conversion.
{{< /callout >}}

## Interaction with overrides

A column with a [type override](/docs/guide/type-overrides) or
[converter](/docs/guide/converters) is converted inline by the generated code, so
**no** SQLite converter is registered for it - registering one would convert the
value twice. Overridden values used as *parameters* still need the adapter, so
that half is still registered.

## Faster date parsing

`speedups: true` swaps the standard-library date parsing for
[`ciso8601`](https://github.com/closeio/ciso8601):

```yaml
options:
  speedups: true
```

This affects the `date` and `datetime`/`timestamp` converters only - `decimal`,
`bool` and `blob` are unchanged. It adds `ciso8601` as a runtime dependency of the
generated code, so only enable it if you install that package.
