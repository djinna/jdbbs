---
title: Maintaining help
summary: How the help pages work, where they live, and how a page goes live.
audience: studio
visibility: admin
status: live
covers:
  - srv/help.go
  - docs/help/README.md
last_verified: efed34a
owner: jenna
order: 10
---

Only you see this page.

- **Where:** `docs/help/*.md` in the prodcal repo, one file per page; the
  file name is the URL. Edits are live within seconds (no rebuild).
- **Going live:** new pages start with `status: draft` and only you can see
  them, with a *draft* badge. Change it to `status: live` to publish.
  `visibility: admin` keeps a page yours for good (like this one).
- **What's not covered yet:** the bottom of [/help/](/help/) lists client
  routes and public tools that no help page explains.
- **Agents:** `/help/llms.txt`, and any page as markdown by adding `.md`.
- **Search** covers title, summary and body, with stemming (*footnote* finds
  *footnotes*). Drafts and admin pages turn up only in your searches.

Conventions: `docs/help/README.md`. Coming next (punch list 8.6 on): a "?"
on every page, Report a nit, one page per Inspect finding, the review queue
at `/admin/help/`, What's new, stats, and the Q&A trial.
