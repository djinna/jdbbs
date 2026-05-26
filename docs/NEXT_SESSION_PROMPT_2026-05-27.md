# Next session — 2026-05-27

Run `jpull` first. Then:

## PHASE 1 — Live smoke of TRK-DEV-005/006 (10 min, browser)
Compile-history panel shipped 2026-05-26 in commit `56e8256`. Migration 015 is
applied on the VM (verified). Service is healthy. What's NOT verified yet is the
real round-trip in a browser.

1. Open `https://jdbbs.exe.xyz/admin/`, go to Typesetting tab, pick The Twitter
   Years (project 7).
2. Select the manuscript in "ts-pdf-book", confirm the new
   "Recent compiles (N)" panel renders under the Compile PDF button with N≥1
   row (the previous 2026-05-25 smoke compile).
3. Edit a spec value (e.g. `typography.base_size_pt` 10 → 12), click Compile PDF,
   wait for status to flip to Done. The panel should auto-refresh and a new row
   should appear at the top.
4. Click the OLDER row's `download` link — confirm it downloads a PDF with the
   pre-edit body size (not the just-compiled one). Filename should suffix the
   older row's `created_at`, not "now."
5. Same loop for EPUB under the Generate EPUB button.
6. Confirm `GET /api/books/<bid>/outputs?include=spec` returns rows with
   `spec_snapshot` populated for the new compile (and `null` for the legacy
   pre-migration row).

If anything misbehaves, the most likely culprits are:
- `bookAuth` returning 401/403 on the unauthenticated browser session →
  re-check `requireAuth` semantics for project-linked books.
- Stale `admin.html` cache → hard reload.
- `tsRefreshHistory` not called → confirm `tsLoadBooks` (around line 2580 in
  admin.html) reaches the new `tsRefreshHistory()` call.

## PHASE 2 — TRK-MIG-006 (corrections pipeline round-trip) (~3 hours)
See `docs/TRACKER.md` entry. Summary:
- Currently corrections are stored in SQLite and exported by hand to YAML and
  fed to `typesetting/scripts/apply-corrections.py` (and the docx variant).
- Goal: after a successful PDF/EPUB generate in `srv/`, materialize the
  corrections-set YAML in-memory and invoke the corresponding patcher
  automatically. A round-tripped correction should appear in the regenerated
  artifact without any manual step.
- Touch points: `srv/corrections.go` (existing CRUD), `srv/books.go::runConversion`
  (insertion point after pandoc-to-typst), `srv/epub.go::runEPUBGeneration`,
  `typesetting/scripts/apply-corrections{,-docx}.py`.
- Acceptance: create a correction in the SPA, recompile, confirm the change
  shows up in the new PDF (and the new EPUB).
- Bonus: capture `correction_set_id` (or a version stamp) into the
  `book_outputs` row alongside `spec_snapshot` from TRK-DEV-006 — same lineage
  shape. Probably warrants a small schema addition (migration 016) but verify
  the data model first.

## PHASE 3 — pick one of:
- **TRK-OPS-006** (5-10 min, mechanical): drop the 12 test/dummy projects from
  the prod DB. Backup → `DELETE FROM projects WHERE id != 7` → verify cascades.
- **TRK-DESIGN-001 (CP-5)** (~2-3 hours): Ghosts parity matrix. The
  release-confidence check after CP-1..CP-4.

## DEPLOY (standard pattern)
```bash
ssh exedev@jdbbs.exe.xyz 'cd /home/exedev/prodcal && git pull --ff-only && \
  go build -o prodcal ./cmd/srv && sudo systemctl restart prodcal && \
  sleep 2 && systemctl is-active prodcal'
curl -sI https://jdbbs.exe.xyz | head -1
```
Don't touch the systemd unit (TRK-OPS-005).

## WRAP
After phases land, update `docs/TRACKER.md` Resume here block and write
`docs/NEXT_SESSION_PROMPT_2026-05-28.md`. Commit + push.

## Open follow-ups (low priority, file when warranted)
- **TRK-DEV-007** — diff-vs-latest UI for the compile-history panel. The
  `?include=spec` API exists; only the renderer is missing. Render minimal
  field-by-field diff between consecutive `spec_snapshot` blobs.
- EPUB spec gaps: `epub.embed_fonts` is parsed but ignored;
  `epub.landmarks` UI field has no consumer. Fold into TRK-DESIGN-003.
