# Next session — 2026-09-25 (security closeout: owner items + recovery bundle)

Read first: `docs/reviews/SESSION-HANDOFF-2026-09-24.md` (all addenda, latest
is the 25 Sep morning one), `docs/reviews/SECURITY-CLOSEOUT-PLAN-2026-09-24.md`,
then on the VM the private `scratch/security-2026-09-24/CONSOLIDATED-REPORT-2026-09-24.md`
and punch list `scratch/run/CHECKLIST.md` §7 (private; never export).

Start-of-session chores:
1. Restart the runpage with this conversation id:
   `tmux kill-session -t runpage; tmux new-session -d -s runpage "cd /home/exedev/prodcal && RUNPAGE_CHAT_CONV=<id> scripts/run-page.sh"`.
2. Check last night's 03:00 backup (`tail ~/backups/backup.log`,
   `cat ~/backups/.LAST-R2-SUCCESS`) and that `rclone lsl r2:jdbbs-backups/db`
   still shows nothing newer than 24 Sep.
3. If Jenna has confirmed the old AgentMail key is deleted: shred
   `~/prodcal/.env.pre-rotate-20260924` and `~/.config/rclone/rclone.conf.pre-rotate-20260924`.

7.16 (app findings) is closed. Agent-side work remaining, in order:
- 7.7 path-only inventory of tracked confidential material (manuscripts,
  customer notes/screenshots, embedded fonts, incl. history) for Jenna's decision.
- 7.5 encrypted recovery bundle (age recipient on VM, Jenna holds the key) →
  7.11 isolated full restore drill (no live email/payment/LLM/callbacks).
- 7.8 prepare repo-scoped deploy keys; 7.9 CI checks (go build/vet/test, shellcheck).
- Pin down the rclone attempt-1 `501 NotImplemented` oddity (upgrade rclone?).

Owner-side items to nudge: 7.12 second-account `/admin/` denial; 7.13 old
AgentMail key deleted + R2 token scope confirmed; `.prodcal-secret` rotation
window; 7.8 register keys; 7.10 retention/lock policy.
