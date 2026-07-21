# sqlc-gen-better-python

[![Codecov](https://codecov.io/gh/rayakame/sqlc-gen-better-python/graph/badge.svg?token=LROCMXW6MC)](https://codecov.io/gh/rayakame/sqlc-gen-better-python)
[![Go coverage](https://img.shields.io/codecov/c/github/rayakame/sqlc-gen-better-python?flag=go&label=go%20coverage)](https://app.codecov.io/gh/rayakame/sqlc-gen-better-python/flags)
[![Python coverage](https://img.shields.io/codecov/c/github/rayakame/sqlc-gen-better-python?flag=python&label=python%20coverage)](https://app.codecov.io/gh/rayakame/sqlc-gen-better-python/flags)
![Python Version from PEP 621 TOML](https://img.shields.io/python/required-version-toml?tomlFilePath=https%3A%2F%2Fraw.githubusercontent.com%2Frayakame%2Fsqlc-gen-better-python%2Fmain%2Fpyproject.toml)
![Ruff](https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/charliermarsh/ruff/main/assets/badge/v2.json)
[![CI](https://github.com/rayakame/sqlc-gen-better-python/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/rayakame/sqlc-gen-better-python/actions/workflows/ci.yml)

A WASM plugin for SQLC allowing the generation of Python code.

The generated code requires **Python 3.12 or newer** (it uses PEP 695 type
aliases and generics, and `enum.StrEnum`).

## Documentation

**https://rayakame.github.io/sqlc-gen-better-python/**

- [Getting Started](https://rayakame.github.io/sqlc-gen-better-python/docs/getting-started/) - install the plugin and generate your first models.
- [Guide](https://rayakame.github.io/sqlc-gen-better-python/docs/guide/) - configuration, drivers, model types, writing queries, and every feature, each with real generated output.
- [Reference](https://rayakame.github.io/sqlc-gen-better-python/docs/reference/) - all configuration options, SQL-to-Python type mappings, and per-driver feature support.

Questions or feedback? Join the [Discord](https://discord.gg/hikari).

> [!NOTE]  
> Every Release before `v1.0.0`, including this one is an beta release. 
> These versions are primarly released for interested people who want to test this plugin and help make it better.
>
> Everything that is implemented works and is being used in production environments already.
> Since `v0.5.0` this includes full support for PostgreSQL enums and a fourth model type, `pydantic`.
> Feel free to lmk any wanted features and I'm going to do my best on implementing them with the time I have rn.

## Example Config

```yaml
# filename: sqlc.yaml
version: "2"
plugins:
  - name: python
    wasm:
      url: https://github.com/rayakame/sqlc-gen-better-python/releases/download/v0.5.1/sqlc-gen-better-python.wasm
      sha256: c7cc470df7625ae3232c2b042060b948180ae784ce3d81c32e8a2c040fe04fa7
sql:
  - engine: "postgresql"
    queries: "query.sql"
    schema: "schema.sql"
    codegen:
      - out: "app/db"
        plugin: python
        options:
          package: "db"
          emit_init_file: true
          sql_driver: "asyncpg"
          model_type: "msgspec"

```

More options at the [`sqlc` config reference](https://docs.sqlc.dev/en/stable/reference/config.html),
and the full plugin option list in the
[configuration reference](https://rayakame.github.io/sqlc-gen-better-python/docs/reference/configuration-options/).

## Features

- **Four model types** - `dataclass`, `attrs`, `msgspec`, or `pydantic`
  ([docs](https://rayakame.github.io/sqlc-gen-better-python/docs/guide/model-types/)).
- **Three drivers** - `asyncpg` for PostgreSQL, `aiosqlite` and `sqlite3` for SQLite
  ([docs](https://rayakame.github.io/sqlc-gen-better-python/docs/guide/drivers/)).
- **Typed query functions** - one module per query file, one function per query
  ([docs](https://rayakame.github.io/sqlc-gen-better-python/docs/guide/writing-queries/)).
- **PostgreSQL enums** as `enum.StrEnum` classes
  ([docs](https://rayakame.github.io/sqlc-gen-better-python/docs/guide/enums/)).
- **Type overrides and converters** - swap a column's Python type, or plug in your
  own encode/decode functions
  ([overrides](https://rayakame.github.io/sqlc-gen-better-python/docs/guide/type-overrides/),
  [converters](https://rayakame.github.io/sqlc-gen-better-python/docs/guide/converters/)).
- **Typed JSON columns** via msgspec structs
  ([docs](https://rayakame.github.io/sqlc-gen-better-python/docs/guide/working-with-json/)).
- **Optional docstrings** in `google`, `numpy`, or `pep257` convention
  ([docs](https://rayakame.github.io/sqlc-gen-better-python/docs/guide/docstrings/)).
- Generated code passes **pyright strict** and **ruff**.

The [sqlc macros](https://docs.sqlc.dev/en/latest/reference/macros.html)
`sqlc.arg`, `sqlc.narg` and `sqlc.embed` are supported (`sqlc.slice` is not).
Which query commands are available depends on the driver - see the
[feature support matrix](https://rayakame.github.io/sqlc-gen-better-python/docs/reference/feature-support/).

## Development

Contributions are very welcome, for more information and help please read
the [contribution guidelines](https://github.com/rayakame/sqlc-gen-better-python/blob/main/CONTRIBUTING.md).

### Changelog

Can be found [here](https://github.com/rayakame/sqlc-gen-better-python/blob/main/CHANGELOG.md)

## Credits

Because of missing documentation about creating these plugins, this work is heavily
inspired by:

- [sqlc-gen-go](https://github.com/sqlc-dev/sqlc-gen-go)
- [sqlc-gen-java](https://github.com/tandemdude/sqlc-gen-java)

Special thanks to [tandemdude](https://github.com/tandemdude) for answering my questions on discord.
