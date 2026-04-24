-- Manuscript transmittal forms (one per project)
CREATE TABLE IF NOT EXISTS transmittals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'draft',
    data TEXT NOT NULL DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_transmittals_project ON transmittals(project_id);

INSERT OR IGNORE INTO migrations (migration_number, migration_name)
VALUES (003, '003-transmittal');
