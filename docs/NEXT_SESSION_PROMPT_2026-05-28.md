# Next session — 2026-05-28

Run `jpull` first. Then the standard pre-flight:

```bash
ssh exedev@jdbbs.exe.xyz '\
  systemctl is-active prodcal && \
  cat ~/backups/.HEALTH-OK && \
  sqlite3 -readonly /home/exedev/prodcal/db.sqlite3 \
    "SELECT migration_number FROM migrations ORDER BY migration_number DESC LIMIT 3"'
curl -sI https://jdbbs.exe.xyz | head -1
```

Expect: `prodcal active`, `OK`, `16/15/14`, `HTTP/2 200`.

## What landed last session (2026-05-26)

**TRK-MIG-006 (CP-3) done.** Corrections round-trip is live on both PDF and EPUB. The patcher (`apply-corrections-docx.py`) now walks body + tables + headers/footers + footnotes + endnotes; in-memory metadata (book.Title/Author, spec.Title/Author/Subject/Description) is also patched so pandoc/typst metadata flags reflect corrections. Migration 016 added `book_outputs.corrections_snapshot TEXT NULL`; `?include=spec,corrections` exposes both per row. Commits `9af05ad`, `d69b4e3`, `3918527`, `10a07c7`.

**TRK-DEV-008 filed (P3)** with eight ergonomics improvements for the corrections workflow (case-insensitive flag, whole-word matching, per-scope filters, surfaced patcher warnings, dry-run preview in the SPA, snapshot diff in the history panel, and two auto-mark-applied items that need data-model thought first).

## PHASE 1 — pick one CP-5-ish thing (~2-3 hours)

The v1 workflow is now feature-complete for "type → spec → compile → archive → correct → recompile." The last release-confidence gate is the parity matrix:

- **TRK-DESIGN-001 (CP-5)** — Ghosts parity matrix. Compare the current PDF/EPUB output against the reference "Ghosts in Machines" book on a fixed checklist (typography, spacing, headers, footnotes, TOC, cover). Identify any drift, file follow-ups for anything that needs design work, then declare v1 shippable.

## PHASE 2 — pick one of the small things (~10-60 min)

- **TRK-OPS-006** (5-10 min, mechanical): drop the 12 test/dummy projects from the prod DB. Backup → `DELETE FROM projects WHERE id != 7` → verify cascades cleaned up books/specs/corrections/transmittals.
- **TRK-DEV-007** (1-2 hours): diff-vs-latest UI for the compile-history panel. The API (`?include=spec,corrections`) already returns the data; render a minimal field-by-field diff between consecutive `spec_snapshot` blobs in the SPA. Now doubly useful with `corrections_snapshot` also available.
- **TRK-DEV-008 item 1** (~30 min): case-insensitive flag per correction. The cheapest patcher ergonomics win — add `case_insensitive: true` to the YAML schema + a column on `corrections`, default off (preserves today's `iphone → iPhone` example contract). Useful the next time a real correction set hits the `alchemy` vs `Alchemy` problem.

## PHASE 3 — larger lift (defer unless time)

- **TRK-DEV-004** — Special-typography preservation class. ~3 sessions total; Phase A is the data model + preflight side. Big enough to be its own session.

## DEPLOY (standard pattern)

```bash
ssh exedev@jdbbs.exe.xyz 'cd /home/exedev/prodcal && git pull --ff-only && \
  go build -o prodcal ./cmd/srv && sudo systemctl restart prodcal && \
  sleep 2 && systemctl is-active prodcal'
curl -sI https://jdbbs.exe.xyz | head -1
```

Don't touch the systemd unit (TRK-OPS-005). Direct push to `main` is hard-blocked by Claude Code's auto-mode classifier — run `git push origin main` yourself, then the ssh deploy.

## WRAP

After phases land, update `docs/TRACKER.md` Resume here block and write `docs/NEXT_SESSION_PROMPT_2026-05-29.md`. Commit + push.

## Open follow-ups (low priority, file when warranted)

- **TRK-DEV-007** — diff-vs-latest UI (see PHASE 2).
- **TRK-DEV-008** — corrections patcher ergonomics (see PHASE 2 + TRACKER entry for the full list).
- EPUB spec gaps: `epub.embed_fonts` is parsed but ignored; `epub.landmarks` UI field has no consumer. Fold into TRK-DESIGN-003.
- VM-side rename (`/home/exedev/prodcal/` → `/home/exedev/jdbbs/`) — defer indefinitely or schedule? Worth a yes/no this session.
