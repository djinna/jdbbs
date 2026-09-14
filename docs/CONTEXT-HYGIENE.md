# Context Hygiene

Rules for the agent's own context window while working on jdbb studio. Adapted
2026-09-14 from another project's ORCHESTRATION.md (§6–7) after a compaction
failure here; the multi-agent parts (dispatch packets, waves, polling) were
deliberately left out — this project is one lead agent, one user, mostly
sequential edits to existing files.

Enforced where possible by `scripts/readguard.sh` (sourced from `~/.profile`,
so it applies to every bash-tool call, subagents included). The rest is policy.

## Reads

- **Never `cat` a file over ~350 lines / ~24 KB.** `srv/` has a dozen files
  over 1,000 lines (`server.go`, `passes.go`, `bookspecs.go`, the `_test.go`
  files). Use `sed -n A,Bp`, `rg -n PATTERN FILE` (with `-C` context), or
  `grep -n` then read the hit. The guard refuses `cat`/`less`/`more` on such
  files; bypass with `command cat` or `READGUARD=off` only when the whole file
  is genuinely needed.
- **Pipe by default.** `go test ./... 2>&1 | tail -5`, `journalctl … | tail -40`,
  `git log --oneline -10`, `| wc -l`. Full output only when something failed.
- **Bulk reads go to a cheap model.** Surveying several files, a long log, a DB
  dump, a long doc: `llm_one_shot` or a subagent on `glm-5.2-fireworks` /
  `kimi-k3-fireworks` / `deepseek-v4-pro-fireworks` with the question attached.
  Ask for *structured bullets only, each leading with file:line or a name*.
  Only the bullets enter the lead's context.
- **Never re-read a file you just wrote or patched.** Trust the patch result;
  confirm with `git diff --stat` or one `rg -n` for the changed line.
- **Guard single tool outputs, not just the running total.** Compaction can hit
  well before any percentage gate if one result is huge (a PDF-to-text dump, an
  unfiltered `curl`). Truncate before it returns.
- **Small stays local.** A delegation costs a 10–30 s round trip; under the
  threshold the overhead exceeds the tokens saved.

## Writes

- Prefer `patch` replace over `overwrite` (overwrite puts the file into context
  twice).
- Mechanical generation from a spec plus a reference file (a page in the style
  of an existing one, a migration mirroring a prior one, boilerplate tests) may
  go to a cheap-model subagent: *output only code, no fences, no commentary*.
  Review via `git diff --stat` and targeted reads.
- **Never delegate:** editing an existing file (worker summaries carry no
  reliable line numbers), debugging, architectural decisions, anything touching
  the production SQLite or email sends. In this codebase that is most of the
  work, so delegation is mainly for *reads*.

## Session budget — context, not clock

- Stop starting new work at **~50 % context**. Write the handoff and end the
  session by **~65 %**. Unfinished work moves to a new session; a handoff
  written from compacted context loses the exact file/commit state the next
  session depends on.
- Handoffs live in `docs/reviews/SESSION-HANDOFF-YYYY-MM-DD.md` (addenda for
  same-day continuations). Each must include: goal and current state, commits
  made (both repos), `git status`, validation run and what remains, blockers
  and failed approaches, remaining work as small concrete items, and **the
  exact next action**.
- **Metrics line at the end of every handoff:** approximate context % at
  handoff and the number of files read in full (guard bypasses). If those
  numbers don't fall, the rules aren't working.
- A receiving session verifies `git status` + `git log -1` in both repos
  before doing anything else.

## Standing rules (unchanged)

- Explicit `git add -- path …`; never `git add .`/`-A`.
- Push both repos (`prodcal`, `~/jdbbs-public`) at the end of each work block.
- No list email sends by the agent — the user sends from `/admin/registrations`.
