# Session prompt — next (2026-05-29)

Run `jpull` first. Standard pre-flight:

```bash
ssh exedev@jdbbs.exe.xyz '\
  systemctl is-active prodcal && \
  sqlite3 -readonly /home/exedev/prodcal/db.sqlite3 \
    "SELECT migration_number FROM migrations ORDER BY migration_number DESC LIMIT 3"'
curl -sI https://jdbbs.exe.xyz | head -1
```

Expect: `prodcal active`, `16/15/14`, `HTTP/2 200`.

## What landed 2026-05-28

- **TRK-DEV-007 done** — diff-vs-previous UI in the compile-history panel. `srv/static/admin.html` only; new JS helpers `tsSpecDiff`, `tsParseCorrectionsYAML`, `tsCorrectionsDiff` + inline expand-in-place panel per row. Empty diffs render explicit copy; legacy pre-snapshot rows degrade gracefully. No API / Go / schema changes.

## Recommended next ticket

Pick whichever is on the user's plate. Reasonable defaults in priority order:

1. **TRK-DESIGN-001 (CP-5, Ghosts parity matrix)** — release-confidence gate after CP-1..CP-4. Note: `docs/GHOSTS_PARITY_2026-05-26.md` already exists from a prior orchestrator subagent run — read it first before adding anything; the work may already be partially done.
2. **TRK-OPS-006** — drop the 12 test/dummy projects from prod DB. Mechanical (backup → `DELETE FROM projects WHERE id != 7` → verify cascades). ~5-10 min. **Open question first:** confirm with user that nothing under projects 1-6/8-12 is worth exporting.
3. **TRK-DEV-004 Phase A** — special-typography preservation class data model. ~1 session for Phase A alone.
4. **TRK-DEV-008 item 1** — case-insensitive `find` flag on corrections; pick only if a real correction set has surfaced the pain.

## Don't touch without re-reading the relevant TRK entry

- prodcal systemd unit (TRK-OPS-005, `Type=notify` + sd_notify).
- `backup-db.sh` env vars (TRK-OPS-007 phase 1 baseline-tuned).

## Open questions still pending for the user

- 12 test projects in prod DB — disposable or any worth exporting first?
- Real alert channel (Discord webhook / ntfy.sh / email) for backup-health failures?
- VM-side rename (`/home/exedev/prodcal/` → `/home/exedev/jdbbs/`, plus systemd unit) — defer indefinitely or schedule?
