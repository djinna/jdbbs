# ProdCal Server Handler Summary

All handlers live in `/home/exedev/prodcal/srv/` spread across multiple files.

---

## Auth System Overview

Three auth levels exist:

1. **exe.dev Admin** — `X-ExeDev-UserID` header (set by proxy). Checked via `requireExeDevAdmin` (HTML redirect) or `requireExeDevAdminAPI` (JSON 401).
2. **Project-level auth** — SHA-256-hashed password tokens in `auth_tokens` table. Checked via cookie `prodcal_auth_{projectID}` or `X-Auth-Token` header.
3. **Client-level auth** — SHA-256-hashed password in `clients` table. Checked via cookie `prodcal_client_{clientSlug}`.

`checkAuth(r, projectID)` cascades: project token → client cookie → `X-ExeDev-UserID` header. If no tokens exist for a project, access is open.

---

## Health Check

### `handleHealthz`
- **Route:** `GET /healthz`
- **File:** `server.go`
- **Auth:** None
- **Does:** Pings DB with `SELECT 1`
- **Response:** `200 {"status":"ok"}` or `503 {"status":"error","detail":"..."}`

---

## Admin Handlers (admin.go)

### `handleAdminDashboard`
- **Route:** `GET /admin/`
- **Auth:** exe.dev admin (HTML redirect to login if missing)
- **Does:** Serves `static/admin.html`
- **Response:** HTML page

### `handleAdminProjectList`
- **Route:** `GET /api/admin/projects`
- **Auth:** exe.dev admin API (401 JSON)
- **Query params:** `archived=1` (optional, filters to archived projects only)
- **Does:** Lists all projects with task counts, auth status, transmittal status
- **Response:** `200` JSON array of `projectSummary` objects:
  ```json
  [{"id", "name", "client_slug", "project_slug", "start_date", "created_at", "updated_at", "archived_at", "task_count", "done_count", "active_count", "has_auth", "has_transmittal", "transmittal_status", "path"}]
  ```

### `handleAdminClientList`
- **Route:** `GET /api/admin/clients`
- **Auth:** exe.dev admin API
- **Does:** Lists all clients with project counts
- **Response:** `200` JSON array of `adminClientSummary`:
  ```json
  [{"slug", "name", "has_auth", "project_count", "created_at"}]
  ```

### `handleAdminCreateClient`
- **Route:** `POST /api/admin/clients`
- **Auth:** exe.dev admin API
- **Request JSON:** `{"name": str, "slug": str, "password": str (optional)}`
- **Does:** Creates a new client. Slug is normalized. Password is SHA-256 hashed.
- **Response:** `201 {"slug", "name", "has_auth", "created_ok"}` or `409` if slug exists

---

## Project Handlers (server.go)

### `handleListProjects`
- **Route:** `GET /api/projects`
- **Auth:** None
- **Does:** Lists all projects
- **Response:** `200` JSON array of `dbgen.Project` objects

### `handleCreateProject`
- **Route:** `POST /api/projects`
- **Auth:** exe.dev admin API
- **Request JSON:** `{"name": str, "start_date": str, "client_slug": str, "project_slug": str}`
- **Does:** Creates project in a transaction, then seeds with standard 31-task book production workflow
- **Response:** `201` project object or `400`/`500`

### `handleGetProject`
- **Route:** `GET /api/projects/{id}`
- **Auth:** None (returns auth status for the frontend to decide)
- **Path params:** `id` (project ID)
- **Does:** Returns project with auth info
- **Response:** `200 {"project": {...}, "has_auth": bool, "authenticated": bool}` or `404`

### `handleGetProjectByPath`
- **Route:** `GET /api/project-by-path/{client}/{project}`
- **Auth:** None (returns auth status)
- **Path params:** `client` (client slug), `project` (project slug)
- **Does:** Looks up project by client+project slug pair
- **Response:** `200 {"project": {...}, "has_auth": bool, "authenticated": bool}` or `404`

### `handleUpdateProject`
- **Route:** `PUT /api/projects/{id}`
- **Auth:** Project-level (`requireAuth`)
- **Path params:** `id`
- **Request JSON:** `{"name": str, "start_date": str, "client_slug": str, "project_slug": str}` — slugs preserved if omitted
- **Response:** `200 {"ok":"true"}` or `401`/`500`

