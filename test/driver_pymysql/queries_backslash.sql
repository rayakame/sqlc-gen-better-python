-- A backslash in the SQL has to survive into the generated Python constant.
-- MySQL reads a backslash in a string literal as an escape, so each doubled
-- backslash below is one literal backslash on the server.

-- name: InsertBackslashRow :exec
INSERT INTO test_slice (id, name, note) VALUES (?, ?, ?);

-- name: GetBackslashPattern :one
SELECT 'a\\tb\\d+' AS pattern;

-- name: GetBackslashNote :one
SELECT note FROM test_slice WHERE note = 'C:\\dir\\name' AND id = ?;
