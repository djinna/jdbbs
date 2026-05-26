# Testing Plan for ProdCal

**STATUS: ✅ COMPLETE** - See commit 4f217cd

Repo: `/home/exedev/prodcal`. Go backend + vanilla JS SPA, SQLite (actual DB at `/home/exedev/db.sqlite3`), systemd service on port 8000.

## Context

The app has grown through ~12 sessions with no automated tests. It now has:
- Project CRUD with tasks, Gantt timeline, budget tracking
- Transmittal forms with version history
- File log (transfer tracking) and journal (call/decision/approval/note entries)
- Client portal with per-client auth + per-project auth
- Admin dashboard (exe.dev admin only)
- Email system (AgentMail API) with 4 email types:
  1. Transmittal email (`POST /api/projects/{id}/transmittal/email`)
  2. Project snapshot email (`POST /api/projects/{id}/snapshot/email`) — schedule + budget + transmittal + recent files + recent journal
  3. Activity digest email (`POST /api/projects/{id}/activity/email`) — file log + journal from last N days
  4. Client weekly digest (`POST /api/clients/{client}/digest/email`) — activity across all client projects

## What to test

### Priority 1: Go backend API tests

Create `srv/server_test.go` (or split into per-feature test files). Use `httptest.NewServer` with an in-memory SQLite database.

**Setup helper** needed:
- Create test server with in-memory DB (`db.Open(":memory:")`)
- Run migrations
- Seed test data (project, tasks, transmittal, file log entries, journal entries, client, auth tokens)
- Return `*httptest.Server` + cleanup func

**Test categories:**

#### Auth tests (`srv/auth_test.go`)
- Request without auth → 401
- Request with valid project cookie → 200
- Request with valid client cookie → 200
- Request with `X-ExeDev-UserID` header → 200
- Client portal auth cascade: client cookie, project cookie, admin header

#### Project CRUD (`srv/project_test.go`)
- Create project → 201, verify fields
- Get project → 200 with auth, 401 without
- Update project → 200, verify changed fields
- Delete project → cascades tasks/transmittal/file-log/journal
- Duplicate project → copies tasks
- List projects → returns all (admin)

#### Task CRUD (`srv/task_test.go`)
- Create task with all budget fields
- Update task status transitions
- Delete task
- List tasks ordered by sort_order
- Milestone flag behavior

#### Transmittal (`srv/transmittal_test.go`)
- Create/update transmittal data (JSON blob)
- Version history: update creates version, list versions, restore version
- Duplicate transmittal

#### File Log (`srv/filelog_test.go`)
- Create file log entry (inbound/outbound)
- List entries ordered by transfer_date DESC
- Delete entry
- Client-level file log (aggregates across projects)
- Direction defaults to "inbound"
- transfer_date defaults to today

#### Journal (`srv/journal_test.go`)
- Create journal entry (call/decision/approval/note types)
- Content required validation
- List entries ordered by created_at DESC
- Delete entry
- Client-level journal (aggregates across projects)
- entry_type defaults to "note"

#### Email endpoints (`srv/email_test.go`)
- All 4 email endpoints return 503 when email not configured
- Recipient validation (must contain @ and .)
- At least one recipient required
- With email configured (mock AgentMail): verify email body contains expected sections
- Snapshot email includes file log + journal sections when data exists
- Snapshot email omits file log + journal sections when empty
- Activity email with ?days=N parameter
- Activity email with no activity → "no activity" message
- Client digest aggregates only projects with activity
- Client digest with no activity across any projects

**For email tests**: Don't call the real AgentMail API. Either:
- Make `EmailConfig.sendEmail` mockable (interface or function field)
- Or just test the HTML/text builders directly (they're exported functions)

### Priority 2: Email content tests

Test the builder functions directly:

```go
func TestBuildSnapshotHTML_WithFileLog(t *testing.T) {
    p := snapshotParams{
        ProjectName: "Test Project",
        FileLog: []fileLogEntry{{Direction: "inbound", Filename: "test.pdf", ...}},
        Journal: []journalEntry{{EntryType: "call", Content: "Discussed timeline", ...}},
    }
    html := buildSnapshotHTML(p)
    // assert contains "Recent Files"
    // assert contains "test.pdf"
    // assert contains "Recent Journal"
    // assert contains "Discussed timeline"
}

func TestBuildSnapshotHTML_EmptyFileLog(t *testing.T) {
    p := snapshotParams{ProjectName: "Test Project"}
    html := buildSnapshotHTML(p)
    // assert NOT contains "Recent Files"
    // assert NOT contains "Recent Journal"
}
```

Same pattern for:
- `buildActivityHTML` / `buildActivityText`
- `buildClientDigestHTML` / `buildClientDigestText`
- `buildSnapshotText` with file log + journal

### Priority 3: Integration/smoke tests

End-to-end flows:
- Create project → add tasks → add file log → add journal → send snapshot email → verify response
- Create project → add activity → send activity email → verify date filtering
- Create client → create 2 projects → add activity to one → send digest → verify only active project included

### Priority 4: Frontend smoke tests (optional, stretch)

If desired, could add basic Playwright or similar tests:
- Auth flow: enter password → see calendar
- Tab switching: Files → Journal
- Email modal opens and has recipient checkboxes
- Client portal loads project cards with email buttons

## Implementation notes

- Go test files go in `srv/` package alongside the code
- Use `testing` + `net/http/httptest` + `strings.Contains` for assertions
- The DB path for tests: use `":memory:"` with `db.Open`
- Migrations embed is in `db/` package — call `db.RunMigrations()` on the test DB
- Email builders are package-level functions in `srv/` — directly testable
- The `snapshotParams`, `activityParams`, `clientDigestParams` structs and `fileLogEntry`, `journalEntry` types are all in the `srv` package
- Server constructor: `srv.New(database, emailConfig)` — pass `nil` for email in non-email tests
- Build/deploy after: `make build && sudo systemctl restart srv`

## Files to touch

- `srv/server_test.go` — test setup helper, smoke tests
- `srv/auth_test.go` — auth cascade tests
- `srv/project_test.go` — CRUD tests
- `srv/task_test.go` — task CRUD
- `srv/filelog_test.go` — file log tests
- `srv/journal_test.go` — journal tests  
- `srv/email_test.go` — email endpoint + builder tests
- `srv/transmittal_test.go` — transmittal tests

## Current file structure reference

```
srv/
  server.go          — routes, auth, project/task handlers
  admin.go           — admin dashboard handler
  client.go          — client portal API handlers
  email.go           — AgentMail client, transmittal email
  snapshot_email.go  — project snapshot email (schedule+budget+tx+files+journal)
  activity_email.go  — activity digest email (files+journal, last N days)
  client_digest_email.go — client weekly digest email
  filelog.go         — file log CRUD handlers
  journal.go         — journal CRUD handlers
  static/
    app.js           — calendar SPA (project view)
    admin.html       — admin dashboard
    client.html      — client portal
    style.css        — shared styles
db/
  db.go              — SQLite open + migration runner
  migrations/        — 001 through 006
```
