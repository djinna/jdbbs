-- Content review 2026-09-13, E-F1 / E-F2.
-- Reorder the house stylesheet so the intro can cite sections in reading order:
-- author-facing mechanics (basics, typography) first, then what the studio does
-- with the file on arrival, then the fiction/nonfiction sections.
-- Old: 1 basics, 2 manuscript handling, 3 voice, 4 dialogue, 5 typography, 6 nonfiction, 7 fixes, 8 don't-change
-- New: 1 basics, 2 typography, 3 manuscript handling, 4 voice, 5 dialogue, 6 nonfiction, 7 fixes, 8 don't-change
UPDATE house_style SET section_ord = CASE section_ord
    WHEN 2 THEN 3
    WHEN 3 THEN 4
    WHEN 4 THEN 5
    WHEN 5 THEN 2
    ELSE section_ord END
  WHERE section_ord IN (2,3,4,5);

-- Manuscript handling: address the author, not the copyeditor.
UPDATE house_style SET body = 'Send manuscripts as **.docx**, not plain text or pasted text. Word files preserve italics, smart quotes, em dashes, and scene breaks; plain-text paste loses them.'
  WHERE section = 'Manuscript handling' AND item_ord = 1 AND kind = 'prose';
UPDATE house_style SET body = 'On our side: we edit in a lossless round-trip (.docx → markdown → .docx or equivalent). Asterisks in the extracted text are the author''s italics; we verify each one is real emphasis before touching it.'
  WHERE section = 'Manuscript handling' AND item_ord = 2 AND kind = 'prose';
UPDATE house_style SET body = 'You get back a **clean edited file plus a changelog** (before → after list and author queries). Review the changelog, not tracked changes — they don''t reliably survive conversion.'
  WHERE section = 'Manuscript handling' AND item_ord = 4 AND kind = 'prose';
