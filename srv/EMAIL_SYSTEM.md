# Email System — Architecture & Reference

All email is sent via [AgentMail](https://agentmail.to) API.
Configured via environment variables `AGENTMAIL_API_KEY` and `AGENTMAIL_INBOX_ID`.
Sender address: `jdbb@agentmail.to`.

## Infrastructure

**`srv/email.go`** — Core email plumbing:
- `EmailConfig` struct holds API key + inbox ID, loaded from env at startup
- `sendEmail(to, cc, subject, textBody, htmlBody)` — low-level AgentMail API call
- `handleEmailStatus()` — `GET /api/email/status` — returns `{"configured": true/false}`
- `handleSendTransmittalEmail()` — `POST /api/projects/{id}/transmittal/email`
- `transmittalEmailData` struct + `buildTransmittalTextSummary()` / `buildTransmittalHTMLSummary()`

The `Server.Email` field is `*EmailConfig` — nil if env vars are missing.
All handlers check `s.Email == nil` and return 503 if not configured.

## Email Pathways

There are **6 email pathways** in two categories:

### Manual (button-triggered, user picks recipients)

All of these use the same SPA modal pattern: default recipients
(`jdbb@agentmail.to` + `j@djinna.com`) plus an editable "Other" row.
First recipient = To, rest = CC.

| # | Trigger | Endpoint | File | Description |
|---|---------|----------|------|-------------|
| 1 | **Email button** on transmittal form | `POST /api/projects/{id}/transmittal/email` | `srv/email.go` | Full transmittal summary (book info, production dates, checklist, design, etc.) |
| 2 | **Email Snapshot button** on calendar toolbar | `POST /api/projects/{id}/snapshot/email` | `srv/snapshot_email.go` | Comprehensive project snapshot (schedule, budget, transmittal status, recent files, recent journal) |
| 3 | **Email button** on Files/Journal tabs | `POST /api/projects/{id}/activity/email` | `srv/activity_email.go` | Activity digest — file transfers + journal entries from last N days (default 7, `?days=N`) |
| 4 | **Weekly Digest button** on client portal | `POST /api/clients/{client}/digest/email` | `srv/client_digest_email.go` | Aggregated digest across ALL projects for a client — last 7 days of files + journal |

### Automatic (server-initiated, no user action)

| # | Trigger | Recipient | File | Description |
|---|---------|-----------|------|-------------|
| 5 | **Client updates transmittal** (auto-save) | `j@djinna.com` | `srv/transmittal_notify.go` | Notification that a client is editing a transmittal. Throttled: max 1 per project per 30 min. Skipped when admin edits (X-ExeDev-UserID header present). |

## Pathway Details

### 1. Transmittal Email (`srv/email.go`)
- Triggered by: "Email" button on `/{client}/{project}/transmittal/` page
- Parses `transmittalEmailData` from the stored JSON blob
- Subject: `Transmittal [FINAL]: Book Title`
- Both HTML (styled tables, checkmarks) and plain text
- Includes link back to transmittal page

### 2. Project Snapshot (`srv/snapshot_email.go`)
- Triggered by: "📧 Email Snapshot" on the calendar toolbar
- Loads: tasks, transmittal, last 10 file log entries, last 10 journal entries
- Computes: % complete, done/active/pending counts, budget variance
- Subject: `Project Snapshot: Name — Apr 5, 2026`
- Most complex email — stat cards, task table, budget table, transmittal card, files table, journal list

### 3. Activity Email (`srv/activity_email.go`)
- Triggered by: "📧 Email" on Files or Journal tab headers
- Scoped to a configurable window (`?days=7`)
- Subject: `Activity Update: Name — Mar 29 – Apr 5, 2026`
- Lighter than snapshot — just file transfers + journal entries

### 4. Client Digest (`srv/client_digest_email.go`)
- Triggered by: "📧 Weekly Digest" on client portal (`/{client}/`)
- Aggregates activity across ALL projects for that client slug
- Only includes projects with activity in the period
- Subject: `Weekly Digest: Client Name — Mar 29 – Apr 5, 2026`
- Auth: client-level password OR any project-level auth for that client

### 5. Transmittal Update Notification (`srv/transmittal_notify.go`)
- **This is the only automatic/server-initiated email**
- Fires inside `handleUpdateTransmittal` (the auto-save handler)
- Throttle: `sync.Mutex` + `map[int64]time.Time`, 30-minute cooldown per project
- Runs in a background goroutine — doesn't block the save response
- On send failure: clears the throttle so it retries on next save
- Skips when `X-ExeDev-UserID` header is present (admin editing via exe.dev proxy)
- Fixed recipient: `j@djinna.com` (hardcoded as `txNotifyRecipient` const)
- Subject: `📋 Transmittal Updated: Book Title (client-slug)`
- Lightweight email — just project/book/author/status + link

## HTML Email Conventions

All HTML emails follow the same pattern for email client compatibility:
- Outer `<table>` wrapper with `#f4f3f9` background
- Inner `640px` table with white background, `12px` border-radius
- Purple gradient header (`#6c63ff` → `#8b83ff`)
- Inline styles everywhere (no `<style>` blocks — email clients strip them)
- Stat cards: colored rounded divs with large number + small label
- Tables: purple header row, alternating `#ffffff` / `#faf9ff` rows
- Footer: `#aaa` text with generation timestamp + link

## Auth Model

Manual emails (1–4) require authentication:
- Project-level cookie (set via password)
- Client-level cookie
- `X-ExeDev-UserID` header (exe.dev admin proxy)

Automatic email (5) inherits auth from the triggering request
(the client already authenticated to call `PUT /api/projects/{id}/transmittal`).

## Environment Variables

```
AGENTMAIL_API_KEY=am_...      # AgentMail Bearer token
AGENTMAIL_INBOX_ID=jdbb@agentmail.to  # Inbox ID for sending
PRODCAL_BASE_URL=https://jdbbs.exe.xyz  # Used for links in emails (auto-derived if unset)
```

## Adding a New Email Pathway

1. Create `srv/new_email.go` with handler + text/HTML builders
2. Register route in `srv/server.go` Handler() method
3. If manual: add SPA button + recipient modal in the relevant `.js` file
4. If automatic: call from the relevant handler, consider throttling
5. Follow inline-style HTML conventions (copy from `snapshot_email.go`)
6. Always provide both `textBody` and `htmlBody`
7. Update this file
