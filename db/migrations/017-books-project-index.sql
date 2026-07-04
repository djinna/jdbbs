-- Migration 017: Index books by project so the per-project book listing
-- (GetBooksByProject) doesn't scan the whole table.

CREATE INDEX IF NOT EXISTS idx_books_project_id ON books(project_id);

INSERT OR IGNORE INTO migrations (migration_number, migration_name)
VALUES (017, '017-books-project-index');
