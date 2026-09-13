-- Content review queue: proposed copy changes from a better-documents review,
-- decided by the admin at /admin/content-review/ before anything is edited.
-- The page proposes; it never writes to the reviewed files.
CREATE TABLE IF NOT EXISTS content_review_items (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    batch        TEXT NOT NULL,                       -- '2026-09-13'
    ord          INTEGER NOT NULL DEFAULT 0,
    ref          TEXT NOT NULL DEFAULT '',            -- 'A-F2' (id in the full report)
    page         TEXT NOT NULL DEFAULT '',            -- route or artifact name
    file         TEXT NOT NULL DEFAULT '',            -- where the change would land
    location     TEXT NOT NULL DEFAULT '',
    current_text TEXT NOT NULL DEFAULT '',
    proposed     TEXT NOT NULL DEFAULT '',
    reason       TEXT NOT NULL DEFAULT '',
    severity     TEXT NOT NULL DEFAULT 'MINOR',       -- CRITICAL | MAJOR | MINOR | NIT
    pass         INTEGER NOT NULL DEFAULT 1,          -- 1..5 (skill pass)
    decision     TEXT NOT NULL DEFAULT 'pending',     -- pending | accepted | edited | rejected
    edited_text  TEXT NOT NULL DEFAULT '',
    note         TEXT NOT NULL DEFAULT '',
    decided_at   TIMESTAMP,
    applied_at   TIMESTAMP,
    created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_content_review_batch ON content_review_items(batch, ord);

INSERT OR IGNORE INTO site_pages (route, title, owner, source, visibility, listed, page_type, status, note) VALUES
 ('/admin/content-review/', 'Content review', 'prodcal', 'srv/static/content-review.html', 'admin', 'unlisted', 'admin', 'live', 'Accept / edit / reject proposed copy changes from docs/reviews/CONTENT-REVIEW-*.md'),
 ('/admin/email-preview/',  'Email template gallery', 'prodcal', 'srv/email_preview.go', 'admin', 'unlisted', 'admin', 'live', 'All outbound templates rendered with fixture data');
