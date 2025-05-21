CREATE EXTENSION IF NOT EXISTS citext;   -- citext
CREATE EXTENSION IF NOT EXISTS ltree;    -- ltree / lquery / ltxtquery

CREATE TABLE IF NOT EXISTS test_postgres_types
(
    /* ───────────── Integer family ───────────── */
    id                    int PRIMARY KEY  NOT NULL,
    serial_test           serial           NOT NULL,
    serial4_test          serial4          NOT NULL,
    bigserial_test        bigserial        NOT NULL,
    smallserial_test      smallserial      NOT NULL,
    int_test              int              NOT NULL,
    bigint_test           bigint           NOT NULL,
    smallint_test         smallint         NOT NULL,

    /* ───────────── Floating‑point ───────────── */
    float_test            float            NOT NULL,
    double_precision_test double precision NOT NULL,
    real_test             real             NOT NULL,

    /* ───────────── Exact numeric ───────────── */
    numeric_test          numeric(12, 4)   NOT NULL,
    money_test            money            NOT NULL,

    /* ───────────── Boolean ───────────── */
    bool_test             boolean          NOT NULL,

    /* ───────────── JSON / JSONB ───────────── */
    json_test             json             NOT NULL,
    jsonb_test            jsonb            NOT NULL,

    /* ───────────── Binary ───────────── */
    bytea_test            bytea            NOT NULL,

    /* ───────────── Date & time ───────────── */
    date_test             date             NOT NULL,
    time_test             time             NOT NULL,
    timetz_test           timetz           NOT NULL,
    timestamp_test        timestamp        NOT NULL,
    timestamptz_test      timestamptz      NOT NULL,
    interval_test interval NOT NULL,

    /* ───────────── Character / text ───────────── */
    text_test             text             NOT NULL,
    varchar_test          varchar(255)     NOT NULL,
    bpchar_test           bpchar(10)                         NOT NULL,
    char_test             char(1)          NOT NULL,
    citext_test           citext           NOT NULL,

    /* ───────────── UUID ───────────── */
    uuid_test             uuid             NOT NULL,

    /* ───────────── Network types ───────────── */
    inet_test             inet             NOT NULL,
    cidr_test             cidr             NOT NULL,
    macaddr_test          macaddr          NOT NULL,
    macaddr8_test         macaddr8         NOT NULL,

    /* ───────────── LTree family ───────────── */
    ltree_test            ltree            NOT NULL,
    lquery_test           lquery           NOT NULL,
    ltxtquery_test        ltxtquery        NOT NULL
);

CREATE TABLE IF NOT EXISTS test_inner_postgres_types
(
    /* ───────────── Integer family ───────────── */
    table_id              int              NOT NULL,
    serial_test           serial           ,
    serial4_test          serial4          ,
    bigserial_test        bigserial        ,
    smallserial_test      smallserial      ,
    int_test              int              ,
    bigint_test           bigint           ,
    smallint_test         smallint         ,

    /* ───────────── Floating‑point ───────────── */
    float_test            float            ,
    double_precision_test double precision ,
    real_test             real             ,

    /* ───────────── Exact numeric ───────────── */
    numeric_test          numeric(12, 4)   ,
    money_test            money            ,

    /* ───────────── Boolean ───────────── */
    bool_test             boolean          ,

    /* ───────────── JSON / JSONB ───────────── */
    json_test             json             ,
    jsonb_test            jsonb            ,

    /* ───────────── Binary ───────────── */
    bytea_test            bytea            ,

    /* ───────────── Date & time ───────────── */
    date_test             date             ,
    time_test             time             ,
    timetz_test           timetz           ,
    timestamp_test        timestamp        ,
    timestamptz_test      timestamptz      ,
    interval_test interval ,

    /* ───────────── Character / text ───────────── */
    text_test             text             ,
    varchar_test          varchar(255)     ,
    bpchar_test           bpchar(10)       ,
    char_test             char(1)          ,
    citext_test           citext           ,

    /* ───────────── UUID ───────────── */
    uuid_test             uuid             ,

    /* ───────────── Network types ───────────── */
    inet_test             inet             ,
    cidr_test             cidr             ,
    macaddr_test          macaddr          ,
    macaddr8_test         macaddr8         ,

    /* ───────────── LTree family ───────────── */
    ltree_test            ltree            ,
    lquery_test           lquery           ,
    ltxtquery_test        ltxtquery
);


CREATE TABLE IF NOT EXISTS test_copy_from
(
    id                    int PRIMARY KEY  NOT NULL,
    float_test            float           NOT NULL ,
    int_test              int NOT NULL
);