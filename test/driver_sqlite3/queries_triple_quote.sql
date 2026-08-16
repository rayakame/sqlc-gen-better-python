-- The SQL goes into a Python literal, which cannot be delimited by a quote run
-- the query itself contains. The first query forces the other delimiter; the
-- second holds both and leaves no block spelling at all.

-- name: GetTripleQuotePattern :one
SELECT 'a"""b' AS pattern FROM test_slice WHERE id = ?;

-- name: GetBothQuotesPattern :one
SELECT 'a"""b''''''c' AS pattern FROM test_slice WHERE id = ?;
