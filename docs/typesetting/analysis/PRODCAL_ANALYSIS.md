# Prodcal Project — Comprehensive Analysis

Generated: 2026-07-12

## 1. Database Schema Evolution

### Migration History (14 migrations)

| # | Name | What it adds |
|---|------|--------------|
| 001 | base | Core tables: `migrations`, `projects`, `tasks`, `auth_tokens`. Full task model with budget fields (orig/curr/actual), milestone flag, status tracking, assignee, word counts, rates. |
| 002 | slugs | Adds `client_slug` + `project_slug` to projects with unique composite index for URL routing (`/vgr/aog/`). |
| 003 | transmittal | Creates `transmittals` table — one per project, JSON blob storage, draft/final status. Unique index on project_id. |
| 004 | transmittal-versions | Creates `transmittal_versions` table for version history. Each save creates a version row. |
| 005 | clients | Creates `clients` table with slug PK, name, password_hash. Enables client-level auth (one password for all client projects). |
| 006 | file-log-journal | Creates `file_log` (file transfer tracking: direction, filename, type, sent_by, received_by) and `journal` (timestamped notes: call/decision/approval/note types). |
| 007 | books | Creates `books` table for uploaded Word files with BLOB storage for source, PDF, and EPUB data. Status tracking + error messages. |
| 008 | book-specs | Creates `book_specs` table — one per project (UNIQUE constraint), JSON blob for typesetting configuration. |
| 009 | book-project-link | ALTERs `books` to add `project_id` FK. ALTERs `book_specs` to add `cover_data` BLOB + `cover_type`. |
| 010 | corrections | Creates `corrections` table for discrete find/replace edit propagation with chapter targeting and applied/skipped status. |
| 011 | project-archive | ALTERs `projects` to add `archived_at` timestamp. Soft-delete pattern instead of hard delete. |
| 012 | manuscript-preflights | Creates `manuscript_preflights` table for persistent preflight reports (summary JSON, report JSON, report HTML). UNIQUE(project_id, book_id). |
| 013 | preflight-history | **Destructive migration**: recreates `manuscript_preflights` without the UNIQUE constraint to allow history (multiple reports per project/book). Uses copy-to-new-table pattern for SQLite. |
| 014 | book-output-history | Creates `book_outputs` table for preserving EPUB/PDF output history per book (BLOB storage per output). |

### Current Schema State — All Tables

**Core entities:**
- `projects` — id, name, start_date, client_slug, project_slug, archived_at, created_at, updated_at
- `tasks` — id, project_id (FK CASCADE), sort_order, assignee, title, is_milestone, orig_weeks, curr_weeks, orig_due, curr_due, actual_done, status, words, words_per_hour, hours, rate, budget_notes, orig_budget, curr_budget, actual_budget, created_at, updated_at
- `clients` — slug (PK), name, password_hash, created_at

**Auth:**
- `auth_tokens` — id, project_id (FK CASCADE), token_hash, label, created_at

**Transmittals:**
- `transmittals` — id, project_id (FK CASCADE, UNIQUE), status, data (JSON), created_at, updated_at
- `transmittal_versions` — id, transmittal_id (FK CASCADE), data (JSON), status, saved_at

**File tracking:**
- `file_log` — id, project_id (FK CASCADE), direction, filename, file_type, sent_by, received_by, notes, transfer_date, created_at
- `journal` — id, project_id (FK CASCADE), entry_type, content, created_at

**Book production:**
- `books` — id, title, author, series, source_filename, source_data (BLOB), pdf_data (BLOB), epub_data (BLOB), status, error_msg, project_id (FK SET NULL), created_at, updated_at
- `book_specs` — id, project_id (FK CASCADE, UNIQUE), data (JSON), cover_data (BLOB), cover_type, created_at, updated_at
- `book_outputs` — id, book_id (FK CASCADE), output_format, output_data (BLOB), source_filename, created_at
- `corrections` — id, project_id (FK CASCADE), find_text, replace_text, chapter, note, status, applied_at, created_at

