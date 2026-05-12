# ProdCal — Comprehensive Technical Analysis

Generated: 2026-04-14

## 1. Executive Summary

ProdCal is a **book production management web application** (~11,100 lines of Go, ~9,300 lines of frontend HTML/JS/CSS) built for "jdbb studio" — a book production house. It manages the full lifecycle of book production projects from manuscript transmittal through typesetting, proofing, and final print delivery.

The app runs as a systemd service on port 8000, proxied through exe.dev's HTTPS infrastructure at `https://jdbbs.exe.xyz/`. It uses SQLite with WAL mode, Go 1.26 with the stdlib `net/http` mux, and sqlc for type-safe database queries.

---

## 2. Architecture

### 2.1 Overall Structure

```
cmd/srv/main.go          ← Binary entrypoint (20 lines)
srv/server.go            ← HTTP routes, core handlers, auth (650+ lines)
srv/admin.go             ← Admin dashboard + project/client list APIs
srv/transmittal.go       ← Transmittal CRUD + versioning + duplication
srv/books.go             ← Book upload, docx→typst→PDF conversion pipeline
srv/bookspecs.go         ← Typesetting spec management, config generation
srv/preflight.go         ← Manuscript preflight analysis (edge case detection)
srv/epub.go              ← EPUB generation pipeline (pandoc-based)
srv/client.go            ← Client portal auth + project listing
srv/corrections.go       ← Find/replace corrections tracking
srv/filelog.go           ← File transfer log
srv/journal.go           ← Project journal (notes, calls, decisions)
srv/email.go             ← Email infrastructure + transmittal email
srv/snapshot_email.go    ← Project snapshot email (comprehensive)
srv/activity_email.go    ← Activity digest email
srv/client_digest_email.go ← Cross-project client digest email
srv/transmittal_notify.go  ← Automatic transmittal update notification
srv/static/              ← Embedded frontend (SPA approach)
db/db.go                 ← SQLite open, migration runner
db/migrations/           ← 14 sequential SQL migrations
db/queries/              ← sqlc query definitions
db/dbgen/                ← sqlc generated Go code
```

### 2.2 HTTP Routing & Middleware

**Router**: Go 1.22+ stdlib `net/http.ServeMux` with method-based routing (`GET /api/projects`, `POST /api/projects/{id}/tasks`, etc.).

**~80 registered routes** organized as:
- `/healthz` — health check
- `/admin/` — admin dashboard (HTML)
- `/api/admin/*` — admin-only APIs
- `/api/projects/*` — project CRUD, tasks, transmittal, book specs, preflight, email, file log, journal, corrections
- `/api/books/*` — book upload/convert/download
- `/api/clients/*` — client portal APIs
- `/api/fonts` — font listing
- `/api/email/status` — email config status
- `/static/*` — CSS/JS assets
- `/{client}/{project}/` — calendar SPA
- `/{client}/{project}/transmittal/` — transmittal SPA
- `/{client}/` — client portal
- `/` — landing page

**No middleware layer** — auth checks are done inline per-handler via `requireAuth()`, `requireExeDevAdmin()`, or `requireExeDevAdminAPI()`.

**URL routing for SPAs**: The catch-all `GET /` handler parses URL path segments to determine which SPA to serve:
- 0 segments → landing page
- 1 segment → client portal
- 2 segments → calendar SPA
- 3 segments with "transmittal" → transmittal SPA
- Other deep paths → static file server

### 2.3 Authentication Model (3-tier)

1. **exe.dev admin auth**: `X-ExeDev-UserID` header set by the exe.dev reverse proxy. Required for admin dashboard, book management, and destructive operations.
2. **Project-level auth**: SHA-256 hashed passwords stored in `auth_tokens` table. Cookie-based (`prodcal_auth_{projectID}`).
3. **Client-level auth**: SHA-256 hashed passwords in `clients` table. Cookie-based (`prodcal_client_{slug}`). Grants access to all projects under that client.

Fallback chain: If no auth tokens exist for a project, it's open access. Otherwise, checks project cookie → client cookie → exe.dev header.

### 2.4 Database Schema & Migrations

**14 migrations** applied sequentially at startup:

