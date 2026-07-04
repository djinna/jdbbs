#!/usr/bin/env bash
# Daily SQLite backup for prodcal with integrity probes.
#
# Behaviour:
#   * Uses sqlite3 .backup for a consistent snapshot (safe with WAL mode).
#   * gzip the result.
#   * Verify the backup is large enough to plausibly contain real data.
#   * Decompress to a tmp file and probe projects + books rowcounts.
#   * Compare size to the previous backup; warn (not fail) on >50% drift.
#   * Retention: daily backups for 30 days, plus one per month indefinitely.
#   * Drop a ~/backups/.LAST-FAILURE sentinel if anything goes wrong so
#     downstream monitoring / out-of-band checks can notice silently
#     broken cron runs.
#
# History: prior to 2026-05-12 this script silently produced 3.5 KB
# "successful" backups for a week after a WorkingDirectory change moved
# the live DB but not the script's hardcoded path. See TRK-OPS-007.

set -euo pipefail

# ---------------- config ----------------
BACKUP_DIR="${BACKUP_DIR:-$HOME/backups}"
DB_PATH="${DB_PATH:-/home/exedev/prodcal/db.sqlite3}"
TIMESTAMP="$(date +%Y%m%d-%H%M%S)"
BACKUP_FILE="${BACKUP_DIR}/prodcal-${TIMESTAMP}.sqlite3"
GZ_FILE="${BACKUP_FILE}.gz"
LOCK_FILE="${BACKUP_DIR}/.lock"
FAILURE_FLAG="${BACKUP_DIR}/.LAST-FAILURE"
SUCCESS_FLAG="${BACKUP_DIR}/.LAST-SUCCESS"
RETAIN_DAILY_DAYS="${RETAIN_DAILY_DAYS:-30}"

# Probe thresholds. These are sized for the current production data
# (12 projects, 6 books — Twitter Years source DOCXs + compiled PDFs
# total ~60 MB). Adjust if the legitimate baseline shifts.
MIN_GZ_BYTES="${MIN_GZ_BYTES:-1048576}"   # 1 MB — anything smaller is suspect
MIN_PROJECTS="${MIN_PROJECTS:-1}"
MIN_BOOKS="${MIN_BOOKS:-1}"
DRIFT_WARN_PCT="${DRIFT_WARN_PCT:-50}"    # warn only; legitimate growth can exceed this

# ---------------- helpers ----------------
die() {
  local msg="$1"
  printf '[%s] FAIL: %s\n' "$(date -u +%FT%TZ)" "$msg" >&2
  # Quarantine any partially-produced artifact. A backup that reaches die()
  # has passed no (or not all) probes, so it must never remain as a valid
  # prodcal-*.sqlite3.gz that sync-to-r2 / restore-drill / prune / backup_status
  # would treat as the newest good backup. Renaming to .BAD drops it from the
  # prodcal-*.sqlite3.gz glob; also clear the uncompressed intermediate.
  if [ -n "${GZ_FILE:-}" ] && [ -f "$GZ_FILE" ]; then
    mv -f "$GZ_FILE" "${GZ_FILE}.BAD" 2>/dev/null || true
  fi
  if [ -n "${BACKUP_FILE:-}" ] && [ -f "$BACKUP_FILE" ]; then
    rm -f "$BACKUP_FILE" 2>/dev/null || true
  fi
  mkdir -p "$(dirname "$FAILURE_FLAG")"
  {
    printf 'time:        %s\n' "$(date -u +%FT%TZ)"
    printf 'script:      %s\n' "$0"
    printf 'host:        %s\n' "$(hostname)"
    printf 'db_path:     %s\n' "$DB_PATH"
    printf 'backup_file: %s\n' "${GZ_FILE:-<unset>}"
    printf 'reason:      %s\n' "$msg"
  } > "$FAILURE_FLAG"
  exit 1
}

cleanup() {
  rm -f "$PROBE_FILE" 2>/dev/null || true
}
PROBE_FILE=""
trap cleanup EXIT

log() { printf '[%s] %s\n' "$(date -u +%FT%TZ)" "$*"; }

stat_size() {
  # Portable file size in bytes.
  if stat -c '%s' "$1" 2>/dev/null; then return; fi
  stat -f '%z' "$1" 2>/dev/null || echo 0
}

mkdir -p "$BACKUP_DIR"

# ---------------- single-instance lock ----------------
exec 9>"$LOCK_FILE"
flock -n 9 || die "another backup is already running (lock: $LOCK_FILE)"

# ---------------- preflight ----------------
[ -f "$DB_PATH" ] || die "source DB not found at $DB_PATH"
src_bytes="$(stat_size "$DB_PATH")"
[ "$src_bytes" -gt 0 ] || die "source DB at $DB_PATH is zero bytes"
log "starting backup: $DB_PATH ($(numfmt --to=iec --suffix=B "$src_bytes" 2>/dev/null || echo "${src_bytes}B"))"

# ---------------- take backup ----------------
sqlite3 "$DB_PATH" ".backup '${BACKUP_FILE}'" \
  || die "sqlite3 .backup failed for $DB_PATH"
[ -f "$BACKUP_FILE" ] || die "backup file $BACKUP_FILE was not created"

gzip "$BACKUP_FILE" \
  || die "gzip failed on $BACKUP_FILE"
