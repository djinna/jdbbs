# Session handoff — 2026-07-04 (pre-launch review + local desktop app)

## Status: shipped & verified
- **5 launch blockers** fixed, deployed to prod (VM `jdbbs.exe.xyz`, `prodcal.service`), and verified live: client-auth gate (401 for anon on protected client), DB pragmas (DSN `_pragma` + `SetMaxOpenConns(1)`), safe restore runbook, spec-autosave race, Go 1.26.4 + `x/net` (govulncheck = 0).
- **Entire HIGH + QOL + NIT tier** from `docs/reviews/LAUNCH-TRIAGE.md` fixed (fan session), committed, pushed, deployed. Includes admin toast system, escaping+rate-limit on email, Files-tab convert/retry, orphan-client fix, favicon/robots, `srv`→`prodcal` doc/unit rename, README rewrite, Docker path removed, blob-dedupe migrations 017/018.
- **Persistent local + desktop app**: `internal/localrun` (loopback admin-header proxy), `cmd/prodcal-local` (headless), `cmd/prodcal-app` (WebKit window, `//go:build darwin`), `scripts/build-mac-app.sh` (double-clickable `ProdCal.app`). Full DOCX→PDF/EPUB+preflight verified locally (pandoc/typst/python-docx/fonts all present).
- Prod = GitHub = local all at the latest `main`. Prod DB: 3 projects, 7 book rows (6 = *Twitter Years* dev/real iterations, 1 = zoo test) — all legit history, **no orphans**.

## Decisions locked
- **GitHub is the single source of truth (hub).** Mac + VM are both consumers. VM only ever `git pull --ff-only`. NOT the reverse (VM-as-hub rejected: it's a live prod checkout, can't build the darwin-only desktop code, single fragile box).
- **Local desktop app may diverge from the VM** in purpose (personal operator tool vs client SaaS). Mechanism: keep shared work on `main`; start a `desktop` branch ONLY when making a VM-incompatible change (e.g. hiding the SaaS/portal/email surface); always push that branch to GitHub (backup); pull engine fixes from `main` selectively. No fork today — `main` safely holds both (desktop shell is `//go:build darwin`, VM never compiles it).
- **Manuscript version history is a FEATURE.** Every `.docx` re-upload + its outputs is traceable history to keep; must always allow a fresh `.docx` upload. Any future `book_outputs` pruning must be conservative (keep all upload rows + latest N outputs per format) — never blunt deletion.
- **VM is canonical for final print output;** local is for drafting/proofing (tool versions can drift).

## Why the local app shows no projects
By design (model A): the desktop app uses its own isolated DB at `~/Library/Application Support/ProdCal/db.sqlite3`, separate from prod. It starts empty. See the DEFERRED task below to seed it.

## DEFERRED — seed local app with prod PROJECT data (transmittals yes, manuscripts NO)
Goal: populate the local app with prod's projects/clients/**transmittals**/tasks/journal/file-log/corrections/book-specs, but WITHOUT the manuscript files or generated outputs (books/book_outputs/preflights — the 352 MB of blobs). One-time snapshot copy (local diverges after; not ongoing sync).

Approach (all on a COPY — never touch the live prod DB):
1. On the VM, copy the DB and strip manuscript tables, then shrink:
   ```bash
   ssh exedev@jdbbs.exe.xyz 'cp /home/exedev/prodcal/db.sqlite3 /tmp/seed.sqlite3 && sqlite3 /tmp/seed.sqlite3 "DELETE FROM book_outputs; DELETE FROM manuscript_preflights; DELETE FROM books; VACUUM;" && gzip -f /tmp/seed.sqlite3 && ls -la /tmp/seed.sqlite3.gz'
   ```
   (Keeps: projects, clients, transmittals, transmittal_versions, tasks, journal, file_log, corrections, book_specs. Drops: books, book_outputs, manuscript_preflights. `sqlite3` CLI has FK off by default, so the explicit DELETEs are safe in any order. Result should be a few MB.)
2. Pull it to the Mac (user runs — ssh/scp is user-run):
   ```bash
   scp exedev@jdbbs.exe.xyz:/tmp/seed.sqlite3.gz ~/Downloads/
   ```
3. Install it as the local DB (stop the app first — SQLite single-writer):
   ```bash
   pkill -f prodcal-app; pkill -f prodcal-local
   D="$HOME/Library/Application Support/ProdCal"
   gunzip -c ~/Downloads/seed.sqlite3.gz > "$D/db.sqlite3"
   rm -f "$D/db.sqlite3-wal" "$D/db.sqlite3-shm"
   cd ~/jd-projects/jdbbs && ./scripts/build-mac-app.sh && open ProdCal.app
   ```
Notes: schema parity holds (prod + local both at migration 018), so the app opens the seeded DB without migration conflicts. `.prodcal-secret` is untouched. Client password hashes come along but are irrelevant locally (the launcher injects the admin header). Book rows are gone, so the Files/Typesetting tabs start empty locally — you re-upload a fresh `.docx` when you want to typeset locally.

## Other open (all optional, non-launch)
- One-time `VACUUM` on prod to reclaim the space 018 freed (file stays ~352 MB until then): `ssh … 'cd /home/exedev/prodcal && sudo systemctl stop prodcal && sqlite3 db.sqlite3 "VACUUM;" && sudo systemctl start prodcal && curl -s localhost:8000/healthz'`. Off-peak; briefly locks the DB. NOTE: real bloat is `book_outputs` HISTORY (book 8 = 11 pdf/7 epub rows), which 018 did NOT prune — leave it (it's history) unless it grows painful.
- R2 restore-drill cron (monthly): `30 4 1 * * /home/exedev/prodcal/scripts/r2-restore-drill.sh`.
- Kill the stale local `prodcal-app` (pid was 93599) if still running an old binary.

## How to run the local app
```bash
cd ~/jd-projects/jdbbs && ./scripts/build-mac-app.sh && open ProdCal.app   # native window, admin UI
# or headless in a browser: ./scripts/run-local.sh   (don't run both at once — same data dir)
```
Full local setup detail (python-docx install, `JDBBS_TYPESETTING_DIR`, tool table): `docs/LOCAL-USAGE.md`.

## Deploy (reminder)
`ssh exedev@jdbbs.exe.xyz 'cd /home/exedev/prodcal && git pull --ff-only origin main && make build && sudo systemctl restart prodcal && sleep 2 && systemctl is-active prodcal && curl -s localhost:8000/healthz'` — prepend `bash scripts/backup-db.sh &&` before any migration-bearing deploy.
