# Session handoff — page normalization shipped; stylesheet split + admin pagination next (2026-09-11)

**Deployed to VM.** Workshop Sep 21–22 (10 days). 8 real registrations + 1 faux (Mike Check).

## Read first
- `docs/PAGE-DESIGN-HOSTING-VISIBILITY-2026-09-11.md` — the page standard (anatomy, 1240 shell, wordmark, chrome, visibility tiers + recipes, inventory).
- `docs/IDEAS.md` — parked ideas.
- `docs/reviews/ORCHESTRATION-RETRO-2026-09-03.md` — fan-out rules (worked well today: 4 agents, one file-set each, no commits by agents).
- Decisions + your answers live in `~/sitemap-review/answers.json` (tool at https://jdbbs.exe.xyz:8102/, `sitemap-review.service`).

## Done this session (commits 5087c68, 451bc81, 63710a5 in prodcal; cd37546…42abad8 in pi-public)
- theme.css/js: `--shell-max` 1240, `.jdbb-shell/.jdbb-prose/.jdbb-footer`, centralized wordmark kerning, sans-only random first face (session-sticky), auto-mount `#theme-bar`, lighter bg.
- Every page normalized (embedded + pi-public + generated companions via `client-raw/cleanup.py`). Inline theme copies in pi-public deleted. `/lg` → 301 `/`. Homepage +15% non-heading type.
- Cohort roster `/2026-pi-symposium` (srv/cohort.go, cohort.html, migration 024 `clients.cohort_slug`, set on Factory Pass redemption). Test `TestCohortRosterGate`.
- Admin **Pages** tab (site_pages registry, `/api/admin/pages` CRUD; srv/sitepages.go).
- Zoom → Discord in registration.go + workshop.html.
- Mike Check smoke account: registration #10, `bookiq@gmail.com`, client `mike-check`, project `smoke-test-manuscript`, coupon PYB-PFSA-R56E redeemed, cohort set, password reset via admin endpoint (emailed). **Uncheck him before real announcements; delete after Sep 22.**

## Next (user-confirmed, in order)

### DONE 2026-09-11 late: outbound-email log + admin Mail report
Migration 027 `outbound_email`; `Server.mail(meta, …)` is the single logging
sender (all 11 call sites migrated; `EmailConfig.sendEmail` is now only a thin
shim — new code must call `s.mail`). `GET /api/admin/email` with filters;
Admin › Mail tab; tracker cards show "Mail: N sent · code announcement ×1 ·
Factory Pass delivered · last …" and a recipient-level "Sent mail" panel.
Backfilled: 6 Sep 4 announcements (joined event_announcements ↔ coupons),
9 registration confirmations, and today's 5 journal entries (one-off SQL).
Tests: `TestOutboundEmailLog`, `TestRegistrationSendsAreLogged`.

### DONE 2026-09-11 late: email template restyle
All 10 outbound HTML templates now wrap in `emailShell()` (`srv/email_shell.go`:
wordmark line, mono kicker, h1, hairline footer; blocks emailKV/emailTable/
emailStats/emailH2/emailButton/emailCode/emailStatus/emailSignoff/emailList).
No gradients, emoji, pills, or "ProdCal" (From name now "jdbb studio").
Text parts unchanged except emoji/brand. Preview gallery with fixture data:
`/admin/email-preview/` (admin-gated; `?part=text` for the plain part).
Build-delivered mail extracted to `buildDeliveredText/HTML`. Any NEW email
template must use emailShell and be added to `emailPreviewFixtures`.

### DONE 2026-09-13: BCC audit copy + permanent smoke persona
- Every send (batch or single) is BCC'd to `PRODCAL_MAIL_BCC` (default
  j@djinna.com), skipped when that address is already To/Cc. Recorded in
  `outbound_email.bcc_addrs` (migration 029), shown in Admin › Mail.
- **Mike Check** (`bookiq@gmail.com`, reg #10, client `mike-check`, project 17,
  pass 2) is PERMANENT — do not delete, keep `consent_email` on. Announcements
  always include his registration (`withSmokeRegistration`,
  `smokeRegistrationEmail` in `srv/registration.go`). Confirmed: snapshot #23
  to him with BCC landed 200.

### THIRD: stable project slugs (user: "clients may well change book titles… maybe lname-000")
Today `uniqueProjectSlug` (`srv/passes.go` ~L444) derives from the manuscript
title, truncated to 24 chars → `building-in-the-wrong-ma`. Title-derived is
fragile: titles change during production and the slug is in emailed URLs.
Proposal to confirm with user: per-client sequence, `{lastname}-{NNN}` →
`/mike-casey/casey-001/`. Alternatives to mention: `/mike-casey/001/` (shorter,
no redundancy) or `/mike-casey/book-001/`. Implementation: `SELECT COUNT(*)+1
FROM projects WHERE client_slug=?` (or MAX of parsed suffix) → zero-pad 3;
keep `projects.name` as the human title (renamable freely). Apply to Factory
redemption path AND the admin New Project modal (offer as default, editable).
Existing project `mike-casey/building-in-the-wrong-ma`: leave as-is unless
user says rename — the URL is already in Mike Casey's fulfillment email.

0. Everything from the 2026-09-11 workplan is shipped. Remaining items are post-workshop polish (see docs/IDEAS.md) and the deploy freeze Sep 19–22.
1. ~~`/stylesheet` split~~ — DONE (commit after 8e07125): `/stylesheet-pi/` tool, `/stylesheet/` public house sheet (62 items, fiction/nonfiction/both, `srv/housestyle.go`).
2. ~~Admin projects list: pagination + filters~~ — DONE (client-side over cached list; search/client/status/sort/per-page + pager; auto-hidden when everything fits).
3. Push both repos to GitHub if not done (VM: `git push git@github.com:djinna/jdbbs.git main`; pi-public has its own remote — check `git remote -v`).

## Gotchas learned
- Browser tool screenshots lag a frame; verify via DOM eval before trusting a "blank" shot. Admin pages need the header: `~/sitemap-review/proxy.py` (loopback :8104, run in tmux) injects it.
- `sudo systemctl restart prodcal` after `make build`; pi-public edits are live immediately.
- Deploy freeze recommended Sep 19–22.
