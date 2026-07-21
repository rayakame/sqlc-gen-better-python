-- name: InsertSliceRow :exec
INSERT INTO test_slice (id, name) VALUES (?, ?);

-- name: GetSliceRows :many
SELECT * FROM test_slice WHERE id IN (sqlc.slice('ids')) ORDER BY id;

-- name: GetSliceRowFiltered :one
SELECT * FROM test_slice WHERE name = ? AND id IN (sqlc.slice('ids')) AND id != ? LIMIT 1;

-- name: CountSliceRows :one
SELECT count(*) FROM test_slice WHERE id IN (sqlc.slice('ids')) OR name IN (sqlc.slice('names'));

-- name: DeleteSliceRows :execrows
DELETE FROM test_slice WHERE id IN (sqlc.slice('ids'));
