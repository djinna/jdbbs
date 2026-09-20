# Email System — Architecture & Reference

All email goes out through one transport function (`EmailConfig.send`, `srv/email.go`),
which speaks to one of two providers:

- **Resend** (since 2026-09-18, preferred) — turned on by `PRODCAL_MAIL_FROM`
  (e.g. `factory@mail.jdbb.studio`); calls go through the exe.dev proxy
  `https://resend.int.exe.xyz` (key injected at the edge) or, with
  `RESEND_API_KEY` + `PRODCAL_RESEND_URL=https://api.resend.com`, direct.
  The sending domain `mail.jdbb.studio` carries our own SPF/DKIM/DMARC
  (records at Porkbun).
- **AgentMail** (fallback when `PRODCAL_MAIL_FROM` is unset) — the original
  transport; `jdbb@agentmail.to` stays as the archive inbox / CC option.
Configured via environment variables `AGENTMAIL_API_KEY` and `AGENTMAIL_INBOX_ID`.
Sender address: `PRODCAL_MAIL_FROM` on Resend, else `jdbb@agentmail.to`. Reply-To is `j@djinna.com` either way.

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

There are **8 email pathways** in two categories (5 manual + 3 automatic):

### Manual (button-triggered, user picks recipients)

All of these use the same SPA modal pattern: default recipients
(`jdbb@agentmail.to` + `j@djinna.com`) plus an editable "Other" row.
First recipient = To, rest = CC.

| # | Trigger | Endpoint | File | Description |
|---|---------|----------|------|-------------|
| 1 | **Email button** on transmittal form | `POST /api/projects/{id}/transmittal/email` | `srv/email.go` | Full transmittal summary (book info, production dates, checklist, design, etc.) |
| 2 | **Email Snapshot button** on calendar toolbar (admin-console buttons removed 2026-09-19, punch list 0.22 — calendar-era digest) | `POST /api/projects/{id}/snapshot/email` | `srv/snapshot_email.go` | Comprehensive project snapshot (schedule, budget, transmittal status, recent files, recent journal) |
| 3 | **Email button** on Files/Journal tabs (admin-console button removed 2026-09-19, 0.22) | `POST /api/projects/{id}/activity/email` | `srv/activity_email.go` | Activity digest — file transfers + journal entries from last N days (default 7, `?days=N`) |
| 4 | **Weekly Digest button** on client portal | `POST /api/clients/{client}/digest/email` | `srv/client_digest_email.go` | Aggregated digest across ALL projects for a client — last 7 days of files + journal |
| 5 | **Send announcement** on workshop cohort tracker | `POST /api/admin/registrations/announce` | `srv/registration.go` | Personalized one-to-one announcement to opted-in, non-declined registrants; logs each batch |

### Automatic (server-initiated, no user action)

| # | Trigger | Recipient | File | Description |
|---|---------|-----------|------|-------------|
| 6 | **Client updates transmittal** (auto-save) | `j@djinna.com` | `srv/transmittal_notify.go` | Notification that a client is editing a transmittal. Throttled: max 1 per project per 30 min. Skipped when admin edits (X-ExeDev-UserID header present). |
| 7 | **Factory Pass fulfilled** / **password reset** / **build delivered** | the pass customer | `srv/passes.go` | Transactional storefront mail: the welcome/sign-in message (also re-sent after an admin password reset) and the per-build receipt. |
| 8 | **Client asks for a sign-in link** (portal / factory gate) | the address they typed, only if it is on file for that client | `srv/login_links.go` | Magic-link sign-in: one-shot `/auth/link?t=…`, 30-minute expiry, replaces typing the password. Silent when the address is unknown or the client has hit 3 live links in 15 min. |

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

### 5. Workshop Cohort Announcement (`srv/registration.go`)
- Triggered by: **Send announcement** on `/admin/registrations`
- Admin selects recipients; the server re-checks `consent_email=1` and excludes declined registrations
- Sends one personalized message per recipient (no exposed To/CC list)
- Updates each successful recipient's `last_emailed_at` and logs the batch in `event_announcements`
- Includes a consent reminder and reply-to-opt-out language
- **Merge fields** (`mergeAnnouncement`): `{{first}}` `{{name}}` `{{code}}` `{{factory_url}}`, and
  conditional blocks `{{#code}}…{{/code}}` (unredeemed code issued), `{{#redeemed}}…{{/redeemed}}`
  (pass redeemed; `{{factory_url}}` is the customer factory page), `{{#nocode}}…{{/nocode}}` (no code).
  A bare `{{code}}`/`{{factory_url}}` for a recipient who has none aborts the whole batch before any send.
- `"preview": true` in the request body returns the merged plain-text letters per recipient without
  sending or logging anything — **Preview merge** on the tracker uses it. Works with the mailer off.
- Bare URLs in the body are linked in the HTML part.

### 6. Transmittal Update Notification (`srv/transmittal_notify.go`)
- **This is the only automatic/server-initiated email**
- Fires inside `handleUpdateTransmittal` (the auto-save handler)
- Throttle: `sync.Mutex` + `map[int64]time.Time`, 30-minute cooldown per project
- Runs in a background goroutine — doesn't block the save response
- On send failure: clears the throttle so it retries on next save
- Skips when `X-ExeDev-UserID` header is present (admin editing via exe.dev proxy)
- Fixed recipient: `j@djinna.com` (hardcoded as `txNotifyRecipient` const)
- Subject: `📋 Transmittal Updated: Book Title (client-slug)`
- Lightweight email — just project/book/author/status + link

