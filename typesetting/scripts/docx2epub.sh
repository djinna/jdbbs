#!/bin/bash
# Convert Word DOCX to EPUB3
# Usage: ./scripts/docx2epub.sh input.docx [output.epub]
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

if [ $# -lt 1 ]; then
    echo "Usage: $0 <input.docx> [output.epub]"
    exit 1
fi

INPUT="$1"
BASENAME="$(basename "${INPUT%.docx}")"
OUTPUT="${2:-$PROJECT_ROOT/output/${BASENAME}.epub}"

echo "Converting: $INPUT → $OUTPUT"

# Build pandoc args
ARGS=(
    -f docx
    -t epub3
    --css "$PROJECT_ROOT/templates/epub/epub-styles.css"
    --split-level=1
    --toc
    --toc-depth=2
    -o "$OUTPUT"
)

# Add embedded fonts if available
if [ -d "$PROJECT_ROOT/fonts/sourcesans/WOFF2" ]; then
    for font in "$PROJECT_ROOT/fonts/sourcesans/WOFF2/"*.woff2; do
        [ -f "$font" ] && ARGS+=(--epub-embed-font "$font")
    done
fi
if [ -d "$PROJECT_ROOT/fonts/jetbrainsmono/fonts/webfonts" ]; then
    for font in "$PROJECT_ROOT/fonts/jetbrainsmono/fonts/webfonts/"*.woff2; do
        [ -f "$font" ] && ARGS+=(--epub-embed-font "$font")
    done
fi

echo "  DOCX → EPUB3..."
pandoc "${ARGS[@]}" "$INPUT"

echo "Done: $OUTPUT"
ls -la "$OUTPUT"
