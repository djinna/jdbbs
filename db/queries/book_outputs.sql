-- name: CreateBookOutput :one
INSERT INTO book_outputs (
    book_id, output_format, output_data, source_filename, spec_snapshot
) VALUES (
    ?, ?, ?, ?, ?
)
RETURNING id, book_id, output_format, output_data, source_filename, spec_snapshot, created_at;

-- name: ListBookOutputs :many
SELECT id, book_id, output_format, source_filename, length(output_data) AS size_bytes, spec_snapshot, created_at
FROM book_outputs
WHERE book_id = ?
ORDER BY created_at DESC, id DESC
LIMIT ?;

-- name: GetBookOutput :one
SELECT id, book_id, output_format, output_data, source_filename, spec_snapshot, created_at
FROM book_outputs
WHERE id = ? AND book_id = ?;
