# JDBBS Tracker

Single source of truth for the review, security/workflow improvements, and
development/testing plan toward "Typst fully working for interior book page
design and production" — staged as: ship Ghosts → reliable single-chapter SPA
pipeline → multi-chapter/anthology support.

This file is GitHub-issues-flavored markdown. Each entry is one section.
Edit only from a clean working tree, push before someone else pulls.

---

## 🟢 Resume here — next session (2026-05-30)

**Last touched:** 2026-05-29 (TRK-DEV-009 — per-chapter EPUB author bylines — landed; TRK-DESIGN-004 fontwork landed in the parallel session if its TRACKER section is also updated).

**Just shipped 2026-05-29 — TRK-DEV-009 done:**
- `srv/epub.go`: `epubSpec.Chapters` + `injectChapterAuthors` post-pandoc zip rewrite. Splices `<p class="chapter-author">{author}</p>` after the first `<h1>` of each chapter XHTML, matched in spine order. `mimetype` invariant preserved; XML escapes; nav/toc/title/cover skipped.
- `srv/epub_chapter_test.go`: five unit tests + a `parseEPUBSpec` chapters test.
- `srv/static/admin.html`: "Chapters (anthology bylines)" repeating-row editor in the EPUB card.
- Empty/missing `data.epub.chapters` → byte-identical EPUB to today's output. No migration (schema is JSON).
- **Next:** TRK-TEST-002 (live Ghosts visual regression) is now meaningful — EPUB diff against `reference/GHOSTS.epub` was previously blocked on this. Or TRK-DESIGN-003 (smart punctuation / body alignment), same `srv/epub.go` pandoc invocation, was deliberately deferred so it didn't conflict with this session.

---

**Earlier 2026-05-26 — three concurrent workstreams landed:**

**Just shipped 2026-05-26 — three concurrent workstreams landed:**

1. **TRK-DEV-007 (diff-vs-previous UI) done.** Compile-history panel rows now show a `diff` link that toggles an inline two-column comparison (spec changes left, corrections changes right) against the immediately-preceding compile of the same format. Empty diffs render explicit copy; legacy pre-snapshot rows degrade gracefully. JS-only landing in `srv/static/admin.html`: helpers `tsSpecDiff`, `tsParseCorrectionsYAML`, `tsCorrectionsDiff`; `tsRenderHistory` upgraded to fetch with `?include=spec,corrections`. Commit `705eb3f`.

2. **TRK-OPS-006 done.** Prod DB cleaned: 12 → 1 projects. Twitter Years (id 7) preserved; the other 11 (Art of Gig, tweetbook, Digital Garden, multiple test/smoke projects) deleted in a single transaction under `PRAGMA foreign_keys = ON;`. All FK cascades fired correctly. Zero orphans. VACUUM ran. Backup at `~/backups/prodcal-20260526-032947.sqlite3.gz`. Full before/after counts in TRK-OPS-006 close.

3. **TRK-DESIGN-001 audited; parity matrix shipped.** Doc at `docs/GHOSTS_PARITY_2026-05-26.md`: 25 ✅ / 10 ⚠️ (need live compile) / 1 ❌ (PDF/X-4 — Typst limitation). **Single biggest gap:** anthology per-chapter author bylines work on the Typst side (`set-story-info()` per chapter in template + main.typ) but break in the EPUB pipeline — `book_specs.data` has no per-chapter author array, so pandoc emits one book-level author. Critical-path for Ghosts-like series after Twitter Years. Three child tickets surfaced: TRK-DEV-009, TRK-DESIGN-004, enriched TRK-TEST-002. Commit `6480e94`.

**Pre-existing context (2026-05-26 earlier — TRK-MIG-006 + TRK-DEV-008):**
- Corrections round-trip live on PDF + EPUB; migration 016 (`book_outputs.corrections_snapshot`) applied; snapshots populated since.
- TRK-DEV-008 filed (P3, patcher ergonomics) — pick when warranted.

