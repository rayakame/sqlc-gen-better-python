-- Public-schema enum
CREATE TYPE mood AS ENUM ('sad', 'ok', 'happy');

-- Primary table — exercises non-null, nullable, enum, array, and the
-- three types that need runtime conversion (bytea, numeric, timestamptz).
CREATE TABLE authors
(
    id      int PRIMARY KEY NOT NULL,
    name    text            NOT NULL,
    bio     text,
    mood    mood            NOT NULL,
    tags    text[]          NOT NULL,
    avatar  bytea,
    rating  numeric(3, 2)   NOT NULL,
    created timestamptz     NOT NULL
);

-- Secondary table — gives us something to JOIN / sqlc.embed against.
CREATE TABLE books
(
    id        int PRIMARY KEY NOT NULL,
    author_id int             NOT NULL,
    title     text            NOT NULL
);

-- Custom schema — same enum name in a different schema (tests
-- schema-qualified naming and that the generator doesn't collide).
CREATE SCHEMA IF NOT EXISTS custom;
CREATE TYPE custom.mood AS ENUM ('sad', 'ok', 'happy');

CREATE TABLE custom.reviews
(
    id      int PRIMARY KEY NOT NULL,
    book_id int             NOT NULL,
    mood    custom.mood     NOT NULL
);
