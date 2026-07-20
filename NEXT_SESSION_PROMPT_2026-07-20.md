# Session handoff — PI Editorial Stylesheet: review the theme/ledger/scope work

Work done 2026-07-20 in `prodcal` (commit `e8781cd`) + `book-production`
(commit `3a1764f`). Service is live on port 8000 behind the exe.dev proxy.
**Both commits are LOCAL only — not yet pushed to GitHub.**

---

## ⭐ START HERE — live links to review

All under `/stylesheet/` (the pages are NOT at the site root; `index.html` /
`authors.html` are just internal filenames served under that path).

- **Editor tool:**            https://jdbbs.exe.xyz/stylesheet/
- **Authors' view:**          https://jdbbs.exe.xyz/stylesheet/authors

Exports (regenerated from the DB, which is canonical):
- Working sheet:              https://jdbbs.exe.xyz/stylesheet/export.html
- **Master (universal only):** https://jdbbs.exe.xyz/stylesheet/export.html?scope=universal
- **Zoothesia only:**          https://jdbbs.exe.xyz/stylesheet/export.html?scope=zoothesia
- Authors' sheet (HTML):      https://jdbbs.exe.xyz/stylesheet/authors.html
- Markdown variants: swap `.html` → `.md` on any of the above.

> Note on scopes today: the DB was reset to a clean slate (all 134 items
> `proposed`, all `universal`). Exports only include **accepted** items, so the
> master/zoothesia export links will look sparse until the team accepts + tags
> items. To *see* the scope split visually right now, open the editor and use
> the **Universal / Zoothesia** filter, or add `?demo=1` (below).

Want to preview the full UI (accepted items, a pending-edit diff, both scopes,
star toggles) without touching real data? Append `?demo=1`:
- https://jdbbs.exe.xyz/stylesheet/?demo=1
- https://jdbbs.exe.xyz/stylesheet/authors?demo=1

In the top-right of every page: the shared font switcher (8 fonts) + dark
toggle — flip dark mode and change the body font to confirm it matches the rest
of jdbbs (admin, transmittal, landing).

---

## What to look for (the three things that changed)

### 1. Shared chrome + ledger redesign
- The page now uses the site-wide `theme.css`/`theme.js`: same `[jdbb] studio`
  masthead, statusline, upper-right font+dark control as the rest of jdbbs.
- Cards are gone. Items are hairline-separated rows; tags are mono bracketed
  text (`[rule]`, `[accepted]`); actions are underlined text-links with a single
  filled button; diffs use house green (added) / red strikethrough (removed).
- Verify: light AND dark, and that switching the font actually changes body type.

### 2. Per-item scope (universal vs zoothesia)
- Each editor item has a scope toggle: **universal · zoothesia**
  (universal = goes to the org-wide master sheet; zoothesia = stays with this
  title's archive). Read-only users see the scope as plain text.
- Toolbar has an **All scopes / Universal / Zoothesia** filter.
- Exports honor `?scope=universal` / `?scope=zoothesia` (see links above),
  with matching titles + download filenames (`pi-master-stylesheet.*`,
  `pi-zoothesia-stylesheet.*`).
- Backing: migration `020-stylesheet-scope.sql`; API `PATCH
  /api/stylesheet/items/{id}` now accepts `scope` (validated, write-gated).

### 3. Process playbook (book-production repo)
- `/home/exedev/book-production/docs/EDITORIAL-STYLESHEET-PROCESS.md` — the
  end-to-end copyedit → review-tool → org master sheet pipeline, the review
  tool reference (URLs, allowlist, seeding, exports), and the universal-vs-title
  scope model with the title #2/#3 workflow.

---

## Files touched
- prodcal `e8781cd`:
  `db/migrations/020-stylesheet-scope.sql`, `srv/stylesheet.go`,
  `srv/static/stylesheet/{index.html,authors.html,style.css,app.js}`,
  `docs-stylesheet-SPEC.md`
- book-production `3a1764f`:
  `docs/EDITORIAL-STYLESHEET-PROCESS.md`

## State
- `prodcal` tree clean; `prodcal.service` active on 8000; `go build`/`vet`/
  `gofmt`/`go test ./srv/...` all green.
- DB reset to clean slate: 134 items `proposed`, all `universal`, 0 pending edits.
- Migrations through 020 applied.

## Open items / next steps
- **Push commits** (hub-and-spoke): from your local machine or over SSH from the
  VM — `prodcal` `e8781cd` → GitHub, then the VM fast-forwards; `book-production`
  `3a1764f` is on branch `prep/test-run-2026-04-05`. Nothing is pushed yet.
- Multi-title seeding: the current seeder replaces the working set. If titles
  must coexist in one DB, extend `db/seed/parse_stylesheet.py` to stamp a
  per-title scope and seed additively (noted in the playbook §3).
- No exe.dev share/CLI changes needed — everything is on the default port.
