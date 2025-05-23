# Changelog
All notable changes to this project will be documented in this file.


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
