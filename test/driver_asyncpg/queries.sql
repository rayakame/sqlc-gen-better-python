-- name: GetOneTestPostgresType :one
SELECT *
FROM test_postgres_types
WHERE id = $1 LIMIT 1;

-- name: GetOneInnerTestPostgresType :one
SELECT *
FROM test_inner_postgres_types
WHERE table_id = $1 LIMIT 1;

-- name: GetOneTestTimestampPostgresType :one
SELECT timestamp_test
FROM test_postgres_types
WHERE id = $1 LIMIT 1;

-- name: GetOneTestByteaPostgresType :one
SELECT bytea_test
FROM test_postgres_types
WHERE id = $1 LIMIT 1;

-- name: GetManyTestPostgresType :many
SELECT *
FROM test_postgres_types
WHERE id = $1;

-- name: GetManyTestIteratorPostgresType :many
SELECT *
FROM test_postgres_types
WHERE id = $1;

-- name: GetManyTestTimestampPostgresType :many
SELECT timestamp_test
FROM test_postgres_types
WHERE id = $1 LIMIT 2;

-- name: GetManyTestByteaPostgresType :many
SELECT bytea_test
FROM test_postgres_types
WHERE id = $1 LIMIT 2;

-- name: GetEmbeddedTestPostgresType :one
SELECT test_postgres_types.*, sqlc.embed(test_inner_postgres_types)
FROM test_postgres_types
         JOIN test_inner_postgres_types ON test_inner_postgres_types.table_id = test_postgres_types.id
WHERE test_postgres_types.id = $1;

-- name: GetAllEmbeddedTestPostgresType :one
SELECT sqlc.embed(test_postgres_types), sqlc.embed(test_inner_postgres_types)
FROM test_postgres_types
         JOIN test_inner_postgres_types ON test_inner_postgres_types.table_id = test_postgres_types.id
WHERE test_postgres_types.id = $1;

-- name: CreateOneTestPostgresType :exec
INSERT INTO test_postgres_types (id,
                                 serial_test,
                                 serial4_test,
                                 bigserial_test,
                                 smallserial_test,
                                 int_test,
                                 bigint_test,
                                 smallint_test,
                                 float_test,
                                 double_precision_test,
                                 real_test,
                                 numeric_test,
                                 money_test,
                                 bool_test,
                                 json_test,
                                 jsonb_test,
                                 bytea_test,
                                 date_test,
                                 time_test,
                                 timetz_test,
                                 timestamp_test,
                                 timestamptz_test,
                                 interval_test,
                                 text_test,
                                 varchar_test,
                                 bpchar_test,
                                 char_test,
                                 citext_test,
                                 uuid_test,
                                 inet_test,
                                 cidr_test,
                                 macaddr_test,
                                 macaddr8_test,
                                 ltree_test,
                                 lquery_test,
                                 ltxtquery_test)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8,
        $9, $10, $11, $12, $13, $14, $15, $16,
        $17, $18, $19, $20, $21, $22, $23, $24,
        $25, $26, $27, $28, $29, $30, $31, $32,
        $33, $34, $35, $36);

-- name: CreateOneTestPostgresInnerType :exec
INSERT INTO test_inner_postgres_types (table_id,
                                       serial_test,
                                       serial4_test,
                                       bigserial_test,
                                       smallserial_test,
                                       int_test,
                                       bigint_test,
                                       smallint_test,
                                       float_test,
                                       double_precision_test,
                                       real_test,
                                       numeric_test,
                                       money_test,
                                       bool_test,
                                       json_test,
                                       jsonb_test,
                                       bytea_test,
                                       date_test,
                                       time_test,
                                       timetz_test,
                                       timestamp_test,
                                       timestamptz_test,
                                       interval_test,
                                       text_test,
                                       varchar_test,
                                       bpchar_test,
                                       char_test,
                                       citext_test,
                                       uuid_test,
                                       inet_test,
                                       cidr_test,
                                       macaddr_test,
                                       macaddr8_test,
                                       ltree_test,
                                       lquery_test,
                                       ltxtquery_test)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8,
        $9, $10, $11, $12, $13, $14, $15, $16,
        $17, $18, $19, $20, $21, $22, $23, $24,
        $25, $26, $27, $28, $29, $30, $31, $32,
        $33, $34, $35, $36);

