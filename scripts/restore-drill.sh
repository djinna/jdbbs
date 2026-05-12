#!/usr/bin/env bash
# Monthly restore drill. Pairs with backup-db.sh.
#
# Picks the newest local backup, decompresses it to a tmp file, opens
# with sqlite3 and runs a battery of read queries that simulate the
# beginning of a real restore. If any query fails or returns absurd
# numbers, exits non-zero and writes ~/backups/.LAST-DRILL-FAILURE.
#
# Cron monthly (e.g., 0 4 1 * * — 04:00 UTC on the 1st of each month,
# after the 03:00 daily backup has had time to land).

set -euo pipefail

BACKUP_DIR="${BACKUP_DIR:-$HOME/backups}"
MIN_PROJECTS="${MIN_PROJECTS:-1}"
MIN_BOOKS="${MIN_BOOKS:-1}"
EXPECTED_TABLES="${EXPECTED_TABLES:-projects books book_specs book_outputs tasks corrections}"
SUCCESS_FLAG="${BACKUP_DIR}/.LAST-DRILL-SUCCESS"
FAILURE_FLAG="${BACKUP_DIR}/.LAST-DRILL-FAILURE"

die() {
  local msg="$1"
  printf '[%s] DRILL FAIL: %s\n' "$(date -u +%FT%TZ)" "$msg" >&2
  mkdir -p "$(dirname "$FAILURE_FLAG")"
  {
    printf 'time:    %s\n' "$(date -u +%FT%TZ)"
    printf 'host:    %s\n' "$(hostname)"
    printf 'reason:  %s\n' "$msg"
    [ -n "${SOURCE:-}" ] && printf 'source:  %s\n' "$SOURCE"
  } > "$FAILURE_FLAG"
  exit 1
}

log() { printf '[%s] %s\n' "$(date -u +%FT%TZ)" "$*"; }

cleanup() {
  rm -f "${PROBE_FILE:-}"
}
PROBE_FILE=""
trap cleanup EXIT

SOURCE="$(find "$BACKUP_DIR" -maxdepth 1 -name 'prodcal-*.sqlite3.gz' -printf '%T@ %p\n' 2>/dev/null \
          | sort -rn | head -1 | cut -d' ' -f2-)"
[ -n "${SOURCE:-}" ] && [ -f "$SOURCE" ] \
  || die "no backup files found in $BACKUP_DIR"

log "drill source: $SOURCE"

PROBE_FILE="$(mktemp -t restore-drill.XXXXXX.sqlite3)"
gunzip -c "$SOURCE" > "$PROBE_FILE" \
  || die "could not decompress $SOURCE"

# 1. All expected tables present.
existing_tables="$(sqlite3 "$PROBE_FILE" ".tables")"
for t in $EXPECTED_TABLES; do
  echo "$existing_tables" | tr -s ' \n\t' '\n' | grep -qx "$t" \
    || die "expected table '$t' not found in restored DB"
done
log "table check: all of [$EXPECTED_TABLES] present"

# 2. Project + book row counts within sanity bounds.
projects_n="$(sqlite3 "$PROBE_FILE" 'SELECT COUNT(*) FROM projects;')"
books_n="$(sqlite3 "$PROBE_FILE" 'SELECT COUNT(*) FROM books;')"
[ "$projects_n" -ge "$MIN_PROJECTS" ] || die "projects=$projects_n < MIN_PROJECTS=$MIN_PROJECTS"
[ "$books_n" -ge "$MIN_BOOKS" ] || die "books=$books_n < MIN_BOOKS=$MIN_BOOKS"
log "rowcount check: projects=$projects_n books=$books_n"

# 3. integrity_check
integ="$(sqlite3 "$PROBE_FILE" 'PRAGMA integrity_check;' | head -1)"
[ "$integ" = "ok" ] || die "integrity_check returned '$integ'"
log "integrity check: ok"

# 4. Simulate a real-world query — what an admin would run after a
#    restore to confirm "the Twitter Years project survived."
sample="$(sqlite3 "$PROBE_FILE" \
  "SELECT id, name FROM projects ORDER BY id LIMIT 5;" 2>&1)"
log "sample projects (first 5):"
printf '%s\n' "$sample" | sed 's/^/    /'

# Success.
rm -f "$FAILURE_FLAG"
{
  printf 'time:        %s\n' "$(date -u +%FT%TZ)"
  printf 'source:      %s\n' "$SOURCE"
  printf 'projects:    %s\n' "$projects_n"
  printf 'books:       %s\n' "$books_n"
  printf 'integrity:   %s\n' "$integ"
} > "$SUCCESS_FLAG"

printf '[%s] DRILL OK source=%s projects=%s books=%s\n' \
  "$(date -u +%FT%TZ)" "$(basename "$SOURCE")" "$projects_n" "$books_n"
