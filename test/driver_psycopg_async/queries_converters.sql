-- name: InsertConverted :exec
INSERT INTO test_converters (id, prefs, maybe_prefs, tags) VALUES ($1, $2, $3, $4);

-- name: GetConverted :one
SELECT * FROM test_converters WHERE id = $1;

-- name: ListConvertedByTags :many
SELECT id FROM test_converters WHERE tags = $1;

-- name: DeleteConverted :exec
DELETE FROM test_converters WHERE id = $1;

-- name: InsertConverterArrayRow :exec
INSERT INTO test_converter_array (id, label) VALUES ($1, $2);

-- name: GetConverterArrayLabel :one
SELECT label FROM test_converter_array WHERE id = $1;

-- name: FindConverterArrayByLabels :many
SELECT id FROM test_converter_array WHERE label = ANY($1::converter_label[]);
