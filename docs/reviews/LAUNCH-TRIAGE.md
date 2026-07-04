# ProdCal Pre-Launch Review — Consolidated Triage

Date: 2026-07-04 · Tree: `aad23e1` (merged VM+GitHub) · Method: 7 parallel review agents (backend, data, email, pipeline, 2× UX, tests/ops) + live browser/curl verification against a local server (fresh DB, migrations 1–16, email off).

**Coverage caveats (read first):**
- **BackendCore** and the dedicated **SecuritySweep** agents had their final output killed by the model provider's content filter (security-heavy text). Their load-bearing findings were **recovered from transcript + reconstructed and re-verified live by Main** — nothing critical was lost (the auth blocker is triple-corroborated). BackendCore reportedly had 3 BLOCKER / 7 HIGH; the extras beyond what's listed here could not be recovered and warrant a targeted re-review.
- **BookPipeline** crashed after reading the code but before producing findings. The shell-out surface was spot-checked by Main (all `exec.Command` use separate-arg form → **no classic shell injection**; uploads bounded at 50 MB / 10 MB). **A focused re-review of `epub.go` zip handling (zip-slip), temp-file cleanup/collisions, and corrections-YAML safety is still owed.**

Severity legend: 🔴 **BLOCKER** (breaks launch / data / security) · 🟡 **HIGH** (fix before launch) · 🟢 **QOL** (workflow win, post-launch OK) · ⚪ **NIT**.

---

## 🔴 Launch blockers — must fix before shipping

### B1. Client password gate is decorative — project & transmittal content is open and editable without it
`srv/server.go:361-397` (`checkAuth`). Returns `true` whenever a project has **zero project-level tokens** (`if err != nil || len(tokens) == 0 { return true }`) — it never falls back to requiring the *client* password. Since neither admin nor portal project-creation mints project tokens, the default state of a password-protected client is: portal listing gated, but anyone with the URL `/{client}/{project}/` or `…/transmittal/` **skips the gate and gets fully editable** transmittal (autosave PUT, version restore, Mark Final) and calendar (task CRUD). `srv/journal.go:157` → `srv/client.go:34` compounds it: any token-less sibling project unlocks the protected client's file-log, journal, and **digest-email** endpoints for anonymous callers.
- **Verified live:** `/empty/` (a passwordless client) renders its full portal — including the Weekly-Digest trigger — with no gate at all.
- **Corroborated by:** ClientPortalUX (BLOCKER), BackendCore, EmailSystem (HIGH), DataLayer.
- **Fix:** in `checkAuth`, when the project's client has a `password_hash`, require the client cookie even if the project has no tokens; include client-level auth in `handleGetProjectByPath.has_auth` so the SPAs show the password screen. Also fix `hasAnyProjectAuthForClient` to skip token-less projects.

### B2. SQLite pragmas are per-connection → FK enforcement OFF + `SQLITE_BUSY` 500s under load
`db/db.go:28-39`; no `SetMaxOpenConns` anywhere. `foreign_keys=ON` and `busy_timeout=1000` are set via `db.Exec` on one pooled connection; they're **per-connection** and the pool is uncapped, so fresh connections run with SQLite defaults (`foreign_keys=OFF`, `busy_timeout=0`). Verified against modernc.org/sqlite v1.39.0 source (DSN `_pragma=` is the only reliable channel).
- **Blast radius:** every `ON DELETE CASCADE`/`SET NULL` silently unenforced on most connections (orphaned `book_outputs` multi-MB blobs on book delete); two concurrent writes → second gets `SQLITE_BUSY` immediately → handler 500. Most likely trigger: blob-heavy book convert/EPUB write overlapping any save.
- **Fix:** `sql.Open("sqlite", path+"?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)&_pragma=journal_mode(wal)")` **and** `db.SetMaxOpenConns(1)` (also makes the `:memory:` test DB safe). — DataLayer (BLOCKER), BackendCore.