### `handleDeleteProject`
- **Route:** `DELETE /api/projects/{id}`
- **Auth:** N/A
- **Does:** Always returns error — deletion is disabled
- **Response:** `405 {"error":"project deletion disabled; archive instead"}`

### `handleArchiveProject`
- **Route:** `POST /api/projects/{id}/archive`
- **Auth:** exe.dev admin API
- **Path params:** `id`
- **Does:** Sets `archived_at` timestamp
- **Response:** `200 {"ok":"true"}`

### `handleRestoreProject`
- **Route:** `POST /api/projects/{id}/restore`
- **Auth:** exe.dev admin API
- **Path params:** `id`
- **Does:** Clears `archived_at`
- **Response:** `200 {"ok":"true"}`

### `handleSeedProject`
- **Route:** `POST /api/projects/{id}/seed`
- **Auth:** Project-level
- **Path params:** `id`
- **Request JSON:**
  ```json
  {
    "project_name": str, "project_start": str,
    "tasks": [{"sort_order", "assignee", "task", "is_milestone", "orig_weeks", "curr_weeks", "orig_due", "curr_due", "actual_done", "words", "words_per_hour", "hours", "rate", "budget_notes", "orig_budget", "curr_budget", "actual_budget"}]
  }
  ```
- **Does:** Bulk-creates tasks from seed data. Optionally updates project start date.
- **Response:** `200 {"ok":true, "count": N}`

### `handleDuplicateProject`
- **Route:** `POST /api/projects/{id}/duplicate`
- **Auth:** Project-level (on source)
- **Path params:** `id` (source project)
- **Request JSON:** `{"name": str, "start_date": str, "client_slug": str, "project_slug": str}`
- **Does:** Creates new project, copies all tasks with date-shifted schedules. Resets status to pending, zeroes budgets.
- **Response:** `200 {"project": {...}, "tasks_copied": N}`

---

## Task Handlers (server.go)

### `handleListTasks`
- **Route:** `GET /api/projects/{id}/tasks`
- **Auth:** Project-level
- **Path params:** `id` (project ID)
- **Response:** `200` JSON array of `dbgen.Task`

### `handleCreateTask`
- **Route:** `POST /api/projects/{id}/tasks`
- **Auth:** Project-level
- **Path params:** `id` (project ID)
- **Request JSON:** `taskInput` struct:
  ```json
  {"sort_order", "assignee", "title", "is_milestone", "orig_weeks", "curr_weeks", "orig_due", "curr_due", "actual_done", "status", "words", "words_per_hour", "hours", "rate", "budget_notes", "orig_budget", "curr_budget", "actual_budget"}
  ```
  `status` defaults to `"pending"` if empty.
- **Response:** `201` task object

