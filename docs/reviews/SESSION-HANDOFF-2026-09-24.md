# Security and recovery handoff — September 24, 2026

**Status: app hardening is deployed; overall security and complete recovery
are not signed off.** Customers should use their existing app login and
Factory pass, not an infrastructure share. No customer passwords, passes,
VM shares, repository settings, or credentials were changed in this session.

This repository is public. This handoff deliberately omits sensitive evidence,
customer details, and descriptions of unresolved exploit paths. Do not commit
the private audit directory or export customer punch-list notes into Git.

## Shipped and validated

- `fcaa112` — explicit administrator email allowlist, including every existing
  project/pass/cohort bypass; missing configuration fails closed. The local
  Mac launcher configures a separate loopback identity.
- Credential creation requires admin; customer project moves/copies cannot
  cross client boundaries. Multipart requests have actual total-body/file
  limits, rather than just multipart RAM thresholds.
- Production HTTP header/idle timeouts; HTTPS project cookies carry `Secure`.
  The service now has `NoNewPrivileges=yes` and `UMask=0077`. These are
  incremental protections, not a dedicated-user/converter sandbox.
- `29fc12e` — offsite sync refuses conflicting overwrites and downloads bytes
  for SHA-256 verification. Offsite drills compare with a local counterpart
  when available. Offline regression coverage exercises matching bytes,
  same-size wrong bytes, and copy failure.
- Both commits pushed; production health reports `0924.29fc12e`.
- Full `go test ./...`, `go vet ./...`, relevant shellcheck, gofmt,
  `git diff --check`, and `python3 scripts/test-backup-verification.py` passed.
- Local live HTTP checks: anonymous admin 401, non-allowlisted authenticated
  identity 403, allowlisted owner 200, anonymous credential creation denied,
  public storefront 200. No customer modifications or transactional emails
  were used for these probes.

External proxy/browser and direct-ingress verification remain outstanding;
the live probes above ran locally and supplied proxy identity headers.

## Blocking work

1. **Offsite recovery:** current downloaded evidence does not match the local
   snapshot. Fresh upload attempts failed with provider authorization errors.
   Do not mistake structural SQLite integrity for source identity/currentness.
   Existing remote objects were not overwritten. Failure markers intentionally
   keep the app backup dashboard at **ACTION NEEDED**.
2. **Owner decisions:** inspect person/team/root/web shares and invitation
   links; confirm existing customer app logins before removing grants. Review
   GitHub publication rights, key scope, collaborators, and branch/CI controls.
3. **Complete recovery:** database snapshots contain manuscript/output BLOBs
   but do not replace encrypted recovery of configuration, secrets, licensed
   fonts, external documents, and toolchain. A clean-machine restore/build
   with outbound integrations disabled has not been demonstrated.
4. **Remaining inbound review:** retire unused services, verify trusted
   ingress externally, establish least-privilege runtime/converter isolation,
   and finish remaining authorization/abuse checks before declaring readiness.

The offsite root cause is unknown. Another writer/environment is a hypothesis,
not a finding. Use a deployment-specific namespace and unique immutable
snapshot identifiers after confirming credentials and all writers; never
repair by blindly overwriting the evidence.

## Private operational handoff

On the VM, read `scratch/security-2026-09-24/PRIVATE-FINDINGS-2026-09-24.md`,
`app-review.md`, `github-review.md`, and `PRIMARY-RESULTS-2026-09-24.md`.
This ignored directory also holds downloaded evidence, test logs, recovery
attempt logs, and the preserved previous backup-status markers. Do not serve
this directory with a static web server. Private punch-list section 7 tracks
owner decisions and remaining work; it was not exported to the public repo.

**Exact next action:** obtain the owner's VM-share review and R2 credential/
writer review, then upload a current snapshot to a new authorized unique key.
Download it into private scratch, compare SHA-256, integrity, meaningful
record/BLOB counts, and complete an isolated recovery drill. Only clear
failure markers after the underlying problem is verified resolved.

## Coordination / repositories

- Primary session `cJFM2DP` owns remediation. Independent reviewer `cK4MWGP`
  contributed publish-safe checklist commit `cb2d5e8`; its detailed report
  remains private. Another parallel reviewer was asked to stop remediation
  and preserve all remote evidence; inspect coordination notes before changes.
- Public-doc repo was unchanged in this session; independent reviewer
  confirmed existing commits through `58c64bf` pushed.
- Both working trees were clean before this handoff was added.
- No app recompile is needed for this handoff commit; deployed code version
  intentionally remains `0924.29fc12e`.

Metrics: approximately 60% context at handoff; zero large-file readguard
bypasses. Bulk app/GitHub/backup review was delegated read-only.

## Closeout planning addendum — September 24, 2026

Jenna has removed the five individual VM shares. Her final `share show` reports
PUBLIC port 8000 and no individual shares; its earlier “now private” deletion
message was inconsistent with that final status. External acceptance checks
remain outstanding.

The user asked to share the remaining closeout work. Independent session
`cK4MWGP` is now coordinating the follow-up. See
`SECURITY-CLOSEOUT-PLAN-2026-09-24.md`; private punch-list section 7 now separates
owner account decisions from agent implementation and records completion
criteria. Detailed findings and the punch-list archive remain unexported.

Next owner actions: approve/set intended GitHub visibility and review working,
bucket-scoped backup credentials without pasting secrets into chat. Next agent
actions: verify authenticated source access, complete the private recovery
inventory, then create/download/hash-verify a current offsite snapshot once
write access is repaired. No new successful offsite recovery or complete
restore is claimed by this planning addendum.

Preparation completed: VM `origin` now uses the existing authenticated SSH
path and fetch succeeded; the broad key was not changed or revoked. A private
metadata-only recovery inventory records 12 roots, two symlinks to review and
five tool versions in `scratch/security-2026-09-24/RECOVERY-INVENTORY-2026-09-24.json`.
No secret values were copied into it. The existing runpage was restarted with
`RUNPAGE_CHAT_CONV=cK4MWGP`, so notes now reach this conversation.

No application code, production database, cloud permissions or backup objects
were changed in this follow-up. Keep the GitHub/R2 account decisions first;
do not infer a repaired backup from the existence of this plan.

Metrics for follow-up: approximately 60% context; zero large-file readguard
bypasses. No additional subagents or live transactional sends.

## Consolidation addendum — September 24, 2026 (afternoon)

Session `cTB3D2B` reconciled all earlier notes. Summary safe for publication:

- The offsite "403" was a client-side behaviour (rclone attempting to create the
  bucket with a bucket-scoped token), not a credential failure. The mismatched
  remote snapshot came from a second host holding a copy of this deployment's
  backup configuration. Backups now write to a per-deployment prefix
  (`scripts/r2-env.sh`, `scripts/r2-init-namespace.sh`); a full upload,
  SHA-256 readback and restore drill succeeded. Locating the second host and
  rotating the shared credential are owner actions tracked privately.
- Two pipeline fixes deployed (`1caf636`, `7e53175`): Typst builds are confined
  to their job directory, and failed-inspection reports escape error output.
- Deployed version `0924.7e53175`; full `go test ./...` green.
- Remaining application findings and owner decisions are in private punch-list
  section 7 (items 7.4–7.16). Private consolidated report:
  `scratch/security-2026-09-24/CONSOLIDATED-REPORT-2026-09-24.md` (not in Git).
