-- name: GetBookSpec :one
SELECT id, project_id, data, created_at, updated_at
FROM book_specs WHERE project_id = ?;

-- name: UpsertBookSpec :one
INSERT INTO book_specs (project_id, data, updated_at)
VALUES (?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(project_id) DO UPDATE SET data = excluded.data, updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: DeleteBookSpec :exec
DELETE FROM book_specs WHERE project_id = ?;

-- name: GetBookSpecCover :one
SELECT id, project_id, cover_data, cover_type FROM book_specs WHERE project_id = ?;

-- name: UpdateBookSpecCover :exec
UPDATE book_specs SET cover_data = ?, cover_type = ?, updated_at = CURRENT_TIMESTAMP WHERE project_id = ?;