**Preflight:**
- `manuscript_preflights` — id, project_id (FK CASCADE), book_id (FK CASCADE), status, summary_json, report_json, report_html, error_msg, source_filename, created_at, updated_at

**Meta:**
- `migrations` — migration_number (PK), migration_name, executed_at

### Relationships

```
clients (slug PK)
  └─ projects (client_slug references clients conceptually, but NO FK constraint)
       ├─ tasks (project_id FK CASCADE)
       ├─ auth_tokens (project_id FK CASCADE)
       ├─ transmittals (project_id FK CASCADE, 1:1)
       │    └─ transmittal_versions (transmittal_id FK CASCADE)
       ├─ file_log (project_id FK CASCADE)
       ├─ journal (project_id FK CASCADE)
       ├─ book_specs (project_id FK CASCADE, 1:1)
       ├─ corrections (project_id FK CASCADE)
       ├─ books (project_id FK SET NULL)
       │    ├─ book_outputs (book_id FK CASCADE)
       │    └─ manuscript_preflights (book_id FK CASCADE)
       └─ manuscript_preflights (project_id FK CASCADE)
```

## 2. SQL Query Patterns & Analysis

### sqlc Configuration
- Engine: SQLite
- Generated package: `dbgen`
- Queries in `db/queries/` → generated Go in `db/dbgen/`

### Query Files

**books.sql** (11 queries): Full CRUD, status updates, PDF/EPUB data retrieval, project association. Uses `RETURNING *` for creates. Separates listing (excludes BLOBs) from data retrieval (includes BLOBs) — good pattern for performance.

**book_specs.sql** (5 queries): Get, Upsert (ON CONFLICT DO UPDATE), Delete, cover get/update. Clean upsert pattern.

**book_outputs.sql** (2 queries): Create + list by book_id and format. Ordered by created_at DESC for history.

**manuscript_preflights.sql** (3 queries): GetLatest (ORDER BY + LIMIT 1), Create, ListAll. Good pattern for history with latest-first ordering.

**visitors.sql** (19 queries): The main query file — projects, tasks, auth tokens. Full CRUD for all. Notable: `GetProjectByPath` filters `archived_at IS NULL`. Separate `ListProjects` (active) and `ListArchivedProjects` queries.

### Issues & Observations

1. **Missing FK constraint**: `projects.client_slug` references `clients.slug` conceptually but has no FK constraint in the schema. The relationship is maintained only by application logic.

2. **BLOB storage in SQLite**: Books store source, PDF, and EPUB data as BLOBs directly in the database. For small-to-medium files this works, but large manuscripts could bloat the DB. The `book_outputs` table compounds this by storing historical outputs.

3. **No query file for file_log/journal/corrections/transmittals**: These tables use raw SQL in Go handlers rather than sqlc-generated code (mixed pattern).

4. **Migration numbering inconsistency**: Migration 011 uses `VALUES (11, 'project archive')` — non-zero-padded number and different naming convention from others.

5. **corrections table has redundant FK**: Has both `FOREIGN KEY (project_id) REFERENCES projects(id)` in CREATE TABLE and separate `REFERENCES projects(id) ON DELETE CASCADE` in column definition. The CASCADE is on the column def; the table-level FK lacks CASCADE — but SQLite takes the first one it sees.

6. **corrections index lacks IF NOT EXISTS**: `CREATE INDEX idx_corrections_project` — will fail if run twice (though the migration tracking prevents re-execution).

## 3. Frontend Architecture

### Architecture: Multi-Page Application with SPA-like Pages

The app uses **server-side routing to distinct HTML pages**, each of which behaves as a mini-SPA:

