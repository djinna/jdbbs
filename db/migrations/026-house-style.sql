-- Migration 026: public house stylesheet (the /stylesheet split, 2026-09-11).
--
-- stylesheet_items stays the internal PI review tool (now at /stylesheet-pi/).
-- house_style is the anonymized, read-only, publisher-wide sheet served at
-- /stylesheet/. Seeded from srv/house_style_seed.json on first request.
CREATE TABLE IF NOT EXISTS house_style (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    section_ord INTEGER NOT NULL,
    section     TEXT NOT NULL,
    item_ord    INTEGER NOT NULL,
    kind        TEXT NOT NULL DEFAULT 'rule',      -- rule | prose
    col1        TEXT NOT NULL DEFAULT '',
    col2        TEXT NOT NULL DEFAULT '',
    col3        TEXT NOT NULL DEFAULT '',
    body        TEXT NOT NULL DEFAULT '',
    book_kind   TEXT NOT NULL DEFAULT 'both',      -- fiction | nonfiction | both
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_house_style_order ON house_style (section_ord, item_ord);

UPDATE site_pages SET route='/stylesheet-pi/', title='PI editorial stylesheet (review tool)', source='srv/static/stylesheet-pi/index.html',
  visibility='public', listed='unlisted', status='keep',
  note='internal review tool; write access allowlisted by exe.dev email. Was /stylesheet/ before 2026-09-11', updated_at=CURRENT_TIMESTAMP
WHERE route='/stylesheet/';
UPDATE site_pages SET route='/stylesheet-pi/authors', source='srv/static/stylesheet-pi/authors.html', listed='unlisted', status='keep', updated_at=CURRENT_TIMESTAMP
WHERE route='/stylesheet/authors';
INSERT OR IGNORE INTO site_pages (route, title, owner, source, visibility, listed, page_type, status, note) VALUES
 ('/stylesheet/', 'House editorial stylesheet', 'prodcal', 'srv/housestyle.go + srv/house_style_seed.json', 'public', 'listed', 'ledger', 'live',
  'anonymized, read-only, fiction/nonfiction/both tabs; linked from home resources');