### B3. Disaster-recovery runbook corrupts the database
`DEPLOY.md:75-82`. Restore steps `sudo systemctl stop srv` (**wrong unit** — prod is `prodcal.service`, so the stop fails and the server keeps running), then `cp` a backup over `db.sqlite3` **while the live WAL server still holds it** → torn state vs. the attached `-wal`/`-shm`, corrupting the freshly restored DB. `gunzip` also consumes the only local backup. The single documented DR path is actively dangerous.
- **Fix:** `systemctl stop prodcal.service` → `zcat`/`gunzip -k` to a temp file → `mv` into place → `rm` stale `db.sqlite3-wal`/`-shm` → start → check `/healthz`. — TestsAndOps (BLOCKER), DataLayer.

### B4. Typesetting spec autosave: switching project discards edits and can overwrite the wrong project's spec
`srv/static/admin.html:1814-1829` (project-switch handler), `2131-2137` (2 s debounce), `2139-2165` (save), `1830-1857` (load). The `tsProject` change handler calls `tsLoadSpec(pid)` without flushing/cancelling the pending autosave, and load unconditionally sets `tsDirty=false`. Two failure modes: (1) **certain data loss** — edit A, switch to B within 2 s → A's edits silently gone; (2) **cross-project overwrite race** — if the debounce fires after `tsProject.value` changes but before B's fetch resolves, A's edited spec is PUT to `/api/projects/B/book-spec`, wholesale replacing B's spec (book specs have **no version history**).
- **Fix:** in the change handler `clearTimeout(tsAutoSaveTimer)`; if dirty, `await` a save bound to the OLD pid before loading the new one. — AdminWorkflowUX (BLOCKER).

### B5. Go toolchain has stdlib vulns reachable from the running server
govulncheck (run by Main): **GO-2026-5039** (`net/textproto`) and **GO-2026-5037** (`crypto/x509`), both reached from `http.Serve` (`srv/server.go:257`), fixed in **go1.26.4**; `go.mod` pins `go 1.26.0`.
- **Fix:** bump the Go toolchain to ≥1.26.4 and rebuild. (Separately, the Dependabot "1 moderate" = `golang.org/x/net v0.48.0` < `v0.55.0`, pulled only via the `sqlc` tool directive — **not** in the server call graph; `go get golang.org/x/net@v0.55.0 && go mod tidy` silences GitHub but is lower priority than the toolchain bump.) — Main (govulncheck), TestsAndOps.

> **Near-blocker:** **Client cookie stores the plaintext password** for 90 days, HttpOnly but **no `Secure` flag** (`srv/client.go:120-133`). Any log/proxy/backup that captures cookies captures client passwords. Fix with a random session token + `Secure:true`. Filed HIGH below; treat as a blocker if any intermediary logs cookies.

---

## 🟡 HIGH — fix before launch

