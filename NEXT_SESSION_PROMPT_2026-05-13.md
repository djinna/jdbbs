# Next session — 2026-05-13

> Mirror at `book-prod/NEXT_SESSION_PROMPT_2026-05-13.md`.
> Canonical state lives in `jdbbs/TRACKER.md`. Open that first.

## TL;DR

Five things shipped on 2026-05-12 (this is the closing snapshot of that marathon):

1. **MIG-001..004** — full migration of book-prod + uncommitted VM state into `prodcal` monorepo (`unify-typesetting` branch, 8 commits, merged as `88e5283`); VM cutover complete.
2. **SEC-008** — 6 Dependabot vulnerabilities resolved (2 critical: grpc auth-bypass, pgx memory-safety), all auto-closed by GitHub.
3. **OPS-003** — silent data-loss-in-progress caught and reversed: the real 117 MB DB lived at `/home/exedev/db.sqlite3` but MIG-004's `WorkingDirectory` change had silently switched the service to an empty Feb-25 test DB; backups had been 3.5 KB for a week. Recovered via `sqlite3 .backup` to the new path.
4. **OPS-005** — systemd orphan-process race root-caused (main.go exited 0 on errors, Type=simple, Restart=always) and fixed (`exit 1` + `Type=notify` + `daemon.SdNotify(SdNotifyReady)` + `Restart=on-failure`). Verified with 5 rapid back-to-back restarts — zero orphans.
5. **OPS-007 phases 1+2+3** — full backup pipeline overhaul:
   - hardened `backup-db.sh` (flock, size probe, rowcount probe, integrity check, drift warn, daily + monthly retention tiers, success/failure sentinels)
   - 3-2-1 satisfied: VM `~/backups/` + Cloudflare R2 (`r2:jdbbs-backups/db/`) + Mac mirror (`jbackup-pull` zsh function)
   - hourly health probe (`check-backups.sh`) + monthly restore drill (`restore-drill.sh`) + Mac-side sentinel-age display
   - 24 pre-fix junk objects culled from all three locations

Plus tracker hygiene: `jdbbs/TRACKER.md` has a pinned "Resume here" section at the top with the live verification block.

## Critical context for the next instance

- **VM**: `exedev@jdbbs.exe.xyz` (Ubuntu 24.04, Go 1.22.2, typst 0.12.0, pandoc 3.1.3)
- **Three repos**:
  - `github.com/djinna/prodcal` — live source of truth, where prodcal binary + typesetting/ live
  - `github.com/djinna/jdbbs` — tracker + future destination of the eventual rename (TRK-MIG-009)
  - `github.com/djinna/book-prod` — historical, mostly inert now
