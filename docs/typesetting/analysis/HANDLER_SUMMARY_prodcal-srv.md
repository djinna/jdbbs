# prodcal/srv Handler Summary

Generated from: client.go, client_digest_email.go, admin.go, email.go, activity_email.go, snapshot_email.go, preflight.go

---

## Auth Patterns (shared infrastructure)

| Pattern | Description |
|---------|-------------|
| **Project-level auth** (`requireAuth`) | Checks cookie `prodcal_auth_{id}` or `X-Auth-Token` header against `auth_tokens` table. If no tokens configured → open access. Also accepts client-level cookie or `X-ExeDev-UserID` header. |
| **Client-level auth** (`checkClientAuth`) | Checks cookie `prodcal_client_{slug}` against `clients.password_hash`. If no password set → open access. |
| **Client-or-project auth** (`checkClientAuthOrProjectAuth`) | Accepts client cookie, any sibling project cookie, or exe.dev admin header. Returns 404 if client not found, 401 if auth fails. |
| **exe.dev admin** (`requireExeDevAdminAPI`) | Requires `X-ExeDev-UserID` header (set by upstream proxy after exe.dev SSO). Returns 401 JSON if missing. |
| **exe.dev admin page** (`requireExeDevAdmin`) | Same check but redirects to `/__exe.dev/login?redirect=...` (302) for HTML page requests. |

---

## client.go — Client Portal Handlers

### `handleClientVerify`
- **Route:** `POST /api/clients/{client}/verify`
- **Purpose:** Verify a client password and set a persistent auth cookie.
- **Path params:** `{client}` — client slug
- **Request body:** `{"password": "string"}`
- **Auth:** None required (this IS the auth endpoint)
- **Response:**
  - `200` → `{"ok": true}` (also sets cookie `prodcal_client_{slug}`, 90-day expiry, HttpOnly, SameSite=Lax)
  - `400` → `{"error": "client slug required"}` or `{"error": "bad request"}`
  - `404` → `{"error": "client not found"}`
  - `401` → `{"error": "invalid password"}`
  - `500` → `{"error": "server error"}`
- **Notes:** If client has no password set (`password_hash == ""`), returns `{"ok": true}` immediately without setting a cookie.

### `handleClientInfo`
- **Route:** `GET /api/clients/{client}`
- **Purpose:** Return client metadata and auth status (whether the requester is authenticated).
- **Path params:** `{client}` — client slug
- **Auth:** None required (informational endpoint)
- **Response:**
  - `200` → `{"slug": "...", "name": "...", "has_auth": bool, "authenticated": bool}`
  - `404` → `{"error": "client not found"}`
  - `500` → `{"error": "server error"}`
- **Notes:** `authenticated` is true if no password is set OR if the client cookie is valid.

### `handleClientProjects`
- **Route:** `GET /api/clients/{client}/projects`
- **Purpose:** List all non-archived projects for a client with summary stats.
- **Path params:** `{client}` — client slug
- **Auth:** Client-level auth OR any project-level auth for a sibling project in the same client.
- **Response:**
  - `200` → Array of:
    ```json
    [{
      "id": 1, "name": "...", "client_slug": "...", "project_slug": "...",
      "start_date": "...", "updated_at": "...",
      "task_count": 10, "done_count": 3, "active_count": 2,
      "has_transmittal": true, "transmittal_status": "draft",
      "path": "/client-slug/project-slug/"
    }]
    ```
  - `404` → `{"error": "client not found"}`
  - `401` → `{"error": "unauthorized"}`
  - `500` → `{"error": "..."}`
- **Notes:** Returns empty array `[]` if no projects. Joins with `tasks` and `transmittals` for aggregate counts.

### `handleClientCreateProject`
- **Route:** `POST /api/clients/{client}/projects`
- **Purpose:** Create a new project under a client, with standard workflow seeding.
- **Path params:** `{client}` — client slug
- **Request body:**
  ```json
  {
    "name": "string (required)",
    "start_date": "string (optional, YYYY-MM-DD)",
    "project_slug": "string (optional, auto-derived from name if empty)"
  }
  ```
