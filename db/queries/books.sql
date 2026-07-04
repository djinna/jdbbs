-- name: ListBooks :many
SELECT id, title, author, series, source_filename, status, error_msg, project_id, created_at, updated_at
FROM books ORDER BY created_at DESC;

-- name: GetBook :one
SELECT * FROM books WHERE id = ?;

-- name: CreateBook :one
INSERT INTO books (title, author, series, source_filename, source_data, project_id, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, 'uploaded', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING *;

-- name: UpdateBookStatus :exec
UPDATE books SET status = ?, error_msg = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: UpdateBookStatusReady :exec
UPDATE books SET status = 'ready', updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: TouchBook :exec
UPDATE books SET updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: UpdateBookProject :exec
UPDATE books SET project_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: DeleteBook :exec
DELETE FROM books WHERE id = ?;

-- name: GetBookPDF :one
SELECT b.id, b.title, o.output_data AS pdf_data
FROM books b
LEFT JOIN book_outputs o ON o.book_id = b.id AND o.output_format = 'pdf'
WHERE b.id = ?
ORDER BY o.created_at DESC, o.id DESC
LIMIT 1;

-- name: GetBookEPUB :one
SELECT b.id, b.title, o.output_data AS epub_data
FROM books b
LEFT JOIN book_outputs o ON o.book_id = b.id AND o.output_format = 'epub'
WHERE b.id = ?
ORDER BY o.created_at DESC, o.id DESC
LIMIT 1;

-- name: GetBooksByProject :many
SELECT id, title, author, series, source_filename, status, error_msg, project_id, created_at, updated_at
FROM books WHERE project_id = ? ORDER BY created_at DESC;

-- name: GetBookProjectID :one
SELECT id, project_id FROM books WHERE id = ?;
