# ProdCal API Reference

Comprehensive reference for all HTTP endpoints in ProdCal.  
**Base URL:** `https://<hostname>.exe.xyz:8000`  
**Content-Type:** All JSON endpoints accept and return `application/json` unless noted.  
**Error format:** `{"error": "message"}` with appropriate HTTP status code.

## Authentication Model

ProdCal has three authentication levels:

| Level | Mechanism | Grants access to |
|-------|-----------|------------------|
| **Admin** | exe.dev proxy sets `X-ExeDev-UserID` header | All endpoints |
| **Project** | Cookie `prodcal_auth_{id}` or `X-Auth-Token` header | Project-scoped data |
| **Client** | Cookie `prodcal_client_{slug}` | Client's projects & data |

Endpoints with no auth configured on the resource are open-access.

---

## Quick Reference

| # | Method | Path | Auth | Description |
|---|--------|------|------|-------------|
| 1 | GET | `/healthz` | None | Health check |
| 2 | GET | `/admin/` | Admin | Admin dashboard HTML |
| 3 | GET | `/api/admin/projects` | Admin | List all projects with stats |
| 4 | GET | `/api/admin/clients` | Admin | List all clients with stats |
| 5 | POST | `/api/admin/clients` | Admin | Create a client |
| 6 | GET | `/api/projects` | ⚠️ None | List all non-archived projects |
| 7 | POST | `/api/projects` | Admin | Create a project |
| 8 | GET | `/api/projects/{id}` | None | Get project (includes auth status) |
| 9 | PUT | `/api/projects/{id}` | Project | Update project metadata |
| 10 | DELETE | `/api/projects/{id}` | — | Disabled (returns 405) |
| 11 | POST | `/api/projects/{id}/archive` | Admin | Archive a project |
| 12 | POST | `/api/projects/{id}/restore` | Admin | Restore an archived project |
| 13 | GET | `/api/projects/{id}/tasks` | Project | List tasks for a project |
| 14 | POST | `/api/projects/{id}/tasks` | Project | Create a task |
| 15 | PUT | `/api/tasks/{id}` | Project | Update a task |
| 16 | DELETE | `/api/tasks/{id}` | Project | Delete a task |
| 17 | POST | `/api/projects/{id}/auth` | Project* | Set project password |
| 18 | DELETE | `/api/projects/{id}/auth` | Admin | Clear all project auth tokens |
| 19 | POST | `/api/projects/{id}/verify` | None | Verify password & set cookie |
| 20 | POST | `/api/projects/{id}/seed` | Project | Bulk-import tasks from JSON |
| 21 | POST | `/api/projects/{id}/duplicate` | Project | Clone project with shifted dates |
| 22 | GET | `/api/project-by-path/{client}/{project}` | None | Look up project by slug pair |
| 23 | GET | `/api/projects/{id}/transmittal` | Project | Get transmittal (auto-creates) |
| 24 | PUT | `/api/projects/{id}/transmittal` | Project | Update transmittal data |
| 25 | GET | `/api/transmittals/{id}/versions` | Project | List transmittal version history |
| 26 | GET | `/api/transmittals/{id}/versions/{vid}` | Project | Get a specific version snapshot |
| 27 | POST | `/api/transmittals/{id}/versions/{vid}/restore` | Project | Restore a version |
| 28 | POST | `/api/transmittals/{id}/duplicate` | Project | Copy transmittal to another project |
| 29 | GET | `/api/books` | Admin | List all books (no blob data) |
| 30 | POST | `/api/books/upload` | Admin | Upload a manuscript (multipart) |
| 31 | POST | `/api/books/{id}/convert` | Admin | Start docx→PDF conversion |
| 32 | GET | `/api/books/{id}/download/{format}` | ⚠️ None | Download PDF or EPUB |
| 33 | PUT | `/api/books/{id}/project` | Admin | Link/unlink book to project |
| 34 | DELETE | `/api/books/{id}` | Admin | Delete a book |
| 35 | GET | `/api/projects/{id}/book-spec` | Admin | Get book spec (auto-creates) |
| 36 | PUT | `/api/projects/{id}/book-spec` | Admin | Update book spec JSON |
| 37 | POST | `/api/projects/{id}/book-spec/pull-transmittal` | Admin | Import transmittal→spec fields |
| 38 | POST | `/api/projects/{id}/book-spec/generate-config` | Admin | Preview Typst config from spec |
| 39 | POST | `/api/projects/{id}/book-spec/cover` | Admin | Upload cover image |
| 40 | GET | `/api/projects/{id}/book-spec/cover` | ⚠️ None | Get cover image |
| 41 | GET | `/api/fonts` | Admin | List available Typst fonts |
| 42 | POST | `/api/projects/{id}/book-spec/word-template` | Admin | Generate .docx style template |
| 43 | POST | `/api/projects/{id}/preflight` | Admin | Run manuscript preflight check |
| 44 | GET | `/api/projects/{id}/preflight` | Admin | Get latest preflight result |
| 45 | GET | `/api/projects/{id}/preflight/report` | Admin | Get preflight HTML report |
| 46 | POST | `/api/books/{id}/generate-epub` | Admin | Start EPUB generation |
| 47 | POST | `/api/projects/{id}/transmittal/email` | Project | Email transmittal summary |
| 48 | POST | `/api/projects/{id}/snapshot/email` | Project | Email project snapshot |
| 49 | POST | `/api/projects/{id}/activity/email` | Project | Email activity digest |
| 50 | GET | `/api/email/status` | None | Check if email is configured |
| 51 | GET | `/api/projects/{id}/file-log` | Project | List file log entries |
| 52 | POST | `/api/projects/{id}/file-log` | Project | Create file log entry |
| 53 | DELETE | `/api/projects/{id}/file-log/{entry}` | Project | Delete file log entry |
| 54 | GET | `/api/projects/{id}/journal` | Project | List journal entries |
| 55 | POST | `/api/projects/{id}/journal` | Project | Create journal entry |
| 56 | DELETE | `/api/projects/{id}/journal/{entry}` | Project | Delete journal entry |
| 57 | GET | `/api/projects/{id}/corrections` | Project | List corrections |
| 58 | POST | `/api/projects/{id}/corrections` | Project | Create a correction |
| 59 | DELETE | `/api/projects/{id}/corrections/{entry}` | Project | Delete a correction |
| 60 | PUT | `/api/corrections/{id}/status` | Project | Update correction status |
| 61 | POST | `/api/projects/{id}/corrections/export` | Project | Export corrections as YAML |
| 62 | GET | `/api/clients/{client}` | None | Get client info & auth status |
| 63 | POST | `/api/clients/{client}/verify` | None | Verify client password |
| 64 | GET | `/api/clients/{client}/projects` | Client | List client's projects |
| 65 | POST | `/api/clients/{client}/projects` | Client | Create project from client portal |
| 66 | GET | `/api/clients/{client}/file-log` | Client | Recent file log across projects |
| 67 | GET | `/api/clients/{client}/journal` | Client | Recent journal across projects |
| 68 | POST | `/api/clients/{client}/digest/email` | Client | Email client activity digest |
| 69 | GET | `/static/*` | None | Static assets (CSS, JS, images) |
| 70 | GET | `/{$}` | None | Landing page HTML |
| 71 | GET | `/{client}/` | None | Client portal SPA |
| 72 | GET | `/{client}/{project}/` | None | Production calendar SPA |
| 73 | GET | `/{client}/{project}/transmittal/` | None | Transmittal form SPA |

