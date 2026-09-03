# HANDOFF — Storefront planning for the jdbbs studio (2026-09-03)

**Status:** planning only. Nothing built, nothing committed beyond this note.

## How we got here (including a wrong turn)

User asked: *"steelman/critique using this tool (or its patterns) to create a
storefront for the jdbbs studio."* The link identifying "this tool" was
accidentally omitted, and the agent **guessed** it meant ProdCal itself
(Go + SQLite + pi-public pages + workshop-registration pattern) and wrote a
full analysis on that premise. Lesson recorded: when the referent of "this
tool" is ambiguous, **stop and ask**.

The intended reference is:

- **One Shot** — https://shiponeshot.com/
- Specifics page the user flagged: https://shiponeshot.com/codex#included

## What One Shot actually is (verified 2026-09-03 by fetching the site)

A commercial **starter kit / boilerplate designed for coding agents**, by
@wbnns (Binns Pte. Ltd.):

- **Included:** passwordless auth (Google / Apple / email OTP), Stripe
  payments (subscriptions + one-off), transactional email, iPhone/Android
  shells (Hotwire Native), multi-user teams/tenancy, privacy tooling
  (export/delete), rate limits, one-command deploy, optional crypto module.
- **Stack:** Rails-flavored — Turbo + Stimulus asset pipeline visible on the
  site; deploys with **Kamal** (self-host, no PaaS). *Not* Go.
- **Agent-first design:** ships `CLAUDE.md` / `AGENTS.md` conventions, its own
  test suite, per-agent landing pages (/codex, /claude, /cursor, /hermes,
  /openclaw). Pitch: agent spends tokens on your product, not on rebuilding
  auth/payments/email.
- **Pricing/license:** $49/mo or $299/yr subscription → private repo access +
  updates while subscribed; copies already downloaded keep working forever.
  License: build/sell unlimited apps, 100% yours; only restriction is
  redistributing the kit itself. 14-day refund.
- **Docs cover:** billing & entitlements (gating features behind payment),
  adapter+fake pattern for integrations, content/SEO/blog, BYO keys,
  upgrading (pulling kit updates into your fork).

Saved raw fetches (ephemeral): `/tmp/sos-home.html`, `/tmp/sos-codex.html`,
`/tmp/sos-license.html`, `/tmp/sos-docs.html` — refetch if gone.

## The actual question to answer next session

Steelman + critique of **buying/using One Shot (or copying its patterns) to
build a storefront for the jdbbs studio**, i.e. roughly three options to weigh:

1. **Buy One Shot** ($299/yr) and let an agent build the store on it —
   accepting a second stack (Rails + Kamal) alongside the Go/SQLite ProdCal
   world, on new hosting.
2. **Copy its patterns, not its code** — the interesting part may be the
   *shape*: agent-readable conventions, auth/payments/email as pre-solved
   "essentials," entitlement gating, tests the agent runs on itself. ProdCal
   already has some of this DNA (AGENTS.md, embedded tests, AgentMail).
3. **Neither** — prior session's (mis-aimed but still relevant) fallback:
   pi-public page(s) + Stripe Payment Links / merchant-of-record ships a
   credible small-press storefront with ~zero new code. See the 2026-09-01
   conversation (slug `steelman-critique-studio-storefront`) for that full
   steelman/critique; its *Phase 2* idea (Stripe webhook → tokenized EPUB
   download from the actual production pipeline) is the one custom thing no
   platform replicates.

### Questions to grill the user on before analyzing

- What is the storefront actually selling? Books (EPUB/print)? Workshop
  seats/subscriptions? Studio services? (Determines whether One Shot's
  subscription machinery is even relevant — it's SaaS-shaped, not
  catalog-shaped.)
- Appetite for a second stack + second host (Rails/Kamal vs. everything-on-
  the-jdbbs-VM)? Who maintains it?
- Custom domain plans? (Affects both paths; jdbbs.exe.xyz vs. a store domain.)
- Is "mobile app shells" or "teams/tenancy" ever wanted, or dead weight?
- Is part of the interest actually *studying One Shot as a product model*
  (agent-first starter kit, docs-as-marketing, per-agent landing pages) rather
  than using it? The studio teaches "protocolize your book" — One Shot is a
  protocolized codebase; there may be a meta-lesson worth more than the kit.

## Pointers

- Prior (mis-premised) analysis: Shelley conversation `cQYGD7H`
  (`steelman-critique-studio-storefront`), 2026-09-01→09-03.
- pi-public pattern: `/home/exedev/pi-public/README.md`, `servePublicDoc` in
  `srv/server.go`.
- Workshop registration flow (the existing proto-checkout): `srv/registration.go`.
- Email pathways: `srv/EMAIL_SYSTEM.md`.
