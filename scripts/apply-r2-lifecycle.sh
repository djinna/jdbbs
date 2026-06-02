#!/usr/bin/env bash
# Apply the Cloudflare R2 lifecycle policy for ProdCal backups.
#
# The VM's rclone R2 token is object-scoped and may not have bucket lifecycle
# admin permission. If this script fails with AccessDenied, apply
# scripts/r2-lifecycle.json in the Cloudflare dashboard or rerun with an R2
# admin token that can PutBucketLifecycleConfiguration.

set -euo pipefail

R2_CONFIG_FILE="${R2_CONFIG_FILE:-$HOME/.config/rclone/rclone.conf}"
R2_REMOTE_SECTION="${R2_REMOTE_SECTION:-r2}"
R2_BUCKET="${R2_BUCKET:-jdbbs-backups}"
LIFECYCLE_FILE="${LIFECYCLE_FILE:-$(dirname "$0")/r2-lifecycle.json}"

python3 - "$R2_CONFIG_FILE" "$R2_REMOTE_SECTION" "$R2_BUCKET" "$LIFECYCLE_FILE" <<'PY'
import configparser
import json
import sys

import boto3
import botocore.exceptions

config_path, section, bucket, lifecycle_path = sys.argv[1:]
config = configparser.ConfigParser()
if not config.read(config_path):
    raise SystemExit(f"could not read rclone config: {config_path}")
if section not in config:
    raise SystemExit(f"missing rclone section [{section}] in {config_path}")
r = config[section]
required = ["endpoint", "access_key_id", "secret_access_key"]
missing = [k for k in required if not r.get(k)]
if missing:
    raise SystemExit(f"missing required rclone keys: {', '.join(missing)}")
with open(lifecycle_path, "r", encoding="utf-8") as f:
    lifecycle = json.load(f)

s3 = boto3.client(
    "s3",
    endpoint_url=r["endpoint"],
    aws_access_key_id=r["access_key_id"],
    aws_secret_access_key=r["secret_access_key"],
    region_name=r.get("region", "auto"),
)
try:
    s3.put_bucket_lifecycle_configuration(
        Bucket=bucket,
        LifecycleConfiguration=lifecycle,
    )
except botocore.exceptions.ClientError as e:
    err = e.response.get("Error", {})
    raise SystemExit(f"put lifecycle failed: {err.get('Code')}: {err.get('Message')}")
print(f"applied lifecycle to {bucket}: {json.dumps(lifecycle, separators=(',', ':'))}")
PY
