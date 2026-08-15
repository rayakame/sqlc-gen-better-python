-- name: InsertDbtypeOverride :exec
INSERT INTO test_dbtype_override (id, happened_at) VALUES (?, ?);

-- name: GetDbtypeOverride :one
SELECT * FROM test_dbtype_override WHERE id = ?;
