-- Migration 018: Store compiled artifacts only in book_outputs (014). The
-- latest PDF/EPUB used to be duplicated on books.pdf_data/epub_data, doubling
-- every backup. Backfill any blob that has no history row (pre-014 compiles),
-- then drop the columns.

INSERT INTO book_outputs (book_id, output_format, output_data, source_filename, created_at)
SELECT b.id, 'pdf', b.pdf_data, b.source_filename, b.updated_at
FROM books b
WHERE b.pdf_data IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM book_outputs o WHERE o.book_id = b.id AND o.output_format = 'pdf'
  );

INSERT INTO book_outputs (book_id, output_format, output_data, source_filename, created_at)
SELECT b.id, 'epub', b.epub_data, b.source_filename, b.updated_at
FROM books b
WHERE b.epub_data IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM book_outputs o WHERE o.book_id = b.id AND o.output_format = 'epub'
  );

ALTER TABLE books DROP COLUMN pdf_data;
ALTER TABLE books DROP COLUMN epub_data;

INSERT OR IGNORE INTO migrations (migration_number, migration_name)
VALUES (018, '018-book-blob-dedupe');
