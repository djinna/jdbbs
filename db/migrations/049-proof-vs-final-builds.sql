-- Migration 049: Proof vs. Final builds (punch list 0.28, 2026-09-20).
-- A build is now either a proof (free, unlimited; the print PDF carries a
-- small PROOF footer on every page and a line on the copyright page) or a
-- final (uses one of the pass's build credits; clean PDF). The EPUB is never
-- marked. books.build_kind is the kind of the in-flight / most recent build,
-- set at request time so a failed proof never refunds a credit that was
-- never debited.
ALTER TABLE book_outputs ADD COLUMN kind TEXT NOT NULL DEFAULT 'final';
ALTER TABLE books ADD COLUMN build_kind TEXT NOT NULL DEFAULT 'final';

CREATE INDEX IF NOT EXISTS idx_book_outputs_kind_created
ON book_outputs(book_id, kind, output_format, created_at DESC, id DESC);

INSERT OR IGNORE INTO migrations (migration_number, migration_name)
VALUES (049, '049-proof-vs-final-builds');
