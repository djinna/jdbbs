# Next session — domains, jdbb.net landing, blyg, then Help (2026-09-29)

Start by reading `AGENTS.md`, `docs/CONTEXT-HYGIENE.md`, then:

- `docs/reviews/HELP-SYSTEM-SCOUT-2026-09-29.md` — Mintlify scout + the 13-point "build our own" list
- `docs/reviews/CUSTOM-DOMAIN-PLAN-2026-09-29.md` — domain facts, switch steps, redirect rules

Work from the punch list: add **section 8 · Domains + Help** to
`scratch/run/CHECKLIST.md` with the rows below; restart the `runpage` tmux
session with the new conversation id (`RUNPAGE_CHAT_CONV`). Build in order;
don't start phase N+1 before N is checked by Jenna unless it's independent.

## Phase 0 — Jenna (DNS + access)

- Cloudflare (jdbb.net zone): CNAME `studio` → `jdbbs.exe.xyz`, **DNS only**.
- From her machine: `ssh exe.dev domain add jdbbs studio.jdbb.net`.
  (This VM's exe.dev key, label `jdbbs-shelley`, is scoped: `ls` shows only
  `yc-agent`, and `domain` is "not allowed by SSH key permissions". So Jenna
  runs `domain add`, not the agent.)
- Blyg VM: Jenna tells us its name. For agent access either tag it with the tag
  this key is scoped to (check `ssh exe.dev ssh-key list` on her machine) or
  run that VM's own Shelley with a pasted brief.
- Verify: `dig +short studio.jdbb.net`, `curl -sI https://studio.jdbb.net/`.

## Phase 1 — Switch app to studio.jdbb.net

Follow CUSTOM-DOMAIN-PLAN steps 3–6. Key checks: client magic-link login,
factory upload, and **whether exe.dev admin login works on the custom host**
(undocumented). If not, admin stays on `jdbbs.exe.xyz` and admin-login links
in emails get a separate admin base URL. `PRODCAL_BASE_URL` in `.env`;
replace the ~15 hard-coded `jdbbs.exe.xyz` strings (srv, scripts,
jdbbs-public); `rel=canonical`; 301 public GETs from exe.xyz, never `/api/`
or admin. Mail stays `factory@mail.jdbb.studio`.

## Phase 2 — Fresh jdbb.net landing page (replaces the Blot stub)

Blot blog at the apex is a stub; posts are not meaningful (Jenna, 29 Sep).
MX → Google: **never touch MX**.

- Draft first, on `studio.jdbb.net` or a preview route; Jenna approves copy
  before any DNS change. Page is a public doc (jdbbs-public, theme.css,
  1240 shell, Terminal Folio). Content: Jenna's book publishing + production
  history. Sources to mine: `/press`, `/field-notes`, `/litmags`, `/factory`,
  `/word-free`, docs/brand, session handoffs. Ask Jenna for facts you can't source.
  Links out: Studio (studio.jdbb.net), Blyg, contact.
- Serving: prodcal serves it by Host (`jdbb.net`, `www.jdbb.net`) — small
  host switch in the router, add a site_pages row.
- Cutover (Jenna, Cloudflare): apex CNAME `@` → `jdbbs.exe.xyz` DNS only +
  DNS → Settings → "Flatten CNAME at root"; `www` CNAME same; then
  `domain add jdbbs jdbb.net` and `… www.jdbb.net`. Remove Blot records only
  at cutover.

## Phase 3 — blyg.jdbb.net (Blygger, Book Factory news)

Blygger (blygger.org) is a protocol: a blyg = a directory of static files +
edit history + RSS, changelog-native ("differential feed"). Reference client
**Blygger Studio** is a Cloudflare Worker (`npm run init` / `deploy`) and
deploys to `blyg.<domain>` with `/studio` for writing — jdbb.net is already
on Cloudflare, so that path needs no VM. Jenna has started one on a separate
exe VM. **Decide first:** repurpose the VM one (CNAME `blyg` →
`<vm>.exe.xyz`, `domain add <vm> blyg.jdbb.net`) or move to the Worker.
Pre-1.0 the wire format changes — keep it upgraded (`npm run upgrade`).
Idea to park: the Help "What's new" changelog is exactly a differential feed;
later it could publish as blyg fragments.

