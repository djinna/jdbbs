-- Migration 014: Preserve EPUB/PDF output history per book

CREATE TABLE IF NOT EXISTS book_outputs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    book_id INTEGER NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    output_format TEXT NOT NULL,
    output_data BLOB NOT NULL,
    source_filename TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_book_outputs_book_format_created
ON book_outputs(book_id, output_format, created_at DESC, id DESC);

INSERT OR IGNORE INTO migrations (migration_number, migration_name)
VALUES (014, '014-book-output-history');
