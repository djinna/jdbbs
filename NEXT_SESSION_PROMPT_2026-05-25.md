# Next session — 2026-05-25

> Snapshot of `TRACKER.md` "Resume here" block as of 2026-05-25.
> Canonical state lives in `TRACKER.md` (this repo, root). Open that first.

## Just shipped — TRK-MIG-009 (canonical repo cutover)

- `djinna/prodcal` HEAD force-pushed into `djinna/jdbbs:main` (commit `83e21f2` absorbs `TRACKER.md`, `MIGRATION_LOG.md`, `NEXT_SESSION_PROMPT_2026-05-13.md`, `CLAUDE.md`; gitignores `.claude/`).
- `djinna/prodcal` and `djinna/book-prod` archived (read-only) on GitHub; redirects from old URLs still resolve.
- Safety tags pushed: `pre-jdbbs-rename-2026-05-25` (prodcal) + `pre-overwrite-2026-05-25` (jdbbs).
- VM remote URL updated `git@github.com:djinna/prodcal.git` → `https://github.com/djinna/jdbbs.git`; pulled cleanly; service stayed active.
- Local clones renamed: `~/jd-projects/prodcal` → `~/jd-projects/jdbbs`; old jdbbs preserved at `~/jd-projects/jdbbs-bootstrap-pre-2026-05-25`.
- Local `~/jd-projects/book-prod` deprecated (single canonical repo now).

**NOT renamed** (deferred as TRK-OPS-008): VM directory `/home/exedev/prodcal/`, systemd unit `prodcal.service`, binary `prodcal`. Cosmetic only, path-sensitive (orphan-race fix per TRK-OPS-005), no user-facing impact.

Plus tracker hygiene: TRK-FLOW-001 (session-prompt proliferation) closed as done — problem dissolved with single-repo cutover.

## Cross-machine setup (do once per Mac)

```bash
# Clone the canonical repo
git clone https://github.com/djinna/jdbbs.git ~/jd-projects/jdbbs

# Add jpull function to ~/.zshrc — fetch + status the canonical repo,
# warn on any drift, print the TRACKER "Resume here" block automatically.
# (See ~/jd-projects/jdbbs/scripts/jpull.sh for the canonical version, or
#  copy from the Mac where it already exists.)

# Bootstrap Claude Code skills (gstack)
cd ~/jd-projects/jdbbs/.claude/skills/gstack && ./setup
```

## Critical context for the next instance

- **VM**: `exedev@jdbbs.exe.xyz` (Ubuntu 24.04, Go 1.22.2, typst 0.12.0, pandoc 3.1.3)
- **GitHub remote**: `https://github.com/djinna/jdbbs.git` (the only live repo)
- **Live DB**: `/home/exedev/prodcal/db.sqlite3` (VM directory is still `prodcal/` — see TRK-OPS-008)
- **systemd unit**: `prodcal.service` (still — `Type=notify`, `Restart=on-failure`, orphan-race fix per TRK-OPS-005)
- **Twitter Years** is project 7; all other projects in the DB are tests (TRK-OPS-006 will drop them).
- **Don't `systemctl daemon-reload` without re-reading TRK-OPS-005** — orphan-race fix depends on `Type=notify` + the binary's `sd_notify` call.

## Verification block — paste this first

```bash
# Local — sync
jpull   # fetch jdbbs + print TRACKER Resume here

# VM-side
ssh exedev@jdbbs.exe.xyz '\
  echo "=== service ===" && \
  systemctl is-active prodcal && \
  systemctl show -p MainPID --value prodcal && \
  sudo ss -ltnp "sport = :8000" | grep -v State && \
  echo "=== sentinels ===" && \
  for f in .LAST-SUCCESS .LAST-R2-SUCCESS .HEALTH-OK .LAST-DRILL-SUCCESS; do \
    echo "--- $f ---"; cat ~/backups/$f 2>/dev/null || echo "MISSING"; done && \
  echo "=== crontab ===" && crontab -l | grep -v ^# && \
  echo "=== DB rowcounts ===" && \
  sqlite3 -readonly /home/exedev/prodcal/db.sqlite3 \
    "SELECT '"'"'projects='"'"' || COUNT(*) FROM projects UNION ALL SELECT '"'"'books='"'"' || COUNT(*) FROM books" && \
  echo "=== R2 ===" && rclone size r2:jdbbs-backups && rclone ls r2:jdbbs-backups/db | sort'

# Frontend smoke
curl -sI https://jdbbs.exe.xyz | head -1

# Local Mac mirror (only on Macs with jbackup-pull set up)
jbackup-pull 2>/dev/null || echo "jbackup-pull not configured on this Mac (TODO)"

# Repo cleanliness
cd ~/jd-projects/jdbbs && git status -b --porcelain | head -3 && git log --oneline -3
```

Expected: prodcal active, MainPID matches listener, all 4 sentinels present with `time:` < 26h old, DB shows projects ≥ 1 + books ≥ 1, R2 has ≥ 1 file under db/, HTTP 200, local repo clean and at canonical HEAD.

## Next priorities (pick one)

### Option A — TRK-OPS-006 (warm-up, ~5-10 min)

Drop the 12 test/dummy projects from the live DB. Twitter Years is project 7; everything else is disposable. Full SQL block in TRACKER under TRK-OPS-006.

### Option B — TRK-DESIGN-001 Ghosts parity (the real milestone)

Compile `manuscripts/ghosts/` on the VM, diff resulting PDF vs `reference/GHOSTS.pdf`. The "does Typst actually work" question. Build parity matrix: page count, trim, margins, font, leading, baseline grid, drop caps, chapter numbering, running heads, page numbers, widow/orphan handling, end matter. File child tickets (TRK-DESIGN-002..N) for any ❌.

### Option C — Phase 1 review (educational)

Bottom-up walkthrough of typesetting subsystem. Start at `typesetting/templates/series-template.typ`. Annotated index in TRACKER as TRK-REV-001.

## Open questions for the user

- **Backup alarm channel.** check-backups.sh writes `.HEALTH-FAIL` but nothing pings the user. Discord webhook? ntfy.sh? Or is `jbackup-pull` weekly sufficient?
- **OPS-006 confirmation.** Re-confirm there's nothing in projects 1, 3, 4, 5, 6, 8-15 worth exporting first.
- **jbackup-pull on second Mac.** Asymmetry observed 2026-05-25 — function exists on one Mac, not this one. Set up everywhere or pick a single "primary mirror" Mac?
- **TRK-OPS-008 (VM rename).** Defer indefinitely, or schedule a focused session?

## Files / paths to NOT touch unless intentionally re-doing them

- `/home/exedev/prodcal/srv.service` — systemd unit; orphan-race fix lives here (TRK-OPS-005)
- `/home/exedev/prodcal/cmd/srv/main.go` — `os.Exit(1)` is load-bearing
- `/home/exedev/prodcal/srv/server.go` — `net.Listen` + `daemon.SdNotify` ordering matters
- `/home/exedev/prodcal/scripts/backup-db.sh` — env thresholds are baseline-tuned to current data
- `/home/exedev/.archive-old-db-pre-recovery/` — original pre-recovery DB; keep for forensics
