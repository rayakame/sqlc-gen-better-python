-- name: GetFieldNaming :one
SELECT *
FROM test_field_namings
WHERE id = $1 LIMIT 1;

-- name: GetJoinedFieldNamings :one
SELECT a.outputs, b.outputs
FROM test_field_namings a
JOIN test_field_namings b ON a.id = b.id
WHERE a.id = $1 LIMIT 1;

-- name: SetFieldNamingOutputs :exec
UPDATE test_field_namings
SET outputs = $2
WHERE id = $1 AND outputs <> sqlc.arg(outputs)::jsonb;
