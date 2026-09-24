#!/usr/bin/env bash
# Create this host's R2 backup namespace once. Refuses to overwrite.
#   scripts/r2-init-namespace.sh            -> db-<hostname>-<8 hex>
#   scripts/r2-init-namespace.sh my-prefix  -> explicit value
set -euo pipefail
BACKUP_DIR="${BACKUP_DIR:-$HOME/backups}"
FILE="${R2_NAMESPACE_FILE:-$BACKUP_DIR/.r2-namespace}"
if [ -f "$FILE" ]; then
  printf 'namespace already set: %s (%s)\n' "$(cat "$FILE")" "$FILE" >&2
  exit 1
fi
name="${1:-db-$(hostname | tr -c 'a-zA-Z0-9-\n' '-' | tr '[:upper:]' '[:lower:]')-$(openssl rand -hex 4)}"
mkdir -p "$BACKUP_DIR"
umask 077
printf '%s\n%s\n' "$name" "$(hostname)" > "$FILE"
printf 'R2 namespace for this host: %s\nwritten to %s\n' "$name" "$FILE"