- **Auth:** Client-level auth (if client is password-protected).
- **Response:**
  - `201` → Full project object from `dbgen.CreateProject`
  - `400` → `{"error": "name required"}` / `{"error": "project slug required"}` / etc.
  - `404` → `{"error": "client not found"}`
  - `401` → `{"error": "client login required"}`
  - `409` → `{"error": "project slug already exists for this client"}`
  - `500` → `{"error": "..."}`
- **Notes:** Uses transaction. Normalizes slug via `normalizeProjectSlug` (lowercased, alphanumeric + hyphens). Seeds with `seedProjectWithStandardWorkflow`.

---

## client_digest_email.go — Client Weekly Digest Email

### `handleSendClientDigest`
- **Route:** `POST /api/clients/{client}/digest/email`
- **Purpose:** Send a weekly digest email summarizing file transfers and journal entries across all projects for a client.
- **Path params:** `{client}` — client slug
- **Request body:**
  ```json
  {
    "recipients": ["email@example.com", "cc@example.com"]
  }
  ```
- **Auth:** `checkClientAuthOrProjectAuth` — client cookie, any sibling project cookie, or exe.dev admin.
- **Response:**
  - `200` → `{"ok": true, "sent_to": [...], "subject": "Weekly Digest: ClientName — Jan 1 – Jan 8, 2025"}`
  - `503` → `{"error": "email not configured ..."}` (if `AGENTMAIL_API_KEY` / `AGENTMAIL_INBOX_ID` not set)
  - `400` → `{"error": "at least one recipient required"}` or invalid email
  - `404` → (from `checkClientAuthOrProjectAuth` if client not found)
  - `401` → (from `checkClientAuthOrProjectAuth` if unauthorized)
  - `500` → `{"error": "failed to send email: ..."}` or query errors
- **Notes:** Always looks at last 7 days. First recipient = To, rest = CC. Sends both HTML and plaintext. Only includes projects that had activity. Builds stats: active projects count, file transfer count, journal entry count.

---

## admin.go — Admin Dashboard Handlers

### `handleAdminDashboard`
- **Route:** `GET /admin/`
- **Purpose:** Serve the admin SPA HTML page.
- **Auth:** exe.dev admin (page-level — redirects to SSO login if not authenticated).
- **Response:**
  - `200` → HTML content (`static/admin.html`)
  - `302` → Redirect to `/__exe.dev/login?redirect=...` if not authed
  - `500` → plain text "internal error"

### `handleAdminProjectList`
- **Route:** `GET /api/admin/projects`
- **Purpose:** List all projects with summary stats (for admin dashboard).
- **Query params:** `archived=1` — if set, returns only archived projects (default: non-archived)
- **Auth:** exe.dev admin API (requires `X-ExeDev-UserID` header).
- **Response:**
  - `200` → Array of:
    ```json
    [{
      "id": 1, "name": "...", "client_slug": "...", "project_slug": "...",
      "start_date": "...", "created_at": "...", "updated_at": "...", "archived_at": "",
      "task_count": 10, "done_count": 3, "active_count": 2,
      "has_auth": true, "has_transmittal": true, "transmittal_status": "final",
      "path": "/client/project/"
    }]
    ```
  - `401` → `{"error": "exe.dev login required"}`
  - `500` → `{"error": "..."}`
- **Notes:** Non-archived sorted by `updated_at DESC`; archived sorted by `archived_at DESC`. Returns empty array if none.

### `handleAdminClientList`
- **Route:** `GET /api/admin/clients`
- **Purpose:** List all clients with project counts.
- **Auth:** exe.dev admin API.
- **Response:**
  - `200` → Array of:
    ```json
    [{
      "slug": "...", "name": "...", "has_auth": true,
      "project_count": 5, "created_at": "..."
    }]
    ```
  - `401` → `{"error": "exe.dev login required"}`
  - `500` → `{"error": "..."}`