-- name: DeleteOneTestPostgresType :exec
DELETE
FROM test_postgres_types
WHERE test_postgres_types.id = $1;


-- name: DeleteOneTestPostgresInnerType :exec
DELETE
FROM test_inner_postgres_types
WHERE test_inner_postgres_types.table_id = $1;

-- name: CreateResultOneTestPostgresType :execresult
INSERT INTO test_postgres_types (id,
                                 serial_test,
                                 serial4_test,
                                 bigserial_test,
                                 smallserial_test,
                                 int_test,
                                 bigint_test,
                                 smallint_test,
                                 float_test,
                                 double_precision_test,
                                 real_test,
                                 numeric_test,
                                 money_test,
                                 bool_test,
                                 json_test,
                                 jsonb_test,
                                 bytea_test,
                                 date_test,
                                 time_test,
                                 timetz_test,
                                 timestamp_test,
                                 timestamptz_test,
                                 interval_test,
                                 text_test,
                                 varchar_test,
                                 bpchar_test,
                                 char_test,
                                 citext_test,
                                 uuid_test,
                                 inet_test,
                                 cidr_test,
                                 macaddr_test,
                                 macaddr8_test,
                                 ltree_test,
                                 lquery_test,
                                 ltxtquery_test)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8,
        $9, $10, $11, $12, $13, $14, $15, $16,
        $17, $18, $19, $20, $21, $22, $23, $24,
        $25, $26, $27, $28, $29, $30, $31, $32,
        $33, $34, $35, $36);

-- name: UpdateResultTestPostgresType :execresult
UPDATE test_postgres_types
SET serial_test = 187
WHERE test_postgres_types.id = $1;

-- name: DeleteOneResultTestPostgresType :execresult
DELETE
FROM test_postgres_types
WHERE test_postgres_types.id = $1;

-- name: CreateRowsOneTestPostgresType :execrows
INSERT INTO test_postgres_types (id,
                                 serial_test,
                                 serial4_test,
                                 bigserial_test,
                                 smallserial_test,
                                 int_test,
                                 bigint_test,
                                 smallint_test,
                                 float_test,
                                 double_precision_test,
                                 real_test,
                                 numeric_test,
                                 money_test,
                                 bool_test,
                                 json_test,
                                 jsonb_test,
                                 bytea_test,
                                 date_test,
                                 time_test,
                                 timetz_test,
                                 timestamp_test,
                                 timestamptz_test,
                                 interval_test,
                                 text_test,
                                 varchar_test,
                                 bpchar_test,
                                 char_test,
                                 citext_test,
                                 uuid_test,
                                 inet_test,
                                 cidr_test,
                                 macaddr_test,
                                 macaddr8_test,
                                 ltree_test,
                                 lquery_test,
                                 ltxtquery_test)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8,
        $9, $10, $11, $12, $13, $14, $15, $16,
        $17, $18, $19, $20, $21, $22, $23, $24,
        $25, $26, $27, $28, $29, $30, $31, $32,
        $33, $34, $35, $36);

-- name: UpdateRowsTestPostgresType :execrows
UPDATE test_postgres_types
SET serial_test = 187
WHERE test_postgres_types.id = $1;

-- name: DeleteOneRowsTestPostgresType :execrows
DELETE
FROM test_postgres_types
WHERE test_postgres_types.id = $1;

-- name: CreateRowsTable :execrows
CREATE TABLE test_create_rows_table
(
    id   int PRIMARY KEY NOT NULL,
    test int             NOT NULL
);



-- name: TestCopyFrom :copyfrom
INSERT INTO test_copy_from (id,
                            float_test, int_test)
VALUES ($1, $2, $3);
