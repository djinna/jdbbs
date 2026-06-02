#!/usr/bin/env bash
# Sync local prodcal backups to Cloudflare R2 for off-VM redundancy.
#
# Pairs with scripts/backup-db.sh: that script creates the local backup
# and probes it; this script ships the result off-VM. Run sync-to-r2.sh
# AFTER backup-db.sh, never in parallel.
#
# Prerequisites:
#   - rclone configured with a remote named "r2" (see docs in TRACKER
#     under TRK-OPS-007). To bootstrap: `rclone config` then select s3
#     -> Cloudflare -> paste account ID + access key + secret.
#   - Local backups exist at $BACKUP_DIR (default ~/backups).
#
# Behaviour:
#   - Mirrors $BACKUP_DIR -> r2:$BUCKET/$PREFIX using `rclone copy`
#     (not sync — we never want to delete remote-side automatically).
#   - Only ships *.sqlite3.gz files (filtered include).
#   - Verifies the newest local backup actually made it to R2 by
#     checking remote size matches.
#   - Drops a $BACKUP_DIR/.LAST-R2-SUCCESS or .LAST-R2-FAILURE sentinel
#     so the same out-of-band monitoring catches missed off-VM runs.

set -euo pipefail

BACKUP_DIR="${BACKUP_DIR:-$HOME/backups}"
R2_REMOTE="${R2_REMOTE:-r2}"
R2_BUCKET="${R2_BUCKET:-jdbbs-backups}"
R2_PREFIX="${R2_PREFIX:-db}"
LOCK_FILE="${BACKUP_DIR}/.r2-lock"
SUCCESS_FLAG="${BACKUP_DIR}/.LAST-R2-SUCCESS"
FAILURE_FLAG="${BACKUP_DIR}/.LAST-R2-FAILURE"

die() {
  local msg="$1"
  printf '[%s] R2 FAIL: %s\n' "$(date -u +%FT%TZ)" "$msg" >&2
  mkdir -p "$(dirname "$FAILURE_FLAG")"
  {
    printf 'time:    %s\n' "$(date -u +%FT%TZ)"
    printf 'script:  %s\n' "$0"
    printf 'host:    %s\n' "$(hostname)"
    printf 'remote:  %s:%s/%s\n' "$R2_REMOTE" "$R2_BUCKET" "$R2_PREFIX"
    printf 'reason:  %s\n' "$msg"
  } > "$FAILURE_FLAG"
  exit 1
}

log() { printf '[%s] %s\n' "$(date -u +%FT%TZ)" "$*"; }

mkdir -p "$BACKUP_DIR"

exec 9>"$LOCK_FILE"
flock -n 9 || die "another R2 sync is already running (lock: $LOCK_FILE)"

command -v rclone >/dev/null 2>&1 || die "rclone not installed"

if ! rclone listremotes 2>/dev/null | grep -qx "${R2_REMOTE}:"; then
  die "rclone remote '$R2_REMOTE' is not configured. Run \`rclone config\` first."
fi

# Probe the remote — fast no-op call that confirms credentials work and
# the bucket exists/accessible. We don't `rclone mkdir` because R2 buckets
# must be created via the dashboard or API.
if ! rclone lsd "${R2_REMOTE}:${R2_BUCKET}" >/dev/null 2>&1; then
  die "cannot list r2:${R2_BUCKET} — bucket missing or credentials wrong. Create the bucket in the Cloudflare R2 dashboard."
fi

newest_local="$(find "$BACKUP_DIR" -maxdepth 1 -name 'prodcal-*.sqlite3.gz' -printf '%T@ %p\n' 2>/dev/null \
                | sort -rn | head -1 | cut -d' ' -f2-)"
[ -n "${newest_local:-}" ] && [ -f "$newest_local" ] \
  || die "no local backup files found in $BACKUP_DIR (looking for prodcal-*.sqlite3.gz)"

newest_size="$(stat -c '%s' "$newest_local" 2>/dev/null || stat -f '%z' "$newest_local")"
newest_name="$(basename "$newest_local")"
log "newest local: $newest_name ($newest_size B)"

log "rclone copy ${BACKUP_DIR}/ -> ${R2_REMOTE}:${R2_BUCKET}/${R2_PREFIX}/ (filter: prodcal-*.sqlite3.gz)"
rclone copy "$BACKUP_DIR/" "${R2_REMOTE}:${R2_BUCKET}/${R2_PREFIX}/" \
  --include 'prodcal-*.sqlite3.gz' \
  --transfers 2 \
  --checkers 4 \
  --stats=0 \
  2>&1 | sed 's/^/  rclone: /' \
  || die "rclone copy failed"

# Verify the newest backup landed in R2 with the same size.
remote_size="$(rclone size --include "$newest_name" --json "${R2_REMOTE}:${R2_BUCKET}/${R2_PREFIX}/" 2>/dev/null \
              | grep -oE '"bytes":[0-9]+' | head -1 | cut -d: -f2)"
if [ -z "$remote_size" ]; then
  die "could not read remote size for $newest_name; the file may not have landed"
fi
if [ "$remote_size" != "$newest_size" ]; then
  die "remote size $remote_size != local size $newest_size for $newest_name"
fi

rm -f "$FAILURE_FLAG"
{
  printf 'time:        %s\n' "$(date -u +%FT%TZ)"
  printf 'remote:      %s:%s/%s/%s\n' "$R2_REMOTE" "$R2_BUCKET" "$R2_PREFIX" "$newest_name"
  printf 'bytes:       %s\n' "$remote_size"
} > "$SUCCESS_FLAG"

printf '[%s] R2 OK file=%s bytes=%s remote=%s:%s/%s\n' \
  "$(date -u +%FT%TZ)" "$newest_name" "$remote_size" "$R2_REMOTE" "$R2_BUCKET" "$R2_PREFIX"
