# Protocol Institute — first client through the whole factory · **Test 2**

**2026-09-17.** Test 1 (14:46–15:10 UTC) started, then aborted after signup — artefacts kept below. Sandbox Stripe (`stripe-test`). Client `production@protocol-institute.org`.
Legend: **YOU** = your turn · **ME** = Shelley's turn · **BOTH** = look together.
Ticked = done. This page refreshes itself every 20 s.

## 0 · Instrumentation — ME

- [x] 0.1 Request log (non-GET + every 4xx/5xx, with who) + slog for login / upload / Inspect / transmittal
- [x] 0.2 `scripts/factory-tail.sh`; running in tmux `factory-tail`
- [x] 0.3 Build, tests green, restart, commit, push
- [x] 0.4 This page live

## 1 · Signup from the sales page — YOU

- [x] 1.1 Open `https://jdbbs.exe.xyz/factory` → **Buy a Factory Pass**. Use a **private window** so nothing from the admin session leaks in
- [x] 1.2 Stripe Checkout: email `production@protocol-institute.org`, **cardholder name is what the sign-in slug comes from** (C1) — `jdixon` is taken by test 1, so use e.g. "Protocol Institute" → `/protocol-institute/`; card `4242 4242 4242 4242`, any future date, any CVC. **No promo code** this time (full-price path)
- [x] 1.3 Land on the success page — note what it tells you to do next
- [x] 1.4 ME — watch: `checkout session created` → `store: fulfilled` → `factory pass fulfilled` → `Your Factory Pass` email; check `/admin/store/` row, pass + client + project created
- [x] 1.5 YOU — open the inbox for `production@…`: pass email arrived? Does it read right? (audit copy BCC'd to j@)

## 2 · Portal + transmittal — YOU

- [ ] 2.1 Follow the link in the pass email, log in with the password it gives
- [ ] 2.2 ME — `client login` logged; no 4xx on the way in
- [ ] 2.3 YOU — Transmittal: title, author, trim, formats, any custom styles. **Save draft** first
- [ ] 2.4 YOU — **Mark Final**
- [ ] 2.5 ME — `transmittal saved status=final`, template built, `template_ready` email sent
- [ ] 2.6 YOU — download the Word template (portal or email link); open it; styles list matches what you declared?

## 3 · Manuscript + Inspect — YOU

- [ ] 3.1 Upload the manuscript **as it is today** (before applying the template) → **Inspect**. This is the real-world Inspect data point
- [ ] 3.2 ME — `manuscript uploaded` (size) → `inspect run` counts; compare with what the report page shows you
- [ ] 3.3 BOTH — read the findings: false positives? unclear labels? anything you'd want it to have caught?
- [ ] 3.4 YOU — apply the template (Organizer or paste into template), re-upload, re-Inspect
- [ ] 3.5 ME — counts dropped as expected; Keep / Strip / Convert decisions recorded

## 4 · Build — YOU

- [ ] 4.1 **Build**
- [ ] 4.2 ME — `book conversion starting` → `complete` (elapsed, sizes), credit 3→2 in the ledger, `Build ready` email
- [ ] 4.3 YOU — download PDF + EPUB. Check: trim, title page, chapter openers, section breaks, running heads, EPUB nav
- [ ] 4.4 BOTH — catches → list below

## 5 · Second title (optional today) — BOTH

- [ ] 5.1 New project under the same client from the portal
- [ ] 5.2 ME — attach a pass to it (API; admin button is on the pre-freeze list)
- [ ] 5.3 Repeat 2–4

## 6 · Wrap — ME

- [ ] 6.1 Fix small catches inline; park big ones
- [ ] 6.2 Handoff addendum; commit, push
- [ ] 6.3 BOTH — freeze / hotfix policy for Sep 19–23

## Test 1 — aborted (artefacts kept)

Client `jdixon`, project 21 `jdixon/book-001` "Obliquities", **pass 6 revoked** (bundle $427: pass + 3 builds + 6 mo storage), store order 3, session `cs_test_…E46W`, pass email received 14:48 UTC (old wording).

## Catches

- **C1 (design question)** Client slug comes from the Stripe cardholder name → `jdixon`, portal at `/jdixon/`. An organisation buying for an author gets a URL named after whoever paid. Options: ask for "press / author name" on the sales form, or let the client rename in the portal.
- **C1b** Slugger is *first initial + surname*: "Protocol Institute" → `pinstitute`. Fine for people, odd for presses/orgs. Same fix as C1.
- **C1c** "Hi Protocol," — `firstName()` on an org name. Part of the C1 fix.
- **C4 (yours, 1 min)** Email footer saved value reads "Replied to this email go to Jenna." — fix in `/admin/docs/` → Email strings (or clear to fall back to the default "Reply to this email to reach Jenna.").
- **Obs.** Pass email "Support" (3 bullets) and the sales page "Support edges" (Preflight Review $75 / Live help / Studio typesetting $800) don't quite match. Decide which is canonical at wrap.
- **C5 (fixed `6486a72`)** Factory link in the pass email not obviously clickable (code style, no underline) → underlined.
- **C2 (fixed `61e1499`)** Pass email quoted the base pass — "3 builds … (6 months)" — on a 6-build, 12-month pass. Now computes from the pass and lists add-ons explicitly. Stray "the" before Discord removed. *Next fulfilment gets the new text; yours is the old one.*
- **C3 (open)** Sales page never mentions the add-ons; they appear only as Stripe `optional_items` on the checkout page. Jenna added storage; Stripe shows **both** add-ons in the session — check the sandbox Payments view for how +3 builds got in.
- **Obs.** Bundle at checkout ($349 + $49 + $29 = $427) fulfilled correctly: pass 6, 6 builds, 12 mo storage, no promo.

## Log excerpts

**Test 2**
```
15:17:59 store: checkout session created kind=pass          POST /api/public/store/checkout 200 361ms
15:18:59 factory pass fulfilled pass_id=7 project=pinstitute/book-001
15:18:59 store: fulfilled amount=34900 promo=""
15:19:00 email sent → production@protocol-institute.org "Your Factory Pass: Obliquities"
```
**Test 1**

```
14:46:08 store: checkout session created kind=pass          POST /api/public/store/checkout 200 448ms
14:48:03 factory pass fulfilled pass_id=6 project=jdixon/book-001
14:48:03 store: fulfilled amount=42700 promo=""
14:48:03 email sent → production@protocol-institute.org "Your Factory Pass: Obliquities"
```
