-- A backslash in the SQL has to survive into the generated Python constant.
-- SQLite string literals have no backslash escape, so every backslash below
-- reaches the database as written.

-- name: InsertBackslashRow :exec
INSERT INTO test_slice (id, name, note) VALUES (?, ?, ?);

-- name: GetBackslashPattern :one
SELECT 'a\tb\d+' AS pattern;

-- name: GetBackslashNote :one
SELECT note FROM test_slice WHERE note = 'C:\dir\name' AND id = ?;