**⚠️** = Missing auth (see known issues below)

---

## 1. Health

### `GET /healthz`

Health check — verifies the database is reachable.

**Auth:** None  
**Response:**
```json
// 200 OK
{"status": "ok"}

// 503 Service Unavailable
{"status": "error", "detail": "..."}
```

---

## 2. Projects

### `GET /api/projects`

List all non-archived projects.

**Auth:** ⚠️ None (should require admin — see known issues)  
**Response:** `200`
```json
[
  {
    "id": 1,
    "name": "Art of Gig",
    "client_slug": "vgr",
    "project_slug": "aog",
    "start_date": "2026-02-23",
    "created_at": "2026-03-01T...",
    "updated_at": "2026-03-08T...",
    "archived_at": ""
  }
]
```

### `POST /api/projects`

Create a new project with the standard 31-task workflow.

**Auth:** Admin  
**Request:**
```json
{
  "name": "Book Title",
  "start_date": "2026-04-01",
  "client_slug": "acme",
  "project_slug": "book-one"
}
```
**Required:** `client_slug`, `project_slug`  
**Response:** `201`
```json
{
  "id": 5,
  "name": "Book Title",
  "client_slug": "acme",
  "project_slug": "book-one",
  "start_date": "2026-04-01",
  "created_at": "...",
  "updated_at": "..."
}
```

### `GET /api/projects/{id}`

Get a single project with auth status info.

