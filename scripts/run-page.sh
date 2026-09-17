#!/usr/bin/env bash
# Render scratch/run/CHECKLIST.md to scratch/run/index.html (auto-refreshing)
# and serve it on :8766 → https://jdbbs.exe.xyz:8766/ . Re-render on change.
set -euo pipefail
cd "$(dirname "$0")/../scratch/run"
render() {
  pandoc CHECKLIST.md -s -o index.html --metadata pagetitle="Run" \
    -V header-includes='<meta http-equiv="refresh" content="20"><style>body{max-width:44rem;margin:2rem auto;padding:0 1rem;font:16px/1.5 system-ui}li{margin:.3rem 0}h2{margin-top:2rem;border-top:1px solid #ddd;padding-top:1rem}code{background:#f3f3f3;padding:0 .25em}</style>'
  sed -i -E 's/\[x\]/<span style="color:green">☑<\/span>/g; s/\[~\]/<span style="color:#c60">◐<\/span>/g; s/\[ \]/☐/g' index.html
}
render
busybox httpd -f -p 8766 -h . &
while sleep 3; do render; done
