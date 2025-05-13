-- name: GetEmbeddedTestPostgresType1 :one
SELECT *, sqlc.embed(test_inner_postgres_types)
FROM test_postgres_types
         JOIN test_inner_postgres_types ON test_inner_postgres_types.table_id = test_postgres_types.id
    LIMIT 1;

-- name: GetEmbeddedTestPostgresType2 :one
SELECT test_postgres_types.*, sqlc.embed(test_inner_postgres_types), test_inner_postgres_types.bool_test
FROM test_postgres_types
         JOIN test_inner_postgres_types ON test_inner_postgres_types.table_id = test_postgres_types.id
    LIMIT 1;

-- name: GetEmbeddedTestPostgresType3 :one
SELECT test_postgres_types.id, test_postgres_types.serial_test, sqlc.embed(test_inner_postgres_types), test_inner_postgres_types.bool_test
FROM test_postgres_types
         JOIN test_inner_postgres_types ON test_inner_postgres_types.table_id = test_postgres_types.id
    LIMIT 1;

-- name: GetEmbeddedTestPostgresType4 :one
SELECT sqlc.embed(test_postgres_types), sqlc.embed(test_inner_postgres_types), test_inner_postgres_types.bool_test
FROM test_postgres_types
         JOIN test_inner_postgres_types ON test_inner_postgres_types.table_id = test_postgres_types.id
    LIMIT 1;

-- name: TTTT :one
SELECT timestamp_test FROM test_postgres_types LIMIT 1;