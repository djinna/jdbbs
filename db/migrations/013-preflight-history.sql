-- Migration 013: Keep manuscript preflight history instead of overwriting latest only

CREATE TABLE IF NOT EXISTS manuscript_preflights_new (
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
    updated_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

INSERT INTO manuscript_preflights_new (
    id, project_id, book_id, status, summary_json, report_json, report_html,
    error_msg, source_filename, created_at, updated_at
)
SELECT id, project_id, book_id, status, summary_json, report_json, report_html,
       error_msg, source_filename, created_at, updated_at
FROM manuscript_preflights;

DROP TABLE manuscript_preflights;
ALTER TABLE manuscript_preflights_new RENAME TO manuscript_preflights;

CREATE INDEX IF NOT EXISTS idx_manuscript_preflights_project
ON manuscript_preflights(project_id, updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_manuscript_preflights_project_book_latest
ON manuscript_preflights(project_id, book_id, updated_at DESC, id DESC);

INSERT OR IGNORE INTO migrations (migration_number, migration_name)
VALUES (013, '013-preflight-history');
