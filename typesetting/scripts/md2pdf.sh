#!/bin/bash
# Convert Markdown to PDF using Pandoc → Typst → PDF
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

if [ $# -lt 1 ]; then
    echo "Usage: $0 <input.md> [output.pdf]"
    exit 1
fi

INPUT="$1"
BASENAME="$(basename "${INPUT%.md}")"
OUTPUT="${2:-$PROJECT_ROOT/output/${BASENAME}.pdf}"
TEMP_TYP="$PROJECT_ROOT/output/${BASENAME}.typ"

echo "Converting: $INPUT → $OUTPUT"

# Step 1: Pandoc Markdown → Typst
echo "  [1/2] Markdown → Typst..."
pandoc -f markdown -t typst "$INPUT" -o "$TEMP_TYP.body"

# Step 2: Wrap with template
cat > "$TEMP_TYP" << 'HEADER'
#import "../templates/series-template.typ": *

// Extract title and author from Pandoc frontmatter if present
#show: book.with(
  title: "Story Collection",
)

// Replace horizontal rule with section break
#let horizontalrule = section-break

// Body content
HEADER

cat "$TEMP_TYP.body" >> "$TEMP_TYP"
rm "$TEMP_TYP.body"

# Step 3: Compile Typst → PDF
echo "  [2/2] Typst → PDF..."
cd "$PROJECT_ROOT"
typst compile \
    --root . \
    --font-path fonts/sourcesans/OTF \
    --font-path fonts/jetbrainsmono/fonts/ttf \
    "$TEMP_TYP" "$OUTPUT"

echo "Done: $OUTPUT"
ls -la "$OUTPUT"
