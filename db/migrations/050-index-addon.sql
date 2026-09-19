-- Migration 050: Back-of-book index add-on (punch list 5.13, 2026-09-20).
-- books.index_json is the drafted/reviewed index document (srv/indexer
-- Index schema); index_status is its state machine:
--   off → drafting → draft → reviewed   (error on a failed draft)
-- passes.index_included is the $100 add-on entitlement (store lookup_key
-- "index", or an admin grant).
ALTER TABLE books ADD COLUMN index_json TEXT;
ALTER TABLE books ADD COLUMN index_status TEXT NOT NULL DEFAULT 'off';
ALTER TABLE passes ADD COLUMN index_included INTEGER NOT NULL DEFAULT 0;

INSERT OR IGNORE INTO migrations (migration_number, migration_name)
VALUES (050, '050-index-addon');
