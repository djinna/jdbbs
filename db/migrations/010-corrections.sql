-- Migration 010: Corrections table for discrete edit propagation

CREATE TABLE IF NOT EXISTS corrections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    find_text TEXT NOT NULL,
    replace_text TEXT NOT NULL,
    chapter TEXT,
    note TEXT,
    status TEXT NOT NULL DEFAULT 'pending',  -- pending, applied, skipped
    applied_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (project_id) REFERENCES projects(id)
);

CREATE INDEX idx_corrections_project ON corrections(project_id, created_at DESC);

INSERT OR IGNORE INTO migrations (migration_number, migration_name)
VALUES (010, '010-corrections');
