#!/usr/bin/env zsh
# jpull — fetch the canonical jdbbs repo, fast-forward if clean, print the
# TRACKER "Resume here" block. Idempotent. Clones if missing.
#
# Install: source this file from ~/.zshrc, e.g.:
#   source ~/jd-projects/jdbbs/scripts/jpull.sh
# Or paste the function body into ~/.zshrc.

jpull() {
  local base="$HOME/jd-projects"
  local repo="jdbbs"
  local dir="$base/$repo"
  local behind ahead dirty
  print -P "%F{cyan}=== $repo ===%f"
  if [[ ! -d "$dir/.git" ]]; then
    print -P "%F{yellow}cloning...%f"
    git -C "$base" clone "https://github.com/djinna/$repo.git" || return 1
  fi
  git -C "$dir" fetch --quiet
  behind=$(git -C "$dir" rev-list --count HEAD..@{u} 2>/dev/null)
  ahead=$(git -C "$dir" rev-list --count @{u}..HEAD 2>/dev/null)
  dirty=$(git -C "$dir" status --porcelain | wc -l | tr -d ' ')
  if [[ "$behind" -gt 0 && "$dirty" -eq 0 && "$ahead" -eq 0 ]]; then
    print -P "%F{green}fast-forward $behind commits%f"
    git -C "$dir" pull --ff-only --quiet
  elif [[ "$behind" -gt 0 ]]; then
    print -P "%F{red}behind $behind, ahead $ahead, dirty $dirty - resolve manually%f"
  else
    print -P "%F{green}up to date%f (ahead $ahead, dirty $dirty)"
  fi
  print -P "%F{cyan}=== TRACKER Resume here ===%f"
  sed -n '/^## .*Resume here/,/^---$/p' "$dir/TRACKER.md" 2>/dev/null | head -60
}
