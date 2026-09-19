#!/usr/bin/env bash
# factory-demo.sh — the six-call factory demo for the workshop / talk (punch list 0.9).
#
#   TOKEN=… P=17 scripts/factory-demo.sh ms.docx            # inspect only (free)
#   TOKEN=… P=17 scripts/factory-demo.sh ms.docx --build    # + one build credit, downloads both outputs
#
# Each step prints the command before running it so the room can read along.
# Never point it at pinstitute (id 22); use mcheck/book-001 (id 17) or a demo pass.
set -euo pipefail
B=${B:-https://jdbbs.exe.xyz}; : "${TOKEN:?set TOKEN}"; : "${P:?set P (project id)}"
MS=${1:?manuscript.docx}; H="Authorization: Bearer $TOKEN"
step(){ printf '\n\033[1;36m$ %s\033[0m\n' "$*"; sleep "${PAUSE:-1}"; }

step "curl \$B/api/projects/\$P/pass"
curl -s -H "$H" "$B/api/projects/$P/pass" | jq -c '{status,builds_used,credits_remaining}'

step "curl -F file=@$MS … \$B/api/books/upload"
BOOK=$(curl -s -H "$H" -F "file=@$MS" -F title="${TITLE:-Demo}" -F author="${AUTHOR:-Workshop}" -F project_id="$P" "$B/api/books/upload" | jq .id)
echo "book id: $BOOK"

step "curl -X POST \$B/api/projects/\$P/preflight {book_id:$BOOK}   # free"
curl -s -H "$H" -X POST "$B/api/projects/$P/preflight" -H 'Content-Type: application/json' -d "{\"book_id\":$BOOK}" | jq -c '.summary, (.book_map.sections | map(.title) | .[:6])'

[[ "${2:-}" == "--build" ]] || { echo; echo "(inspect only — add --build to spend a credit)"; exit 0; }

step "curl -X POST \$B/api/books/$BOOK/convert {format:both, kind:proof}   # free; kind:final uses one credit"
curl -s -H "$H" -X POST "$B/api/books/$BOOK/convert" -H 'Content-Type: application/json' -d '{"format":"both","kind":"proof"}' | jq -c .

step "poll \$B/api/books/$BOOK until ready"
until curl -s -H "$H" "$B/api/books/$BOOK" | jq -e '.status=="ready" or .status=="error"' >/dev/null; do printf .; sleep 3; done; echo
curl -s -H "$H" "$B/api/books/$BOOK" | jq -c '{status, error, outputs: [.outputs[]? | {format,size_bytes}]}'

step "download"
mkdir -p "${OUT:-out}"
curl -s -H "$H" -o "${OUT:-out}/book.pdf"  "$B/api/books/$BOOK/download/pdf"
curl -s -H "$H" -o "${OUT:-out}/book.epub" "$B/api/books/$BOOK/download/epub"
ls -l "${OUT:-out}"
