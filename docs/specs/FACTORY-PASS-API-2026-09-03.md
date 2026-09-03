# Factory Pass — API contract (v1, workshop build)

Shared contract between backend (`srv/passes.go`), customer UI
(`srv/static/factory.html`), admin tracker (`srv/static/registrations.html`),
and the public offer page (`pi-public/factory.html`). **Change here first.**

Decisions (2026-09-03): 3 included builds; unlimited preflights; 6-month
storage; failed builds refunded; no Typst source; Stripe deferred (fulfillment
is a single function with `source` = `coupon` | `admin` | later `stripe`).

## Tables (migration 023)

```sql
passes(
  id, project_id UNIQUE NOT NULL → projects,
  sku TEXT ('factory-pass'), source TEXT ('coupon'|'admin'|'stripe'),
  coupon_id NULL → coupons, customer_email TEXT, customer_name TEXT,
  builds_included INT (3), builds_used INT (0), builds_extra INT (0),
  fulfilled_at, expires_at (fulfilled_at + 6 months),
  status TEXT ('active'|'expired'|'purged'|'revoked'),
  note TEXT)
pass_ledger(id, pass_id, book_id NULL, delta INT (-1 debit, +1 refund, +N pack),
  reason TEXT ('build'|'build_failed_refund'|'pack'|'grant'|'revoke'),
  created_at)
coupons(id, code UNIQUE, sku, max_redemptions INT (1), redeemed_count INT (0),
  expires_at NULL, registration_id NULL → event_registrations,
  issued_to_email TEXT, note TEXT, created_at)
```

`credits_remaining = builds_included + builds_extra - builds_used`.
`live = status='active' AND expires_at > now`.

## Public

`POST /api/public/redeem` — honeypot field `company`, per-IP limiter (5/10min).
```json
{"code":"PYB-7K3M-QX2A","name":"…","email":"…","title":"…","author":"…","company":""}
→ 200 {"ok":true,"portal_url":"https://host/{client}/{project}/factory/",
        "client_slug":"…","project_slug":"…"}
→ 400 {"error":"…"}  (bad/expired/used code, missing fields)
```
Fulfillment: create client (slug from name, unique-ified; random 12-char
password, hashed), project (slug from title), pass, coupon.redeemed_count++,
send fulfillment email (portal URL, client password, first steps). If
`email` matches an `event_registrations` row, link `coupons.registration_id`
if not already set. Reject with 400 if the code is already used — do NOT
silently re-fulfil.

`GET /api/public/config` — existing; add `"factory": {"builds_included":3,
"storage_months":6}` for page copy.

## Customer (auth = existing `requireAuth(projectID)`: project token cookie,
## client password cookie, or admin header)

`GET  /api/projects/{id}/pass` →
```json
{"exists":true,"status":"active","live":true,"builds_included":3,"builds_extra":0,
 "builds_used":1,"credits_remaining":2,"expires_at":"2027-03-03T…Z",
 "customer_name":"…"}
```
`{"exists":false}` for projects with no pass (admin-only projects behave as before).

`GET  /api/projects/{id}/books` → `[ListBooksRow…]` for that project only
(existing `ListBooksByProject`).

`POST /api/books/upload` (multipart: file,title,author,series,project_id) —
now `requireAuth(project_id)` when `project_id` present **and the project has a
live pass**; else admin. Non-admin: `project_id` is mandatory.

`POST /api/books/{id}/detect-chapters`, `POST /api/projects/{id}/preflight`,
`GET /api/projects/{id}/preflight`, `GET /api/projects/{id}/preflight/report`
— `requireAuth(project)` + live pass; else admin. Unmetered.

`POST /api/books/{id}/convert` — `requireAuth(project)` + live pass +
`credits_remaining > 0` (else `402 {"error":"no builds remaining", "credits_remaining":0}`);
one in-flight build per project (`409`). Debit ledger `-1 build` before
starting; `failConversion` writes `+1 build_failed_refund`. Admin header:
gating skipped **but ledger still debited/refunded if a pass exists** (so
workshop-instructor-run builds count and the dogfood data is real).

`GET /api/books/{id}/download/{format}`, `/outputs`, `/outputs/{oid}/download`
— unchanged (already project-auth).

Page route: `GET /{client}/{project}/factory/` → serves
`static/factory.html` (same pattern as `/transmittal/`; assets under it served
from static). The page resolves its project via the existing
`GET /api/project-by-path/{client}/{project}`.

## Admin (`requireExeDevAdminAPI`)

`POST /api/admin/coupons` `{"registration_id":12,"note":"…"}` or
`{"issued_to_email":"…","note":"…"}` → `{"code":"PYB-…","id":…}`.
Code format: `PYB-XXXX-XXXX`, uppercase, no 0/O/1/I. `expires_at` default
2026-09-30.
`GET  /api/admin/coupons` → list with redemption status + pass/project link.
`GET  /api/admin/passes` → list (project, client, credits, expires, status).
`POST /api/admin/passes/{id}/grant` `{"builds":3,"reason":"pack"}` → ledger +N.
`GET  /api/admin/registrations` — existing rows gain `coupon_code`,
`coupon_redeemed_at` (LEFT JOIN) so the tracker can show issued/redeemed.

## Emails (new pathway #6 in `srv/EMAIL_SYSTEM.md`)

1. **Fulfillment** (on redeem): subject "Your Factory Pass: {title}". Portal
   URL, client slug + password, what's included (3 builds, unlimited
   preflights, until {expires}), "start with the transmittal", support edges
   (see below).
2. **Build delivered** (on successful convert, to pass.customer_email): links to
   PDF / EPUB / preflight report; credits remaining.
Expiry warnings (T-30/T-7) and purge: **deferred**, nothing expires before
March 2027.

## Support edges (copy, used by page + emails)

- Included: the workshop sessions themselves (Sep 21–22).
- After: email support is **not** included. Self-serve = preflight report +
  docs. Live support: **USD 100/hr**, booked in advance, 30-min minimum.
- Reply-To on transactional mail should point at an address that autoresponds
  with those edges (config, not code).
