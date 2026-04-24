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
  
  if [[ ! -f "$input" ]]; then
    error "Input file not found: $input"
  fi
  
  log "Converting $input → $output"
  mkdir -p "$(dirname "$output")"
  
  pandoc "$input" \
    --lua-filter=scripts/docx-to-typst.lua \
    -t typst \
    -o "$output"
  
  log "Created $output"
  warn "Edit the file to fill in title/author metadata"
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
  
  typst compile --font-path fonts/ "$input" "$output"
  
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
  
  if [[ -z "$docx" ]]; then
    error "Usage: ./scripts/build.sh full manuscript.docx"
  fi
  
  local basename="$(basename "${docx%.docx}")"
  local typ="src/${basename}.typ"
  local pdf="output/${basename}.pdf"
  
  convert "$docx" "$typ"
  compile_pdf "$typ" "$pdf"
  
  log "Complete! Output: $pdf"
}

# Show help
help() {
  cat << HELP
Book Production Build Script

Usage: ./scripts/build.sh [command] [options]

Commands:
  convert <input.docx> [output.typ]   Convert Word to Typst
  compile <input.typ> [output.pdf]    Compile Typst to PDF  
  epub <input.typ> [output.epub]      Compile to EPUB (experimental)
  preview <input.typ>                 Live preview with auto-reload
  full <input.docx>                   Full pipeline: docx → typ → pdf
  help                                Show this help

Examples:
  ./scripts/build.sh convert manuscripts/novel.docx
  ./scripts/build.sh compile src/novel.typ
  ./scripts/build.sh full manuscripts/novel.docx
  ./scripts/build.sh preview src/novel.typ

HELP
}

# Main dispatch
case "${1:-help}" in
  convert)  convert "$2" "$3" ;;
  compile)  compile_pdf "$2" "$3" ;;
  epub)     compile_epub "$2" "$3" ;;
  preview)  preview "$2" ;;
  full)     full "$2" ;;
  help|*)   help ;;
esac