**Auth / security**
- **H1. Plaintext password in client cookie** — `srv/client.go:120-133`; see near-blocker above.
- **H2. Wrong client code renders a convincing fake empty portal** — `srv/static/client.html` `boot()` catch (~916) sets `state.error` but the portal view never shows it; a typo'd code shows a dashboard titled with the typo, zero stats, "No projects yet." A real client concludes their projects were deleted. Verified: `/api/clients/{typo}` → 404, portal still renders. Fix: render an explicit "we don't recognize that client code" state.
- **H3. Migrations are non-transactional and non-idempotent** — `db/db.go:106-114`; a mid-file failure leaves partial DDL unrecorded, so re-run hits `duplicate column`/missing-table and **the server never boots** without manual DB surgery (worst case: 013's DROP/RENAME rebuild). Fix: `BEGIN IMMEDIATE…COMMIT` per file + runner-side recording; `IF NOT EXISTS` on 010's index.

**Data / backups**
- **H4. Failed-probe backups poison the whole chain** — `scripts/backup-db.sh`; rowcount/integrity/threshold failures `die` but leave the bad `.gz` in `BACKUP_DIR`, which sync-to-r2 ships, restore-drill drills, prune counts, and backup_status reports as newest. Fix: `mv` bad artifacts to `.BAD` in `die()`.
- **H5. Prune can delete the last good local daily during a failure streak** — `scripts/prune-backups.sh:40-60`; also sweeps manual `prodcal-pre-migration.*` snapshots and contradicts the 30-day retention story (`KEEP_RECENT=10`). Fix: never prune while `.LAST-FAILURE` exists; tighten the delete glob to `prodcal-[0-9]{8}-[0-9]{6}`.
- **H6. `/api/public/summary` queries a nonexistent `archived` column; error swallowed** — `srv/server.go:304` (`WHERE COALESCE(archived,0)=0`; column is `archived_at` per migration 011; Scan error discarded with `_ =`). **Verified live:** returns `projects_active:0` with a live non-archived project. Landing page's active count is permanently 0. Fix: `WHERE archived_at IS NULL` and stop discarding Scan errors.

**Email** (all from EmailSystem)
- **H7. Transmittal email injects client-controlled content unescaped into HTML** — `srv/email.go` `buildTransmittalHTMLSummary` never imports `html`; client-editable transmittal fields flow raw into HTML sent to `jdbb@agentmail.to`/`j@djinna.com` (phishing vector). Every other builder escapes. Fix: `html.EscapeString` all user-derived values (~12 sites).
- **H8. Outbound spam relay** — all 4 manual email endpoints accept arbitrary recipient lists with **no rate limit / no cap**; combined with open-access projects (B1), anyone can send unlimited ProdCal-branded mail from `jdbb@agentmail.to`. Fix: per-IP/per-project rate limit + recipient cap (≤10).
- (The client-digest auth bypass is the same root cause as B1.)

**Admin workflow** (all from AdminWorkflowUX)
- **H9. Files tab dead-ends** — uploaded books have no Convert button, errored books no Retry (`startConvert`/`retryConvert` are dead code, `admin.html:1606/1652`); unlinked books can't be converted from the UI at all (`books.go:92` upload doesn't auto-convert). Fix: render Convert/Retry, wire the dead functions, show ErrorMsg inline.
- **H10. Custom-styles editor loses focus on every keystroke** — `admin.html:2492-2524` rebuilds `container.innerHTML` inside the Name `input` handler. Feature effectively unusable. Fix: update the row in place.
- **H11. "Delete Project" button: confirm dialog, then silently does nothing** — `app.js:783-787` calls `DELETE` with no try/catch; `server.go:563` always 405s ("archive instead"). Unhandled rejection, modal stays open. Fix: remove it or replace with Archive.
- **H12. Corrections Save ignores response status** — `admin.html:3253-3272`; on 4xx/5xx the typed find/replace is cleared and lost, list unchanged. Same on status-cycle/delete. Fix: check `res.ok`, keep form populated on failure.
- **H13. New project with a brand-new client slug creates an orphan client** — modal invites new slugs (`admin.html:1196`) but no `clients` row is created (`server.go:426-473`); portal 404s, client invisible in list. Fix: `INSERT OR IGNORE INTO clients` in the create tx, or warn in the modal.
- **H14. Auth-expiry handling checks 403 but server sends 401/302** — `admin.html:1026-1030` vs `admin.go:42`; on expiry the operator gets `projects.reduce is not a function` re-shown every 30 s instead of a login prompt. Fix: treat any `!r.ok` as login-required.
- **H15. Renamed project slugs leave stale URLs → blank page** — `app.js:1317-1336`; old path 404s, uncaught, `#app` stays empty. Fix: catch and render "Project not found."

**Client-facing hygiene** (ClientPortalUX)
- **H16. Internal notes & session logs are publicly routable** — `srv/static/notes.html` + `notes-2026-02-23-*.html` + `epub-workflow-test-checklist.html` served at `/static/…` (verified live 200); `notes.html` links a **real prod client** `/vgr/aog/` and the admin URL. Combined with B1, hands visitors a likely-editable real project URL. Fix: move them out of `srv/static/` before launch.
- **H17. No 404s; `/favicon.ico` & `/robots.txt` return portal HTML; landing has no favicon** — `server.go:190-210` soft-404s every path to 200 `client.html` (verified live); `landing.html` has no `<link rel=icon>`. Fix: short-circuit `/favicon.ico`, `/robots.txt`, `/apple-touch-icon*`; add favicon; pair with H2.
- **H18. 5 unshipped landing forks publicly routable** — `landing-{alt,editorial,tease,navbar,alt-nav}.html` at `/static/…` (verified live). Not linked from any shipped page (nav is clean), but indexable; `landing-tease.html` renders **fake "healthy/accepting/ready" status rows** on the prod domain. Fix: delete the rejected forks once a variant is chosen. **Ship `landing.html`** (runner-up `landing-alt`; never `landing-tease`).

---

