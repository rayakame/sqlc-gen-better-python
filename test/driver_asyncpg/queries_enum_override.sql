-- name: InsertEnumOverride :exec
INSERT INTO test_enum_override (id, mood_test) VALUES ($1, $2);

-- name: GetEnumOverrideMood :one
SELECT mood_test FROM test_enum_override WHERE id = $1;
