---
title: Type mappings
description: >-
  The built-in SQL-to-Python type mappings for PostgreSQL, SQLite, and MySQL, and what nullable columns and arrays map to.
weight: 20
prev: /docs/reference/configuration-options
next: /docs/reference/feature-support
---

How the plugin maps SQL types to Python types, per engine. These are the
built-in defaults; a [type override](/docs/guide/type-overrides) or
[converter](/docs/guide/converters) replaces the type for a specific column or
SQL type.

## How a column's type is built

The tables below give the **base** Python type for a SQL type. Two modifiers are
applied on top, based on the column:

- **Nullable** (the column is not `NOT NULL`): `| None` is appended, e.g.
  `str | None`.
- **Array / list** (a Postgres array column, or a `sqlc.slice` parameter): the
  base type is wrapped as `collections.abc.Sequence[T]`.

Both can combine: a nullable `bytea[]` becomes
`collections.abc.Sequence[memoryview] | None`.

## PostgreSQL

| SQL type | Python type |
|---|---|
| `smallint`, `integer`, `bigint` (and `int2`/`int4`/`int8`, `smallserial`/`serial`/`bigserial`, `pg_catalog.*` forms) | `int` |
| `real`, `double precision` (`float4`/`float8`) | `float` |
| `numeric` | `decimal.Decimal` |
| `money` | `str` |
| `boolean` (`bool`) | `bool` |
| `json`, `jsonb` | `str` |
| `bytea` | `memoryview` |
| `date` | `datetime.date` |
| `time`, `timetz` | `datetime.time` |
| `timestamp`, `timestamptz` | `datetime.datetime` |
| `interval` | `datetime.timedelta` |
| `text`, `varchar`, `char`/`bpchar`, `citext` | `str` |
| `uuid` | `uuid.UUID` |
| `inet`, `cidr`, `macaddr`, `macaddr8` | `str` |
| `ltree`, `lquery`, `ltxtquery` | `str` |
| a user-defined `ENUM` type | the generated [enum class](/docs/guide/enums), e.g. `enums.Mood` |
| anything else | `typing.Any` |

{{< callout type="info" >}}
  A SQL type the plugin does not recognize falls back to `typing.Any`. If that
  happens for a type you use often, add a [type override](/docs/guide/type-overrides)
  or [converter](/docs/guide/converters) so it gets a real Python type.
{{< /callout >}}

## SQLite

SQLite type names are matched case-insensitively, by exact name first and then by
prefix (so `varchar(255)` matches `varchar`, and `decimal(10,5)` matches
`decimal`).

| SQL type | Python type |
|---|---|
| `int`, `integer`, `tinyint`, `smallint`, `mediumint`, `bigint`, `int2`, `int8`, `bigserial` | `int` |
| `real`, `double`, `double precision`, `float`, `numeric` | `float` |
| `decimal`, `decimal(p,s)` | `decimal.Decimal` |
| `bool`, `boolean` | `bool` |
| `date` | `datetime.date` |
| `datetime`, `timestamp` | `datetime.datetime` |
| `text`, `clob`, `json`, `character`, `varchar`, `varyingcharacter`, `nchar`, `nativecharacter`, `nvarchar` | `str` |
| `blob` | `memoryview` |
| anything else | `typing.Any` |

For the two SQLite drivers, several of these types also need runtime adapters and
converters (registered in the generated code) to round-trip correctly - see
[SQLite type conversion](/docs/guide/sqlite-type-conversion).

## MySQL

MySQL type names are matched case-insensitively.

| SQL type | Python type |
|---|---|
| `tinyint(1)`, `bool`, `boolean` | `bool` |
| `tinyint`, `smallint`, `mediumint`, `int`, `integer`, `bigint`, `year`, `serial` (signed or `unsigned`) | `int` |
| `float`, `double`, `double precision`, `real` | `float` |
| `decimal`, `dec`, `fixed`, `numeric` | `decimal.Decimal` |
| `char`, `varchar`, `tinytext`, `text`, `mediumtext`, `longtext` | `str` |
| `binary`, `varbinary`, `tinyblob`, `blob`, `mediumblob`, `longblob`, `bit` | `memoryview` |
| `date` | `datetime.date` |
| `datetime`, `timestamp` | `datetime.datetime` |
| `time` | `datetime.timedelta` |
| `json` | `str` |
| an inline `ENUM(...)` or `SET(...)` column | the generated [enum class](/docs/guide/enums#mysql) - only single-valued sets round-trip |
| anything else | `typing.Any` |

{{< callout type="info" >}}
  Two mappings follow what the PyMySQL-family drivers actually return rather
  than the SQL standard: `time` is a `datetime.timedelta` (MySQL `TIME` values
  span more than 24 hours), and `tinyint(1)` is `bool` only when the display
  width is literally 1 in the DDL - a plain `tinyint` stays `int`.
{{< /callout >}}
