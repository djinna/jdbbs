#!/bin/bash
# Convert Word DOCX to PDF using Pandoc → Typst → PDF
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

if [ $# -lt 1 ]; then
    echo "Usage: $0 <input.docx> [output.pdf]"
    echo ""
    echo "Converts Word documents to PDF using the series template."
    echo "Use Word styles for best results - see docs/word-styles.md"
    exit 1
fi

INPUT="$1"
BASENAME="$(basename "${INPUT%.docx}")"
OUTPUT="${2:-$PROJECT_ROOT/output/${BASENAME}.pdf}"
TEMP_TYP="$PROJECT_ROOT/output/${BASENAME}.typ"

echo "Converting: $INPUT → $OUTPUT"

# Step 1: Pandoc DOCX → Typst
echo "  [1/2] DOCX → Typst..."
pandoc -f docx -t typst \
    --wrap=none \
    "$INPUT" -o "$TEMP_TYP.body"

# Step 2: Wrap with template
cat > "$TEMP_TYP" << 'HEADER'
#import "../templates/series-template.typ": *

#show: book.with(
  title: "Untitled",
)

// Map Word horizontal rules to section breaks
#let horizontalrule = section-break

// Style mappings for common Word elements
// Heading 1 → Chapter title (handled by template)
// Heading 2 → Section heading
// Heading 3 → Subsection
// Normal → Body paragraph
// Quote → Block quote
// Code → Code block

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
