#!/usr/bin/env bash
# Shared Cloudflare R2 settings for the backup scripts. Source, don't run:
#   . "$(dirname "$0")/r2-env.sh"
#
# Why this exists (2026-09-24 security review):
#   * The R2 token is bucket-scoped and cannot create buckets. rclone's
#     single-object `copyto` to a not-yet-seen key issues CreateBucket and
#     gets 403 AccessDenied, which looked like broken credentials. Directory
#     `rclone copy` lists first and never hits it. RCLONE_S3_NO_CHECK_BUCKET
#     stops rclone trying.
#   * A second host holding a copy of this VM's rclone config + crontab was
#     writing a stale database under the shared `db/prodcal-YYYYMMDD-030001`
#     keys at ~03:01 UTC, overwriting this VM's verified upload. Every
#     deployment therefore writes to its own prefix, read from a file created
#     once per host (scripts/r2-init-namespace.sh). Env R2_PREFIX overrides.

BACKUP_DIR="${BACKUP_DIR:-$HOME/backups}"
R2_REMOTE="${R2_REMOTE:-r2}"
R2_BUCKET="${R2_BUCKET:-jdbbs-backups}"
R2_NAMESPACE_FILE="${R2_NAMESPACE_FILE:-$BACKUP_DIR/.r2-namespace}"
export RCLONE_S3_NO_CHECK_BUCKET=true

# File format: line 1 prefix, line 2 hostname that created it. A VM copy
# (`exe.dev cp`) inherits the file but gets a new hostname, so it refuses to
# write into the original's prefix instead of silently clobbering it.
R2_NAMESPACE_HOST=""
if [ -z "${R2_PREFIX:-}" ] && [ -f "$R2_NAMESPACE_FILE" ]; then
  R2_PREFIX="$(sed -n 1p "$R2_NAMESPACE_FILE" | tr -d '[:space:]')"
  R2_NAMESPACE_HOST="$(sed -n 2p "$R2_NAMESPACE_FILE" | tr -d '[:space:]')"
fi

# r2_require_prefix: fail closed rather than fall back to the shared `db/`.
# Callers define die() first.
r2_require_prefix() {
  if [ -z "${R2_PREFIX:-}" ]; then
    die "no R2 prefix: set R2_PREFIX or create $R2_NAMESPACE_FILE (scripts/r2-init-namespace.sh)"
  fi
  if [ -n "$R2_NAMESPACE_HOST" ] && [ "$R2_NAMESPACE_HOST" != "$(hostname)" ]; then
    die "R2 namespace '$R2_PREFIX' belongs to host '$R2_NAMESPACE_HOST', this is '$(hostname)' — a VM copy? Disable cron here or run scripts/r2-init-namespace.sh after removing $R2_NAMESPACE_FILE"
  fi
  case "$R2_PREFIX" in
    */*|.*|"") die "invalid R2 prefix '$R2_PREFIX' (single path segment, no leading dot)";;
  esac
}
