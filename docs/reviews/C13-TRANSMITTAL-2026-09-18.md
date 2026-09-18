# C13 — Transmittal rewritten for the factory (punch-list 5.4)

2026-09-18. Sub-items C6, C7, C10, C11, C12 of
`docs/runs/RUN-2026-09-17-protocol-institute.md` (lines 87–116), one commit each.
The rest of C13 (Book Design trim picker, Page Proofs, Deliverables, Subrights)
is **not** in this block — see TODOs.

Rule kept throughout: a saved transmittal is a JSON blob; **every old key still
loads and is preserved on save**, even where its field is no longer shown.
`setField`/`getField` only touch the path being edited, and the server stores
the blob as sent, so nothing a client typed is lost.

## What changed, per sub-item

### a) C6 — Production section gone (`cd1af7f`)
- Removed the section: Mechs Delivery, Weeks in Prod., Bound Book Date, and the
  second Transmittal Date.
- **Print Run** now sits in Book Information (still key `production.print_run`).
- New optional **Target date** (`production.target_date`), also in the studio
  email summary (`srv/email.go`) and reset on Duplicate.
- Legacy, read-only-in-JSON: `production.transmittal_date`, `production.mechs_delivery`,
  `production.weeks_in_production`, `production.bound_book_date`.

### b) C7 — stats from the manuscript, not typed (`db9bcbb`)
- **Parts** stays a typed field (`checklist_stats.parts`): a number ≥ 1 = the
  book has parts; blank / "none" = no parts. `pullTransmittalIntoSpec` now writes
  it to `spec.structure.parts` — before this the count never left the transmittal,
  so `specHasParts` only ever saw hand-edited specs.