| File | URL Pattern | Role | Lines |
|------|------------|------|-------|
| `index.html` | `/{client}/{project}/` | Calendar shell — loads `app.js` | 27 |
| `app.js` | (loaded by index.html) | Full calendar SPA: timeline/table/budget/files/journal tabs | 1,295 |
| `admin.html` | `/admin/` | **Self-contained** admin dashboard with inline `<style>` + `<script>` | 2,780 |
| `client.html` | `/{client}/` | Client portal — inline JS, project cards, activity tabs | 919 |
| `landing.html` | `/` | Public landing page — inline JS, client links | 343 |
| `transmittal.html` | `/{client}/{project}/transmittal/` | Transmittal shell — loads `transmittal.js` + `transmittal.css` | 28 |
| `transmittal.js` | (loaded by transmittal.html) | Full transmittal form SPA | 1,295 |
| `style.css` | (shared stylesheet) | Design system with light/dark mode | 1,156 |
| `transmittal.css` | (transmittal-specific) | Transmittal form + email modal styles | 873 |

### Tab System
- **Calendar page** (`app.js`): Tabs for Timeline, Table, Budget, Files, Journal, Typesetting — controlled by JS state, switching `display: none/block` on `.tab-panel` divs.
- **Admin page** (`admin.html`): Active/Archived tabs for project lists, plus Clients section.
- **Client portal** (`client.html`): Project cards + Recent Activity sub-tabs (Files/Journal).

### Virtual DOM / Rendering
- Uses a custom `h()` hyperscript function to build DOM elements imperatively.
- Full re-render on state change (not diffing — rebuilds DOM subtrees).
- State management via plain JS objects (`state.project`, `state.tasks`, etc.).
- Auto-save with debouncing (600ms for transmittal).

### Theme System
- Light/dark mode via `html.dark` class.
- Theme persisted in `localStorage` (`prodcal-theme-v1`).
- Flash-prevention script runs inline before body to set class early.
- Font switching (IBM Plex Serif, IBM Plex Sans, Literata) via theme bar.

### Key Frontend Patterns
- **No build step**: All JS is vanilla, no bundler, no TypeScript.
- **Embedded static**: Files are embedded into the Go binary via `embed` package.
- **Cache busting**: Manual version query params on CSS/JS links (`?v=20260409b`).
- **Modals**: Used for task editing, email recipients, file log entry, journal entry.
- **Inline styles in admin.html**: The admin page is completely self-contained (~2,780 lines) with all CSS and JS inline. This duplicates the design system variables.

## 4. CSS Design System

### Design Tokens (CSS Custom Properties)
- **Colors**: `--bg`, `--surface`, `--border`, `--text`, `--text-secondary`, `--text-muted`, `--accent`, `--green`, `--red`, `--yellow`, `--orange`
- **Badge variants**: `--badge-green-bg`, `--badge-red-bg`, `--badge-yellow-bg`, `--badge-dim-bg`
- **UI**: `--progress-bg`, `--toggle-bg`, `--hover-tint`, `--modal-backdrop`, `--radius`
- **Decorative**: `--noise-opacity`, `--grid-color`, `--gradient-wash`, `--theme-bg`

