# docs/help — the studio's help pages

Every `*.md` file here (except this README) is one page at
`/help/<file-name-without-.md>`, rendered by `srv/help.go` in the standard
shell. The server re-reads this directory when anything changes: **an edited
page is live within a couple of seconds, no rebuild.** Commit edits like any
other doc.

File names are the URL: lower-case letters, digits and hyphens only
(`send-your-manuscript.md` → `/help/send-your-manuscript`).

## Frontmatter

```yaml
---
title: Send your manuscript           # page title (else the first "# " line)
summary: One sentence, shown in the index, search and llms.txt.   # required by the test
audience: authors                     # authors | api | studio — who it's written for
group: Start here                     # optional index heading; else derived (below)
visibility: public                    # public | admin      (anything else = admin)
status: draft                         # live | draft        (anything else = draft)
routes: ["/{client}/{project}/factory/"]   # site_pages routes this page explains
covers:                               # source files it describes (self-update, 8.9)
  - srv/static/factory.js
last_verified: 8c163d2                # commit the text was last checked against
owner: jenna
order: 20                             # sort within the group (default 100)
---
```

- **Tiers.** `visibility: admin` or `status: draft` → only the admin sees the
  page (404 for everyone else, in the page, index, search, llms.txt and `.md`).
  New pages start as `draft`; Jenna flips them to `live`.
- **Groups.** Index headings come from `group:`, else: `audience: api` →
  *For developers*; admin/studio → *Studio admin*; else the site_pages tier of
  the first route (client → *Your books*, public → *The studio*).
- **routes** must exist in `site_pages` (the test checks). The admin view of
  `/help/` lists client routes and public tools that no page names yet.
- **Links** between pages: write `[Inspect](inspect.md)` — they render as
  `/help/inspect` and also work on GitHub. The test fails on links to pages
  that don't exist.

## Voice

Studio voice, second person ("you"), non-technical, short. **Never "Word" —
say "your editor"** (the file is still a `.docx`). Examples use *Sample Press*,
never a real client. See `docs/NEXT_SESSION_PROMPT_2026-09-29.md` Phase 4.

## For agents

`/help/llms.txt` indexes the live pages; every page is also at
`/help/<slug>.md`. Search: `/help/search?q=` (HTML), `/api/help/search?q=`
(JSON: `{query, results:[{slug,title,summary,snippet,url}]}`).
