CREATE TABLE IF NOT EXISTS test_mysql_types
(
    /* ------------- Integer family ------------- */
    id                    bigint PRIMARY KEY NOT NULL,
    int_test              int                NOT NULL,
    integer_test          integer            NOT NULL,
    mediumint_test        mediumint          NOT NULL,
    smallint_test         smallint           NOT NULL,
    tinyint_test          tinyint            NOT NULL, -- plain tinyint stays int
    bigint_test           bigint             NOT NULL,
    int_unsigned_test     int unsigned       NOT NULL,
    bigint_unsigned_test  bigint unsigned    NOT NULL,
    year_test             year               NOT NULL,
    /* ------------- Boolean (tinyint(1) and its aliases) ------------- */
    tinyint1_test         tinyint(1)         NOT NULL,
    bool_test             bool               NOT NULL,
    boolean_test          boolean            NOT NULL,
    /* ------------- Floating-point ------------- */
    float_test            float              NOT NULL,
    double_test           double             NOT NULL,
    double_precision_test double precision   NOT NULL,
    real_test             real               NOT NULL,
    /* ------------- Exact numeric (decimal) ------------- */
    decimal_test          decimal(12,4)      NOT NULL,
    numeric_test          numeric(10,2)      NOT NULL,
    /* ------------- Character / text ------------- */
    char_test             char(10)           NOT NULL,
    varchar_test          varchar(255)       NOT NULL,
    tinytext_test         tinytext           NOT NULL,
    text_test             text               NOT NULL,
    mediumtext_test       mediumtext         NOT NULL,
    longtext_test         longtext           NOT NULL,
    /* ------------- Binary ------------- */
    binary_test           binary(16)         NOT NULL,
    varbinary_test        varbinary(255)     NOT NULL,
    tinyblob_test         tinyblob           NOT NULL,
    blob_test             blob               NOT NULL,
    mediumblob_test       mediumblob         NOT NULL,
    longblob_test         longblob           NOT NULL,
    bit_test              bit(8)             NOT NULL,
    /* ------------- Date & time (time maps to timedelta) ------------- */
    date_test             date               NOT NULL,
    datetime_test         datetime           NOT NULL,
    datetime6_test        datetime(6)        NOT NULL,
    timestamp_test        timestamp          NOT NULL,
    time_test             time               NOT NULL,
    /* ------------- JSON (kept as str) ------------- */
    json_test             json               NOT NULL,
    /* ------------- Inline enum and set ------------- */
    -- '24h' and '_hidden' pin the digit- and underscore-leading constant
    -- names of the synthesized test_mysql_types_mood enum class.
    mood                  enum('sad','ok','happy','24h','_hidden') NOT NULL,
    -- SET columns become StrEnums like enum columns (sqlc materializes
    -- both). Only single-valued sets round-trip, see the docs.
    tag                   set('alpha','beta','gamma') NOT NULL
);

CREATE TABLE IF NOT EXISTS test_inner_mysql_types
(
    table_id              bigint PRIMARY KEY NOT NULL,
    int_test              int,
    integer_test          integer,
    mediumint_test        mediumint,
    smallint_test         smallint,
    tinyint_test          tinyint,
    bigint_test           bigint,
    int_unsigned_test     int unsigned,
    bigint_unsigned_test  bigint unsigned,
    year_test             year,
    tinyint1_test         tinyint(1),
    bool_test             bool,
    boolean_test          boolean,
    float_test            float,
    double_test           double,
    double_precision_test double precision,
    real_test             real,
    decimal_test          decimal(12,4),
    numeric_test          numeric(10,2),
    char_test             char(10),
    varchar_test          varchar(255),
    tinytext_test         tinytext,
    text_test             text,
    mediumtext_test       mediumtext,
    longtext_test         longtext,
    binary_test           binary(16),
    varbinary_test        varbinary(255),
    tinyblob_test         tinyblob,
    blob_test             blob,
    mediumblob_test       mediumblob,
    longblob_test         longblob,
    bit_test              bit(8),
    date_test             date,
    datetime_test         datetime,
    datetime6_test        datetime(6),
    timestamp_test        timestamp,
    time_test             time,
    json_test             json,
    mood                  enum('sad','ok','happy','24h','_hidden'),
    tag                   set('alpha','beta','gamma')
);

CREATE TABLE IF NOT EXISTS test_type_override
(
    id                    bigint PRIMARY KEY NOT NULL,
    text_test             text
);

-- Enum column with a py_type override: the override wins over the
-- synthesized enum class, and parameters convert back through it.
CREATE TABLE IF NOT EXISTS test_enum_override
(
    id                    bigint PRIMARY KEY NOT NULL,
    mood_test             enum('sad','ok','happy') NOT NULL
);

-- Uppercase type names and precision variants exercise the SQL-type
-- normalization. The version-comment query in queries_case.sql lives on
-- this table too.
CREATE TABLE IF NOT EXISTS test_case_sensitivity
(
    id                    bigint PRIMARY KEY NOT NULL,
    upper_dt              DATETIME           NOT NULL,
    prec_dec              DECIMAL(10,2)      NOT NULL
);

-- A column named like the implicit first argument of generated functions.
CREATE TABLE IF NOT EXISTS test_reserved_args
(
    id                    bigint PRIMARY KEY NOT NULL,
    conn                  varchar(64)        NOT NULL
);

-- :execlastid reads cursor.lastrowid from the AUTO_INCREMENT key. Serial
-- is the bigint unsigned AUTO_INCREMENT alias.
CREATE TABLE IF NOT EXISTS test_execlastid
(
    id                    serial PRIMARY KEY,
    name                  varchar(64)        NOT NULL
);

-- Plural column name: field names must NOT be singularized (only table
-- names and embed fields are). Ported from PR 164.
CREATE TABLE IF NOT EXISTS test_field_namings
(
    id                    bigint PRIMARY KEY NOT NULL,
    outputs               json               NOT NULL
);

-- Backtick-quoted identifiers that are not valid Python names (issue 160).
CREATE TABLE IF NOT EXISTS test_invalid_identifiers
(
    id                    bigint PRIMARY KEY NOT NULL,
    `3p%`                 text,
    `new notes`           text               NOT NULL,
    `%pct`                text
);

-- Digit-leading table name: the class gets a Model prefix (Model3RdPartyStat).
CREATE TABLE IF NOT EXISTS `3rd_party_stats`
(
    id                    bigint PRIMARY KEY NOT NULL,
    total                 bigint             NOT NULL
);

-- Variable-length IN lists via sqlc.slice: the /*SLICE:name*/ placeholder in
-- the SQL constant is expanded at call time, one "%s" per element.
CREATE TABLE IF NOT EXISTS test_slice
(
    id                    bigint PRIMARY KEY NOT NULL,
    name                  varchar(64)        NOT NULL,
    note                  varchar(64)
);

CREATE TABLE IF NOT EXISTS test_converters
(
    id                    bigint PRIMARY KEY NOT NULL,
    prefs                 json               NOT NULL,
    maybe_prefs           json,
    tags                  text               NOT NULL
);
