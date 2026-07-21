---
title: Docstrings
weight: 80
prev: /docs/guide/working-with-json
next: /docs/guide/sqlite-type-conversion
---

The plugin can document everything it generates - models, enums, and query
functions - in the docstring convention of your choice.

```yaml
options:
  docstrings: "google"        # none | google | numpy | pep257
  docstrings_emit_sql: true   # embed each query's SQL (default true)
```

The default is `none`, which emits no docstrings at all.

## The three conventions

The same `:one` query, generated under each convention:

{{< tabs >}}

  {{< tab name="google" >}}
````python
async def get_field_naming(conn: ConnectionLike, *, id_: int) -> models.TestFieldNaming | None:
    """Fetch one from the db using the SQL query with `name: GetFieldNaming :one`.

    ```sql
    SELECT id, outputs
    FROM test_field_namings
    WHERE id = $1 LIMIT 1
    ```

    Args:
        conn:
            Connection object of type `ConnectionLike` used to execute the query.
        id_: int.

    Returns:
        Result of type `models.TestFieldNaming` fetched from the db. Will be `None` if not found.
    """
````
  {{< /tab >}}

  {{< tab name="numpy" >}}
````python
async def get_field_naming(conn: ConnectionLike, *, id_: int) -> models.TestFieldNaming | None:
    """Fetch one from the db using the SQL query with `name: GetFieldNaming :one`.

    ```sql
    SELECT id, outputs
    FROM test_field_namings
    WHERE id = $1 LIMIT 1
    ```

    Parameters
    ----------
    conn : ConnectionLike
        Connection object of type `ConnectionLike` used to execute the query.
    id_ : int

    Returns
    -------
    models.TestFieldNaming
        Result fetched from the db. Will be `None` if not found.

    """
````
  {{< /tab >}}

  {{< tab name="pep257" >}}
````python
async def get_field_naming(conn: ConnectionLike, *, id_: int) -> models.TestFieldNaming | None:
    """Fetch one from the db using the SQL query with `name: GetFieldNaming :one`.

    ```sql
    SELECT id, outputs
    FROM test_field_namings
    WHERE id = $1 LIMIT 1
    ```

    Arguments:
    conn -- Connection object of type `ConnectionLike` used to execute the query.
    id_ -- int.

    Returns:
    models.TestFieldNaming -- Result fetched from the db. Will be `None` if not found.
    """
````
  {{< /tab >}}

{{< /tabs >}}

## Embedded SQL

By default each query function's docstring includes the exact SQL it runs, in a
fenced `sql` block - handy when reading generated code or hovering a call in your
editor. Set `docstrings_emit_sql: false` to leave it out.

{{< callout type="info" >}}
  `:copyfrom` functions never embed SQL, since they do not execute a statement -
  they stream rows via the driver's bulk copy API.
{{< /callout >}}

## Models and enums

Models and enums are documented too. Under `numpy`, the generated model carries an
`Attributes` section:

```python
@attrs.define()
class TestFieldNaming:
    """Model representing TestFieldNaming.

    Attributes
    ----------
    id_ : int
    outputs : str

    """

    id_: int
    outputs: str
```

## Linting generated code

If you run a docstring linter (ruff's `pydocstyle` rules, for example) over the
generated package, set its convention to match the one you generated, or every
file will report style violations. In `ruff.toml`:

```toml
[lint.pydocstyle]
convention = "google"
```
