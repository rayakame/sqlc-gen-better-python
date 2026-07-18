# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

A sqlc WASM plugin written in Go that generates Python database code (models + query functions + enums) from SQL. The plugin is compiled to `wasip1/wasm` and executed by `sqlc generate`. Supported Python drivers: `asyncpg`, `aiosqlite`, `sqlite3`; model types: `dataclass`, `attrs`, `msgspec`, `pydantic`.

## Commands

### Go (plugin source)

```
make tests      # go test -shuffle=on ./...
make fmt        # golangci-lint fmt
make lint       # golangci-lint run
make lint-fix   # auto-fix a subset of linters
make pipelines  # lint-fix + fmt + lint (default goal)
```

### Building the WASM plugin

After any Go change, the plugin must be rebuilt before `sqlc generate` picks it up:

```
scripts\build\build.bat   # Windows (scripts/build/build.sh on Unix)
```

This builds with `GOOS=wasip1 GOARCH=wasm`, computes the new SHA-256, patches the `sha256:` field in the root `sqlc.yaml` and every `test/driver_*/sqlc.yaml` (sqlc refuses to run the plugin on a hash mismatch), and copies the `.wasm` into each test driver directory.

### Python (verification of generated code)

Python tooling is uv + nox. One-time setup: `uv sync --group dev`. Requires `sqlc` on PATH.

```
uv run nox                    # all default sessions
uv run nox -s asyncpg         # regenerate test/driver_asyncpg via sqlc, then pyright + ruff on it
uv run nox -s aiosqlite       # same for aiosqlite
uv run nox -s sqlite3         # same for sqlite3
uv run nox -s asyncpg_check   # `sqlc diff` variant: verifies committed generated code is up to date (CI uses these)
uv run nox -s pyright ruff    # type-check / lint the test suite itself
uv run nox -s pytest          # runtime tests (needs postgres, see below)
```

Extra pytest args pass through after `--`, e.g. `uv run nox -s pytest -- test/driver_asyncpg/msgspec/test_msgspec_classes.py -k test_name`.

pytest needs a local PostgreSQL, configured via the `POSTGRES_URI` env var (default `postgresql://root:187187@localhost:5432/root`). CONTRIBUTING.md has a `docker run` one-liner for it.

### Changelog

Every PR needs a changie fragment: `make changelog` (or `changie new`).

## Architecture

Entry point: `plugin/main.go` -> `codegen.Run(internal.Handler)`. The whole generation pipeline lives in `internal/handler.go`:

1. **`internal/config`** - parses and validates plugin options from the `GenerateRequest` (driver, model type, docstrings, overrides, ...). Enum-like constants in `constants.go`.
2. **`internal/types`** - engine-specific SQL-type -> Python-type mapping (`postgresql.go`, `sqlite.go`), selected by `GetTypeConversionFunc(engine)`.
3. **`internal/transform`** - turns the sqlc catalog/queries into the IR: `BuildEnums()`, `BuildTables()`, `BuildQueries(tables)`.
4. **`internal/model`** - the IR structs (`Enum`, `Table`, `Query`, ...) plus naming logic: initialisms, table-name singularization (jinzhu/inflection), and Python reserved-word escaping (`reserved.go`).
5. **`internal/driver`** - the `Driver` interface (`driver.go`) with one implementation per Python driver (`asyncpg.go`, `aiosqlite.go`/`sqlite3.go` sharing `sqlite_base.go`). A driver knows whether it is async, its connection type, which query commands (`:one`, `:many`, `:copyfrom`, ...) it supports, and emits the query function bodies and the `QueryResults` helper class.
6. **`internal/render`** - file-level orchestration: produces `models.py`, `enums.py`, and one queries module per query file, resolves imports (`imports.go`, including the `typing.TYPE_CHECKING` block), returns `[]*plugin.File`.
7. **`internal/writer`** - `CodeWriter`, the low-level indented-Python emitter (lines, headers, docstrings in google/numpy/pep257 conventions).

`internal/log` is a debug logger; with the `debug: true` plugin option its buffer is emitted as an extra `log.txt` output file.

### Test layout

`test/driver_<driver>/` each contain a `sqlc.yaml` that generates the full matrix - 3 model types x (`classes`/`functions`, i.e. `emit_classes` on/off) - into committed subdirectories like `msgspec/classes/`. The generated Python is committed on purpose: CI regenerates it (`*_check` sessions use `sqlc diff`), type-checks it with pyright in strict mode, lints it with ruff, and runs runtime pytest suites (`test_*.py` next to the generated packages). When a Go change alters generated output, rebuild the wasm, rerun the driver sessions, and commit the regenerated files.

The root `sqlc.yaml` generating into `test/` directly is a scratch config for quick manual `sqlc generate` runs during development.
