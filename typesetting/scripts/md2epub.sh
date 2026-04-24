#!/bin/bash
# Convert Markdown to EPUB3
# Usage: ./scripts/md2epub.sh input.md [output.epub]
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

if [ $# -lt 1 ]; then
    echo "Usage: $0 <input.md> [output.epub] [--title TITLE] [--author AUTHOR]"
    exit 1
fi

INPUT="$1"
shift

BASENAME="$(basename "${INPUT%.md}")"
OUTPUT="$PROJECT_ROOT/output/${BASENAME}.epub"
TITLE="$BASENAME"
AUTHOR=""

# Parse optional args
while [[ $# -gt 0 ]]; do
    case $1 in
        --title)
            TITLE="$2"
            shift 2
            ;;
        --author)
            AUTHOR="$2"
            shift 2
            ;;
        *.epub)
            OUTPUT="$1"
            shift
            ;;
        *)
            shift
            ;;
    esac
done

echo "Converting: $INPUT → $OUTPUT"

# Build pandoc args
ARGS=(
    -f markdown
    -t epub3
    --css "$PROJECT_ROOT/templates/epub/epub-styles.css"
    --split-level=1
    --toc
    --toc-depth=2
    -o "$OUTPUT"
)

# Add metadata if provided
if [ -n "$TITLE" ]; then
    ARGS+=(--metadata "title=$TITLE")
fi
if [ -n "$AUTHOR" ]; then
    ARGS+=(--metadata "author=$AUTHOR")
fi

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

echo "  Markdown → EPUB3..."
pandoc "${ARGS[@]}" "$INPUT"

echo "Done: $OUTPUT"
ls -la "$OUTPUT"
