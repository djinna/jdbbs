# Orchestration retro — the Factory Pass fan-out (2026-09-03)

Four subagents in parallel: `backend-passes`, `factory-ui`, `tracker-codes`,
`offer-page`. Two finished in ~15 min (tracker, offer page). Two ran ~90+ min
each (backend, UI). All output was good; the problem was **duration and
opacity**, not quality.

## What went long, and why

| Agent | Brief said | Agent produced | Root cause |
|---|---|---|---|
| backend-passes | "favor simple, tested… minimum" — but listed ~8 handlers, 2 emails, gating on 5 endpoints, migration, sqlc, admin JOIN, EMAIL_SYSTEM doc, *and* a 9-case test list | 1,068-line passes.go, **1,354-line test file (30 tests)**, 6 commits | The brief was a *whole feature*, not a task. "Simple" is unenforceable; a concrete test list is a floor, not a ceiling. Then I appended scope mid-flight (EPUB chaining, /factory route). |
| factory-ui | 7 page sections, login flow, polling, 402/409 handling, outputs history, screenshots on 5 states | 1,122-line factory.js + mock server + 5 screenshots + a notes doc | Same: a whole page with every state enumerated. No line/time budget. It also built a mock API to screenshot "populated" state — nice, unasked. |
| tracker-codes (glm) | 4 numbered edits to one file | 8 insertions | Tightly scoped, single file, small model. ✔ |
| offer-page (opus) | 1 file, mirror an existing file, 8 sections | 1 file | Single file with a template to copy. ✔ |

Pattern: **duration ∝ number of distinct endpoints/states in the brief**, not
difficulty. Big-model agents given an open-ended "feature" brief will expand
to fill it with tests, docs, mocks and edge cases — all defensible, all slow.

## Rules for the next fan-out

1. **Budget in the brief, explicitly.** "≤ 300 lines of new Go, ≤ 6 tests,
   stop and report at 25 minutes even if unfinished." Agents honour concrete
   numbers; they don't honour "simple."
2. **One agent = one file (or one vertical slice ≤ 3 files).** Split
   backend into: (a) migration + sqlc, (b) redeem + fulfill + email, (c) gating
   + ledger, (d) admin endpoints. Chain them (each takes the prior's commit)
   or run (b)(c)(d) in parallel once (a) lands. Each is a 15-min job.
3. **Contract first, then fan** — this worked (docs/specs/…). Keep doing it;
   it's what let 4 agents build against unimplemented endpoints without drift.
4. **Never append scope to a running agent.** Queue it as a new, small agent
   after the first reports. (I sent the EPUB-chaining + /factory route to
   backend-passes mid-run; that added ~30 min and made the final commit
   heterogeneous.)
5. **Disjoint `git add` paths, stated in the brief.** Two agents staging
   concurrently in one repo swept each other's files into commits
   (`0c42445`, `fac9a28` carry UI files under backend messages). Tell each
   agent: "`git add` only your listed paths; never `git add -A`."
6. **Poll with short timeouts and read the progress summary.** The 600-s
   status check showed backend had only finished the migration after ~40
   min — that was the moment to split, not wait.
7. **Model to task:** glm/kimi-class for single-file edits with an exact
   spec; opus-class for pages that need taste or for anything touching auth.
   Don't hand opus a one-file edit — it will over-deliver.
8. **Tests: name the cases and say "these and no more."** The floor becomes
   the ceiling only if you say so.
9. **Ask for a smoke, not a showcase.** "One screenshot of the empty state"
   vs. "screenshots" → the UI agent built a mock API and shot five states.

## Would I split differently now?

Yes — 7 agents, ~15 min each, two waves:

- Wave 1 (parallel): migration+sqlc · offer page · tracker · factory.html+css shell
- Wave 2 (parallel, after wave 1 commit): redeem/fulfill/email · gating/ledger ·
  admin endpoints · factory.js
- Wave 3 (me): route wiring, EPUB-in-build decision, smoke test.

Same wall-clock as this run at best, but each piece reviewable, no
mid-flight scope changes, and no 90-minute black boxes.
