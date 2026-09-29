# Help system scout — Mintlify and friends (2026-09-29)

Scouting only; nothing built. Prompted by Jenna finding Mintlify via
https://docs.twill.ai/overview.

## Mintlify, as of today (official pricing page, checked 29 Sep 2026)

| | Starter (free) | Pro ($450/mo annual, $540 monthly) | Enterprise |
|---|---|---|---|
| Editors | 5 | unlimited | unlimited |
| Hosting, custom domain, web editor, Git sync, search, custom components + CSS/JS | yes | yes | yes |
| llms.txt, per-page `.md`, MCP server ("agent optimizations") | yes | yes | yes |
| Reader auth | Mintlify accounts only | password | OAuth / JWT / SSO |
| AI Assistant (Q&A) | no | 25 credits/answer | same |
| Agent + **Automations** (the self-updating part) | no | 250 credits/update, 10k credits/mo incl. | same |
| Feedback, analytics, webhooks, preview deploys, grammar checks | no | yes | yes |
| Self-hosting / static export | no | no | yes |

Third-party pricing write-ups disagree wildly (old Hobby/$250 Pro tiers are
still floating around); trust the official page and re-check before any decision.

**Verdict:** Starter would host a good-looking public docs site for free, but
the two things that make it interesting to us are paywalled — automations
(self-updating) and AI Q&A — and gating help behind ProdCal client logins
needs JWT auth (Enterprise). It would also live outside our app with its own
React chrome, breaking the one-shell / shared-nav rule (AGENTS.md
"Two sides, one app"). Possible narrow use: a free Starter site for the
*public* Factory API/CLI developer docs. Not for in-app help.

## What the Twill site shows (the pattern worth copying)

- Grouped sidebar (Get started / Customization / Integrations / Account),
  Ctrl-K search, separate *Documentation* and *API reference* tabs.
- Banner telling agents to fetch `/llms.txt`; every page also served as
  `<page>.md` (verified `/overview.md` → 200). Docs are written for humans
  and agents at once.
- Release notes, pricing and security as first-class doc pages.

## Alternatives

- **Hosted:** GitBook (free plan with Git sync), ReadMe (~$250/mo), Fern
  (free plan with AI credits — per Fern's own comparison), Document360, Archbee.
- **Open-source, self-hosted:** Docusaurus (React), Starlight (Astro, Pagefind
  search), Fumadocs (Next.js), MkDocs Material, Scalar (API reference).
  All add a Node/Python build and a second visual system to maintain.
- **Ours:** Go + pandoc (already used by `/admin/runs/`, `renderRunMarkdown`)
  + SQLite. Closest to how the app already works.

## If we built our own — what we'd have to handle

1. **Content as files.** `docs/help/*.md` with frontmatter: `title`,
   `audience` (public/client/admin), `routes` (pages it explains), `covers`
   (source globs it describes), `last_verified` (commit), `owner`.
2. **Nav manifest.** Mintlify's `docs.json` equivalent — or derive groups from
   the `site_pages` registry so help and routes can't drift.
3. **Render in our shell.** theme.css, masthead, 1240, `data-admin-nav` /
   `data-client-nav`. Reuse the runs renderer.
4. **Visibility tiers.** Public / client / admin, enforced server-side with the
   existing auth; client help must never leak admin pages.
5. **Search.** SQLite FTS5 (verify modernc build has it) or a small
   client-side index.
6. **Agent-readable.** `/help/llms.txt` and `/help/<slug>.md`; later maybe an
   MCP tool. Fits the factory API/CLI story.
7. **Contextual "?"** on each page → its help page, with a convergence test
   (like `nav_convergence_test.go`): every route has help or an explicit opt-out.
   Both ends: each feature's admin and customer view.
8. **Feedback loop.** Thumbs / "this is wrong" → punch-list inbox (Section 0).
   We already have that pipe.
9. **Self-updating (the Mintlify "automation" pattern).** Trigger on deploy,
   merge, or schedule → diff since each page's `last_verified` → map changed
   files to pages via `covers` → agent drafts edits → branch + punch-list item
   for Jenna to approve. Never auto-publish. Cheap staleness check without an
   agent: a test that warns when covered files changed after `last_verified`.
   Mintlify's own advice applies: narrow scope, written procedure, explicit paths.
10. **Screenshots rot fastest.** Scripted headless-browser capture per page, or
    avoid screenshots.
11. **What's new / changelog** from commits or punch-list exports.
12. **Web editing ("wiki").** Admin doc editor (`/admin/docs/`) already edits
    and auto-commits `jdbbs-public` HTML; extend to help markdown. Clients
    suggest via feedback, not edit.
13. **AI Q&A (optional, last).** Retrieval over help pages + LLM; per-answer
    cost, needs guardrails and "don't know" behaviour. Defer.

Rough size: 1–6 are a few days (mostly done pieces); 7–8 a day or two;
9 is the real project and the one to pilot on a handful of pages first.

## Open questions for Jenna

- Audience first: clients/authors using the portal+factory, Jenna and
  collaborators on the admin side, or outside developers calling the factory API?
- "Wiki" = who edits: only Jenna + agent, or clients too?
- Public, gated, or both?

## Sources

- https://www.mintlify.com/pricing (read directly 29 Sep 2026)
- https://www.mintlify.com/docs/guides/use-automations.md
- https://docs.twill.ai/llms.txt
- buildwithfern.com, ferndesk.com, gitbook.com comparison posts (vendor-biased)
