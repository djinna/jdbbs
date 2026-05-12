# Prodcal reorientation note — 2026-04-12

Purpose
- Capture the current reorientation research after the interrupted SSH session.
- Give the next session a clean starting point without needing to reconstruct state again.

## Current shell / repo baseline
- Working directory at time of note: `/home/exedev/prodcal`
- Branch state: `main...origin/main [ahead 1, behind 1]`
- Latest local commits:
  - `250bffe` transmittal UX fixes + epub test session checkpoint
  - `fdd987f` fix(transmittal): harden custom styles and epub metadata flow
  - `7faab32` feat(admin): add new project and new client affordances
  - `adf36dd` fix(projects): seed standard workflow for new calendars
  - `a07bc72` fix(transmittal): add 3-state checklist UX, epub isbn, and clearer field states

## Current uncommitted worktree state
Not clean. At the time of reorientation, these files were modified:
- `db/dbgen/models.go`
- `db/dbgen/visitors.sql.go`
- `db/queries/visitors.sql`
- `srv/admin.go`
- `srv/admin_test.go`
- `srv/auth_test.go`
- `srv/bookspecs.go`
- `srv/bookspecs_test.go`
- `srv/client.go`
- `srv/server.go`
- `srv/server_test.go`
- `srv/static/admin.html`
- `srv/static/app.js`
- `srv/static/client.html`
- `srv/static/landing.html`
- `srv/static/notes.html`
- `srv/static/transmittal.js`
- new migration: `db/migrations/011-project-archive.sql`

Implication
- The repo is no longer in a narrow “just deploy the transmittal fix” state.
- Before any deploy/reset/rebase, inspect and separate intentional in-progress work from safe-to-ship work.

## What had already been completed before the interruption
From prior Prodcal sessions, the transmittal work had already reached this state:
- Explicit 3-state checklist/backmatter UX implemented:
  - In manuscript now
  - Coming later
  - Not included
- `isbn_epub` added to book information
- Visible autosave copy added to transmittal UI
- Duplicate/reset logic clears stale checklist statuses and dates
- Frontend compatibility logic preserves behavior for older saved transmittals

## Prior verification already completed
Previously verified in earlier sessions:
- Targeted transmittal tests passed
- Broader transmittal/client tests passed
- Manual QA session progressed through EPUB workflow checks and confirmed:
  - Phase 1: Transmittal Capture — PASS
  - Phase 2: Pull Into Spec — PASS
  - Phase 3: Export Word Template — PASS
  - Phase 4: Prepare Manuscript in Word — pending manual user work at that time
  - Phase 5: EPUB-First Run — pending at that time
  - Phase 6: Print / Typst — pending at that time

## Follow-up fixes/changes already noted in repo docs
From `docs/checklists/2026-04-11-epub-test-session.md`:
- Applied during testing:
  - Page Proofs contact placeholder changed from “Address / Tel” to “Email”
  - Files & Delivery updated to Deliverables, with modernized delivery options
- Queued, not yet implemented there:
  - auto-update `project_slug` when `book.title` changes in transmittal UI

## Important deploy/environment notes
- Embedded static assets require rebuild/restart after JS changes:
  - `make build && sudo systemctl restart srv.service`
- A stale-binary / stale-static issue had previously been observed on this host.
- Historical systemd hazard exists: duplicate units `srv.service` and `prodcal.service` can race for port 8000.
- More recent session note said `srv.service` was the live unit after verification; older notes recorded `prodcal.service` as listener during an earlier conflict. Re-check live listener before assuming.
- Public base URL remains `https://jdbbs.exe.xyz`

## Best next-step framing for a fresh session
Use a phase-gated restart:
1. Inspect current diff by area and identify what changed since the transmittal/EPUB checkpoint.
2. Separate:
   - safe deployable work
   - still-in-progress work
   - anything needing stash/branching/rebase resolution
3. Only then decide whether to build/restart/deploy.

## Suggested first prompt for the next clean session
“Read `docs/notes/2026-04-12-reorientation-and-clean-restart.md`, inspect the current diff by area, and tell me what is safe to commit/deploy versus still in progress.”
