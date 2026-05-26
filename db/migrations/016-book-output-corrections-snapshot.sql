-- Migration 016: Capture the corrections ledger (YAML) in effect at compile
-- time on each book_outputs row. Pairs with spec_snapshot (015): together they
-- describe everything an artifact was built from.

ALTER TABLE book_outputs ADD COLUMN corrections_snapshot TEXT;

INSERT OR IGNORE INTO migrations (migration_number, migration_name)
VALUES (016, '016-book-output-corrections-snapshot');
