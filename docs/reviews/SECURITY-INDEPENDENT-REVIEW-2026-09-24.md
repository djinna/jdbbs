# ProdCal independent security acceptance checklist — September 24, 2026

**Customers should receive application access, not infrastructure access.**
This is a release/recovery checklist, not a certification that every control
has passed. Reassess after the coordinated hardening work and a complete
offsite restore.

This file is intentionally safe to publish. Detailed operational evidence,
customer information, credentials, and unremediated exploit descriptions
belong in private storage, not a public source repository.

## Access model and customer migration

exe.dev Web sharing grants access to the VM's private HTTPS services; it is
not an application role. Root sharing additionally grants SSH, Terminal, and
Shelley. A Factory pass grants application entitlements for a customer's
project; the customer's login and the pass's validity are separate checks.
The public `/factory` page is an entry point, not a credential.

1. Explicitly authorize application administrators; never equate being
   authenticated by exe.dev with being a studio administrator.
2. Check each customer's existing project, login, pass status, expiry, and
   credits. Preserve those records rather than recreating them unnecessarily.
3. Test a customer's own-project access and rejection from another customer's
   project and all administrator endpoints.
4. Remove unnecessary person/team VM shares and unused invitation links.
   Removing a link does not remove people who already accepted it.
5. Test customer sign-in without an exe.dev VM share. Keep the public entry
   point reachable; do not make the whole VM proxy private as a substitute
   for application authorization.
6. Review prior Root access separately. Revocation does not reverse previous
   copies, credential use, or persistent changes made inside the VM.

The owner inspects sharing **from their own computer**, not inside the VM:

```sh
ssh exe.dev share show jdbbs
```

Reference: exe.dev documentation, `docs/sharing.md`, `docs/proxy.md`, and
`docs/login-with-exe.md`, consulted September 24, 2026.

## Inbound acceptance criteria

- Every listener has a named purpose, owner, expected audience, and shutdown
  policy. Preview servers never serve a repository root or secret/data tree.
- Administrator authorization is tested separately from authentication;
  client/project/cohort boundaries have negative tests.
- Credential creation, recovery, and replacement require the correct
  authority. A missing credential must not enable anonymous takeover.
- Trusted proxy headers cannot be supplied over an unintended direct path.
  Verify from outside the VM as well as locally.
- Upload/body limits apply before expensive parsing; authentication abuse,
  callbacks, and document-processing resource limits are reviewed.
- The application and untrusted document processors run with least privilege.
  Private editing/notes surfaces are not assumed safe merely because their
  URLs are obscure.

## GitHub acceptance criteria

- Confirm intended repository visibility and permission to distribute every
  tracked manuscript, customer note, image, and licensed asset.
- Separate public source/documentation from confidential operational data.
  Deleting a current file is not removal from Git history or existing clones.
- Restrict VM GitHub access to the necessary repository and permissions.
- Verify collaborators, deploy keys, apps, tokens, account MFA, branch rules,
  review requirements, and CI/secret scanning in authenticated GitHub settings.
  An anonymous API result or successful Git push does not verify all of these.

## Complete recovery acceptance criteria

- Inventory both database and filesystem state. Include manuscript and
  generated-output BLOBs, external documents, required licensed fonts,
  configuration, signing secrets, service definitions, and build dependencies.
- Keep source history and configuration recovery independent of the VM.
  Back up secrets in encrypted, access-controlled storage, never Git.
- Verify an offsite object by downloading it, comparing its cryptographic hash
  to the local snapshot, checking SQLite integrity, and checking meaningful
  record/BLOB counts. Fresh filenames and successful uploads are insufficient.
- Use deployment-specific object namespaces and unique snapshot identifiers;
  review all writers. A different environment must not replace production's
  recovery point.
- Verify actual daily/monthly retention and deletion protection at the storage
  provider; a checked-in lifecycle policy is not evidence it is active.
- Keep a recovery copy protected from credentials available to the running app.
- Schedule offsite restore drills and alert on failed or stale *successful
  restores*, not just failed backup jobs.
- Restore into an isolated environment with outbound email, payment activity,
  and callbacks disabled. Verify customer/project/pass records, source-file
  downloads, an existing output, and a new document build.
- Agree and measure recovery objectives. Initial proposed targets: at most
  24 hours of data loss and restoration within one business day; these are
  goals, not demonstrated guarantees.

## Coordination and limits

This independent review was coordinated with the concurrent primary security
session. It did not change production credentials, customer passes, sharing
settings, repository visibility, or database contents. Remediation deployment
and final verification belong to the primary session.

Do not mark the system ready until both the access-boundary tests and a
complete, current offsite recovery drill pass.
