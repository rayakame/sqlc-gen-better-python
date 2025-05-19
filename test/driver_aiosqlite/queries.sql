-- name: GetOneTestSqliteTypes :many
SELECT id FROM test_sqlite_types WHERE date_test = ? AND bool_test = ? AND datetime_test = ? AND decimal_test = ?;