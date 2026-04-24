-- Version history for transmittal forms
CREATE TABLE IF NOT EXISTS transmittal_versions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    transmittal_id INTEGER NOT NULL REFERENCES transmittals(id) ON DELETE CASCADE,
    data TEXT NOT NULL DEFAULT '{}',
    status TEXT NOT NULL DEFAULT 'draft',
    saved_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_transmittal_versions_tid ON transmittal_versions(transmittal_id);

INSERT OR IGNORE INTO migrations (migration_number, migration_name)
VALUES (004, '004-transmittal-versions');
