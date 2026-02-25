# Email Integration for File Log + Journal

Repo: `/home/exedev/prodcal`. See `SESSION-SUMMARY.txt` for full context. Go backend + vanilla JS SPA, SQLite, systemd service on port 8000.

## What exists

**Email infra** (srv/email.go, srv/snapshot_email.go):
- AgentMail API client, configured via `.env` (AGENTMAIL_API_KEY + AGENTMAIL_INBOX_ID)
- `POST /api/projects/{id}/transmittal/email` — sends transmittal summary
- `POST /api/projects/{id}/snapshot/email` — sends project snapshot (schedule overview stat cards, full task table with status badges and overdue highlights, budget summary with variance, transmittal status)
- Recipient picker modal pattern in the SPA: default recipients (jdbb@agentmail.to, j@djinna.com) + custom "Other" field. First recipient = To, rest = CC.
- Both HTML and plain-text bodies built in Go string builders

**File Log** (srv/filelog.go, session 10):
- `file_log` table: direction, filename, file_type, sent_by, received_by, notes, transfer_date
- `GET/POST /api/projects/{id}/file-log`, `DELETE .../file-log/{entry}`
- Client-level: `GET /api/clients/{client}/file-log`

**Journal** (srv/journal.go, session 10):
- `journal` table: entry_type (call/decision/approval/note), content, created_at
- `GET/POST /api/projects/{id}/journal`, `DELETE .../journal/{entry}`
- Client-level: `GET /api/clients/{client}/journal`

UI for both lives in Files and Journal tabs on the calendar page (srv/static/app.js), plus a Recent Activity section on the client portal (srv/static/client.html).

## What to build

Three integration points, in priority order:

### 1. Add File Log + Journal to the project snapshot email

The "📧 Email Snapshot" button (already on the calendar toolbar) sends a comprehensive project status email. It currently covers schedule, budget, and transmittal. Add two new sections:

**Recent Files** section — last 10 file log entries for this project:
- Table: date, direction (↓ In / ↑ Out), filename, type, from→to
- Show count: "3 files logged" or similar
- Skip if no entries

**Recent Journal** section — last 10 journal entries for this project:
- Reverse-chronological list with type badge (📞 Call, ⚖️ Decision, ✅ Approval, 📝 Note), date, and content
- Skip if no entries

Both HTML and plain-text builders need updating. The data loading happens in `handleSendProjectSnapshot` (srv/snapshot_email.go) — add queries there, thread through `snapshotParams`, render in `buildSnapshotHTML` and `buildSnapshotText`.

### 2. Activity digest email — new endpoint + button

A focused email that's just the recent activity, not the full snapshot. Useful for a quick "what happened this week" update.

`POST /api/projects/{id}/activity/email` — sends file log + journal entries from the last N days (default 7, accept `?days=N`).

Add an "📧 Email" button to both the Files tab and the Journal tab — small icon button in the tab header bar. Clicking it opens the same recipient-picker modal pattern, then hits the activity email endpoint.

The email body:
- Subject: "Activity Update: {Project Name} — {date range}"
- Header with project name and date range
- File transfers section (if any)
- Journal entries section (if any)
- "No activity in this period" if both are empty
- Footer with link to project

### 3. Client-level weekly digest (stretch goal)

A digest email covering ALL projects for a client. Not per-project — it aggregates.

`POST /api/clients/{client}/digest/email` — sends a digest for the last 7 days across all client projects.

Sections per project (only show projects with activity):
- Project name as header
- File transfers for that project
- Journal entries for that project

Plus a summary header: "3 files transferred, 5 journal entries across 2 projects this week"

Add a "📧 Weekly Digest" button to the client portal dashboard.

This one's a stretch goal — do #1 and #2 first.

## Implementation notes

- The snapshot email HTML builder (srv/snapshot_email.go) uses inline styles for email client compatibility — follow the same pattern
- Color scheme: purple (#6c63ff) headers, alternating row backgrounds, badge colors matching the SPA
- The SPA modal pattern for recipients is in app.js — look at `renderSnapshotEmailModal` and `snapshotRecipients` for the pattern to copy
- The `sendEmail(to, cc, subject, textBody, htmlBody)` function on `EmailConfig` handles the AgentMail API call
- Auth: same cascade as everything else — project token / client cookie / exe.dev admin
- After building, `make build && sudo systemctl restart srv`
