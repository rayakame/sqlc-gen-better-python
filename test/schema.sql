-- Public schema (default)
CREATE TYPE mood AS ENUM ('sad', 'ok', 'happy');

CREATE TABLE test_enum
(
    id int PRIMARY KEY NOT NULL,
    b boolean NOT NULL,
    b2 boolean,
    m  mood NOT NULL
);

-- Custom schema
CREATE SCHEMA IF NOT EXISTS custom;

CREATE TYPE custom.mood AS ENUM ('sad', 'ok', 'happy');

CREATE TABLE custom.test_enum
(
    id int PRIMARY KEY NOT NULL,
    b boolean NOT NULL,
    b2 boolean,
    m  custom.mood NOT NULL
);