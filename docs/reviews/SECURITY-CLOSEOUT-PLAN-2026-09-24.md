# ProdCal security closeout plan — September 24, 2026

**Finish by demonstrating recovery and least-privilege access, not just by
changing settings.** Jenna owns account-level decisions; Shelley owns
implementation and reproducible verification. The private shared punch list,
section 7, holds operational status and notes.

This plan is safe to publish. Keep incident evidence, customer inventories,
credentials and recovery artifacts out of source control. Do not export the
private punch-list archive to a public repository.

## 1. Contain unnecessary publication — Jenna, then Shelley

- **Jenna:** recommended immediate step is making the app repository private.
  Check GitHub's visibility-change notice before confirming. This concerns
  source distribution, not the VM-served application's public front door.
- **Shelley:** verify authenticated repository access before deployment needs
  it; review tracked and historical manuscript, customer-note, image and font
  paths. Present a publication inventory without copying their contents into
  another report.
- **Together:** decide which materials belong in a separate public tree.
  Private visibility does not erase public forks, earlier clones or downloads.
  Any historical cleanup/force-push needs its own approval and coordination.

**Done when:** intended visibility is confirmed, confidential content is
separated from intentional publication, and deployment can fetch the source.

## 2. Establish a current recovery point — Jenna, then Shelley

- **Jenna:** inspect backup-bucket access and other writers. If replacement
  credentials are needed, use bucket-scoped object read/write credentials
  rather than account administration. Enter secrets through secure interactive
  configuration, not chat or Git.
- **Shelley:** preserve existing objects and evidence; use a deployment-specific
  destination and uniquely identified snapshot. Download the new object,
  compare SHA-256 to its source snapshot, run SQLite integrity checks, and
  compare meaningful record and BLOB counts/hashes.
- **Together:** choose provider-enforced retention/deletion protection and an
  independent recovery copy. Client-side no-overwrite flags are not immutable
  storage.

**Done when:** a fresh offsite object is demonstrably the intended current
snapshot, not merely a well-formed database with a recent filename.

## 3. Cover the whole application — Shelley, with Jenna holding recovery keys

- Inventory SQLite and its source/output/cover BLOBs; Git revisions and bundles;
  external documents; licensed fonts; configuration and signing secrets;
  service definitions; and document-toolchain installation details.
- Encrypt the recovery package. Keep the decryption key outside the VM and
  independently recoverable; only the public encryption recipient needs to be
  available to the automated backup job.
- Automate daily backups and verification, agree retention, and alert on stale
  successful backup/restore verification as well as explicit failures.
- Restore into isolation, with live email, payments, LLM calls and callbacks
  prevented. Check customer/pass records, file hashes, existing downloads,
  authorization boundaries, and one new document build.
- Record elapsed recovery time and agree recovery objectives. Initial proposed
  targets are at most 24 hours of data loss and restoration within one business
  day; they remain unproven until measured.

**Done when:** an operator can follow the runbook on a replacement environment
and recover both customer data and a working book-production pipeline.

## 4. Reduce ongoing access and change risk — shared

- Prepare repository-scoped GitHub access for each repository. Test the new
  path before retiring an existing account-wide key; preserve required
  agent/deployment work without granting unrelated-repository access.
- Review collaborators, apps, tokens and MFA. Add suitable CI and enable
  branch/ruleset controls supported by the repository's plan.
- Review surviving VM listeners and trusted ingress from outside the VM;
  isolate application/converter execution from operator credentials.
- Confirm public pages work signed out, non-admins cannot enter administration,
  and customers cannot cross project/client boundaries.
- Agree any credential/session rotation after reviewing exposure and
  replacement dependencies; do not silently invalidate customer sessions.

**Done when:** old unnecessary access is retired, negative authorization tests
pass, and ongoing changes/backups have observable safeguards.

## Working agreement

Work from private punch-list items 7.2–7.13. Mark an item complete only with its
evidence, not when a script is written or an account setting is requested.
Do not delete old recovery objects, rewrite history, revoke an active key, or
change customer entitlements without the relevant approval.

Documentation consulted September 24, 2026: GitHub “Setting repository
visibility”; Cloudflare R2 “API tokens”; exe.dev sharing/proxy documentation.