## 🟢 QOL — high-value workflow wins — ✅ ALL RESOLVED 2026-07-04

**Top 5 operator-daily wins (AdminWorkflowUX ranking):** (1) custom-styles focus loss [H10, fixed earlier]; (2) spec autosave flush [B4, fixed earlier]; (3) Convert/Retry in Files tab [H9, fixed earlier]; (4) ✅ **one toast system on the admin dashboard** — `showToast` ported into `admin.html` (inline script + CSS); every routine success/failure `alert()` replaced with toasts, silent successes now confirmed (archive/restore/upload/link/client-create/passwords/corrections); `confirm()` kept only for genuinely destructive decisions; (5) auto-create client [H13, fixed earlier] + ✅ friendly 409s (below).

- ✅ Backup "ACTION NEEDED" mirrored into the admin header: red `.backup-alert` badge shown when `/api/admin/backup-status` is not ok (problems in tooltip; click scrolls to footer detail).
- ✅ Upload success message no longer erased in the same tick — success is a toast that survives the form reset.
- ✅ Modals with typed input no longer close on bare backdrop click (confirm-if-dirty); `beforeunload` warns on dirty modal forms or unsaved spec (`tsDirty`).
- ✅ Slug collisions → 409 `"a project with that client/project slug already exists"` via `isUniqueErr` in create/update/duplicate (`server.go`); covered by `TestCreateProjectSlugCollision409` / `TestUpdateProjectSlugCollision409`.
- ✅ 30 s dashboard refresh skips the list re-render while a modal is open or a form control is focused (health/backup polls still run).
- ✅ `handleDuplicateProject` wrapped in one transaction; any task-copy failure rolls back the whole duplicate; `tasks_copied` = rows actually inserted (`TestDuplicateProjectReportsTasksCopied`, `TestDuplicateProjectSlugCollisionRollsBack`).
- ✅ Transmittal update/restore/duplicate each run in a single tx; version-insert errors now fail the request (500 + rollback); `transmittal_versions` capped at 50 per transmittal, pruned in-tx (`TestTransmittalVersionCapEnforced`, `TestTransmittalVersionInsertFailureFailsUpdate`).
- ✅ `.LAST-DRILL-FAILURE` / `.LAST-R2-MONTHLY-FAILURE` / `.LAST-R2-DRILL-FAILURE` now surface in `/api/admin/backup-status` problems (and the new admin header badge); **R2 restore path now drilled** by new `scripts/r2-restore-drill.sh` (newest R2 object → gunzip → integrity_check → sentinel; monthly cron documented in DEPLOY.md).
- ✅ Blob double-storage removed: migration 018 backfills `books.pdf_data`/`epub_data` into `book_outputs` then drops the columns; pipelines write once (`CreateBookOutput`), downloads serve the newest output row. Upgrade covered by `TestBlobDedupeMigrationBackfill`; fresh boot 001–018 verified.
- ✅ Font persisted: first visit still picks a random typeface, then it sticks — `{font, dark}` in `prodcal-theme-v1` shared across landing/portal/calendar/transmittal.
- ✅ Password screens unified on the portal-gate pattern (same card/wording, inline "Invalid password", forgot-mailto kept, no `alert()`).
- ✅ Hardcoded prod URLs/emails removed from client-facing pages: relative `/` and `/admin/` links (also fixes the desktop app); human contact fetched from new `GET /api/public/config` (`CONTACT_EMAIL` env, default j@djinna.com) with hardcoded fallback; `jdbb@agentmail.to` archive recipient intentionally kept (system address). Zero `jdbbs.exe.xyz` occurrences remain in static files.
- ✅ Passwordless clients: portal project creation now 403s unless the caller is admin (`TestPortalCreateProjectPasswordlessClientAdminOnly`); protected-client + cookie flow unchanged.
- ✅ Email: Reply-To (`PRODCAL_MAIL_REPLY_TO`, default j@djinna.com) + From display name (`PRODCAL_MAIL_FROM_NAME`, default ProdCal, via AgentMail headers map) on all pathways; digest gets `List-Unsubscribe` header + footer line; transmittal HTML brought to text parity with inline styles (Gmail-safe), escaping preserved; AgentMail error bodies logged server-side, callers get generic "email send failed"; `snapshotFormatDate` escapes unparseable input (`TestTransmittalTextHTMLParity`, `TestSnapshotFormatDateFallback`).
- ✅ Docs/ops: `README.md` rewritten as a real ProdCal readme; `.githooks/pre-commit` now shellchecks staged `.sh` (skips zsh + missing-tool gracefully); `Makefile` `stop/start/restart` implemented as `systemctl … prodcal` wrappers; Docker path deleted (`Dockerfile` + `.dockerignore` removed — see NIT below).
- ✅ Mobile (390×844): `@media (max-width:480px)` rules — dash-header wraps, theme-bar bottom-anchored (no overlap), toast stack lifted, landing h1 scaled. Verified headless at 390×844.

