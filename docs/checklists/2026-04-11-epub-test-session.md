# EPUB Workflow Test Session — 2026-04-11

## Test Plan
See: `docs/checklists/2026-04-09-epub-workflow-test-checklist.html`

## Results So Far

| Phase | Description | Result |
|-------|-------------|--------|
| 1 | Transmittal Capture (ISBN EPUB, custom styles, persistence) | PASS |
| 2 | Pull Into Spec (transmittal → spec bridge) | PASS |
| 3 | Export Word Template (download .docx, verify styles) | PASS |
| 4 | Prepare Manuscript in Word | PENDING — user doing manually |
| 5 | EPUB-First Run | PENDING |
| 6 | Print / Typst | PENDING |

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

## Queued Fixes (not yet implemented)

### Auto-update project_slug when book.title changes
- When user edits book title on transmittal, the project slug should update
- Implementation: in `setField()` in transmittal.js, when `path === 'book.title'`, slugify the value and call `PUT /api/projects/{id}` with new `project_slug`
- Backend already supports this endpoint
- Not a blocker — cosmetic/admin-side only

## Deploy Notes
- Binary is embedded-static, so every JS change requires: `make build && sudo systemctl restart srv.service`
- Live unit is `srv.service` (not `prodcal.service`) — confirmed this session
- After restart, verified via `readlink /proc/$PID/exe` showing non-deleted binary

## Next Steps
1. User prepares manuscript in Word (Phase 4) — applying tweet/metadata-p/metadata-c styles + deliberate local formatting noise
2. Resume with Phase 5 (EPUB-First Run) and Phase 6 (Typst/PDF)
3. After all phases pass, batch-fix the slug auto-update
4. Git commit all changes
