# Next session — 2026-09-24 (security review continuation)

Read first: `docs/reviews/SESSION-HANDOFF-2026-09-24.md` (all addenda),
`docs/reviews/SECURITY-CLOSEOUT-PLAN-2026-09-24.md`, then on the VM the private
`scratch/security-2026-09-24/CONSOLIDATED-REPORT-2026-09-24.md` and punch list
`scratch/run/CHECKLIST.md` §7 (private section; never export).

Start-of-session chores:
1. Restart the runpage with this conversation id:
   `tmux kill-session -t runpage; tmux new-session -d -s runpage "cd /home/exedev/prodcal && RUNPAGE_CHAT_CONV=<id> scripts/run-page.sh"`.
2. Check last night's 03:00 backup wrote to the per-host prefix with SHA
   verification (`tail ~/backups/backup.log`, `cat ~/backups/.LAST-R2-SUCCESS`),
   and that nothing new appeared under the old shared `db/` prefix
   (`rclone lsl r2:jdbbs-backups/db | sort -k2 | tail -2`).
3. Confirm `git fetch` from GitHub still works now the repo is private.
4. If Jenna has confirmed the old AgentMail key is deleted: shred
   `~/prodcal/.env.pre-rotate-20260924` and `~/.config/rclone/rclone.conf.pre-rotate-20260924`.

Then continue 7.16 in order: callback DNS pinning (dial-time address check,
`srv/books.go` callback transport) → DOCX zip budgets + exec deadlines/queues
→ duplication policy + build-credit race. Each with a regression test; deploy
only when `books.status='converting'` count is 0; push after each commit.

Owner-side items to nudge: 7.12 second-account `/admin/` denial; 7.13 old
AgentMail key deleted + R2 token scope confirmed; decide `.prodcal-secret`
rotation window; 7.8 GitHub deploy keys.
