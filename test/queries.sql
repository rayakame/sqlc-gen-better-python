-- ─── :exec / :execresult / :execrows / :copyfrom ────────────────────────

-- name: InsertAuthor :exec
INSERT INTO authors (id, name, bio, mood, tags, avatar, rating, created)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: UpdateAuthorName :execresult
UPDATE authors SET name = $2 WHERE id = $1;

-- name: DeleteAuthorsByMood :execrows
DELETE FROM authors WHERE mood = $1;

-- name: CopyAuthors :copyfrom
INSERT INTO authors (id, name, mood, tags, rating, created)
VALUES ($1, $2, $3, $4, $5, $6);


-- ─── :one returning a tracked table (should reference models.X) ─────────

-- name: GetAuthor :one
SELECT * FROM authors WHERE id = $1;

-- name: GetBook :one
SELECT * FROM books WHERE id = $1;

-- name: GetReview :one
SELECT * FROM custom.reviews WHERE id = $1;


-- ─── :one returning a scalar ────────────────────────────────────────────

-- name: GetAuthorName :one
SELECT name FROM authors WHERE id = $1;

-- name: GetAuthorBio :one
SELECT bio FROM authors WHERE id = $1;

-- name: GetAuthorMood :one
SELECT mood FROM authors WHERE id = $1;

-- name: GetAuthorTags :one
SELECT tags FROM authors WHERE id = $1;

-- name: GetAuthorAvatar :one
SELECT avatar FROM authors WHERE id = $1;


-- ─── :one returning an inline-defined struct (no table match) ───────────

-- name: GetAuthorIdAndName :one
SELECT id, name FROM authors WHERE id = $1;

-- name: GetBookWithAuthorName :one
SELECT b.id, b.title, a.name AS author_name
FROM books b
         JOIN authors a ON a.id = b.author_id
WHERE b.id = $1;


-- ─── sqlc.embed() ───────────────────────────────────────────────────────

-- name: GetBookWithAuthor :one
SELECT sqlc.embed(books), sqlc.embed(authors)
FROM books
         JOIN authors ON authors.id = books.author_id
WHERE books.id = $1;

-- name: ListBooksWithAuthor :many
SELECT sqlc.embed(books), sqlc.embed(authors)
FROM books
         JOIN authors ON authors.id = books.author_id;


-- ─── :many ──────────────────────────────────────────────────────────────

-- name: ListAuthors :many
SELECT * FROM authors;

-- name: ListAuthorIds :many
SELECT id FROM authors;

-- name: ListAuthorsByMood :many
SELECT * FROM authors WHERE mood = $1;

-- name: ListBookTitlesByAuthor :many
SELECT b.title, a.name AS author_name
FROM books b
         JOIN authors a ON a.id = b.author_id
WHERE a.id = $1;

-- name: ListReviewsForBook :many
SELECT * FROM custom.reviews WHERE book_id = $1;


-- ─── Misc ───────────────────────────────────────────────────────────────

-- name: InsertReview :exec
INSERT INTO custom.reviews (id, book_id, mood)
VALUES ($1, $2, $3);

-- name: CountAuthors :one
SELECT COUNT(*) FROM authors;