- **Notes:** Sorted by `created_at DESC, slug ASC`.

### `handleAdminCreateClient`
- **Route:** `POST /api/admin/clients`
- **Purpose:** Create a new client organization.
- **Request body:**
  ```json
  {
    "name": "string (required)",
    "slug": "string (required, normalized)",
    "password": "string (optional, hashed with SHA-256 if provided)"
  }
  ```
- **Auth:** exe.dev admin API.
- **Response:**
  - `201` → `{"slug": "...", "name": "...", "has_auth": true/false, "created_ok": true}`
  - `400` → `{"error": "name required"}` or `{"error": "slug required"}`
  - `409` → `{"error": "client slug already exists"}`
  - `401` → `{"error": "exe.dev login required"}`
  - `500` → `{"error": "..."}`

---

## email.go — Transmittal Email + Email Status

### `handleSendTransmittalEmail`
- **Route:** `POST /api/projects/{id}/transmittal/email`
- **Purpose:** Email a formatted transmittal summary to recipients.
- **Path params:** `{id}` — project ID (int64)
- **Request body:**
  ```json
  {
    "recipients": ["to@example.com", "cc1@example.com"]
  }
  ```
- **Auth:** Project-level auth (`requireAuth`).
- **Response:**
  - `200` → `{"ok": true, "sent_to": [...], "subject": "Transmittal [FINAL]: Book Title"}`
  - `503` → `{"error": "email not configured ..."}`
  - `400` → bad ID, bad request, invalid email, or no recipients
  - `401` → `{"error": "unauthorized"}`
  - `404` → `{"error": "transmittal not found"}`
  - `500` → parse error or send failure
- **Notes:** Loads transmittal from DB (`transmittals` table). Renders both HTML and plaintext versions. HTML includes styled tables for book info, production dates, component checklist, design specs. Subject format: `Transmittal [STATUS]: Book Title`. First recipient = To, rest = CC.

### `handleEmailStatus`
- **Route:** `GET /api/email/status`
- **Purpose:** Check if email sending is configured.
- **Auth:** None.
- **Response:**
  - `200` → `{"configured": true/false}`
- **Notes:** Returns true if both `AGENTMAIL_API_KEY` and `AGENTMAIL_INBOX_ID` env vars are set.

---

## activity_email.go — Activity Report Email

### `handleSendActivityEmail`
- **Route:** `POST /api/projects/{id}/activity/email`
- **Purpose:** Email an activity report (file transfers + journal entries) for a project over a configurable time window.
- **Path params:** `{id}` — project ID
- **Query params:** `days` — lookback period (1–90, default 7)
- **Request body:**
  ```json
  {
    "recipients": ["to@example.com", "cc@example.com"]
  }
  ```
- **Auth:** Project-level auth (`requireAuth`).
- **Response:**
  - `200` → `{"ok": true, "sent_to": [...], "subject": "Activity Update: ProjectName — Jan 1 – Jan 8, 2025"}`
  - `503` → email not configured
  - `400` → bad ID, bad request, invalid email, no recipients
  - `401` → unauthorized
  - `404` → project not found
  - `500` → query/send errors
- **Notes:** Queries `file_log` and `journal` tables for entries since the cutoff date. HTML email includes stat cards (file transfer count, journal entry count), file transfer table (date, direction, filename, type, from→to), and journal entries with emoji icons (📞 call, ⚖️ decision, ✅ approval, 📝 default). Also generates plaintext fallback.

---

## snapshot_email.go — Full Project Snapshot Email

### `handleSendProjectSnapshot`
- **Route:** `POST /api/projects/{id}/snapshot/email`
- **Purpose:** Email a comprehensive project snapshot including schedule, budget, transmittal status, recent files, and journal.
- **Path params:** `{id}` — project ID
- **Request body:**
  ```json
  {
    "recipients": ["to@example.com", "cc@example.com"]
  }
  ```
