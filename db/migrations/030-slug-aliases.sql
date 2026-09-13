-- Old client/project URL paths that should keep working after a rename.
-- old_project_slug = '' means a client-level alias (the whole /{client}/ tree).
CREATE TABLE IF NOT EXISTS slug_aliases (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    old_client_slug  TEXT NOT NULL,
    old_project_slug TEXT NOT NULL DEFAULT '',
    new_client_slug  TEXT NOT NULL,
    new_project_slug TEXT NOT NULL DEFAULT '',
    created_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (old_client_slug, old_project_slug)
);
