-- Migration 009: Link books to projects, cover image storage

ALTER TABLE books ADD COLUMN project_id INTEGER REFERENCES projects(id) ON DELETE SET NULL;

ALTER TABLE book_specs ADD COLUMN cover_data BLOB;
ALTER TABLE book_specs ADD COLUMN cover_type TEXT NOT NULL DEFAULT '';

INSERT OR IGNORE INTO migrations (migration_number, migration_name)
VALUES (009, '009-book-project-link');
