---
title: Configuration options
weight: 10
prev: /docs/reference
next: /docs/reference/type-mappings
---

Every option the plugin accepts, set under `options:` in the plugin's `codegen`
block. See [Configuration](/docs/guide/configuration) for where this block lives
in `sqlc.yaml`.

## Options

`package`, `sql_driver`, and `emit_init_file` are required; everything else is
optional.

| Option | Type | Default | Description |
|---|---|---|---|
| `package` | string | *required* | Name of the generated package. |
| `sql_driver` | string | *required* | One of `asyncpg`, `aiosqlite`, `sqlite3`. Must match the engine (`asyncpg` -> `postgresql`; the other two -> `sqlite`). |
| `emit_init_file` | bool | *required* | Whether to emit an `__init__.py` in the package. Must be set explicitly. Set `false` only if the package already has one. |
| `model_type` | string | `dataclass` | One of `dataclass`, `attrs`, `msgspec`, `pydantic`. See [Model types](/docs/guide/model-types). |
| `initialisms` | list[string] | `["id"]` | Identifier segments to upper-case, e.g. `app_id` -> `AppID`. |
| `emit_exact_table_names` | bool | `false` | If `true`, model names mirror table names; otherwise plural table names are singularized. |
| `emit_classes` | bool | `false` | If `true`, query functions become methods on a `Querier` class. See [Writing queries](/docs/guide/writing-queries). |
| `inflection_exclude_table_names` | list[string] | `[]` | Table names to leave un-singularized (only used when `emit_exact_table_names` is `false`). |
| `omit_unused_models` | bool | `false` | If `true`, tables not referenced by any query are not turned into models. |
| `omit_typechecking_block` | bool | `false` | If `true`, non-builtin types are imported at module level instead of inside a `typing.TYPE_CHECKING` block. |
| `docstrings` | string | `none` | One of `none`, `google`, `numpy`, `pep257`. See [Docstrings](/docs/guide/docstrings). |
| `docstrings_emit_sql` | bool | `true` | Include each query's SQL in its docstring. Ignored when `docstrings` is `none`. |
| `query_parameter_limit` | int | *unset* | When set to a non-negative value, queries with more parameters than the limit bundle them into a single `params` argument. Unset or negative never bundles (except `:copyfrom`, which always uses a params class). |
| `omit_kwargs_limit` | int | `0` | Queries with this many parameters or fewer do not require keyword arguments. `0` makes every parameter keyword-only. Must not be negative. |
| `speedups` | bool | `false` | Use faster third-party libraries for type conversion. Currently affects `sqlite3`/`aiosqlite` only (uses `ciso8601`). See [SQLite type conversion](/docs/guide/sqlite-type-conversion). |
| `overrides` | list[Override] | `[]` | Replace the Python type of matching columns. See [Type overrides](/docs/guide/type-overrides). |
| `converters` | list[Converter] | `[]` | Named encode/decode function pairs referenced by an override. See [Converters](/docs/guide/converters). |
| `indent_char` | string | `" "` (space) | Character used for one indent step. |
| `chars_per_indent_level` | int | `4` | Number of `indent_char`s per indent level. |
| `debug` | bool | `false` | Emit a `log.txt` debug log during `sqlc generate`. |

## Object schemas

Some options take structured objects rather than scalars.

### `py_type`

Describes a Python type and how to import it. Used by an `override` and by a
`converter`.

| Field | Type | Description |
|---|---|---|
| `type` | string | *required* - the type expression used in annotations, e.g. `UserString` or `collections.UserString`. |
| `import` | string | Module to import, e.g. `collections`. |
| `package` | string | Name to import from `import` (`from <import> import <package>`). If empty, emits `import <import>`. |

### `override`

An entry in `overrides`. Matches columns either by SQL type or by column name;
specifying both, or neither, is an error.

| Field | Type | Description |
|---|---|---|
| `db_type` | string | Match a SQL type exactly (case-insensitive), e.g. `text`, `pg_catalog.int4`. Mutually exclusive with `column`. |
| `column` | string | Match `[catalog.][schema.]table.column`; wildcards are supported. Mutually exclusive with `db_type`. |
| `py_type` | py_type | Replacement type. Required unless `converter` is set. |
| `converter` | string | Name of a `converters` entry, which supplies the type and its encode/decode functions instead of `py_type`. |

When both a `column` and a `db_type` override could match, the column match wins.

### `converter`

An entry in `converters`. See [Converters](/docs/guide/converters).

| Field | Type | Description |
|---|---|---|
| `name` | string | *required* - unique name, referenced by an override's `converter`. |
| `py_type` | py_type | *required* - the Python type the converter produces and accepts. |
| `to_db` | string | *required* - dotted path to a function that turns the Python value into the column's wire type. |
| `from_db` | string | *required* - dotted path to a function that turns the wire value into the Python type. |
