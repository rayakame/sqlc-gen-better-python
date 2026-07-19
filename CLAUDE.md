# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Ground rules

- NEVER add Claude as an author, co-author, or contributor to commits. No
  `Co-Authored-By` trailers, no "Generated with" lines. Commits are authored
  solely by the repository owner.
- NEVER commit (or push) unless explicitly told to. Finish the work, validate
  it, and report; the user decides when to commit.
- Use only plain ASCII characters in everything you write (code, comments,
  docs, SQL, commit messages). No em-dashes, curly quotes, arrows, or box
  drawing. Non-ASCII in .sql files has corrupted sqlc's byte-offset parameter
  rewriting before.

## What this is

A sqlc WASM plugin written in Go that generates Python database code (models +
query functions + enums) from SQL. The plugin is compiled to `wasip1/wasm` and
executed by `sqlc generate`. Supported Python drivers: `asyncpg`, `aiosqlite`,
`sqlite3`; model types: `dataclass`, `attrs`, `msgspec`, `pydantic`.

Generated code targets Python 3.12+ (PEP 695 type aliases and generics,
`enum.StrEnum`). Generated output must be deterministic and byte-identical
across runs: CI compares committed fixtures with `sqlc diff`, checks them with
`ruff format --check`, pyright strict, and ruff, and runs runtime pytest
suites against real databases.

## Commands

### Go (plugin source)

```
make tests      # go test -shuffle=on ./...
make fmt        # golangci-lint fmt
make lint       # golangci-lint run
make lint-fix   # auto-fix a subset of linters
make pipelines  # lint-fix + fmt + lint (default goal)
```

There is pre-existing golangci-lint debt (~40 issues); when linting your own
changes, check that no NEW issues appear in the files you touched.

### Building the WASM plugin

After any Go change, the plugin must be rebuilt before `sqlc generate` picks
it up:

```
scripts\build\build.bat   # Windows (scripts/build/build.sh on Unix)
```

This builds with `GOOS=wasip1 GOARCH=wasm`, computes the new SHA-256, patches
the `sha256:` field in the root `sqlc.yaml` and every `test/driver_*/sqlc.yaml`
(sqlc refuses to run the plugin on a hash mismatch), and copies the `.wasm`
into each test driver directory.

### Python (verification of generated code)

Python tooling is uv + nox. One-time setup: `uv sync --group dev`. Requires
`sqlc` on PATH and Python >= 3.12.

```
uv run nox                    # all default sessions
uv run nox -s asyncpg         # regenerate test/driver_asyncpg via sqlc, then pyright + ruff on it
uv run nox -s aiosqlite       # same for aiosqlite
uv run nox -s sqlite3         # same for sqlite3
uv run nox -s asyncpg_check   # `sqlc diff` variant: verifies committed generated code is up to date (CI uses these)
uv run nox -s pyright ruff    # type-check / lint the test suite itself
uv run nox -s pytest          # runtime tests (needs postgres, see below)
```

Extra pytest args pass through after `--`, e.g.
`uv run nox -s pytest -- test/driver_asyncpg/msgspec/test_msgspec_classes.py -k test_name`.

pytest needs a local PostgreSQL, configured via the `POSTGRES_URI` env var
(default `postgresql://root:187187@localhost:5432/root`). CONTRIBUTING.md has
a `docker run` one-liner for it.

The full verification loop after a generator change: `go build ./...` ->
rebuild wasm -> `uv run nox` -> `uv run nox -s asyncpg_check sqlite3_check
aiosqlite_check` -> commit the regenerated fixtures together with the Go
change (when told to commit).

### Changelog

Every PR needs a changie fragment: `make changelog` (or `changie new`).
Fragments live in `.changes/unreleased/`.

## Architecture

Entry point: `plugin/main.go` -> `codegen.Run(internal.Handler)`. The whole
generation pipeline lives in `internal/handler.go`:

1. **`internal/config`** - parses and validates plugin options from the
   `GenerateRequest` (driver, model type, docstrings, overrides, ...).
   Enum-like constants in `constants.go`, override matching in `overrides.go`.
2. **`internal/types`** - engine-specific SQL-type -> Python-type mapping
   (`postgresql.go`, `sqlite.go`), selected by `GetTypeConversionFunc(engine)`.
3. **`internal/transform`** - turns the sqlc catalog/queries into the IR:
   `BuildEnums()`, `BuildTables()`, `BuildQueries(tables)`,
   `FilterUnusedModels()`. `type.go` builds `PyType` and normalizes
   `SQLType` (lowercased once here; every downstream consumer relies on it).
