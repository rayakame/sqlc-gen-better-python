-- name: InsertCaseRow :exec
INSERT INTO test_case_sensitivity (id, upper_dt, prec_dec) VALUES (?, ?, ?);

-- name: GetCaseRow :one
SELECT * FROM test_case_sensitivity WHERE id = ?;

-- Placeholder inside an executable /*! version comment: the body is live
-- SQL to both MySQL and sqlc, so the ? must become a real %s.
-- name: CountCaseRows :one
SELECT count(*) FROM test_case_sensitivity /*! WHERE id >= ? */;
