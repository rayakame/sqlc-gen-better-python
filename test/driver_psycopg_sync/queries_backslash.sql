-- A backslash in the SQL has to survive into the generated Python constant.
-- PostgreSQL string literals are standard-conforming, so a backslash below is
-- an ordinary character.

-- name: GetBackslashPattern :one
SELECT 'a\tb\d+'::text AS pattern;

-- name: MatchBackslashPath :one
SELECT 'C:\dir\name'::text = sqlc.arg(probe)::text AS matches;
