# Factory page (customer UI) — build notes, 2026-09-03

Files owned by this work (nothing else touched):

- `srv/static/factory.html` — static shell, readable with JS broken
- `srv/static/factory.js` — all behaviour (~1100 lines, vanilla, no build step)
- `srv/static/factory.css` — layout only; every color/type value is a theme.css token

Head block is copied verbatim from `transmittal.html` (font preload + theme
bootstrap `<script>` + relative `theme.css` / `style.css` hrefs), so the page
works under `/{client}/{project}/factory/` where assets resolve to `static/`.

## Assumptions where the contract was silent

1. **`GET /api/projects/{id}/preflight` needs `?book_id=`.** The contract lists
   the endpoint bare. The handler (`handleGetManuscriptPreflight`) 400s without
   `book_id`, so the page always sends `?book_id=<newest book>`.

2. **Books JSON is PascalCase with a `sql.NullInt64` project id.**
   `handleListProjectBooks` returns `[]GetBooksByProjectRow` and that struct has
   no json tags, so keys are `ID/Title/Author/Series/SourceFilename/Status/
   ErrorMsg/ProjectID/CreatedAt/UpdatedAt` and `ProjectID` is
   `{"Int64":42,"Valid":true}`. `normBook()` accepts that **and** a
   hypothetical snake_case future shape, so adding json tags later won't break
   the page.

3. **Book status vocabulary.** The contract says poll while `'converting'` and
   finish on `'done'`; the pipeline actually writes `uploaded → converting →
   ready | error`. The page treats `converting|building` as in-flight and
   `ready|done` as success, so either spelling works.

4. **"Newest book = current."** Sorted by `CreatedAt` desc with `id` as
   tiebreak. Only the newest is inspected/built; older uploads stay listed
   read-only. The contract doesn't define "current", and one project can hold
   several uploads.

5. **403 means "no pass / expired".** `requirePassAccess` returns 403 (not 402)
   for a missing or expired pass. Undocumented in the contract, so the page
   treats 403 on an action as "my pass data is stale": it re-reads
   `/pass`, re-renders, and shows the entitlement reason. 403 on `/pass`
   itself is treated as `exists:false` (read-only).

6. **Preflight summary shape** read from `preflightResponse` in `srv/preflight.go`:
   `{exists, status, updated_at, summary:{total,high,medium,low,by_type{}},
   images[], report_url, error}`. Severity buckets are relabelled for
   non-technical readers (high → "worth fixing", medium → "worth a look",
   low → "just noting"); `by_type` keys get a plain-English lookup table
   (`TYPE_LABELS`) with a underscore-stripping fallback for unknown keys.

7. **Outputs shape** from `handleListBookOutputs`: snake_case
   `{id, book_id, output_format, size_bytes, created_at}`. Newest of each format
   is the "latest" download; the remainder render as "Earlier builds".

8. **`customer_name` is not the book author.** The page prefills the author
   field from the previous upload first, then `pass.customer_name`; the title
   prefills from the project name (the project was created from the title given
   at redemption) and falls back to the filename.

## One thing the parent should decide

**The build may not produce an EPUB.** The page's button says
"Build PDF + EPUB" per the contract, and the contract's deliverables are
"print PDF + EPUB". But `runConversion` only writes a `pdf` output row — EPUB
comes from a separate `POST /api/books/{id}/convert`-adjacent endpoint,
`POST /api/books/{id}/generate-epub`, which the contract does **not** loosen for
customers (it's still `requireExeDevAdminAPI`).

The page fails soft: if only a PDF output exists it shows
"EPUB — not built yet" plus a line telling the author to reload and then email,
explicitly promising they won't spend another build. But for the Sep 21 workshop
one of these needs to happen on the backend side:

- have `runConversion` also generate the EPUB (one build = both files), **or**
- loosen `generate-epub` to `requirePassAccess` and I'll chain it after the
  build (unmetered), **or**
- change the button/deliverable copy to PDF-only.

Option 1 matches the contract's wording and needs no page change.

## Smoke test

No server needed — see `scratch/factory-smoke/README.md`. `mock.js` fakes the
contract's endpoints; `?state=fresh|nopass|expired|nocredits|building|locked|pdfonly`
walks the edges. With no mock at all every API call 404s, which is the
"server is down" rehearsal: the page still renders all five steps, the footer,
and a visible banner (screenshot `factory-no-api.png`).

Verified: populated, fresh (no uploads), no pass (read-only), expired,
zero credits, mid-build, password gate, PDF-without-EPUB, 390px mobile
(no horizontal overflow), and dark mode.
