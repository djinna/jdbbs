# Next session: triage the 15 Dependabot alerts on `djinna/jdbbs` (ProdCal)

_Created 2026-08-15. Context: pre-existing dependency vulnerabilities surfaced
during the workshop-registration push; unrelated to that change. Separate
errand, low urgency, but worth clearing._

GitHub reports **15 vulnerabilities** on `main` (7 critical, 3 high, 5
moderate): https://github.com/djinna/jdbbs/security/dependabot — all
pre-existing, none introduced by recent work. This is a Go web app; the repo is
at `/home/exedev/prodcal` on the VM. Direct deps are minimal
(`github.com/coreos/go-systemd/v22`, `github.com/webview/webview_go`,
`golang.org/x/crypto`, `modernc.org/sqlite`), so most alerts are almost
certainly **transitive** module CVEs. `go.mod` pins `go 1.26.4`; the VM has
`GOTOOLCHAIN=auto`.

**Goal:** get the alert count down safely, without breaking the build or the
doc pipeline (pandoc/typst/python-docx).

## Approach
1. Read the Dependabot list first (ask the user to paste it, or fetch via `gh`
   if it's authed) so you know exactly which modules/CVEs — don't blind-bump.
2. Prefer minimal targeted bumps: `go get <module>@<fixed-version>` per
   advisory, then `go mod tidy`. A broad `go get -u ./... && go mod tidy` is
   fine but review the diff.
3. There's a `dependabot/go_modules/go_modules-a3c8a40308` branch on origin —
   check whether it's a useful batched bump to merge or just noise.
4. **Must stay green:** `go build ./...`, `go vet ./...`,
   `gofmt -l cmd srv db`, `go test ./...`. Some pipeline tests
   (python-docx/typst/pandoc) only pass on the VM — run the full `go test ./...`
   there.
5. Flag (don't force) anything that isn't a clean fix: a CVE with no fixed
   version yet, or a major-version bump with API changes — especially
   `modernc.org/sqlite`, which is the DB driver; test carefully.

## Git flow (hub-and-spoke, per AGENTS.md)
Commit locally → push to GitHub `origin` → on the VM
`git pull --ff-only && make build && sudo systemctl restart prodcal`.
The user runs all `ssh exedev@jdbbs.exe.xyz '...'` commands themselves — give
single concatenated one-liners and have them paste the output back. Verify the
service comes back `active` and `/healthz` is ok after restart.

## Deliverable
A short summary of what was bumped, what count remains and why (with the
"can't fix yet / risky" ones called out), committed with a clear message.