| # | Name | Tables/Changes |
|---|------|---------------|
| 001 | base | `migrations`, `projects`, `tasks`, `auth_tokens` |
| 002 | slugs | Add `client_slug`, `project_slug` to projects |
| 003 | transmittal | `transmittals` (JSON blob per project) |
| 004 | transmittal-versions | `transmittal_versions` (version history) |
| 005 | clients | `clients` table for client-level auth |
| 006 | file-log-journal | `file_log`, `journal` tables |
| 007 | books | `books` table (BLOB storage for docx/pdf/epub) |
| 008 | book-specs | `book_specs` table (typesetting config JSON) |
| 009 | book-project-link | Add `project_id` to books, cover data to specs |
| 010 | corrections | `corrections` table |
| 011 | project-archive | Add `archived_at` to projects |
| 012 | manuscript-preflights | `manuscript_preflights` table |
| 013 | preflight-history | Recreate preflights without UNIQUE constraint |
| 014 | book-output-history | `book_outputs` table for PDF/EPUB history |

**Key design decisions**:
- Transmittal and book spec data stored as JSON text blobs (schema-less, flexible)
- Binary files (docx, pdf, epub, cover images) stored as SQLite BLOBs
- WAL mode enabled with 1s busy timeout
- Foreign keys enforced

### 2.5 Template/Frontend Rendering

**Embedded via `//go:embed static/*`** — all static files compiled into the binary.

**SPA architecture**: Four distinct SPAs, each a single HTML file:
- `admin.html` (2,780 lines) — admin dashboard with inline JS
- `index.html` (27 lines) — calendar SPA shell, loads `app.js`
- `transmittal.html` (28 lines) — transmittal SPA shell, loads `transmittal.js`
- `client.html` (919 lines) — client portal with inline JS
- `landing.html` (343 lines) — landing page

**No server-side template rendering** — all HTML is served statically, all data fetched via JSON APIs.

---

## 3. Features

### 3.1 Core Entities

| Entity | Description |
|--------|------------|
| **Projects** | Book production projects with name, dates, client/project slugs |
| **Tasks** | Production workflow steps with scheduling, budgeting, assignees |
| **Transmittals** | Manuscript transmittal forms (JSON blob with 15+ sections) |
| **Books** | Uploaded docx manuscripts + generated PDF/EPUB outputs |
| **Book Specs** | Typesetting configuration (fonts, margins, page size, headings) |
| **Corrections** | Find/replace entries for text corrections |
| **File Log** | Record of file transfers between parties |
| **Journal** | Project notes, call logs, decisions, approvals |
| **Clients** | Client organizations with auth |
| **Manuscript Preflights** | Automated manuscript analysis reports |

### 3.2 Workflows

**Project Lifecycle**:
- Create project (admin or client portal) → auto-seed 31-task standard workflow
- Duplicate project with date-shifted tasks and zeroed budgets
- Archive/restore projects (soft delete)
- Delete disabled (returns 405)

**Manuscript Transmittal**:
- Client fills out comprehensive form (book info, production dates, checklist, backmatter, illustrations, permissions, page iv, subrights, editing instructions, design specs, cover info, file archives, proofs)
- Auto-save with 5-minute throttled version snapshots
- Automatic notification email to admin when client saves
- Version history with restore capability
- Duplicate transmittal to another project (clears book-specific fields)

**Book Production Pipeline** (docx → PDF):
1. Upload docx manuscript
2. Run manuscript preflight (Python `detect-edge-cases.py`)
3. Configure typesetting spec (pull from transmittal or manual)
4. Generate Typst config from spec
5. Convert: pandoc docx→typst (with custom Lua filter) → typst compile → PDF
6. Download PDF

**EPUB Pipeline**:
1. pandoc docx→epub3 with metadata from spec
2. Custom CSS generated from spec (section breaks, font sizes)
3. Cover image from book spec
4. Output history preserved

**Email System** (6 pathways):
1. Manual: Transmittal summary, Project snapshot, Activity digest, Client weekly digest
2. Automatic: Transmittal update notification (throttled 30min/project)

### 3.3 CRUD Operations

