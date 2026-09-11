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

### FIRST: outbound-email log + admin "Mail" report (user asked for this explicitly, 2026-09-11 late)
Why: user was surprised twice today by mail they couldn't see (Sep 4 code
announcements only exist in a doc; today's transmittal notification). Sends are
only in the systemd journal, which rotates. Every send must be auditable.
- Migration 027: `outbound_email(id, sent_at, to_addrs, cc_addrs, subject, kind, ref_type, ref_id, status_code, error, triggered_by)`.
  `kind` = template name (registration_confirm, registration_admin_alert, announcement, factory_pass, client_password, transmittal_update, snapshot, activity, …); `ref_*` = registration/project/client id; `triggered_by` = admin email header, 'client', 'public', or 'system'.
- Hook point: the single sender in `srv/email.go` (the func around line 60–125 that logs `"email sent"`). Record success AND failure; make `sendEmail` take a small `emailMeta` struct or add a wrapper, then thread kind/ref through the ~8 call sites (`grep -n 'email sent\|s.sendEmail\|SendEmail' srv/*.go`).
- API: `GET /api/admin/email?kind=&to=&since=&limit=` (admin-gated).
- UI: "Mail" tab in `admin.html` (table: when · to · subject · kind · status · ref link) AND a "Sent mail" panel at the bottom of `/admin/registrations` filtered to workshop kinds (announcement, registration_confirm, factory_pass) with a per-registrant column "last mailed / codes sent".
- Backfill: insert the six Sep 4 announcement rows from `docs/reviews/FACTORY-PASS-SESSION-A-DRY-RUN-2026-09-04.md` (kind=announcement, status 200, note "backfilled from doc") so the report is honest about history.
- Test: send path writes a row on success and on AgentMail failure.

Also flagged, not yet acted on: Toby Shorin (reg #8, Sep 10) has NO code; Andrea Leiter (reg #2) has a code never emailed (consent_email=false). Ask user how to handle before Sep 21. Email templates are off-brand (purple gradient, emoji, "ProdCal") — offered restyle, user hasn't answered.

0. Everything from the 2026-09-11 workplan is shipped. Remaining items are post-workshop polish (see docs/IDEAS.md) and the deploy freeze Sep 19–22.
1. ~~`/stylesheet` split~~ — DONE (commit after 8e07125): `/stylesheet-pi/` tool, `/stylesheet/` public house sheet (62 items, fiction/nonfiction/both, `srv/housestyle.go`).
2. ~~Admin projects list: pagination + filters~~ — DONE (client-side over cached list; search/client/status/sort/per-page + pager; auto-hidden when everything fits).
3. Push both repos to GitHub if not done (VM: `git push git@github.com:djinna/jdbbs.git main`; pi-public has its own remote — check `git remote -v`).

## Gotchas learned
- Browser tool screenshots lag a frame; verify via DOM eval before trusting a "blank" shot. Admin pages need the header: `~/sitemap-review/proxy.py` (loopback :8104, run in tmux) injects it.
- `sudo systemctl restart prodcal` after `make build`; pi-public edits are live immediately.
- Deploy freeze recommended Sep 19–22.