**Auth:** None (returns auth status for the SPA to decide)  
**Path params:** `id` — project ID  
**Response:** `200`
```json
{
  "project": { /* project object */ },
  "has_auth": true,
  "authenticated": false
}
```

### `PUT /api/projects/{id}`

Update project metadata.

**Auth:** Project  
**Path params:** `id` — project ID  
**Request:**
```json
{
  "name": "New Name",
  "start_date": "2026-05-01",
  "client_slug": "acme",
  "project_slug": "new-slug"
}
```
Omitted `client_slug`/`project_slug` are preserved from existing values.  
**Response:** `200` `{"ok": "true"}`

### `DELETE /api/projects/{id}`

**Disabled.** Always returns `405 Method Not Allowed`.  
Use `POST /api/projects/{id}/archive` instead.

### `POST /api/projects/{id}/archive`

Soft-archive a project (sets `archived_at` timestamp).

**Auth:** Admin  
**Response:** `200` `{"ok": "true"}`

### `POST /api/projects/{id}/restore`

Restore an archived project.

**Auth:** Admin  
**Response:** `200` `{"ok": "true"}`

### `POST /api/projects/{id}/seed`

Bulk-import tasks from JSON. Updates project start date if provided.

**Auth:** Project  
**Request:**
```json
{
  "project_name": "Book",
  "project_start": "2026-04-01",
  "tasks": [
    {
      "sort_order": 1,
      "assignee": "JD",
      "task": "Review manuscript",
      "is_milestone": false,
      "orig_weeks": 1,
      "curr_weeks": 1,
      "orig_due": "2026-04-08",
      "curr_due": "2026-04-08",
      "actual_done": "",
      "words": 0,
      "words_per_hour": 0,
      "hours": 0,
      "rate": 0,
      "budget_notes": "",
      "orig_budget": 0,
      "curr_budget": 0,
      "actual_budget": 0
    }
  ]
}
```
**Response:** `200` `{"ok": true, "count": 31}`

### `POST /api/projects/{id}/duplicate`

Clone a project: copies all tasks with date-shifted schedules and zeroed budgets.

**Auth:** Project (on source)  
**Request:**
```json
{
  "name": "Book Two",
  "start_date": "2026-06-01",
  "client_slug": "acme",
  "project_slug": "book-two"
}
```
**Required:** `name`, `client_slug`, `project_slug`  
**Response:** `200`
```json
{
  "project": { /* new project object */ },
  "tasks_copied": 31
}
```
**Side effects:** Dates are shifted by the delta between old and new `start_date`. `actual_done` is cleared and all statuses reset to `"pending"`.

### `GET /api/project-by-path/{client}/{project}`

Look up a project by its client + project slug pair. Used by the SPA to resolve URL paths.

**Auth:** None (returns auth status)  
**Path params:** `client`, `project` — slug strings  
**Response:** `200`
```json
{
  "project": { /* project object */ },
  "has_auth": true,
  "authenticated": false
}
```
**Errors:** `404` if no matching project.

---

## 3. Tasks

### `GET /api/projects/{id}/tasks`

List all tasks for a project, ordered by `sort_order`.

**Auth:** Project  
**Response:** `200`
```json
[
  {
    "id": 1,
    "project_id": 1,
    "sort_order": 1,
    "assignee": "JD",
    "title": "Review manuscript",
    "is_milestone": 0,
    "orig_weeks": 1.0,
    "curr_weeks": 1.0,
    "orig_due": "2026-04-08",
    "curr_due": "2026-04-08",
    "actual_done": "",
    "status": "pending",
    "words": 0,
    "words_per_hour": 0,
    "hours": 0.0,
    "rate": 0.0,
    "budget_notes": "",
    "orig_budget": 0.0,
    "curr_budget": 0.0,
    "actual_budget": 0.0,
    "created_at": "...",
    "updated_at": "..."
  }
]
```

### `POST /api/projects/{id}/tasks`

Create a single task.

**Auth:** Project  
**Request:** Same fields as task object (all optional except implicit `project_id` from path). `status` defaults to `"pending"`.  
**Response:** `201` — the created task object.

### `PUT /api/tasks/{id}`

Update a task. Looks up the task's `project_id` for auth.