### `handleUpdateTask`
- **Route:** `PUT /api/tasks/{id}`
- **Auth:** Project-level (looked up from task's project)
- **Path params:** `id` (task ID)
- **Request JSON:** Same `taskInput` struct
- **Response:** `200 {"ok":"true"}` or `404`

### `handleDeleteTask`
- **Route:** `DELETE /api/tasks/{id}`
- **Auth:** Project-level (looked up from task's project)
- **Path params:** `id` (task ID)
- **Response:** `200 {"ok":"true"}` or `404`

---

## Auth Handlers (server.go)

### `handleSetAuth`
- **Route:** `POST /api/projects/{id}/auth`
- **Auth:** If tokens exist, must be authenticated. If no tokens, open access (first-time setup).
- **Path params:** `id`
- **Request JSON:** `{"password": str}`
- **Does:** Creates auth token (SHA-256 hash), sets HttpOnly cookie `prodcal_auth_{id}` (1 year)
- **Response:** `200 {"ok":"true"}`

### `handleClearAuth`
- **Route:** `DELETE /api/projects/{id}/auth`
- **Auth:** exe.dev admin API
- **Path params:** `id`
- **Does:** Deletes all auth tokens for project, clears cookie
- **Response:** `200 {"ok":"true"}`

### `handleVerifyAuth`
- **Route:** `POST /api/projects/{id}/verify`
- **Auth:** None (this IS the auth endpoint)
- **Path params:** `id`
- **Request JSON:** `{"password": str}`
- **Does:** Verifies password against stored hash, sets cookie on success
- **Response:** `200 {"ok":"true"}` or `401 {"error":"invalid password"}`

---

## Transmittal Handlers (transmittal.go)

### `handleGetTransmittal`
- **Route:** `GET /api/projects/{id}/transmittal`
- **Auth:** Project-level
- **Path params:** `id` (project ID)
- **Does:** Returns transmittal for project; auto-creates with defaults if missing
- **Response:** `200 {"id", "project_id", "status", "data": <raw JSON>, "created_at", "updated_at"}`

### `handleUpdateTransmittal`
- **Route:** `PUT /api/projects/{id}/transmittal`
- **Auth:** Project-level
- **Path params:** `id` (project ID)
- **Request JSON:** `{"status": str, "data": <JSON object>}` — status defaults to "draft"
- **Does:** Snapshots current version (throttled to 5-min intervals), then updates. Triggers admin notification email (throttled 30 min) if not an exe.dev user.
- **Response:** `200 {"ok":true}`

### `handleListTransmittalVersions`
- **Route:** `GET /api/transmittals/{id}/versions`
- **Auth:** Project-level (note: `{id}` here is treated as project ID in the route match)
- **Path params:** `id` (project ID)
- **Does:** Lists up to 100 transmittal version snapshots
- **Response:** `200` JSON array of `{"id", "status", "title", "saved_at"}`

### `handleGetTransmittalVersion`
- **Route:** `GET /api/transmittals/{id}/versions/{vid}`
- **Auth:** Project-level
- **Path params:** `id` (project ID), `vid` (version ID)
- **Response:** `200 {"id", "status", "data": <raw JSON>, "saved_at"}` or `404`

### `handleRestoreTransmittalVersion`
- **Route:** `POST /api/transmittals/{id}/versions/{vid}/restore`
- **Auth:** Project-level
- **Path params:** `id` (project ID), `vid` (version ID)
- **Does:** Snapshots current state (bypasses throttle), then overwrites with version data
- **Response:** `200 {"ok":true}` or `404`

### `handleDuplicateTransmittal`
- **Route:** `POST /api/transmittals/{id}/duplicate`
- **Auth:** Project-level (both source and target)
- **Path params:** `id` (source project ID)
- **Request JSON:** `{"target_project_id": int64}`
- **Does:** Copies transmittal to target project, clearing book-specific fields (title, ISBNs, dates, checklist statuses) while keeping publisher/design defaults
- **Response:** `200 {"ok":true, "target_project_id": N}` or `409` if target already has one

---

## Books Handlers (books.go)

### `handleListBooks`
- **Route:** `GET /api/books`
- **Auth:** exe.dev admin API
- **Does:** Lists all books (without blob data)
- **Response:** `200` JSON array of book metadata

### `handleUploadBook`
- **Route:** `POST /api/books/upload`
- **Auth:** exe.dev admin API
- **Request:** Multipart form: `file` (required, max 50MB), `title` (required), `author` (required), `series` (optional), `project_id` (optional)
- **Response:** `201 {"id", "title", "author", "status"}`

### `handleConvertBook`
- **Route:** `POST /api/books/{id}/convert`
- **Auth:** exe.dev admin API
- **Path params:** `id` (book ID)
- **Does:** Kicks off async docx → Typst → PDF pipeline (pandoc + typst compile). Marks book as "converting".
- **Response:** `200 {"status":"converting"}` (async — check book status later)

### `handleDownloadBook`
- **Route:** `GET /api/books/{id}/download/{format}`
- **Auth:** None (public download)
- **Path params:** `id` (book ID), `format` (`pdf` or `epub`)
- **Response:** Binary file download with `Content-Disposition` header, or `404`

### `handleLinkBookProject`
- **Route:** `PUT /api/books/{id}/project`
- **Auth:** exe.dev admin API
- **Path params:** `id` (book ID)
- **Request JSON:** `{"project_id": int64|null}`
- **Does:** Links/unlinks a book to a project
- **Response:** `200 {"ok":"true"}`

### `handleDeleteBook`
- **Route:** `DELETE /api/books/{id}`
- **Auth:** exe.dev admin API
- **Path params:** `id` (book ID)
- **Response:** `200 {"ok":"true"}`

---

## Book Spec Handlers (bookspecs.go)

### `handleGetBookSpec`
- **Route:** `GET /api/projects/{id}/book-spec`
- **Auth:** exe.dev admin API
- **Path params:** `id` (project ID)
- **Does:** Returns book spec; auto-creates with extensive defaults if missing. Checks for cover image.
- **Response:** `200 {"id", "project_id", "data": <raw JSON>, "has_cover": bool, "created_at", "updated_at"}`

### `handleUpdateBookSpec`
- **Route:** `PUT /api/projects/{id}/book-spec`
- **Auth:** exe.dev admin API
- **Path params:** `id` (project ID)
- **Request JSON:** `{"data": <JSON object>}` — the full spec data blob
- **Response:** `200 {"ok":true, "updated_at": str}`

### `handlePullTransmittalToSpec`
- **Route:** `POST /api/projects/{id}/book-spec/pull-transmittal`
- **Auth:** exe.dev admin API
- **Path params:** `id` (project ID)
- **Does:** Imports fields from the project's transmittal into its book spec (metadata, design, front/back matter checklist, custom styles, etc.)
- **Response:** `200 {"ok":true, "data": <merged spec JSON>}`

### `handleGenerateConfig`
- **Route:** `POST /api/projects/{id}/book-spec/generate-config`
- **Auth:** exe.dev admin API
- **Path params:** `id` (project ID)
- **Does:** Converts the spec into Typst config override code
- **Response:** `200 {"config": "<typst code string>"}`

### `handleUploadCover`
- **Route:** `POST /api/projects/{id}/book-spec/cover`
- **Auth:** exe.dev admin API
- **Path params:** `id` (project ID)
- **Request:** Multipart form: `cover` file (JPEG or PNG, max 10MB)
- **Does:** Stores cover image in `book_specs` table
- **Response:** `200 {"ok":true, "size": N, "type": str}`

### `handleGetCover`
- **Route:** `GET /api/projects/{id}/book-spec/cover`
- **Auth:** None (public)
- **Path params:** `id` (project ID)
- **Response:** Raw image bytes with appropriate Content-Type, or `404`

### `handleListFonts`
- **Route:** `GET /api/fonts`
- **Auth:** exe.dev admin API
- **Does:** Runs `typst fonts --font-path <fontsDir>` and categorizes results
- **Response:** `200` JSON array of `{"family": str, "category": str}` where category is serif/sans-serif/monospace/other

### `handleGenerateWordTemplate`
- **Route:** `POST /api/projects/{id}/book-spec/word-template`
- **Auth:** exe.dev admin API
- **Path params:** `id` (project ID)
- **Does:** Runs Python script to generate a styled .docx template from the book spec. Validates no duplicate custom style names.
- **Response:** Binary .docx file download, or `400` (duplicate styles) / `404` / `500`

---

## Preflight Handlers (preflight.go)

### `handleRunManuscriptPreflight`
- **Route:** `POST /api/projects/{id}/preflight`
- **Auth:** exe.dev admin API
- **Path params:** `id` (project ID)
- **Request JSON:** `{"book_id": int64}`
- **Does:** Writes book's source .docx to temp file, runs Python edge-case detector script, appends undeclared style warnings by comparing against book spec. Stores HTML+JSON report in DB.
- **Response:** `200` `preflightResponse`:
  ```json
  {"exists":true, "project_id", "book_id", "status": "ready"|"error", "source_filename", "updated_at", "summary": {"total", "high", "medium", "low", "by_type"}, "images": [...], "report_url", "history": [{"id","status","updated_at","report_url","latest"}], "error": str|null}
  ```

### `handleGetManuscriptPreflight`
- **Route:** `GET /api/projects/{id}/preflight`
- **Auth:** exe.dev admin API
- **Path params:** `id` (project ID)
- **Query params:** `book_id` (required, int)
- **Does:** Returns latest preflight result + history for a book
- **Response:** `200` `preflightResponse` (same shape as above, or `{"exists":false}`)

### `handleGetManuscriptPreflightReport`
- **Route:** `GET /api/projects/{id}/preflight/report`
- **Auth:** exe.dev admin API
- **Path params:** `id` (project ID)
- **Query params:** `book_id` (required), `preflight_id` (optional — defaults to latest)
- **Does:** Serves the stored HTML preflight report
- **Response:** `text/html` document, or `404`

---

## EPUB Handlers (epub.go)

### `handleGenerateEPUB`
- **Route:** `POST /api/books/{id}/generate-epub`
- **Auth:** exe.dev admin API
- **Path params:** `id` (book ID)
- **Does:** Kicks off async EPUB generation via pandoc (docx → epub3). Uses book spec for metadata, cover image, custom CSS, TOC depth, section breaks.
- **Response:** `200 {"status":"generating_epub"}` (async)

---

## Email Handlers

### `handleEmailStatus` (email.go)
- **Route:** `GET /api/email/status`
- **Auth:** None
- **Does:** Reports whether email is configured
- **Response:** `200 {"configured": bool}`

### `handleSendTransmittalEmail` (email.go)
- **Route:** `POST /api/projects/{id}/transmittal/email`
- **Auth:** Project-level
- **Path params:** `id` (project ID)
- **Request JSON:** `{"recipients": ["email@...", ...]}`
- **Does:** Loads transmittal data, builds text+HTML summary, sends via AgentMail API. First recipient is "to", rest are "cc".
- **Response:** `200 {"ok":true, "sent_to": [...], "subject": str}` or `503` (email not configured)

### `handleSendProjectSnapshot` (snapshot_email.go)
- **Route:** `POST /api/projects/{id}/snapshot/email`
- **Auth:** Project-level
- **Path params:** `id` (project ID)
- **Request JSON:** `{"recipients": ["email@...", ...]}`
- **Does:** Builds comprehensive project snapshot email including: schedule overview (% complete, task table), budget summary (orig/curr/actual), transmittal status, recent 10 file log entries, recent 10 journal entries. Both HTML and plain text.
- **Response:** `200 {"ok":true, "sent_to": [...], "subject": str}` or `503`

### `handleSendActivityEmail` (activity_email.go)
- **Route:** `POST /api/projects/{id}/activity/email`
- **Auth:** Project-level
- **Path params:** `id` (project ID)
- **Query params:** `days` (optional, default 7, max 90)
- **Request JSON:** `{"recipients": ["email@...", ...]}`
- **Does:** Sends file log + journal entries from the last N days
- **Response:** `200 {"ok":true, "sent_to": [...], "subject": str}` or `503`

### `handleSendClientDigest` (client_digest_email.go)
- **Route:** `POST /api/clients/{client}/digest/email`
- **Auth:** Client-level or any project auth for client
- **Path params:** `client` (client slug)
- **Request JSON:** `{"recipients": ["email@...", ...]}`
- **Does:** Aggregates last 7 days of activity across ALL projects for a client. Shows per-project file log and journal entries.
- **Response:** `200 {"ok":true, "sent_to": [...], "subject": str}` or `503`

---

## File Log Handlers (filelog.go)

### `handleListFileLog`
- **Route:** `GET /api/projects/{id}/file-log`
- **Auth:** Project-level
- **Path params:** `id` (project ID)
- **Response:** `200` JSON array of `fileLogEntry`:
  ```json
  [{"id", "project_id", "direction", "filename", "file_type", "sent_by", "received_by", "notes", "transfer_date", "created_at"}]
  ```

### `handleCreateFileLog`
- **Route:** `POST /api/projects/{id}/file-log`
- **Auth:** Project-level
- **Path params:** `id` (project ID)
- **Request JSON:** `{"direction": str (default "inbound"), "filename": str, "file_type": str, "sent_by": str, "received_by": str, "notes": str, "transfer_date": str (default today)}`
- **Response:** `201` fileLogEntry object

### `handleDeleteFileLog`
- **Route:** `DELETE /api/projects/{id}/file-log/{entry}`
- **Auth:** Project-level
- **Path params:** `id` (project ID), `entry` (entry ID)
- **Response:** `200 {"ok":"true"}`

### `handleClientFileLog`
- **Route:** `GET /api/clients/{client}/file-log`
- **Auth:** Client-level or project-level for any project in client
- **Path params:** `client` (client slug)
- **Query params:** `limit` (optional, default 20, max 100)
- **Does:** Returns recent file log entries across all projects for this client
- **Response:** `200` JSON array of fileLogEntry + `project_name` field

---

## Journal Handlers (journal.go)

### `handleListJournal`
- **Route:** `GET /api/projects/{id}/journal`
- **Auth:** Project-level
- **Path params:** `id` (project ID)
- **Response:** `200` JSON array of `journalEntry`:
  ```json
  [{"id", "project_id", "entry_type", "content", "created_at"}]
  ```

### `handleCreateJournal`
- **Route:** `POST /api/projects/{id}/journal`
- **Auth:** Project-level
- **Path params:** `id` (project ID)
- **Request JSON:** `{"entry_type": str (default "note"), "content": str (required)}`
- **Response:** `201` journalEntry object or `400` if content empty

### `handleDeleteJournal`
- **Route:** `DELETE /api/projects/{id}/journal/{entry}`
- **Auth:** Project-level
- **Path params:** `id` (project ID), `entry` (entry ID)
- **Response:** `200 {"ok":"true"}`

### `handleClientJournal`
- **Route:** `GET /api/clients/{client}/journal`
- **Auth:** Client-level or project-level for any project in client
- **Path params:** `client` (client slug)
- **Query params:** `limit` (optional, default 20, max 100)
- **Does:** Returns recent journal entries across all client projects
- **Response:** `200` JSON array of journalEntry + `project_name` field

---

## Corrections Handlers (corrections.go)

### `handleListCorrections`
- **Route:** `GET /api/projects/{id}/corrections`
- **Auth:** Project-level
- **Path params:** `id` (project ID)
- **Response:** `200` JSON array of `correctionEntry`:
  ```json
  [{"id", "project_id", "find_text", "replace_text", "chapter", "note", "status", "applied_at", "created_at"}]
  ```

### `handleCreateCorrection`
- **Route:** `POST /api/projects/{id}/corrections`
- **Auth:** Project-level
- **Path params:** `id` (project ID)
- **Request JSON:** `{"find_text": str (required), "replace_text": str, "chapter": str, "note": str}`
- **Response:** `201` correctionEntry object

### `handleDeleteCorrection`
- **Route:** `DELETE /api/projects/{id}/corrections/{entry}`
- **Auth:** Project-level
- **Path params:** `id` (project ID), `entry` (correction ID)
- **Response:** `200 {"ok":"true"}`

### `handleUpdateCorrectionStatus`
- **Route:** `PUT /api/corrections/{id}/status`
- **Auth:** Project-level (looked up from correction's project)
- **Path params:** `id` (correction ID)
- **Request JSON:** `{"status": "pending"|"applied"|"skipped"}`
- **Does:** Updates status; sets `applied_at` to today if status is "applied"
- **Response:** `200 {"ok":"true"}` or `400` (invalid status) / `404`

### `handleExportCorrections`
- **Route:** `POST /api/projects/{id}/corrections/export`
- **Auth:** Project-level
- **Path params:** `id` (project ID)
- **Does:** Generates YAML file of all corrections
- **Response:** `application/x-yaml` file download (`corrections-project-{id}.yaml`)

---

## Client Handlers (client.go)

### `handleClientInfo`
- **Route:** `GET /api/clients/{client}`
- **Auth:** None (returns auth status)
- **Path params:** `client` (client slug)
- **Does:** Returns client info and whether caller is authenticated
- **Response:** `200 {"slug", "name", "has_auth": bool, "authenticated": bool}` or `404`

### `handleClientVerify`
- **Route:** `POST /api/clients/{client}/verify`
- **Auth:** None (this IS the auth endpoint)
- **Path params:** `client` (client slug)
- **Request JSON:** `{"password": str}`
- **Does:** Verifies password, sets `prodcal_client_{slug}` cookie (90 days)
- **Response:** `200 {"ok":true}` or `401`

### `handleClientProjects`
- **Route:** `GET /api/clients/{client}/projects`
- **Auth:** Client-level or any project auth for client
- **Path params:** `client` (client slug)
- **Does:** Lists non-archived projects for client with task counts, transmittal status
- **Response:** `200` JSON array of:
  ```json
  [{"id", "name", "client_slug", "project_slug", "start_date", "updated_at", "task_count", "done_count", "active_count", "has_transmittal", "transmittal_status", "path"}]
  ```

### `handleClientCreateProject`
- **Route:** `POST /api/clients/{client}/projects`
- **Auth:** Client-level (if password-protected)
- **Path params:** `client` (client slug)
- **Request JSON:** `{"name": str (required), "start_date": str, "project_slug": str (auto-derived from name if omitted)}`
- **Does:** Creates project under client, seeds with standard workflow. Normalizes slug.
- **Response:** `201` project object or `409` (slug exists)

---

## Handler Count Summary

| Category | Count |
|----------|-------|
| Health | 1 |
| Admin | 4 |
| Projects | 9 |
| Tasks | 4 |
| Auth | 3 |
| Transmittal | 6 |
| Books | 6 |
| Book Spec | 8 |
| Preflight | 3 |
| EPUB | 1 |
| Email | 5 |
| File Log | 4 |
| Journal | 4 |
| Corrections | 5 |
| Client | 4 |
| **Total** | **67** |
