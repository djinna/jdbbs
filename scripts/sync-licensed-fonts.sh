#!/usr/bin/env bash
# Sync print-licensed fonts from this Mac to the VM. The font files are
# never committed to git (TRK-DESIGN-002) — this rsync is the only way they
# reach the VM's typst compile. Run on demand after dropping new families
# into typesetting/fonts/licensed/.
set -euo pipefail

SRC="${HOME}/jd-projects/jdbbs/typesetting/fonts/licensed/"
DST="exedev@jdbbs.exe.xyz:/home/exedev/prodcal/typesetting/fonts/licensed/"

if [ ! -d "$SRC" ] || [ -z "$(find "$SRC" -type f ! -name 'README.md' -print -quit 2>/dev/null)" ]; then
  echo "No licensed fonts at $SRC (only README) — nothing to sync."
  exit 0
fi

echo "Syncing $SRC -> $DST"
# --exclude README.md: the convention doc stays local; the VM only needs fonts.
rsync -avz --delete --exclude 'README.md' "$SRC" "$DST"

echo "Remote licensed/ contents:"
ssh exedev@jdbbs.exe.xyz "ls -laR /home/exedev/prodcal/typesetting/fonts/licensed/"
