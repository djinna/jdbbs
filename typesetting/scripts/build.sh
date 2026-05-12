#!/bin/bash
# Build script for book production
# Usage: ./scripts/build.sh [command] [options]

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_DIR"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log() { echo -e "${GREEN}[build]${NC} $1"; }
warn() { echo -e "${YELLOW}[warn]${NC} $1"; }
error() { echo -e "${RED}[error]${NC} $1"; exit 1; }

# Convert Word to Typst
convert() {
  local input="$1"
  local output="${2:-src/$(basename "${input%.docx}.typ")}"
  local decisions_file="${3:-${EDGE_DECISIONS_FILE:-}}"

  if [[ ! -f "$input" ]]; then
    error "Input file not found: $input"
  fi

  if [[ -n "$decisions_file" && ! -f "$decisions_file" ]]; then
    error "Decisions file not found: $decisions_file"
  fi

  log "Converting $input → $output"
  mkdir -p "$(dirname "$output")"

  local pandoc_args=(
    "$input"
    --lua-filter=scripts/docx-to-typst-enhanced.lua
    -t typst
    -o "$output"
  )

  local decisions_map_file=""
  if [[ -n "$decisions_file" ]]; then
    log "Applying edge-case decisions from: $decisions_file"
    decisions_map_file="$(mktemp /tmp/edge-decisions-map.XXXXXX.tsv)"

    python3 - "$decisions_file" "$decisions_map_file" << 'PY'
import json
import sys

src, dst = sys.argv[1], sys.argv[2]
with open(src, 'r', encoding='utf-8') as f:
    payload = json.load(f)

decisions = payload.get('decisions', payload if isinstance(payload, list) else [])

SUPPORTED_TYPES = {"manual_list", "colored_text", "highlighted_text"}

with open(dst, 'w', encoding='utf-8') as out:
    for item in decisions:
        if not isinstance(item, dict):
            continue

        item_type = item.get('type')
        if item_type not in SUPPORTED_TYPES:
            continue

        text = item.get('text')
        decision = item.get('decision')
        if not text or not decision:
            continue

        clean_text = str(text).replace('\t', ' ').replace('\n', ' ').strip()
        clean_type = str(item_type).strip().lower()
        clean_decision = str(decision).strip().lower()
        out.write(f"{clean_type}\t{clean_decision}\t{clean_text}\n")
PY

    pandoc_args+=(--metadata="edge-decisions-map-file:$decisions_map_file")
  fi

  pandoc "${pandoc_args[@]}"

  if [[ -n "$decisions_map_file" && -f "$decisions_map_file" ]]; then
    rm -f "$decisions_map_file"
  fi

  log "Created $output"
  warn "Edit the file to fill in title/author metadata"
}

# Generate edge-case review artifacts (HTML + JSON)
review_edges() {
  local input="$1"
  local output="${2:-output/$(basename "${input%.docx}")-edge-review.html}"

  if [[ ! -f "$input" ]]; then
    error "Input file not found: $input"
  fi

  log "Detecting edge cases for $input"
  mkdir -p "$(dirname "$output")"

  python3 scripts/detect-edge-cases.py "$input" -o "$output" --json

  log "Review report: $output"
  log "Raw detection JSON: ${output%.html}.json"
  warn "Open the HTML report, make Keep/Strip/Convert decisions, then export edge_case_decisions.json"
}

# Compile Typst to PDF
compile_pdf() {
  local input="${1:-src/book.typ}"
  local output="${2:-output/$(basename "${input%.typ}.pdf")}"
  
  if [[ ! -f "$input" ]]; then
    error "Input file not found: $input"
  fi
  
  log "Compiling $input → $output"
  mkdir -p "$(dirname "$output")"
  
  typst compile --root . --font-path fonts/ "$input" "$output"
  
  log "Created $output"
}

# Compile Typst to EPUB (via Pandoc)
compile_epub() {
  local input="${1:-src/book.typ}"
  local output="${2:-output/$(basename "${input%.typ}.epub")}"
  
  log "Compiling $input → $output (via intermediate)"
  mkdir -p "$(dirname "$output")"
  
  # Typst doesn't output EPUB directly, so we go through HTML
  # This is a simplified approach; may need refinement
  warn "EPUB export is experimental"
  
  # For now, suggest using the CSS directly with original docx
  log "For EPUB, consider: pandoc manuscript.docx --css=templates/epub/epub-styles.css -o output/book.epub"
}

# Preview with live reload
preview() {
  local input="${1:-src/book.typ}"
  
  log "Starting preview for $input"
  typst watch --font-path fonts/ "$input"
}

# Full pipeline: docx → typ → pdf
full() {
  local docx="$1"
  local decisions_file="${2:-${EDGE_DECISIONS_FILE:-}}"

  if [[ -z "$docx" ]]; then
    error "Usage: ./scripts/build.sh full manuscript.docx [edge_case_decisions.json]"
  fi

  local basename="$(basename "${docx%.docx}")"
  local typ="src/${basename}.typ"
  local pdf="output/${basename}.pdf"

  convert "$docx" "$typ" "$decisions_file"
  compile_pdf "$typ" "$pdf"

  log "Complete! Output: $pdf"
}

# Show help
help() {
  cat << HELP
Book Production Build Script

Usage: ./scripts/build.sh [command] [options]

Commands:
  convert <input.docx> [output.typ] [decisions.json]
                                      Convert Word to Typst (optional edge decisions)
  review <input.docx> [output.html]   Detect edge cases and generate review artifacts
  compile <input.typ> [output.pdf]    Compile Typst to PDF
  epub <input.typ> [output.epub]      Compile to EPUB (experimental)
  preview <input.typ>                 Live preview with auto-reload
  full <input.docx> [decisions.json]  Full pipeline: docx → typ → pdf
  help                                Show this help

Examples:
  ./scripts/build.sh review manuscripts/novel.docx
  ./scripts/build.sh convert manuscripts/novel.docx
  ./scripts/build.sh convert manuscripts/novel.docx src/novel.typ edge_case_decisions.json
  EDGE_DECISIONS_FILE=edge_case_decisions.json ./scripts/build.sh full manuscripts/novel.docx
  ./scripts/build.sh preview src/novel.typ

HELP
}

# Main dispatch
case "${1:-help}" in
  convert)  convert "$2" "$3" "$4" ;;
  review)   review_edges "$2" "$3" ;;
  compile)  compile_pdf "$2" "$3" ;;
  epub)     compile_epub "$2" "$3" ;;
  preview)  preview "$2" ;;
  full)     full "$2" "$3" ;;
  help|*)   help ;;
esac