**Auth:** Project (of the task's project)  
**Path params:** `id` — task ID  
**Request:** Same fields as task object.  
**Response:** `200` `{"ok": "true"}`  
**Errors:** `404` if task not found.

### `DELETE /api/tasks/{id}`

Delete a task.

**Auth:** Project (of the task's project)  
**Path params:** `id` — task ID  
**Response:** `200` `{"ok": "true"}`  
**Errors:** `404` if task not found.

---

## 4. Auth

### `POST /api/projects/{id}/auth`

Set a password on a project. If no auth exists yet, anyone can set the first password. If auth already exists, requires existing auth.

**Auth:** Project (if auth already configured) or None (first-time setup)  
**Request:**
```json
{"password": "secret123"}
```
**Response:** `200` `{"ok": "true"}`  
**Side effects:** Sets `prodcal_auth_{id}` cookie (1-year expiry, HttpOnly, SameSite=Lax).

### `DELETE /api/projects/{id}/auth`

Clear all auth tokens for a project.

**Auth:** Admin  
**Response:** `200` `{"ok": "true"}`  
**Side effects:** Clears the auth cookie.

### `POST /api/projects/{id}/verify`

Verify a password against stored tokens.

**Auth:** None  
**Request:**
```json
{"password": "secret123"}
```
**Response:** `200` `{"ok": "true"}`  
**Side effects:** Sets `prodcal_auth_{id}` cookie on success.  
**Errors:** `401` if password doesn't match.

---

## 5. Transmittals

Manuscript transmittal forms — structured JSON documents tracking book components, production dates, checklist items, and editorial instructions.

### `GET /api/projects/{id}/transmittal`

Get the transmittal for a project. Auto-creates with default template if none exists.

**Auth:** Project  
**Response:** `200`
```json
{
  "id": 1,
  "project_id": 1,
  "status": "draft",
  "data": {
    "book": {"author": "", "title": "", ...},
    "production": {...},
    "checklist": [...],
    "backmatter": [...],
    "illustrations": {...},
    "permissions": {...},
    "page_iv": {...},
    "subrights": {...},
    "editing": {...},
    "design": {...},
    "cover": {...},
    "files": {...},
    "proofs": {...},
    "custom_styles": [],
    "other_instructions": ""
  },
  "created_at": "...",
  "updated_at": "..."
}
```

### `PUT /api/projects/{id}/transmittal`

Update the transmittal data and/or status.

**Auth:** Project  
**Request:**
```json
{
  "status": "final",
  "data": { /* full transmittal JSON */ }
}
```
`status` defaults to `"draft"`. `data` defaults to the template if empty.  
**Response:** `200` `{"ok": true}`  
**Side effects:**
- Snapshots current state as a version (throttled to 1 per 5 minutes).
- If the updater is not an admin (no `X-ExeDev-UserID`), sends a notification email to the admin (throttled to 1 per 30 minutes).

### `GET /api/transmittals/{id}/versions`

List version history for a transmittal (up to 100 most recent).

**Auth:** Project  
**Path params:** `id` — project ID (despite the URL saying "transmittals")  
**Response:** `200`
```json
[
  {
    "id": 5,
    "status": "draft",
    "title": "Book Title",
    "saved_at": "2026-03-08 14:30:00"
  }
]
```

### `GET /api/transmittals/{id}/versions/{vid}`

Get the full data of a specific version.

**Auth:** Project  
**Path params:** `id` — project ID, `vid` — version ID  
**Response:** `200`
```json
{
  "id": 5,
  "status": "draft",
  "data": { /* full transmittal JSON */ },
  "saved_at": "..."
}
```

### `POST /api/transmittals/{id}/versions/{vid}/restore`

Restore a previous version. Always snapshots the current state first (bypasses throttle).

**Auth:** Project  
**Path params:** `id` — project ID, `vid` — version ID  
**Response:** `200` `{"ok": true}`

### `POST /api/transmittals/{id}/duplicate`

Copy a transmittal to another project. Clears book-specific fields (title, ISBNs, dates) while preserving publisher/house defaults.

**Auth:** Project (on both source and target)  
**Path params:** `id` — source project ID  
**Request:**
```json
{"target_project_id": 5}
```
**Response:** `200` `{"ok": true, "target_project_id": 5}`  
**Errors:** `409` if target already has a transmittal.

---

## 6. Books

Manuscript upload, PDF conversion (docx→Typst→PDF), and EPUB generation.

### `GET /api/books`

List all books without blob data.

**Auth:** Admin  
**Response:** `200`
```json
[
  {
    "id": 1,
    "title": "The Book",
    "author": "Author Name",
    "series": "",
    "source_filename": "manuscript.docx",
    "status": "ready",
    "project_id": 1,
    "created_at": "...",
    "updated_at": "..."
  }
]
```

### `POST /api/books/upload`

Upload a manuscript file (typically .docx).

**Auth:** Admin  
**Content-Type:** `multipart/form-data`  
**Form fields:**
- `file` (required) — the manuscript file, max 50 MB
- `title` (required) — book title
- `author` (required) — author name
- `series` (optional) — series name
- `project_id` (optional) — link to a project

**Response:** `201`
```json
{
  "id": 4,
  "title": "The Book",
  "author": "Author Name",
  "status": "uploaded"
}
```

### `POST /api/books/{id}/convert`

Start the docx→Typst→PDF conversion pipeline.

**Auth:** Admin  
**Response:** `200` `{"status": "converting"}`  
**Side effects:** Runs conversion asynchronously. Pipeline: pandoc (docx→typst with custom Lua filter) → typst compile → store PDF. Uses the linked project's book spec for typography settings.

### `GET /api/books/{id}/download/{format}`

Download the generated PDF or EPUB.

**Auth:** ⚠️ None (see known issues)  
**Path params:** `id` — book ID, `format` — `pdf` or `epub`  
**Response:** Binary file with appropriate Content-Type and Content-Disposition headers.  
**Errors:** `404` if format not generated yet.

### `PUT /api/books/{id}/project`

Link or unlink a book to a project.

**Auth:** Admin  
**Request:**
```json
{"project_id": 1}
```
Send `{"project_id": null}` to unlink.  
**Response:** `200` `{"ok": "true"}`

### `DELETE /api/books/{id}`

Delete a book and all its data.

**Auth:** Admin  
**Response:** `200` `{"ok": "true"}`

### `POST /api/books/{id}/generate-epub`

Start EPUB generation from the book's source docx.

**Auth:** Admin  
**Response:** `200` `{"status": "generating_epub"}`  
**Side effects:** Runs asynchronously. Uses pandoc with epub3 output, project spec settings for metadata/CSS/cover image.

---

## 7. Book Specs

Typography and design specifications for book production. Controls the Typst template configuration.

### `GET /api/projects/{id}/book-spec`

Get the book spec for a project. Auto-creates with defaults if none exists.

**Auth:** Admin  
**Response:** `200`
```json
{
  "id": 1,
  "project_id": 1,
  "data": {
    "metadata": {"title": "", "author": "", ...},
    "page": {"trim": "us-digest", "width_in": 5.5, "height_in": 8.5, ...},
    "typography": {"body_font": "Libertinus Serif", "base_size_pt": 10, ...},
    "headings": {...},
    "elements": {...},
    "running_heads": {...},
    "front_matter": {...},
    "back_matter": {...},
    "typesetting": {...},
    "custom_styles": [],
    "epub": {...}
  },
  "has_cover": false,
  "created_at": "...",
  "updated_at": "..."
}
```

### `PUT /api/projects/{id}/book-spec`

Update the book spec JSON.

**Auth:** Admin  
**Request:**
```json
{"data": { /* full spec JSON */ }}
```
**Response:** `200` `{"ok": true, "updated_at": "..."}`

### `POST /api/projects/{id}/book-spec/pull-transmittal`

Import fields from the project's transmittal into the book spec. Maps: book info→metadata, design→page/typesetting, checklist→front_matter/back_matter, custom_styles.

**Auth:** Admin  
**Response:** `200`
```json
{"ok": true, "data": { /* updated spec */ }}
```
**Errors:** `404` if no transmittal exists.

### `POST /api/projects/{id}/book-spec/generate-config`

Preview the Typst configuration code that would be generated from the current spec.

**Auth:** Admin  
**Response:** `200`
```json
{"config": "\n// Project-specific config overrides...\n#let config = merge-config((\n  page-width: 5.5in,\n  ...\n))"}
```

### `POST /api/projects/{id}/book-spec/cover`

Upload a cover image.

**Auth:** Admin  
**Content-Type:** `multipart/form-data`  
**Form fields:**
- `cover` (required) — JPEG or PNG image, max 10 MB

**Response:** `200`
```json
{"ok": true, "size": 245000, "type": "image/jpeg"}
```

### `GET /api/projects/{id}/book-spec/cover`

Serve the cover image.

**Auth:** ⚠️ None (see known issues)  
**Response:** Binary image with Content-Type header. Cached for 1 hour.  
**Errors:** `404` if no cover uploaded.

### `GET /api/fonts`

List available font families from Typst + the project font directory.

**Auth:** Admin  
**Response:** `200`
```json
[
  {"family": "Libertinus Serif", "category": "serif"},
  {"family": "Source Sans 3", "category": "sans-serif"},
  {"family": "JetBrains Mono", "category": "monospace"}
]
```

### `POST /api/projects/{id}/book-spec/word-template`

Generate a styled .docx template from the book spec. Includes custom styles as named Word styles.

**Auth:** Admin  
**Response:** Binary `.docx` file with Content-Disposition header.  
**Errors:** `400` if duplicate custom style names are detected.

---

## 8. Manuscript Preflight

Automated analysis of .docx manuscripts for edge cases, style issues, and image inventory.

### `POST /api/projects/{id}/preflight`

Run a preflight check on a book's source manuscript.

**Auth:** Admin  
**Request:**
```json
{"book_id": 4}
```
**Response:** `200`
```json
{
  "exists": true,
  "project_id": 1,
  "book_id": 4,
  "status": "ready",
  "source_filename": "manuscript.docx",
  "updated_at": "2026-04-14T...",
  "summary": {
    "total": 12,
    "high": 2,
    "medium": 5,
    "low": 5,
    "by_type": {"undeclared_custom_style": 2, "image_inventory": 3, ...}
  },
  "images": [{"type": "image_inventory", ...}],
  "report_url": "/api/projects/1/preflight/report?book_id=4&preflight_id=1",
  "history": [
    {"id": 1, "status": "ready", "updated_at": "...", "report_url": "...", "latest": true}
  ]
}
```
**Side effects:** Writes source docx to temp dir, runs Python detector script, cross-references custom styles against book spec, stores results in DB.

### `GET /api/projects/{id}/preflight`

Get the latest preflight result for a book.

**Auth:** Admin  
**Query params:** `book_id` (required)  
**Response:** Same shape as POST response, or `{"exists": false}` if no preflight has been run.

### `GET /api/projects/{id}/preflight/report`

Get the full HTML preflight report.

**Auth:** Admin  
**Query params:** `book_id` (required), `preflight_id` (optional — defaults to latest)  
**Response:** `text/html` — the rendered report.  
**Errors:** `404` if no report found.

---

## 9. File Log

Track file transfers between production participants.

### `GET /api/projects/{id}/file-log`

List all file log entries, newest first.

**Auth:** Project  
**Response:** `200`
```json
[
  {
    "id": 1,
    "project_id": 1,
    "direction": "inbound",
    "filename": "manuscript-v2.docx",
    "file_type": "docx",
    "sent_by": "Author",
    "received_by": "JD",
    "notes": "Final revision",
    "transfer_date": "2026-04-10",
    "created_at": "2026-04-10T..."
  }
]
```

### `POST /api/projects/{id}/file-log`

Create a file log entry.

**Auth:** Project  
**Request:**
```json
{
  "direction": "inbound",
  "filename": "manuscript-v2.docx",
  "file_type": "docx",
  "sent_by": "Author",
  "received_by": "JD",
  "notes": "Final revision",
  "transfer_date": "2026-04-10"
}
```
`direction` defaults to `"inbound"`, `transfer_date` defaults to today.  
**Response:** `201` — the created entry.

### `DELETE /api/projects/{id}/file-log/{entry}`

Delete a file log entry.

**Auth:** Project  
**Path params:** `entry` — entry ID  
**Response:** `200` `{"ok": "true"}`

---

## 10. Journal

Free-form project notes, decisions, and call records.

### `GET /api/projects/{id}/journal`

List all journal entries, newest first.

**Auth:** Project  
**Response:** `200`
```json
[
  {
    "id": 1,
    "project_id": 1,
    "entry_type": "note",
    "content": "Discussed cover design options",
    "created_at": "2026-04-10T..."
  }
]
```

### `POST /api/projects/{id}/journal`

Create a journal entry.

**Auth:** Project  
**Request:**
```json
{
  "entry_type": "note",
  "content": "Discussed cover design options"
}
```
`entry_type` defaults to `"note"`. Other types: `"call"`, `"decision"`, `"approval"`.  
`content` is required.  
**Response:** `201` — the created entry.

### `DELETE /api/projects/{id}/journal/{entry}`

Delete a journal entry.

**Auth:** Project  
**Path params:** `entry` — entry ID  
**Response:** `200` `{"ok": "true"}`

---

## 11. Corrections

Find-and-replace corrections tracked for typesetting rounds.

### `GET /api/projects/{id}/corrections`

List all corrections, newest first.

**Auth:** Project  
**Response:** `200`
```json
[
  {
    "id": 1,
    "project_id": 1,
    "find_text": "teh",
    "replace_text": "the",
    "chapter": "Ch. 3",
    "note": "Recurring typo",
    "status": "pending",
    "applied_at": "",
    "created_at": "..."
  }
]
```

### `POST /api/projects/{id}/corrections`

Create a correction.

**Auth:** Project  
**Request:**
```json
{
  "find_text": "teh",
  "replace_text": "the",
  "chapter": "Ch. 3",
  "note": "Recurring typo"
}
```
`find_text` is required.  
**Response:** `201` — the created correction.

### `DELETE /api/projects/{id}/corrections/{entry}`

Delete a correction.

**Auth:** Project  
**Path params:** `entry` — correction ID  
**Response:** `200` `{"ok": "true"}`

### `PUT /api/corrections/{id}/status`

Update the status of a correction.

**Auth:** Project (of the correction's project)  
**Path params:** `id` — correction ID  
**Request:**
```json
{"status": "applied"}
```
Valid values: `"pending"`, `"applied"`, `"skipped"`.  
**Response:** `200` `{"ok": "true"}`  
**Side effects:** Setting `"applied"` records the current date in `applied_at`.

### `POST /api/projects/{id}/corrections/export`

Export all corrections as a YAML file.

**Auth:** Project  
**Response:** `application/x-yaml` file download.
```yaml
# Corrections export
corrections:
  - find: "teh"
    replace: "the"
    chapter: "Ch. 3"
    # Recurring typo
```

---

## 12. Email

All email endpoints use AgentMail API. Requires `AGENTMAIL_API_KEY` and `AGENTMAIL_INBOX_ID` environment variables.

### `GET /api/email/status`

Check if email is configured.

**Auth:** None  
**Response:** `200`
```json
{"configured": true}
```

### `POST /api/projects/{id}/transmittal/email`

Send the transmittal as a formatted email summary.

**Auth:** Project  
**Request:**
```json
{"recipients": ["editor@example.com", "author@example.com"]}
```
First recipient is "To", rest are "CC".  
**Response:** `200`
```json
{"ok": true, "sent_to": ["..."], "subject": "Transmittal [FINAL]: Book Title"}
```
**Errors:** `503` if email not configured.

### `POST /api/projects/{id}/snapshot/email`

Send a full project snapshot — tasks, budget summary, transmittal overview, recent activity.

**Auth:** Project  
**Request:**
```json
{"recipients": ["team@example.com"]}
```
**Response:** `200` `{"ok": true, "sent_to": [...], "subject": "..."}`

### `POST /api/projects/{id}/activity/email`

Send an activity digest — file transfers and journal entries within a time window.

**Auth:** Project  
**Query params:** `days` (optional, default `7`, max `90`)  
**Request:**
```json
{"recipients": ["team@example.com"]}
```
**Response:** `200` `{"ok": true, "sent_to": [...], "subject": "Activity Update: ..."}`

### `POST /api/clients/{client}/digest/email`

Send a cross-project activity digest for all of a client's projects.

**Auth:** Client  
**Request:**
```json
{"recipients": ["client@example.com"]}
```
**Response:** `200` `{"ok": true, "sent_to": [...], "subject": "..."}`

---

## 13. Clients

Client portal — scoped access to a publisher's/author's projects.

### `GET /api/clients/{client}`

Get client info and auth status.

**Auth:** None  
**Path params:** `client` — client slug  
**Response:** `200`
```json
{
  "slug": "vgr",
  "name": "Venkatesh Rao",
  "has_auth": true,
  "authenticated": false
}
```

### `POST /api/clients/{client}/verify`

Verify client password and set auth cookie.

**Auth:** None  
**Request:**
```json
{"password": "secret"}
```
**Response:** `200` `{"ok": true}`  
**Side effects:** Sets `prodcal_client_{slug}` cookie (90-day expiry).  
**Errors:** `401` if password wrong. `404` if client not found.

### `GET /api/clients/{client}/projects`

List the client's non-archived projects with task and transmittal stats.

**Auth:** Client (or any project-level auth for a project in this client)  
**Response:** `200`
```json
[
  {
    "id": 1,
    "name": "Art of Gig",
    "client_slug": "vgr",
    "project_slug": "aog",
    "start_date": "2026-02-23",
    "updated_at": "...",
    "task_count": 31,
    "done_count": 5,
    "active_count": 2,
    "has_transmittal": true,
    "transmittal_status": "draft",
    "path": "/vgr/aog/"
  }
]
```

### `POST /api/clients/{client}/projects`

Create a new project from the client portal. Seeds with standard workflow.

**Auth:** Client  
**Request:**
```json
{
  "name": "New Book",
  "start_date": "2026-06-01",
  "project_slug": "new-book"
}
```
`project_slug` auto-derived from `name` if omitted.  
**Response:** `201` — the created project object.  
**Errors:** `409` if slug already exists.

### `GET /api/clients/{client}/file-log`

Recent file log entries across all of a client's projects.

**Auth:** Client  
**Query params:** `limit` (optional, default `20`, max `100`)  
**Response:** `200`
```json
[
  {
    "id": 1,
    "project_id": 1,
    "direction": "inbound",
    "filename": "...",
    "project_name": "Art of Gig",
    ...
  }
]
```

### `GET /api/clients/{client}/journal`

Recent journal entries across all of a client's projects.

**Auth:** Client  
**Query params:** `limit` (optional, default `20`, max `100`)  
**Response:** `200`
```json
[
  {
    "id": 1,
    "project_id": 1,
    "entry_type": "note",
    "content": "...",
    "project_name": "Art of Gig",
    ...
  }
]
```

---

## 14. Admin

Admin-only endpoints for the dashboard at `/admin/`.

### `GET /admin/`

Serve the admin dashboard SPA.

**Auth:** Admin (redirects to exe.dev login if not authenticated)  
**Response:** `text/html`

### `GET /api/admin/projects`

List all projects with aggregated stats (task counts, transmittal status, auth status).

**Auth:** Admin  
**Query params:** `archived=1` — show archived projects instead  
**Response:** `200`
```json
[
  {
    "id": 1,
    "name": "Art of Gig",
    "client_slug": "vgr",
    "project_slug": "aog",
    "start_date": "2026-02-23",
    "created_at": "...",
    "updated_at": "...",
    "archived_at": "",
    "task_count": 31,
    "done_count": 15,
    "active_count": 2,
    "has_auth": true,
    "has_transmittal": true,
    "transmittal_status": "final",
    "path": "/vgr/aog/"
  }
]
```

### `GET /api/admin/clients`

List all clients with project counts.

**Auth:** Admin  
**Response:** `200`
```json
[
  {
    "slug": "vgr",
    "name": "Venkatesh Rao",
    "has_auth": true,
    "project_count": 3,
    "created_at": "..."
  }
]
```

### `POST /api/admin/clients`

Create a new client.

**Auth:** Admin  
**Request:**
```json
{
  "name": "New Publisher",
  "slug": "newpub",
  "password": "optional-password"
}
```
Slug is normalized (lowercase, hyphens only). Password is optional.  
**Response:** `201`
```json
{
  "slug": "newpub",
  "name": "New Publisher",
  "has_auth": true,
  "created_ok": true
}
```
**Errors:** `409` if slug already exists.

---

## 15. Pages (HTML)

These routes serve single-page applications.

| Route | Page | Description |
|-------|------|-------------|
| `GET /{$}` | Landing | Marketing/info landing page |
| `GET /{client}/` | Client portal | Client's project list with auth gate |
| `GET /{client}/{project}/` | Calendar SPA | Production calendar & task management |
| `GET /{client}/{project}/transmittal/` | Transmittal SPA | Manuscript transmittal form |
| `GET /static/*` | Static assets | CSS, JS, images (embedded in binary) |

Slug-based routing: paths without trailing slashes redirect with `301`.  
Sub-paths like `/{client}/{project}/style.css` serve static assets.

---

## Known Issues

| Issue | Severity | Route | Description |
|-------|----------|-------|-------------|
| Missing auth on project list | HIGH | `GET /api/projects` | Returns all projects to unauthenticated requests |
| Missing auth on book download | HIGH | `GET /api/books/{id}/download/{format}` | Anyone with a book ID can download PDFs/EPUBs |
| Missing auth on cover image | LOW | `GET /api/projects/{id}/book-spec/cover` | Cover images are publicly accessible |
| SHA-256 password hashing | HIGH | `checkAuth()`, `handleAdminCreateClient()` | Uses SHA-256 instead of bcrypt |
| API key in git history | HIGH | `SESSION-SUMMARY.txt` | AgentMail API key committed to repo |

---

*Generated 2026-04-14. 73 routes documented (71 registered + 2 catch-all patterns).*
