# Session handoff — 29 Sep 2026 (domains: Phase 1 done, Phase 2 drafted)

Conversation `c4F4BIB`. Plan of record: `docs/NEXT_SESSION_PROMPT_2026-09-29.md`.

## State

- **Punch list 7** live at https://jdbbs.exe.xyz:8766/ (list 6 archived as
  `docs/runs/PUNCHLIST-2026-09-29-list6.md`). Section 8 = Domains + Help;
  help build blocks are **8.5–8.14** (handoff's 8.1–8.10 renumbered).
  8.1 ✓ · 8.2 ✓ · 8.3 ◐ · 8.4 ✓.
- **8.2 Phase 1 done.** `PRODCAL_BASE_URL=https://studio.jdbb.net` in `.env`
  (backup `scratch/env-backup-2026-09-29`); `email_footer` setting updated in
  DB. `srv/canonical_host.go`: 301 public GETs on jdbbs.exe.xyz → studio
  (not `/api/`, `/admin`, `/static/`, `/healthz`, non-GET); `Link rel=canonical`
  elsewhere. Hard-coded hosts replaced in srv, scripts, jdbbs-public.
  Verified: magic link, factory upload (test book 69 on `mcheck`), admin login;
  Jenna confirmed the old-host redirect + admin fallback in her browser.
  The VM cannot reach its own `*.exe.xyz` URL (no hairpin) — test old-host
  behaviour with `curl -H 'Host: jdbbs.exe.xyz' localhost:8000…` or ask Jenna.
- **8.3 Phase 2 drafted.** `jdbbs-public/jdbb-net.html`, preview
  https://studio.jdbb.net/jdbb-net (site_pages 053, draft, unlisted).
  `srv/landing_host.go` (in the handler chain, inert until DNS): jdbb.net/ →
  page, www → apex 301, studio paths → studio 301, anything else → 302 `/`.
  Facts mined to `scratch/landing/FACTS.md`; 5 orange `[ fill ]` placeholders
  wait on Jenna (see 8.3 note). Nav-convergence test exempts `jdbb-net.html`.
- **8.4 blyg.** blyg.jdbb.net live on `jd-blyg` (Jenna did CNAME + domain add).
  Decided: publishing credential = **blyg API key in an exe.dev integration**
  (Jenna took the recommendation). Not built yet.
- **runpage fix:** click ☐/◐ → ☑, ☑ → ☐; alt/shift-click = ◐; ticked items stay
  in place until reload (was: ◐ → ☐ first, and items jumped into the done fold).

## Commits

- prodcal: `4e024d0` archive list 6 · `982fd1c` studio.jdbb.net switch ·
  `ee8837c` runpage ticks · `b6c3005` list-7 export · landing-host commit +
  this handoff (see `git log`). Deployed build `0929.b6c3005` + landing host.
- jdbbs-public: `c72cdfb` hosts → studio.jdbb.net · landing draft commit.

## Next actions (in order)

1. **8.3:** fold Jenna's answers into `jdbb-net.html` (edit on disk = live;
   no rebuild). On her copy approval → give her the cutover steps (in the 8.3
   note); after `domain add`, verify `curl -sI https://jdbb.net/` and
   `https://www.jdbb.net/` and flip site_pages `/jdbb-net` status to live.
   **Never touch MX.**
2. **8.4 follow-up:** check whether Blygger Studio has an API-key/token auth
   (`https://jd-blyg.int.exe.xyz/` via the integration); if not, write a brief
   for jd-blyg's Shelley to add one. Jenna wraps the key in an integration.
3. Stripe dashboard business website → studio.jdbb.net (Jenna, optional).
4. Then **8.5 Help 1** per the plan.

Metrics: context ≈ 50 % at handoff; files read in full: 2 (press.html via
`command cat`, FACTS.md).