| Entity | Create | Read | Update | Delete |
|--------|--------|------|--------|--------|
| Projects | ✅ | ✅ | ✅ | ❌ (archive only) |
| Tasks | ✅ | ✅ (list) | ✅ | ✅ |
| Transmittals | ✅ (auto) | ✅ | ✅ | ❌ |
| Books | ✅ (upload) | ✅ | ✅ (convert) | ✅ |
| Book Specs | ✅ (auto) | ✅ | ✅ | via sqlc only |
| Corrections | ✅ | ✅ | ✅ (status) | ✅ |
| File Log | ✅ | ✅ | ❌ | ✅ |
| Journal | ✅ | ✅ | ❌ | ✅ |
| Clients | ✅ | ✅ | ❌ | ❌ |
| Preflights | ✅ (run) | ✅ | ❌ | ❌ |

---

## 4. Code Quality

### 4.1 Go Idioms & Patterns

**Good practices**:
- Uses Go 1.22+ stdlib mux with method+path routing (no third-party router)
- sqlc for type-safe database access (generated code)
- `//go:embed` for static assets
- `context.Context` propagation through handlers
- Structured logging via `log/slog`
- Proper use of `sql.NullInt64`, `sql.NullString`, `sql.NullTime`
- Transaction usage for create-project-with-seed operations
- Clean JSON API patterns (`jsonOK`, `jsonErr` helpers)

**Patterns observed**:
- Handler methods on `*Server` struct (dependency injection via struct)
- Background goroutines for long-running conversions (PDF, EPUB)
- Throttled notifications with `sync.Mutex` + time map
- Embedded default data as Go string literals (transmittal template, spec defaults)

### 4.2 Error Handling

**Generally good**: Most database errors are caught and returned as JSON errors with appropriate HTTP status codes.

**Gaps**:
- `runConversion()` and `runEPUBGeneration()` run in goroutines but only log errors — no way for the client to poll for completion status (only book status field)
- Some error paths silently ignore errors (e.g., `_ = q.UpdateProject(...)` in seed handler)
- `json.NewEncoder(w).Encode()` errors are never checked
- `w.Write(data)` return values ignored in template serving

### 4.3 Input Validation

**Present but minimal**:
- Email addresses validated with basic `@` and `.` check
- Project/client slugs normalized via `normalizeProjectSlug()`
- Correction status validated against whitelist
- Book upload has 50MB limit
- Cover upload has 10MB limit and JPEG/PNG type check

**Missing**:
- No validation on date formats (accepts any string)
- No validation on task field ranges (negative hours, etc.)
- No sanitization of JSON blob content in transmittals/specs
- No length limits on text fields
- Body decode errors return generic "bad request" without specifics

### 4.4 Security Considerations

**SQL Injection**: ✅ **Safe** — All database access uses parameterized queries (sqlc generated code + `?` placeholders for raw queries).

**XSS**: ⚠️ **Mixed**
- Server-rendered HTML emails use `html.EscapeString()` ✅
- SPA frontend constructs DOM via `document.createElement()` (safe) rather than `innerHTML` ✅
- But: the preflight report HTML is served directly from DB (`w.Write([]byte(row.ReportHtml))`) — if the Python script or docx contains malicious content, it could be XSS risk ⚠️

**CSRF**: ❌ **No CSRF protection**
- No CSRF tokens
- State-changing operations (POST/PUT/DELETE) rely only on cookies + JSON Content-Type
- The `SameSite=Lax` cookie setting provides some protection against cross-site POST, but not against same-site attacks

**Auth**:
- Passwords stored as SHA-256 hashes (no salt, no bcrypt/scrypt/argon2) ⚠️
- Auth cookies store the **plaintext password** as the cookie value, and re-hash on each request ⚠️
- Cookie `MaxAge` of 1 year for project auth, 90 days for client auth
- No rate limiting on password attempts
- Hardcoded admin email recipient: `j@djinna.com`
- `.env` file contains API key in plaintext (committed to analysis but gitignored)

**Path Traversal**: The `sanitizeFilename()` function strips dangerous characters from download filenames ✅

**Secrets**: The `.env` file with `AGENTMAIL_API_KEY` is gitignored ✅ but present on disk.

