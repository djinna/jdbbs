#!/usr/bin/env bash
# Local dev tooling setup for jdbbs (Go server: cmd/, srv/, db/).
# Idempotent — safe to re-run. Run from anywhere inside the repo.
#
#   bash scripts/dev-setup.sh          # check tools, activate git hooks, verify
#   bash scripts/dev-setup.sh --brew   # also `brew install` any missing tools
#
# See docs/DEV-SETUP.md for the full picture (incl. optional python-docx parity).
export PATH="/opt/homebrew/bin:/usr/local/bin:$PATH"   # arm64 + intel Homebrew

cd "$(git rev-parse --show-toplevel)" || { echo "not in a git repo"; exit 1; }

WITH_BREW=0
[ "${1:-}" = "--brew" ] && WITH_BREW=1

echo "== jdbbs dev setup =="

# 1. Required tools -----------------------------------------------------------
missing=()
for t in go shellcheck; do
  command -v "$t" >/dev/null 2>&1 || missing+=("$t")
done

if [ ${#missing[@]} -gt 0 ]; then
  if [ "$WITH_BREW" = "1" ] && command -v brew >/dev/null 2>&1; then
    echo "installing: ${missing[*]}"
    brew install "${missing[@]}"
  else
    echo "missing tools: ${missing[*]}"
    echo "  install with:  brew install ${missing[*]}"
    echo "  or re-run:     bash scripts/dev-setup.sh --brew"
    exit 1
  fi
fi
echo "tools: go $(go version | awk '{print $3}'), shellcheck $(shellcheck --version 2>/dev/null | awk '/version:/{print $2}')"

# 2. Activate the versioned git hooks ----------------------------------------
git config core.hooksPath .githooks
chmod +x .githooks/* 2>/dev/null || true
echo "git hooks: core.hooksPath -> .githooks (pre-commit active)"

# 3. Verify the Go inner loop ------------------------------------------------
echo "-- go build ./... --"
if go build ./...; then echo "  build: ok"; else echo "  build: FAILED (see errors above)"; fi
echo "-- go vet ./... --"
if go vet ./...; then echo "  vet: ok"; else echo "  vet: FAILED (see errors above)"; fi
unfmt=$(gofmt -l cmd srv db 2>/dev/null)
if [ -z "$unfmt" ]; then
  echo "gofmt: clean"
else
  echo "gofmt: unformatted files ->"
  while IFS= read -r f; do printf '  %s\n' "$f"; done <<< "$unfmt"
  echo "  fix: gofmt -w cmd srv db"
fi

echo "== done =="
echo "Pure-Go checks (build/vet/gofmt + most tests) run locally."
echo "Pipeline tests (python-docx / typst / pandoc) are VM-side by default — see docs/DEV-SETUP.md."
