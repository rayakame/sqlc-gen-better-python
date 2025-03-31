-- name: GetAuthor :one
SELECT id, name
FROM authors
WHERE id = ? LIMIT 1;

-- name: ListAuthors :many
SELECT *
FROM authors
WHERE id IN (sqlc.slice('ids'))
ORDER BY name;

-- name: CreateAuthor :one
INSERT INTO authors (name, bio)
VALUES (?, ?) RETURNING *;

-- name: UpsertAuthorName :one
UPDATE authors
SET name = CASE WHEN sqlc.arg(set_name) THEN ? ELSE name END RETURNING *;

-- name: UpdateAuthor :exec
UPDATE authors
set name = ?,
    bio  = ?
WHERE id = ?;

-- name: DeleteAuthor :exec
DELETE
FROM authors
WHERE id = ?;

-- name: UpdateAuthorT :one
UPDATE authors
SET name = coalesce(sqlc.narg('name'), name),
    bio  = coalesce(sqlc.narg('bio'), bio)
WHERE id = sqlc.arg('id') RETURNING *;