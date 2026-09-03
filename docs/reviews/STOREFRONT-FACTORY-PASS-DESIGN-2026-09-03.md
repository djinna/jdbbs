# Storefront: the "Factory Pass" — access + deliverables design (2026-09-03)

**Status:** thinking document. Decisions marked **[DECIDE]** need the user.
Supersedes the One Shot detour in `HANDOFF-storefront-planning-2026-09-03.md`:
the product is now concrete, so the question is no longer "which kit" but
"what exactly is sold, who can press which button, and what do they walk away
with." Target dogfood: the Protocolize Your Book workshop, **Sep 21–22, 2026**
(18 days), via free-access codes for the ~8 attendees.

## 0. Ground truth (read from the code today)

| Fact | Where | Consequence |
|---|---|---|
| Upload / convert / preflight / detect-chapters are **admin-only** (`requireExeDevAdminAPI`) | `srv/books.go`, `srv/preflight.go` | No non-admin can run the factory. This is *the* gap. |
| All build UI lives in `admin.html`; client portal + project SPA have none | `srv/static/` | A customer-facing "build" surface must be created. |
| Client auth = slug + shared password (cookie); project tokens; admin via `X-ExeDev-UserID` header | `srv/server.go checkAuth` | Reusable as-is for v1; no new auth primitive required. |
| Clients may create their own projects if the client has a password | `srv/client.go handleClientCreateProject` | Onboarding pattern exists. |
| Source docx + PDF/EPUB + output history are SQLite blobs | migrations 007, 014 | "Storage" = rows; expiry is a policy + purge job, not infra. |
| `event_registrations` has emails, consent, announcement history | migration 021/022, `registrations.html` | Codes can be issued per-attendee from the tracker. |
| Public form pattern: pi-public HTML → `/api/public/*` with honeypot + IP rate limit | `srv/registration.go` | Redemption form reuses this exactly. |
| Email pathways via AgentMail | `srv/EMAIL_SYSTEM.md` | Fulfillment/expiry mails are cheap to add (pathway #6). |
| Licensed fonts: PDF embedding OK, EPUB guarded | `typesetting/fonts/licensed/README.md` | Customer PDFs are license-clean. |

## 1. What is being sold — define the unit precisely

**Proposal: the *Factory Pass* — one manuscript, one project, through the
whole protocol.** Not "one conversion." The workshop's own framing is the
lifecycle (transmittal → template → preflight → build), and that's where the
value is; the pandoc→typst compile is seconds of CPU.

A pass contains:

1. **One project** (transmittal + generated Word authoring template + book spec).
2. **Unlimited preflights.** Preflight is the negotiation step; making it free
   steers people toward fixing their Word file instead of burning builds.
3. **N included builds** (a build = one successful convert producing PDF+EPUB).
   Failed builds (pandoc/typst error) don't count.
4. **Storage/liveness for 6 months**: project stays rebuildable and
   downloadable; output history retained.
5. **Deliverables** (§3).

**[DECIDE] N.** Reasoning: the constraint is not compute, it's (a) support
scope creep — the fear is an unbounded "typeset my book until I love it"
consultancy — and (b) a clean line for "this is a different manuscript." Options:

| Model | Pros | Cons |
|---|---|---|
| **Count only: N=5 builds, then buy more** | Simplest to meter and explain; matches "charge for addl runs" | Penalizes honest iteration on a messy source |
| Time only: unlimited builds for 30 days | Generous, encourages the fix loop | Doesn't produce "addl run" revenue; hard to stop the 31st-day grumble |
| **Hybrid (recommended): 5 builds included; extra builds in packs of 5; the pass also expires at 6 mo regardless of credits** | Both levers; credits are the *only* thing metered in code | Two numbers to communicate |

Recommendation: **5 included, +5 pack as an add-on SKU, preflights free,
failed builds refunded.** Tune N after watching the workshop cohort's actual
build counts (log every build with a reason field — that data is the point of
the dogfood).

"Different manuscript" rule: a pass is bound to one project. Re-uploading a
revised Word file of the *same* title is a rebuild (uses a credit). A new title
= new pass. Enforced softly (one project per pass) — don't try to detect
"substantively different."

## 2. Access — who presses the button

### Options considered

**A. Concierge (zero code).** Customer emails the .docx; admin uploads and
builds; customer downloads from the existing portal. Works for 8 attendees with
the instructor in the room. Rejected as the *product*: it's not a storefront and
it doesn't test anything the storefront needs. Keep as the fallback if §5's
build slips.

**B. Extend the existing client/project model (recommended for v1).**
Fulfillment creates a `client` (slug, generated password) + `project` +
`pass`. Loosen upload/convert/preflight/detect from admin-only to
`requireAuth(projectID)` **plus an entitlement check** (project has a live
pass with credits). Reuses `checkAuth`, the portal URL scheme
`/{client}/{project}/`, and the cookie flow. Needs: server gating + a
customer-facing build panel (§5).

**C. Passwordless magic-link (the One Shot pattern).** Identity = email; login
via emailed one-time link → session cookie. Better UX than a shared password,
and AgentMail is already wired. It's a *login* change, orthogonal to B — B's
entitlement work is needed either way. Defer: do B for the workshop; if
passwords annoy the cohort, C is a bounded follow-up (one table, two endpoints).

**D. Ship software instead (the Mac app).** "Buy the factory, run it locally."
Different product — no storage, no re-do metering, no server. Not this.

### Recommended v1 flow (coupon path == paid path minus payment)

```
[/factory page]  ── "Have a code?" ──▶  POST /api/public/redeem
                                         {code, email, name, title, author}
                                                │
                                                ▼
                                    fulfillPass(source="coupon")
                                    • create client (slug from name, random pw)
                                    • create project (slug from title)
                                    • create pass (5 credits, expires +6mo)
                                    • email: portal URL + password + "how to start"
                                    • link to event_registrations row if email matches
                                                │
                                                ▼
                         /{client}/{project}/   ← customer: upload → preflight → build → download
```

Later, `POST /api/stripe/webhook` (checkout.session.completed) calls the
**same `fulfillPass(source="stripe")`**. The workshop therefore dogfoods every
part of the product except the payment rail — which is the right thing, since
the rail is the most commoditized piece.

### Codes for the workshop

- **Unique per attendee**, not one shared code. Issued from the cohort tracker
  (`registrations.html` gets an "Issue pass code" action), pre-bound to the
  registration email. Redemption with a different email still works but is
  flagged. Gives attribution and lets you revoke one without the others.
- Coupon fields: `code, sku, max_redemptions=1, redeemed_count, expires_at
  (Sep 30), registration_id NULL, note`.
- Recommend sending the code in the pre-workshop prep email (announcement
  machinery already exists), asking attendees to redeem **before session 1**
  so onboarding failures surface with days of slack, not live.

**[DECIDE]** Stripe now or after? My recommendation: **after.** Build
fulfillment + entitlement + redemption now (must work on Sep 21 regardless);
add Stripe Checkout + webhook in the following weeks. Stripe's own promotion
codes could replace the coupon table later, but a 100%-off Stripe Checkout
still drags attendees through a Stripe form, and it puts a third-party
integration on the critical path 18 days out.

### Gating rules (server)

- `POST /api/books/upload`, `/detect-chapters`, `/api/projects/{id}/preflight*`:
  `requireAuth(projectID)` and project has a **live** pass (not expired).
- `POST /api/books/{id}/convert`: same, **and** `credits_remaining > 0`.
  Debit on start; refund on `failConversion`. Ledger row per build.
- Book must be linked to the project the caller is authorized for (the current
  book→project link is optional; for pass projects make it mandatory on upload).
- Admin bypass unchanged.
- Rate-limit builds per project (e.g. 1 in flight, ≤10/hour) independent of
  credits — credits fence *scope*, the limiter fences *abuse*.

## 3. Deliverables — what they walk away with

| Deliverable | Today | For the pass |
|---|---|---|
| Print PDF | ✔ (admin builds) | ✔ download from portal; every build kept in output history |
| EPUB | ✔ | ✔ same |
| Generated Word authoring template (from transmittal) | ✔ (transmittal flow) | ✔ — call it out explicitly; workshop page already promises it |
| Preflight report (HTML) | admin-only | expose to customer — it *is* the "what to fix" guidance and reduces support |
| Typst source | internal | **[DECIDE]** not by default (template + licensed-font references). Could be a paid "source bundle" SKU later. |
| Transmittal (the filled spec) | ✔ readable | ✔ — it's their protocol record |
| A single "delivery" email per successful build | ✗ | ✔ — links to PDF/EPUB/report; the receipt of the thing they bought (`EMAIL_SYSTEM.md` pathway #6) |

**Expectation-setting copy is a deliverable too.** The factory produces what the
protocol allows; garbage-in yields a correctly typeset pile of garbage. The
preflight is the honest broker. The page should say: *"Deliverable = a
correctly typeset PDF+EPUB of the manuscript as it conforms to your
transmittal. Preflight tells you what doesn't conform."* That sentence is the
refund policy.

## 4. Storage / expiry — the 6 months

- `pass.expires_at = fulfilled_at + 6 months`.
- Warning emails at T-30d and T-7d (consent-aware; they bought it, so
  transactional).
- At expiry: project → **read-only** (no build/upload); downloads still work
  for a **30-day grace**; then purge blobs (source, outputs) and keep metadata
  (title, dates, build count) for the customer's and your records.
- **Extension SKU**: +6 months (cheap). **Reactivation**: buying a build pack
  on an expired-but-not-purged pass also extends it. After purge, it's a new
  pass.
- Privacy framing: unpublished manuscripts are the customer's IP. A hard
  expiry + purge is a *feature*; say so on the page. **Check** that purge
  propagates to whatever `backup_status.go` snapshots — if backups retain
  blobs for 90 days, the policy must say so.

## 5. Minimum build for Sep 21 (ordered by risk)

1. **Schema (1 migration):** `passes`, `pass_ledger`, `coupons`. ~1 hr.
2. **`fulfillPass()` + `POST /api/public/redeem`** (honeypot + limiter, same
   as register) + fulfillment email. ~half day.
3. **Server gating** per §2 (loosen 5 handlers, add entitlement check, debit/
   refund in `runConversion`). ~half day. Tests: auth matrix + credit ledger.
4. **Customer build panel** in the project SPA (`index.html`/`app.js`): upload
   .docx, "Inspect" (preflight) with report link, "Build" with status polling,
   outputs list with downloads, credits + expiry badge. This is the biggest
   chunk and the biggest UX risk — lift the DOM/JS from `admin.html`'s
   typesetting tab rather than redesign. ~1–2 days.
5. **`/factory` pi-public page**: the offer, what's included, the constraint,
   the redeem form. Copy is most of the work. ~half day.
6. **Cohort tracker**: "Issue code" button + code column. ~2 hrs.
7. Expiry job + warning emails — **can slip past the workshop** (nothing
   expires before March). Log a follow-up.
8. Stripe Checkout + webhook → `fulfillPass("stripe")` — **after**.

Fallback if (4) isn't ready by ~Sep 17: ship (1)(2)(3)(5)(6) and run builds
from admin during session 3 (Option A), with credits still debited so the
ledger data is real.

## 6. Open questions **[DECIDE]**

1. Price of a pass, of a +5 build pack, of +6 mo storage. (Anchors: what would
   you charge to typeset one book by hand? The pass should be the "no human
   hours" tier.)
2. N included builds — 5? See §1.
3. Login: shared password (B) is fine for the workshop; do you want magic-link
   (C) before public launch?
4. Typst source as a deliverable: never / paid SKU / included?
5. Does a pass include *any* human time (one review email? none?) — the
   support fence matters more than the compute fence.
6. Store domain: everything here assumes `jdbbs.exe.xyz/factory`. Custom
   domain affects Stripe branding and the emails, nothing else.
7. Anthologies (multi-author, `chapters` metadata) — same pass or a different
   SKU? Suggest: same pass for v1, revisit.

## 7. What One Shot was actually good for

Its useful lesson survives: **entitlement gating as a first-class table**
(§2 gating rules), **fulfillment as one function with multiple sources**
(coupon/stripe/admin-grant), and **the agent-readable conventions** you already
have. Nothing here needs its code.
