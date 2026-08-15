-- name: InsertConverted :exec
INSERT INTO test_converters (id, prefs, maybe_prefs, tags) VALUES (?, ?, ?, ?);

-- name: GetConverted :one
SELECT * FROM test_converters WHERE id = ?;

-- name: ListConvertedByTags :many
SELECT id FROM test_converters WHERE tags = ?;

-- name: DeleteConverted :exec
DELETE FROM test_converters WHERE id = ?;
