# sqlc-gen-better-python

[![Codecov](https://codecov.io/gh/rayakame/sqlc-gen-better-python/graph/badge.svg?token=LROCMXW6MC)](https://codecov.io/gh/rayakame/sqlc-gen-better-python)
![Python Version from PEP 621 TOML](https://img.shields.io/python/required-version-toml?tomlFilePath=https%3A%2F%2Fraw.githubusercontent.com%2Frayakame%2Fsqlc-gen-better-python%2Fmain%2Fpyproject.toml)
![Ruff](https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/charliermarsh/ruff/main/assets/badge/v2.json)
[![CI](https://github.com/rayakame/sqlc-gen-better-python/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/rayakame/sqlc-gen-better-python/actions/workflows/ci.yml)

A WASM plugin for SQLC allowing the generation of Python code.


> [!NOTE]  
> This is currently being worked on. It is far from being ready for any kind of release, let alone a stable one.  
> Please wait for the v1 release; before that, this plugin is likely to not work.

## Example Config

```yaml
# filename: sqlc.yaml
version: "2"
plugins:
  - name: python
    wasm:
      url: https://github.com/rayakame/sqlc-gen-better-python/releases/download/v0.4.5/sqlc-gen-better-python.wasm
      sha256: 9b42f7cffda942388470e777af957ec8ab03cbea9124ff8c49bef05f9d00cb66
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

More options at the [`sqlc` config reference](https://docs.sqlc.dev/en/stable/reference/config.html)

## Configuration Options

| Name                             | Type           | Required | Description                                                                                                                                                                                                               |
|----------------------------------|----------------|----------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `package`                        | string         | yes      | The name of the package where the generated files will be located                                                                                                                                                         |
| `emit_init_file`                 | bool           | yes      | If set to to `false` there will be no `__init__.py` file created in the package that you specified. Only set this to false if you know that you already have a `__init__.py` file otherwise the generated code wont work. |
| `sql_driver`                     | string         | no       | The name of the sql driver you want to use. Defaults to `aiosqlite`. Valid options are listed [here](#feature-support)                                                                                                    |
| `model_type`                     | string         | no       | The model type you want to use. This can be one of `dataclass`, `msgspec` or `attrs`. Defaults to `dataclass`                                                                                                             |
| `initialisms`                    | list[string]   | no       | An array of [initialisms](https://google.github.io/styleguide/go/decisions.html#initialisms) to upper-case. For example, `app_id` becomes `AppID`. Defaults to `["id"]`.                                                  |
| `emit_exact_table_names`         | bool           | no       | If `true`, model names will mirror table names. Otherwise sqlc attempts to singularize plural table names.                                                                                                                |
| `emit_classes`                   | bool           | no       | If `true`, every query function will be put into a class called `Querier`. Otherwise every function will be a standalone function.                                                                                        |
| `inflection_exclude_table_names` | list[string]   | no       | An array of table names that should not be turned singular. Only applies if `emit_exact_table_names` is `false`.                                                                                                          |
| `omit_unused_models`             | bool           | no       | If set to `true` and there are models/tables that are not used in any query, they wont be turned into models.                                                                                                             |
| `omit_typechecking_block`        | bool           | no       | If set to `true`, will not wrap all non-builtin types behind a `typing.TYPE_CHECKING` block. Defaults to `false`                                                                                                          |
| `docstrings`                     | string         | no       | If set, there will be docstrings generated in the selected format. This can be one of `google`, `numpy`, `pep257` and `none`. `none` will not generate any docstrings.                                                    |
| `docstrings_emit_sql`            | bool           | no       | If set to `false` the SQL code for each query wont be included in the docstrings. This defaults to `true` but is not used when `docstrings` is not set or set to `none`                                                   |
| `query_parameter_limit`          | integer        | no       | Not yet implemented.                                                                                                                                                                                                      |
| `omit_kwargs_limit`              | integer        | no       | This can be used to set a limit where any query with less or equal amounts of parameters will not require kwargs for the parameters. This defaults to `0` which makes every query require kwargs for their parameters.    |
| `speedups`                       | bool           | no       | If set to `true` the plugin will use other librarys for type conversion. Needs extra dependecys to be installed. This option currently only affects `sqlite3` & `aiosqlite` and uses the library `ciso8601`               |
| `overrides`                      | list[Override] | no       | A list of [type overrides](#type-overrides).                                                                                                                                                                              |
| `debug`                          | bool           | no       | If set to `true`, there will be debug logs generated into a `log.txt` file when executing `sqlc generate`. Defaults to `false`                                                                                            |

### Type Overrides

Similar to `sqlc-gen-go` this plugin supports overriding types with your own. You can either override the type of every
column that has a specific sql type, or you can overwrite the type of specific columns.

```yaml
# filename: sqlc.yaml
# ...
options:
  # ...
  overrides:
    - db_type: text
      py_type:
        import: collections
        package: UserString
        type: UserString
    - column: table_name.text_column
      py_type:
        import: collections
        type: collections.UserString

```

## Feature Support

Every [sqlc macro](https://docs.sqlc.dev/en/latest/reference/macros.html) is supported.
The supported [query commands](https://docs.sqlc.dev/en/latest/reference/query-annotations.html) depend on the SQL
driver you are using, supported commands are listed below.
> Every `:batch*` command is not supported by this plugin and probably will never be.

> Prepared Queries are not planned for the near future, but will be implemented sooner or later

|           | `:exec` | `:execresult` | `:execrows` | `:execlastid` | `:many` | `:one` | `:copyfrom` |
|-----------|---------|---------------|-------------|---------------|---------|--------|-------------|
| aiosqlite | yes     | yes           | yes         | yes           | yes     | yes    | no          |
| sqlite3   | yes     | yes           | yes         | yes           | yes     | yes    | no          |
| asyncpg   | yes     | yes           | yes         | no            | yes     | yes    | yes         |
| psycopg2  | no      | no            | no          | no            | no      | no     | no          |
| mysql     | no      | no            | no          | no            | no      | no     | no          |

## Development

A roadmap of what is planned & worked on can be found [here](https://github.com/users/rayakame/projects/1/).

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