**Live state:**
- `prodcal` service active on `jdbbs.exe.xyz`, MainPID==listener (Type=notify), DB at `/home/exedev/prodcal/db.sqlite3`. Projects=1, books=6 (all linked to Twitter Years #7). Backup pipeline 3-2-1 satisfied; sentinels green.
- DB size grew from 117MB (TRK-MIG-009 cutover, 2026-05-12) to 276MB (2026-05-26) — `book_outputs.output_data` BLOBs from accumulating compile artifacts. Tracked as TRK-OPS-009 (retention policy, P3).

**Run before doing anything else** (use `jpull` on the Mac first):

```bash
jpull
ssh exedev@jdbbs.exe.xyz '\
  systemctl is-active prodcal && \
  cat ~/backups/.HEALTH-OK && \
  sqlite3 -readonly /home/exedev/prodcal/db.sqlite3 \
    "SELECT migration_number FROM migrations ORDER BY migration_number DESC LIMIT 3"'
curl -sI https://jdbbs.exe.xyz | head -1
```

Expect: `prodcal active`, `OK`, `16/15/14`, `HTTP/2 200`.

**Next priorities — release-confidence track for Ghosts-like anthologies:**

1. ~~**TRK-DEV-009**~~ — DONE 2026-05-29.
2. **TRK-TEST-002 (P1, ~2-3 hours)** — Live Ghosts compilation + visual regression. Run `manuscripts/ghosts/main.typ` through Typst on VM (or local), pandoc-compile the same source to EPUB, diff page-by-page against `reference/GHOSTS.pdf`/`.epub`. Closes the 10 ⚠️ cells in the parity matrix. Now unblocked (DEV-009 + DESIGN-004 both in).
3. **TRK-DESIGN-004 (P1, ~2-3 hours)** — CJK/Thai font bundling decision. Ghosts has CJK (Hiragino) and Thai (Thonburi) content; current code has CSS rules but no Typst fallback. Decide: bundle commercial fonts (TRK-DESIGN-002 covers licensing) or pick OFL alternatives.
4. **TRK-DESIGN-003 (P2, ~2-3 hours)** — Typography drift audit + smart-punctuation conversion. Body alignment (justify vs ragged-right) is the one confirmed CSS drift item; smart-punctuation conversion not yet wired in pandoc invocations. Likely fold in during a TEST-002 visual-diff session.

**Track 2 (after DEV-009 ships):**

5. **TRK-DEV-004** — Special-typography preservation class (data model + preflight + pipeline). ~3 sessions; Phase A is data model only.
6. **TRK-DEV-008 (any single item)** — corrections patcher ergonomics; item 1 (case-insensitive flag) is the cheapest unlock; item 4 (surface patcher warnings) is the highest-leverage.
7. **TRK-OPS-009 (P3)** — book_outputs retention policy. ~30 min when growth becomes uncomfortable.

After DEV-009 + TEST-002 + DESIGN-004 ship, v1 workflow is "complete enough to ship Ghosts-like titles." Translation layer (v2) is **TRK-TRANS-001..009** — see `docs/PRODUCTION_ROADMAP_2026-05-25.md`.

**Open questions for the user:**

- TRK-DESIGN-002 — commercial font licenses (Plantin MT Pro, Proxima Nova, Hiragino, Thonburi): user said yes earlier. Locate license docs, decide bundling vs subs.
- Want a real alert channel (Discord webhook / ntfy.sh / email) for backup-health failures?
- VM-side rename (`/home/exedev/prodcal/` → `/home/exedev/jdbbs/`, plus systemd unit) — defer indefinitely or schedule?

**Do NOT touch without re-reading the relevant TRK entry:**
- prodcal systemd unit (TRK-OPS-005, `Type=notify` + sd_notify).
- `backup-db.sh` env vars (TRK-OPS-007 phase 1 baseline-tuned).

---

## 🗂 Earlier resume block — 2026-05-25 (TRK-MIG-009 cutover)

**Just shipped — TRK-MIG-009 (canonical repo cutover):**
- `djinna/prodcal` HEAD force-pushed into `djinna/jdbbs:main` (commit `83e21f2` absorbs TRACKER.md, MIGRATION_LOG.md, NEXT_SESSION_PROMPT_2026-05-13.md, CLAUDE.md; gitignores `.claude/`).
- `djinna/prodcal` and `djinna/book-prod` archived (read-only) on GitHub; redirects from old URLs still resolve.
- Safety tags pushed: `pre-jdbbs-rename-2026-05-25` (on prodcal) and `pre-overwrite-2026-05-25` (on jdbbs).
- VM remote URL updated `git@github.com:djinna/prodcal.git` → `https://github.com/djinna/jdbbs.git`; pulled cleanly; service stayed active throughout.
- Local clones renamed: `~/jd-projects/prodcal` → `~/jd-projects/jdbbs`; old `~/jd-projects/jdbbs` preserved as `~/jd-projects/jdbbs-bootstrap-pre-2026-05-25`.
- Local `~/jd-projects/book-prod` clone deprecated (see FLOW-001).
- **NOT renamed** (deferred — invisible to users, low value, would require systemd unit touch which is path-sensitive per TRK-OPS-005): VM directory `/home/exedev/prodcal/`, systemd unit `prodcal.service`, binary name `prodcal`. These are internal-only.

**Live state (assumed healthy until verified):**
- `prodcal` service active on `jdbbs.exe.xyz`, MainPID==listener (Type=notify), DB at `/home/exedev/prodcal/db.sqlite3` (117 MB, projects=12, books=6, Twitter Years is project 7)
- Backup pipeline 3-2-1 satisfied
- All four sentinels green; `~/backups/.HEALTH-OK` < 1h old

**Run before doing anything else** (use `jpull` on the Mac to refresh local first):

```bash
# Local
jpull                                  # fetch jdbbs (+ archived clones if you still have them)

# VM-side smoke
ssh exedev@jdbbs.exe.xyz '\
  systemctl is-active prodcal && \
  cat ~/backups/.HEALTH-OK && \
  echo "--- crontab ---" && crontab -l | grep -v ^# && \
  echo "--- DB rowcounts ---" && \
  sqlite3 -readonly /home/exedev/prodcal/db.sqlite3 \
    "SELECT '"'"'projects='"'"' || COUNT(*) FROM projects UNION ALL SELECT '"'"'books='"'"' || COUNT(*) FROM books" && \
  echo "--- R2 ---" && rclone size r2:jdbbs-backups'
curl -sI https://jdbbs.exe.xyz | head -1
jbackup-pull   # only on Macs where this function is configured
```

**Next priorities (see `docs/PRODUCTION_ROADMAP_2026-05-25.md` for the full v1→v2 path):**

1. **TRK-DEV-002 (CP-1) + TRK-MIG-007 (CP-4)** — Wire spec→compile pipeline for Typst, verify Libertinus on VM. The keystone session: makes the admin SPA actually produce books. ~3-4 hours. Unblocks DESIGN-001 and TEST-001.
2. **TRK-DEV-003 (CP-2) + TRK-DESIGN-003 (CP-6)** — Same for EPUB + audit typography drift (e.g. body-text alignment). ~3 hours.
3. **TRK-MIG-006 (CP-3)** — Expose corrections API + round-trip a real correction through to a regenerated artifact. ~3 hours.
4. **TRK-DESIGN-001 (CP-5)** — Ghosts parity matrix. Release-confidence check after CP-1..CP-4. ~2-3 hours.
5. **TRK-OPS-006** — warm-up: drop 12 test/dummy projects from prod DB. ~5-10 min, mechanical.

After CP-1..CP-5 ship, v1 workflow is "complete." Translation layer (v2) is **TRK-TRANS-001..009** — see `docs/PRODUCTION_ROADMAP_2026-05-25.md` for the v2 phases and `docs/translation layer 2026-05-25.md` for the original design discussion.

**Open questions for the user:**

- Want a real alert channel (Discord webhook, ntfy.sh, exe.dev email forward) for backup-health failures? Currently observable only via `jbackup-pull`.
- Are the 12 test projects 100% disposable, or should any be exported / kept for reference before delete? (Earlier confirmed: only Twitter Years is real.)
- Set up `jbackup-pull` on the second Mac too? (Asymmetry observed 2026-05-25.)
- VM-side rename (`/home/exedev/prodcal/` → `/home/exedev/jdbbs/`, plus systemd unit) — defer indefinitely, or schedule as a discrete cutover session?

**Do NOT touch without re-reading the relevant TRK entry:**
- prodcal systemd unit — restart pattern is documented in TRK-OPS-005, depends on `Type=notify` + `sd_notify` ready signal
- `backup-db.sh` env vars — `MIN_GZ_BYTES`/`MIN_PROJECTS` thresholds are baseline-tuned, see TRK-OPS-007 phase 1

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
  - [TRK-OPS-008 — VM-side rename: prodcal → jdbbs](#trk-ops-008--vm-side-rename-prodcal--jdbbs-directory-systemd-unit-binary-scripts)
  - [TRK-OPS-009 — book_outputs retention policy](#trk-ops-009--book_outputs-retention-policy)
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
  - [TRK-DESIGN-003 — Typography drift audit + smart-punctuation conversion](#trk-design-003--typography-drift-audit--smart-punctuation-conversion-epub-and-typst)
  - [TRK-DESIGN-004 — CJK/Thai font bundling & fallback strategy](#trk-design-004--cjkthai-font-bundling--fallback-strategy)
  - [TRK-DESIGN-002 — Commercial font licensing & bundling](#trk-design-002--commercial-font-licensing--bundling)
- [Dev (DEV)](#dev-dev)
  - [TRK-DEV-001 — `series-template.typ` config override mechanism (consolidate)](#trk-dev-001--series-templatetyp-config-override-mechanism-consolidate)
  - [TRK-DEV-002 — Wire spec → Typst compile pipeline (CP-1, KEYSTONE)](#trk-dev-002--wire-spec--typst-compile-pipeline-cp-1-keystone)
  - [TRK-DEV-004 — Special-typography preservation class](#trk-dev-004--special-typography-preservation-class-data-model--preflight--pipeline)
  - [TRK-DEV-003 — Wire spec → EPUB compile pipeline (CP-2)](#trk-dev-003--wire-spec--epub-compile-pipeline-cp-2)
  - [TRK-DEV-007 — Diff-vs-previous UI in compile-history panel](#trk-dev-007--diff-vs-previous-ui-in-compile-history-panel)
  - [TRK-DEV-009 — Per-chapter author in EPUB spec + pipeline (anthology critical-path)](#trk-dev-009--per-chapter-author-in-epub-spec--pipeline-anthology-critical-path)
  - [TRK-DEV-005 — Compile-history panel in admin SPA](#trk-dev-005--compile-history-panel-in-admin-spa)
  - [TRK-DEV-006 — Snapshot spec JSON into book_outputs per compile](#trk-dev-006--snapshot-spec-json-into-book_outputs-per-compile)
  - [TRK-DEV-010 — Wire `--epub-embed-font` into pandoc invocation for Noto CJK/Thai bundle](#trk-dev-010--wire---epub-embed-font-into-pandoc-invocation-for-noto-cjkthai-bundle)
  - [TRK-DEV-011 — Admin SPA UX polish (project workflow gaps)](#trk-dev-011--admin-spa-ux-polish-project-workflow-gaps)
  - [TRK-DEV-008 — Corrections patcher ergonomics](#trk-dev-008--corrections-patcher-ergonomics)
- [Translation (TRANS) — v2](#translation-trans--v2)
  - [TRK-TRANS-001..009 — see PRODUCTION_ROADMAP_2026-05-25.md](#translation-trans--v2)
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
- status: done
- priority: P3
- created: 2026-05-12
- updated: 2026-05-25
- refs: prodcal `83e21f2` (absorption commit, force-pushed into jdbbs:main); safety tags `pre-jdbbs-rename-2026-05-25` (prodcal) + `pre-overwrite-2026-05-25` (jdbbs)

**Done 2026-05-25.** Chose **Option B (clean cutover)** with safety tags. Steps executed:

1. Absorbed jdbbs-unique files (TRACKER.md, MIGRATION_LOG.md, NEXT_SESSION_PROMPT_2026-05-13.md, CLAUDE.md) into prodcal as commit `83e21f2`; pushed to djinna/prodcal.
2. Tagged both repos at pre-cutover HEADs (`pre-jdbbs-rename-2026-05-25`, `pre-overwrite-2026-05-25`); pushed tags.
3. Force-pushed prodcal `83e21f2` into djinna/jdbbs:main (overwriting the 24-commit bootstrap history, which remains accessible via `pre-overwrite-2026-05-25` tag).
4. Renamed local clones: `~/jd-projects/prodcal` → `~/jd-projects/jdbbs`; old jdbbs preserved at `~/jd-projects/jdbbs-bootstrap-pre-2026-05-25`.
5. VM: updated remote URL `git@github.com:djinna/prodcal.git` → `https://github.com/djinna/jdbbs.git`; `git fetch && git pull --ff-only` brought in the absorption commit; service stayed active throughout (no systemd touch).
6. Archived `djinna/prodcal` and `djinna/book-prod` on GitHub (read-only; redirects preserved).

**Deliberately deferred** (low value, would require systemd touch which is path-sensitive per TRK-OPS-005):
- VM directory rename `/home/exedev/prodcal/` → `/home/exedev/jdbbs/`
- systemd unit rename `prodcal.service` → `jdbbs.service`
- binary name change `prodcal` → `jdbbs`
- backup script path updates (`backup-db.sh`, `sync-to-r2.sh`, `check-backups.sh`)
- crontab path updates

Filed as **TRK-OPS-008** (open, P3) for if/when it becomes worth doing.

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
- status: done
- priority: P2
- created: 2026-05-12
- updated: 2026-05-26
- refs: commits 9af05ad, d69b4e3, 3918527, 10a07c7; db/migrations/016-book-output-corrections-snapshot.sql; srv/corrections_apply.go; srv/books.go::runConversion + srv/epub.go::runEPUBGeneration; typesetting/scripts/apply-corrections-docx.py

**Done 2026-05-26.** Both compile pipelines materialize pending corrections (`status='pending'`) as YAML in-process and patch the source docx via `apply-corrections-docx.py` before pandoc runs. Migration 016 adds `book_outputs.corrections_snapshot TEXT NULL` pairing with the spec_snapshot from DEV-006; `GET /api/books/{id}/outputs?include=spec,corrections` returns both per row.

**Hardening that happened in the loop:**
- Patcher originally walked only `doc.paragraphs` — extended to body + tables + headers/footers + footnotes + endnotes. Footnotes/endnotes are plain `Part` instances in python-docx (no `.element`), so they're parsed from `.blob` with lxml, mutated, and written back via `part.blob =` before `doc.save`. Each non-body scope is wrapped in its own try/except so a single edge case can't take out the body pass.
- In-memory metadata patch: pandoc receives author/title via `--metadata` flags pulled from the `books` table (not docx content). Without patching `book.Title`/`book.Author` (and `spec.Title/Author/Subject/Description`) before they reach pandoc, the EPUB title page rendered unpatched values. Same patch applied in the PDF pipeline since the typst template reads `book.Title`/`Author` too.
- Status flips (pending → applied) stay manual: every compile starts from the original docx and re-applies all pending entries, so auto-flipping would silently drop the fix on the next compile. The user marks applied when the canonical Word doc subsumes the fix.
- Failure to patch (python deps missing, malformed correction, etc.) logs a warn and continues with the unpatched source — a typo ledger shouldn't block a build.

**Verified live** with `Venkatesh → Venkat` (exercises body byline, footnote citation, pandoc metadata, typst metadata) and `alchemy → al-TEST` (body-local sanity check). Case-sensitive matching is intentional (matches the `iphone → iPhone` example in the script docstring) — capitalized variants like `Alchemy` need their own correction row. Future ergonomics filed as **TRK-DEV-008**.

### TRK-MIG-007 — Verify Libertinus Serif on VM (or bundle)

- area: MIG
- status: done
- priority: P1
- created: 2026-05-12
- updated: 2026-05-25
- refs: jdbbs/typesetting/fonts/libertinus/, commit d451aa4

**Done 2026-05-25.** Audited VM: `fc-list | grep -i libertinus` returned empty, `~/prodcal/typesetting/fonts/` contained only jetbrainsmono + sourcesans. Took the preferred path: bundled `Libertinus-7.040` (alerque/libertinus, OFL) Serif family (Regular/Bold/Italic/BoldItalic/Semibold/SemiboldItalic) into `typesetting/fonts/libertinus/OTF/`. No code change needed — `typst compile --font-path typesetting/fonts` (srv/books.go:282) scans recursively. Live verification deferred to TRK-DEV-002 smoke test.

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
- status: done
- priority: P2
- created: 2026-05-12
- updated: 2026-05-26
- refs: backup `~/backups/prodcal-20260526-032947.sqlite3.gz` on VM

**Done 2026-05-26.** Fresh backup taken via `scripts/backup-db.sh` (probe: projects=12 books=6 integrity=ok, 261MB gzipped). Executed `DELETE FROM projects WHERE id != 7` in a single transaction with `PRAGMA foreign_keys = ON;`.

**Important gotcha for future similar work:** SQLite's `PRAGMA foreign_keys` defaults to OFF per-connection. Without enabling it explicitly, FK cascades silently no-op — the DELETE succeeds but leaves orphans in every dependent table. Always set the pragma at the start of any session that depends on cascades.

**Before → after row counts:**

| Table | Before | After | Cascade source |
|---|---:|---:|---|
| projects | 12 | 1 | (target of DELETE) |
| books | 6 | 6 | `project_id` ON DELETE SET NULL — all already linked to #7, none orphaned |
| book_specs | 5 | 1 | ON DELETE CASCADE |
| transmittals | 9 | 1 | ON DELETE CASCADE |
| transmittal_versions | 27 | 12 | via transmittals CASCADE |
| corrections | 2 | 1 | ON DELETE CASCADE |
| tasks | 166 | 0 | ON DELETE CASCADE (project 7 had no tasks) |
| manuscript_preflights | 12 | 12 | all 12 linked to project 7 |
| journal | 4 | 0 | ON DELETE CASCADE |
| file_log | 4 | 0 | ON DELETE CASCADE |
| auth_tokens | 3 | 0 | ON DELETE CASCADE |
| book_outputs | 22 | 22 | via books (books retained → outputs retained) |

Zero orphans verified post-commit. VACUUM ran.

**Note:** Deleted project names were Art of Gig, tweetbook, The Digital Garden, test (twice), Smoke Test 2026-04-08, 9apr26 test, Admin 9 Apr Seed Check, test 3, test 01, test 02. None named "Ghosts" — the Ghosts manuscript lives at `manuscripts/ghosts/` in the repo and `reference/GHOSTS.{pdf,epub}` for the golden output, not as a DB project. The Typst direct-compile path doesn't need a DB project; if/when the admin SPA path is needed for Ghosts (e.g. after TRK-DEV-009 to test multi-author EPUB end-to-end), a fresh Ghosts project would be created from the manuscript DOCX upload.

**DB size note:** post-VACUUM size is 276MB (up from 117MB at TRK-MIG-009 cutover 2026-05-12). Growth dominated by `book_outputs.output_data` BLOBs from compile-history retention. Tracked as TRK-OPS-009.

### TRK-OPS-007 — Backup hygiene: integrity probes + off-VM replication

- area: OPS / SEC
- status: done (phase 1 + 2 + 3)
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

**Phase 3 — observability (done 2026-05-12).**

Three additions to close the "silent backup rot" gap:

- **`scripts/check-backups.sh`** (cron hourly at minute 17). Examines the four sentinels. FAILS if any `.LAST-FAILURE` or `.LAST-R2-FAILURE` is present, OR if `.LAST-SUCCESS` / `.LAST-R2-SUCCESS` is missing or older than `$MAX_AGE_HOURS` (default 26h). On success writes `.HEALTH-OK` with the timestamps of the most recent backup + R2 push. On failure writes `.HEALTH-FAIL` with details. Verified 2026-05-12 18:41 with 4 scenarios:
  - positive (everything fresh) → `HEALTH OK`
  - .LAST-SUCCESS touched to 48h ago → `STALE — last updated 48h 00m ago (max 26h)`
  - synthetic .LAST-FAILURE present → reports the reason
  - back to clean state → `HEALTH OK`
- **`scripts/restore-drill.sh`** (cron monthly at 04:00 UTC on the 1st). Picks the newest local backup, decompresses to a tmp file, asserts (a) all six expected tables present, (b) projects + books rowcounts ≥ thresholds, (c) `PRAGMA integrity_check` returns `ok`, (d) logs a sample of the first 5 projects. Verified 2026-05-12 18:41: all checks passed against `prodcal-20260512-183831.sqlite3.gz`.
- **Mac-side `jbackup-pull` extended** to (a) rsync all eight sentinel files, not just `.LAST-SUCCESS` / `.LAST-FAILURE`, and (b) compute human-readable ages of each (e.g. `.HEALTH-OK (0h 1m ago)`) by parsing the embedded `time:` field as UTC.

**Cron now installed on VM:**
```cron
# Daily backup + R2 push at 03:00 UTC.
0 3 * * * /home/exedev/prodcal/scripts/backup-db.sh >> /home/exedev/backups/backup.log 2>&1; /home/exedev/prodcal/scripts/sync-to-r2.sh >> /home/exedev/backups/backup.log 2>&1
# Hourly health probe.
17 * * * * /home/exedev/prodcal/scripts/check-backups.sh >> /home/exedev/backups/backup-health.log 2>&1
# Monthly restore drill on the 1st at 04:00 UTC.
0 4 1 * * /home/exedev/prodcal/scripts/restore-drill.sh >> /home/exedev/backups/backup.log 2>&1
```

**Cleanup completed 2026-05-12 18:51 UTC:**
- 8 pre-fix 3.5 KB artifacts (`prodcal-20260505-030001.sqlite3.gz` through `prodcal-20260512-030001.sqlite3.gz`) deleted from VM `~/backups/`, R2 `r2:jdbbs-backups/db/`, and Mac `~/backups-jdbbs/`.
- Each of the three locations now holds the same 3 real 109 MB backups from today's recovery + verification runs (171252, 172054, 183831).
- check-backups.sh after cull: `HEALTH OK`.

**Still open (small follow-ups, low priority):**
- check-backups.sh writes to a log; no actual alarm channel (email/pager/webhook) is wired up. With `jbackup-pull` on the Mac surfacing the sentinels, this is observable but not push-notified. Could add a Discord/ntfy.sh webhook later if/when usage justifies it.

**RPO/RTO targets:**
- RPO: ≤ 24 hours (daily backups); upgrade to ≤ 1 hour later with litestream if we go heavier on usage.
- RTO: ≤ 30 minutes for a fresh VM (spin up + pull repo + `rclone copy r2:jdbbs-backups/db/<latest> .` + gunzip + start systemd).

### TRK-OPS-008 — VM-side rename: prodcal → jdbbs (directory, systemd unit, binary, scripts)

- area: OPS
- status: open
- priority: P3
- created: 2026-05-25
- updated: 2026-05-25
- refs: TRK-MIG-009 (deferred this work), TRK-OPS-005 (orphan-race fix is path-sensitive)

Cosmetic cleanup deferred from MIG-009. After 2026-05-25, GitHub repo is `djinna/jdbbs`, local clone is `~/jd-projects/jdbbs`, but VM is still under `/home/exedev/prodcal/` with `prodcal.service` running `./prodcal` binary. Nothing user-facing depends on these names (jdbbs.exe.xyz is nginx-fronted). Items to do if/when:

1. Stop `prodcal.service`.
2. `mv /home/exedev/prodcal /home/exedev/jdbbs`.
3. Edit `srv.service` → `WorkingDirectory=/home/exedev/jdbbs`, `Environment=JDBBS_TYPESETTING_DIR=/home/exedev/jdbbs/typesetting`, `ExecStart=/home/exedev/jdbbs/jdbbs` (after renaming the binary). Rename `srv.service` → `jdbbs.service` and re-symlink in `/etc/systemd/system/`.
4. `go build -o jdbbs ./cmd/srv` (or just `mv prodcal jdbbs`).
5. Update path references in `scripts/backup-db.sh`, `scripts/sync-to-r2.sh`, `scripts/check-backups.sh`, `scripts/restore-drill.sh` (grep `/home/exedev/prodcal` and `db.sqlite3` location).
6. Update crontab paths (`crontab -l | sed 's|/prodcal|/jdbbs|g' | crontab -`).
7. `daemon-reload`, `enable`, `start jdbbs.service`. Verify with the standard verification block.
8. Confirm next backup cron run produces a valid sentinel.

Risk: TRK-OPS-005's orphan-race fix depends on `Type=notify` + `sd_notify` ordering — re-validate after the unit-file rewrite. Plan ≥30 min and have a rollback (revert the unit file, mv directory back).

### TRK-OPS-009 — book_outputs retention policy

- area: OPS
- status: open
- priority: P3
- created: 2026-05-26
- updated: 2026-05-26
- refs: db/migrations/014-book-output-history.sql; TRK-DEV-005 close note ("keep everything for now"); 2026-05-26 observation: DB grew from 117MB to 276MB in two weeks

**Context.** Every PDF/EPUB compile persists the full artifact bytes into `book_outputs.output_data` (DEV-005). Over a couple of weeks of dev compiling, the DB grew from 117MB (cutover 2026-05-12) to 276MB (2026-05-26). Most of that is `book_outputs` BLOBs.

Not urgent — backups are still under 300MB gzipped, R2 cost is trivial — but worth a policy before it reaches a real pain point. Options:

1. **Keep latest N per (book_id, output_format).** Simple `DELETE` cron, e.g. keep 20 of each. Loses history beyond that window.
2. **Keep all "marked" outputs + latest N otherwise.** Add `book_outputs.pinned BOOLEAN DEFAULT 0`; users mark via UI; pruning skips pinned rows.
3. **Move blobs to filesystem after N days.** `book_outputs.output_data` becomes a filesystem path for older rows. More complex; defers the size pressure rather than removing it.
4. **Do nothing.** Plausible for the next 6-12 months at current growth rate.

**Recommendation:** option 2, but file as a small implementation ticket only when DB size hits ~1GB or backup duration becomes noticeable. Until then, this is just a watch item.

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
- status: done
- priority: P3
- created: 2026-05-12
- updated: 2026-05-25

**Done 2026-05-25** via TRK-MIG-009. djinna/book-prod archived, local `~/jd-projects/book-prod` deprecated, single canonical repo (jdbbs) means there's only one place per-session prompts can land. Convention going forward: TRACKER.md "Resume here" block is the live entry point; per-session prompts (`NEXT_SESSION_PROMPT_YYYY-MM-DD.md`) are end-of-session snapshots committed to jdbbs root, not mirrored anywhere.

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
- status: in-progress (audit done; live-compile verification + child tickets remain)
- priority: P1
- created: 2026-05-12
- updated: 2026-05-26
- refs: `docs/GHOSTS_PARITY_2026-05-26.md` (full audit, 25✅/10⚠️/1❌); `reference/GHOSTS.pdf` (136 pages, 353.811 × 546.567 pt = 4.91 × 7.59 in, PDF/X-4, InDesign 21.0); `manuscripts/ghosts/main.typ` (per-chapter Typst sources with `set-story-info()` configured for 9 chapters)

**2026-05-26 audit done (subagent, read-only).** Full matrix in `docs/GHOSTS_PARITY_2026-05-26.md`. Summary: 25 ✅ items match code, 10 ⚠️ need live-compile verification, 1 ❌ (PDF/X-4 — Typst limitation; post-process via Ghostscript possible).

**Critical finding:** anthology per-chapter author bylines work on the Typst side (`set-story-info(title:, author:)` per chapter in template; main.typ configures 9 chapters with different authors) but **break in the EPUB pipeline** — `book_specs.data` schema has no per-chapter author array, so pandoc emits one book-level author for the whole EPUB regardless of source markup. Filed as **TRK-DEV-009** (P1, critical-path for Ghosts-like anthologies).

**Child tickets filed/enriched:**
- TRK-DEV-009 — Per-chapter author in EPUB spec + pipeline (new, P1)
- TRK-DESIGN-004 — CJK/Thai font bundling (new, P1) — Ghosts has multilingual content (Hiragino, Thonburi)
- TRK-TEST-002 — enriched with concrete visual-regression approach (existing, P1) — live compile + page-by-page diff vs reference
- TRK-DESIGN-003 — already covers body-alignment drift + smart-punctuation (existing, P2)

**Closure criteria:** the 10 ⚠️ cells get resolved via TRK-TEST-002 live compilation; the 1 ❌ (PDF/X-4) is accepted as can't-have unless a post-process step is added later. Final close happens after TRK-DEV-009 + TRK-DESIGN-004 + TRK-TEST-002 all ship.

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

**Original action (superseded by 2026-05-26 audit above):** ~~spend a focused pass with InDesign PDF + a rendered Typst PDF side-by-side; fill remaining `❓` cells; produce a final cant-have list.~~ The detailed audit is in `docs/GHOSTS_PARITY_2026-05-26.md`; remaining ❓/⚠️ cells will close via TRK-TEST-002 (live compile + diff).

### TRK-DESIGN-003 — Typography drift audit + smart-punctuation conversion (EPUB and Typst)

- area: DESIGN
- status: open
- priority: P2
- created: 2026-05-25
- updated: 2026-05-26
- refs: typesetting/templates/epub/epub-styles.css; book-prod-archived-2026-05-25/TYPOGRAPHY_REFINEMENT_PROMPT.md; PRODUCTION_ROADMAP_2026-05-25.md (CP-6); srv/books.go:230 + srv/epub.go:132 (pandoc invocations); typesetting/scripts/{docx2epub.sh,build.sh}

Two related typography-polish concerns, same session.

**Part A — CSS drift audit.** Audit current EPUB CSS vs the InDesign-derived spec in the (archived) refinement prompt. Confirmed gap: body text is `text-align: justify` in current CSS; prompt specifies left-aligned ("originals use ragged right"). Other potential drift (margins, indent values, line-heights) — needs systematic walk-through. **Decision per-style:** change CSS, expose as a toggle in book_specs, or leave (and update the spec doc).

**Part B — Smart-punctuation conversion (added 2026-05-26).** Neither pandoc invocation (`srv/books.go:230` for DOCX→Typst, `srv/epub.go:132` for DOCX→EPUB) has `+smart` on the input format; same for the shell-script variants. So straight quotes pass through verbatim, `--` stays as two hyphens, `...` stays as three dots. Fix: add `+smart` to the `--from` extension list in both Go invocations (and the shell scripts if they're still on a code path). Specifically converts:

- `'` (apostrophe + open/close single) → `'` / `'`
- `"` (open/close double) → `"` / `"`
- `--` → en-dash `–`
- `---` → em-dash `—`
- `...` → ellipsis `…`

Verify the conversion doesn't bleed into declared special-typography blocks (TRK-DEV-004) — preformatted content with literal `--` or `...` (terminal transcripts, tweet handles, ASCII art) must NOT be smart-converted. Pandoc's `smart` extension applies broadly, so the Lua filter or a downstream pass may need to detect preserved blocks and revert. Coordinate with TRK-DEV-004 Phase C/D.

**v2 locale wrinkle (for awareness, not for this ticket):** Pandoc's `smart` is locale-aware via `--metadata=lang:fr` etc. — French gets « », German „ ", etc. The per-language Typst templates (TRK-TRANS-004) and per-language EPUB stylesheets (TRK-TRANS-005) will need this metadata wired through from the spec. TRK-TRANS-006 (validators) will need to verify locale-appropriate punctuation is actually present in output.

**Acceptance:**
- A manuscript with straight quotes and `--` produces curly quotes and em-dash in both PDF and EPUB.
- A declared special-typography block (e.g. terminal transcript) preserves literal characters.
- EPUB CSS audit produces a documented list of drift items with per-item decisions.

Fold into the same session as TRK-DEV-003 (now closed) — or schedule independently after the diagnostic loop (TRK-DEV-005/006) lands.

### TRK-DESIGN-004 — CJK/Thai font bundling & fallback strategy

- area: DESIGN
- status: done (2026-05-25, VM smoke pending; EPUB embed-font wiring deferred to TRK-DEV-010)
- priority: P1
- created: 2026-05-26
- updated: 2026-05-25
- refs: TRK-DESIGN-001, TRK-DESIGN-002, `docs/GHOSTS_PARITY_2026-05-26.md` §"CJK Glyphs" / "Thai Glyphs"; `typesetting/templates/epub/epub-styles.css`; `typesetting/templates/series-template.typ`; `typesetting/fonts/noto/`; `manuscripts/ghosts/08 Loyalty.md`

**Resolution (2026-05-25).**

*Audit.* Scanned `manuscripts/ghosts/*.md` for CJK + Thai codepoints. **Only one chapter has multilingual content:** `08 Loyalty.md` line 39 — `**Lin Store 47---林家商店47號---ร้านยามาลิน สาขา ๔๗**`. Five Han characters (Traditional Chinese — `林家商店` Lin Family Store, Taiwanese setting; **no hiragana/katakana**, so the parity doc's "Japanese" framing was wrong) and 17 Thai characters including digits `๔๗`. Bundled TC + Thai serif only — no Japanese, no SC, no Korean.

*Bundled fonts under `typesetting/fonts/noto/`:*
- `CJK-TC/NotoSerifTC-{Regular,Bold}.otf` — Noto Serif TC v2.003 (Adobe/Google OFL), ~16MB total. Source: https://github.com/notofonts/noto-cjk/releases/tag/Serif2.003 (15_NotoSerifTC.zip → SubsetOTF/TC/).
- `Thai/NotoSerifThai-{Regular,Bold}.ttf` — Noto Serif Thai v2.002 (OFL), uncondensed Regular + Bold, ~180KB. Source: https://github.com/notofonts/thai/releases/tag/NotoSerifThai-v2.002.
- `OFL.txt` license shipped alongside each family.

*Typst wiring.* `typesetting/templates/series-template.typ` base typography `set text` now passes a font array — primary `config.body-font` (Libertinus Serif) with `"Noto Serif TC"` and `"Noto Serif Thai"` as fallbacks. Typst's `--font-path typesetting/fonts` recursive scan (`srv/books.go::runConversion`) discovers `noto/CJK-TC/*.otf` + `noto/Thai/*.ttf` automatically — no Go change required.

*EPUB wiring.* `typesetting/templates/epub/epub-styles.css`: added four `@font-face` declarations at top of file (`src: url(NotoSerifTC-Regular.otf)` etc — pandoc convention is to embed by basename). Updated `.chinese` family to lead with `Noto Serif TC` (then existing Hiragino / PingFang / Microsoft YaHei). Updated `.thai` to lead with `Noto Serif Thai` (then Thonburi / Leelawadee). Body and `p` stacks now include the Noto families between Libertinus and Georgia so inline multilingual runs (the Ghosts ch. 8 line is bold body text mixing English + TC + Thai without a wrapping `.chinese`/`.thai` class) cascade correctly.

*Deferred — pandoc `--epub-embed-font` flags (filed as TRK-DEV-010).* To make EPUBs self-contained on readers without OS-installed Noto, `srv/epub.go` needs `--epub-embed-font=<path>` args for each bundled font (4 lines near line 144). Not done in this session because TRK-DEV-009 was concurrent on the same file. Until then EPUB rendering relies on OS-installed Noto / Hiragino / Thonburi fallback — works on macOS readers and on Linux readers with `fonts-noto-cjk` + `fonts-noto-thai` packages installed (the prodcal VM does not currently have these — Docker / VM provisioning may want them as belt-and-suspenders).

*Smoke test status.* No local `typst` binary on this Mac; the Typst smoke specified in the session prompt (step 3) was deferred to deploy. Verify on VM with:
```bash
ssh exedev@jdbbs.exe.xyz 'cat > /tmp/font-smoke.typ <<EOF
#set text(font: ("Libertinus Serif", "Noto Serif TC", "Noto Serif Thai"))
English. 林家商店. ภาษาไทย.
EOF
cd /home/exedev/prodcal && typst compile --font-path typesetting/fonts /tmp/font-smoke.typ /tmp/font-smoke.pdf'
```

**Original problem.** Reference GHOSTS.pdf embeds HiraKakuPro-W3 (Japanese) and Thonburi (Thai) fonts for multilingual content (the Khlongs chapter and other passages). Current state:
- **Typst:** no CJK/Thai font bundled in `typesetting/fonts/`; would fall back to OS defaults (works on macOS, unpredictable on Linux/VM)
- **EPUB:** CSS classes `.chinese` / `.thai` reference Hiragino/Thonburi by name with fallback stacks; e-reader rendering depends on whether the user's device has the fonts

Without resolution, the Ghosts compile will produce ❌ for CJK/Thai cells in the parity matrix (tofu boxes or wrong-glyph substitution).

**Decision points (need user input):**

1. **Bundle the commercial fonts (Hiragino, Thonburi)?** User indicated they have licenses for Plantin + Proxima (TRK-DESIGN-002). Hiragino is Apple-system on macOS — licensing for embedding in distributed EPUB/PDF is restrictive. Thonburi same. *Probably* requires OFL alternatives even if user owns macOS-distributable copies.
2. **Pick OFL alternatives:** Noto Serif CJK JP (and SC/TC for Chinese variants) is the standard OFL CJK font; Noto Sans Thai for Thai. Metrics differ from Hiragino/Thonburi — visual drift in the parity diff is expected.
3. **Embed in EPUB?** Once chosen, add `@font-face` to `epub-styles.css` and ship the font files inside the EPUB package. This guarantees rendering on any e-reader but inflates EPUB size by ~5-15MB per CJK family.

**Suggested approach:**

1. Audit `manuscripts/ghosts/` for actual CJK/Thai content extent — Unicode block scan via `python -c "import unicodedata; ..."`. Confirms what fonts are actually needed.
2. Bundle Noto Serif CJK JP + Noto Sans Thai (OFL) into `typesetting/fonts/noto/`. Same shape as the Libertinus bundle (TRK-MIG-007).
3. Update `series-template.typ` font fallback chain to include the new families (Typst's `text` `font:` accepts a list).
4. Update `epub-styles.css` `.chinese` / `.thai` classes to use the bundled families as primary; declare `@font-face` if shipping inside EPUB.
5. Smoke test against the Khlongs chapter specifically.

**Effort:** ~2-3 hours including font download + license verification + smoke test. **Blocks:** TRK-TEST-002 visual regression (CJK/Thai cells can't be evaluated without working fonts). **Related:** TRK-DESIGN-002 (broader font-licensing question — owner has Plantin/Proxima but probably not Hiragino/Thonburi distribution rights).

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

### TRK-DEV-002 — Wire spec → Typst compile pipeline (CP-1, KEYSTONE)

- area: DEV
- status: done
- priority: P1
- created: 2026-05-25
- updated: 2026-05-26
- refs: srv/books.go:189-321 (runConversion), srv/books.go:455 (buildTypstConfig), srv/bookspecs.go:514 (specToTypstConfig), srv/static/admin.html:2598 (Compile PDF handler), commits 688195a (config plumbing fix), 792c76d (timestamped filenames)

**Closed 2026-05-26.** Live smoke against The Twitter Years (book 8, project 7): base_size_pt 10→13 produced visibly larger body text (page count 533→570) AND the tweet block restyled per custom-style override. Two bugs surfaced and fixed during smoke:

1. **`book.with()` wasn't receiving caller's merged config** — only the template module's default-config — so body-font/base-size/margin overrides silently no-opped. Custom styles (e.g. tweet-p) worked because they close over `config` in the caller scope. Fix (688195a): pass `config: config,` to `book.with(...)`. This was THE bug that made it look like the pipeline wasn't wired even after the per-book convert path was confirmed end-to-end.
2. **PDF download cached in the browser** — same URL between compiles, no Cache-Control header, so back-to-back compiles served identical bytes from cache. Fix (792c76d): `Cache-Control: no-store` + timestamp-suffixed `Content-Disposition` filename (`{title}-{YYYYMMDD-HHMMSS}.pdf` from `books.updated_at`).

Follow-ups filed: TRK-DEV-005 (compile-history panel — `book_outputs` already populated, no UI), TRK-DEV-006 (snapshot spec JSON per output row for diffable lineage).

**Resolution 2026-05-25.** Fresh-session audit showed the 2026-05-25 "current state" section below was stale — the seam was already wired in a prior session and not closed out in the tracker. Verified end-to-end trace:

- `srv/books.go::runConversion` line 232 calls `s.buildTypstConfig(bid, book)`.
- `buildTypstConfig` (line 455) resolves `book.ProjectID` → `q.GetBookSpec` → returns `specToTypstConfig(data)`.
- `specToTypstConfig` (srv/bookspecs.go:514) — shared between the existing `handleGenerateConfig` endpoint and `runConversion`, so the optional "refactor to extract a helper" in step 4 is already realized.
- Result is interpolated into the header replacement (line 232-260) just after the `#import "…series-template.typ": *` line, so the spec's `#let config = merge-config((…))` override takes effect before `#show: book.with(…)`.
- SPA `ts-compile-pdf` button (admin.html:2598) calls `tsSaveSpec()` first if dirty, then `POST /api/books/{bid}/convert` → triggers the spec-aware path.

**Step 2 (new `POST /api/projects/{id}/book-spec/compile` endpoint) deliberately skipped:** redundant with the per-book convert route which already does spec-driven compile. Adding it would duplicate route plumbing without changing behavior. If a project-scoped "compile all books in this project" batch op is wanted later, file a fresh ticket.

**Remaining = live smoke (folded into TRK-MIG-007 cycle):** push d451aa4 (Libertinus bundle), redeploy, edit a Twitter Years spec value (e.g. base size), recompile, eyeball the diff in the PDF. If the diff is visible and Libertinus renders, mark this fully done.

#### Current state (audited 2026-05-25 — start here, don't re-discover)

**Already shipped, reuse:**
- `srv/bookspecs.go::handleGenerateConfig` (line ~454) — given a project, produces a Typst config dict. Output format matches what `series-template.typ::merge-config` expects.
- `srv/bookspecs.go` — full CRUD for `book_specs`: GET, PUT, pull-transmittal, generate-word-template, upload-cover.
- `typesetting/templates/series-template.typ` (lines 13-90) — `default-config` dict + `merge-config()` (line 80-86). Callers can merge overrides before importing.
- `db/migrations/008-book-specs.sql` — `book_specs` table, one per project.
- `db/migrations/009-book-project-link.sql` — `books.project_id` FK (needed to look up spec from book).
- Admin SPA "Compile PDF" button exists in `srv/static/admin.html` (~line 550-700). Currently calls preflight endpoint, not a compile endpoint.
- `typesetting/scripts/build.sh` — Typst-to-PDF wrapper. Likely the right shell to invoke from Go.

**Missing seam — this ticket:**
- No call to `handleGenerateConfig` from `runConversion()` in `srv/books.go`.
- `runConversion()` writes a bare `main.typ` with no spec-derived config import.
- No `POST /api/projects/{id}/book-spec/compile` endpoint.
- Admin SPA "Compile PDF" button is wired to preflight, not to compile.

**Watch out:**
- VM systemd unit `prodcal.service` is path-sensitive (TRK-OPS-005) — do NOT touch the unit file. Just push new binary + restart.
- VM dir is still `/home/exedev/prodcal/` (TRK-OPS-008 deferred). Local path is `~/jd-projects/jdbbs/` — same content, different parent name.
- Single canonical repo as of 2026-05-25 (TRK-MIG-009). Push to `djinna/jdbbs`, deploy by `ssh exedev@jdbbs.exe.xyz 'cd /home/exedev/prodcal && git pull && go build -o prodcal ./cmd/srv && sudo systemctl restart prodcal'`.

#### Implementation steps

1. **In `srv/books.go::runConversion()`**: after determining output dir, look up the book's project's book_spec; if present, render `config.typ` (same shape as handleGenerateConfig output, but written to a file) into the working dir; prepend `#import "config.typ": *` (or `merge-config`-pattern import) to generated `main.typ`. Fall back to default-config when no spec exists.
2. **New endpoint** `POST /api/projects/{id}/book-spec/compile` in `srv/bookspecs.go`: load spec → resolve linked book(s) → invoke runConversion() with PDF target → return JSON `{pdf_url, log}` (or stream the PDF). Mirror the existing generate-word-template endpoint's shape.
3. **Wire admin SPA**: in `srv/static/admin.html`, find the existing "Compile PDF" button in the Typesetting tab; change its handler to POST the new endpoint; surface log + download link in the existing output area.
4. **Refactor opportunity (optional, defer if time-pressed)**: extract config-rendering from `handleGenerateConfig` into a `renderTypstConfig(spec)` helper that both the existing endpoint and `runConversion()` call. Avoids duplication.
5. **Smoke test end-to-end on the VM**: pick a real book (Twitter Years has real content), edit spec in admin UI, click Compile PDF, download artifact, eyeball it. Check that spec changes (base size, leading, section break style) actually appear in the PDF.

#### Acceptance checks
- Spec change in admin UI → recompile → visible diff in PDF.
- Compile endpoint returns within 30s for a single-chapter book.
- Default-config fallback works for projects with no spec.
- No regression to existing `handleGenerateConfig` callers.
- Service restart is clean (no orphan race per TRK-OPS-005).

**Effort:** 3-4 hours. **Blocks:** TRK-DESIGN-001 (Ghosts parity needs working compile), TRK-TEST-001 (compile fixtures need working compile). **Fold in:** TRK-MIG-007 (verify/bundle Libertinus on VM during this session — `ssh exedev@jdbbs.exe.xyz 'fc-list | grep -i libertinus'`; if absent, `apt install fonts-libertinus`).

### TRK-DEV-004 — Special-typography preservation class (data model + preflight + pipeline)

- area: DEV
- status: open
- priority: P2
- created: 2026-05-25
- updated: 2026-05-25
- refs: docs/notes/2026-04-12-special-typography-preflight-and-ascii-preservation.md (original design discussion); docs/notes/2026-04-13-typst-pipeline-rewire-and-testing-retro.md (the Twitter Years QA session where this surfaced); docs/notes/2026-04-14-preflight-redesign-parked.md (preflight UI redesign that came out of the same period)

**Problem.** Layout-sensitive content (ASCII art, tweet snapshots, poem-like blocks, terminal transcripts, preserved-formatting reproductions) gets silently normalized somewhere between manuscript → output. During the April 2026 Twitter Years QA sessions this caused visible breakage in PDF output and remains an open semantic gap. The original note's framing is the right one: this isn't "bad spacing" — it's *intentional preformatted content* that needs to be declared and preserved.

**Why this matters now.** Twitter Years (project 7, book 6) is the first real book targeted at the v1 pipeline (TRK-DEV-002). A tweet-feed book is dense with this exact content class. If TRK-DEV-002's smoke test produces a PDF where tweet snapshots are mangled, the bug is here, not in the compile wiring.

**Scope — three layers:**

1. **Data model.** Add a structured "special typography" section to the transmittal (and surface it in book_specs). Per item:
   - label / short name
   - manuscript location or chapter
   - description of intent
   - expected treatment
   - Word style to apply
   - EPUB treatment (CSS class or block-level rule)
   - print treatment (Typst directive or named style)
   - notes / unresolved questions

2. **Preflight.** Extend the existing preflight (`scripts/detect-edge-cases.py`, redesigned 2026-04-14 per parked note) to output:
   - declared special-typography blocks (from transmittal/spec) — confirmed present in manuscript
   - observed layout-sensitive content NOT declared (suspicious — needs editorial review)
   - undeclared items block the compile until resolved or explicitly waived

3. **Pipeline preservation.** Both DOCX→Typst (via `docx-to-typst-enhanced.lua`) and DOCX→EPUB (via the spec→CSS path being built in TRK-DEV-003) must consume the declared-items list and emit the right wrapper (semantic Typst block / EPUB div with the right class). No silent normalization of whitespace, line breaks, or block boundaries inside a declared item.

**Approach (suggested phasing — each could be a separate sub-session):**

- **Phase A** — data model only: extend transmittal JSON schema; add UI section in `srv/static/transmittal.js`; ship without changing pipeline behavior. ~1 session.
- **Phase B** — preflight integration: declared items survive into the preflight report; undeclared layout-sensitive content gets flagged. ~1 session.
- **Phase C** — pipeline preservation in Typst path. Add Lua-filter handling for declared blocks; verify ASCII art and tweet snapshots round-trip. Should be done WITH or AFTER TRK-DEV-002 ships. ~1 session.
- **Phase D** — pipeline preservation in EPUB path. CSS classes + Pandoc filter. Should be done WITH or AFTER TRK-DEV-003 ships. ~1 session.

**Acceptance check** (Twitter Years smoke test):
- Transmittal has at least one declared special-typography item (e.g. "tweet feed snippet").
- Manuscript contains that item.
- Preflight confirms declared item present; no undeclared layout-sensitive content flagged.
- Compiled PDF preserves the item's spacing, line breaks, and block structure exactly.
- Compiled EPUB does the same (lossy-by-design only where reflow demands it; no silent whitespace collapse).

**Relationship to v2 (translation layer):** the same preservation contract will need to survive translation (the bilingual side-by-side format from TRK-TRANS-002 must keep declared blocks intact, and per-language Typst templates need lang-agnostic handling of preformatted blocks). Worth keeping in mind during Phase A's schema design — the spec for a declared item probably needs a `translatable: bool` flag so things like ASCII art (don't translate) are distinguished from tweet snapshots (translate the text inside, preserve the structure).

### TRK-DEV-003 — Wire spec → EPUB compile pipeline (CP-2)

- area: DEV
- status: done
- priority: P1
- created: 2026-05-25
- updated: 2026-05-26
- refs: srv/epub.go::handleGenerateEPUB (line 20), runEPUBGeneration (line 47); srv/static/admin.html:2637 (Generate EPUB handler); commit 56e8256

**Closed 2026-05-26.** Fresh-session audit (same pattern as TRK-DEV-002 closure) showed the seam was already wired end-to-end in a prior session, just not closed. Verified call trace:

- `srv/epub.go::handleGenerateEPUB` (line 20) → `runEPUBGeneration` (line 47).
- Line 75-93: when `book.ProjectID.Valid`, calls `q.GetBookSpec(ctx, book.ProjectID.Int64)` → `parseEPUBSpec(dbSpec.Data, book)` → `q.GetBookSpecCover` writes cover.{jpg,png} → `spec.CoverImage`.
- Line 97-101: `spec.buildCSS()` renders custom.css; passed as Pandoc `--css=` (overlay approach, resolves TRK-MIG-005).
- Line 108-127: spec values populate pandoc args — title / author / language / toc-depth / subject / description + `--epub-cover-image`.
- admin.html:2637-2667: "Generate EPUB" button calls `tsSaveSpec()` first then `POST /api/books/{id}/generate-epub`.
- admin.html:585-634: all `epub.*` spec fields (toc_depth, chapter_break, section_break, body_font_size, custom_css, language, subject, description) persist into `book_specs.data` via tsSaveSpec.

**Minor follow-up gaps (non-blocking, file if needed):** `epub.embed_fonts` is parsed but not acted on; `epub.landmarks` UI field has no spec-side consumer. Both surface in TRK-DESIGN-003 territory (CSS drift audit) — fold there.

### TRK-DEV-005 — Compile-history panel in admin SPA

- area: DEV
- status: done
- priority: P2
- created: 2026-05-26
- updated: 2026-05-26
- refs: commit 56e8256; db/queries/book_outputs.sql (ListBookOutputs, GetBookOutput); srv/books.go (bookAuth, handleListBookOutputs, handleDownloadBookOutput); srv/static/admin.html (ts-pdf-history, ts-epub-history panels)

**Done 2026-05-26.** Backend: `ListBookOutputs(book_id, limit=20)` returns metadata + `length(output_data) AS size_bytes` (no bytes on the list path); `GetBookOutput(id, book_id)` enforces ownership before streaming. New routes `GET /api/books/{id}/outputs` (?include=spec to fold in TRK-DEV-006 snapshot) and `GET /api/books/{id}/outputs/{output_id}/download` (per-output Content-Disposition uses the row's `created_at` so older artifacts download with their actual compile timestamp, not the latest). Auth via new `bookAuth` helper (project auth if linked, admin if unlinked) — mirrors `handleDownloadBook`. Admin SPA: small history panel under each Compile button; auto-refreshes on book select change + after each successful compile via existing polling loop.
- refs: db/migrations/014-book-output-history.sql, db/dbgen/book_outputs.sql.go (CreateBookOutput already populated per compile), srv/books.go:301 + srv/epub.go:146 (writers), srv/static/admin.html (Typesetting tab consumer)

**Context.** Every PDF/EPUB compile already inserts a `book_outputs` row with the artifact bytes, format, source_filename, and created_at — verified live 2026-05-25 against project 7. The archive exists; nothing surfaces it. Result: users (a) can't compare across compiles, (b) only ever see the latest artifact via `GET /api/books/{id}/download/{format}` (which returns whatever's in `books.pdf_data`/`epub_data`), and (c) have no way to retrieve a previous compile if a spec edit makes things worse.

**Scope.**

1. **Query:** add `ListBookOutputs(book_id, limit)` to db/queries/book_outputs.sql + regenerate dbgen. Returns id, format, source_filename, created_at, length(output_data). Index already exists (idx_book_outputs_book_format_created).
2. **Endpoint:** `GET /api/books/{id}/outputs` → JSON array of the above (no bytes). Same auth as the existing download.
3. **Per-output download:** `GET /api/books/{id}/outputs/{output_id}/download` → streams `book_outputs.output_data` for that row with the same Content-Disposition timestamp suffix the latest-download endpoint now uses (commit 792c76d). Confirm output_id belongs to the book before serving (path traversal hygiene).
4. **UI:** small panel under each Compile button in the Typesetting tab. Renders the last ~20 outputs as a list: timestamp · format · size · download link. Live-refresh after a successful compile completes (the existing polling loop already knows when status flips to ready).
5. **Retention:** for now, keep everything (the DB is 117 MB total — output bytes are the bulk of it but not yet a problem). When it does become a problem, add a TRK-OPS retention ticket (e.g. keep latest N per book + all marked-archived).

**Acceptance:** compile twice with different spec values; both PDFs appear in the panel; downloading the older row produces a PDF that matches the earlier spec, not the latest.

**Effort:** ~1 hour. **Pairs well with:** TRK-DEV-006 (spec snapshot per row makes the history actually diagnostic — without it you see two PDFs but can't tell what produced them).

### TRK-DEV-006 — Snapshot spec JSON into book_outputs per compile

- area: DEV
- status: done
- priority: P3
- created: 2026-05-26
- updated: 2026-05-26
- refs: commit 56e8256; db/migrations/015-book-output-spec-snapshot.sql; srv/books.go::buildTypstConfig (now returns (config, rawJSON)); srv/epub.go (captures dbSpec.Data → specSnapshot)

**Done 2026-05-26.** Migration 015 adds `book_outputs.spec_snapshot TEXT NULL` (legacy rows stay NULL). Both write sites (`runConversion` for PDF, `runEPUBGeneration` for EPUB) pass the raw `book_specs.data` JSON to `CreateBookOutput` via `nullStringFrom`. `buildTypstConfig` was refactored to return `(configString, rawJSON)` so the snapshot is captured from the same fetch the compile already uses — no double-read. `GET /api/books/{id}/outputs?include=spec` returns `spec_snapshot` per row for future diff-vs-latest UI (TRK-DEV-007 when warranted).

**Follow-up (defer):** **TRK-DEV-007** — diff-vs-latest UI rendering a minimal field-by-field spec comparison between consecutive compiles (e.g. `base_size_pt: 10 → 13, body_font unchanged`). API already returns the data; UI side only. File when it becomes the bottleneck.
- refs: db/migrations/014-book-output-history.sql (will need migration 015), srv/books.go::runConversion (write site), srv/bookspecs.go::specToTypstConfig

**Context.** TRK-DEV-005 surfaces the artifact archive. This ticket makes each archived artifact self-documenting: what spec values produced this PDF? Without it, when a user compiles 10 times across spec edits and one PDF looks right, they can't recover the spec that produced it.

**Scope.**

1. **Migration 015:** add `book_outputs.spec_snapshot TEXT NULL` (JSON blob copied from `book_specs.data` at compile time). NULL for legacy rows; new compiles populate it.
2. **Write site:** in `runConversion` (srv/books.go) and the EPUB equivalent (srv/epub.go), after the buildTypstConfig lookup, pass the raw spec JSON through to CreateBookOutput. May want a small helper that returns both `(configString, rawJSON)` to avoid double-fetching.
3. **API:** extend `GET /api/books/{id}/outputs` (TRK-DEV-005) to optionally include `spec_snapshot` when `?include=spec` is passed, so the UI can show a diff between consecutive compiles.
4. **UI (later):** "diff vs latest" link on each history row, rendering a minimal field-by-field diff (e.g. "base_size_pt: 10 → 13, body_font unchanged"). Out of scope for this ticket — file as TRK-DEV-007 when the basic snapshot lands.

**Acceptance:** new compiles persist spec; API returns it; legacy rows still listable with `spec_snapshot: null`.

**Effort:** ~1.5 hours. **Blocked by:** TRK-DEV-005 (no panel = nowhere to expose snapshot diff). **Related:** TRK-MIG-006 (corrections round-trip) — both want artifact lineage; consider whether `book_outputs` should also reference the correction_set_id that was active at compile time.

### TRK-DEV-007 — Diff-vs-previous UI in compile-history panel

- area: DEV
- status: done
- priority: P2
- created: 2026-05-26
- updated: 2026-05-28
- closed: 2026-05-28
- refs: srv/static/admin.html (compile-history panel under PDF + EPUB Compile buttons); `GET /api/books/{id}/outputs?include=spec,corrections` (already returns both snapshots); commits 56e8256 (DEV-005/006), 9af05ad-10a07c7 (MIG-006 corrections_snapshot)

**Done 2026-05-28.** JS-only landing in `srv/static/admin.html`. `tsRenderHistory` now fetches with `?include=spec,corrections` and caches each panel's rows in `tsHistoryRows[el.id]`. Each row except the oldest visible gets a `diff` link that toggles an inline two-column panel (spec diff left, corrections diff right). Helpers: `tsSpecDiff` walks two JSON objects recursively emitting `{path, before, after}` for changed leaves with stable sort; `tsParseCorrectionsYAML` parses the `%q`-quoted YAML emitted by `renderCorrectionsYAML` via `JSON.parse('"' + raw + '"')`; `tsCorrectionsDiff` keys on `find` and returns `{added, removed, changed}` with per-field deltas. Empty-diff state renders explicit copy ("no spec changes" / "no corrections changes" / "no spec or corrections changes since … (no snapshots recorded)"). Legacy pre-snapshot rows: diff link greys out via opacity and the renderer emits "earlier compile had no spec snapshot" rather than throwing. Unit-tested the three helpers under node before pushing. No new API, no Go changes, no schema change.

**Context.** Both `spec_snapshot` (DEV-006, migration 015) and `corrections_snapshot` (MIG-006, migration 016) are now persisted per compile. The history panel lists rows + lets you download artifacts but nothing surfaces *what changed* between two compiles. Right now: user compiles twice, downloads both PDFs, opens them side-by-side, tries to remember which spec edits and which corrections were active. The data exists; only the UI gap is left.

**Scope.**

1. **API:** no new endpoint needed. `GET /api/books/{id}/outputs?include=spec,corrections` already returns the data per row.
2. **UI helper (JS):** small diff function that compares two JSON objects field-by-field and emits a flat list of `{ path, before, after }` for changed fields. Spec diff = nested JSON walk (typography.base_size_pt: 10 → 12). Corrections diff = set-difference on YAML-parsed entries (added/removed/modified rows).
3. **History panel additions:**
   - Each row in the panel gets a "diff" button (except the oldest visible row).
   - Click → modal or expand-in-place showing two columns: spec diff (left), corrections diff (right). "Compared to compile from <earlier timestamp>".
   - Default diff target is the immediately-preceding row (chronologically). Optional: a dropdown to compare against any other row in the panel.
   - Empty diff renders as "no spec or corrections changes since [timestamp]" — distinguishes "we compiled twice with same inputs" (a thing to know) from "we never recorded a snapshot" (legacy rows from before migration 015/016).
   - Legacy rows with `spec_snapshot: null` show "(no snapshot recorded)" in the diff target dropdown and disable the diff button when they're the target.

**Format suggestion** for the field-list diff:

```
typography.base_size_pt:  10 → 12
typography.body_font:     "Libertinus Serif" → "Plantin MT Pro"
running_heads.verso:      "author" → "title"
```

For corrections:

```
+ added:   { find: "Venkatesh", replace: "Venkat", status: "pending" }
- removed: { find: "iphone", replace: "iPhone", status: "applied" }
~ changed: { find: "alchemy" } status: "pending" → "applied"
```

**Acceptance:**

- After two compiles with different spec values, the diff button on the newer row shows the field changes.
- After two compiles with different corrections (one added between compiles), the diff shows the added correction.
- An empty diff is rendered explicitly, not as a blank panel.
- Legacy pre-snapshot rows are handled gracefully.

**Effort:** ~1-2 hours. **Pairs well with:** TRK-DEV-008 item 4 (surface patcher warnings in the panel) — same UI zone, but distinct enough to file as separate tickets; do them in separate sessions to keep merge surface clean. **Don't bundle with:** TRK-DEV-004 (special-typography) or any other admin.html-heavy work in the same session.

### TRK-DEV-009 — Per-chapter author in EPUB spec + pipeline (anthology critical-path)

- area: DEV
- status: done
- priority: P1
- created: 2026-05-26
- updated: 2026-05-29
- closed: 2026-05-29
- refs: `docs/GHOSTS_PARITY_2026-05-26.md` §"Single Biggest Gap"; `srv/epub.go::handleGenerateEPUB` (`buildEPUBMetadata`, line ~85+); `srv/bookspecs.go` (`parseEPUBSpec`); `manuscripts/ghosts/main.typ` (Typst-side reference: `set-story-info(title:, author:)` per chapter); `typesetting/templates/series-template.typ` (per-chapter rendering already supported); `db/migrations/008-book-specs.sql` (where `book_specs.data` schema lives)

**The single biggest gap from the Ghosts parity audit.**

**Current state:**
- Typst pipeline: ✅ anthology support working. `series-template.typ` exposes `set-story-info(title:, author:)`; `manuscripts/ghosts/main.typ` configures 9 different authors across 9 chapters; running headers correctly state-track per-chapter author (verso) and title (recto).
- EPUB pipeline: ❌ broken for anthologies. `book_specs.data.epub` schema has only a singular `author` field; pandoc receives one `--metadata=author=…` value for the whole book; chapter divisions are preserved but per-chapter authorship is lost.

**Impact:** Any multi-author book (Ghosts, Librarians, future series titles) ships an EPUB with one author across all chapters. Twitter Years (single author) ships fine. The first time the EPUB fails is the first multi-author book to ship.

**Approach:**

1. **Schema extension.** Add `chapters` array to `book_specs.data.epub`:
   ```jsonc
   "epub": {
     ...existing fields...,
     "chapters": [
       { "title": "Soda Sweet as Blood", "author": "Spencer Nitkey", "file": "01-soda.md" },
       { "title": "In Every Lifetime", "author": "Lara Dal Molin", "file": "02-lifetime.md" },
       ...
     ]
   }
   ```
   No migration needed — `book_specs.data` is JSON, schema-flexible. UI exposes it as a repeating row editor in the Typesetting tab's EPUB section.

2. **Population sources.** Three ways the chapters array gets filled:
   - **Manual entry in admin SPA.** Always available.
   - **Pull from transmittal.** Extend `handlePullTransmittalToSpec` to map `transmittal.chapters[*]` → `spec.epub.chapters[*]` if the transmittal has structured chapter data.
   - **Auto-detect from manuscript.** Lower priority; would scan uploaded DOCX heading-style runs + paragraphs following them. Defer to a follow-up ticket.

3. **EPUB generation wiring.** Two integration points in `srv/epub.go`:
   - **Pandoc metadata:** pandoc supports per-chapter `<h1>` author via custom Lua filter or via XHTML postprocessing. Cleanest: emit chapters as separate XHTML files via pandoc's `--split-level=1` (already happens) and inject a `<p class="chapter-author">{{author}}</p>` block per chapter from the spec.
   - **EPUB OPF metadata:** the EPUB spec doesn't natively support per-chapter authors at the package-metadata level, but most readers honor in-content `<p class="chapter-author">` styling. The existing `.chapter-author` CSS class (epub-styles.css lines 102-111) handles rendering — only the *content injection* is missing.

4. **UI in admin SPA.** Add a "Chapters" subsection under EPUB settings in the Typesetting tab. Reuse the corrections-table pattern (add/remove rows; per-row title + author + optional file field). Auto-save like other spec fields.

5. **Smoke test:** Upload Ghosts DOCX as a new project (id != 7, won't conflict with Twitter Years); configure 9 chapters with authors; compile EPUB; verify per-chapter bylines render in Calibre or epubcheck.

**Acceptance:**
- A book with `spec.epub.chapters` populated produces an EPUB with each chapter showing its own author byline as a `.chapter-author` paragraph.
- A book with empty `chapters` (or pre-DEV-009 spec) falls back to current behavior: book-level `spec.author` for the whole EPUB.
- Round-trip with Ghosts source produces an EPUB whose chapter-by-chapter author bylines match the reference GHOSTS.epub.

**Effort:** ~3-4 hours. **Blocks:** TRK-TEST-002 visual regression (EPUB diff is meaningless without this). **Related:** TRK-DEV-004 (anthology-aware spec is also where special-typography per-chapter declarations might live — schema design should consider how they'd nest).

**Closed 2026-05-29.** Implemented all three layers (schema → backend → UI). No migration: `book_specs.data.epub.chapters` is just a new JSON field; empty/missing array preserves today's single-author behavior exactly (Twitter Years untouched).

**Call trace:**
- `srv/epub.go`: added `epubChapter` struct + `Chapters` slice on `epubSpec`. `parseEPUBSpec` reads `data.epub.chapters[]` and drops fully-empty rows. After `pandoc` completes and the EPUB is read, if `len(spec.Chapters) > 0`, `injectChapterAuthors` rewrites the zip: for each XHTML entry (skipping `nav.xhtml`/`toc.xhtml`/`title_page.xhtml`/`cover.xhtml`) that contains an `<h1>`, splice `<p class="chapter-author">{author}</p>` immediately after the first `</h1>`. Match is by *spine order* (zip order), one-to-one with `spec.Chapters[i]`. Surplus entries on either side are left alone. `mimetype` stays first and `zip.Store`d. Bylines XML-escape author names. A default `.chapter-author` CSS rule is emitted only when chapters are configured.
- `srv/epub_chapter_test.go`: five unit tests covering the happy path, no-chapters no-op, nav-skip, XML escaping, and `mimetype`-first invariant. Plus a `parseEPUBSpec` test for the new `chapters` field.
- `srv/static/admin.html`: new "Chapters (anthology bylines)" subsection in the EPUB card, above "Layout". Repeating-row editor (title / author / file) following the `tsRenderCustomStyles` pattern: `tsRenderEpubChapters` renders + wires input/remove handlers; "+ Add chapter" appends and re-renders. `tsPopulateForm` calls the render. Auto-save piggybacks on `tsMarkDirty` → debounced `tsSaveSpec` (no separate endpoint).

**Non-goals deferred (per session prompt):** font bundling (TRK-DESIGN-004, concurrent); smart-punctuation (TRK-DESIGN-003); live visual regression with Ghosts manuscript (TRK-TEST-002); auto-detecting chapters from DOCX heading runs; transmittal → spec.epub.chapters mapping (the `handlePullTransmittalToSpec` extension noted in §2 above remains open as a follow-up if/when transmittals start carrying structured chapter data).

**Smoke validation:** Go isn't installed on the Mac dev box; correctness validation is unit-test coverage + post-deploy `go build` on the VM. Visual confirmation against `reference/GHOSTS.epub` deferred to TRK-TEST-002 (which is the natural follow-up now that this is in).

### TRK-DEV-010 — Wire `--epub-embed-font` into pandoc invocation for Noto CJK/Thai bundle

- area: DEV
- status: open
- priority: P2
- created: 2026-05-25
- updated: 2026-05-25
- refs: TRK-DESIGN-004 (parent — bundled the fonts but deferred this wiring to avoid concurrent-edit conflict with TRK-DEV-009); `srv/epub.go` line ~144; `typesetting/fonts/noto/`

**Problem.** TRK-DESIGN-004 landed Noto Serif TC + Noto Serif Thai under `typesetting/fonts/noto/` and added `@font-face url(NotoSerifTC-Regular.otf)` etc. to `epub-styles.css`. The CSS expects the font files to live inside the generated EPUB package (pandoc's convention is to reference by basename when `--epub-embed-font` packages them). Without the flag, the `@font-face` declarations point to missing files inside the EPUB and readers fall through to the family-stack OS fallbacks (Hiragino on macOS, system Noto if installed on Linux, possibly tofu otherwise).

**Action.** Add four flags to the pandoc `args` slice in `srv/epub.go::handleGenerateEPUB`, near where `--css=` is appended (~line 144):

```go
args = append(args,
    "--epub-embed-font=typesetting/fonts/noto/CJK-TC/NotoSerifTC-Regular.otf",
    "--epub-embed-font=typesetting/fonts/noto/CJK-TC/NotoSerifTC-Bold.otf",
    "--epub-embed-font=typesetting/fonts/noto/Thai/NotoSerifThai-Regular.ttf",
    "--epub-embed-font=typesetting/fonts/noto/Thai/NotoSerifThai-Bold.ttf",
)
```

Verify path resolution — pandoc resolves relative to `cmd.Dir` (currently `tmpDir`). May need absolute paths via `filepath.Join(repoRoot, ...)` or chdir to repo root. Smoke: compile any book to EPUB, unzip and confirm `EPUB/fonts/*.otf` are present and CSS resolves.

**Effort:** ~15 min including smoke. **Pickup:** as soon as TRK-DEV-009 lands (concurrent file). **Closes:** the "Deferred" item in TRK-DESIGN-004.

### TRK-DEV-011 — Admin SPA UX polish (project workflow gaps)

- area: DEV
- status: open
- priority: P3
- created: 2026-05-26
- updated: 2026-05-26
- refs: surfaced during 2026-05-26 user setup for TEST-002 (Ghosts project creation walkthrough)

User feedback while creating a new project + uploading manuscript revealed three workflow gaps. Filed as one ticket since they're all admin SPA polish surfaced in the same session; pick when warranted, not as a batch.

1. **Calendar widget on New Project form.** The new-project form (where you name the project + assign client) should include a date-picker for project start date, not require editing it after the fact from the project detail view. Currently you create the project, then navigate into it to set scheduling. Minor friction; common path.

2. **No manuscript upload on the transmittal/project detail page.** Manuscript uploads live on the top-level `Files` tab (`srv/static/admin.html` line 450, `tab-files`); the project-detail/transmittal view at `/pi/{slug}` has no upload affordance. The user reasonably expects to upload-while-in-context-of-the-project. Options:
   - Add a small "Upload manuscript" affordance on the project detail page that pre-fills `project_id` and links into the Files workflow.
   - Or: surface "Books in this project" on the project detail page with an upload-here button.
   - Either way, current state forces a back-out-to-top-level → select-project-in-dropdown → upload flow that's easy to miss when new to the SPA.

3. **Transmittal Date field doesn't propagate to Calendar tab.** Setting "Transmittal Date" in either Book Information or Production sections of the transmittal does not appear in the project's Calendar tab. Also: there are TWO Transmittal Date fields on the transmittal page (one under Book Information, one under Production) — unclear which is canonical, and neither hooks into Calendar. Investigate which field the Calendar tab reads from (or whether it reads at all); fix the wire-up; consider collapsing the two fields if they represent the same concept.

**Effort estimates:** item 1 ~30 min, item 2 ~1-2 hours (depends on how much existing Files-tab plumbing is reused), item 3 ~30-60 min (investigation + fix). None are blocking the v1 release-confidence track; pick as ergonomics warrant.

### TRK-DEV-008 — Corrections patcher ergonomics

- area: DEV
- status: open
- priority: P3
- created: 2026-05-26
- updated: 2026-05-26
- refs: typesetting/scripts/apply-corrections-docx.py; srv/corrections_apply.go; srv/corrections.go (CRUD + status); admin SPA corrections ledger

Follow-ups surfaced by the TRK-MIG-006 round-trip. None is blocking — the patcher is correct and propagates pending corrections to PDF + EPUB. These are usability gaps that show up the moment a real correction set has any variety in it.

**Candidate improvements (pick when warranted, not as a batch):**

1. **Case-insensitive flag per correction.** Today matching is pure `str.find()` — `alchemy` won't catch `Alchemy` and you need a second row. Add an optional `case_insensitive: true` to the YAML schema + a `case_insensitive INTEGER NOT NULL DEFAULT 0` column to the `corrections` table. The script keeps the default behavior (the `iphone → iPhone` example in the docstring relies on case-sensitivity) and only loosens when the flag is set.
2. **Whole-word matching flag.** Same shape — `whole_word: true` so `alchemy → al-TEST` doesn't accidentally rewrite a substring like `alchemystudio`. Regex with `\b…\b`.
3. **Per-scope filters.** Some fixes are footnote-only ("citation typo"), some metadata-only. Optional `scope:` field (`body | footnote | metadata | all`) on each correction row. Default `all` preserves today's behavior.
4. **Surface patcher warnings in the SPA.** Today `WARN corrections: patcher failed; compiling from unpatched source` only shows in `journalctl`. Stash the last patcher stderr (or a count of corrections that returned NOT FOUND) on the compile's `book_outputs` row and render a yellow badge in the history panel. Pairs with TRK-DEV-007 (diff-vs-latest UI).
5. **Dry-run preview in the ledger UI.** Add a "preview" button per correction row that invokes the patcher with `--dry-run` against the latest docx and shows match count + first 3 contexts. Catches typos in the `find:` field before they ship a no-op into the next compile.
6. **`corrections_snapshot` diff in the history panel.** API already returns `corrections_snapshot` via `?include=corrections` (DEV-006 shape). UI side only: show which corrections were active for each compile.
7. **Reconsider auto-marking applied.** Current contract: status flips stay manual because every compile re-applies all pending entries from the original docx, so auto-flipping would silently drop the fix on the next compile. *If* the data model later captures "version of source docx that already includes this fix" (e.g. a `applied_to_source_revision` field), auto-flip becomes safe — but only then.
8. **Patch the source docx in DB once the canonical Word doc is re-uploaded.** Adjacent to (7): when a new manuscript is uploaded that already contains the fix, allow batch-marking applied via a UI affordance ("This upload includes corrections #3, #7, #12 — mark applied?").

**Acceptance per item:** each lands as its own small change — case-insensitive is ~20 lines of script + a migration; the SPA additions are larger and probably wait for a real production correction set.

**Effort:** items 1–2 are ~30 min each; items 3–6 are 1–2 hours each; items 7–8 need data-model thought first.

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
- updated: 2026-05-26
- refs: TRK-DESIGN-001, `docs/GHOSTS_PARITY_2026-05-26.md` (closes 10 ⚠️ cells); `manuscripts/ghosts/main.typ` + chapter `.typ` files; `reference/GHOSTS.pdf` + `reference/GHOSTS.epub`; `typesetting/scripts/build-ghosts.sh`

**Scope.** Live-compile Ghosts end-to-end (Typst + EPUB) and page-by-page diff vs reference. The DESIGN-001 audit established what the code does; this ticket establishes what the rendered output actually looks like.

**Approach:**

1. **Typst PDF compile.** On VM: `cd /home/exedev/prodcal/manuscripts/ghosts && typst compile --font-path ../../typesetting/fonts main.typ /tmp/ghosts-typst.pdf` (or invoke via `typesetting/scripts/build-ghosts.sh` which already wires the right paths). Expect 136-ish pages.
2. **EPUB compile.** Two options:
   - Direct: pandoc the same source chapters with the bundled EPUB stylesheet.
   - Via admin SPA: upload Ghosts DOCX as a project, configure spec to mark it anthology (post-TRK-DEV-009 only), compile EPUB. This is the "real" path but blocked on DEV-009 for fair per-chapter-author comparison.
3. **PDF page-by-page diff.** `pdftoppm` both PDFs to PNG @ 150 dpi; ImageMagick `compare -metric AE -fuzz 5%` per page; emit a per-page-delta CSV. Threshold for justified-text noise: tune empirically (start at AE < 0.5% of pixels).
4. **EPUB diff.** `epubcheck` for structural validity. Unzip both, diff per-chapter XHTML (semantic: chapter title, author byline, body presence). Render EPUB to PDF via Calibre (`ebook-convert`) for visual diff.
5. **Close the 10 ⚠️ cells.** Each gets ✅ or ❌ with evidence (per-page deltas, screenshots, observations). ❌ cells spawn child tickets.

**Pre-requisites (must ship before this is meaningful):**
- **TRK-DEV-009** — per-chapter EPUB author. Without this, EPUB diff will show one author across all chapters; can't validate anthology behavior.
- **TRK-DESIGN-004** — CJK/Thai font availability on compile host. Without bundled or installed fonts, glyphs will render as tofu and inflate the diff metric.
- **TRK-DESIGN-002** — Plantin MT Pro + Proxima Nova bundling (user has licenses). With open-source subs (Libertinus + Source Sans 3), metric/visual differences inflate the diff. Could be done after a first pass-with-subs to get a baseline.

**Sample reference PNGs already exist:** `reference/pdf_pages/` (or equivalent — see GHOSTS_PARITY doc for the actual file inventory the subagent found). Extend to all 136 pages.

**Effort:** ~2-3 hours when prerequisites are met. Probably best done as a dedicated session after DEV-009 + DESIGN-004 + DESIGN-003 land.

### TRK-TEST-003 — VM smoke script + cron

- area: TEST / OPS
- status: open
- priority: P2
- created: 2026-05-12
- updated: 2026-05-12

`scripts/smoke.sh` on VM: one DOCX→PDF + one MD→PDF + one MD→EPUB + one corrections apply; exits nonzero on failure. Cron daily, send email on failure (AgentMail is already wired).

---

## Translation (TRANS) — v2

Translation-layer tickets per `docs/PRODUCTION_ROADMAP_2026-05-25.md` §"v2". All P3 (post-v1). Source design discussion: `docs/translation layer 2026-05-25.md`. Recommended architecture: Variant E (orchestrator) + Variant F (cross-lingual consistency). Skip Variant G (MT-with-rubber-stamp).

### TRK-TRANS-001 — Per-title manifest schema (translation section)

- area: TRANS
- status: open
- priority: P3
- created: 2026-05-25
- updated: 2026-05-25

Design `manifest.yaml` translations section: `translations.{lang}.{status, translator, reviewer, glossary, style_guide, mt_engine, llm_pass, target_pub_date, locked_at}`. Treat `zh-Hans` and `zh-Hant` as separate. Document es-ES vs es-419 policy. Add `translations` DB table linked to `book_id`. ~1 session.

### TRK-TRANS-002 — Bilingual side-by-side output format prototype

- area: TRANS
- status: open
- priority: P3
- created: 2026-05-25
- updated: 2026-05-25

Per the source doc: the format translators receive determines whether they tolerate the workflow or route around it. Prototype Markdown table layout vs XLIFF; ship to one real translator for feedback before locking in. ~1 session for prototype, ~1 for refinement.

### TRK-TRANS-003 — MT + LLM draft pipeline (fr first)

- area: TRANS
- status: open
- priority: P3
- created: 2026-05-25
- updated: 2026-05-25
- blocked-by: TRK-TRANS-001, TRK-TRANS-002

Pick `fr` first (best MT pair, easiest reviewer recruitment). End-to-end: chapter-aware chunker, glossary loader, DeepL call, Claude pass with glossary + style guide + prior-chunk context, bilingual output writer. ~2 sessions per language.

### TRK-TRANS-004 — Per-language Typst templates

- area: TRANS
- status: open
- priority: P3
- created: 2026-05-25
- updated: 2026-05-25

`series-template-fr.typ` (nbsp before `;:?!`, guillemets, em-dash dialogue, hyphenation, accented capitals). Then `-es.typ` (¡¿, dialogue dashes), then `-zh-Hans.typ` (CJK fonts, fullwidth punctuation, line-breaking). Each ~1 session; zh is much larger.

### TRK-TRANS-005 — Per-language EPUB stylesheets

- area: TRANS
- status: open
- priority: P3
- created: 2026-05-25
- updated: 2026-05-25

`epub-styles-{fr,es,zh-Hans}.css` with appropriate font stacks (Source Han Serif / Noto Serif CJK for zh) and `<html lang="…">` tagging. ~1 session per language.

### TRK-TRANS-006 — Per-language automated validators

- area: TRANS
- status: open
- priority: P3
- created: 2026-05-25
- updated: 2026-05-25

Regex-based checks (not LLM job): French bare `:;?!` without nbsp; Spanish `?` without leading `¿`; Chinese halfwidth punctuation in CJK runs. Run as part of compile. ~1 session.

### TRK-TRANS-007 — ISBN/ONIX/cover registry per language

- area: TRANS
- status: open
- priority: P3
- created: 2026-05-25
- updated: 2026-05-25

At 20-50 books × 3 langs × 2 formats = 60-200 ISBNs/year + ONIX records. DB schema, admin UI, ONIX generator per distributor. ~2-3 sessions; meaningful design surface.

### TRK-TRANS-008 — Cross-lingual glossary system

- area: TRANS
- status: open
- priority: P3
- created: 2026-05-25
- updated: 2026-05-25

Per-series glossary in git (`glossaries/{series}-{lang}.yaml`). Hermes/orchestrator commits updates with translator attribution. Backlist re-translation flagging. Terminology drift detection across volumes. ~2 sessions.

### TRK-TRANS-009 — Translation orchestrator agent

- area: TRANS
- status: open
- priority: P3
- created: 2026-05-25
- updated: 2026-05-25

The hermes-style coordinator: spawns translation jobs on `ms:final`, posts to Slack on draft-ready ("87k words, ~40-60hr review, side-by-side at translations/fr/draft-v1.md"), diffs MTPE vs LLM draft (flags hotspots for native reviewer), weekly digest cron. Significant agent-design work — start with a separate planning session before any code.

### TRK-TRANS-OQ — Open questions before v2 starts

- area: TRANS
- status: open
- priority: P2 (decision before any v2 build work)
- created: 2026-05-25
- updated: 2026-05-25

Source doc identifies these as load-bearing before building:
1. **Source-author contracts** — verify they permit MT-then-edit workflow. Older contracts often silent; newer ones may explicitly prohibit. Check before building, not after.
2. **MTPE industry politics** — IAPTI, European/national translator orgs have evolving positions and rates. Talk to your translators about how they want to work; don't impose MTPE top-down.
3. **Chinese MT engine choice** — DeepL/Google/Baidu/DeepSeek/Qwen — get zh translators to A/B before locking in.
4. **Language priority** — which target first? Doc recommends `fr`; user has not chosen.

---

## Decisions log

(Append-only. Each decision is dated, summarized, and refs the entries it locks down.)

- **2026-05-12** — Tracker lives in `jdbbs/TRACKER.md` (single source of truth), even before the VM cutover.
- **2026-05-12** — Commercial fonts (Plantin MT Pro, Proxima Nova) will be bundled in `typesetting/fonts/` with license docs (per user). Open substitutes remain as fallback.
- **2026-05-12** — **Strategy reversal**: reconcile in `prodcal` (the live, substantive repo), not in `jdbbs`. The VM's prodcal has ~30 commits + significant uncommitted work that's missing from jdbbs (including security hardening commit `3c2256d`, preflight system, email notifications, archive lifecycle, custom-style hardening, API docs). jdbbs has only 3 substantive commits worth porting (typesetting subdir import, Phase 2 path fixes, Phase 3.1 trim registry). Doing the reconciliation in prodcal is strictly fewer ports. Final rename `prodcal → jdbbs` is deferred (TRK-MIG-009).
- **2026-05-12** — Cutover style: "just do it" — minutes of downtime acceptable, deploy and fix-forward, no parallel staging.
- **2026-05-25** — Single canonical repo: prodcal and book-prod archived, jdbbs becomes the only live repo (TRK-MIG-009 done). VM dir/service rename deferred (TRK-OPS-008).
- **2026-05-25** — v1 critical path consolidated in `docs/PRODUCTION_ROADMAP_2026-05-25.md`. The three older planning docs (TEST_PLAN_2026-03-09, TYPOGRAPHY_REFINEMENT_PROMPT, TYPST_FRONTEND_PLAN) are superseded by it. Originals preserved in `book-prod-archived-2026-05-25/`.
- **2026-05-25** — Translation layer architecture: Variant E (orchestrator) + Variant F (cross-lingual consistency). Skip Variant G (MT-with-rubber-stamp). Tracked as TRK-TRANS-001..009 (v2, all P3).
- **2026-05-25** — Working notes moved from repo root to `docs/` for tidiness (TRACKER.md, MIGRATION_LOG.md, NEXT_SESSION_PROMPT_*.md, etc.). `scripts/jpull.sh` updated to read from either location.
- **2026-05-26** — Orchestrator fan-out validated: three concurrent workstreams (DEV-007 in a fresh Claude Code session; OPS-006 + DESIGN-001 audit in the orchestrator session via direct shell + Explore subagent) landed cleanly. Coordination point: shared working tree on one Mac means parallel sessions can't simultaneously modify the same file; TRACKER.md is the recurring contention surface. Resolution: orchestrator defers TRACKER edits until parallel sessions commit, then applies its updates on top.
- **2026-05-26** — Ghosts parity audit defines the v1 release-confidence track. Critical-path tickets surfaced: TRK-DEV-009 (per-chapter EPUB author), TRK-DESIGN-004 (CJK/Thai fonts). TRK-TEST-002 (live visual regression) gates final close of TRK-DESIGN-001.
- **2026-05-26** — SQLite gotcha confirmed: `PRAGMA foreign_keys = ON;` must be set explicitly per connection or FK cascades silently no-op. Affects any future bulk-cleanup work and any code that relies on cascades.
