# Local dev setup

How to get a Mac ready to work on the `jdbbs` Go server (`cmd/`, `srv/`, `db/`).
Tested on Apple Silicon; works on Intel too (the hook/script probe both Homebrew prefixes).

## TL;DR — new machine

```bash
git pull
brew install go shellcheck
bash scripts/dev-setup.sh
```

That's it. The script activates the git hook and verifies the toolchain. Re-running
it is safe (idempotent).

## What you get

A fast local inner loop — no ssh round-trip for the everyday checks:

| Check | Command |
|-------|---------|
| Type check | `go build ./...` |
| Lint | `go vet ./...` |
| Format | `gofmt -l cmd srv db` (fix: `gofmt -w cmd srv db`) |
| Shell lint | `shellcheck scripts/*.sh typesetting/scripts/*.sh` |

Plus editor intelligence: with `go` installed, `gopls` gives you inline type errors,
go-to-definition, autocomplete, and format-on-save in your editor.

## The git hook

`scripts/dev-setup.sh` runs `git config core.hooksPath .githooks`, which points git
at the **version-controlled** hook in [`.githooks/pre-commit`](../.githooks/pre-commit).
Because it lives in the repo, every machine runs the same hook and it updates with
`git pull` — no copying into `.git/hooks/`, no hand-pasting.

The hook blocks a commit only on problems *in the change you're committing*:
- staged `.go` files that aren't `gofmt`-clean (scoped to staged files, so pre-existing
  unformatted files elsewhere never block you)
- a `go vet` failure

Bypass once with `git commit --no-verify`. Undo the activation with
`git config --unset core.hooksPath`.

## Hybrid toolchain — what runs where

Pure-Go work runs locally. A few integration tests shell out to the document
pipeline (`pandoc`, `python3`/`python-docx`, `typst`), which lives on the VM
(`exedev@jdbbs.exe.xyz:/home/exedev/prodcal`). Those tests fail locally with
`ModuleNotFoundError: docx` unless you install the pipeline deps.

**Default:** run pure-Go checks locally; run the full `go test ./...` on the VM,
where the whole pipeline is assembled. This is deliberate — the VM is the
parity/deploy environment (`origin/main` deploys from there).

### Optional: full local test parity

If you want `go test ./...` fully green locally, install the pipeline deps. Note
that macOS Homebrew Python is "externally managed" (PEP 668), so a bare
`pip3 install` is refused. Use a virtualenv:

```bash
python3 -m venv .venv
source .venv/bin/activate
pip install python-docx
# typst + pandoc if you also build books locally:
brew install typst pandoc
```

The Go code invokes a bare `python3`, so the venv must be **active** in the shell
where you run `go test` for the word-template test to find `docx`. Most day-to-day
Go work doesn't touch these tests — skip this unless you need them.

## Why a script, not just steps

Manual steps drift between machines and rot. A committed script is one command that
does the same thing everywhere, and the hook being version-controlled means there's
one source of truth to improve. Update it once, every machine gets it on `git pull`.
