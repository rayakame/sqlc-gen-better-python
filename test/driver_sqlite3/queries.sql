-- name: InsertOneSqliteType :exec
INSERT INTO test_sqlite_types (
    id, int_test, bigint_test, smallint_test, tinyint_test, int2_test, int8_test, bigserial_test,
    blob_test, real_test, double_test, double_precision_test, float_test, numeric_test, decimal_test,
    boolean_test, bool_test, date_test, datetime_test, timestamp_test,
    character_test, varchar_test, varyingcharacter_test, nchar_test, nativecharacter_test,
    nvarchar_test, text_test, clob_test, json_test
) VALUES (
             ?, ?, ?, ?, ?, ?, ?, ?,
             ?, ?, ?, ?, ?, ?, ?,
             ?, ?, ?, ?, ?,
             ?, ?, ?, ?, ?, ?, ?, ?, ?
         );

-- name: InsertOneInnerSqliteType :exec
INSERT INTO test_inner_sqlite_types (
    table_id, int_test, bigint_test, smallint_test, tinyint_test, int2_test, int8_test, bigserial_test,
    blob_test, real_test, double_test, double_precision_test, float_test, numeric_test, decimal_test,
    boolean_test, bool_test, date_test, datetime_test, timestamp_test,
    character_test, varchar_test, varyingcharacter_test, nchar_test, nativecharacter_test,
    nvarchar_test, text_test, clob_test, json_test
) VALUES (
             ?, ?, ?, ?, ?, ?, ?, ?,
             ?, ?, ?, ?, ?, ?, ?,
             ?, ?, ?, ?, ?,
             ?, ?, ?, ?, ?, ?, ?, ?, ?
         );

-- name: GetOneSqliteType :one
SELECT * FROM test_sqlite_types WHERE id = ?;

-- name: GetOneInnerSqliteType :one
SELECT * FROM test_inner_sqlite_types WHERE table_id = ?;

-- name: GetOneDate :one
SELECT date_test FROM test_sqlite_types WHERE id = ? AND date_test = ?;

-- name: GetOneDatetime :one
SELECT datetime_test FROM test_sqlite_types WHERE id = ? AND datetime_test = ?;

-- name: GetOneTimestamp :one
SELECT timestamp_test FROM test_sqlite_types WHERE id = ? AND timestamp_test = ?;

-- name: GetOneBool :one
SELECT bool_test FROM test_sqlite_types WHERE id = ? AND bool_test = ?;

-- name: GetOneBoolean :one
SELECT boolean_test FROM test_sqlite_types WHERE id = ? AND boolean_test = ?;

-- name: GetOneDecimal :one
SELECT decimal_test FROM test_sqlite_types WHERE id = ? AND decimal_test = ?;

-- name: GetOneBlob :one
SELECT blob_test FROM test_sqlite_types WHERE id = ? AND blob_test = ?;

-- name: GetManySqliteType :many
SELECT * FROM test_sqlite_types WHERE id = ?;

-- name: GetManyInnerSqliteType :many
SELECT * FROM test_inner_sqlite_types WHERE table_id = ?;

-- name: GetManyDate :many
SELECT date_test FROM test_sqlite_types WHERE id = ? AND date_test = ?;

-- name: GetManyDatetime :many
SELECT datetime_test FROM test_sqlite_types WHERE id = ? AND datetime_test = ?;

-- name: GetManyTimestamp :many
SELECT timestamp_test FROM test_sqlite_types WHERE id = ? AND timestamp_test = ?;

-- name: GetManyBool :many
SELECT bool_test FROM test_sqlite_types WHERE id = ? AND bool_test = ?;

-- name: GetManyBoolean :many
SELECT boolean_test FROM test_sqlite_types WHERE id = ? AND boolean_test = ?;

-- name: GetManyDecimal :many
SELECT decimal_test FROM test_sqlite_types WHERE id = ? AND decimal_test = ?;