### Visual Identity
- **Typography**: IBM Plex Serif (body), IBM Plex Sans (UI), Literata (option). Serif-first.
- **Background effects**: SVG noise texture overlay + gradient wash + dashed vertical grid lines.
- **Color palette**: Blue accent (#2563eb light / #5b9aff dark), muted greens/reds/yellows for status.
- **Layout**: Max-width 1100px centered content.

### Dark Mode
- Full dark mode with separate token values.
- Smooth 0.3s transitions on background/color changes.
- All status colors shift to lighter variants in dark mode.

### Issue: Style Duplication
- `admin.html` contains ~400+ lines of inline CSS that largely duplicate `style.css` and add admin-specific styles. Any design token change needs to be made in both places.
- `landing.html` also has inline styles duplicating the core design tokens.
- `client.html` has its own inline styles too.

## 5. Test Coverage & Quality

### Test Infrastructure
- **`server_test.go`**: Core test helpers — `testServer()` creates in-memory SQLite DB + runs migrations + creates `httptest.Server`. Helper functions: `apiRequest`, `apiRequestAdmin`, `decodeJSON`, `itoa`.
- Tests run against real HTTP endpoints via `httptest`, hitting the full handler stack.
- No mocking framework — uses simple function fields (e.g., `s.preflightRunner`) for injectable behavior.

### Test Files & Coverage

| File | Tests | What's Covered |
|------|-------|----------------|
| `server_test.go` | 4 | Health check, project CRUD, archive, admin-only archive guard |
| `auth_test.go` | 5 | Auth required, verify password, no-auth project, cookie auth, remove auth |
| `admin_test.go` | 5 | Client CRUD, duplicate slug, archive/restore, admin HTML content checks |
| `client_test.go` | 3 | Client auth for project creation, task seeding (31 tasks), client project listing |
| `task_test.go` | 3 | Task CRUD, milestones, default status |
| `transmittal_test.go` | 2 | Transmittal defaults (ISBN, checklist status, custom_styles), duplicate transmittal resets |
| `filelog_test.go` | 3 | File log CRUD, defaults (direction/date), outbound entries |
| `journal_test.go` | 4 | Journal CRUD, content required validation, default type, all entry types |
| `email_test.go` | 8 | Email not configured (503), recipient validation, snapshot/activity HTML builders with/without data, activity text, emoji mapping |
| `bookspecs_test.go` | 2 | Pull transmittal maps custom styles, maps shared typesetting fields |
| `custom_styles_roundtrip_test.go` | 1 | Custom style type/description preservation through save/load cycle |
| `preflight_test.go` | 5 | Missing preflight (exists=false), admin auth required, stored HTML report, history + latest report URL, preflight success with summary, undeclared style warning, declared style usage |
| `word_template_test.go` | 3 | Pull transmittal maps EPUB ISBN, duplicate custom style names rejected, unique styles allowed |

**Total: ~48 test functions across 13 files.**

### Quality Assessment
- **Strengths**: Tests are integration-level, hitting real endpoints with real DB migrations. Good coverage of auth cascade, CRUD operations, edge cases (duplicates, defaults, missing data).
- **Gaps**:
  - No tests for `corrections` CRUD.
  - No tests for `books` upload/download/convert endpoints.
  - No tests for `epub.go` generation.
  - No tests for transmittal version history API.
  - No tests for the client digest email.
  - No tests for transmittal notification throttling.
  - No frontend/browser tests.
  - No load/performance tests.

## 6. Documentation Completeness

### Primary Docs
- **`DEPLOY.md`**: Comprehensive deployment guide — build, migrations, seeding, backup, restore, Docker, monitoring, architecture diagram. Well-written.
- **`EMAIL_SYSTEM.md`**: Excellent reference — all 6 email pathways documented with endpoints, triggers, auth, HTML conventions, and "how to add a new pathway" guide.
- **`CHECKPOINTS.md`**: Lightweight git tag convention for marking verified states.
- **`SESSION-SUMMARY.txt`**: Detailed session-by-session changelog from session 5 through 10. Includes credentials, URLs, conversation chain, fix/build list.
- **`TEAM-UPDATE.txt`**: User-facing email to team for initial testing.

### Planning Docs
- **`NEXT-SESSION-EMAIL-INTEGRATION.md`**: Detailed implementation plan for email integration (marked COMPLETE).
- **`NEXT-SESSION-TESTING-PLAN.md`**: Testing plan (marked COMPLETE — ref commit 4f217cd).
- **`docs/plans/2026-04-09-transmittal-custom-styles-plan.md`**: Thorough 6-task implementation plan for custom styles pipeline.

### Notes
- **`docs/notes/2026-04-08-transmittal-resume.md`**: Mid-session resume notes with TDD progress.
- **`docs/notes/2026-04-12-reorientation-and-clean-restart.md`**: State capture after interrupted SSH session.
- **`docs/notes/2026-04-12-special-typography-preflight-and-ascii-preservation.md`**: Deep product/architectural note on special typography handling. Very thorough.
- **`docs/notes/2026-04-13-typst-pipeline-rewire-and-testing-retro.md`**: Retrospective on QA efficiency.
- **`docs/notes/2026-04-14-preflight-redesign-parked.md`**: Explicit "park this work" decision note.

### Checklists
- **`docs/checklists/2026-04-11-epub-test-session.md`**: Phase-gated EPUB workflow test results.

### Documentation Quality
- **Excellent operational docs**: DEPLOY.md and EMAIL_SYSTEM.md are production-ready.
- **Strong session memory**: SESSION-SUMMARY.txt provides continuity across coding sessions.
- **Good decision records**: The notes capture *why* decisions were made, not just what.
- **Gap**: No API reference doc. Endpoints are documented piecemeal across SESSION-SUMMARY.txt and EMAIL_SYSTEM.md but there's no single API spec.
- **Gap**: No README.md at repo root.

## 7. Infrastructure & Deployment

### Build
- **Makefile**: Simple `go build -o prodcal ./cmd/srv` + clean + test.
- **Go module**: `srv.exe.dev`, Go 1.26.0, SQLite via `modernc.org/sqlite` (pure Go, no CGO).
- **sqlc**: Code generation from SQL queries → `db/dbgen/`.

### Deployment
- **systemd service** (`srv.service`): Runs as `exedev` user, WorkingDirectory `/home/exedev`, listens on `:8000`.
- **Environment**: `.env` file loaded via `EnvironmentFile` — contains AgentMail API key.
- **Proxy**: exe.dev proxy routes `jdbbs.exe.xyz` → port 8000.
- **Database**: SQLite at `/home/exedev/db.sqlite3` (note: WorkingDirectory is `/home/exedev`, not `/home/exedev/prodcal`).

### Backup
- **Script**: `scripts/backup-db.sh` — uses `sqlite3 .backup` (safe with WAL mode), gzip compression, 7-day retention.
- **Cron**: Daily at 3 AM.

### Docker
- **Dockerfile**: Multi-stage build (golang:1.26 → debian:bookworm-slim). Installs ca-certificates + sqlite3. Volume at `/app/data`. Not used in production currently.

### .gitignore / .dockerignore
- Properly ignores: binary, DB files, .env, seed data, WAL/SHM files.

## 8. Issues, Gaps & Inconsistencies

### Critical / High Priority

1. **API Key in .env is readable**: The AgentMail API key `am_us_8de7...` is in a `.env` file that's gitignored but present on disk. The `.env` is chmod 600 per docs, which is good, but the key appears in SESSION-SUMMARY.txt which IS tracked in git. **Security concern**: if SESSION-SUMMARY.txt is in the repo, the API key is exposed.

2. **Password in SESSION-SUMMARY.txt**: The file contains plaintext passwords (`artofgig2026`, `willwrite2026`). If this file is committed to git, these credentials are in version history.

3. **Duplicate systemd units**: Both `srv.service` and `prodcal.service` exist and can race for port 8000. This has caused issues in multiple sessions (documented in notes). Needs cleanup.

4. **No FK from projects to clients**: `projects.client_slug` has no foreign key to `clients.slug`. Orphaned projects can exist for non-existent clients.

### Medium Priority

5. **BLOB storage scalability**: Storing full book files (DOCX, PDF, EPUB) + their history as BLOBs in SQLite will cause DB size to grow quickly. The backup script copies the entire DB daily. Consider moving BLOBs to filesystem with DB metadata.

6. **CSS duplication across pages**: Design tokens and base styles are duplicated in `admin.html` (~400 lines), `landing.html`, `client.html`. Changes must be made in 4+ places. Should extract shared CSS into `style.css` and have all pages reference it.

7. **Mixed query patterns**: Some tables use sqlc-generated queries (`books`, `book_specs`, `book_outputs`, `manuscript_preflights`, `projects`, `tasks`, `auth_tokens`) while others use raw SQL in handlers (`file_log`, `journal`, `corrections`, `transmittals`, `clients`). Inconsistent and makes it harder to reason about data access.

8. **No transmittal/file_log/journal/corrections sqlc queries**: These entities lack generated query files. Adding them would improve type safety and consistency.

9. **admin.html is 2,780 lines**: This monolithic file contains HTML, CSS, and JS all inline. It's the largest frontend file and would benefit from extraction.

### Low Priority / Cosmetic

10. **Migration 011 naming inconsistency**: Uses bare `11` instead of `011` and a different naming style (`'project archive'` vs `'011-project-archive'`).

11. **No README.md**: The repo lacks a root README. DEPLOY.md partially covers this.

12. **Cache busting is manual**: Version query params like `?v=20260409b` must be updated by hand. Since static files are embedded in the binary, this is partially mitigated (rebuild = new binary), but the query params in transmittal.html/index.html could go stale.

13. **Test helper `testServerForHTML`**: Defined in `admin_test.go` but is essentially a convenience wrapper. Minor code organization issue.

14. **Missing test coverage**: No tests for books upload/convert, EPUB generation, corrections CRUD, transmittal version history, client digest email, transmittal notification throttling.

15. **`corrections` table FK duplication**: Has both column-level and table-level foreign key definitions for `project_id`.

16. **WorkingDirectory mismatch**: `srv.service` sets `WorkingDirectory=/home/exedev` but the repo and binary are in `/home/exedev/prodcal`. The DB path is relative (`db.sqlite3`), so it resolves to `/home/exedev/db.sqlite3` — documented in SESSION-SUMMARY.txt but could confuse new developers.

## 9. Architecture Summary

```
Client Browser
    │
    ├── GET / → landing.html (branded landing page)
    ├── GET /admin/ → admin.html (exe.dev auth required)
    ├── GET /{client}/ → client.html (client portal)
    ├── GET /{client}/{project}/ → index.html + app.js (calendar SPA)
    ├── GET /{client}/{project}/transmittal/ → transmittal.html + transmittal.js
    │
    └── API calls (JSON)
         ├── /api/projects/* — CRUD, archive/restore
         ├── /api/projects/{id}/tasks — task CRUD
         ├── /api/projects/{id}/transmittal — transmittal CRUD
         ├── /api/projects/{id}/file-log — file log CRUD
         ├── /api/projects/{id}/journal — journal CRUD
         ├── /api/projects/{id}/book-spec — book spec CRUD
         ├── /api/projects/{id}/preflight — manuscript preflight
         ├── /api/projects/{id}/*/email — email sending
         ├── /api/books/* — book upload/convert/download
         ├── /api/clients/* — client auth/portal
         ├── /api/admin/* — admin endpoints
         └── /healthz — health check

Go Backend (srv package, ~7,000 lines)
    ├── server.go (1,059 lines) — routing, auth middleware, project/task handlers
    ├── bookspecs.go (990 lines) — spec management, transmittal→spec bridge
    ├── snapshot_email.go (814 lines) — project snapshot email builder
    ├── preflight.go (530 lines) — manuscript preflight system
    ├── transmittal.go (470 lines) — transmittal API
    ├── email.go (462 lines) — AgentMail client, transmittal email
    ├── books.go (446 lines) — book upload/convert
    ├── client_digest_email.go (390 lines) — client weekly digest
    ├── activity_email.go (369 lines) — activity digest email
    ├── client.go (328 lines) — client portal API
    ├── epub.go (287 lines) — EPUB generation
    ├── admin.go (216 lines) — admin dashboard
    ├── corrections.go (210 lines) — corrections CRUD
    ├── transmittal_notify.go (180 lines) — auto-notification
    ├── journal.go (180 lines) — journal handlers
    └── filelog.go (157 lines) — file log handlers

SQLite Database (WAL mode)
    ├── 14 migrations (auto-run on startup)
    ├── 5 sqlc query files → generated Go code
    └── BLOB storage for books, covers, outputs
```

The project is a well-structured, pragmatic production tool with strong operational documentation and decent test coverage. The main technical debts are CSS duplication across pages, inconsistent data access patterns (sqlc vs raw SQL), and BLOB storage that may not scale well for heavy book production use.
