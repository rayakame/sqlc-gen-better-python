-- name: InsertCaseRow :exec
INSERT INTO test_case_sensitivity (id, upper_dt, prec_dec) VALUES (?, ?, ?);

-- name: GetCaseRow :one
SELECT upper_dt, prec_dec FROM test_case_sensitivity WHERE id = ?;

-- name: InsertReservedArg :exec
INSERT INTO test_reserved_args (id, conn) VALUES (?, ?);

-- name: GetReservedArg :one
SELECT id FROM test_reserved_args WHERE conn = ?;