4. **`internal/model`** - the IR structs (`Enum`, `Table`, `Query`, `PyType`,
   ...) plus naming logic: initialisms, table-name singularization
   (jinzhu/inflection; exclusions match bare AND schema-qualified names),
   Python reserved-word escaping (`reserved.go`), and `DedupName`.
5. **`internal/driver`** - the `Driver` interface (`driver.go`) with
   `asyncpg.go` for PostgreSQL and `sqlite_base.go` as the single
   implementation for BOTH sqlite drivers (parameterized by module name +
   async flag; there are no separate aiosqlite/sqlite3 files). A driver knows
   which query commands it supports and emits query function bodies and the
   `QueryResults` class. `conversion.go` holds the ordered sqlite
   adapter/converter spec table; adapters are registered for parameter types,
   converters for non-overridden return types, and the import resolver
   mirrors exactly what `WriteConversionSetup` emits. `rowbuilder.go` builds
   row-decoding expressions (overrides/enums convert inline; lists convert
   element-wise). `common.go` has the shared signature/param expansion;
   `convertParamExpr` converts overridden params back to their DefaultType
   (element-wise for lists) and is also used by the copyfrom emitter.
6. **`internal/render`** - file-level orchestration: produces `models.py`
   (always, even when empty), `enums.py` (when enums exist), and one queries
   module per query file; resolves imports (`imports.go`, including the
   `typing.TYPE_CHECKING` block and the omit_typechecking_block layout),
   returns `[]*plugin.File`.
7. **`internal/writer`** - `CodeWriter`, the low-level indented-Python emitter
   (lines, headers, docstrings in google/numpy/pep257 conventions,
   `QueryResults` class skeleton, line-length-aware call wrapping).

`internal/log` is a debug logger; with the `debug: true` plugin option its
buffer is emitted as an extra output file.

### Invariants that bite

- Generated output must be byte-identical to the committed fixtures AND a
  no-op under `ruff format` (line-length 320; anything longer is emitted
  pre-exploded with magic trailing commas via `writer.MaxLineLength`,
  `FitsLine`, `WriteWrappedCall`).
- Import resolution must match emission exactly: a name referenced at runtime
  by emitted code needs a runtime import; an annotation-only name belongs in
  the TYPE_CHECKING block (annotations are lazy via
  `from __future__ import annotations`). pyright strict + ruff (select ALL)
  on the fixtures are the enforcement.
- Type aliases in generated code must be lazy (PEP 695 `type` statements):
  with `omit_typechecking_block: true` they execute at module level, and
  `asyncpg.Connection[...]` is a stub-only generic that raises TypeError if
  evaluated eagerly.
- sqlite `register_converter` is process-global; per-module registration
  emits only what that module needs (params -> adapters, non-overridden
  returns -> converters).
- Column overrides do not attach to `ANY($1::type[])` parameters (sqlc does
  not link them to the column); only `db_type` overrides reach those.

### Test layout

`test/driver_<driver>/` each contain a `sqlc.yaml` generating a matrix of
codegen blocks (4 model types x `classes`/`functions`, plus special-purpose
blocks like `omit_tc/` for omit_typechecking_block coverage) from shared
schema/query files into committed subdirectories. Query .sql files map 1:1 to
generated Python modules, so edge cases get their own query file when they
need an isolated module (e.g. adapter-only vs converter-only sqlite
conversion setups). The generated Python is committed on purpose: CI
regenerates it (`*_check` sessions use `sqlc diff`), type-checks it with
pyright in strict mode, lints it with ruff, and runs runtime pytest suites
(`test_*.py` next to the generated packages). When a Go change alters
generated output, rebuild the wasm, rerun the driver sessions, and commit the
regenerated files (when told to commit).

Keep .sql files ASCII-only: multi-byte characters in comments corrupt sqlc's
byte-offset parameter rewriting and can silently drop `?` placeholders.

The root `sqlc.yaml` generating into `test/` directly is a scratch config for
quick manual `sqlc generate` runs during development.

### Lint config gotchas

- `ruff.toml` is the root config; per-directory `ruff.toml` files under
  `test/driver_*/<dir>/` extend it to set the matching docstring convention.
  Never define `[lint.per-file-ignores]` in an extending config: it REPLACES
  the root's table for that subtree instead of merging. Add patterns to the
  root config instead.
- Validate ruff changes with `--no-cache`; stale caches have masked real
  failures locally that CI then caught.
- pyright is configured in `pyproject.toml` (pythonVersion 3.12, strict). Do
  not use `executionEnvironments`; it breaks third-party import resolution.
