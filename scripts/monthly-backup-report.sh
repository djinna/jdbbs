#!/usr/bin/env bash
# Send a monthly ProdCal backup status report via AgentMail.

set -euo pipefail

BACKUP_DIR="${BACKUP_DIR:-$HOME/backups}"
PRODCAL_DIR="${PRODCAL_DIR:-/home/exedev/prodcal}"
RECIPIENTS="${BACKUP_REPORT_RECIPIENTS:-j@djinna.com}"
AGENTMAIL_API_KEY="${AGENTMAIL_API_KEY:-}"
AGENTMAIL_INBOX_ID="${AGENTMAIL_INBOX_ID:-}"

if [ -z "$AGENTMAIL_API_KEY" ] || [ -z "$AGENTMAIL_INBOX_ID" ]; then
  # shellcheck disable=SC1091
  [ -f "$PRODCAL_DIR/.env" ] && set -a && . "$PRODCAL_DIR/.env" && set +a
fi

log() { printf '[%s] %s\n' "$(date -u +%FT%TZ)" "$*"; }
die() { log "FAIL: $*" >&2; exit 1; }

[ -n "${AGENTMAIL_API_KEY:-}" ] || die "AGENTMAIL_API_KEY not set"
[ -n "${AGENTMAIL_INBOX_ID:-}" ] || die "AGENTMAIL_INBOX_ID not set"
command -v python3 >/dev/null 2>&1 || die "python3 not installed"

report_json="$(mktemp -t prodcal-backup-report.XXXXXX.json)"
trap 'rm -f "$report_json"' EXIT

BACKUP_DIR="$BACKUP_DIR" python3 - > "$report_json" <<'PY'
import glob
import json
import os
import subprocess
import time
from datetime import datetime, timezone
from pathlib import Path

backup_dir = Path(os.environ["BACKUP_DIR"])
now = time.time()

def sentinel(name):
    p = backup_dir / name
    if not p.exists():
        return {"exists": False}
    text = p.read_text(errors="replace")
    out = {"exists": True, "path": str(p), "text": text.strip()}
    for line in text.splitlines():
        if ":" in line:
            k, v = line.split(":", 1)
            out[k.strip()] = v.strip()
    return out

files = sorted(backup_dir.glob("prodcal-*.sqlite3.gz"))
local_bytes = sum(p.stat().st_size for p in files)
newest = max(files, key=lambda p: p.stat().st_mtime) if files else None

def age_hours(p):
    return (now - p.stat().st_mtime) / 3600 if p else None

r2_size = None
r2_error = None
try:
    cp = subprocess.run(["rclone", "size", "r2:jdbbs-backups/db", "--json"], text=True, capture_output=True, timeout=120)
    if cp.returncode == 0:
        r2_size = json.loads(cp.stdout)
    else:
        r2_error = (cp.stderr or cp.stdout).strip()
except Exception as e:
    r2_error = str(e)

sent = {name: sentinel(name) for name in [
    ".LAST-SUCCESS", ".LAST-FAILURE", ".LAST-R2-SUCCESS", ".LAST-R2-FAILURE", ".HEALTH-OK", ".HEALTH-FAIL", ".LAST-DRILL-SUCCESS", ".LAST-DRILL-FAILURE", ".LAST-R2-MONTHLY-SUCCESS", ".LAST-R2-MONTHLY-FAILURE"
]}

problems = []
if not newest:
    problems.append("No local prodcal backup files found")
elif age_hours(newest) > 30:
    problems.append(f"Newest local backup is {age_hours(newest):.1f}h old")
for fail in [".LAST-FAILURE", ".LAST-R2-FAILURE", ".HEALTH-FAIL", ".LAST-DRILL-FAILURE", ".LAST-R2-MONTHLY-FAILURE"]:
    if sent[fail]["exists"]:
        problems.append(f"{fail} is present")
for ok in [".LAST-SUCCESS", ".LAST-R2-SUCCESS", ".HEALTH-OK"]:
    if not sent[ok]["exists"]:
        problems.append(f"{ok} missing")
if r2_error:
    problems.append("Could not read R2 size: " + r2_error[:200])

out = {
    "generated_at": datetime.now(timezone.utc).isoformat(timespec="seconds"),
    "ok": not problems,
    "problems": problems,
    "local": {
        "dir": str(backup_dir),
        "count": len(files),
        "bytes": local_bytes,
        "newest": newest.name if newest else None,
        "newest_age_hours": age_hours(newest),
    },
    "r2": {
        "remote": "r2:jdbbs-backups/db",
        "size": r2_size,
        "error": r2_error,
    },
    "sentinels": sent,
}
print(json.dumps(out, indent=2))
PY

