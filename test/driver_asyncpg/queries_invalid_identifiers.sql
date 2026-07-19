-- name: InsertInvalidIdentifiers :exec
INSERT INTO test_invalid_identifiers (id, "3p%", "new notes") VALUES ($1, $2, $3);

-- name: GetInvalidIdentifiers :one
SELECT * FROM test_invalid_identifiers WHERE id = $1;
