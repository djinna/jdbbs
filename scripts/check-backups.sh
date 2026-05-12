#!/usr/bin/env bash
# Backup health check. Pairs with backup-db.sh + sync-to-r2.sh.
#
# Examines the four sentinel files those scripts maintain and exits
# non-zero with a clear summary if anything is wrong. Intended for an
# hourly cron, with output appended to ~/backups/backup-health.log so
# silently-broken backups are caught within an hour rather than a week.
#
# Sentinels (all in $BACKUP_DIR):
#   .LAST-SUCCESS       — written by backup-db.sh on every successful run
#   .LAST-FAILURE       — written by backup-db.sh when anything went wrong
#   .LAST-R2-SUCCESS    — written by sync-to-r2.sh on successful push
#   .LAST-R2-FAILURE    — written by sync-to-r2.sh when the push failed
#
# Health rules:
#   - .LAST-FAILURE / .LAST-R2-FAILURE present -> FAIL (loud)
#   - .LAST-SUCCESS missing or > MAX_AGE_HOURS old -> FAIL
#   - .LAST-R2-SUCCESS missing or > MAX_AGE_HOURS old -> FAIL
#   - everything fine -> OK + drops .HEALTH-OK sentinel so the Mac side
#     can show the last health-check time without re-opening logs.

set -euo pipefail

BACKUP_DIR="${BACKUP_DIR:-$HOME/backups}"
MAX_AGE_HOURS="${MAX_AGE_HOURS:-26}"    # daily cron at 03:00 UTC -> 26h gives ~2h slack
HEALTH_FLAG="${BACKUP_DIR}/.HEALTH-OK"
HEALTH_FAIL_FLAG="${BACKUP_DIR}/.HEALTH-FAIL"

NOW="$(date -u +%s)"
problems=()

# Pretty-print seconds as "Hh Mm".
fmt_age() {
  local s=$1
  printf '%dh %02dm' $((s / 3600)) $(((s % 3600) / 60))
}

mtime() {
  if [ -f "$1" ]; then stat -c '%Y' "$1" 2>/dev/null || stat -f '%m' "$1"; else echo ""; fi
}

# 1. Failure sentinels must be absent.
for f in .LAST-FAILURE .LAST-R2-FAILURE; do
  if [ -f "${BACKUP_DIR}/${f}" ]; then
    reason="$(grep '^reason:' "${BACKUP_DIR}/${f}" | head -1 | sed 's/^reason: *//')"
    problems+=("${f}: ${reason:-(no reason given)}")
  fi
done

# 2. Success sentinels must exist and be fresh.
for f in .LAST-SUCCESS .LAST-R2-SUCCESS; do
  path="${BACKUP_DIR}/${f}"
  if [ ! -f "$path" ]; then
    problems+=("${f}: MISSING — has backup-db.sh / sync-to-r2.sh ever succeeded?")
    continue
  fi
  m="$(mtime "$path")"
  age_s=$(( NOW - m ))
  age_h=$(( age_s / 3600 ))
  if [ "$age_h" -gt "$MAX_AGE_HOURS" ]; then
    problems+=("${f}: STALE — last updated $(fmt_age $age_s) ago (max ${MAX_AGE_HOURS}h)")
  fi
done

# Report.
NOW_HUMAN="$(date -u +%FT%TZ)"
if [ ${#problems[@]} -eq 0 ]; then
  rm -f "$HEALTH_FAIL_FLAG"
  {
    printf 'time:        %s\n' "$NOW_HUMAN"
    printf 'status:      OK\n'
    [ -f "${BACKUP_DIR}/.LAST-SUCCESS" ] && \
      printf 'last_backup: %s\n' "$(grep '^time:' "${BACKUP_DIR}/.LAST-SUCCESS" | head -1 | awk '{print $2}')"
    [ -f "${BACKUP_DIR}/.LAST-R2-SUCCESS" ] && \
      printf 'last_r2:     %s\n' "$(grep '^time:' "${BACKUP_DIR}/.LAST-R2-SUCCESS" | head -1 | awk '{print $2}')"
  } > "$HEALTH_FLAG"
  printf '[%s] HEALTH OK\n' "$NOW_HUMAN"
  exit 0
fi

rm -f "$HEALTH_FLAG"
{
  printf 'time:     %s\n' "$NOW_HUMAN"
  printf 'host:     %s\n' "$(hostname)"
  printf 'status:   FAIL\n'
  printf 'problems:\n'
  for p in "${problems[@]}"; do
    printf '  - %s\n' "$p"
  done
} > "$HEALTH_FAIL_FLAG"

printf '[%s] HEALTH FAIL (%d problems):\n' "$NOW_HUMAN" "${#problems[@]}" >&2
for p in "${problems[@]}"; do
  printf '  - %s\n' "$p" >&2
done
exit 1
