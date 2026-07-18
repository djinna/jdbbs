-- Migration 019: PI editorial stylesheet review tool.
-- Items seeded from the Zoothesia stylesheet; DB is canonical going forward.
-- Team members (allowlisted by email) can accept/reject/edit items and tag
-- some as "author-facing" to derive a tighter authors' sheet.

CREATE TABLE IF NOT EXISTS stylesheet_items (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    section_ord   INTEGER NOT NULL DEFAULT 0,
    section       TEXT    NOT NULL DEFAULT '',
    item_ord      INTEGER NOT NULL DEFAULT 0,
    kind          TEXT    NOT NULL DEFAULT 'prose',   -- 'rule' | 'prose'
    col1          TEXT    NOT NULL DEFAULT '',
    col2          TEXT    NOT NULL DEFAULT '',
    col3          TEXT    NOT NULL DEFAULT '',
    body          TEXT    NOT NULL DEFAULT '',
    status        TEXT    NOT NULL DEFAULT 'proposed', -- proposed|accepted|rejected
    author_facing INTEGER NOT NULL DEFAULT 0,
    status_by     TEXT    NOT NULL DEFAULT '',
    status_at     TIMESTAMP,
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_stylesheet_items_order
    ON stylesheet_items (section_ord, item_ord);

CREATE TABLE IF NOT EXISTS stylesheet_edits (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    item_id      INTEGER NOT NULL REFERENCES stylesheet_items(id) ON DELETE CASCADE,
    col1         TEXT NOT NULL DEFAULT '',
    col2         TEXT NOT NULL DEFAULT '',
    col3         TEXT NOT NULL DEFAULT '',
    body         TEXT NOT NULL DEFAULT '',
    note         TEXT NOT NULL DEFAULT '',
    proposed_by  TEXT NOT NULL DEFAULT '',
    proposed_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    state        TEXT NOT NULL DEFAULT 'pending'  -- pending|accepted|rejected
);

CREATE INDEX IF NOT EXISTS idx_stylesheet_edits_item
    ON stylesheet_edits (item_id, state);

INSERT OR IGNORE INTO migrations (migration_number, migration_name) VALUES (019, '019-stylesheet');
