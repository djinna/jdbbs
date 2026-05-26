# Session prompt — TRK-DEV-007 (diff-vs-previous UI)

> Use this as the kick-off prompt for a fresh Claude Code session.
> Concurrent-safe with TRK-DESIGN-001 (observational, docs-only) and
> TRK-OPS-006 (DB-only) — those are being handled by the orchestrator
> in parallel. Do NOT pick up DEV-004 or DEV-008 in this session —
> they share admin.html and will merge-conflict.

Run `jpull` first. Then standard pre-flight:

```bash
ssh exedev@jdbbs.exe.xyz '\
  systemctl is-active prodcal && \
  sqlite3 -readonly /home/exedev/prodcal/db.sqlite3 \
    "SELECT migration_number FROM migrations ORDER BY migration_number DESC LIMIT 3"'
curl -sI https://jdbbs.exe.xyz | head -1
```

Expect: `prodcal active`, `16/15/14`, `HTTP/2 200`.

## What landed since 2026-05-26

- **TRK-MIG-006 done** — corrections auto-apply in both PDF + EPUB pipelines. Migration 016 added `book_outputs.corrections_snapshot TEXT NULL`. Commits 9af05ad, d69b4e3, 3918527, 10a07c7.
- **TRK-DEV-008 filed** — patcher ergonomics follow-ups (P3, pick when warranted, don't batch).

## TASK — TRK-DEV-007 (diff-vs-previous UI)

Full plan in `docs/TRACKER.md` under TRK-DEV-007. Summary:

Both `spec_snapshot` (DEV-006) and `corrections_snapshot` (MIG-006) are persisted per compile. The history panel lists them but surfaces nothing about *what changed* between two compiles. Right now: user compiles twice, downloads both PDFs, opens them side-by-side, tries to remember which spec edits and which corrections were active. The data exists; only the UI gap is left.

**No new API needed** — `GET /api/books/{id}/outputs?include=spec,corrections` already returns both per row.

### Implementation steps

1. **Verify snapshots are populated** (~2 min):

   ```bash
   ssh exedev@jdbbs.exe.xyz 'sqlite3 -readonly /home/exedev/prodcal/db.sqlite3 \
     "SELECT id, output_format, created_at, length(spec_snapshot), length(corrections_snapshot) \
      FROM book_outputs WHERE book_id = (SELECT id FROM books WHERE project_id=7 LIMIT 1) \
      ORDER BY created_at DESC LIMIT 5"'
   ```

   Newest rows should have both lengths > 0; pre-migration-015/016 rows will be NULL. If everything is NULL, something's broken upstream — stop and ask.

2. **JS helper — shallow spec diff** (~30 min):
   - Walk two JSON objects recursively, emit a flat list of `{ path, before, after }` for changed leaf fields.
   - Skip identical values. Treat missing-on-one-side as add/remove.
   - Format paths dotted: `typography.base_size_pt`, `running_heads.verso`.

3. **JS helper — corrections diff** (~20 min):
   - Parse both YAML blobs (or both already-JSON-encoded if that's the snapshot shape — verify).
   - Compute set-difference on entries keyed by `find` text (or row ID if the YAML carries it).
   - Emit added / removed / status-changed entries.

4. **UI in admin.html** (~30-45 min):
   - In the compile-history panel (under PDF + EPUB Compile buttons), add a "diff" button to every row except the oldest visible.
   - Default diff target = immediately-preceding row chronologically. Optional polish: dropdown to compare against any other row in the panel — defer if time-pressed.
   - On click, expand-in-place (or modal — pick what reads cleaner in your zone of the SPA) with two columns: spec diff left, corrections diff right.
   - Format example for spec changes:
     ```
     typography.base_size_pt:  10 → 12
     typography.body_font:     "Libertinus Serif" → "Plantin MT Pro"
     ```
   - Format example for corrections:
     ```
     + added:   { find: "Venkatesh", replace: "Venkat" }
     - removed: { find: "iphone", replace: "iPhone" }
     ~ changed: { find: "alchemy" } status: pending → applied
     ```
   - Empty diff renders explicitly: "no spec or corrections changes since <earlier timestamp>". Distinguishes "compiled twice with same inputs" from "snapshot not recorded."
   - Legacy rows with NULL snapshot: render "(no snapshot recorded)" in the dropdown / disable diff button when they're the target.

5. **Smoke test live** (~15 min):
   - Open `https://jdbbs.exe.xyz/admin/`, Typesetting tab, Twitter Years (project 7).
   - Compile once.
   - Edit a spec field (e.g. base_size_pt 10 → 12), compile again.
   - Click diff on the newer row → expect to see `typography.base_size_pt: 10 → 12`.
   - Add a correction, compile again. Click diff → expect to see `+ added` entry.
   - Mark that correction applied (no recompile needed), edit something unrelated, compile. Click diff → expect to see corrections status change AND any spec change.
   - Try diffing against a legacy pre-snapshot row — should degrade gracefully, not throw.

### Acceptance (paste from TRACKER for explicit check)

- After two compiles with different spec values, the diff button on the newer row shows the field changes.
- After two compiles with different corrections, the diff shows the added/removed/changed entries.
- An empty diff is rendered explicitly, not as a blank panel.
- Legacy pre-snapshot rows are handled gracefully (no JS error).

## DEPLOY

```bash
ssh exedev@jdbbs.exe.xyz 'cd /home/exedev/prodcal && git pull --ff-only && \
  go build -o prodcal ./cmd/srv && sudo systemctl restart prodcal && \
  sleep 2 && systemctl is-active prodcal'
curl -sI https://jdbbs.exe.xyz | head -1
```

(JS-only changes don't require Go rebuild, but rebuilding is harmless and confirms nothing else regressed. Static files are embedded into the binary.)

Don't touch the systemd unit (TRK-OPS-005). Direct push to `main` is blocked by auto-mode classifier — push yourself, then ssh deploy.

## WRAP-UP TASKS (after smoke passes)

1. Update `docs/TRACKER.md`: mark TRK-DEV-007 done with the call trace / file refs.
2. Update Resume here block: note DEV-007 done, advance to whichever is next on the user's plate.
3. Write `docs/NEXT_SESSION_PROMPT_2026-05-27.md` (or whatever date applies) — likely seeding TRK-DEV-004 Phase A (special-typography data model) or whatever the parallel DESIGN-001 work surfaces.
4. Commit, push, optionally redeploy if you made code changes.

## NON-GOALS / WHAT NOT TO DO

- **Don't add a new API endpoint** — `?include=spec,corrections` already returns everything you need.
- **Don't bundle TRK-DEV-008 items** even though the corrections-warnings ticket (item 4) lives in the same panel zone. Separate session, separate merge surface.
- **Don't refactor the history-panel JS just because you're in there.** Add the diff feature; resist the cleanup urge unless you find something genuinely broken.
- **Don't touch `srv/books.go` or `srv/epub.go`** — DESIGN-001 (orchestrator's parallel work) may surface child tickets that touch them; staying clear avoids conflicts.

## CONCURRENT-WORK AWARENESS

At session-start time, two other workstreams may be in flight:
- **TRK-DESIGN-001 (Ghosts parity matrix)** — orchestrator running this in a subagent. Output is docs/TRACKER.md additions + possibly a new `docs/GHOSTS_PARITY_2026-05-26.md`. No code changes expected.
- **TRK-OPS-006 (drop test projects)** — orchestrator running this directly. DB-only changes on VM; nothing in the repo.

If you find your `git pull` brings in TRACKER changes mid-session, rebase locally and continue. If you find a conflict on `docs/TRACKER.md`, accept both sides (each session is adding distinct entries — DEV-007 close vs DESIGN-001 children — and the merge is almost always trivial).
