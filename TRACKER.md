# JDBBS Tracker

Single source of truth for the review, security/workflow improvements, and
development/testing plan toward "Typst fully working for interior book page
design and production" — staged as: ship Ghosts → reliable single-chapter SPA
pipeline → multi-chapter/anthology support.

This file is GitHub-issues-flavored markdown. Each entry is one section.
Edit only from a clean working tree, push before someone else pulls.

---

## Conventions

- **IDs** are `TRK-AREA-NNN`. Areas:
  - `MIG` migration (prodcal + book-prod → jdbbs)
  - `OPS` operations (VM, systemd, nginx, backups, deploy)
  - `REV` code review (correctness, quality)
  - `SEC` security
  - `FLOW` workflow / dev ergonomics
  - `DEV` development tasks (features, refactors)
  - `TEST` testing tasks (infra, fixtures, suites)
  - `DESIGN` design parity & typography (Ghosts golden)
- **Status:** `open` | `in-progress` | `blocked` | `done` | `dropped`
- **Priority:** `P0` (production-blocking) | `P1` (must-do for cutover/ship) | `P2` (should-do) | `P3` (nice-to-have)
- **Refs:** repo paths, commit SHAs, doc anchors, URLs
- **Update protocol:** when changing status, also bump `updated:` and append a dated note at the bottom of the entry.
- **Splitting trigger:** when this file passes ~100 entries or a section passes ~30 active items, split into `planning/{migration,ops,review,security,workflow,dev,test,design}.md`.

---

## Strategy (revised 2026-05-12)

**Direction:** reconcile in `prodcal` (the live, substantive repo), then migrate to `jdbbs` only at the very end as a rename / final push.

**Why the reversal:** initial plan assumed jdbbs was the truth and book-prod + prodcal were stale upstreams. Investigation showed the opposite — the live VM `prodcal` has **~30 commits and significant uncommitted work** that's not in jdbbs (security hardening, preflight system, email notifications, archive lifecycle, custom-style hardening, API docs). jdbbs has only ~3 substantive commits worth keeping (typesetting subdir import, Phase 2 path fixes, Phase 3.1 trim registry). It's strictly fewer ports going **jdbbs → prodcal** than the other way.

**Phases:**

1. **Stabilize source of truth (this week)**
   - Push VM HEADs to GitHub: `/home/exedev/prodcal/` → `djinna/prodcal`, `/home/exedev/book-production/` → `djinna/book-prod`. Backup + reachability.
   - Commit & push the uncommitted VM work (modified Lua / Python / Typst, untracked notes, preflight test suite).
2. **Reconcile in `prodcal`**
   - Port jdbbs's typesetting subdir import (`2127bca`) — bring `typesetting/` into prodcal.
   - Port jdbbs's Phase 2 path fixes (`ddccd22`) — `typesettingRoot()` resolver, Dockerfile, Makefile, CI.
   - Port jdbbs's Phase 3.1 trim registry (`488c01e`) — protocolized preset etc.
   - Land the 3 modified book-production files + the preflight test suite into the unified tree.
3. **Cutover (drop the orphan, refresh systemd to point at the unified prodcal binary, point `db.sqlite3` correctly).**
4. **Rename to jdbbs at the end** — either GitHub rename `prodcal` → `jdbbs` (preserves history), or final force-push into `djinna/jdbbs` with prodcal then archived. Decide at the end.

**This tracker still lives at `jdbbs/TRACKER.md`** for now — it'll get committed into the final repo regardless of name. If reconciliation work needs a tracker visible to the prodcal codebase during reconciliation, mirror as needed.

---

## Table of Contents

