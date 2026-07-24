-- name: InsertEnumOverride :exec
INSERT INTO test_enum_override (id, mood_test) VALUES ($1, $2);

-- name: GetEnumOverrideMood :one
SELECT mood_test FROM test_enum_override WHERE id = $1;

-- name: ListEnumOverrideByIds :many
SELECT id, mood_test FROM test_enum_override WHERE id = ANY($1::int[]);

-- name: CountEnumOverrideByMoods :one
SELECT count(*) FROM test_enum_override WHERE mood_test = ANY($1::test_mood[]);
