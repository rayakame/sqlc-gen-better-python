-- name: InsertSliceRow :exec
INSERT INTO test_slice (id, name, note) VALUES (?, ?, ?);

-- name: GetSliceRows :many
SELECT * FROM test_slice WHERE id IN (sqlc.slice('ids')) ORDER BY id;

-- name: GetSliceRowFiltered :one
SELECT * FROM test_slice WHERE name = ? AND id IN (sqlc.slice('ids')) AND id != ? LIMIT 1;

-- name: GetSliceRowsByNotes :many
SELECT * FROM test_slice WHERE note IN (sqlc.slice('notes')) ORDER BY id;

-- name: GetFirstSliceName :one
SELECT name FROM test_slice WHERE id IN (sqlc.slice('ids')) OR name IN (sqlc.slice('names')) ORDER BY id LIMIT 1;

-- name: GetSliceRowsByNameOrNote :many
SELECT * FROM test_slice WHERE name IN (sqlc.slice('names')) OR note IN (sqlc.slice('names')) ORDER BY id;

-- name: DeleteSliceRows :execrows
DELETE FROM test_slice WHERE id IN (sqlc.slice('ids'));
