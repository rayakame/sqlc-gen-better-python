-- name: InsertInvalidIdentifiers :exec
INSERT INTO test_invalid_identifiers (id, "3p%", "new notes") VALUES ($1, $2, $3);

-- name: GetInvalidIdentifiers :one
SELECT * FROM test_invalid_identifiers WHERE id = $1;

-- name: InsertThirdPartyStat :exec
INSERT INTO "3rd_party_stats" (id, total) VALUES ($1, $2);

-- name: GetThirdPartyStat :one
SELECT * FROM "3rd_party_stats" WHERE id = $1;
