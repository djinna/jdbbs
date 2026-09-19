-- 6.1 B, 2026-09-19: per-project interactive style sheets.
--
-- Each project (book) gets its own instance of the house editorial style sheet
-- (table house_style, migration 026). Rows are copied in on first open; the
-- client accepts our defaults (the encouraged path — 'accepted' is the default
-- status), rejects or edits individual rules, and adds their own.
--
-- Status model, deliberately small:
--   accepted — our default, accepted (seeded state)
--   rejected — does not apply to this book; stays visible, struck, omitted from the export
--   edited   — client changed the text; house_id keeps the original reachable for Restore
--   added    — the client's own rule (house_id NULL)

CREATE TABLE IF NOT EXISTS project_stylesheets (
    project_id INTEGER PRIMARY KEY REFERENCES projects(id) ON DELETE CASCADE,
    book_kind  TEXT NOT NULL DEFAULT 'both',   -- both | fiction | nonfiction
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS project_stylesheet_items (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id  INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    house_id    INTEGER,                        -- house_style.id, NULL for client additions
    section_ord INTEGER NOT NULL DEFAULT 0,
    section     TEXT NOT NULL DEFAULT '',
    item_ord    INTEGER NOT NULL DEFAULT 0,
    kind        TEXT NOT NULL DEFAULT 'rule',   -- rule | prose
    col1        TEXT NOT NULL DEFAULT '',
    col2        TEXT NOT NULL DEFAULT '',
    col3        TEXT NOT NULL DEFAULT '',
    body        TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'accepted',
    note        TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_pss_items_project
    ON project_stylesheet_items(project_id, section_ord, item_ord, id);

-- One copy of a given house rule per project (re-seeding a book kind is an
-- INSERT ... WHERE NOT EXISTS; the index makes the invariant explicit).
CREATE UNIQUE INDEX IF NOT EXISTS idx_pss_items_house
    ON project_stylesheet_items(project_id, house_id) WHERE house_id IS NOT NULL;

INSERT OR IGNORE INTO site_pages (route, title, owner, source, visibility, listed, page_type, status, note) VALUES
 ('/{client}/{project}/stylesheet/', 'Project style sheet', 'prodcal', 'srv/static/project-stylesheet.html', 'client', 'nav', 'client', 'live',
  'Per-project instance of the house editorial style sheet (/stylesheet/). Seeded on first open; accept all / reject / edit / add; effective sheet at /{client}/{project}/stylesheet/index.md. API /api/projects/{id}/stylesheet* behind requireAuth. srv/project_stylesheet.go, migration 046.');