- **Auth:** Project-level auth (`requireAuth`).
- **Response:**
  - `200` → `{"ok": true, "sent_to": [...], "subject": "Project Snapshot: ProjectName — Jan 8, 2025"}`
  - `503` → email not configured
  - `400` → bad ID, bad request, invalid email, no recipients
  - `401` → unauthorized
  - `404` → project not found
  - `500` → query/send errors
- **Email sections:**
  1. **Schedule Overview** — % complete, done/active/pending counts as stat cards
  2. **Task Schedule** — full task table (task name, assignee, status badge, due date; milestones marked with ◆; overdue rows highlighted red with ⚠)
  3. **Budget Summary** — original/current/actual totals + variance (under/over budget)
  4. **Transmittal Status** (if exists) — book title, author, publisher, production dates, checklist progress (X/Y items)
  5. **Recent Files** — last 10 file log entries (date, direction, filename, type, from→to)
  6. **Recent Journal** — last 10 journal entries with emoji + entry type + date + content
- **Notes:** Most comprehensive email. Loads from `projects`, `tasks`, `transmittals`, `file_log`, `journal` tables. Computes budget variance. Identifies overdue tasks.

---

## preflight.go — Manuscript Preflight Analysis

### `handleGetManuscriptPreflight`
- **Route:** `GET /api/projects/{id}/preflight`
- **Purpose:** Get the latest manuscript preflight result (or check if one exists).
- **Path params:** `{id}` — project ID
- **Query params:** `book_id` (required, int64)
- **Auth:** exe.dev admin API (`requireExeDevAdminAPI`).
- **Response:**
  - `200` (no preflight) → `{"exists": false}`
  - `200` (has preflight) →
    ```json
    {
      "exists": true,
      "project_id": 1,
      "book_id": 42,
      "status": "ready" | "error",
      "source_filename": "manuscript.docx",
      "updated_at": "2025-01-08T...",
      "summary": {
        "total": 15, "high": 2, "medium": 5, "low": 8,
        "by_type": {"undeclared_custom_style": 2, "image_inventory": 5, ...}
      },
      "images": [{...image_inventory findings...}],
      "report_url": "/api/projects/1/preflight/report?book_id=42&preflight_id=99",
      "history": [
        {"id": 99, "status": "ready", "updated_at": "...", "report_url": "...", "latest": true},
        {"id": 98, "status": "ready", "updated_at": "...", "report_url": "...", "latest": false}
      ],
      "error": "..." // only if status=error
    }
    ```
  - `400` → bad id or bad book_id
  - `401` → exe.dev login required
  - `500` → query error

### `handleGetManuscriptPreflightReport`
- **Route:** `GET /api/projects/{id}/preflight/report`
- **Purpose:** Serve the HTML preflight report (for viewing in browser/iframe).
- **Path params:** `{id}` — project ID
- **Query params:**
  - `book_id` (required, int64)
  - `preflight_id` (optional, int64) — if omitted, serves the latest report
- **Auth:** exe.dev admin API.
- **Response:**
  - `200` → `Content-Type: text/html; charset=utf-8` with the stored HTML report
  - `400` → bad id, bad book_id, bad preflight_id
  - `404` → not found
  - `401` → exe.dev login required
  - `500` → query error

### `handleRunManuscriptPreflight`
- **Route:** `POST /api/projects/{id}/preflight`
- **Purpose:** Run the manuscript preflight analysis (edge-case detection) on a book's DOCX source.
- **Path params:** `{id}` — project ID
- **Request body:**
  ```json
  {
    "book_id": 42
  }
  ```
- **Auth:** exe.dev admin API.
- **Response:**
  - `200` → Same shape as `handleGetManuscriptPreflight` (with `exists: true`, full summary, images, history)
  - `400` → bad id, bad request, book_id required, no source file
  - `404` → book not found
  - `401` → exe.dev login required
  - `500` → various errors
