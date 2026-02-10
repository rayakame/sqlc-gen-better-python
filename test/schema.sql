CREATE TABLE test_postgres_types
(
    /* ───────────── Integer family ───────────── */
    id                    int PRIMARY KEY  NOT NULL,
    serial_test           serial           NOT NULL,
    timestamp_test        timestamp        NOT NULL
);

CREATE TYPE mood AS ENUM ('sad', 'ok', 'happy');

CREATE TABLE test_inner_postgres_types
(
    /* ───────────── Integer family ───────────── */
    table_id              int              NOT NULL,
    /* ───────────── Boolean ───────────── */
    bool_test             boolean          NOT NULL
);

CREATE TABLE test_enum
(
    /* ───────────── Integer family ───────────── */
    id int PRIMARY KEY NOT NULL,
    m  mood            NOT NULL
);


