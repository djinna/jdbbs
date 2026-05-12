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
- status: open
- priority: P1
- created: 2026-05-12
- updated: 2026-05-12
- refs: VM `~/db.sqlite3` 122 MB, `~/db.sqlite3-wal` 31 MB, `~/backups/`

WAL has not been checkpointed for a while (31 MB unmerged). Check: `~/backups/` exists but content unknown — verify nightly backup is running, copies are recent and restorable. Define RPO/RTO. Consider `litestream` for continuous replication.

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
- status: open
- priority: P1
- created: 2026-05-12
- updated: 2026-05-12
- refs: https://github.com/djinna/prodcal/security/dependabot

Surfaced during the TRK-MIG-001 push. Six open alerts on prodcal default branch:

| # | Severity | Package | Summary |
|---|---|---|---|
| #4 | **critical** | `google.golang.org/grpc` | Authorization bypass via missing leading slash in `:path` |
| #5 | **critical** | `github.com/jackc/pgx/v5` | Memory-safety vulnerability |
| #1 | medium | `golang.org/x/crypto` | `ssh` unbounded memory consumption |
| #2 | medium | `golang.org/x/crypto` | `ssh/agent` panic on malformed message |
| #3 | low | `filippo.io/edwards25519` | `MultiScalarMult` invalid result when receiver is not identity |
| #6 | low | `github.com/jackc/pgx/v5` | SQL injection via dollar-quoted placeholder confusion |

**Triage notes:**

- pgx is likely a transitive dep (prodcal uses SQLite). Verify in `go.sum`; if unused, removing the transitive chain is the cleanest fix.
- grpc auth-bypass: also probably transitive (no obvious gRPC server). Verify and remove if unused; otherwise bump.
- `golang.org/x/crypto` — used by Go std-lib-adjacent code and possibly directly. Bump to a clean release.
- `edwards25519` — verify; likely transitive.

**Action:** `go list -m all | grep -E "pgx|grpc|crypto|edwards25519"` to identify reachable paths. Bump or remove. Run `go test ./...`. Confirm Dependabot closes alerts after push.

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
