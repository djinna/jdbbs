#!/usr/bin/env bash
# Monthly R2 restore drill. Pairs with sync-to-r2.sh and restore-drill.sh.
#
# restore-drill.sh proves the newest LOCAL backup restores; this script
# proves the off-VM copy does too — the actual disaster path. Pulls the
# newest daily backup object from R2 (rclone, same remote/bucket/prefix
# sync-to-r2.sh ships to), decompresses it to a tmp file, and runs
# PRAGMA integrity_check plus a sanity rowcount against the result.
#
# Writes $BACKUP_DIR/.LAST-R2-DRILL-SUCCESS on success and
# $BACKUP_DIR/.LAST-R2-DRILL-FAILURE on failure (both read by
# /api/admin/backup-status monitoring).
#
# Cron monthly (e.g., 30 4 1 * * — 04:30 UTC on the 1st of each month,
# after the 04:00 local restore drill has finished).

set -euo pipefail

BACKUP_DIR="${BACKUP_DIR:-$HOME/backups}"
R2_REMOTE="${R2_REMOTE:-r2}"
R2_BUCKET="${R2_BUCKET:-jdbbs-backups}"
R2_PREFIX="${R2_PREFIX:-db}"
MIN_PROJECTS="${MIN_PROJECTS:-1}"
SUCCESS_FLAG="${BACKUP_DIR}/.LAST-R2-DRILL-SUCCESS"
FAILURE_FLAG="${BACKUP_DIR}/.LAST-R2-DRILL-FAILURE"

die() {
  local msg="$1"
  printf '[%s] R2 DRILL FAIL: %s\n' "$(date -u +%FT%TZ)" "$msg" >&2
  mkdir -p "$(dirname "$FAILURE_FLAG")"
  {
    printf 'time:    %s\n' "$(date -u +%FT%TZ)"
    printf 'host:    %s\n' "$(hostname)"
    printf 'remote:  %s:%s/%s\n' "$R2_REMOTE" "$R2_BUCKET" "$R2_PREFIX"
    printf 'reason:  %s\n' "$msg"
    [ -n "${OBJECT:-}" ] && printf 'object:  %s\n' "$OBJECT"
  } > "$FAILURE_FLAG"
  exit 1
}

log() { printf '[%s] %s\n' "$(date -u +%FT%TZ)" "$*"; }

cleanup() {
  rm -f "${PULL_FILE:-}" "${PROBE_FILE:-}"
}
PULL_FILE=""
PROBE_FILE=""
OBJECT=""
trap cleanup EXIT

command -v rclone >/dev/null 2>&1 || die "rclone not installed"
command -v sqlite3 >/dev/null 2>&1 || die "sqlite3 not installed"

rclone listremotes 2>/dev/null | grep -qx "${R2_REMOTE}:" \
  || die "rclone remote '$R2_REMOTE' is not configured. Run \`rclone config\` first."

# Newest daily backup object. Names embed the timestamp
# (prodcal-YYYYMMDD-HHMMSS.sqlite3.gz), so lexical order is chronological
# order; the prodcal-20 prefix (rclone filter + grep belt-and-suspenders)
# keeps prodcal-monthly-* anchors out.
OBJECT="$(rclone lsf --files-only --include 'prodcal-20*.sqlite3.gz' \
          "${R2_REMOTE}:${R2_BUCKET}/${R2_PREFIX}/" 2>/dev/null \
          | grep '^prodcal-20.*\.sqlite3\.gz$' | LC_ALL=C sort | tail -1 || true)"
[ -n "$OBJECT" ] \
  || die "no backup objects found at ${R2_REMOTE}:${R2_BUCKET}/${R2_PREFIX}/ (credentials wrong, bucket missing, or nothing synced yet)"

log "drill object: ${R2_REMOTE}:${R2_BUCKET}/${R2_PREFIX}/${OBJECT}"

PULL_FILE="$(mktemp -t r2-restore-drill.XXXXXX.sqlite3.gz)"
rclone copyto "${R2_REMOTE}:${R2_BUCKET}/${R2_PREFIX}/${OBJECT}" "$PULL_FILE" --stats=0 \
  || die "rclone copyto failed for $OBJECT"

PROBE_FILE="$(mktemp -t r2-restore-drill.XXXXXX.sqlite3)"
gunzip -c "$PULL_FILE" > "$PROBE_FILE" \
  || die "could not decompress $OBJECT"

# 1. integrity_check
integ="$(sqlite3 "$PROBE_FILE" 'PRAGMA integrity_check;' | head -1)" \
  || die "integrity_check query failed (pulled object is not a SQLite DB?)"
[ "$integ" = "ok" ] || die "integrity_check returned '$integ'"
log "integrity check: ok"

# 2. Sanity rowcount — a restore-worthy backup has real projects in it.
projects_n="$(sqlite3 "$PROBE_FILE" 'SELECT COUNT(*) FROM projects;')" \
  || die "rowcount query failed (projects table missing?)"
[ "$projects_n" -ge "$MIN_PROJECTS" ] || die "projects=$projects_n < MIN_PROJECTS=$MIN_PROJECTS"
log "rowcount check: projects=$projects_n"

# Success.
rm -f "$FAILURE_FLAG"
{
  printf 'time:       %s\n' "$(date -u +%FT%TZ)"
  printf 'object:     %s:%s/%s/%s\n' "$R2_REMOTE" "$R2_BUCKET" "$R2_PREFIX" "$OBJECT"
  printf 'projects:   %s\n' "$projects_n"
  printf 'integrity:  %s\n' "$integ"
} > "$SUCCESS_FLAG"

printf '[%s] R2 DRILL OK object=%s projects=%s\n' \
  "$(date -u +%FT%TZ)" "$OBJECT" "$projects_n"
