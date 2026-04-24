-- Base schema
CREATE TABLE IF NOT EXISTS migrations (
    migration_number INTEGER PRIMARY KEY,
    migration_name TEXT NOT NULL,
    executed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Projects
CREATE TABLE IF NOT EXISTS projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    start_date TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Tasks
CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    assignee TEXT NOT NULL DEFAULT '',
    title TEXT NOT NULL DEFAULT '',
    is_milestone INTEGER NOT NULL DEFAULT 0,
    orig_weeks REAL NOT NULL DEFAULT 0,
    curr_weeks REAL NOT NULL DEFAULT 0,
    orig_due TEXT NOT NULL DEFAULT '',
    curr_due TEXT NOT NULL DEFAULT '',
    actual_done TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending',  -- pending, in_progress, done
    words INTEGER NOT NULL DEFAULT 0,
    words_per_hour INTEGER NOT NULL DEFAULT 0,
    hours REAL NOT NULL DEFAULT 0,
    rate REAL NOT NULL DEFAULT 0,
    budget_notes TEXT NOT NULL DEFAULT '',
    orig_budget REAL NOT NULL DEFAULT 0,
    curr_budget REAL NOT NULL DEFAULT 0,
    actual_budget REAL NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tasks_project ON tasks(project_id, sort_order);

-- Auth tokens (simple shared passwords per project)
CREATE TABLE IF NOT EXISTS auth_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    label TEXT NOT NULL DEFAULT 'default',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT OR IGNORE INTO migrations (migration_number, migration_name)
VALUES (001, '001-base');