[ -f "$GZ_FILE" ] || die "compressed file $GZ_FILE missing after gzip"

gz_bytes="$(stat_size "$GZ_FILE")"
log "wrote $GZ_FILE ($(numfmt --to=iec --suffix=B "$gz_bytes" 2>/dev/null || echo "${gz_bytes}B"))"

# ---------------- size probe ----------------
if [ "$gz_bytes" -lt "$MIN_GZ_BYTES" ]; then
  die "backup size ${gz_bytes}B < MIN_GZ_BYTES (${MIN_GZ_BYTES}B). The DB or its path is likely wrong."
fi

# ---------------- rowcount probe ----------------
PROBE_FILE="$(mktemp -t backup-probe.XXXXXX.sqlite3)"
zcat "$GZ_FILE" > "$PROBE_FILE" \
  || die "could not decompress $GZ_FILE for probe"

projects_n="$(sqlite3 "$PROBE_FILE" 'SELECT COUNT(*) FROM projects;' 2>&1)" \
  || die "could not read projects table from backup: $projects_n"
books_n="$(sqlite3 "$PROBE_FILE" 'SELECT COUNT(*) FROM books;' 2>&1)" \
  || die "could not read books table from backup: $books_n"
integrity="$(sqlite3 "$PROBE_FILE" 'PRAGMA integrity_check;' 2>&1)" \
  || die "PRAGMA integrity_check failed on backup: $integrity"

log "probe: projects=$projects_n books=$books_n integrity=$integrity"

[ "$projects_n" -ge "$MIN_PROJECTS" ] \
  || die "backup has $projects_n projects, MIN_PROJECTS=$MIN_PROJECTS"
[ "$books_n" -ge "$MIN_BOOKS" ] \
  || die "backup has $books_n books, MIN_BOOKS=$MIN_BOOKS"
[ "$integrity" = "ok" ] \
  || die "integrity check returned '$integrity' (expected 'ok')"

# ---------------- drift comparison vs previous backup (warn only) ----------------
prev_gz="$(find "$BACKUP_DIR" -maxdepth 1 -name 'prodcal-*.sqlite3.gz' \
            ! -name "$(basename "$GZ_FILE")" -printf '%T@ %p\n' 2>/dev/null \
            | sort -rn | head -1 | cut -d' ' -f2- || true)"
if [ -n "${prev_gz:-}" ] && [ -f "${prev_gz}" ]; then
  prev_bytes="$(stat_size "$prev_gz")"
  if [ "$prev_bytes" -gt 0 ]; then
    diff_pct=$(( (gz_bytes > prev_bytes ? gz_bytes - prev_bytes : prev_bytes - gz_bytes) * 100 / prev_bytes ))
    log "previous backup: $(basename "$prev_gz") (${prev_bytes}B), drift=${diff_pct}%"
    if [ "$diff_pct" -gt "$DRIFT_WARN_PCT" ]; then
      log "WARN: backup size drift ${diff_pct}% exceeds ${DRIFT_WARN_PCT}% (was ${prev_bytes}B, now ${gz_bytes}B). Inspect manually if unexpected."
    fi
  fi
fi

# ---------------- retention ----------------
# Daily tier: delete >$RETAIN_DAILY_DAYS days old.
# Monthly tier: keep the FIRST backup of each calendar month indefinitely.
log "retention: keeping daily backups for $RETAIN_DAILY_DAYS days, plus first-of-month indefinitely"

# Build a set of files to preserve as "monthly anchor": for each YYYYMM
# present in $BACKUP_DIR, pick the oldest-timestamped file.
preserve_list="$(mktemp)"
trap 'cleanup; rm -f "$preserve_list"' EXIT

for ym in $(find "$BACKUP_DIR" -maxdepth 1 -name 'prodcal-*.sqlite3.gz' -printf '%f\n' \
            | sed -E 's/^prodcal-([0-9]{6}).*$/\1/' | sort -u); do
  anchor="$(find "$BACKUP_DIR" -maxdepth 1 -name "prodcal-${ym}*.sqlite3.gz" -printf '%f\n' \
            | sort | head -1)"
  [ -n "$anchor" ] && echo "$BACKUP_DIR/$anchor" >> "$preserve_list"
done

# Then delete anything older than RETAIN_DAILY_DAYS that isn't in the preserve list.
while IFS= read -r f; do
  if ! grep -qxF "$f" "$preserve_list"; then
    log "retain: removing old daily $f"
    rm -f "$f"
  fi
done < <(find "$BACKUP_DIR" -maxdepth 1 -name 'prodcal-*.sqlite3.gz' -mtime "+${RETAIN_DAILY_DAYS}" -print)

# ---------------- success ----------------
rm -f "$FAILURE_FLAG"
{
  printf 'time:        %s\n' "$(date -u +%FT%TZ)"
  printf 'backup_file: %s\n' "$GZ_FILE"
  printf 'size_bytes:  %s\n' "$gz_bytes"
  printf 'projects:    %s\n' "$projects_n"
  printf 'books:       %s\n' "$books_n"
} > "$SUCCESS_FLAG"

printf '[%s] OK backup_file=%s size_bytes=%s projects=%s books=%s\n' \
  "$(date -u +%FT%TZ)" "$GZ_FILE" "$gz_bytes" "$projects_n" "$books_n"
