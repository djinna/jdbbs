# Help 5 (punch list 8.9): review loop. Design, 1 Oct 2026

Designed on 1 Oct, before the privacy and terms work took the session. Nothing
is built yet. Decisions 7–9 come from `docs/NEXT_SESSION_PROMPT_2026-09-29.md`
(Phase 4).

## Shape

**The server detects, queues, reviews and applies. The agent (Shelley) drafts.**
There is no LLM in the server for this: drafts need repo context, and alerts
already reach the agent's chat through the runpage push. The factory watch
plan (section 9) uses the same model.

1. **Sweep** (`srv/help_review.go`). Runs on startup when HEAD ≠ the last
   swept head (that is, on each deploy), roughly 60 s after boot, plus a weekly
   ticker. For each page with `covers` + `last_verified`: `git log
   <last_verified>..HEAD -- <covers>` (run in the git root of `helpDir()`; skip
   quietly if it isn't a repo, as on the Mac or in tests). Flag a missing covered
   file or an unknown commit. Record the result in `help_sweeps` (trigger, head,
   JSON report). Push one Section 0 line through the nit pipe (`nitInboxURL()`)
   only when the stale set changed, or when the last push is ≥ 7 days old.
   **Live pages go to review. Drafts are the agent's to edit directly**, since
   Jenna reviews a whole draft when she publishes it. Disable with
   `PRODCAL_HELP_SWEEP=off`. Start it from `New()`, not from tests.
2. **Proposals** (`help_proposals`): slug, kind (edit | new | publish | verify),
   category, summary, source (sweep:<head> | nit:<id> | agent | admin),
   base_hash + base_text, proposed (the whole file), verified_at (the commit;
   it becomes `last_verified` on approve), status (pending | approved |
   rejected | auto | conflict | withdrawn), decided_*, note, permission_id,
   commit_hash, notified_at.
   Admin API: `GET/POST /api/admin/help/proposals`,
   `POST …/{id}/approve {content?, grant?:{category,description}}`,
   `POST …/{id}/reject {note}`, `GET /api/admin/help/stale`,
   `POST /api/admin/help/publish {slug}` (flips a draft to live: the 30
   drafts need a button, not a file edit).
3. **Apply**: check base_hash (a mismatch means conflict), write
   tmp → rename, set `last_verified`, then `git add -- f && git commit -m … -- f`
   (that commit mode leaves other staged work alone; the docs editor already
   commits this way, and there's no pre-commit hook on the VM). No push. If git
   fails, keep the file and note it on the proposal.
4. **Standing permissions** (`standing_permissions`: scope help | app, category,
   description, example, granted_by, proposal_id, revoked_at, uses,
   last_used_at). Seed: help/typo and help/broken-link (decision 9). A
   proposal posted with `auto:true` is applied at once only if an active
   permission matches. Typo and broken-link also get mechanical checks (a
   token diff with small edit distance; only hrefs changed). Log it, push a
   Section 0 line, no email. Section 9.5 hot fixes reuse the table
   (scope `app`).
5. **`/admin/help/`** (static `srv/static/help-admin.html` like docs-editor):
   Waiting for you (diff from a Go token diff served by the API; Approve / Edit
   / Reject; tickbox "You can fix things like this yourself from here on"),
   Stale pages, Drafts (Publish), Nits (`/api/admin/nits`), Standing
   permissions (Revoke), Recent log. Needs an ADMIN_NAV entry in
   `srv/static/theme.js`, a site_pages row, an entry in the
   nav_convergence_test admin list, and `routes: ["/admin/help/"]` on
   maintaining-help.
6. **Email pathway #10** (`srv/help_review_email.go`): one notice to Jenna
   (`txNotifyRecipient`), debounced: send once the oldest un-notified pending
   proposal is ≥ 5 min old. It lists them and links `adminLoginURL(base,
   "/admin/help/")`. Kind `help_review`. Add it to email_preview fixtures, and
   update `srv/EMAIL_SYSTEM.md` (the count becomes 10: 5 manual + 5 automatic).

First sweep caveat: every page has `last_verified: efed34a`, and Help 2–4 have
changed covered files since. Expect a noisy first report. Only
maintaining-help is live. Update it in the same commit: it still lists
"/admin/help/" as coming next.
