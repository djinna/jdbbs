#!/usr/bin/env bash
# Shared checklist page on :8766 → https://jdbbs.exe.xyz:8766/
# Source of truth: scratch/run/CHECKLIST.md (edit by hand or via the page's tick boxes).
# Notes from the page land in scratch/run/notes.json. See scripts/runpage/server.py.
set -euo pipefail
cd "$(dirname "$0")"
exec python3 runpage/server.py
