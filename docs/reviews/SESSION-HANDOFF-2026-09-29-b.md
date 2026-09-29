# Session handoff — 29 Sep 2026 (b): blyg brief, Help 1–4 built

Conversation `cOVVAAY` (runpage tmux restarted with `RUNPAGE_CHAT_CONV=cOVVAAY`).
Plan of record still `docs/NEXT_SESSION_PROMPT_2026-09-29.md` (Phase 4).
Previous: `SESSION-HANDOFF-2026-09-29.md`.

## State

- **8.3 landing (waiting on Jenna).** No answers yet. Filled only Books:
  "About 300, over 30 years" (from her blyg bio). 5 questions + cutover steps
  posted on the 8.3 note. Never touch MX.
- **8.4 follow-up done.** Blygger Studio 0.7.0 has **no** API key (upstream
  `/api` = owner cookie only; stable publishing API is upstream roadmap 1.8).
  Brief for jd-blyg's Shelley: `docs/briefs/JD-BLYG-PUBLISH-TOKEN-2026-09-29.md`
  (token via `X-Blyg-Token`, because the existing jd-blyg integration already
  uses `Authorization`; second integration `blyg-publish` → `https://blyg.jdbb.net`).
  Also flags: blyg origin still `jd-blyg.exe.xyz`, author link → old host.
- **8.5 Help 1 [~]** `srv/help.go`: `/help/`, `/help/{slug}`, `.md` twins,
  `/help/llms.txt`, `/help/search`, `/api/help/search`. docs/help/*.md read
  from disk and reloaded on change (no rebuild). Tiers: `visibility: admin` /
  `status: draft` → admin only everywhere; unknown values fail closed.
  In-memory FTS5 (porter). goldmark added to go.mod. site_pages 054.
  Seed pages (all **draft**): start-here, sign-in, transmittal,
  send-your-manuscript, inspect, build-and-download, factory-pass,
  style-sheet, factory-api; maintaining-help (admin, live).
  Fix: upload box said 100 MB, server enforces 50 MB.
- **8.6 Help 2 [~]** `/api/help/for?path=` + theme.js `helpLinks()`: masthead
  "?" + `<a data-help=slug hidden>` "learn more" (factory steps 2/3/5/6),
  shown only for pages the viewer can open. `helpOptOut` + `TestHelpRouteCoverage`.
- **8.7 Help 3 [~]** Report a nit: `POST /api/nits` (honeypot `website`, 5/IP/10 min,
  200/day), table `nits` (055), pushes `[nit·help|app #id]` to runpage `/add`
  (`PRODCAL_NIT_INBOX_URL`, default 127.0.0.1:8766/add). theme.js footer link +
  `<dialog>`; `GET /api/admin/nits`. Smoke-tested on prod, row removed.
- **8.8 Help 4 [~]** 20 `docs/help/finding-*.md` drafts (subagent-written,
  spot-checked one; `TestHelpEveryFindingHasAPage`), Inspect rows get
  "learn more". `TestVocabularyNeverWord` over help + client UI/messages;
  ~12 strings fixed (export hints now say ".docx" instead of product menus —
  asked Jenna on 8.8 whether to keep literal menu names).

## Commits

prodcal: `efed34a` blyg brief · `351bd0a` Help 1 · `c6da9a6` Help 2 ·
`1ae536e` Help 3 · `1b77c7a` Help 4 · this handoff + punchlist export.
jdbbs-public: `25bf19c` landing book count · factory-api error text.
Deployed: build `0929.1ae536e`+ working tree = `1b77c7a` (restarted after).

## Validation

`go vet`, gofmt, `go test ./...` green on the VM at `1b77c7a`. Browser-checked
help index/page/table, "?" + learn-more on the factory page, nit dialog +
submit, via a preview instance (scratch/help/preview, copied DB, drafts
flipped live, `PRODCAL_NIT_INBOX_URL=off`). The VM browser can't send admin
headers — use a preview instance like that to see drafts.

## Waiting on Jenna

- 8.3 facts + copy approval → cutover (hers, Cloudflare + `domain add`).
- 8.4 paste the brief into jd-blyg's Shelley; then create `blyg-publish` integration.
- 8.5/8.8 read drafts; flip `status: live` (or tell the agent which). Open
  checks on 8.5 note: Copyright style meaning; Inspect "images too small to
  print" claim vs code; 8 UI labels the detector never emits.
- 8.8 menu-name wording.

## Next actions (in order)

1. Fold in any 8.3/8.5/8.8 answers from notes.
2. **8.9 Help 5**: self-update loop (covers/last_verified diff → drafts),
   `/admin/help/` review (diff, Approve/Edit/Reject; list nits from
   `/api/admin/nits`), approval email (read + update `srv/EMAIL_SYSTEM.md`
   first — new pathway), standing permissions. Needs ADMIN_NAV entry +
   site_pages row (nav_convergence_test).
3. 8.10 What's new → 8.11 stats + daily email → 8.12 Q&A → 8.13 API section → 8.14 screenshots.

Metrics: context ≈ 60 % at handoff; files read in full: 1 (scratch/help/FACTS.md, via sed in two halves).
