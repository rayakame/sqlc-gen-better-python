-- name: InsertInvalidIdentifiers :exec
INSERT INTO test_invalid_identifiers (id, `3p%`, `new notes`) VALUES (?, ?, ?);

-- sqlc's star expansion drops the backtick quoting on MySQL, so the
-- columns are listed explicitly.
-- name: GetInvalidIdentifiers :one
SELECT id, `3p%`, `new notes`, `%pct` FROM test_invalid_identifiers WHERE id = ?;

-- name: InsertThirdPartyStat :exec
INSERT INTO `3rd_party_stats` (id, total) VALUES (?, ?);

-- name: GetThirdPartyStat :one
SELECT * FROM `3rd_party_stats` WHERE id = ?;
