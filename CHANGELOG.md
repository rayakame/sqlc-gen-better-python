# Changelog
All notable changes to this project will be documented in this file.


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
