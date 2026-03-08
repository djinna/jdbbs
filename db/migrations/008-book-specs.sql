-- Migration 008: Book specs table for typesetting configuration

CREATE TABLE IF NOT EXISTS book_specs (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    data       TEXT    NOT NULL DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id)
);

INSERT OR IGNORE INTO migrations (migration_number, migration_name)
VALUES (008, '008-book-specs');
