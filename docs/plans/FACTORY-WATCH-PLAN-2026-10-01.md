# Factory watch — "teammate, not tool" (plan, 2026-10-01)

Source idea (Jenna's punch-list inbox, 1 Oct): *don't push a release without
a thread that watches it, reads the logs and writes a doc. Worst case you
learn what really happens in production; best case you get a hot fix
automatically.* Applied to the Book Factory: every deploy gets a watcher, the
factory floor gets a watcher, and each alert goes to the right one of us.
**Plan only. Nothing built yet.** Punch list section 9.

## 1. What we already have (assessed 1 Oct)

| Signal | Goes to | How | Notes |
|---|---|---|---|
| Transmittal edited by a client (#6) | **Jenna** | email, max 1 per project per 30 min | hard-coded j@djinna.com |
| Every build, proof/final/failed (#9 QC) | **Jenna** | email per build | still ON in `.env`; was meant for workshop week only (0.31) |
| Store go-live note | **Jenna** | email | broken: never sent (0.31) |
| Store sale / Stripe webhook | ? | none found | verify: does a sale alert anyone? |
| Factory Floor `/admin/factory/` | **Jenna** | live board + feed, pull | she has to open it |
| `factory-tail` (tmux, → `scratch/run/factory.log`) | **me** | log file | passive: I only see it when I read it |
| `journalctl -u prodcal` WARN/ERROR | **me** | log | passive, same |
| Punch-list notes, Section 0 items | **both** | page + push to my chat | the only thing that wakes me |
| Report a nit (8.7) | **both** | Section 0 + chat | new 29 Sep |
| Customer mail (#7, #8) | the customer | email | transactional; no copy to either of us |

**Gap:** nothing watches a deploy, and nothing tells me about a production
error or a stuck customer unless you tell me. You get one email per build,
which is noisy, and nothing when something actually breaks.

## 2. The plan

1. **Routing table (decide first).**
   - **Me**, immediately (chat push): everything: errors, failed builds,
     stuck customers, deploy regressions. I triage.
   - **You**, immediately (email): only P1s. That means the site is down, a
     payment or fulfilment failed, a customer's *final* failed twice, or I'm
     about to roll back.
   - **You**, daily: one digest that folds in the per-build QC mail and the 8.11 usage
     email: builds, failures and what I did about them, nits, stuck customers.
   - **Both** (Section 0): anything needing a decision or a reply to a customer.
2. **Deploy watch.** After every `systemctl restart`, a watcher (subagent or
   scheduled wake-up) runs for ~60 min. It smoke-tests key routes (portal,
   factory, help, admin, a proof build on the Mike Check book), tails journal
   + factory events, and compares error rates with the hour before. It writes
   `docs/runs/DEPLOY-YYYY-MM-DD-<build>.md` and posts a one-line verdict on
   the punch list. On a regression it fixes forward, or proposes a rollback to
   the last checkpoint tag. **Rollback needs your OK.**
3. **Floor watch (watching what customers do).** A rule pass over
   `factory_events` every few minutes. It looks for: repeated build failures,
   upload rejected, Inspect "worth fixing" with no new upload for a day,
   sign-in links denied repeatedly, an expired pass being used, a build stuck
   in `converting`. Each hit becomes a `[watch]` Section 0 row and a push to my
   chat with the log context attached. I diagnose, then draft a customer reply
   for you or a fix.
4. **Automatic hot fixes, within limits.** I fix things myself only in
   categories you've allowed (same standing-permission store as Help 8.9:
   e.g. copy typos, a wrong error message, a missing redirect). Every one is
   logged in the deploy doc and the daily digest. Anything else waits for you.
5. **Server-side alert hook.** One `alert(level, text)` helper in prodcal
   with the same pipe as nits (runpage `/add` → my chat). P1s also email you.
   The QC email becomes a digest line, closing 0.31.
6. **Outside check (optional).** The VM can't reach its own public URL, so
   add a tiny external uptime ping (another exe VM or a free monitor) on
   `studio.jdbb.net/healthz` that emails you on downtime.

## 3. Open questions for Jenna

- Is the P1 list right? Anything else you want emailed immediately?
- Daily digest time (HKT morning?) and one email or two?
- Which hot-fix categories may I do without asking, to start?
- Keep the per-build QC email until the digest exists, or turn it off now?
