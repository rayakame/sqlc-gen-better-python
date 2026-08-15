-- sqlc numbers SQLite placeholders as soon as a query uses a named argument,
-- and those indexes stop matching once a sqlc.slice marker expands to more
-- (or fewer) than one placeholder. Every query here mixes the two.

-- name: GetNamedSliceRows :many
SELECT * FROM test_slice WHERE id IN (sqlc.slice('ids')) AND name = sqlc.arg(wanted) ORDER BY id;

-- name: GetNamedSliceRowsReused :many
SELECT * FROM test_slice WHERE id IN (sqlc.slice('ids')) AND name = sqlc.arg(wanted) AND (note IS NULL OR note != sqlc.arg(wanted)) AND id IN (sqlc.slice('ids')) ORDER BY id;

-- name: GetNamedSliceRowsArgFirst :many
SELECT * FROM test_slice WHERE name = sqlc.arg(wanted) AND id IN (sqlc.slice('ids')) ORDER BY id;

-- name: CountNamedSliceRows :one
SELECT count(*) FROM test_slice WHERE id IN (sqlc.slice('ids')) AND name = sqlc.arg(wanted);
