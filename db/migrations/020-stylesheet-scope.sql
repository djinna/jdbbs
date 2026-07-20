-- Migration 020: per-item scope for the editorial stylesheet.
-- `scope` distinguishes rules that belong in the org-wide PI master stylesheet
-- (universal — carried forward to future titles) from rules specific to the
-- current title (zoothesia — kept with that book's archive and retired after).
-- Default 'universal' so existing seeded rows are treated as candidates for the
-- master sheet until an editor narrows them to a specific title.

ALTER TABLE stylesheet_items
    ADD COLUMN scope TEXT NOT NULL DEFAULT 'universal';  -- 'universal' | 'zoothesia'

CREATE INDEX IF NOT EXISTS idx_stylesheet_items_scope
    ON stylesheet_items (scope);

INSERT OR IGNORE INTO migrations (migration_number, migration_name) VALUES (020, '020-stylesheet-scope');