- **Live DB**: `/home/exedev/prodcal/db.sqlite3` (after OPS-003 recovery — was `/home/exedev/db.sqlite3` before MIG-004's WorkingDirectory change)
- **Twitter Years** is project 7; all other projects in the DB are tests/dummies (TRK-OPS-006 will drop them).
- **Don't `systemctl daemon-reload` without re-reading TRK-OPS-005** — the orphan-race fix depends on `Type=notify` + the binary's `sd_notify` call, and a misconfigured unit will look healthy while not actually being.

## Verification block — paste this first

```bash
# VM-side
ssh exedev@jdbbs.exe.xyz '\
  echo "=== prodcal service ===" && \
  systemctl is-active prodcal && \
  systemctl show -p MainPID --value prodcal && \
  sudo ss -ltnp "sport = :8000" | grep -v State && \
  echo "=== last backup sentinels ===" && \
  for f in .LAST-SUCCESS .LAST-R2-SUCCESS .HEALTH-OK .LAST-DRILL-SUCCESS; do \
    echo "--- $f ---"; cat ~/backups/$f 2>/dev/null || echo "MISSING"; done && \
  echo "=== crontab ===" && crontab -l | grep -v ^# && \
  echo "=== DB rowcounts ===" && \
  sqlite3 -readonly /home/exedev/prodcal/db.sqlite3 \
    "SELECT '"'"'projects='"'"' || COUNT(*) FROM projects UNION ALL SELECT '"'"'books='"'"' || COUNT(*) FROM books" && \
  echo "=== R2 ===" && rclone size r2:jdbbs-backups && rclone ls r2:jdbbs-backups/db | sort'

# Frontend smoke
curl -sI https://jdbbs.exe.xyz | head -1

# Local Mac mirror + cross-check
jbackup-pull

# Repo cleanliness on all three repos
for d in ~/jd-projects/{prodcal,jdbbs,book-prod}; do
  echo "=== $d ===" && cd "$d" && \
  git status -b --porcelain | head -3 && \
  git log --oneline -3
done
```

Expected output: prodcal active, MainPID matches listener, all 4 sentinels present with `time:` < 26h old (or freshly post-cron at 03:00 UTC), DB shows projects ≥ 1 + books ≥ 1, R2 has ≥ 1 file under db/, HTTP 200, Mac mirror has files matching VM count.

If any of those fail: **stop and triage before doing new work.** Most likely culprit if something's off tomorrow: the 03:00 UTC cron didn't fire (timezone or systemd-cron). Check `journalctl -u cron --since "03:00"`.

## Next priorities (pick one)

### Option A — TRK-OPS-006 (warm-up, ~5-10 min)

Drop the 12 test/dummy projects from the live DB. Twitter Years is project 7; everything else is disposable. Pattern:

```bash
ssh exedev@jdbbs.exe.xyz
cd /home/exedev/prodcal
# Snapshot first (we have 4 backups already, but extra is free)
bash scripts/backup-db.sh
# Then delete
sqlite3 db.sqlite3 <<SQL
BEGIN;
SELECT id, name FROM projects WHERE id != 7;
DELETE FROM book_outputs WHERE book_id IN (SELECT id FROM books WHERE project_id != 7);
DELETE FROM book_specs WHERE book_id IN (SELECT id FROM books WHERE project_id != 7);
DELETE FROM books WHERE project_id != 7;
DELETE FROM tasks WHERE project_id != 7;
DELETE FROM corrections WHERE book_id NOT IN (SELECT id FROM books);
DELETE FROM projects WHERE id != 7;
COMMIT;
VACUUM;
SELECT COUNT(*) FROM projects, books;
SQL
```

After: confirm via the admin UI; run `bash scripts/backup-db.sh && bash scripts/sync-to-r2.sh` to capture the post-cleanup state in all three backup locations.

### Option B — TRK-DESIGN-001 Ghosts parity (the real milestone)

Compile `manuscripts/ghosts/` on the VM, diff resulting PDF vs `reference/GHOSTS.pdf`. This is the "does Typst actually work for production-grade typesetting" question.

Steps:
1. `cd /home/exedev/prodcal/typesetting/scripts && bash build-ghosts.sh` (already wired up post-MIG-003).
2. Compare output vs `/home/exedev/prodcal/reference/GHOSTS.pdf`.
3. Build a parity matrix in TRACKER as TRK-DESIGN-001: page count, trim, margins, font, leading, baseline grid, drop caps, chapter numbering, running heads, page numbers, widow/orphan handling, end matter.
4. For each row: ✅ matches / ❌ visible difference / ⚠️ close-enough-to-defer.
5. File child tickets (TRK-DESIGN-002..N) for any ❌.

### Option C — Phase 1 review (educational, lower urgency)

Bottom-up walkthrough of the typesetting subsystem. Start at `typesetting/templates/series-template.typ`. Output: an annotated index in TRACKER under TRK-REV-001 covering: each file's responsibility, hot paths, technical debt observations, recommended refactors. No code changes; just understanding.

## Open questions for the user

- **Backup alarm channel.** check-backups.sh writes `.HEALTH-FAIL` but nothing pings the user. Worth wiring a Discord webhook or ntfy.sh push notification? Or is `jbackup-pull` weekly check sufficient?
- **OPS-006 confirmation.** Earlier confirmed only Twitter Years (project 7) is real. Re-confirm there's nothing in projects 1, 3, 4, 5, 6, 8-15 worth exporting first.
- **MIG-009 rename timing.** When is the right moment to rename `prodcal` → `jdbbs` across all three (repo, binary, systemd unit, backup filename prefix)? Probably after Ghosts parity (DESIGN-001) so the user has confidence in the typesetting layer before renaming things.

## Files / paths to NOT touch unless intentionally re-doing them

- `/home/exedev/prodcal/srv.service` — systemd unit; orphan-race fix lives here
- `/home/exedev/prodcal/cmd/srv/main.go` — `os.Exit(1)` is load-bearing
- `/home/exedev/prodcal/srv/server.go` — `net.Listen` + `daemon.SdNotify` ordering matters
- `/home/exedev/prodcal/scripts/backup-db.sh` — env thresholds are baseline-tuned to current data
- `/home/exedev/.archive-old-db-pre-recovery/` — original pre-recovery DB; keep for forensics
