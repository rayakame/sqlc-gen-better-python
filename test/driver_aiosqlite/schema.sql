CREATE TABLE test_sqlite_types
(
    /* ───────────── Integer family ───────────── */
    id                    integer PRIMARY KEY NOT NULL,
    int_test              int                 NOT NULL, -- covers integer / mediumint …
    bigint_test           bigint              NOT NULL, -- covers unsignedbigint …
    smallint_test         smallint            NOT NULL,
    tinyint_test          tinyint             NOT NULL,
    int2_test             int2                NOT NULL,
    int8_test             int8                NOT NULL,
    bigserial_test        bigserial           NOT NULL,
    /* ───────────── Binary (blob) ───────────── */
    blob_test             blob                NOT NULL,
    /* ───────────── Floating-point / numeric ───────────── */
    real_test             real                NOT NULL,
    double_test double NOT NULL,
    double_precision_test "double precision"  NOT NULL,
    float_test            float               NOT NULL,
    numeric_test          numeric             NOT NULL,
    /* ───────────── Exact numeric (decimal) ───────────── */
    decimal_test          decimal(12, 4)      NOT NULL,
    /* ───────────── Boolean ───────────── */
    boolean_test          boolean             NOT NULL,
    bool_test             bool                NOT NULL,
    /* ───────────── Date & time ───────────── */
    date_test             date                NOT NULL,
    datetime_test         datetime            NOT NULL,
    timestamp_test        timestamp           NOT NULL,
    /* ───────────── Character / text ───────────── */
    character_test        character(10)       NOT NULL,
    varchar_test          varchar(255)        NOT NULL,
    varyingcharacter_test varyingcharacter (255) NOT NULL,
    nchar_test            nchar(10)           NOT NULL,
    nativecharacter_test  nativecharacter (10) NOT NULL,
    nvarchar_test         nvarchar(255) NOT NULL,
    text_test             text                NOT NULL,
    clob_test             clob                NOT NULL,
    json_test             json                NOT NULL
)