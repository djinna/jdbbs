-- Migration 011: archive projects instead of deleting them
ALTER TABLE projects ADD COLUMN archived_at TIMESTAMP;

CREATE INDEX IF NOT EXISTS idx_projects_archived_at ON projects(archived_at, updated_at DESC);

INSERT OR IGNORE INTO migrations (migration_number, migration_name)
VALUES (11, 'project archive');
