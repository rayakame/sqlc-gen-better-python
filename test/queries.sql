
-- name: TestExecute :exec
INSERT INTO test_enum (id, b, b2, m)
VALUES ($1, $2, $3, $4);

