-- name: InsertOneMysqlType :exec
INSERT INTO test_mysql_types (
    id, int_test, integer_test, mediumint_test, smallint_test, tinyint_test, bigint_test,
    int_unsigned_test, bigint_unsigned_test, year_test,
    tinyint1_test, bool_test, boolean_test,
    float_test, double_test, double_precision_test, real_test,
    decimal_test, numeric_test,
    char_test, varchar_test, tinytext_test, text_test, mediumtext_test, longtext_test,
    binary_test, varbinary_test, tinyblob_test, blob_test, mediumblob_test, longblob_test, bit_test,
    date_test, datetime_test, datetime6_test, timestamp_test, time_test,
    json_test, mood, tag
) VALUES (
             ?, ?, ?, ?, ?, ?, ?,
             ?, ?, ?,
             ?, ?, ?,
             ?, ?, ?, ?,
             ?, ?,
             ?, ?, ?, ?, ?, ?,
             ?, ?, ?, ?, ?, ?, ?,
             ?, ?, ?, ?, ?,
             ?, ?, ?
         );

-- name: InsertOneInnerMysqlType :exec
INSERT INTO test_inner_mysql_types (
    table_id, int_test, integer_test, mediumint_test, smallint_test, tinyint_test, bigint_test,
    int_unsigned_test, bigint_unsigned_test, year_test,
    tinyint1_test, bool_test, boolean_test,
    float_test, double_test, double_precision_test, real_test,
    decimal_test, numeric_test,
    char_test, varchar_test, tinytext_test, text_test, mediumtext_test, longtext_test,
    binary_test, varbinary_test, tinyblob_test, blob_test, mediumblob_test, longblob_test, bit_test,
    date_test, datetime_test, datetime6_test, timestamp_test, time_test,
    json_test, mood, tag
) VALUES (
             ?, ?, ?, ?, ?, ?, ?,
             ?, ?, ?,
             ?, ?, ?,
             ?, ?, ?, ?,
             ?, ?,
             ?, ?, ?, ?, ?, ?,
             ?, ?, ?, ?, ?, ?, ?,
             ?, ?, ?, ?, ?,
             ?, ?, ?
         );

-- name: GetOneMysqlType :one
SELECT * FROM test_mysql_types WHERE id = ?;

-- name: GetOneInnerMysqlType :one
SELECT * FROM test_inner_mysql_types WHERE table_id = ?;

-- name: GetManyMysqlType :many
SELECT * FROM test_mysql_types WHERE id = ?;

-- name: GetManyInnerMysqlType :many
SELECT * FROM test_inner_mysql_types WHERE table_id = ?;

-- name: GetManyNullableInnerMysqlType :many
SELECT * FROM test_inner_mysql_types WHERE table_id = ? AND int_test <=> ?;

-- name: GetOneDate :one
SELECT date_test FROM test_mysql_types WHERE id = ? AND date_test = ?;

-- name: GetOneDatetime :one
SELECT datetime_test FROM test_mysql_types WHERE id = ? AND datetime_test = ?;

-- name: GetOneTime :one
SELECT time_test FROM test_mysql_types WHERE id = ? AND time_test = ?;

-- name: GetOneBool :one
SELECT tinyint1_test FROM test_mysql_types WHERE id = ? AND tinyint1_test = ?;

-- name: GetOneDecimal :one
SELECT decimal_test FROM test_mysql_types WHERE id = ? AND decimal_test = ?;

-- name: GetOneBlob :one
SELECT blob_test FROM test_mysql_types WHERE id = ? AND blob_test = ?;

-- name: GetOneBit :one
SELECT bit_test FROM test_mysql_types WHERE id = ?;

-- name: GetOneYear :one
SELECT year_test FROM test_mysql_types WHERE id = ?;

-- name: GetOneJson :one
SELECT json_test FROM test_mysql_types WHERE id = ?;

-- name: GetOneMood :one
SELECT mood FROM test_mysql_types WHERE id = ? AND mood = ?;

-- name: GetOneTag :one
SELECT tag FROM test_mysql_types WHERE id = ?;

-- name: GetManyDate :many
SELECT date_test FROM test_mysql_types WHERE id = ? AND date_test = ?;

-- name: GetManyTime :many
SELECT time_test FROM test_mysql_types WHERE id = ? AND time_test = ?;

-- name: GetManyBool :many
SELECT tinyint1_test FROM test_mysql_types WHERE id = ? AND tinyint1_test = ?;

-- name: GetManyDecimal :many
SELECT decimal_test FROM test_mysql_types WHERE id = ? AND decimal_test = ?;

-- name: GetManyMood :many
SELECT mood FROM test_mysql_types WHERE mood = ? ORDER BY id;

-- Parameterless :many with literal percents: QueryResults always passes its
-- args tuple, so the constant must arrive with doubled "%%".
-- name: ListMonths :many
SELECT DATE_FORMAT(datetime_test, '%Y-%m') AS month FROM test_mysql_types ORDER BY id;

-- name: CountMysqlTypes :one
SELECT count(*) FROM test_mysql_types;

-- name: UpdateVarcharTest :execrows
UPDATE test_mysql_types SET varchar_test = ? WHERE id = ?;

-- name: DeleteOneMysqlType :exec
DELETE FROM test_mysql_types WHERE id = ?;

-- name: AllMysqlTypesCursor :execresult
SELECT * FROM test_mysql_types;

-- name: InsertExecLastId :execlastid
INSERT INTO test_execlastid (name) VALUES (?);

-- name: GetExecLastIdName :one
SELECT name FROM test_execlastid WHERE id = ?;

-- name: InsertTypeOverride :exec
INSERT INTO test_type_override (id, text_test) VALUES (?, ?);

-- name: GetTypeOverride :one
SELECT * FROM test_type_override WHERE id = ?;

-- name: GetReservedArg :one
SELECT * FROM test_reserved_args WHERE conn = ?;

-- name: InsertReservedArg :exec
INSERT INTO test_reserved_args (id, conn) VALUES (?, ?);
