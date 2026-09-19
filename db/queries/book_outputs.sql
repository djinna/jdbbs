-- name: CreateBookOutput :one
INSERT INTO book_outputs (
    book_id, output_format, output_data, source_filename, spec_snapshot, corrections_snapshot, kind
) VALUES (
    ?, ?, ?, ?, ?, ?, ?
)
RETURNING id, book_id, output_format, output_data, source_filename, spec_snapshot, corrections_snapshot, created_at, kind;

-- name: ListBookOutputs :many
SELECT id, book_id, output_format, source_filename, length(output_data) AS size_bytes, spec_snapshot, corrections_snapshot, created_at, kind
FROM book_outputs
WHERE book_id = ?
ORDER BY created_at DESC, id DESC
LIMIT ?;

-- name: GetBookOutput :one
SELECT id, book_id, output_format, output_data, source_filename, spec_snapshot, corrections_snapshot, created_at, kind
FROM book_outputs
WHERE id = ? AND book_id = ?;

-- name: GetLatestBookOutputByKind :one
-- Newest artifact of one format and kind (proof | final) for a book; backs
-- GET /api/books/{id}/download/{format}?kind=final.
SELECT id, book_id, output_format, output_data, source_filename, created_at, kind
FROM book_outputs
WHERE book_id = ? AND output_format = ? AND kind = ?
ORDER BY created_at DESC, id DESC
LIMIT 1;

-- name: CountProofOutputsByProjectSince :one
-- Proof PDFs built for a project's books since a moment: the per-project
-- proof rate limit (proofs are free, so this is the only brake).
SELECT COUNT(*) FROM book_outputs o
JOIN books b ON b.id = o.book_id
WHERE b.project_id = ? AND o.kind = 'proof' AND o.output_format = 'pdf' AND o.created_at >= ?;

-- name: PruneBookOutputsByKind :exec
-- Keep the newest `keep` outputs of one format+kind for a book; delete the
-- rest. Used for proof PDFs, which are unlimited and would otherwise pile up.
DELETE FROM book_outputs
WHERE book_outputs.book_id = ?1 AND book_outputs.output_format = ?2 AND book_outputs.kind = ?3
  AND book_outputs.id NOT IN (
    SELECT o.id FROM book_outputs AS o
    WHERE o.book_id = ?1 AND o.output_format = ?2 AND o.kind = ?3
    ORDER BY o.created_at DESC, o.id DESC
    LIMIT ?4
  );

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
