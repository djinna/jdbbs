# EPUB Workflow Test Session — 2026-04-11

## Test Plan
See: `docs/checklists/2026-04-09-epub-workflow-test-checklist.html`

## Results So Far

| Phase | Description | Result |
|-------|-------------|--------|
| 1 | Transmittal Capture (ISBN EPUB, custom styles, persistence) | PASS |
| 2 | Pull Into Spec (transmittal → spec bridge) | PASS |
| 3 | Export Word Template (download .docx, verify styles) | PASS |
| 4 | Prepare Manuscript in Word | PASS |
| 5 | EPUB-First Run | PASS |
| 6 | Print / Typst | FAIL for this manuscript via current upload-triggered convert path; moved into separate workflow review |

## Fixes Applied During Testing

### 1. Page Proofs: "Address / Tel" → "Email" (committed in JS)
- File: `srv/static/transmittal.js` line ~1108
- Changed reviewer contact placeholder from "Address / Tel" to "Email"

### 2. Files & Delivery → Deliverables (committed in JS)
- File: `srv/static/transmittal.js`, `renderFilesSection()`
- Replaced outdated section with modern deliverables:
  - Print-ready PDF (interior)
  - EPUB file
  - Cover files (print + digital)
  - Typst source files
  - Final manuscript (Word/DOCX)
  - Fonts used (if licensable)
- Printer Delivery simplified to: PDF/X or Other (with text field)
- Data key changed: `files.archives` → `files.deliverables`
- Old archive values on existing transmittals won't carry over (intentional)

## EPUB Findings From Continued Testing

### Workflow clarification
- Desired workflow is now explicit:
  - Files page: upload only
  - Typesetting page: explicit Generate EPUB action
  - Typst/PDF generation: separate workflow review, not auto-triggered on upload
- Current app behavior does not match that intent yet: Files upload still auto-triggers `/api/books/{id}/convert`

### Phase 4 result
- User completed the Word manuscript preparation step manually.
- Test manuscript used custom styles:
  - `tweet-p`
  - `metadata-p`
  - `metadata-c`
- User also left deliberate local-formatting noise for edge-case review.

### Phase 5 result
- EPUB build: PASS
- Generated artifact reviewed from local docs directory:
  - `docs/The-Twitter-Years-200722-template.epub`
- Structural preservation of custom styles: PASS
  - EPUB XHTML contains:
    - `<div data-custom-style="tweet-p">…</div>`
    - `<div data-custom-style="metadata-p">…</div>`
    - `<span data-custom-style="metadata-c">…</span>`
- Conclusion: custom style identity survives the DOCX → EPUB path.

### Styling review outcome
- By default, EPUB output preserved the wrappers but applied almost no visual styling.
- Tested custom CSS that produced an acceptable first-pass rendering:

```css
div[data-custom-style="tweet-p"] p {
  font-family: sans-serif;
  font-size: 0.9em;
  margin-top: 0.4em;
  margin-bottom: 0em;  
  text-align: left;
}

div[data-custom-style="metadata-p"] p {
  font-family: sans-serif;
  font-size: 0.6em;
  margin-top: 0em;
  margin-bottom: 0.4em;
  text-align: left;
  color: #666;
}

span[data-custom-style="metadata-c"] {
  font-family: sans-serif;
  font-size: 0.7em;
  color: #666;
}
```

- User feedback after a couple of test cycles: this is a good stopping point for the EPUB custom-style CSS review.

### Stray local formatting review
- Not yet confirmed as caught by a clear review/report path.
- EPUB artifact alone did not demonstrate a surfaced warning/report for deliberate local formatting noise.
- Treat this as still needing explicit follow-up in the edge-case review work.

### Print / Typst finding
- Current upload-triggered convert path failed on this manuscript during Typst compilation.
- Failure mode: generated `chapter.typ` still contained markdown-like heading/footnote syntax that Typst rejected.
- Since Typst should not be auto-triggered from upload anyway, this was deferred into a separate workflow review rather than treated as a blocker for EPUB-first validation.

## Queued Fixes / Follow-ups

### 1. Disable auto-processing on Files upload
- Uploading a DOCX should create/link the book record only.
- Do not auto-trigger Typst/PDF conversion on upload.
- EPUB generation should remain an explicit action from the Typesetting page.
- Typst/PDF flow should be reviewed separately after this workflow change.

### 2. Review special manuscript edge cases after upload workflow fix
- Return to edge-case handling after upload behavior is corrected.
- Known next item mentioned during testing: some ASCII art is currently broken.

### 3. Auto-update project_slug when book.title changes
- When user edits book title on transmittal, the project slug should update.
- Implementation: in `setField()` in transmittal.js, when `path === 'book.title'`, slugify the value and call `PUT /api/projects/{id}` with new `project_slug`
- Backend already supports this endpoint.
- Not a blocker — cosmetic/admin-side only.

## Deploy Notes
- Binary is embedded-static, so every JS change requires rebuild + restart of the actual serving unit.
- Current host reality during this session: `prodcal.service` was the winning listener on port 8000.
- Duplicate unit hazard still exists: `srv.service` and `prodcal.service` can race.
- Verify live binary with `/proc/$PID/exe` before assuming a deploy is active.

## Next Steps
1. Disable auto-processing on Files upload so upload is upload-only.
2. Keep EPUB generation explicit from Typesetting.
3. Review Typst/PDF as a separate workflow, not as part of upload.
4. Return to special manuscript edge cases, including broken ASCII art.
5. Batch-fix the slug auto-update later.
