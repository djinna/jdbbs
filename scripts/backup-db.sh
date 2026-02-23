#!/usr/bin/env bash
# Daily SQLite backup for prodcal
# Keeps 7 days of backups in ~/backups/
set -euo pipefail

BACKUP_DIR="$HOME/backups"
DB_PATH="/home/exedev/prodcal/db.sqlite3"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
BACKUP_FILE="$BACKUP_DIR/prodcal-$TIMESTAMP.sqlite3"

mkdir -p "$BACKUP_DIR"

# Use sqlite3 .backup for a consistent copy (safe with WAL mode)
sqlite3 "$DB_PATH" ".backup '$BACKUP_FILE'"

# Compress
gzip "$BACKUP_FILE"

# Remove backups older than 7 days
find "$BACKUP_DIR" -name 'prodcal-*.sqlite3.gz' -mtime +7 -delete

echo "Backup complete: ${BACKUP_FILE}.gz"