## Phase 4 — Help system (punch list 8.x)

Audience order: DIY clients in the app → API users → Jenna + future
collaborators. Only Jenna edits (assistants later). Content public
("garage door open"); admin-tier help and drafts stay admin-only; no real
client names/data — examples use "Sample Press".

Decisions (29 Sep):

1. `/help/` on the app host; studio voice, "you", non-technical.
2. `/factory/api` becomes the Help API section (one source); `/word-free`,
   `/field-guide`, `/factory/terms` stay, linked.
3. **Vocabulary: never "Word" — say "your editor"** (the file is still a
   `.docx`). Add a copy lint/test over client-facing help + UI strings.
4. DIY slice: portal → factory (transmittal lives in it) → **Inspect** →
   builds/downloads → style sheet. Calendar excluded (see open question).
5. **Inspect findings first**: one help entry per finding type (why flagged,
   how to fix in any editor), linked from the finding. Most important content.
6. "?" per page + "learn more" beside findings/fields; slide-out panel later.
7. Self-update runs on each deploy + weekly staleness sweep.
8. Drafts → `/admin/help/` review (diff, Approve/Edit/Reject) + a Section 0
   line + **email to Jenna linking the approval page** (new email pathway:
   read and update `srv/EMAIL_SYSTEM.md`).
9. Agent may auto-publish **typo fixes and broken-link fixes** (logged, shown in
   the daily report). Each approval has a tickbox **"You can fix things like
   this yourself from here on"** → stored standing permission (category +
   example), listed and revocable in `/admin/help/`; agent consults it.
10. Test: every client/public route in site_pages has a help page or explicit
    opt-out (like `nav_convergence_test.go`). Add the missing transmittal row.
11. Nits: anyone, no login, honeypot + rate limit; page, selected text,
    comment, optional email.
12. One "Report a nit" for help errors and app bugs → Section 0, tagged.
13. Optional "fixed, thanks" reply through the email-consent guard.
14. Public "What's new"; agent drafts from commits, Jenna approves. First
    entries: the building of the help system itself.
15. **2–3 screenshots** only (suggest: factory upload, an Inspect finding with
    its fix, downloads), Sample Press demo, scripted capture so they regenerate.
16. Stats in our DB: views, thumbs, searches incl. zero-result.
17. **AI Q&A**: shown to 50% of visitors (sticky per visitor) as an A/B;
    grounded only in help pages, cites them, "don't know" → offer a nit.
    Log every question. **Kill switch**: admin toggle + automatic
    circuit breaker (per-IP rate limit, global daily cap) that hides it.
    LLM via `https://llm.int.exe.xyz`.
    **One daily usage email to Jenna** covering 16 + 17 + nits + pending drafts
    + auto-fixes.
18. Editor role for assistants: deferred.

Build order: 8.1 files + render in shell + tiers + nav manifest from
site_pages + FTS search + `llms.txt`/`.md` → 8.2 "?" links + route test →
8.3 nits → 8.4 Inspect content slice + vocabulary lint → 8.5 self-update loop +
`/admin/help/` + approval email + standing permissions → 8.6 What's new →
8.7 stats + daily email → 8.8 Q&A A/B + kill switch → 8.9 API section →
8.10 screenshots. Walk both sides (admin + customer) in the browser before
closing each block.

## Open questions for Jenna

- **Project calendar for DIY clients.** Agent's view: don't sell it as a $100
  add-on (a paid add-on is a support promise = handholding). Keep it for
  studio / full-service clients; hide it from DIY factory clients by default,
  no help pages in v1. Revisit if DIY clients ask for it.
- Blyg: repurpose VM vs Blygger Studio Worker?
- Landing page: facts/history she wants told; any titles or clients she
  doesn't want named.
