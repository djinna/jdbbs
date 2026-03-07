-- name: ListBooks :many
SELECT id, title, author, series, source_filename, status, error_msg, created_at, updated_at
FROM books ORDER BY created_at DESC;

-- name: GetBook :one
SELECT * FROM books WHERE id = ?;

-- name: CreateBook :one
INSERT INTO books (title, author, series, source_filename, source_data, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, 'uploaded', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING *;

-- name: UpdateBookStatus :exec
UPDATE books SET status = ?, error_msg = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: UpdateBookPDF :exec
UPDATE books SET pdf_data = ?, status = 'ready', updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: UpdateBookEPUB :exec
UPDATE books SET epub_data = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: DeleteBook :exec
DELETE FROM books WHERE id = ?;

-- name: GetBookPDF :one
SELECT id, title, pdf_data FROM books WHERE id = ?;

-- name: GetBookEPUB :one
SELECT id, title, epub_data FROM books WHERE id = ?;
