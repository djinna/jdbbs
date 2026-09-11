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
1. ~~`/stylesheet` split~~ — DONE (commit after 8e07125): `/stylesheet-pi/` tool, `/stylesheet/` public house sheet (62 items, fiction/nonfiction/both, `srv/housestyle.go`).
2. **Admin projects list: pagination + filters** (status, client, search). Best-guess UX.
3. Push both repos to GitHub if not done (VM: `git push git@github.com:djinna/jdbbs.git main`; pi-public has its own remote — check `git remote -v`).

## Gotchas learned
- Browser tool screenshots lag a frame; verify via DOM eval before trusting a "blank" shot. Admin pages need the header: `~/sitemap-review/proxy.py` (loopback :8104, run in tmux) injects it.
- `sudo systemctl restart prodcal` after `make build`; pi-public edits are live immediately.
- Deploy freeze recommended Sep 19–22.