python3 - "$report_json" "$RECIPIENTS" "$AGENTMAIL_INBOX_ID" "$AGENTMAIL_API_KEY" <<'PY'
import html
import json
import sys
import urllib.error
import urllib.request

report_path, recipients, inbox_id, api_key = sys.argv[1:]
report = json.load(open(report_path, encoding="utf-8"))
status = "OK" if report["ok"] else "ACTION NEEDED"
subject = f"ProdCal backup report — {status} — {report['generated_at'][:10]}"

def human(n):
    if n is None:
        return "n/a"
    units = ["B", "KiB", "MiB", "GiB", "TiB"]
    n = float(n)
    for u in units:
        if n < 1024 or u == units[-1]:
            return f"{n:.1f} {u}" if u != "B" else f"{int(n)} B"
        n /= 1024

local = report["local"]
r2 = report["r2"]
r2_size = r2.get("size") or {}
problems = report["problems"]
lines = [
    f"ProdCal backup report — {status}",
    f"Generated: {report['generated_at']} UTC",
    "",
    "Local backups:",
    f"- Count: {local['count']}",
    f"- Size: {human(local['bytes'])}",
    f"- Newest: {local['newest']} ({local['newest_age_hours']:.1f}h old)" if local.get("newest") else "- Newest: none",
    "",
    "R2 backups:",
    f"- Remote: {r2['remote']}",
    f"- Objects: {r2_size.get('count', 'n/a')}",
    f"- Size: {human(r2_size.get('bytes'))}",
    "",
    "Problems:" if problems else "Problems: none",
]
for p in problems:
    lines.append(f"- {p}")
lines += ["", "Sentinels:"]
for name, sent in report["sentinels"].items():
    if sent.get("exists"):
        lines.append(f"- {name}: {sent.get('time', 'present')}")
    else:
        lines.append(f"- {name}: missing")
text = "\n".join(lines)

html_body = """
<div style="font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;max-width:680px;margin:0 auto;color:#111827">
  <h1 style="font-size:22px;margin-bottom:4px">ProdCal backup report — {status}</h1>
  <p style="color:#6b7280;margin-top:0">Generated {generated} UTC</p>
  <h2 style="font-size:16px">Local backups</h2>
  <ul><li>Count: {local_count}</li><li>Size: {local_size}</li><li>Newest: {newest}</li></ul>
  <h2 style="font-size:16px">R2 backups</h2>
  <ul><li>Remote: {remote}</li><li>Objects: {r2_count}</li><li>Size: {r2_size}</li></ul>
  <h2 style="font-size:16px">Problems</h2>
  {problems_html}
  <h2 style="font-size:16px">Sentinels</h2>
  <pre style="background:#f3f4f6;padding:12px;border-radius:6px;white-space:pre-wrap">{sentinels}</pre>
</div>
""".format(
    status=html.escape(status),
    generated=html.escape(report["generated_at"]),
    local_count=local["count"],
    local_size=human(local["bytes"]),
    newest=html.escape(f"{local['newest']} ({local['newest_age_hours']:.1f}h old)" if local.get("newest") else "none"),
    remote=html.escape(r2["remote"]),
    r2_count=r2_size.get("count", "n/a"),
    r2_size=human(r2_size.get("bytes")),
    problems_html="<p>None.</p>" if not problems else "<ul>" + "".join(f"<li>{html.escape(p)}</li>" for p in problems) + "</ul>",
    sentinels=html.escape("\n".join(f"{k}: {v.get('time', 'present') if v.get('exists') else 'missing'}" for k, v in report["sentinels"].items())),
)

payload = json.dumps({
    "to": [r.strip() for r in recipients.split(",") if r.strip()],
    "subject": subject,
    "text": text,
    "html": html_body,
}).encode()
req = urllib.request.Request(
    f"https://api.agentmail.to/v0/inboxes/{inbox_id}/messages/send",
    data=payload,
    headers={"Content-Type": "application/json", "Authorization": "Bearer " + api_key},
    method="POST",
)
try:
    with urllib.request.urlopen(req, timeout=30) as resp:
        print(f"sent backup report to {recipients}: HTTP {resp.status}")
except urllib.error.HTTPError as e:
    raise SystemExit(f"AgentMail HTTP {e.code}: {e.read().decode(errors='replace')}")
PY
