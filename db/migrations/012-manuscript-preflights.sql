-- Migration 012: Persistent manuscript preflight reports for Typesetting

CREATE TABLE IF NOT EXISTS manuscript_preflights (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    book_id INTEGER NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'ready',
    summary_json TEXT NOT NULL DEFAULT '{}',
    report_json TEXT NOT NULL DEFAULT '[]',
    report_html TEXT NOT NULL DEFAULT '',
    error_msg TEXT NOT NULL DEFAULT '',
    source_filename TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now')),
    UNIQUE(project_id, book_id)
);

CREATE INDEX IF NOT EXISTS idx_manuscript_preflights_project
ON manuscript_preflights(project_id, updated_at DESC);

INSERT OR IGNORE INTO migrations (migration_number, migration_name)
VALUES (012, '012-manuscript-preflights');
