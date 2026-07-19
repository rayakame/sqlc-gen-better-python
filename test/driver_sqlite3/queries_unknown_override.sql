-- name: InsertUnknownOverride :exec
INSERT INTO test_unknown_override (id, happened_at) VALUES (?, ?);

-- name: GetUnknownOverride :one
SELECT happened_at FROM test_unknown_override WHERE id = ?;