### 7. Factory Pass Mail (`srv/passes.go`)
- **Automatic/server-initiated**, transactional (the customer bought this; no consent flag applies)
- Two messages, both to `passes.customer_email`:
  1. **Fulfillment / password reset** — sent from `fulfillPass` callers (`POST /api/public/redeem`, `POST /api/admin/passes`) and re-sent by `POST /api/admin/clients/{slug}/password`. Subject: `Your Factory Pass: {title}`. Carries the portal URL (`/{client}/{project}/factory/`), the client sign-in slug, the generated 12-character password, what's included (3 builds, unlimited preflights, expiry date), first steps, and the support edges. Password-reset responses also return the replacement password once to the authenticated admin, so recovery still works if mail is unavailable; plaintext is never stored.
  2. **Build delivered** — sent from `runConversion` after a successful **final** build only (0.28, 2026-09-20: proof builds are free, unlimited and silent — no email, no credit, PROOF footer on the PDF). Subject: `Final files ready: {title}` (both) or `Final print PDF ready: {title}`. Links to the newest final file(s) (`?kind=final`) plus the preflight report; nudges "read the EPUB first"; states "Finals remaining: N of M. Proofs are free and unlimited."
- Fulfillment mail is fire-and-forget in its own goroutine, so the mailer can never fail a redemption; password-reset mail is synchronous so the tracker can report whether the new credential was actually sent; the build mail runs on the conversion goroutine (already off the request path)
- `s.Email == nil` (local/dev/test) → log a warning and skip; fulfillment still succeeds
- Support-edge copy lives in one place, `passSupportEdges(workshop bool)`, shared by both mails; the factory page mirrors it (`data-workshop` / `data-post-workshop` in `factory.html`, gated in `factory.js`)
- **Workshop switchover (2026-09-20):** all Sep 21–22 workshop copy — the fulfillment mail's "workshop code" intro, the session-1/2 sentences in First steps, the "Live help during the workshop sessions" support edge, and the registration auto-reply's schedule — is gated on `workshopEndsAt` (`srv/passes.go`, end of day Tue Sep 22 HKT, same instant WORKSHOP49 expires). After that it drops out automatically; nothing to hand-edit. `/admin/email-preview/` shows the `*_after_switchover` variants.
- Expiry warnings (T-30/T-7) and purge notices are **deferred** — nothing expires before March 2027
- The fulfilment mail now leads with "Easiest way in: open your portal and enter this email address — we'll send you a sign-in link" above the password line, so pathway 8 is the expected first login; the password stays valid as the fallback.

### 8. Client Sign-in Link (`srv/login_links.go`) — added 2026-09-18, punch list 5.7
- **Automatic/server-initiated**, transactional. Triggered by the customer on `/{client}/` or `/{client}/{project}/factory/`: the auth gate's primary path is "Enter your email → we'll send you a sign-in link"; the password box sits behind "Have a password? Use it instead". The old "Forgot/Lost the password?" mailto links are gone — the same email form replaces them.
- `POST /api/public/login-link` `{client, email}` → **always** `200 {ok:true}` (no account enumeration). A link is created and mailed only when the address (case-insensitive) is one we already know for that client: `clients.email` (new column, migration 042), a Factory Pass `customer_email` on one of the client's projects, a store order for such a pass, or the coupon / workshop registration the pass was redeemed from.
- Subject: `Your sign-in link for {client name}`. Body: the link, "works once and for 30 minutes", "if it has expired, go back to your portal and enter your email again", "if you didn't ask for this, ignore it". Kind `login_link`, ref `client:{slug}`, triggered_by `public`. Fire-and-forget goroutine (the request has already answered ok).
- Token: 32 random bytes, base64url in the URL; only the sha256 hex is stored (`login_links.token_hash`). `GET /auth/link?t=…` verifies hash + not expired + not used, marks it used (guarded UPDATE so a double-click resolves to one winner), sets the **same** `prodcal_client_{slug}` cookie a password login sets (`setClientAuthCookie`, shared with `handleClientVerify`), and 302s to `/{client}/`. Anything else → a themed "That sign-in link has expired" page (401) with a button back to `/{client}/` to request a new one.
- Rate limit: max 3 unused links per client per 15 minutes; over the limit the endpoint still answers ok and simply does not send (logged as a warning).
- Logged with `slog` (`login link issued` / `login link redeemed` / `login link rejected`, never the token) and as factory events `login.link_sent`, `login.link_denied`, `login` (detail "via emailed link") so the admin Floor feed shows sign-ins.
- **Caveat:** like every other pathway the Mail report (`outbound_email`) keeps a copy of the body, which includes the link. Links are single-use and expire in 30 minutes, so the exposure window is small, but it is there.
- Previews: `/admin/email-preview/login_link`. Tests: `srv/login_links_test.go`.

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

Manual emails (1–5) require authentication:
- Project-level cookie (set via password)
- Client-level cookie
- `X-ExeDev-UserID` header (exe.dev admin proxy)

Automatic email (6) inherits auth from the triggering request
(the client already authenticated to call `PUT /api/projects/{id}/transmittal`).
Automatic email (7) is addressed from the pass row, not from the request, so a
build run by the admin still mails the customer who owns the pass.
Automatic email (8) requires no auth at all — it *is* the way in — which is why
it never reveals whether the address or the client exists, and never mails an
address it does not already have on file.

## Environment Variables

```
PRODCAL_MAIL_FROM=factory@mail.jdbb.studio  # Resend on; unset → AgentMail
AGENTMAIL_API_KEY=am_...      # AgentMail Bearer token (fallback transport + archive inbox)
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
