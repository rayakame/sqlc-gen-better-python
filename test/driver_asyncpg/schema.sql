CREATE TABLE test_postgres_types
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

CREATE TABLE test_inner_postgres_types
(
    /* ───────────── Integer family ───────────── */
    table_id              int              NOT NULL,
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
    bpchar_test           bpchar(10)       NOT NULL,
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