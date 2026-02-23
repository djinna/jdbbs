-- Add client/project slug fields for URL paths like /vgr/aog/
ALTER TABLE projects ADD COLUMN client_slug TEXT NOT NULL DEFAULT '';
ALTER TABLE projects ADD COLUMN project_slug TEXT NOT NULL DEFAULT '';

CREATE UNIQUE INDEX IF NOT EXISTS idx_projects_path ON projects(client_slug, project_slug);

INSERT OR IGNORE INTO migrations (migration_number, migration_name)
VALUES (002, '002-slugs');
