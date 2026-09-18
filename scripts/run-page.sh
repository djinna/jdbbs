#!/usr/bin/env bash
# Render scratch/run/CHECKLIST.md to scratch/run/index.html (auto-refreshing)
# and serve it on :8766 → https://jdbbs.exe.xyz:8766/ . Re-render on change.
set -euo pipefail
cd "$(dirname "$0")/../scratch/run"
render() {
  pandoc CHECKLIST.md -s -o index.html --metadata pagetitle="Run" \
    -V header-includes='<meta http-equiv="refresh" content="20"><style>body{max-width:44rem;margin:2rem auto;padding:0 1rem;font:16px/1.5 system-ui;background:#111;color:#ddd}a{color:#8ab4f8}li{margin:.3rem 0}h2{margin-top:2rem;border-top:1px solid #333;padding-top:1rem}code{background:#222;color:#e8e8e8;padding:0 .25em}strong{color:#fff}</style>'
  sed -i -E 's/\[x\]/<span style="color:#5c5">☑<\/span>/g; s/\[~\]/<span style="color:#fa5">◐<\/span>/g; s/\[ \]/☐/g' index.html
}
render
busybox httpd -f -p 8766 -h . &
while sleep 3; do render; done