-- name: GetManyBlob :many
SELECT blob_test FROM test_sqlite_types WHERE id = ? AND blob_test = ?;

-- name: DeleteOneSqliteType :exec
DELETE
FROM test_sqlite_types
WHERE test_sqlite_types.id = ?;

-- name: DeleteOneTestInnerSqliteType :exec
DELETE FROM test_inner_sqlite_types
WHERE test_inner_sqlite_types.table_id = ?;

-- name: InsertResultOneSqliteType :execresult
INSERT INTO test_sqlite_types (
    id, int_test, bigint_test, smallint_test, tinyint_test, int2_test, int8_test, bigserial_test,
    blob_test, real_test, double_test, double_precision_test, float_test, numeric_test, decimal_test,
    boolean_test, bool_test, date_test, datetime_test, timestamp_test,
    character_test, varchar_test, varyingcharacter_test, nchar_test, nativecharacter_test,
    nvarchar_test, text_test, clob_test, json_test
) VALUES (
             ?, ?, ?, ?, ?, ?, ?, ?,
             ?, ?, ?, ?, ?, ?, ?,
             ?, ?, ?, ?, ?,
             ?, ?, ?, ?, ?, ?, ?, ?, ?
         );

-- name: UpdateResultOneSqliteType :execresult
UPDATE test_sqlite_types
SET int_test = 187
WHERE test_sqlite_types.id = ?;

-- name: DeleteResultOneSqliteType :execresult
DELETE
FROM test_sqlite_types
WHERE test_sqlite_types.id = ?;

-- name: InsertRowsOneSqliteType :execrows
INSERT INTO test_sqlite_types (
    id, int_test, bigint_test, smallint_test, tinyint_test, int2_test, int8_test, bigserial_test,
    blob_test, real_test, double_test, double_precision_test, float_test, numeric_test, decimal_test,
    boolean_test, bool_test, date_test, datetime_test, timestamp_test,
    character_test, varchar_test, varyingcharacter_test, nchar_test, nativecharacter_test,
    nvarchar_test, text_test, clob_test, json_test
) VALUES (
             ?, ?, ?, ?, ?, ?, ?, ?,
             ?, ?, ?, ?, ?, ?, ?,
             ?, ?, ?, ?, ?,
             ?, ?, ?, ?, ?, ?, ?, ?, ?
         );

-- name: UpdateRowsOneSqliteType :execrows
UPDATE test_sqlite_types
SET int_test = 187
WHERE test_sqlite_types.id = ?;

-- name: DeleteRowsOneSqliteType :execrows
DELETE
FROM test_sqlite_types
WHERE test_sqlite_types.id = ?;

-- name: CreateRowsTable :execrows
CREATE TABLE test_create_rows_table
(
    id   int PRIMARY KEY NOT NULL,
    test int             NOT NULL
);

-- name: InsertLastIdOneSqliteType :execlastid
INSERT INTO test_sqlite_types (
    id, int_test, bigint_test, smallint_test, tinyint_test, int2_test, int8_test, bigserial_test,
    blob_test, real_test, double_test, double_precision_test, float_test, numeric_test, decimal_test,
    boolean_test, bool_test, date_test, datetime_test, timestamp_test,
    character_test, varchar_test, varyingcharacter_test, nchar_test, nativecharacter_test,
    nvarchar_test, text_test, clob_test, json_test
) VALUES (
             ?, ?, ?, ?, ?, ?, ?, ?,
             ?, ?, ?, ?, ?, ?, ?,
             ?, ?, ?, ?, ?,
             ?, ?, ?, ?, ?, ?, ?, ?, ?
         );

-- name: UpdateLastIdOneSqliteType :execlastid
UPDATE test_sqlite_types
SET int_test = 187
WHERE test_sqlite_types.id = ?;

-- name: DeleteLastIdOneSqliteType :execlastid
DELETE
FROM test_sqlite_types
WHERE test_sqlite_types.id = ?;



