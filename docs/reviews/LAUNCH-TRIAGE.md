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

## 🟢 QOL — high-value workflow wins (post-launch OK)

**Top 5 operator-daily wins (AdminWorkflowUX ranking):** (1) fix custom-styles focus loss [H10]; (2) flush-or-confirm spec autosave on project switch [B4]; (3) Convert/Retry in Files tab [H9]; (4) **one toast system on the admin dashboard** — `showToast` (`app.js:38-51`, the d4bb543/24dea1b work) exists **only in the calendar SPA**; `admin.html` has none, so archive/restore/upload/link/client-create are silently confirmed or use `alert()`; (5) auto-create client + friendly 409s [H13].

- Backup "ACTION NEEDED" status is buried in the footer colophon (`admin.html:875`) — mirror a red indicator into the header. — Admin/Data
- Upload success message erased in the same tick (`admin.html:1544/1517`). — Admin
- All modals close on backdrop click; no `beforeunload` anywhere → unsaved-changes loss. — Admin
- Slug collisions surface as raw `UNIQUE constraint failed…` SQLite errors (`server.go:459/553`); map to 409 (pattern exists at `admin.go:214`). — Admin/Data
- 30 s full dashboard refresh re-renders the list mid-interaction (`admin.html:3304`). — Admin
- `handleDuplicateProject` is a multi-statement write with **no transaction**; `tasks_copied` reports source count, not successes (`server.go:1024-1125`). — Data
- Transmittal write paths race (read-snapshot-write, no tx); version-insert errors silently discarded; `transmittal_versions` grows unbounded (`transmittal.go`). — Data
- Restore-drill & monthly-anchor failures are invisible to monitoring (`.LAST-DRILL-FAILURE` / `.LAST-R2-MONTHLY-FAILURE` unread); **R2 disaster-restore path is never drilled**. — Data
- Compiled PDF/EPUB blobs stored **twice** (`books.{pdf,epub}_data` AND `book_outputs`), inflating every backup. — Data
- Random typeface on every page load (landing/client/transmittal/app all pick a random font, incl. Menlo) — brand roulette; font never persisted. — ClientPortal
- Three inconsistent password screens; transmittal/calendar gates use `alert()` and drop the "forgot" affordance. — ClientPortal
- Hardcoded prod URLs / personal emails in client-facing pages (`client.html:307` Back → `https://jdbbs.exe.xyz/`; footer Admin → absolute prod login; `jdbb@agentmail.to`/`j@djinna.com` hardcoded). **Also breaks the local/desktop-app case.** — ClientPortal
- Passwordless clients: anyone can create projects (and, once email is live, send mail) — `client.go:246`. — ClientPortal/Email
- Email: no Reply-To / display name / List-Unsubscribe; transmittal HTML drops sections the text version has and uses a `<style>` block Gmail strips; 500s forward raw AgentMail error bodies to end users; `snapshotFormatDate` returns unvalidated input raw (secondary injection sink). — Email
- Docs/ops: `README.md` is unmodified "Go Shelley Template" boilerplate; `.githooks/pre-commit` never runs shellcheck though `dev-setup.sh` requires it; `Makefile` declares phantom `stop/start/restart` targets; `.dockerignore` misses `scratch/`+`reference/` (MBs). — Tests
- Mobile (390×844): `client.html .dash-header` no-wrap flex + fixed `.theme-bar` likely overlap; only one narrow-viewport rule in `style.css`. *(Needs a visual pass.)* — ClientPortal

---

## ⚪ NIT (batch later)
- `esc()` defined twice in `admin.html`; `.doc` accepted but pipeline needs `.docx`; `cycleStatus` keeps stale `ActualDone`; `toISOString` UTC ±1-day date shift; New Client password input is `type="text"`; empty-state portal still offers Weekly Digest; landing gives no hint where the client code comes from. — Admin/ClientPortal
- `/api/transmittals/{id}` route param is actually a **project** id (consistent today, misleading). — Admin/Tests/Data
- `db/queries/visitors.sql` misnamed (holds projects/tasks/auth queries; no visitors table). — Data
- `:memory:` test DB + connection pool = separate DB per conn (latent parallel-test flakes). — Data
- `PRAGMA synchronous=NORMAL` is the recommended WAL pairing; `books(project_id)` unindexed; `r2-lifecycle.json` wraps a single predicate in `Filter.And`. — Data
- Stale `srv`-unit references (real unit: `prodcal.service`): `AGENTS.md:15-16`, `README.md:21/25/28/31/34/41`, `DEPLOY.md:8/18/36/78/81`, `handoff.md:30/45`, the `srv.service` filename itself, `notes-2026-02-23-*.html:82`. `CLAUDE.md` already correct. — Tests
- `Dockerfile` can't build (COPYs nonexistent `seed_data.json`; DB not on the `/app/data` VOLUME; runtime image lacks pandoc/typst/python-docx); `DEPLOY.md` seeding curls 401 (missing admin header). Delete Docker path or fix. — Tests

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
