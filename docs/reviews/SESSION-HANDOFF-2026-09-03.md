# Session handoff — Factory Pass shipped; pre-workshop fixes (2026-09-03)

**Deployed to VM this session.** Workshop: Sep 21–22 (18 days). ~8 attendees
will each get a coupon and use the customer Factory page live.

## What exists now (read these first)
- `docs/specs/FACTORY-PASS-API-2026-09-03.md` — the contract everything was built to.
- `docs/reviews/STOREFRONT-FACTORY-PASS-DESIGN-2026-09-03.md` — product design.
- `docs/reviews/FACTORY-PASS-PRICING-TAM-2026-09-03.md` — $149 pass, +3 builds $49, review $75, live $100/hr.
- `docs/reviews/FACTORY-PASS-SMOKE-2026-09-03.md` — first e2e run, what passed, what didn't.
- `docs/reviews/ORCHESTRATION-RETRO-2026-09-03.md` — how to fan out subagents without 90-min black boxes. **Apply its rules if you fan out again.**
- Code: `srv/passes.go` (+`passes_test.go`), migration 023, `srv/static/factory.{html,js,css}`, `jdbbs-public/factory.html` (→ `/factory`), tracker code issuance in `registrations.html`.

Flow: admin issues code in `/admin/registrations` → attendee redeems at
`/factory` → gets client+project+pass, emailed password → works at
`/{client}/{project}/factory/`. Build = PDF+EPUB, 3 included, failures refunded.

Context: **no books were built on prod between late May and today** (client #1
had its own delays), which is why the pandoc 3.1.3 breakage went unnoticed.
Now pandoc 3.11; see DEPLOY.md "Pipeline tools".

---

## Session A — blockers before Sep 21 (do in this order)

### 1. Password recovery / reset (BLOCKER)
Fulfilment emails the only copy of the generated client password. Today
"Lost the password?" on the Factory gate is a **mailto** to j@djinna.com, and
admin has **no reset endpoint** — during the smoke I had to bcrypt by hand.
With 8 people live on Zoom, one lost password = one stuck attendee.

Minimum: `POST /api/admin/clients/{slug}/password` (admin) that sets a new
random password (reuse `generateClientPassword` + `hashPassword` in
`srv/passes.go`/`server.go`) and re-sends the fulfilment-style email
(`sendPassFulfilmentEmail` or similar in passes.go — reuse it), plus a
"Reset & resend password" button on the tracker card in `registrations.html`
(it already knows the coupon → pass → client). Also confirm that if an
attendee redeems *before* email is configured/working, the admin can recover.

Better (post-workshop): magic-link login (email → one-time link → client
cookie). One table, two endpoints; AgentMail already wired. Don't do this
before the workshop.

Also check: does `s.Email` exist on prod (`.env`)? Fulfilment and delivery
mails are silently skipped when nil — smoke log showed the WARN path. Send
one real redeem to yourself on prod and read the email.

### 2. EPUB size — 14 MB for a 3-chapter text file
`srv/epub.go` ~line 181 unconditionally embeds NotoSerifTC Regular+Bold
(8 MB each) and Noto Thai for CJK/Thai fallback. For a Latin-script book
that's 16 MB of dead weight; Amazon charges delivery fees by size on the 70%
royalty plan and the download feels wrong to a customer.

Fix: make embedding conditional on script detection. The preflight already
produces a `language_script` finding (`srv/preflight.go`) — reuse that
detector (or a cheap Unicode-range scan of the docx text) and only embed the
families the text actually needs. Keep the `/licensed/` guard. Add a test:
Latin-only docx → no font files in the zip; a CJK sample → TC fonts present.
Target: text-only EPUB < 500 KB.

### 3. Customer-facing build errors are raw Typst traces
The Twitter Years docx failed with `unknown variable: tweet-p` + a Typst
source excerpt, verbatim in the UI. Attendees' Word files *will* carry custom
styles the template doesn't know. Fix in `failConversion` (srv/books.go):
classify the error and store a customer message in `error_msg` while logging
the raw trace:
- `unknown variable: X` → "Your file uses a Word style (X) that isn't in your
  template. Inspect lists styles not in your transmittal — remove or remap
  it, or ask us to add it."
- pandoc failures → "We couldn't read this Word file. Re-save it as .docx
  from Word and try again."
- anything else → generic + a support pointer.
Keep the raw error for admin (add an `error_detail` column or log-only).
Also (#5 in the smoke notes): when status is `ready` but `error_msg` holds
the non-fatal "EPUB failed" note, the UI should render a warning, not a
failure.

### After 1–3: dry run
Issue yourself a coupon on prod, redeem, upload a *messy* real .docx, inspect,
build, download both files, lose the password, reset it. Then issue the 8
codes and put them in the prep announcement (tracker → announcements) with
"redeem before session 1."

---

## Session B — QOL (after A)
- Delivery email links to `/download/epub` even for PDF-only builds.
- Redeem/verify rate limiting review; the `Lost the password?` link should
  become the reset flow from A.1.
- Admin passes view (`GET /api/admin/passes` exists; no UI beyond tracker).
- `builds_used` visible in the tracker per attendee — that's the dogfood data
  for tuning N and pricing.
- Expiry job + T-30/T-7 emails (nothing expires before Mar 2027).
- Stripe Checkout → `fulfillPass("stripe")` (deferred by decision).
- Preflight Review SKU ($75) needs only a mailto + a paid link for now.
- The two agent commits `0c42445`/`fac9a28` contain UI files under backend
  messages — fine, just don't be confused by `git log -- srv/static/factory.js`.
- pandoc pinned in DEPLOY.md; consider a `make doctor` target that checks
  pandoc ≥3.2, typst 0.12, python-docx.
