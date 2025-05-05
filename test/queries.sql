-- name: UpdateAuthor :exec
UPDATE authors
set name = $1,
    bio  = $2
WHERE id = $3;

-- name: DeleteAuthor :exec
DELETE
FROM authors
WHERE id = $1;

-- name: GetStudentAndScore :one
SELECT sqlc.embed(students), sqlc.embed(test_scores)
FROM students
         JOIN test_scores ON test_scores.student_id = students.id
WHERE students.id = $1;

-- name: GetStudentAndScores :many
SELECT sqlc.embed(students), sqlc.embed(test_scores)
FROM students
         JOIN test_scores ON test_scores.student_id = students.id;

-- name: ListAuthors :many
SELECT authors.id
FROM authors
WHERE id IN (sqlc.slice('ids'))
ORDER BY name;


-- name: ListAuthors2 :many
SELECT authors.id
FROM authors
WHERE id = $1
ORDER BY name;