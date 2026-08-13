-- name: InsertEnumOverride :exec
INSERT INTO test_enum_override (id, mood_test) VALUES (?, ?);

-- name: GetEnumOverrideMood :one
SELECT mood_test FROM test_enum_override WHERE id = ?;

-- name: ListEnumOverrideByIds :many
SELECT id, mood_test FROM test_enum_override WHERE id IN (sqlc.slice('ids')) ORDER BY id;
