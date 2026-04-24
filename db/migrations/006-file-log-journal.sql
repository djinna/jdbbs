-- Migration 006: File Log + Project Journal tables

-- File Log: structured record of file transfers (Dropbox handoffs)
CREATE TABLE IF NOT EXISTS file_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    direction TEXT NOT NULL DEFAULT 'inbound',  -- inbound / outbound
    filename TEXT NOT NULL DEFAULT '',
    file_type TEXT NOT NULL DEFAULT '',          -- .docx, .pdf, .epub, .tiff, .jpg, .png, .eps
    sent_by TEXT NOT NULL DEFAULT '',
    received_by TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    transfer_date TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_file_log_project ON file_log(project_id, transfer_date DESC);

-- Project Journal: timestamped notes for calls, decisions, approvals
CREATE TABLE IF NOT EXISTS journal (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    entry_type TEXT NOT NULL DEFAULT 'note',  -- call / decision / approval / note
    content TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_journal_project ON journal(project_id, created_at DESC);

INSERT OR IGNORE INTO migrations (migration_number, migration_name)
VALUES (006, '006-file-log-journal');
