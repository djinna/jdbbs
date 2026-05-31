# gstack

Use the `/browse` skill from gstack for all web browsing. Never use `mcp__claude-in-chrome__*` tools.

Available skills (gstack v1.55):
- **Browser / QA:** `/browse`, `/qa`, `/qa-only`, `/design-review`, `/benchmark`, `/canary`, `/scrape`, `/skillify`, `/open-gstack-browser`, `/pair-agent`, `/setup-browser-cookies`
- **Plan & review:** `/office-hours`, `/spec`, `/autoplan`, `/plan-ceo-review`, `/plan-eng-review`, `/plan-design-review`, `/plan-devex-review`, `/review`, `/devex-review`, `/codex`, `/plan-tune`
- **Design:** `/design-consultation`, `/design-shotgun`, `/design-html`
- **Ship & deploy:** `/ship`, `/land-and-deploy`, `/landing-report`, `/setup-deploy`
- **Debug & quality:** `/investigate`, `/health`, `/cso`
- **Docs:** `/document-release`, `/document-generate`, `/make-pdf`
- **Context & retros:** `/context-save`, `/context-restore`, `/learn`, `/retro`
- **Safety:** `/careful`, `/freeze`, `/guard`, `/unfreeze`
- **gbrain memory:** `/setup-gbrain`, `/sync-gbrain`
- **iOS:** `/ios-qa`, `/ios-fix`, `/ios-clean`, `/ios-sync`, `/ios-design-review`
- **Meta:** `/gstack-upgrade`, `/benchmark-models`

If gstack skills are not working, run `cd .claude/skills/gstack && ./setup` to build the binary and register skills.

# VM ssh — user-run, not assistant-run

The harness's auto-mode classifier blocks `ssh exedev@jdbbs.exe.xyz '...'` from the assistant (treats prod VM access as needing explicit approval). Don't try to run ssh yourself. Instead, write a single concatenated shell command (one line, `&&`-chained or `;`-chained, multi-step pre-flights and compiles wrapped in one ssh invocation) and ask the user to paste the output back. The user has shell access and runs it themselves; you stay read-only on local files and operate on what they paste in.

Examples of commands to give the user:
- Pre-flight: `ssh exedev@jdbbs.exe.xyz 'systemctl is-active prodcal && git -C /home/exedev/prodcal log --oneline -1 && typst --version'`
- Compile + scp back: `ssh exedev@jdbbs.exe.xyz 'cd /home/exedev/prodcal && git pull --ff-only && typst compile --root . --font-path typesetting/fonts manuscripts/ghosts/main.typ /tmp/ghosts-typst-v2.pdf && pdfinfo /tmp/ghosts-typst-v2.pdf | grep -E "Pages|Page size"' && scp exedev@jdbbs.exe.xyz:/tmp/ghosts-typst-v2.pdf scratch/`

# Scratch / throwaway files

For throwaway intermediate files (assembled DOCX for testing, downloaded PDFs for diffing, etc.) use `./scratch/` at the repo root. It's gitignored. Don't use `/tmp` — macOS Finder hides it by default, which is friction when the user needs to interact with the file via UI (file picker, etc.). Don't scatter `tmp/` subdirs across the source tree (e.g. inside `manuscripts/`) — they pollute the source areas.

## Health Stack

Used by `/health` and as the local dev inner loop:

- typecheck: go build ./...
- lint: go vet ./...
- format: gofmt -l cmd srv db   (fix with: gofmt -w cmd srv db)
- test: go test ./...
- shell: shellcheck scripts/*.sh typesetting/scripts/*.sh

**Hybrid toolchain.** Go + shellcheck run locally (`brew install go shellcheck`) for the fast inner loop (build/vet/gofmt/shell + most tests). The doc pipeline (`python-docx`, `typst`, `pandoc`) lives on the VM; a few integration tests shell out to it (e.g. `TestWordTemplateGeneration*` needs `python-docx`) and fail locally with `ModuleNotFoundError: docx` unless you install those deps. Pipeline-touching tests are the VM's parity job; run `go test ./...` on the VM (`exedev@jdbbs.exe.xyz:/home/exedev/prodcal`) for the full green. A local `.git/hooks/pre-commit` enforces gofmt + go vet (bypass with `git commit --no-verify`).