- **Process:**
  1. Loads book from DB (requires `source_data` blob to be non-empty)
  2. Loads/creates book spec for declared custom styles
  3. Writes DOCX to temp file
  4. Writes declared styles JSON to temp file
  5. Runs `python3 scripts/detect-edge-cases.py` (or injected test runner)
  6. Appends `undeclared_custom_style` warnings by cross-referencing observed styles against declared styles from the book spec
  7. Appends `declared_custom_style_used` confirmations for declared styles that appear in the manuscript
  8. Builds summary (total/high/medium/low counts, by-type breakdown)
  9. Stores HTML report, JSON findings, summary in `manuscript_preflights` table
  10. Returns full response with history
- **Notes:** On python error, stores status="error" with error message but still creates a DB record. Uses `bookProdRoot` (`/home/exedev/book-production`) for script path. Supports overriding the runner via `s.preflightRunner` (for testing).

---

## Helper Types Referenced

```go
type fileLogEntry struct {
    ID, ProjectID         int64
    Direction, Filename   string  // "inbound"|"outbound"
    FileType, SentBy      string
    ReceivedBy, Notes     string
    TransferDate, CreatedAt string
}

type journalEntry struct {
    ID, ProjectID         int64
    EntryType, Content    string  // EntryType: "call"|"decision"|"approval"|"note"
    CreatedAt             string
}

type transmittalEmailData struct {
    Book        { Author, Title, Subtitle, Publisher, Editor, ISBNPaper, ISBNCloth }
    Production  { TransmittalDate, MechsDelivery, WeeksInProd, BoundBookDate, PrintRun }
    ChecklistStats { Parts, Chapters, WordsChars, MSPP, EstBookPP }
    Checklist   []{ Component, HereNow bool, ToComeWhen }
    Backmatter  []{ Component, HereNow bool, ToComeWhen }
    Editing     { CopyeditingLevel, SpecialCharacters, Instructions }
    Design      { Trim, EstPages, Complexity }
    OtherInstructions string
}
```

## Non-Handler Utility Functions (in these files)

| Function | File | Purpose |
|----------|------|---------|
| `normalizeProjectSlug` | client.go | Normalizes a string to a URL-safe slug (lowercase, alphanumeric + hyphens) |
| `hasAnyProjectAuthForClient` | client.go | Checks if user has project-auth cookie for ANY project in the client |
| `checkClientAuth` | client.go | Verifies client-level cookie against stored password hash |
| `getClientSlugForProject` | client.go | Looks up client_slug for a project ID |
| `LoadEmailConfig` | email.go | Loads AgentMail config from env vars |
| `sendEmail` | email.go | Sends via AgentMail API (POST to api.agentmail.to) |
| `buildTransmittalTextSummary` | email.go | Builds plaintext transmittal email |
| `buildTransmittalHTMLSummary` | email.go | Builds HTML transmittal email |
| `buildClientDigestHTML/Text` | client_digest_email.go | Builds digest email content |
| `buildActivityHTML/Text` | activity_email.go | Builds activity email content |
| `activityJournalEmoji` | activity_email.go | Maps journal entry types to emoji |
| `buildSnapshotHTML/Text` | snapshot_email.go | Builds snapshot email content |
| `snapshotFormatMoney` | snapshot_email.go | Formats float to `$X,XXX.XX` |
| `snapshotFormatDate` | snapshot_email.go | Formats `YYYY-MM-DD` to `Jan 2, 2006` |
| `isOverdue` | snapshot_email.go | Checks if a task's due date is past today |
| `declaredCustomStyles` | preflight.go | Extracts declared styles from book spec JSON |
| `appendUndeclaredStyleWarnings` | preflight.go | Cross-references observed vs declared styles |
| `buildPreflightSummary` | preflight.go | Aggregates findings into severity counts |
| `preflightReportURL` | preflight.go | Constructs report URL with query params |
| `preflightHistoryEntries` | preflight.go | Converts DB rows to history response entries |
| `requireExeDevAdmin` | admin.go | Page-level exe.dev auth (redirects) |
| `requireExeDevAdminAPI` | admin.go | API-level exe.dev auth (returns 401 JSON) |
