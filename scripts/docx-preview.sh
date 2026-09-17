#!/usr/bin/env bash
# Render a .docx to PDF + PNG pages with headless LibreOffice for a visual
# check on the VM (installed 2026-09-17: libreoffice-writer-nogui + MS core
# fonts, so Georgia/Arial/Courier New render as themselves).
#   scripts/docx-preview.sh scratch/template22.docx [dpi]
# Writes <name>.pdf and <name>-N.png next to the input.
set -euo pipefail
in="${1:?usage: docx-preview.sh file.docx [dpi]}"
dpi="${2:-70}"
dir="$(dirname "$in")"
base="$(basename "${in%.*}")"
timeout 180 soffice --headless --convert-to pdf --outdir "$dir" "$in" >/dev/null
pdftoppm -r "$dpi" -png "$dir/$base.pdf" "$dir/$base"
pdfinfo "$dir/$base.pdf" | grep -E 'Pages|Page size'
ls "$dir/$base"-*.png
