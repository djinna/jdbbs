#!/usr/bin/env bash
# Build and run the ProdCal local single-user launcher (model A).
#
# Any arguments are passed through to the binary, e.g.:
#   ./scripts/run-local.sh -addr 127.0.0.1:8055 -data ./scratch/local-data
#
# See docs/LOCAL-USAGE.md for details.
set -euo pipefail

cd "$(dirname "$0")/.."

go build -o prodcal-local ./cmd/prodcal-local
exec ./prodcal-local "$@"
