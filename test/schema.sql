CREATE TABLE authors
(
    id   INTEGER PRIMARY KEY,
    name text NOT NULL,
    bio  text
);

CREATE TABLE students (
      id   bigserial PRIMARY KEY,
      name text,
      age  integer
);

CREATE TABLE test_scores (
     student_id bigint,
     score integer,
     grade text
);