---

## ⚪ NIT — ✅ ALL RESOLVED 2026-07-04
- ✅ single `esc()` in `admin.html` (hardened for null/undefined); ✅ `accept=".docx"` only + validation regex tightened; ✅ `cycleStatus` clears `actual_done` when leaving done; ✅ all `toISOString` date defaults replaced with local-timezone formatters (`todayLocalISO`/`localDateStr`); ✅ New Client password input `type="password"` + `autocomplete="new-password"`; ✅ empty-state portal digest already guarded (verified live); ✅ landing hints where the client code comes from (welcome email / contact link).
- ✅ `/api/transmittals/{id}` project-id semantics documented at the route registrations (URLs kept stable deliberately; all handlers use `projectIDFromPath`).
- ✅ `db/queries/visitors.sql` → `db/queries/projects.sql` (dbgen regenerated with sqlc v1.30.0).
- ✅ `:memory:` test-DB pool flake: moot since B2's `SetMaxOpenConns(1)`.
- ✅ `synchronous=NORMAL` already in the DSN since B2; ✅ `books(project_id)` indexed (migration 017); ✅ `r2-lifecycle.json` single predicate unwrapped from `Filter.And`.
- ✅ Stale `srv`-unit refs fixed in `README.md`/`DEPLOY.md`/`handoff.md`; `srv.service` renamed `prodcal.service`; notes html already removed from static in H16.
- ✅ Docker path deleted (`Dockerfile`, `.dockerignore`); `DEPLOY.md` seeding curls carry `-H 'X-ExeDev-UserID: admin'`.

---

## Cross-cutting themes
1. **The proxy is the security model.** `X-ExeDev-UserID` = admin, and `checkAuth` fails open — the app assumes exe.dev in front of it. B1 + the header trust are the same class: *nothing behind the proxy re-validates*. Decide this explicitly before launch.
2. **Silent failures everywhere.** Swallowed Scan errors (H6), ignored `res.ok` (H12), discarded version-insert errors, `applyCorrectionsIfAny` "logged but doesn't fail the compile" — the app frequently reports success on failure. This is the highest-leverage *pattern* to fix.
3. **No feedback primitive on admin.** One toast system would erase a whole cluster of QOL findings.
4. **Test coverage holes on mutating paths:** transmittal version restore, the entire corrections feature (incl. the silent-failure apply that can ship an unpatched book), backup status — all pure-Go-testable, all untested. No CI exists.

## Still owed (coverage gaps)
- **BookPipeline re-review:** `epub.go` zip handling (zip-slip), temp-file cleanup/collisions under concurrent builds, corrections-YAML corruption safety. (Shell-injection ruled out; upload sizes bounded.)
- **BackendCore's un-recovered findings:** it reported 3 BLOCKER / 7 HIGH; items beyond B1/B2/H6 were lost to the provider filter — worth a fresh targeted pass on `server.go`/`client.go` error handling, panics, and resource leaks.

---

## Phase-2 re-review results (done directly by Main, post-fix)

