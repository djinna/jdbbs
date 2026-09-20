#!/usr/bin/env bash
# store-go-live.sh — flip the Factory Pass store from Stripe sandbox to live.
#
# What "live" means here: the app talks to Stripe through the exe.dev proxy.
# PRODCAL_STRIPE_URL=https://stripe-test.int.exe.xyz  → test key (sandbox)
# PRODCAL_STRIPE_URL unset                             → https://stripe.int.exe.xyz, live key
# So the flip is: delete that one line from .env, restart prodcal, and the
# catalog/promos get created on the live account at boot (store.ensureCatalog).
#
# Scheduled by prodcal-store-live.timer for Tue 22 Sep 2026 16:00 UTC
# (= Wed 23 Sep 00:00 Hong Kong): workshop attendees build free through
# Tuesday HKT, billing starts Wednesday. Safe to run twice: if the line is
# already gone it says so and exits 0.
#
#   scripts/store-go-live.sh            # the real thing
#   scripts/store-go-live.sh --dry-run  # edits a copy, restarts nothing, emails nobody
set -euo pipefail

REPO=/home/exedev/prodcal
ENV_FILE="$REPO/.env"
DRY=0; [ "${1:-}" = "--dry-run" ] && DRY=1
log() { printf '%s store-go-live: %s\n' "$(date -u +%FT%TZ)" "$*"; }

if [ $DRY = 1 ]; then
  cp "$ENV_FILE" "$REPO/scratch/env.dry-run"; ENV_FILE="$REPO/scratch/env.dry-run"
  log "DRY RUN — working on $ENV_FILE"
fi

if ! grep -q '^PRODCAL_STRIPE_URL=' "$ENV_FILE"; then
  log "already live: no PRODCAL_STRIPE_URL in $ENV_FILE — nothing to do"
  exit 0
fi

# Workshop week ran with the "no finals left" refusal waived (studio setting
# finals_gate=off, 2026-09-20). Billing on = gate back on; deleting the row
# restores the compiled default ("on"). Picked up by the restart below.
if [ $DRY = 0 ]; then
  sqlite3 "$REPO/db.sqlite3" "DELETE FROM studio_settings WHERE key='finals_gate'" && log "finals_gate: back to default (on)"
else
  log "DRY RUN — would delete studio_settings.finals_gate (gate back on)"
fi

cp "$ENV_FILE" "$ENV_FILE.pre-live.$(date -u +%Y%m%dT%H%M%SZ)"
sed -i '/^PRODCAL_STRIPE_URL=/d' "$ENV_FILE"
log "removed PRODCAL_STRIPE_URL (backup beside .env)"

if [ $DRY = 1 ]; then
  log "DRY RUN — would: systemctl restart prodcal; then verify + email"
  diff <(grep -v '^PRODCAL_STRIPE_URL=' "$REPO/.env") "$ENV_FILE" && log "DRY RUN — resulting .env is exactly current minus that line: OK"
  exit 0
fi

sudo systemctl restart prodcal
sleep 4
STATUS=$(systemctl is-active prodcal || true)
CATALOG=$(journalctl -u prodcal --since '-1 min' --no-pager -o cat | grep -E 'store: catalog' | tail -1 || true)
CFG=$(curl -s -m 10 http://localhost:8000/api/public/store/config | head -c 400 || true)
log "prodcal=$STATUS | $CATALOG"
log "store config: $CFG"

# Tell Jenna. Uses the app's own AgentMail inbox; failure here is not fatal.
set +u
# shellcheck disable=SC1091
source "$REPO/.env"
set -u
if [ -n "${AGENTMAIL_API_KEY:-}" ] && [ -n "${AGENTMAIL_INBOX_ID:-}" ]; then
  BODY="Store flipped to live Stripe at $(date -u +'%a %d %b %Y %H:%M UTC') (00:00 Hong Kong).

prodcal: $STATUS
$CATALOG

Check: https://jdbbs.exe.xyz/admin/store/ should say 'Store is on'; the Catalog table should list prices, and the Stripe dashboard (live mode) should show the Factory Pass product and the WORKSHOP49 / PROTOCOL50 coupons.

If anything looks wrong: put PRODCAL_STRIPE_URL=https://stripe-test.int.exe.xyz back into /home/exedev/prodcal/.env and 'sudo systemctl restart prodcal'. A copy of the previous .env is beside it (.env.pre-live.*)."
  python3 - "$BODY" <<'PY' || log "email failed (non-fatal)"
import json, os, sys, urllib.request
body = sys.argv[1]
req = urllib.request.Request(
    f"https://api.agentmail.to/v0/inboxes/{os.environ['AGENTMAIL_INBOX_ID']}/messages/send",
    data=json.dumps({"to": ["j@djinna.com"], "subject": "jdbb studio store is LIVE — billing on", "text": body}).encode(),
    headers={"Authorization": "Bearer " + os.environ['AGENTMAIL_API_KEY'], "Content-Type": "application/json"})
with urllib.request.urlopen(req, timeout=20) as r:
    print("email sent", r.status)
PY
fi
log "done"
