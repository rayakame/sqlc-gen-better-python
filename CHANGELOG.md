# Changelog
All notable changes to this project will be documented in this file.


## v0.6.0 - 2026-07-21
### Added
* [#186](https://github.com/rayakame/sqlc-gen-better-python/pull/186) Converters: a named pair of your own `to_db` / `from_db` functions that serialize and deserialize a column value, referenced by an override. Overrides could only convert by calling the type itself, which does not work for JSON and similar formats. The functions receive and return the type the column would have had without the override, are never called with `None` (nullable columns stay guarded), and convert list columns element-wise. ([Rayakame](https://github.com/Rayakame))
* [#208](https://github.com/rayakame/sqlc-gen-better-python/pull/208) Documentation site at https://rayakame.github.io/sqlc-gen-better-python/, with a getting started page, a guide covering configuration, drivers, model types, writing queries, enums, type overrides, converters, working with JSON, docstrings, SQLite type conversion and naming, and a reference for every configuration option, the SQL to Python type mappings of both engines and per-driver feature support. Every generated code example is taken from the committed test fixtures. The README now links to the site instead of documenting each feature inline. ([Rayakame](https://github.com/Rayakame))
### Fixed
* [#185](https://github.com/rayakame/sqlc-gen-better-python/pull/185) Overrides on columns whose SQL type the plugin does not recognise are passed to the driver unconverted, so their Python type is now included in `QueryResultsArgsType`; previously pyright strict rejected such a value as a `:many` query parameter. Types used only as parameter annotations also stay lazy instead of being forced into a runtime import. ([Rayakame](https://github.com/Rayakame))
* [#211](https://github.com/rayakame/sqlc-gen-better-python/pull/211) `sqlc.slice` works on the sqlite drivers: generated functions now expand the `/*SLICE:name*/` placeholder at call time, one `?` per element (`NULL` for an empty sequence, so `IN (NULL)` matches no rows), and unpack the sequence into the positional arguments in SQL text order. Previously the placeholder was left in the SQL and binding the sequence raised `sqlite3.ProgrammingError`. Overridden and converted slice parameters keep their element-wise conversion. On PostgreSQL use `= ANY($1::type[])` instead, which sqlc itself intends there. ([Rayakame](https://github.com/Rayakame))

## v0.5.1 - 2026-07-20
### Added
* [#177](https://github.com/rayakame/sqlc-gen-better-python/pull/177) CI now runs the Go test suite and golangci-lint on every pull request, so the Go checks are no longer local-only conventions. ([Rayakame](https://github.com/Rayakame))
* [#177](https://github.com/rayakame/sqlc-gen-better-python/pull/177) Python 3.14 support: the test tooling accepts and CI now tests Python 3.14. This bumps the asyncpg floor to 0.31.0, the first release with Python 3.14 wheels. ([Rayakame](https://github.com/Rayakame))
* [#180](https://github.com/rayakame/sqlc-gen-better-python/pull/180) Go unit tests: internal/model, internal/config, internal/log, internal/types and internal/writer are now covered to 100% of statements, raising the Go coverage flag from 0.5% to about 43%. ([Rayakame](https://github.com/Rayakame))
* [#182](https://github.com/rayakame/sqlc-gen-better-python/pull/182) Go unit tests for internal/transform, internal/driver, internal/render and the handler complete the suite: overall Go statement coverage is now 99.8%, with every package at 100%. ([Rayakame](https://github.com/Rayakame))
### Changed
* [#177](https://github.com/rayakame/sqlc-gen-better-python/pull/177) The contributing guidelines have been rewritten from scratch to match the current development workflow: prerequisites, repository layout, the WASM build scripts, all nox sessions, the runtime test setup and the full loop for generator changes are now documented. changie is also set up as a Go tool, so `make changelog` works without installing anything. Closes issue 88 ([Rayakame](https://github.com/Rayakame))
* [#177](https://github.com/rayakame/sqlc-gen-better-python/pull/177) All pre-existing golangci-lint findings in the plugin source have been fixed; `make lint` now passes with zero issues. This is a pure internal refactor, generated code is unchanged. ([Rayakame](https://github.com/Rayakame))

## v0.5.0 - 2026-07-19
### Added
* [#148](https://github.com/rayakame/sqlc-gen-better-python/pull/148) Support for PostgreSQL enums - enum columns now generate an `enums.py` module with `str`-based enum classes, including schema-qualified naming and runtime value coercion ([rayakame](https://github.com/rayakame))
* [#148](https://github.com/rayakame/sqlc-gen-better-python/pull/148) New `model_type` option `pydantic` generating `pydantic.BaseModel` models (requires pydantic >= 2.9) ([rayakame](https://github.com/rayakame))
* [#148](https://github.com/rayakame/sqlc-gen-better-python/pull/148) `query_parameter_limit` is now implemented and opt-in: when set to a non-negative value, queries with more parameters than the limit take a single `params: <Query>Params` argument; unset or negative values keep parameters expanded ([rayakame](https://github.com/rayakame))
### Changed
* [#148](https://github.com/rayakame/sqlc-gen-better-python/pull/148) Complete rewrite of the code generator internals into a config/transform/render/driver pipeline ([rayakame](https://github.com/rayakame))
### Fixed
* [#148](https://github.com/rayakame/sqlc-gen-better-python/pull/148) `:many` query functions no longer subscript the generic at runtime (`QueryResults(...)` instead of `QueryResults[T](...)`). The subscripted call went through `typing`'s `_GenericAlias.__call__` and raised-and-swallowed an `AttributeError` on every invocation (the class uses `__slots__`, so `__orig_class__` was never actually stored), costing roughly 10x the plain constructor call while providing no runtime or type-checking benefit - the function's return annotation already carries the full `QueryResults[T]` type ([rayakame](https://github.com/rayakame))
* [#148](https://github.com/rayakame/sqlc-gen-better-python/pull/148) sqlite modules now register only the conversion halves they need: parameters register an adapter, non-overridden return columns register a converter. Previously every used conversion type registered both, so a module could install a global `register_converter` as a side effect - changing what overridden return columns receive under `PARSE_DECLTYPES`. Import resolution follows the same split: `ciso8601` is imported exactly when an emitted converter body references it, `datetime`/`decimal` stay in the `TYPE_CHECKING` block when the emitted setup does not reference them at runtime, and `from <package> import enums` is now also emitted when an overridden enum parameter is converted back via `enums.X(...)` at runtime ([rayakame](https://github.com/rayakame))
* [#148](https://github.com/rayakame/sqlc-gen-better-python/pull/148) Type overrides now convert correctly in every position: array parameters convert element-wise instead of passing the whole list into the driver-type constructor, and `:copyfrom` records convert overridden columns back to the driver type before `copy_records_to_table` (previously the raw override value was sent and failed to encode) ([rayakame](https://github.com/rayakame))
* [#148](https://github.com/rayakame/sqlc-gen-better-python/pull/148) sqlite SQL type names are now matched case-insensitively and with precision suffixes: a column declared `DATETIME` registers its adapter/converter pair (previously it was silently skipped and returned raw strings under `PARSE_DECLTYPES`), and `decimal(10,2)` is annotated `decimal.Decimal` instead of `typing.Any` ([rayakame](https://github.com/rayakame))
* [#148](https://github.com/rayakame/sqlc-gen-better-python/pull/148) A SQL column named `conn` (or `self` with `emit_classes`) no longer generates a duplicate function parameter (a `SyntaxError` in the generated module); the parameter is renamed `conn_2`. `:many` query functions with array parameters now type-check: `QueryResultsArgsType` gained a `Sequence` member. The `:execrows` docstring on the sqlite3 driver now correctly documents `-1` for non-DML statements ([rayakame](https://github.com/rayakame))
* [#148](https://github.com/rayakame/sqlc-gen-better-python/pull/148) `omit_typechecking_block: true` no longer crashes generated asyncpg modules at import time: the `ConnectionLike` alias is lazy (PEP 695), module-level import order satisfies E402, and `typing` is only imported where it is actually referenced. `models.py` is emitted even when all models are filtered out, matching v0.4.x. `inflection_exclude_table_names` entries match both the bare and the schema-qualified table name (v0.4.x only matched the qualified form) ([rayakame](https://github.com/rayakame))
* [#172](https://github.com/rayakame/sqlc-gen-better-python/pull/172) Row and Params class field names are no longer singularized: a plural column like `outputs` generated an `output` field on v0.4.x, so the field no longer matched the column. Only table names and embed fields (which hold one row of the joined table) are singularized. Note for upgrades: plural-named columns get their actual names back. Diagnosed by Mic92 in PR 164 ([Mic92](https://github.com/Mic92))
* [#173](https://github.com/rayakame/sqlc-gen-better-python/pull/173) Type overrides on SQL types the plugin does not know (e.g. a custom `JULIANDAY` sqlite type) no longer emit a `typing.Any(value)` conversion that raised `TypeError` at runtime; the values pass through unconverted. `db_type` overrides also match case-insensitively, so configs written against the DDL casing keep working. Reported by gazpachoking in issue 161 ([gazpachoking](https://github.com/gazpachoking))
* [#174](https://github.com/rayakame/sqlc-gen-better-python/pull/174) Quoted SQL identifiers that are not valid Python names no longer generate invalid Python. Column fields and parameters sanitize invalid characters to underscores (`"new notes"` -> `new_notes`), digit-leading names get a `column_` prefix (`"3p%"` -> `column_3p_`), table names that would produce a digit-leading, keyword or empty class name get a `Model` prefix, and collisions are deduplicated (classes with a bare digit suffix). Field names (not parameters) also prefix leading underscores because attrs and pydantic treat such fields specially, so a `_meta` column becomes `column__meta`. Digit- or underscore-leading enum values get a `VALUE_` prefix since pyright strict treats leading-underscore members as private. `:copyfrom` escapes quoted identifiers in the emitted column list. Reported by AlexanderHott in issue 160 ([AlexanderHott](https://github.com/AlexanderHott))
### Breaking
* [#138](https://github.com/rayakame/sqlc-gen-better-python/pull/138) Config option `sql_driver` is now required instead of defaulting to asyncpg ([rayakame](https://github.com/rayakame))
* [#148](https://github.com/rayakame/sqlc-gen-better-python/pull/148) Model and row class fields now escape reserved names like function parameters already did - a column named `id` generates the field `id_` ([rayakame](https://github.com/rayakame))
* [#148](https://github.com/rayakame/sqlc-gen-better-python/pull/148) Generated code now requires Python 3.12+. Type aliases use the lazy PEP 695 `type` statement (which also makes them safe at module level with `omit_typechecking_block`), `QueryResults` uses PEP 695 generic-class syntax, and enums subclass `enum.StrEnum` instead of `(str, enum.Enum)` ([rayakame](https://github.com/rayakame))

## v0.4.5 - 2025-09-03
### Added
* [#122](https://github.com/rayakame/sqlc-gen-better-python/pull/122) Add `omit_typechecking_blocks` option ([davfsa](https://github.com/davfsa))
* [#124](https://github.com/rayakame/sqlc-gen-better-python/pull/124) Added config option `omit_kwargs_limit`. ([rayakame](https://github.com/rayakame))
### Dependencys
* [#125](https://github.com/rayakame/sqlc-gen-better-python/pull/125) Updated sqlc to v1.30.0 used for testing the plugin ([rayakame](https://github.com/rayakame))

## v0.4.4 - 2025-05-30
### Changed
* [#112](https://github.com/rayakame/sqlc-gen-better-python/pull/112) Improved `:execrows` performance for `asyncpg` and added speedup option for `:execrows` ([rayakame](https://github.com/rayakame))
### Fixed
* [#112](https://github.com/rayakame/sqlc-gen-better-python/pull/112) Added `columns` kwarg to `:copyfrom` for `asyncpg` to fix inserts for columns with default values ([rayakame](https://github.com/rayakame))

## v0.4.3 - 2025-05-28
### Fixed
* [#109](https://github.com/rayakame/sqlc-gen-better-python/pull/109) Fixed missing model import when using `:copyfrom` cmd. ([rayakame](https://github.com/rayakame))

## v0.4.2 - 2025-05-25
### Added
* [#104](https://github.com/rayakame/sqlc-gen-better-python/pull/104) Enabled ruff `preview` config option.  ([rayakame](https://github.com/rayakame))
* [#105](https://github.com/rayakame/sqlc-gen-better-python/pull/105) Added support for type overrides, allowing users to specify their own python types for specific sql types. ([rayakame](https://github.com/rayakame))

## v0.4.1 - 2025-05-23
### Fixed
* [#97](https://github.com/rayakame/sqlc-gen-better-python/pull/97) Added `None` to `QueryResultsArgsType` ([rayakame](https://github.com/rayakame))

## v0.4.0 - 2025-05-21
### Added
* [#59](https://github.com/rayakame/sqlc-gen-better-python/pull/59) Added hyperlink to github profile of contributors in the changelog ([rayakame](https://github.com/rayakame))
* [#63](https://github.com/rayakame/sqlc-gen-better-python/pull/63) Added strict output tests & ci to prevent bugs from happening ([rayakame](https://github.com/rayakame))
* [#66](https://github.com/rayakame/sqlc-gen-better-python/pull/66) Added an example sqlc.yaml to README.md ([AlexanderHOtt](https://github.com/AlexanderHOtt))
* [#69](https://github.com/rayakame/sqlc-gen-better-python/pull/69) Added config option to auto generate docstrings for generated python code. ([rayakame](https://github.com/rayakame))
* [#74](https://github.com/rayakame/sqlc-gen-better-python/pull/74) Added support for query annotations `execrows` and `execresult` for driver `asyncpg` ([rayakame](https://github.com/rayakame))
* [#75](https://github.com/rayakame/sqlc-gen-better-python/pull/75) Code coverage tooling ([rayakame](https://github.com/rayakame))
* [#74](https://github.com/rayakame/sqlc-gen-better-python/pull/74) Added `ConnectionLike` instead of `asyncpg.Connection` which allows also using connection pools ([AlexanderHott](https://github.com/AlexanderHott))
* [#82](https://github.com/rayakame/sqlc-gen-better-python/pull/82) Added tests for `aiosqlite` driver with 100% coverage ([rayakame](https://github.com/rayakame))
* [#86](https://github.com/rayakame/sqlc-gen-better-python/pull/86) Added tests for `sqlite3` driver with 100% coverage ([rayakame](https://github.com/rayakame))
* [#86](https://github.com/rayakame/sqlc-gen-better-python/pull/86) Brought `sqlite3` back to full compatibility ([rayakame](https://github.com/rayakame))
* [#87](https://github.com/rayakame/sqlc-gen-better-python/pull/87) Added support for `:copyfrom` for driver `asyncpg` ([rayakame](https://github.com/rayakame))
### Changed
* [#63](https://github.com/rayakame/sqlc-gen-better-python/pull/63) Removed unnecessary `msgspec.field()` and `attrs.field()` for models. ([rayakame](https://github.com/rayakame))
* [#74](https://github.com/rayakame/sqlc-gen-better-python/pull/74) `:many` queries now return `QueryResults` allowing both iteration over rows and fetching rows. ([rayakame](https://github.com/rayakame))
### Deprecated
* [#63](https://github.com/rayakame/sqlc-gen-better-python/pull/63) Added `id` to reserved keywords so that it will appear as `id_` ([rayakame](https://github.com/rayakame))
### Fixed
* [#70](https://github.com/rayakame/sqlc-gen-better-python/pull/70) Fixed the `uv sync` command in CONTRIBUTING.md and added an example docker command to create a postgres instance for testing.  ([AlexanderHOtt](https://github.com/AlexanderHOtt))
* [#82](https://github.com/rayakame/sqlc-gen-better-python/pull/82) Brought `aiosqlite` back to full compatibility ([rayakame](https://github.com/rayakame))
* [#82](https://github.com/rayakame/sqlc-gen-better-python/pull/82) Fixed incorrect typing of query function arguments for nullable fields. ([null-domain](https://github.com/null-domain))

## v0.3.1 - 2025-05-07
### Fixed
* [#50](https://github.com/rayakame/sqlc-gen-better-python/pull/50) Fixed missing `__init__` return type annotation and connection parameter type when using `asyncpg` driver. (tandemdude)
* [#53](https://github.com/rayakame/sqlc-gen-better-python/pull/53) Wrong deserialization of `datetime.datetime` when using `asyncpg` (rayakame)
* [#53](https://github.com/rayakame/sqlc-gen-better-python/pull/53) Fixed unnecessary type conversion when returning data from queries using `asyncpg` (rayakame)

## v0.3.0 - 2025-05-05
### Added
* [#37](https://github.com/rayakame/sqlc-gen-better-python/pull/37) Added `debug` config option to enable debug output. (rayakame)
* [#38](https://github.com/rayakame/sqlc-gen-better-python/pull/38) Added documentation for every configuration option in the `README`. (rayakame)
* [#39](https://github.com/rayakame/sqlc-gen-better-python/pull/39) Added `emit_init_file` configuration option to control `__init__.py` creation. (rayakame)
* [#40](https://github.com/rayakame/sqlc-gen-better-python/pull/40) Added support for `msgspec` model type. (rayakame)
### Fixed
* [#41](https://github.com/rayakame/sqlc-gen-better-python/pull/41) Fixed missing empty lines when using `asyncpg` driver. (rayakame)

## v0.2.0 - 2025-05-05
### Added
* [#29](https://github.com/rayakame/sqlc-gen-better-python/pull/29) Added early driver support for `asyncpg`. Only has support for `exec`, `many` and `one` (rayakame)
### Fixed
* [#31](https://github.com/rayakame/sqlc-gen-better-python/pull/31) Missing return statements for `:execresult`, `:execrows` and `:execlastid` for `aiosqlite` and `sqlite3` (rayakame)

## v0.1.0 - 2025-04-01
### Added
* [#17](https://github.com/rayakame/sqlc-gen-better-python/pull/17) Added support for driver `sqlite3` (rayakame)
* [#21](https://github.com/rayakame/sqlc-gen-better-python/pull/21) Added support for `sqlc.embed()` (rayakame)
### Changed
* [#20](https://github.com/rayakame/sqlc-gen-better-python/pull/20) Query functions now don't take param-structs (rayakame)

## v0.0.1 - 2025-03-31
### Added
* [#13](https://github.com/rayakame/sqlc-gen-better-python/pull/13) Added `emit_classes` config option that, if enabled, puts all the queries into classes (rayakame)
* [#14](https://github.com/rayakame/sqlc-gen-better-python/pull/14) Added changelog functionality (rayakame)
