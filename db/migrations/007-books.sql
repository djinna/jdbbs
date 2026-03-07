-- Migration 007: Books table

-- Books: stores uploaded Word files and generated PDF/EPUB outputs
CREATE TABLE IF NOT EXISTS books (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    author TEXT NOT NULL,
    series TEXT NOT NULL DEFAULT '',
    source_filename TEXT NOT NULL DEFAULT '',
    source_data BLOB,
    pdf_data BLOB,
    epub_data BLOB,
    status TEXT NOT NULL DEFAULT 'uploaded',
    error_msg TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_books_status_created ON books(status, created_at DESC);

INSERT OR IGNORE INTO migrations (migration_number, migration_name)
VALUES (007, '007-books');