### 4.5 Test Coverage

**2,549 lines of tests** across 11 test files:
- `server_test.go` — helper `testServer()` setup, health check
- `admin_test.go` — admin project/client list, create client
- `auth_test.go` — auth flow (set/verify/clear passwords, open access)
- `client_test.go` — client verify, projects, file log, journal
- `bookspecs_test.go` — spec CRUD, pull from transmittal, cover upload
- `preflight_test.go` — preflight run with mocked runner
- `filelog_test.go` — file log CRUD
- `journal_test.go` — journal CRUD
- `task_test.go` — task CRUD
- `transmittal_test.go` — transmittal CRUD, versions, duplicate
- `email_test.go` — email status check
- `custom_styles_roundtrip_test.go` — custom style mapping
- `word_template_test.go` — word template generation

**All tests pass** (cached). Tests use in-memory SQLite databases with full migration setup.

**Not tested**:
- Book upload/conversion (requires pandoc/typst binaries)
- EPUB generation (requires pandoc binary)
- Email sending (would need mock HTTP server)
- The actual frontend JavaScript
- Concurrent access patterns

---

## 5. Infrastructure

### 5.1 Deployment

- **systemd service** (`srv.service`): runs as `exedev` user, `WorkingDirectory=/home/exedev`
- Binary: `/home/exedev/prodcal/prodcal` (17MB, statically compiled)
- **Currently INACTIVE** — service is dead due to port 8000 bind conflict (another process holds it)
- Build: `make build` → `go build -o prodcal ./cmd/srv`
- Restart: `make build && sudo systemctl restart srv`

### 5.2 Database

- **Production DB**: `/home/exedev/db.sqlite3` (109MB, fully migrated through 014)
- **Stale copy**: `/home/exedev/prodcal/db.sqlite3` (80KB, only migrated through 006)
- **Empty files**: `prodcal.db`, `prodcal.sqlite3` (0 bytes)
- The binary opens `db.sqlite3` relative to CWD, and systemd sets `WorkingDirectory=/home/exedev`
- Daily backups at 3 AM to `~/backups/` with 7-day retention
- Contains 4 books with source docx blobs (~7MB each) + generated PDFs (~8MB each)

### 5.3 External Dependencies (Runtime)

- **pandoc**: Used for docx→typst and docx→epub3 conversion
- **typst**: Used for typst→PDF compilation
- **Python 3**: Used for manuscript preflight (`detect-edge-cases.py`) and Word template generation (`generate-word-template.py`)
- Scripts located in `/home/exedev/book-production/scripts/`
- Fonts in `/home/exedev/book-production/fonts/`
- Templates in `/home/exedev/book-production/templates/`

### 5.4 Static Assets

All embedded in the binary via `//go:embed`:
- `style.css` (1,156 lines) — shared design system
- `transmittal.css` (873 lines) — transmittal-specific styles
- `app.js` (1,295 lines) — calendar SPA
- `transmittal.js` (1,295 lines) — transmittal SPA
- 8 HTML files (various)

---

## 6. Issues & Gaps

### 6.1 Critical Issues

1. **Service is DOWN**: `srv.service` is inactive due to port 8000 bind conflict. The restart counter was at 798 attempts.

2. **Stale DB in repo**: `/home/exedev/prodcal/db.sqlite3` has only 6 of 14 migrations applied and appears to be an old copy. The real production DB is at `/home/exedev/db.sqlite3`. If anyone runs the server from the prodcal directory instead of /home/exedev, they'll use the wrong database.

3. **Password stored in cookies**: Auth cookies contain the plaintext password as their value. The server hashes it on each request to compare. If cookies are intercepted, the password is exposed.

4. **Unsalted SHA-256 for passwords**: Should use bcrypt/scrypt/argon2 with per-password salts.

### 6.2 Security Issues

5. **No CSRF protection**: All state-changing endpoints lack CSRF tokens.

6. **No rate limiting**: Password verification endpoints have no brute-force protection.

7. **Preflight report XSS**: `handleGetManuscriptPreflightReport` serves HTML directly from the database with no sanitization.

