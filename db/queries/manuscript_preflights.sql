-- name: GetLatestManuscriptPreflight :one
SELECT id, project_id, book_id, status, summary_json, report_json, report_html, error_msg, source_filename, created_at, updated_at
FROM manuscript_preflights
WHERE project_id = ? AND book_id = ?
ORDER BY updated_at DESC, id DESC
LIMIT 1;

-- name: CreateManuscriptPreflight :one
INSERT INTO manuscript_preflights (
    project_id, book_id, status, summary_json, report_json, report_html,
    error_msg, source_filename, updated_at
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
RETURNING id, project_id, book_id, status, summary_json, report_json, report_html, error_msg, source_filename, created_at, updated_at;

-- name: ListManuscriptPreflights :many
SELECT id, project_id, book_id, status, summary_json, report_json, report_html, error_msg, source_filename, created_at, updated_at
FROM manuscript_preflights
WHERE project_id = ? AND book_id = ?
ORDER BY updated_at DESC, id DESC;
