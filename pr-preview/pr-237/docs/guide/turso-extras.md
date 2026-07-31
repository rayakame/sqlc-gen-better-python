# Turso extras

Turso adds features SQLite does not have. sqlc's sqlite parser rejects their
DDL (`CREATE MATERIALIZED VIEW`, `CREATE DOMAIN`, `BEGIN CONCURRENT`), but
they still work with generated code. The pattern:

{{< callout type="info" >}}
  **Shim the shape in the sqlc schema, create the real object at runtime.**
  sqlc only needs column names and types to generate; the driver does not care
  what the object actually is.
{{< /callout >}}

Everything on this page was verified against pyturso 0.7.1. Experimental
features are enabled per connection:
`turso.connect("app.db", experimental_features="views,custom_types")`.

## Materialized views

Give sqlc a plain `CREATE VIEW` with the same columns; it generates a model
and query functions from it. At runtime, create the real materialized view:

```sql
-- sqlc schema shim
CREATE VIEW user_stats AS
SELECT status, count(*) AS n
FROM users
GROUP BY status;

-- name: GetUserStats :many
SELECT status, n FROM user_stats ORDER BY status;
```

```python
conn = turso.connect("app.db", experimental_features="views")
conn.execute("CREATE MATERIALIZED VIEW user_stats AS SELECT status, count(*) AS n FROM users GROUP BY status")

stats = queries.get_user_stats(conn)()  # incrementally maintained by turso
```

## Change data capture

Turso writes changes to a normal `turso_cdc` table. Declare it in the sqlc
schema and query the change feed with generated functions; only the enabling
`PRAGMA` has to stay in runtime code (sqlc drops `PRAGMA` queries):

```sql
CREATE TABLE turso_cdc
(
    change_id     integer PRIMARY KEY NOT NULL,
    change_time   integer NOT NULL,
    change_txn_id integer,
    change_type   integer NOT NULL, -- 1 insert, 0 update, -1 delete
    table_name    text    NOT NULL,
    id            integer NOT NULL,
    before        blob,
    after         blob,
    updates       blob
);

-- name: ListChanges :many
SELECT change_id, change_type, table_name, id FROM turso_cdc ORDER BY change_id;
```

```python
conn.execute("PRAGMA capture_data_changes_conn('full')")
queries.create_user(conn, id_=1, name="ada")
changes = queries.list_changes(conn)()  # typed rows, one per change
```

## Concurrent writes (MVCC)

Generated functions never manage transactions, so they run unmodified inside
`BEGIN CONCURRENT` / `COMMIT`:

```python
conn = turso.connect("app.db", experimental_features="mvcc")
conn.execute("PRAGMA journal_mode = 'mvcc'").fetchall()

conn.execute("BEGIN CONCURRENT")
queries.create_user(conn, id_=1, name="ada")
conn.execute("COMMIT")
```

{{< callout type="warning" >}}
  Two gotchas: pyturso needs `experimental_features="mvcc"` at connect time -
  the pragma alone is not enough - and the pragma only takes effect once its
  result row is **fetched** (the `.fetchall()` above). Without either you get
  "Concurrent transaction mode is only supported when MVCC is enabled".
{{< /callout >}}

## Enum-like domains

Turso has no enum type, but a `CREATE DOMAIN` with a `CHECK` on a `STRICT`
table enforces the value set database-side. Shim the column as `text` for
sqlc and map it to your own `enum.StrEnum` with a
[type override](/docs/guide/type-overrides) - the same experience the plugin's
[PostgreSQL enums](/docs/guide/enums) provide:

```python
# app_enums.py
class Status(enum.StrEnum):
    ACTIVE = "active"
    BANNED = "banned"
```

```yaml
overrides:
  - column: users.status
    py_type:
      import: app_enums
      package: Status
      type: Status
```

```python
conn = turso.connect("app.db", experimental_features="custom_types")
conn.execute("CREATE DOMAIN status_d AS text CHECK (value IN ('active', 'banned'))")
conn.execute("CREATE TABLE users (id integer PRIMARY KEY NOT NULL, name text NOT NULL, status status_d NOT NULL) STRICT")

queries.create_user(conn, id_=1, name="ada", status=Status.ACTIVE)
user = queries.get_user(conn, id_=1)
assert user.status is Status.ACTIVE  # a real enum member
conn.execute("INSERT INTO users VALUES (2, 'x', 'nonsense')")  # IntegrityError
```

{{< callout type="info" >}}
  All of these Turso features are experimental and pre-1.0 - the SQL surface
  and the pyturso flags may change. The generated code is agnostic to them;
  only your runtime setup would need updating.
{{< /callout >}}