8. **Hardcoded notification recipient**: `j@djinna.com` is hardcoded in `transmittal_notify.go`.

### 6.3 Data Integrity Issues

9. **109MB SQLite with BLOBs**: Storing multi-MB docx/pdf/epub files as SQLite BLOBs is a scaling concern. The DB is already 109MB with only 4 books.

10. **No vacuum/optimization**: No VACUUM or OPTIMIZE scheduled for the SQLite database.

11. **Background conversion status**: PDF/EPUB conversions run in goroutines but there's no WebSocket/SSE/polling mechanism. The client must manually refresh to see if conversion completed.

### 6.4 Missing Features

12. **No file log update**: File log entries can only be created and deleted, not edited.

13. **No journal update**: Journal entries can only be created and deleted, not edited.

14. **No client update/delete**: Clients can only be created, never modified or removed.

15. **No book spec delete from API**: Only available via raw sqlc, not exposed as an endpoint.

16. **No pagination**: All list endpoints return all records. The admin project list, task list, file log, and journal could grow large.

17. **No search/filter**: No text search across projects, tasks, journal entries, etc.

18. **No user/role management**: All admins are equal (any exe.dev user is admin). No user-specific permissions.

### 6.5 Code Quality Issues

19. **Raw SQL in handlers**: Several handlers (admin project list, admin client list, archive/restore, transmittal handlers, corrections, file log, journal) use raw SQL strings instead of sqlc-generated queries. This bypasses the type-safety that sqlc provides.

20. **Duplicated HTML builders**: The email HTML generation code (~2,500 lines across 4 files) builds HTML via string concatenation. This is error-prone and hard to maintain. A template engine would be better.

21. **Admin HTML is 2,780 lines**: The admin dashboard is a massive single HTML file with inline JS. This is hard to maintain.

22. **Duplicated SPA utilities**: `app.js` and `transmittal.js` share identical `$`, `$$`, `h`, and `api` helper functions (copy-pasted). Should be shared.

23. **Hardcoded paths**: `bookProdRoot = "/home/exedev/book-production"` and font/template paths are hardcoded constants in `books.go`.

24. **Standard workflow hardcoded**: The 31-task seed workflow with specific dates (2026-02-23 through 2026-07-09) and assignees (NW, JD, VR, PR) is hardcoded in `server.go`.

25. **`UpdateBookStatus` error field**: The sqlc query `UpdateBookStatus` takes an `error_msg` param, but the `error_msg` field is only set on failure. On conversion start (`status = "converting"`), the previous error message remains.

### 6.6 Operational Issues

26. **No graceful shutdown**: The server uses `http.ListenAndServe` which doesn't support graceful shutdown. Background goroutines (conversions) could be interrupted.

27. **No request logging**: There's no HTTP request logging middleware. Only specific events are logged via slog.

28. **No metrics/monitoring**: No Prometheus/StatsD/etc. metrics exposed.

29. **No connection pooling config**: SQLite is single-writer; the `database/sql` pool defaults may cause contention under load.

30. **Docker not used**: Dockerfile exists but is not used in production. The runtime dependencies (pandoc, typst, python3) are not containerized.

---

## 7. Code Statistics

| Category | Lines |
|----------|-------|
| Go (hand-written) | ~9,800 |
| Go (sqlc generated) | ~1,300 |
| Go (tests) | ~2,550 |
| **Total Go** | **~11,100** |
| HTML templates | ~4,640 |
| JavaScript | ~2,590 |
| CSS | ~2,030 |
| SQL (migrations) | ~180 |
| SQL (queries) | ~120 |
| **Total frontend** | **~9,260** |
| **Grand total** | **~20,500** |

## 8. Database Content (Production)

| Table | Records |
|-------|--------|
| Projects | 12 (5 active, 7 archived) |
| Tasks | 166 |
| Transmittals | 9 |
| Books | 4 |
| Book Outputs | 6 |
| Corrections | 1 |
| File Log | 4 |
| Journal | 4 |
| Manuscript Preflights | 12 |
| Clients | 3 (vgr, sw, pi) |
| Auth Tokens | (per-project) |
| Migrations | 14 |

Database size: 109MB (dominated by book BLOB data)