- [Migration (MIG)](#migration-mig)
  - [TRK-MIG-001 — Push VM HEADs to GitHub (prodcal + book-prod)](#trk-mig-001--push-vm-heads-to-github-prodcal--book-prod)
  - [TRK-MIG-002 — Commit uncommitted VM work into git](#trk-mig-002--commit-uncommitted-vm-work-into-git)
  - [TRK-MIG-003 — Port jdbbs deltas into prodcal (3 substantive commits)](#trk-mig-003--port-jdbbs-deltas-into-prodcal-3-substantive-commits)
  - [TRK-MIG-004 — Cutover: refresh systemd to unified prodcal, drop orphan](#trk-mig-004--cutover-refresh-systemd-to-unified-prodcal-drop-orphan)
  - [TRK-MIG-009 — Final rename / move prodcal → jdbbs](#trk-mig-009--final-rename--move-prodcal--jdbbs)
  - [TRK-MIG-005 — Decide EPUB strategy (Go handler vs shell script)](#trk-mig-005--decide-epub-strategy-go-handler-vs-shell-script)
  - [TRK-MIG-006 — Wire corrections pipeline (SQLite → YAML → patchers)](#trk-mig-006--wire-corrections-pipeline-sqlite--yaml--patchers)
  - [TRK-MIG-007 — Verify Libertinus Serif on VM (or bundle)](#trk-mig-007--verify-libertinus-serif-on-vm-or-bundle)
  - [TRK-MIG-008 — Fix scripts/backup-db.sh path](#trk-mig-008--fix-scriptsbackup-db-sh-path)
- [Ops (OPS)](#ops-ops)
  - [TRK-OPS-001 — Stale prodcal process holding :8000; systemd in restart loop](#trk-ops-001--stale-prodcal-process-holding-8000-systemd-in-restart-loop)
  - [TRK-OPS-002 — Verify TLS termination and reverse proxy path](#trk-ops-002--verify-tls-termination-and-reverse-proxy-path)
  - [TRK-OPS-003 — SQLite WAL hygiene + backup verification](#trk-ops-003--sqlite-wal-hygiene--backup-verification)
  - [TRK-OPS-004 — `.env` on disk in /home/exedev/prodcal/](#trk-ops-004--env-on-disk-in-homeexedevprodcal)
- [Security (SEC)](#security-sec)
  - [TRK-SEC-008 — Resolve 6 Dependabot vulnerabilities on `djinna/prodcal`](#trk-sec-008--resolve-6-dependabot-vulnerabilities-on-djinnaprodcal)
  - [TRK-SEC-006 — Port bcrypt + auth hardening from prodcal `3c2256d`](#trk-sec-006--port-bcrypt--auth-hardening-from-prodcal-3c2256d)
  - [TRK-SEC-007 — Read `docs/API.md` — known security issues already flagged](#trk-sec-007--read-docsapimd--known-security-issues-already-flagged)
  - [TRK-SEC-001 — Admin SPA auth model audit](#trk-sec-001--admin-spa-auth-model-audit)
  - [TRK-SEC-002 — Public binding on :8000 and proxy bypass risk](#trk-sec-002--public-binding-on-8000-and-proxy-bypass-risk)
  - [TRK-SEC-003 — jdbbs repo is public — secret scan](#trk-sec-003--jdbbs-repo-is-public--secret-scan)
  - [TRK-SEC-004 — File upload safety audit (DOCX / YAML / images)](#trk-sec-004--file-upload-safety-audit-docx--yaml--images)
  - [TRK-SEC-005 — Command-injection audit in shell pipeline scripts](#trk-sec-005--command-injection-audit-in-shell-pipeline-scripts)
- [Workflow (FLOW)](#workflow-flow)
  - [TRK-FLOW-001 — Session-prompt proliferation hygiene](#trk-flow-001--session-prompt-proliferation-hygiene)
  - [TRK-FLOW-002 — Verify CI on jdbbs is running and useful](#trk-flow-002--verify-ci-on-jdbbs-is-running-and-useful)
  - [TRK-FLOW-003 — Pre-commit secret scanning](#trk-flow-003--pre-commit-secret-scanning)
- [Review (REV)](#review-rev)
  - [TRK-REV-001 — `prodcal` binary committed to djinna/prodcal repo (16.5 MB)](#trk-rev-001--prodcal-binary-committed-to-djinnaprodcal-repo-165-mb)
- [Design / Typography (DESIGN)](#design--typography-design)
  - [TRK-DESIGN-001 — Ghosts InDesign → Typst parity matrix](#trk-design-001--ghosts-indesign--typst-parity-matrix)
  - [TRK-DESIGN-002 — Commercial font licensing & bundling](#trk-design-002--commercial-font-licensing--bundling)
- [Dev (DEV)](#dev-dev)
  - [TRK-DEV-001 — `series-template.typ` config override mechanism (consolidate)](#trk-dev-001--series-templatetyp-config-override-mechanism-consolidate)
- [Test (TEST)](#test-test)
  - [TRK-TEST-001 — End-to-end fixture pipeline (DOCX/MD → PDF + EPUB)](#trk-test-001--end-to-end-fixture-pipeline-docxmd--pdf--epub)
  - [TRK-TEST-002 — Visual regression for Ghosts golden](#trk-test-002--visual-regression-for-ghosts-golden)
  - [TRK-TEST-003 — VM smoke script + cron](#trk-test-003--vm-smoke-script--cron)

---

## Migration (MIG)

### TRK-MIG-001 — Push VM HEADs to GitHub (prodcal + book-prod)

- area: MIG
- status: done
- priority: P1
- created: 2026-05-12
- updated: 2026-05-12
- refs: backup tags `backup/origin-main-pre-reconcile-2026-05-12` on both repos
- blocks: TRK-MIG-003

**Done 2026-05-12.** Recap:

- **prodcal**: VM main (`d0abea0`) force-with-lease pushed to `djinna/prodcal` main. Old origin/main (`3a8df8a`, the jdbbs kickoff merge) preserved at backup tag. 7 commits of substantive VM work now reachable.
- **book-production**: prep branch `prep/test-run-2026-04-05` updated on origin (6 new commits). origin/main left at `a595088`; reconciliation between prep branch and main is a separate task (note as follow-up below).
- Backup tags both pushed.

**Follow-up sub-items (open, low priority):**

1. Decide what to do with book-prod `prep/test-run-2026-04-05` long-term — merge into main (PR), keep as a parking branch, or rebase onto main. Either way, this is now a problem confined to the upstream that will be archived after TRK-MIG-009.
2. Local clones now reflect VM HEADs:
   - `/Users/jd2025/jd-projects/prodcal` HEAD `d0abea0`
   - `/Users/jd2025/jd-projects/book-prod origin/prep/test-run-2026-04-05` HEAD `0e6e51e`

### TRK-MIG-002 — Commit uncommitted VM work into git

- area: MIG
- status: done
- priority: P1
- created: 2026-05-12
- updated: 2026-05-12
- blocks: TRK-MIG-003

**Done 2026-05-12.** All previously-uncommitted VM work captured and pushed.

**book-production** (7 new commits on `origin/prep/test-run-2026-04-05`):

1. `0d6e46b` lua: tweet-p / metadata-p / metadata-c custom style mappings + content preservation fix
2. `0530215` md-to-chapter: footnotes, custom-style passthrough, URL/escape fixes (+355/-275)
3. `f80dd04` series-template: heading-align / spacing / footnote refinements (+68/-19, new config keys)
4. `7dc25b6` gitignore: agent state dirs, python caches, test-venv, tmp files
5. `18f3488` docs/analysis: 6 context docs relocated under docs/analysis/ (HANDLER_SUMMARY_prodcal-srv, HANDLER_SUMMARY_books-bookspecs-epub, PRODCAL_ANALYSIS, PROJECT_SUMMARY, SCRIPT_ANALYSIS, TEMPLATE_ANALYSIS)
6. `9b41854` test-cases: preflight detector fixtures (declared-styles.json, test.docx, test-report.{html,json}) + sample-chapter.{md,pdf,typ} as minimal E2E fixture
7. `e0e39e1` src: Twitter Years Typst template + edge-cases test fixture

**prodcal** (3 new commits on `origin/main`):

1. `07d06a0` docs/notes: 3 engineering retros (reorientation, special-typography-preflight, typst-pipeline-rewire-and-testing-retro)
2. `4d102c7` docs/analysis: TECHNICAL-ANALYSIS.md + HANDLER_SUMMARY_srv.md relocated under docs/analysis/
3. `0d1d84f` manuscripts/twitter-years: relocate The-Twitter-Years-{0237,0240}.docx + 200722-template.{docx,epub} from root/docs/ into the canonical manuscripts/ home

Both VM working trees now clean. Local clones synced.

### TRK-MIG-003 — Port jdbbs deltas into prodcal (3 substantive commits)

- area: MIG
- status: done
- priority: P1
- created: 2026-05-12
- updated: 2026-05-12
- refs: jdbbs `2127bca`, `ddccd22`, `488c01e`; prodcal `unify-typesetting` branch merged as `88e5283` on `main`
- blocked-by: TRK-MIG-001, TRK-MIG-002

**Done 2026-05-12.** 8 commits on `unify-typesetting`, merged --no-ff to `main`, pushed.

**Source content (5 commits):**
- `01941af` Import typesetting/ subdir (92 files, +6992 LOC) — templates / scripts / filters / fonts / test / test-fixtures. Sourced from `djinna/book-prod` prep@`e0e39e1`; apply-corrections{,-docx}.py grafted from main@`a595088` (prep forks before corrections system landed).
- `fbe897a` Import manuscripts/ (34 files, +10390 LOC) — ghosts/, samples/, and the Twitter Years template.
- `ab63936` Import reference/ (249 files, +86751 LOC) — GHOSTS.pdf, TT.pdf, LIBRARIANS.pdf, EPUBs, extracted internals.
- `c69bcb2` Import corrections/ — example-ghosts.yaml from book-prod main.
- `bca79fb` Import docs/typesetting/ — typography/workflow/author-prep refs filed under docs/typesetting/ to avoid name collisions with prodcal's existing docs/.

**Code ports (3 commits):**
- `48b3d07` Phase 2 path resolver + Phase 3.1 trim registry in Go (srv/books.go, srv/bookspecs.go, srv/preflight.go).
  - `typesettingRoot()` resolver replaces 4 hardcoded `bookProdRoot=/home/exedev/book-production` constants.
  - Resolution order: `JDBBS_TYPESETTING_DIR` env → `./typesetting` → walk parents → fallback.
  - Helpers: `typstFilterPath()` (points to `typesetting/filters/docx-to-typst-enhanced.lua` in new layout), `seriesTemplatePath()`, `fontsDirPath()`.
  - `trimRegistry` map with `protocolized` publisher preset (124.8×192.8mm from reference PDF measurement); `defaultSpecData` switches to protocolized.
- `704c904` admin.html UI mirror — Publisher Presets optgroup, `trimDisplayNames`, `fmtPicas` + `trimTooltip` helpers, comparison strip slot for Protocolized, page-preview mm display + hover tooltip.
- `3fcaaa1` build-ghosts.sh + srv.service path fixes — script-relative `REPO_ROOT`, `src/ghosts/`→`manuscripts/ghosts/`, srv.service `WorkingDirectory=/home/exedev/prodcal` + explicit `JDBBS_TYPESETTING_DIR`.

**Verification:** `go vet ./...` clean, `go build ./...` clean, `go test ./srv/...` ok (3.778s).

**Deferred to follow-up tasks:**
- Dockerfile multi-stage (Typst+Pandoc+Python+Libertinus) — not needed for VM systemd deploy; deferred.
- `.github/workflows/ci.yml` — deferred to TRK-FLOW-001.
- Makefile additions (typeset-deps target) — deferred.

### TRK-MIG-004 — Cutover: refresh systemd to unified prodcal, drop orphan

- area: MIG
- status: done
- priority: P1
- created: 2026-05-12
- updated: 2026-05-12
- refs: VM `pid=2597479` orphan (now dead); `/etc/systemd/system/prodcal.service` refreshed; new main PID 3266422
- blocks: TRK-OPS-001
- blocked-by: TRK-MIG-003

**Done 2026-05-12.** Clean cutover. Total downtime ~3 seconds (the time between SIGTERM and the new bind).

Steps executed:
1. `git fetch + git checkout main + git pull --ff-only` on VM — 9 commits fast-forwarded to `88e5283`.
2. `go build -o prodcal ./cmd/srv` — 17M binary, fresh mtime, clean compile.
3. `sudo cp srv.service /etc/systemd/system/prodcal.service` — new unit with `WorkingDirectory=/home/exedev/prodcal`, `Environment=JDBBS_TYPESETTING_DIR=/home/exedev/prodcal/typesetting`, `After=network.target`.
4. `sudo systemctl daemon-reload`.
5. `sudo systemctl stop prodcal.service` — the failing-restart loop went inactive.
6. `sudo kill 2597479` — orphan responded to SIGTERM; `ps -p 2597479` confirmed dead. (Had been holding `:8000` since 2026-05-06.)
7. `ss -ltnp 'sport = :8000'` — empty.
8. `sudo systemctl start prodcal.service` — main PID 3266422, status `active (running)`, listening on `:8000`.
9. Smoke tests: `curl http://localhost:8000/` → HTTP 200, 14397b in 878 microseconds. `/api/v1/projects` → HTTP 404 (auth-gated, as expected).
10. Journal confirms clean startup: 14 DB migrations applied (007→014), base URL configured, email configured, server listening.

Risk tolerance per user (minutes of downtime acceptable) easily met. No data loss; SQLite DB intact at 122MB.

**Follow-up handled separately:**
- TRK-MIG-008 (backup-db.sh path) — still open.
- TRK-OPS-003 (WAL checkpoint, 31MB uncheckpointed) — still open.
- TRK-SEC-008 (6 Dependabot vulnerabilities, 2 critical) — newly visible after first push; still open.

### TRK-MIG-009 — Final rename / move prodcal → jdbbs

- area: MIG
- status: blocked
- priority: P3
- created: 2026-05-12
- updated: 2026-05-12
- blocked-by: TRK-MIG-004; plus 1–2 weeks of stable post-cutover operation

Once the unified prodcal is stable and feature-complete in production, decide on the final home:

- **Option A (preserve history):** GitHub Settings → rename `djinna/prodcal` → `djinna/jdbbs`. Archive the current `djinna/jdbbs` (push a tag of its current state first). Update VM remotes and systemd paths.
- **Option B (clean cutover):** Force-push prodcal's final state into `djinna/jdbbs`, archive `djinna/prodcal`. Loses prodcal git history (or preserve it in a tag).
- **Option C (defer):** keep working under the prodcal name indefinitely, archive jdbbs as a stale draft, accept that the public name is `prodcal` and the URL stays `jdbbs.exe.xyz`.

Recommendation: pick after we see what reconciliation feels like. Don't decide today.

### TRK-MIG-005 — Decide EPUB strategy (Go handler vs shell script)

- area: MIG
- status: open
- priority: P2
- created: 2026-05-12
- updated: 2026-05-12
- refs: jdbbs/srv/epub.go, jdbbs/typesetting/scripts/{docx2epub,md2epub}.sh; MIGRATION_LOG.md §"Open Phase 3" item 2

Two divergent EPUB paths produce different CSS / font handling. **Action:** pick one source of truth — recommend Go handler shells out to script(s) so CLI and web app share output. Capture decision in DECISIONS section once made.

### TRK-MIG-006 — Wire corrections pipeline (SQLite → YAML → patchers)

- area: MIG
- status: open
- priority: P2
- created: 2026-05-12
- updated: 2026-05-12
- refs: jdbbs/srv/corrections.go, jdbbs/typesetting/scripts/apply-corrections{,-docx}.py; MIGRATION_LOG.md §"Open Phase 3" item 3

Currently: corrections stored in SQLite + manually exported as YAML + manually fed to `apply-corrections.py`. **Action:** after EPUB/DOCX generation in `srv/`, materialize YAML in-memory and invoke the corresponding patcher.

### TRK-MIG-007 — Verify Libertinus Serif on VM (or bundle)

- area: MIG
- status: open
- priority: P1
- created: 2026-05-12
- updated: 2026-05-12
- refs: jdbbs/typesetting/fonts/, jdbbs/Dockerfile; MIGRATION_LOG.md §"Open Phase 3" item 4
- blocks: TRK-MIG-003 (only if absent)

Run `fc-list | grep -i libertinus` on VM. If absent: `apt install fonts-libertinus`, or bundle in `typesetting/fonts/libertinus/` for full self-containment (preferred — removes a runtime dependency).

### TRK-MIG-008 — Fix scripts/backup-db.sh path

- area: MIG
- status: open
- priority: P1
- created: 2026-05-12
- updated: 2026-05-12
- refs: jdbbs/scripts/backup-db.sh; MIGRATION_LOG.md §"Open Phase 3" item 5

Script still references `/home/exedev/prodcal/db.sqlite3`. Update to `/home/exedev/jdbbs/db.sqlite3`. Also verify cron is wired and `~/backups/` is being populated (it exists on VM; check freshness).

---

## Ops (OPS)

### TRK-OPS-001 — Stale prodcal process holding :8000; systemd in restart loop

- area: OPS
- status: open
- priority: P1
- created: 2026-05-12
- updated: 2026-05-12
- refs: VM `pid=2597479` (alive 6d 3h), `systemctl status prodcal`, `journalctl -u prodcal`
- resolution-via: TRK-MIG-003

The systemd unit `prodcal.service` is failing every 5 seconds with `bind: address already in use`. An orphan `/home/exedev/prodcal/prodcal` (PID 2597479) detached from systemd is still serving the SPA. Effect: no deploy attempt since the loop began has taken effect; the live binary is stale.

Resolution = cutover (TRK-MIG-003). Not fixing in place because we're replacing the service.

### TRK-OPS-002 — Verify TLS termination and reverse proxy path

- area: OPS
- status: open
- priority: P2
- created: 2026-05-12
- updated: 2026-05-12
- refs: VM `/etc/nginx/sites-enabled/default`, exe.dev HTTPS proxy

nginx on the VM has only a generic `_` server block — it doesn't terminate TLS for `jdbbs.exe.xyz`. Per jdbbs/README.md, the exe.dev HTTPS proxy handles TLS and forwards to `localhost:8000`. **Action:** confirm exe.dev proxy config (where is it managed?), check cert renewal path, document recovery story if exe.dev's proxy fails or the DNS is hijacked.

### TRK-OPS-003 — SQLite WAL hygiene + backup verification

- area: OPS
- status: done
- priority: P1
- created: 2026-05-12
- updated: 2026-05-12
- refs: live DB `/home/exedev/prodcal/db.sqlite3` (117 MB), backup `/home/exedev/backups/prodcal-20260512-171252.sqlite3.gz`, archive `/home/exedev/.archive-old-db-pre-recovery/`, defensive copy `/home/exedev/db-recovery-20260512T171017Z/`

**Done 2026-05-12.** What started as a WAL/backup hygiene task surfaced a critical recovery situation that had to be resolved end-to-end.

**Pre-state (root-cause discovery):**
- Live DB lived at `/home/exedev/db.sqlite3` (117 MB main + 30 MB WAL — uncheckpointed since at least May 5).
- The May 12 cutover (TRK-MIG-004) changed systemd `WorkingDirectory` from `/home/exedev` to `/home/exedev/prodcal`. The binary opens `db.sqlite3` relative-path, so it transparently switched to `/home/exedev/prodcal/db.sqlite3` — an 80 KB **empty Feb-25 test DB** that happened to be sitting in the prodcal repo dir.
- Live service ran against an empty DB for ~30 minutes (16:39 → 17:10). Users hitting `https://jdbbs.exe.xyz` saw an empty book list.
- The daily backup script (`scripts/backup-db.sh`) had hardcoded `DB_PATH="/home/exedev/prodcal/db.sqlite3"` from a prior layout assumption. So **every backup in `~/backups/` from May 5 onward was 3.5 KB** — the script had been backing up an empty DB silently for a week.

**Recovery executed:**
1. Defensive copies of BOTH DBs (`/home/exedev/db.sqlite3*` and the empty `/home/exedev/prodcal/db.sqlite3*`) into `/home/exedev/db-recovery-20260512T171017Z/`.
2. `sudo systemctl stop prodcal`.
3. `mv /home/exedev/prodcal/db.sqlite3{,.feb25-empty.bak}` (set the empty DB aside).
4. `sqlite3 /home/exedev/db.sqlite3 ".backup '/home/exedev/prodcal/db.sqlite3'"` — atomic, consolidates the 30 MB WAL into a single 117 MB file at the new location.
5. `PRAGMA integrity_check;` → `ok`.
6. `sudo systemctl start prodcal` — service active, MainPID 3270316.
   - **Orphan-process pattern recurred** during this restart (see new TRK-OPS-005 below); resolved inline by killing the orphan PID so systemd's MainPID could take over cleanly.
7. Verified counts: projects=12, books=6, book_specs=5, corrections=1, book_outputs=6, manuscript_preflights=12, transmittals=9.
8. HTTP smoke: GET / → 200 in ~400 µs. Twitter Years (project #7) data fully restored.
9. Backup script: no edit needed (its hardcoded path now happens to be correct for the new layout). Manual run produced `prodcal-20260512-171252.sqlite3.gz` with verified real data inside (projects=12, books=6 inside the gzipped backup).
10. Old DB archived to `/home/exedev/.archive-old-db-pre-recovery/` so future operators can't get confused.

**Follow-ups still open:**
- TRK-OPS-005 — systemd orphan-process race during restart (see below).
- TRK-OPS-006 — DB cleanup: 12 of 13 projects are tests/dummies; only #7 "The Twitter Years: 2007–22" is canonical.
- TRK-OPS-007 — Hardening: backup integrity verification (size threshold, rowcount probe), off-VM replication. Consider `litestream` for continuous replication.

### TRK-OPS-005 — systemd orphan-process race during restart

- area: OPS
- status: done
- priority: P2
- created: 2026-05-12
- updated: 2026-05-12
- refs: prodcal `6ff2c7e9` (fix), TRK-MIG-004 / TRK-OPS-003 (triggers)

**Root cause (found 2026-05-12):**
- `cmd/srv/main.go` printed errors to stderr and **returned normally — exit 0**. A `bind: address already in use` looked to systemd like `status=0/SUCCESS`.
- Combined with `Restart=always`, every "successful exit" triggered a 5-second-later restart. If the previous instance's listener was still alive, the new instance would fail to bind, exit 0 again — looking healthy from systemd's perspective.
- `Type=simple` had no positive ready signal, so when the binary DID bind successfully, there was a brief window where systemd could misregister MainPID as 0 and trigger an unnecessary auto-restart anyway.

**Fix (commit 6ff2c7e9):**
1. `cmd/srv/main.go`: `run()` error → `os.Exit(1)` (was: print and exit 0).
2. `srv.Server.Serve()`: split `http.ListenAndServe` into explicit `net.Listen` first, then `daemon.SdNotify(SdNotifyReady)`, then `http.Serve(listener, handler)`. Adds `github.com/coreos/go-systemd/v22 v22.7.0` (pure Go, MIT).
3. `srv.service`: `Type=simple` → `Type=notify` with `NotifyAccess=main`; `Restart=always` → `Restart=on-failure`.

**Verification (2026-05-12 17:18 UTC):**
- 5 rapid `sudo systemctl restart prodcal.service` cycles in a row. Every cycle:
  - service active
  - `MainPID == listener PID` (no orphans)
  - HTTP 200
- Journal on every restart: `INFO listener bound addr=:8000 pid=…` → `INFO sd_notify ready sent to systemd` → `INFO starting server`.
- Zero `bind: address already in use` after the fix landed.
- Data integrity preserved across all 5 cycles (projects=12, books=6 — Twitter Years).

### TRK-OPS-006 — Clean up 12 test/dummy projects in production DB

- area: OPS
- status: open
- priority: P2
- created: 2026-05-12
- updated: 2026-05-12
- refs: live DB `/home/exedev/prodcal/db.sqlite3`

Only project #7 "The Twitter Years: 2007–22" is canonical (per user 2026-05-12). The other 12 — Art of Gig, tweetbook, The Digital Garden, test, Smoke Test, 9apr26 test, test (again), Admin 9 Apr Seed Check, test 3, test 01, test 02 — are leftover test/dummy projects.

**Action:**
1. Pre-cleanup snapshot: `sqlite3 db.sqlite3 ".backup 'backups/pre-cleanup-$(date +%Y%m%dT%H%M%SZ).sqlite3'"`.
2. In one transaction: `DELETE FROM projects WHERE id != 7`. Verify any cascading deletes (book_specs.project_id, books.project_id, manuscript_preflights.project_id, transmittals.project_id) are wired correctly via FK ON DELETE or need explicit DELETEs.
3. Verify project #7's books survive (should — only that project has books) and that the SPA renders correctly with one project visible.
4. Reclaim space with `VACUUM` if the deletion freed substantial DB rows.

Risk: low. Reversible from backups if needed.

### TRK-OPS-007 — Backup hygiene: integrity probes + off-VM replication

- area: OPS / SEC
- status: done (phase 1 + 2); phase 3 (monitoring + restore drill) open
- priority: P2
- created: 2026-05-12
- updated: 2026-05-12
- refs: prodcal `07b7f910` (script hardened), `scripts/backup-db.sh`, `~/backups/{.LAST-SUCCESS,.LAST-FAILURE}`

**Phase 1 — in-script probes (done 2026-05-12).** Local hardening complete; off-VM replication still open.

**Script (`scripts/backup-db.sh`) now does:**
- single-instance lock via `flock`; concurrent run aborts with a clear message
- preflight on source DB (exists, non-empty)
- post-gzip **size probe** — fail if backup < `$MIN_GZ_BYTES` (default 1 MB); catches the exact bug we hit on 2026-05-12 where 3.5 KB "backups" landed for a week
- decompress to a tmp file, **rowcount probe** on `projects` + `books`; fail if either below `$MIN_PROJECTS` / `$MIN_BOOKS`
- `PRAGMA integrity_check`; fail if not `ok`
- **drift comparison** vs previous backup; warn (not fail) if size delta > `$DRIFT_WARN_PCT` (default 50%)
- **retention tiers**: daily backups for `$RETAIN_DAILY_DAYS` (default 30 days), plus the first-of-month for each calendar month kept indefinitely
- **success sentinel**: `~/backups/.LAST-SUCCESS` with size + rowcounts + timestamp; cleared `.LAST-FAILURE` on success
- **failure sentinel**: `~/backups/.LAST-FAILURE` with reason, timestamp, paths

**Verified 2026-05-12 17:21 UTC** with 4 tests on VM:
| # | Scenario | Result |
|---|---|---|
| 1 | Real 117 MB DB → 109 MB gz; expect OK | OK; projects=12, books=6, integrity=ok, `.LAST-SUCCESS` written |
| 2 | Point at empty Feb-25 backup (80 KB → 3.5 KB gz); expect size fail | FAIL: `3522B < 1048576B`; `.LAST-FAILURE` written |
| 3 | Real DB but `MIN_PROJECTS=99999`; expect rowcount fail | FAIL: `12 < 99999`; `.LAST-FAILURE` written |
| 4 | Run while another instance holds the lock; expect concurrent abort | FAIL: `another backup is already running (lock: …)` |

**Phase 2 — off-VM replication (done 2026-05-12).**

Approach: **rclone copy to Cloudflare R2** for daily off-VM, plus a Mac-side `rsync` alias for an air-gapped on-demand mirror. Together with the VM-local backup, this hits the 3-2-1 rule:

| Copy | Location | Cadence |
|---|---|---|
| 1 (live) | `/home/exedev/prodcal/db.sqlite3` on jdbbs.exe.xyz | continuous |
| 2 (local backup) | `~/backups/prodcal-*.sqlite3.gz` on jdbbs.exe.xyz | daily 03:00 UTC + on-demand |
| 3a (off-VM) | `r2:jdbbs-backups/db/` on Cloudflare R2 | daily 03:00 UTC + on-demand |
| 3b (off-VM) | `~/backups-jdbbs/` on your Mac | on-demand via `jbackup-pull` |

**Cloudflare R2 setup:**
- bucket: `jdbbs-backups` (created via R2 dashboard)
- prefix: `db/`
- API token: `Object Read & Write` scoped to `jdbbs-backups`, configured locally as rclone remote `r2`
- script: prodcal `scripts/sync-to-r2.sh` — flock, rclone copy (not sync, so deletes aren't propagated), post-copy size verification, separate `.LAST-R2-SUCCESS` / `.LAST-R2-FAILURE` sentinels (so a partial — local OK + R2 push failed — is visible)
- cron: same line as backup-db.sh, joined with `;` so a failed local backup still attempts to push the most recent file off-VM (better to ship yesterday's backup than nothing)

**Mac-side mirror (`~/.zshrc::jbackup-pull`):**
- rsync with `--delete --delete-excluded` and a filter for `prodcal-*.sqlite3.gz` + sentinels
- automatically runs a local sqlite probe (projects + books + integrity_check) on the newest pulled backup using `file:?immutable=1` URI so the WAL/shm sidecar dance doesn't error
- usage: `jbackup-pull` (defaults to `~/backups-jdbbs/`)

**Verified end-to-end 2026-05-12 18:38 UTC:**
- backup-db.sh produced a fresh 109 MB local backup with sentinel
- sync-to-r2.sh pushed it; remote size matched local size (113,752,384 B)
- 3 newest backups confirmed in `r2:jdbbs-backups/db/`
- All four sentinels in the expected state (.LAST-SUCCESS present, .LAST-R2-SUCCESS present, both .LAST-FAILURE absent)

**Still open (phase 3):**
- Monitoring: cron job (or exe.dev observability) that checks `~/backups/.LAST-FAILURE` and `.LAST-R2-FAILURE`, plus age of the `.LAST-SUCCESS` sentinels (alarm if > 26 hours old).
- Monthly restore drill: separate cron that decompresses the latest backup into `/tmp/restore-test.sqlite3`, runs a battery of read queries, asserts row counts.
- 8 small (3.5 KB) artifact backups from the pre-fix era now live in R2 indefinitely; cull them via `rclone deletefile` once we're sure they're not needed.

**RPO/RTO targets:**
- RPO: ≤ 24 hours (daily backups); upgrade to ≤ 1 hour later with litestream if we go heavier on usage.
- RTO: ≤ 30 minutes for a fresh VM (spin up + pull repo + `rclone copy r2:jdbbs-backups/db/<latest> .` + gunzip + start systemd).

### TRK-OPS-004 — `.env` on disk in /home/exedev/prodcal/

- area: OPS / SEC
- status: open
- priority: P2
- created: 2026-05-12
- updated: 2026-05-12
- refs: VM `/home/exedev/prodcal/.env` (152 bytes, mode 600)

Small `.env` file on disk. Mode 600 is correct; **action:** read contents during cutover prep, confirm what env vars are needed (likely `AGENTMAIL_API_KEY`, `AGENTMAIL_INBOX_ID`), migrate to `/home/exedev/jdbbs/.env` or systemd `EnvironmentFile=`. Confirm `.env` is in `.gitignore` (it is — confirmed in prodcal/.gitignore).

---

## Security (SEC)

### TRK-SEC-008 — Resolve 6 Dependabot vulnerabilities on `djinna/prodcal`

- area: SEC
- status: done
- priority: P1
- created: 2026-05-12
- updated: 2026-05-12
- refs: prodcal `7176e9a9` (go.mod/go.sum bump); GitHub auto-closed all 6 alerts at 2026-05-12T16:56:30Z
- pr/commit: https://github.com/djinna/prodcal/commit/7176e9a9

**Done 2026-05-12.** Patched, pushed, GitHub auto-detected fixes from go.mod, alerts state=fixed across the board. Patched binary deployed to VM (PID 3268939 active running).

| # | Severity | Package | Bump | GHSA |
|---|---|---|---|---|
| #4 | **critical** | `google.golang.org/grpc` | v1.75.0 → **v1.79.3** | GHSA-p77j-4mvh-x3m3 (authz bypass via missing leading slash in `:path`) |
| #5 | **critical** | `github.com/jackc/pgx/v5` | v5.7.5 → **v5.9.2** | GHSA-9jj7-4m8r-rfcm (memory-safety) |
| #1 | medium | `golang.org/x/crypto` | v0.39.0 → **v0.46.0** | GHSA-j5w8-q4qc-rx2x (ssh unbounded mem) |
| #2 | medium | `golang.org/x/crypto` | v0.39.0 → **v0.46.0** | GHSA-f6x5-jh6r-wrfv (ssh/agent OOB read panic) |
| #3 | low | `filippo.io/edwards25519` | v1.1.0 → **v1.1.1** | GHSA-fw7p-63qq-7hpr (MultiScalarMult identity-receiver) |
| #6 | low | `github.com/jackc/pgx/v5` | v5.7.5 → **v5.9.2** | GHSA-j88v-2chj-qfwx (SQL-injection via $-quoted placeholder) |

**Notes from the bump:**
- `x/crypto` was always a direct import (`srv/server.go` uses `golang.org/x/crypto/bcrypt`) but had been recorded as indirect; `go mod tidy` now promotes it to the direct require block, which is the correct shape.
- The grpc bump tripped a transient downgrade to `v1.79.0-dev` because `sqlc` transitively pinned that pre-release when x/crypto was bumped first. Re-asserting `grpc@v1.79.3` after the x/crypto bump resolves cleanly. Verified via `go list -m google.golang.org/grpc`.
- Knock-on: `x/net` v0.41.0→v0.48.0, `x/sync` v0.16.0→v0.19.0, `x/sys` v0.34.0→v0.39.0, `x/text` v0.26.0→v0.32.0, `protobuf` v1.36.8→v1.36.10, `genproto/{api,rpc}` 2025-07-07→2025-12-02, `cel.dev/expr` v0.24.0→v0.25.1. All from `go mod tidy`; no source code changes needed.
- Verified: `go vet ./...` clean, `go build ./...` clean, `go test ./srv/...` ok (3.7s local, 4.3s VM).
- Deploy: VM rebuilt, systemd restarted, HTTP 200 in 738µs post-restart.

### TRK-SEC-006 — Port bcrypt + auth hardening from prodcal `3c2256d`

- area: SEC / MIG
- status: open
- priority: P0
- created: 2026-05-12
- updated: 2026-05-12
- refs: prodcal VM commit `3c2256d security: bcrypt passwords, auth on downloads/covers/project-list, remove secrets from tracking`

VM commit `3c2256d` already did substantial security hardening — but **only in the VM's prodcal**, not in jdbbs, and likely not pushed to GitHub `djinna/prodcal` either. Contents:

1. SHA-256 → bcrypt (cost 12) for all password/token hashing; new `hashPassword()` / `checkPassword()` replace `hashToken()`.
2. Auth required on book download (`GET /api/books/{id}/download/{format}`) via new `GetBookProjectID` sqlc query.
3. Auth required on cover image (`GET /api/projects/{id}/book-spec/cover`).
4. Auth required on project list (`GET /api/projects`) via `requireExeDevAdminAPI()`.
5. Removed `SESSION-SUMMARY.txt` and `TEAM-UPDATE.txt` from git tracking; added `.hermes/` to `.gitignore`.

**Plus that same commit carried over** the preflight system, book output history, admin dashboard updates, custom style presets, and typography refinements.

**Action:** ensure `3c2256d` lands in the unified prodcal (TRK-MIG-001 push will get it onto GitHub; TRK-MIG-003 keeps it). Then verify each hardening point still works end-to-end after the typesetting subdir is added.

### TRK-SEC-007 — Read `docs/API.md` — known security issues already flagged

- area: SEC
- status: open
- priority: P1
- created: 2026-05-12
- updated: 2026-05-12
- refs: prodcal VM `docs/API.md` (1264 LOC, commit `a3a7845`)

A comprehensive API reference exists on the VM that "flags known security issues" per its commit message. Read it, capture each flagged issue as its own TRK-SEC-NNN entry, then triage by priority. This is probably the fastest way to seed the full security backlog without re-doing analysis.

### TRK-SEC-001 — Admin SPA auth model audit

- area: SEC
- status: open
- priority: P0
- created: 2026-05-12
- updated: 2026-05-12
- refs: jdbbs/srv/static/admin.html, jdbbs/srv/*.go handlers

`https://jdbbs.exe.xyz/admin/` is publicly reachable. Need to confirm: is there any auth gate (basic, OAuth, session)? Read `srv/server.go` (or wherever routes are mounted) for middleware. If no auth exists, this is the highest-priority security finding.

### TRK-SEC-002 — Public binding on :8000 and proxy bypass risk

- area: SEC
- status: open
- priority: P1
- created: 2026-05-12
- updated: 2026-05-12
- refs: VM `ss -tlnp` (`prodcal` listens `*:8000`), jdbbs srv.service flag `-listen :8000`

The Go server binds `0.0.0.0:8000`. If exe.dev's edge proxy is the auth/TLS gate, anyone who finds the VM's IP can bypass it. **Action:** bind `127.0.0.1:8000` instead; have exe.dev proxy connect over the private network or via SSH-port-forward. Or: put auth in the Go app (TRK-SEC-001) so bypass doesn't matter.

### TRK-SEC-003 — jdbbs repo is public — secret scan

- area: SEC
- status: open
- priority: P0
- created: 2026-05-12
- updated: 2026-05-12
- refs: github.com/djinna/jdbbs (public)

Run `gitleaks detect` and `trufflehog filesystem` against the jdbbs working tree and full git history. Anything found = rotate + force-push history rewrite. (`prodcal` is private; lower urgency there but still scan.)

### TRK-SEC-004 — File upload safety audit (DOCX / YAML / images)

- area: SEC
- status: open
- priority: P1
- created: 2026-05-12
- updated: 2026-05-12
- refs: jdbbs/srv/books.go, jdbbs/srv/bookspecs.go, jdbbs/srv/corrections.go

User-supplied DOCX → Pandoc → Lua → Typst → PDF. User-supplied YAML → Python → python-docx → DOCX. Plus image uploads. Per-handler audit: size limits, MIME sniffing, path traversal in saved filenames, sandbox for shell-outs (pandoc/typst/python don't sandbox themselves).

### TRK-SEC-005 — Command-injection audit in shell pipeline scripts

- area: SEC
- status: open
- priority: P1
- created: 2026-05-12
- updated: 2026-05-12
- refs: jdbbs/typesetting/scripts/*.sh

`docx2pdf.sh`, `docx2epub.sh`, `md2pdf.sh`, `md2epub.sh`, `build.sh`, `build-ghosts.sh` — audit every variable expansion. Quote everything. Reject filenames containing `..`, `$`, backticks, newlines. Same audit for the Go `exec.Command` calls in `srv/`.

---

## Workflow (FLOW)

### TRK-FLOW-001 — Session-prompt proliferation hygiene

- area: FLOW
- status: open
- priority: P3
- created: 2026-05-12
- updated: 2026-05-12
- refs: VM `/home/exedev/book-production/NEXT_SESSION_PROMPT_*.md` (10+ files)

Migration convention dropped these from jdbbs; but they keep accumulating in the legacy `book-production/` clone on the VM. Once that clone is retired (TRK-MIG-003), the problem goes away. Add a convention in jdbbs/AGENTS.md: rolling `WORKLOG.md` or this TRACKER file replaces per-session prompt files.

### TRK-FLOW-002 — Verify CI on jdbbs is running and useful

- area: FLOW
- status: open
- priority: P2
- created: 2026-05-12
- updated: 2026-05-12
- refs: jdbbs/.github/workflows/ci.yml

Per MIGRATION_LOG, CI is Go 1.26 + vet + test + python-docx + pyyaml + pandoc. **Action:** verify last run is green; add `typst check` step for templates; add `shellcheck` for typesetting/scripts; add `ruff` for python.

### TRK-FLOW-003 — Pre-commit secret scanning

- area: FLOW / SEC
- status: open
- priority: P2
- created: 2026-05-12
- updated: 2026-05-12

Add `pre-commit` hook with `gitleaks` (or `trufflehog`) + `shellcheck` + `ruff` + `gofmt`. Cheap insurance; runs locally before push.

---

## Review (REV)

### TRK-REV-001 — `prodcal` binary committed to djinna/prodcal repo (16.5 MB)

- area: REV
- status: open
- priority: P3
- created: 2026-05-12
- updated: 2026-05-12
- refs: prodcal/prodcal (16.5 MB Linux/amd64 binary)

A built Go binary lives at the prodcal repo root. Tracked in git, in CI artifacts forever. Once prodcal is archived (TRK-MIG-004) this stops mattering; flagged for awareness. Don't repeat the pattern in jdbbs.

---

## Design / Typography (DESIGN)

### TRK-DESIGN-001 — Ghosts InDesign → Typst parity matrix

- area: DESIGN
- status: open
- priority: P1
- created: 2026-05-12
- updated: 2026-05-12
- refs: book-prod/reference/GHOSTS.pdf (136 pages, 353.811 × 546.567 pt = 4.91 × 7.59 in, PDF/X-4, InDesign 21.0)

The published Ghosts PDF is the golden parity target. Embedded fonts: Plantin MT Pro (R/I), Proxima Nova (B/SB/M), Menlo, HiraKakuPro-W3, Thonburi (last two for CJK/Thai glyphs).

**Feature matrix (to be filled):**

| Feature | InDesign | Typst today | Status | Notes |
|---|---|---|---|---|
| Trim 4.91 × 7.59 in | yes | yes | ✅ | added as `protocolized` preset in trim registry |
| Body font (Plantin MT Pro 10pt) | yes | open sub (Libertinus) | ⚠️ | license owned per user; bundle |
| Heading font (Proxima Nova) | yes | open sub (Source Sans 3) | ⚠️ | license owned per user; bundle |
| Mono (Menlo) | yes | open sub (JetBrains Mono) | ⚠️ | license unclear; decide |
| CJK / Thai glyphs (HiraKakuPro-W3, Thonburi) | yes | unknown | ❓ | which content uses them? |
| First-paragraph no-indent | yes | yes | ✅ | template feature |
| Justified + hyphenated | yes | yes | ✅ | |
| Running heads verso/recto | yes | yes | ✅ | configurable |
| Drop caps | ? | ? | ❓ | inspect InDesign PDF |
| Small caps | ? | ? | ❓ | inspect |
| Section break ornament | yes (breve?) | yes (configurable) | ✅ | confirm choice |
| ToC layout | yes | yes | ❓ | side-by-side comparison needed |
| Image placement / captions | yes (8 images) | yes (images.typ) | ❓ | side-by-side needed |
| Copyright page layout | yes | yes | ❓ | side-by-side needed |
| Footnotes style | ? | ? | ❓ | check if any in Ghosts |
| OpenType: oldstyle figures | ? | font-dependent | ❓ | needs font support |
| OpenType: ligatures | ? | font-dependent | ❓ | likely fine |
| Widow / orphan control | yes | partial (Typst-limited) | ⚠️ | likely cant-have |
| PDF/X-4 prepress output | yes | not directly (PDF/A possible) | ❌ | likely cant-have; or post-process via Ghostscript |
| ICC color profile embedded | yes | unknown | ❓ | needs check |

**Action:** spend a focused pass with InDesign PDF + a rendered Typst PDF side-by-side; fill remaining `❓` cells; produce a final cant-have list.

### TRK-DESIGN-002 — Commercial font licensing & bundling

- area: DESIGN
- status: open
- priority: P1
- created: 2026-05-12
- updated: 2026-05-12
- refs: TRK-DESIGN-001

User has licenses for Plantin MT Pro + Proxima Nova. **Action:**

1. Locate license documents.
2. Confirm permitted distribution mode (server-side embedded? per-output PDF embedded? bundled in repo?).
3. Add fonts to `jdbbs/typesetting/fonts/{plantin,proximanova}/` with a `LICENSE.txt` per family.
4. Update `series-template.typ` defaults to use licensed names when present, fall back to open subs otherwise.
5. Update `Dockerfile` to skip `fonts-libertinus` if Plantin is present.

---

## Dev (DEV)

### TRK-DEV-001 — `series-template.typ` config override mechanism (consolidate)

- area: DEV
- status: open
- priority: P2
- created: 2026-05-12
- updated: 2026-05-12
- refs: jdbbs/typesetting/templates/series-template.typ; TYPST_FRONTEND_PLAN.md (archived) §"Open Questions" #1

Three plausible patterns for spec → template config override (param to `#book()`, separate `config.typ`, shadowed `#let`). The trim-registry phase landed a partial pattern via `merge-config`. **Action:** read the current state, document the chosen pattern in `docs/TYPOGRAPHY.md`, kill any duplicate plumbing.

---

## Test (TEST)

### TRK-TEST-001 — End-to-end fixture pipeline (DOCX/MD → PDF + EPUB)

- area: TEST
- status: open
- priority: P1
- created: 2026-05-12
- updated: 2026-05-12

Smallest-possible fixture inputs (1-page DOCX, 1-page Markdown) in `test/fixtures/`. End-to-end:

1. Upload fixture via SPA → assert resulting PDF page count, fonts, trim.
2. Spec round-trip: save → reload → byte-diff generated `config.typ`.
3. Lua filter unit fixtures (pandoc + filter on tiny DOCX, assert output).
4. Python unit tests for `apply-corrections*.py`, `md-to-chapter.py`, `generate-word-template.py`.
5. Go table tests for `specToTypstConfig`, `parseTrim`, font picker.

### TRK-TEST-002 — Visual regression for Ghosts golden

- area: TEST
- status: open
- priority: P1
- created: 2026-05-12
- updated: 2026-05-12
- refs: TRK-DESIGN-001

Build Ghosts Typst PDF, render each page to PNG via `pdftoppm`, diff against pre-rendered InDesign pages (already in `reference/new_uploads/pdf_samples/ghosts-{010..012}.png` — extend the set), threshold tuned for justified-text noise. Tools: `diff-pdf` or ImageMagick `compare -metric`.

### TRK-TEST-003 — VM smoke script + cron

- area: TEST / OPS
- status: open
- priority: P2
- created: 2026-05-12
- updated: 2026-05-12

`scripts/smoke.sh` on VM: one DOCX→PDF + one MD→PDF + one MD→EPUB + one corrections apply; exits nonzero on failure. Cron daily, send email on failure (AgentMail is already wired).

---

## Decisions log

(Append-only. Each decision is dated, summarized, and refs the entries it locks down.)

- **2026-05-12** — Tracker lives in `jdbbs/TRACKER.md` (single source of truth), even before the VM cutover.
- **2026-05-12** — Commercial fonts (Plantin MT Pro, Proxima Nova) will be bundled in `typesetting/fonts/` with license docs (per user). Open substitutes remain as fallback.
- **2026-05-12** — **Strategy reversal**: reconcile in `prodcal` (the live, substantive repo), not in `jdbbs`. The VM's prodcal has ~30 commits + significant uncommitted work that's missing from jdbbs (including security hardening commit `3c2256d`, preflight system, email notifications, archive lifecycle, custom-style hardening, API docs). jdbbs has only 3 substantive commits worth porting (typesetting subdir import, Phase 2 path fixes, Phase 3.1 trim registry). Doing the reconciliation in prodcal is strictly fewer ports. Final rename `prodcal → jdbbs` is deferred (TRK-MIG-009).
- **2026-05-12** — Cutover style: "just do it" — minutes of downtime acceptable, deploy and fix-forward, no parallel staging.