- **Chapters / Words / Images** are read-only, filled from the newest upload's
  Inspect book map (`GET /api/projects/{id}/books` → newest → `GET …/preflight?book_id=`).
  `BookMap` gained `words`, `images`, per-section `words`, and `counted` (false
  on reports stored before this change → the row says "Inspect again for a word
  count" instead of showing 0). No file → "counted at upload" with a link to the
  factory page.
- Legacy, read-only-in-JSON: `checklist_stats.chapters`, `words_chars`, `ms_pp`, `est_book_pp`.
- Interim CSS underline fix from `3233362` kept; new `.tx-stats-grid`, `.tx-readonly-value`.

### c) C10 — Editing → Typography notes (`a1f60d6`)
- Removed: Developmental Edit (select), Instructions for Developmental Editor,
  Level of Copyediting, Instructions for Copyeditor.
- Kept with new help text: Special Characters, Mathematical Formulas, Custom Styles.
- Legacy, read-only-in-JSON: `editing.developmental_edit`, `editing.developmental_instructions`,
  `editing.copyediting_level`, `editing.instructions`. (`pullTransmittalIntoSpec`
  still copies the two instruction texts into `spec.typesetting.*` if present.)

### d) C11 — Permissions → Rights + Terms page (`1d377a7`, jdbbs-public `8bc45da`)
- Removed: Permissions (reprint) status + date, Consents status + date.
- New: one courtesy paragraph and **one attestation checkbox** — "Everything in
  this manuscript is mine or I have permission to reprint it. The factory
  typesets what I send; clearing rights is my responsibility." Stored as
  `permissions.attested` (bool) + `permissions.attested_at` (ISO timestamp);
  reset on Duplicate.
- New public page **`/factory/terms`** served by `servePublicDoc` from
  `/home/exedev/jdbbs-public/factory-terms.html` (same masthead/footer/theme as
  `/factory`); `site_pages` row in `db/migrations/043-factory-terms-page.sql`.
  Ten plain-English points: delivered, builds count, failed builds not counted,
  rights are the author's, attestation, no editing, data kept six months
  (+30 days read-only, then purged), cover & fonts, refunds, changes.
  Linked from the attestation line and from `/factory` → Storage & privacy.
- Legacy, read-only-in-JSON: `permissions.reprint_status`, `reprint_when`,
  `consents_status`, `consents_when`.

### e) C12 — Pub Info & © → Copyright page builder (`a2b4068`)
Fields (key → what it is):
- `page_iv.copyright_year`, `page_iv.held_by` (rights holder; blank = author)
- publisher and ISBNs are **shown read-only from Book Information** (`book.publisher`,
  `book.isbn_paper`, `book.isbn_epub` — reused, not duplicated)
- `page_iv.publisher_city`, `page_iv.edition_line`
- `cover.credit` — **moved** from the Cover section into this one (same key)
- `page_iv.interior_credit` — blank prints "Typeset by jdbb studio in {typeface}";
  `{typeface}` is filled from `spec.typography.body_font` at build time
  (`interiorCredit()` in `srv/books.go`; `copyright_page_text()` in
  `generate-word-template.py`)
- `page_iv.loc_line`, `page_iv.printed_in`, `page_iv.additional_notices` (textarea)
- **Live preview** of the assembled page under the fields (`refreshCopyrightPreview`
  hooked into `setField` for `page_iv.*`, `book.*`, `cover.*`).

Plumbing:
- `pullTransmittalIntoSpec` maps the new keys to `spec.metadata.*` and (new)
  `tx.cover.credit → spec.cover.credit` — `frontMatterTypst` was already reading
  `spec.cover.credit`, but nothing ever wrote it.
- `frontMatterTypst` passes `publisher-city, edition-line, interior-credit,
  loc-line, printed-in, notices`; `copyright-page-generated` in
  `series-template.typ` sets them in trade order (title · © · publisher, city ·
  edition · cover credit · interior credit · legacy credit lines · LoC · ISBNs ·
  notices · printed in). `escapeTypstString` now escapes newlines (notices are
  multi-line).
- `generate-word-template.py`'s Copyright paragraph is built by the same rules
  (`copyright_page_text`), so the Word template shows what the build will print.
- Tests: `TestFrontMatterTypstCompiles` extended with the new fields (typst
  compile + pdftotext); new `TestFrontMatterTypstCopyrightBuilder`.
- Legacy: `page_iv.credit` is still mapped to `metadata.credit_lines` and printed
  if present; `page_iv.other_credit` / `page_iv.photo_credit` are **not printed**
  (they never were) but are shown read-only in the section when non-empty, with
  a note to move the text into Additional notices.

## Removed fields — plain list
Production: Mechs Delivery, Weeks in Production, Bound Book Date, duplicate
Transmittal Date · Checklist stats: Chapters, Words/Chars, MS pp, Est. Book pp
(now counted) · Editing: Developmental Edit, Instructions for Developmental
Editor, Level of Copyediting, Instructions for Copyeditor · Permissions:
reprint status/date, consents status/date · Pub Info: Credit Line, Other Credit,
Photo Credit (as inputs; text still visible if present) · Cover: Cover credit
(moved to Copyright page, not removed).

## Tested
- Browser (admin proxy :8799 and plain :8000) on `/prot/zoo/transmittal/`: every
  section renders, edits autosave, reload shows them, legacy keys survive a save
  (checked the JSON via `GET /api/projects/14/transmittal`).
- `POST /api/projects/14/book-spec/pull-transmittal` → `structure.parts`,
  `cover.credit`, new `metadata.*` present.
- `GET /api/projects/14/word-template` → 200 with the new Copyright block
  (needed the transmittal final and the zoo's duplicate "Code" custom style —
  pre-existing data — renamed for the test; both restored to draft/original).
- `go test ./srv/` — all pass except the pre-existing `TestNavConvergence`
  failure on the untracked `jdbbs-public/2026-pi-symposium/factory-talk.html`
  (another session's file; not touched).
- `/factory/terms` 200, `/factory` links to it.

## TODOs / not done here
- **EPUB has no generated copyright page.** The EPUB filter drops a typed
  Copyright block (P4) but nothing generates one — the print PDF gets p. iv, the
  EPUB gets nothing. Add a generated `copyright.xhtml` (epub:type copyright-page)
  from the same fields; ~1–2 h in `srv/epub.go`.
- Terms page: **a lawyer has not read it** (the page says so). Refund wording
  (point 9: unused pass refundable, used pass not) is a first draft for Jenna to
  confirm. After Sep 23, set `terms_of_service_url` in Stripe and
  `consent_collection.terms_of_service` at Checkout (C11 note).
- Cover credit prints as typed: "James Langdon" prints bare. Either auto-prefix
  "Cover design by" when the text has no verb, or rely on the placeholder.
- Word count is `strings.Fields` over the DOCX paragraphs — fine for "about how
  long"; not Word's exact number.
- `calcCompletion` (progress bar) still counts Book fields, checklist rows and
  `design.trim` / `design.complexity`; adjust when Book Design (C13 remainder) lands.
- The studio email (`srv/email.go`) still prints legacy Production / stats /
  editing values when present; harmless, tidy with the C13 remainder.
- C13 remainder untouched: Book Design trim picker + reuse dropdown, delete
  Page Proofs / Deliverables / Subrights, front-matter menu, design guidance.

## Addendum, same day — Format section (was Book Design)

Jenna's inbox note: the trim radios still said "DON'T CARE", and PPI / spine
width assumed we know the paper. We don't; the printer does.

- Section renamed **Format**. Lead line says the one physical decision that is
  the author's is the page size ("which printers call the trim").
- "What kind of book is it, as an object?" (`design.trim_guidance`) moved to
  the top, with civilian placeholders.
- Trim choice is now a stacked list with one plain sentence each:
  **Small** 5½ × 8½ · **Medium** 6 × 9 ("if you are unsure, choose this") ·
  **Large** 8½ × 11 · **Let the studio choose** (value `studio`; legacy
  `dont_care` selects this) · **Exact size** with a W × H text box.
- `pullTransmittalIntoSpec`: `studio`/`dont_care` no longer overwrite
  `page.trim` (previously wrote the literal string "dont_care" into the spec);
  `parseTrim` now understands free-form `W x H` inches (`7 x 10`, `6.14 × 9.21 in`),
  so "Exact size" really sets the page. `7 x 10` added to the registry.
- Dropped inputs: Est. Book pp, PPI, Spine Width, Text Complexity (jdbb tier),
  Outside Designer, Reuse Previous Book. Replaced by one paragraph: spine width
  depends on the printer's paper (PPI); give the printer trim + page count when
  the PDF is final, they return a cover template; bring that back if you want
  help with the cover. Old keys still load/save; the new-transmittal default
  drops them. Progress bar counts `design.trim` only.
- Tests: `TestParseTrimFreeForm`, `TestStudioTrimLeavesPageAlone`.
