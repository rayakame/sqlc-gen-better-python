-- name: GetOverrideHappenedAt :one
SELECT happened_at FROM test_override_conversion WHERE id = ?;
