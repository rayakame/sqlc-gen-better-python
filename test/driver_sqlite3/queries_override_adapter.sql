-- name: InsertOverrideConversion :exec
INSERT INTO test_override_conversion (id, price, happened_at) VALUES (?, ?, ?);

-- name: GetOverridePrice :one
SELECT price FROM test_override_conversion WHERE id = ?;
