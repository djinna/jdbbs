# Next session — 2026-05-26

> Drafted 2026-05-25 (end of session that closed TRK-MIG-009 docs reorg
> and shipped TRK-DEV-004). The parallel session that closed TRK-DEV-002
> + TRK-MIG-007 also filed TRK-DEV-005 and TRK-DEV-006 as the natural
> follow-ups. This prompt sequences those next.

## TL;DR

Run `jpull` first. Then three phases in order:

1. **Audit TRK-DEV-003** (15 min, read-only). May already be done — close it or scope it.
2. **Ship TRK-DEV-005** (~1 hour). Compile-history panel.
3. **Ship TRK-DEV-006** (~1.5 hours). Snapshot spec JSON per compile.

The two implementation tickets together = the diagnostic loop. Every subsequent QA session (Twitter Years smoke, Ghosts parity, special-typography testing) needs to compare PDFs across spec edits. Right now you can't. After this session: every compile listable, downloadable, and self-documenting.

## Why this sequencing

- **DEV-003 audit first** because the parallel session's pattern in the last cycle (commit `bcbbf2d`) was: "this ticket is actually already wired in a prior session, just not closed." Evidence DEV-003 may be the same: `srv/epub.go::handleGenerateEPUB` already calls `GetBookSpec`, `parseEPUBSpec`, and `GetBookSpecCover`. Free 3 hours back if the audit closes it.
- **DEV-005 + DEV-006 as the next ticket** because every other open ticket (TRK-DEV-004 Twitter Years smoke, TRK-DESIGN-001 Ghosts parity, TRK-DESIGN-003 CSS drift audit, TRK-MIG-006 corrections round-trip) needs the diagnostic loop to be productive.
- **TRK-DEV-002 live smoke** uncovered a silent `book.with(config:)` no-op that was masking real spec changes — this is exactly the class of bug the diagnostic loop catches on the first comparison instead of the tenth. April 2026's "Phase 6 Print/Typst pending" frustration was very likely this same bug.

## Phase 1 — Audit TRK-DEV-003 (15 min, read-only)

TRK-DEV-003 plan-of-record: wire spec → EPUB compile pipeline. Quick evidence it may already be done: `srv/epub.go::handleGenerateEPUB` already calls `GetBookSpec`, `parseEPUBSpec`, and `GetBookSpecCover`.

Read `srv/epub.go` end-to-end. Confirm what the EPUB pipeline actually consumes from `book_spec`. Then either:

- **(a)** close TRK-DEV-003 with a documented call trace (same pattern as commit `bcbbf2d` closed TRK-DEV-002), or
- **(b)** scope down to whatever's actually missing and update the ticket.

**Do NOT start coding EPUB changes** — this phase is audit + close/rescope only.

## Phase 2 — Ship TRK-DEV-005 (~1 hour)

Compile-history panel. Full plan in `docs/TRACKER.md` under TRK-DEV-005. Summary:

1. Add `ListBookOutputs(book_id, limit)` to `db/queries/book_outputs.sql` + regenerate dbgen.
2. Endpoint: `GET /api/books/{id}/outputs` → JSON array (no bytes). Fields: id, format, source_filename, created_at, length(output_data). Index `idx_book_outputs_book_format_created` already exists.
3. Per-output download: `GET /api/books/{id}/outputs/{output_id}/download` — stream `output_data` with the same timestamp-suffixed `Content-Disposition` as the latest-download endpoint (commit `792c76d`). Verify `output_id` belongs to `book_id` before serving (path traversal hygiene).
4. Admin SPA: small panel under each Compile button in the Typesetting tab. Last ~20 outputs as: `timestamp · format · size · download link`. Live-refresh after a successful compile completes (the existing polling loop knows when status flips to ready).

**Acceptance:** compile twice with different spec values; both PDFs appear in the panel; downloading the older row produces a PDF that matches the earlier spec.

## Phase 3 — Ship TRK-DEV-006 (~1.5 hours)

Snapshot spec JSON into book_outputs per compile. Full plan in `docs/TRACKER.md` under TRK-DEV-006. Summary:

1. **Migration 015:** add `book_outputs.spec_snapshot TEXT NULL` (JSON blob copied from `book_specs.data` at compile time). NULL for legacy rows.
2. **Write site:** in `srv/books.go::runConversion` AND `srv/epub.go`'s equivalent, after the `buildTypstConfig` lookup, persist the raw spec JSON alongside the artifact. A small helper returning `(configString, rawJSON)` avoids double-fetching.
3. **API:** extend `GET /api/books/{id}/outputs` (TRK-DEV-005) to include `spec_snapshot` when `?include=spec` is passed.

**Acceptance:** new compiles persist spec; API returns it on request; legacy rows still listable with `spec_snapshot: null`.

## Deploy

Standard pattern:

```bash
ssh exedev@jdbbs.exe.xyz
cd /home/exedev/prodcal
git pull
go build -o prodcal ./cmd/srv
sudo systemctl restart prodcal
systemctl is-active prodcal && curl -sI https://jdbbs.exe.xyz | head -1
```

**Don't touch the systemd unit** (TRK-OPS-005 — orphan-race fix is path-sensitive).

## Wrap-up tasks (after Phase 3)

1. Update `docs/TRACKER.md` "Resume here" block — new live state + next priority (likely TRK-DEV-004 Phase A, or TRK-DESIGN-003 if the diagnostic loop is the priority before more spec work).
2. Write `docs/NEXT_SESSION_PROMPT_2026-05-27.md` per the handoff convention.
3. Commit + push.

## Live state at hand-off (verify with the standard VM block)

- VM at commit `c9aed90`; `prodcal.service` active.
- TRK-DEV-002 + TRK-MIG-007 closed. Spec → Typst → PDF pipeline live-smoked on project 7 (Twitter Years).
- TRK-DEV-004 (special-typography preservation) open; will bite during Twitter Years smoke of any tweet-snapshot-heavy chapter — but **not** something to start here.

## What NOT to do this session

- Don't start TRK-DEV-004 implementation (special-typography). It needs the diagnostic loop first to verify preservation.
- Don't start TRK-DESIGN-001 (Ghosts parity). Ghosts is anthology — needs CP-8 work first or you'll get false-negative "Typst broke" signals that are actually "anthology unsupported."
- Don't change the systemd unit, backup scripts, or VM directory paths (TRK-OPS-005, TRK-OPS-008, TRK-OPS-007).
