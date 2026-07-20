-- name: InsertAnyParam :exec
INSERT INTO test_any_param (id, tag) VALUES (?, ?);

-- name: ListAnyParamIds :many
SELECT id FROM test_any_param WHERE tag = ?;
