# sqlc-gen-better-python

[![Codecov](https://codecov.io/gh/rayakame/sqlc-gen-better-python/graph/badge.svg?token=LROCMXW6MC)](https://codecov.io/gh/rayakame/sqlc-gen-better-python)
[![Go coverage](https://img.shields.io/codecov/c/github/rayakame/sqlc-gen-better-python?flag=go&label=go%20coverage)](https://app.codecov.io/gh/rayakame/sqlc-gen-better-python/flags)
[![Python coverage](https://img.shields.io/codecov/c/github/rayakame/sqlc-gen-better-python?flag=python&label=python%20coverage)](https://app.codecov.io/gh/rayakame/sqlc-gen-better-python/flags)
![Python Version from PEP 621 TOML](https://img.shields.io/python/required-version-toml?tomlFilePath=https%3A%2F%2Fraw.githubusercontent.com%2Frayakame%2Fsqlc-gen-better-python%2Fmain%2Fpyproject.toml)
![Ruff](https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/charliermarsh/ruff/main/assets/badge/v2.json)
[![CI](https://github.com/rayakame/sqlc-gen-better-python/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/rayakame/sqlc-gen-better-python/actions/workflows/ci.yml)

`sqlc-gen-better-python` is a [sqlc](https://sqlc.dev) plugin that turns your
SQL schema and queries into modern, fully typed Python database code: models,
typed query functions, and enums. You keep writing SQL; the Python stays in
sync with it.

You write:

```sql
-- name: GetUser :one
SELECT * FROM users WHERE id = $1;
```

and get back:

```python
async def get_user(conn: ConnectionLike, *, id_: int) -> models.User | None:
    row = await conn.fetchrow(GET_USER, id_)
    if row is None:
        return None
    return models.User(id_=row[0], name=row[1])
```

No ORM, no hand-written row unpacking. Generated code targets **Python 3.12 or
newer** and passes pyright (strict) and ruff.

## Documentation

**https://sqlc-gen-better-python.rayakame.dev/**

- [Getting Started](https://sqlc-gen-better-python.rayakame.dev/docs/getting-started/) - install the plugin and generate your first models.
- [Guide](https://sqlc-gen-better-python.rayakame.dev/docs/guide/) - every feature, each with real generated output.
- [Reference](https://sqlc-gen-better-python.rayakame.dev/docs/reference/) - all options, type mappings, and per-driver feature support.

Questions or feedback? Join the [Discord](https://discord.gg/hikari).

## Features

- **Four model types** - `dataclass`, `attrs`, `msgspec`, or `pydantic`
  ([docs](https://sqlc-gen-better-python.rayakame.dev/docs/guide/model-types/)).
- **Nine drivers** - `asyncpg`, `psycopg_async`, and `psycopg_sync` for
  PostgreSQL, `aiosqlite` and `sqlite3` for SQLite, `asyncmy` and `pymysql`
  for MySQL, plus experimental `turso_async` and `turso_sync` for
  [Turso](https://github.com/tursodatabase/turso)
  ([docs](https://sqlc-gen-better-python.rayakame.dev/docs/guide/drivers/)).
- **Typed query functions** - one module per query file, one function per query
  ([docs](https://sqlc-gen-better-python.rayakame.dev/docs/guide/writing-queries/)).
- **PostgreSQL enum types and MySQL enum columns** as `enum.StrEnum` classes
  ([docs](https://sqlc-gen-better-python.rayakame.dev/docs/guide/enums/)).
- **Type overrides and converters** - swap a column's Python type, or plug in your
  own encode/decode functions
  ([overrides](https://sqlc-gen-better-python.rayakame.dev/docs/guide/type-overrides/),
  [converters](https://sqlc-gen-better-python.rayakame.dev/docs/guide/converters/)).
- **Typed JSON columns** via msgspec structs
  ([docs](https://sqlc-gen-better-python.rayakame.dev/docs/guide/working-with-json/)).
- **Optional docstrings** in `google`, `numpy`, or `pep257` convention
  ([docs](https://sqlc-gen-better-python.rayakame.dev/docs/guide/docstrings/)).

Every [sqlc macro](https://docs.sqlc.dev/en/latest/reference/macros.html) is
supported. Which query commands are available depends on the driver - see the
[feature support matrix](https://sqlc-gen-better-python.rayakame.dev/docs/reference/feature-support/).

## Example config

```yaml
# filename: sqlc.yaml
version: "2"
plugins:
  - name: python
    wasm:
      url: https://github.com/rayakame/sqlc-gen-better-python/releases/download/v0.9.0/sqlc-gen-better-python.wasm
      sha256: d1787aa32e61f2e73c81a4f93b3e5a9beeec918952cff3183fb96313057f58a8
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

> [!TIP]
> No `sqlc` yet? Besides the [official installation methods](https://docs.sqlc.dev/en/latest/overview/install.html),
> `uv add --dev sqlc-bin` (or `pip install sqlc-bin`) installs
> [`sqlc-bin`](https://pypi.org/project/sqlc-bin/), the unmodified official
> binaries as a pinnable Python package - no Go toolchain required.

More options at the [`sqlc` config reference](https://docs.sqlc.dev/en/stable/reference/config.html),
and the full plugin option list in the
[configuration reference](https://sqlc-gen-better-python.rayakame.dev/docs/reference/configuration-options/).

## Used by

[<img src="docs/static/images/used-by/nmarkov.png" alt="nMarkov logo" height="72">](https://nmarkov.xyz/)

**[nMarkov](https://nmarkov.xyz/)** - a Discord chatbot that learns from your
server's messages and generates its own.

Using `sqlc-gen-better-python` in your project? [Open an issue](https://github.com/rayakame/sqlc-gen-better-python/issues)
to get listed here.

## Development

Contributions are very welcome, for more information and help please read
the [contribution guidelines](https://github.com/rayakame/sqlc-gen-better-python/blob/main/CONTRIBUTING.md).

### Changelog

Can be found [here](https://github.com/rayakame/sqlc-gen-better-python/blob/main/CHANGELOG.md)

## Credits

Special thanks to [tandemdude](https://github.com/tandemdude) for answering my questions on discord.
