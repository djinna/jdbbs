#!/usr/bin/env bash
# Prune local ProdCal SQLite backups without affecting off-VM R2 copies.
#
# Default policy:
#   * keep the newest 10 local backups for fast restores;
#   * keep the first backup from each calendar month as a local anchor;
#   * delete only files matching prodcal-*.sqlite3.gz in BACKUP_DIR.
#
# Intended to run weekly from cron. It is also safe to run manually:
#   DRY_RUN=1 scripts/prune-backups.sh

set -euo pipefail

BACKUP_DIR="${BACKUP_DIR:-$HOME/backups}"
# KEEP_RECENT is a LOCAL fast-restore cache size, NOT the retention policy.
# Authoritative retention is owned by backup-db.sh (daily backups kept 30 days
# + the first backup of each month indefinitely) and mirrored off-VM in R2.
# This prune only trims the on-disk copy: 10 recent + monthly anchors is enough
# for a quick local restore without pinning the full 30-day set to this VM.
KEEP_RECENT="${KEEP_RECENT:-10}"
KEEP_MONTHLY="${KEEP_MONTHLY:-1}"
DRY_RUN="${DRY_RUN:-0}"
LOCK_FILE="${BACKUP_DIR}/.prune-lock"
FAILURE_FLAG="${BACKUP_DIR}/.LAST-FAILURE"

log() { printf '[%s] %s\n' "$(date -u +%FT%TZ)" "$*"; }
die() { log "FAIL: $*" >&2; exit 1; }

[ -d "$BACKUP_DIR" ] || die "backup dir not found: $BACKUP_DIR"
[[ "$KEEP_RECENT" =~ ^[0-9]+$ ]] || die "KEEP_RECENT must be an integer"

# Never prune during a backup failure streak. backup-db.sh drops a
# .LAST-FAILURE sentinel when a run fails its probes; aging out the last
# known-good local daily while fresh backups are broken could leave no local
# restore point. Bail out cleanly until the next good backup clears the flag.
if [ -f "$FAILURE_FLAG" ]; then
  log "refusing to prune: $FAILURE_FLAG present (backup failure streak); leaving local backups intact"
  exit 0
fi

exec 9>"$LOCK_FILE"
flock -n 9 || die "another prune is already running (lock: $LOCK_FILE)"

# Deletion candidates: ONLY auto-generated, timestamped backups
# (prodcal-YYYYMMDD-HHMMSS.sqlite3.gz). Manual snapshots such as
# prodcal-pre-migration.sqlite3.gz do not match this pattern and are never swept.
mapfile -t backups < <(find "$BACKUP_DIR" -maxdepth 1 -type f -name 'prodcal-*.sqlite3.gz' -printf '%f\n' | grep -E '^prodcal-[0-9]{8}-[0-9]{6}\.sqlite3\.gz$' | sort)
if [ "${#backups[@]}" -eq 0 ]; then
  log "no prodcal backup files found in $BACKUP_DIR"
  exit 0
fi

preserve_file="$(mktemp -t prodcal-prune-preserve.XXXXXX)"
all_file="$(mktemp -t prodcal-prune-all.XXXXXX)"
trap 'rm -f "$preserve_file" "$all_file"' EXIT

printf '%s\n' "${backups[@]/#/$BACKUP_DIR/}" > "$all_file"

# Keep newest N backups by timestamped filename.
if [ "$KEEP_RECENT" -gt 0 ]; then
  printf '%s\n' "${backups[@]}" \
    | sort -r \
    | head -n "$KEEP_RECENT" \
    | sed "s#^#$BACKUP_DIR/#" \
    >> "$preserve_file"
fi

# Keep the first backup from each calendar month as an anchor.
if [ "$KEEP_MONTHLY" = "1" ]; then
  printf '%s\n' "${backups[@]}" \
    | sed -E 's/^prodcal-([0-9]{6}).*$/\1/' \
    | sort -u \
    | while IFS= read -r ym; do
        anchor="$(printf '%s\n' "${backups[@]}" | grep -E "^prodcal-${ym}[0-9]{2}-[0-9]{6}\.sqlite3\.gz$" | sort | head -1 || true)"
        [ -n "$anchor" ] && printf '%s/%s\n' "$BACKUP_DIR" "$anchor"
      done \
    >> "$preserve_file"
fi

sort -u -o "$preserve_file" "$preserve_file"

kept=0
deleted=0
freed=0
while IFS= read -r f; do
  if grep -qxF "$f" "$preserve_file"; then
    kept=$((kept + 1))
    continue
  fi

  size="$(stat -c '%s' "$f" 2>/dev/null || echo 0)"
  if [ "$DRY_RUN" = "1" ]; then
    log "would remove $(basename "$f") (${size} bytes)"
  else
    log "removing $(basename "$f") (${size} bytes)"
    rm -f -- "$f"
  fi
  deleted=$((deleted + 1))
  freed=$((freed + size))
done < "$all_file"

if [ "$DRY_RUN" = "1" ]; then
  log "DRY RUN complete: would_delete=$deleted would_free_bytes=$freed keep=$kept total=${#backups[@]}"
else
  log "OK deleted=$deleted freed_bytes=$freed keep=$kept total=${#backups[@]}"
fi
