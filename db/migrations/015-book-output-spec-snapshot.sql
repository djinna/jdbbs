-- Migration 015: Capture the book_spec JSON in effect at compile time on each
-- book_outputs row, so a historical artifact is self-documenting.

ALTER TABLE book_outputs ADD COLUMN spec_snapshot TEXT;

INSERT OR IGNORE INTO migrations (migration_number, migration_name)
VALUES (015, '015-book-output-spec-snapshot');
