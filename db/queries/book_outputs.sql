-- name: CreateBookOutput :one
INSERT INTO book_outputs (
    book_id, output_format, output_data, source_filename, spec_snapshot, corrections_snapshot
) VALUES (
    ?, ?, ?, ?, ?, ?
)
RETURNING id, book_id, output_format, output_data, source_filename, spec_snapshot, corrections_snapshot, created_at;

-- name: ListBookOutputs :many
SELECT id, book_id, output_format, source_filename, length(output_data) AS size_bytes, spec_snapshot, corrections_snapshot, created_at
FROM book_outputs
WHERE book_id = ?
ORDER BY created_at DESC, id DESC
LIMIT ?;

-- name: GetBookOutput :one
SELECT id, book_id, output_format, output_data, source_filename, spec_snapshot, corrections_snapshot, created_at
FROM book_outputs
WHERE id = ? AND book_id = ?;

-- name: PruneBookOutputs :exec
-- Keep the newest `keep` outputs of one format for a book; delete the rest.
-- Used for EPUBs, which are unlimited per pass and would otherwise pile up.
DELETE FROM book_outputs
WHERE book_outputs.book_id = ?1 AND book_outputs.output_format = ?2
  AND book_outputs.id NOT IN (
    SELECT o.id FROM book_outputs AS o
    WHERE o.book_id = ?1 AND o.output_format = ?2
    ORDER BY o.created_at DESC, o.id DESC
    LIMIT ?3
  );
