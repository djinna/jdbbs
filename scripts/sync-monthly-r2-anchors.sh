#!/usr/bin/env bash
# Copy one local ProdCal backup per calendar month to a non-expiring R2 prefix.
#
# Pair with an R2 lifecycle rule that expires daily backups under db/ after
# 180 days. Monthly anchors live under db-monthly/ and are not matched by
# that lifecycle rule.

set -euo pipefail

BACKUP_DIR="${BACKUP_DIR:-$HOME/backups}"
R2_REMOTE="${R2_REMOTE:-r2}"
R2_BUCKET="${R2_BUCKET:-jdbbs-backups}"
R2_MONTHLY_PREFIX="${R2_MONTHLY_PREFIX:-db-monthly}"
LOCK_FILE="${BACKUP_DIR}/.r2-monthly-lock"
SUCCESS_FLAG="${BACKUP_DIR}/.LAST-R2-MONTHLY-SUCCESS"
FAILURE_FLAG="${BACKUP_DIR}/.LAST-R2-MONTHLY-FAILURE"

log() { printf '[%s] %s\n' "$(date -u +%FT%TZ)" "$*"; }

die() {
  local msg="$1"
  log "R2 MONTHLY FAIL: $msg" >&2
  mkdir -p "$(dirname "$FAILURE_FLAG")"
  {
    printf 'time:    %s\n' "$(date -u +%FT%TZ)"
    printf 'script:  %s\n' "$0"
    printf 'host:    %s\n' "$(hostname)"
    printf 'remote:  %s:%s/%s\n' "$R2_REMOTE" "$R2_BUCKET" "$R2_MONTHLY_PREFIX"
    printf 'reason:  %s\n' "$msg"
  } > "$FAILURE_FLAG"
  exit 1
}

[ -d "$BACKUP_DIR" ] || die "backup dir not found: $BACKUP_DIR"
command -v rclone >/dev/null 2>&1 || die "rclone not installed"

exec 9>"$LOCK_FILE"
flock -n 9 || die "another monthly R2 anchor sync is already running (lock: $LOCK_FILE)"

mapfile -t months < <(find "$BACKUP_DIR" -maxdepth 1 -type f -name 'prodcal-*.sqlite3.gz' -printf '%f\n' \
  | sed -E 's/^prodcal-([0-9]{6}).*$/\1/' \
  | sort -u)

[ "${#months[@]}" -gt 0 ] || die "no local backup files found in $BACKUP_DIR"

copied=0
for ym in "${months[@]}"; do
  anchor="$(find "$BACKUP_DIR" -maxdepth 1 -type f -name "prodcal-${ym}*.sqlite3.gz" -printf '%f\n' | sort | head -1)"
  [ -n "$anchor" ] || continue
  src="$BACKUP_DIR/$anchor"
  dst="${R2_REMOTE}:${R2_BUCKET}/${R2_MONTHLY_PREFIX}/${anchor}"
  log "copying monthly anchor $anchor -> $dst"
  rclone copyto "$src" "$dst" --stats=0 2>&1 | sed 's/^/  rclone: /' || die "rclone copyto failed for $anchor"
  copied=$((copied + 1))
done

rm -f "$FAILURE_FLAG"
{
  printf 'time:    %s\n' "$(date -u +%FT%TZ)"
  printf 'remote:  %s:%s/%s\n' "$R2_REMOTE" "$R2_BUCKET" "$R2_MONTHLY_PREFIX"
  printf 'anchors: %s\n' "$copied"
} > "$SUCCESS_FLAG"

log "R2 MONTHLY OK anchors=$copied remote=${R2_REMOTE}:${R2_BUCKET}/${R2_MONTHLY_PREFIX}"
