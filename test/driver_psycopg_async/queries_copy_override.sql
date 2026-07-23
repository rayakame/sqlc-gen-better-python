-- name: CopyOverrideRows :copyfrom
INSERT INTO test_copy_override (id, amount, "co""l") VALUES ($1, $2, $3);

-- name: CountCopyOverrideRows :one
SELECT count(*) FROM test_copy_override;

-- name: DeleteCopyOverrideRows :exec
DELETE FROM test_copy_override;
