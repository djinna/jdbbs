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

extract_title() {
    python3 - "$INPUT" "$BASENAME" <<'PY'
import sys
from docx import Document

input_path = sys.argv[1]
default_title = sys.argv[2]

try:
    doc = Document(input_path)
except Exception:
    print(default_title)
    raise SystemExit(0)

core_title = (doc.core_properties.title or "").strip()
if core_title:
    print(core_title)
    raise SystemExit(0)

for paragraph in doc.paragraphs:
    text = (paragraph.text or "").strip()
    if text:
        print(text)
        raise SystemExit(0)

print(default_title)
PY
}

EPUB_TITLE="${EPUB_TITLE:-$(extract_title)}"

echo "Converting: $INPUT → $OUTPUT"
echo "  EPUB title: $EPUB_TITLE"

# Build pandoc args
ARGS=(
    -f docx
    -t epub3
    --metadata "title=$EPUB_TITLE"
    --css "$PROJECT_ROOT/templates/epub/epub-styles.css"
    --split-level=1
    --toc
    --toc-depth=2
    -o "$OUTPUT"
)

# EPUB font policy: do not embed fonts by default (small file sizes, reader defaults).
# Optional override for exceptional cases:
#   EPUB_EMBED_FONTS=1 ./scripts/docx2epub.sh input.docx output.epub
if [ "${EPUB_EMBED_FONTS:-0}" = "1" ]; then
    echo "  WARNING: embedding EPUB fonts (EPUB_EMBED_FONTS=1)"
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
fi

echo "  DOCX → EPUB3..."
pandoc "${ARGS[@]}" "$INPUT"

echo "Done: $OUTPUT"
ls -la "$OUTPUT"