**BookPipeline gap — CLOSED. No new blocker.**
- **Command injection: none.** Every pandoc/typst/python call uses `exec.Command` with separate argv elements (e.g. `--metadata=title:%s` is a single arg); no shell is invoked, so user-controlled titles/filenames cannot escape (`books.go:298/364`, `epub.go:183`, `chapter_detect.go:413`, `bookspecs.go:849`, `corrections_apply.go:94`, `preflight.go:106`).
- **Zip handling: no zip-slip.** `injectChapterAuthors` + `stripMimetypeExtraField` (`epub.go:367/479`) operate entirely in memory (`zip.NewReader` over a `[]byte` → `bytes.Buffer`); nothing is extracted to a filesystem path, and the epub processed is the one pandoc just generated.
- **Temp files: safe.** `runConversion`/`runEPUBGeneration` use `os.MkdirTemp` with a per-book random suffix + `defer os.RemoveAll`; no predictable paths or cross-build collisions. Uploads bounded (50 MB `books.go:22`, 10 MB `bookspecs.go:951`).
- **Real issue (already tracked):** `applyCorrectionsIfAny` (`corrections_apply.go:104`) — "logged but does NOT fail the compile" ships an UNPATCHED book while reporting success. Silent-failure theme; not a security hole (original content preserved).
- **Verdict:** a hostile or malformed DOCX cannot escape into the shell or break the server — worst case is a conversion marked failed via `failConversion`.

**BackendCore:** re-verified the recovered items. checkAuth fail-open (fixed B1), DB FK-off/busy (fixed B2), plaintext cookie (fixed H1) are resolved. Remaining known backend items are HIGH/QOL and tracked: `archived`/`archived_at` swallowed error (H6), migration non-transactionality (H3), duplicate-project no-tx (QOL). BackendCore's original yield claimed 3 BLOCKER / 7 HIGH but was truncated by the provider filter; the 3rd blocker could not be reconstructed and I found no evidence of an un-captured backend blocker beyond B1/B2. A focused manual pass on `server.go` error handling remains the one residual item.

## Blockers fixed this session (verified)
- **B1/B1b + H1** — `checkAuth` now fails closed on DB error and requires the client password when the client is protected (even for token-less projects); `has_auth` reflects client-level gating; client cookie is now an HMAC session token, not the plaintext password, with `Secure` when served over https. Live-verified: anonymous GET transmittal on a protected client → 401; with cookie → 200; cookie value is a token.
- **B2** — SQLite pragmas moved to the DSN (`_pragma=busy_timeout(5000)&foreign_keys(1)&journal_mode(WAL)&synchronous(NORMAL)`) + `SetMaxOpenConns(1)`.
- **B3** — DEPLOY.md restore runbook rewritten (`prodcal.service`, `zcat` without consuming the backup, clears stale `-wal`/`-shm`).
- **B4** — spec autosave flushes the outgoing project's save bound to its id on switch; autosave timer bound to the loaded pid.
- **B5** — Go toolchain → 1.26.4 + `golang.org/x/net@v0.55.0`; govulncheck now reports 0 vulnerabilities.

## QOL + NIT closure session (2026-07-04, post-HIGH)

Entire 🟢 QOL + ⚪ NIT tier resolved (annotated ✅ inline above). Method: 7 parallel fix agents with disjoint file ownership (admin.html / client-facing static / transmittal.go / email / ops-scripts / docs / db-blob-dedupe) + backend edits and integration by Main + a dedicated test agent.

**Verification:** `go build ./...`, `go vet ./...`, `gofmt` clean; `shellcheck` clean on all touched scripts; `go test ./...` fully green locally (incl. `TestWordTemplateGenerationAllowsUniqueCustomStyleNames` — python-docx now installed on the Mac). Browser-smoked at desktop + 390×844: admin (zero console errors, toasts, backup badge, 409 paths, duplicate-tx), landing/portal/calendar/transmittal (font persistence across pages+reloads, unified gates, relative links, transmittal prefill → autosave → reload round-trip). Fresh-DB boot applies migrations 001–018; upgrade path covered by `TestBlobDedupeMigrationBackfill` and exercised live on the desktop-app data dir.

**Also this session (local desktop-app track):** startup WARNINGs for missing pipeline tools (pandoc/typst/python-docx) in `internal/localrun`; `scripts/build-mac-app.sh` producing a double-clickable `ProdCal.app` (Info.plist bakes `JDBBS_TYPESETTING_DIR` + Homebrew `PATH`); LOCAL-USAGE.md app-bundle section; full operator E2E verified locally: DOCX upload → convert → print PDF (26,906 B `%PDF-`) → EPUB (14.1 MB, embedded fonts) → preflight (findings JSON). Note: the untracked repo-root "01 Aster's Story (copyedited).docx" is actually an HTML file, not a DOCX — the pipeline rejects it cleanly; E2E used a real generated DOCX fixture.
