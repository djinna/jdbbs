-- name: CreateBookOutput :one
INSERT INTO book_outputs (
    book_id, output_format, output_data, source_filename
) VALUES (
    ?, ?, ?, ?
)
RETURNING id, book_id, output_format, output_data, source_filename, created_at;

-- name: ListBookOutputs :many
SELECT id, book_id, output_format, output_data, source_filename, created_at
FROM book_outputs
WHERE book_id = ? AND output_format = ?
ORDER BY created_at DESC, id DESC;
