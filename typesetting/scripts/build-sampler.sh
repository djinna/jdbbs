#!/usr/bin/env bash
# build-sampler.sh — rebuild the customer typography sampler
# (srv/static/samples/typography-sampler.pdf + chip/enlarged PNGs) shown on
# the transmittal's Typeface row. Punch list 0.8 part 3, 2026-09-19.
#
# Runs Typst directly (no API build, no credits). The page setting for each
# pairing comes from the same Go code the real build uses
# (cmd/typoconfig → srv.SamplerTemplate), so the sampler cannot drift from
# production output. Fonts are read from typesetting/fonts and only ever
# subset-embedded in the PDF; nothing is copied out.
#
#   typesetting/scripts/build-sampler.sh            # from anywhere
#   SAMPLER_MS=path/to/manuscript.typ …             # other source text
#
# Needs: go, typst ≥0.12, pdfunite, pdftoppm (poppler). Optional: pngquant /
# optipng (used if present).
#
# Source text: one story from the Ghosts in Machines anthology as converted
# by the pipeline (scratch/typo/ghosts2.typ; a customer manuscript, so it
# lives in the gitignored scratch dir and is not committed).

set -euo pipefail

ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
cd "$ROOT"

MS=${SAMPLER_MS:-scratch/typo/ghosts2.typ}
STORY=${SAMPLER_STORY:-"The House That Paid Its Own Bills"}
NEXT_STORY=${SAMPLER_NEXT_STORY:-"Latency"}
AUTHOR=${SAMPLER_AUTHOR:-"Elizabeth Maher"}
BOOK_TITLE="Ghosts in Machines"
TRIM=us-trade          # 6 × 9
SIZE=standard
BUILD=scratch/sampler/build   # must sit under $ROOT: Typst imports are root-absolute
OUT=srv/static/samples
FONTS=typesetting/fonts

if [ ! -f "$MS" ]; then
  echo "build-sampler: manuscript $MS not found (set SAMPLER_MS)" >&2
  exit 1
fi
for tool in go typst pdfunite pdftoppm; do
  command -v "$tool" >/dev/null || { echo "build-sampler: need $tool" >&2; exit 1; }
done

rm -rf "$BUILD"
mkdir -p "$BUILD"
# Build the config helper once (before touching $OUT: the server embeds
# srv/static/*, and an empty samples dir would break the go build).
go build -o "$BUILD/typoconfig" ./cmd/typoconfig
mkdir -p "$OUT"

# ── Source text ─────────────────────────────────────────────────────────────
# The whole story (chapter opener + enough text for four pages), minus the
# pipeline's editorial-review comments. Two conversion warts in the first
# pages are tidied for the sampler only: escaped Markdown emphasis, hyphen
# bullets (set as a proper list) and escaped straight quotes (let Typst's
# smart quotes curl them).
awk -v s="= $STORY" -v n="= $NEXT_STORY" '$0==s{f=1} $0==n{f=0} f' "$MS" \
  | grep -v '^// Editorial review' \
  | sed -e 's/^\\- /- /' -e 's/\\"/"/g' -e "s/\\\\'/'/g" \
  | perl -0pe 's/\\_([^_]*?)\\_/#emph[$1]/gs' \
  > "$BUILD/story.typ"
[ -s "$BUILD/story.typ" ] || { echo "build-sampler: story '$STORY' not found in $MS" >&2; exit 1; }

# Text-only excerpts for the section-break / paragraph pages: body text with
# the chapter opener and sub-heads stripped, so the page reads text → break →
# text. Both start mid-book (folio 29, a recto) so a running head is carried.
#   breaks.typ  — prologue, #section-break, the next scene
#   paras.typ   — the dialogue-heavy scene that follows (no break on the page)
awk '/^#first-para\[/{c++} c>=1 && c<=2' "$BUILD/story.typ" | grep -v '^===\|^<' > "$BUILD/breaks.typ"
awk '/^#first-para\[/{c++} c==2' "$BUILD/story.typ" | grep -v '^===\|^<' > "$BUILD/paras.typ"

# ── Helpers ──────────────────────────────────────────────────────────────────
# pairing key → display name, body face, heading face (mirror srv/typochoices.go)
pairing_name() { case $1 in classic) echo "Open classic";; house) echo "Studio house";; literary) echo "Literary";; esac; }
pairing_faces() { case $1 in classic) echo "Libertinus Serif & Source Sans 3";; house) echo "Plantin MT Pro & Proxima Nova";; literary) echo "EB Garamond";; esac; }
break_name() { case $1 in space) echo "White space";; breve) echo "Breve";; ornament) echo "Ornament";; esac; }

# write_template DIR PAIRING BREAK PARAS → prints "base/leading" (pt) from the config
write_template() {
  local dir=$1
  "$BUILD/typoconfig" -out "$dir" -trim "$TRIM" -pairing "$2" -size "$SIZE" -break "$3" -paragraphs "$4" -v 2>"$dir/config.txt" >/dev/null \
    || { cat "$dir/config.txt" >&2; return 1; }
  awk '/base-size:/{gsub(/[^0-9.]/,"",$2); b=$2} END{printf "%s/%.1f pt", b, b*1.28}' "$dir/config.txt"
}

# doc HEADER-VARIANT DIR CAPTION BODYFILE [PRELUDE] → $dir/doc.typ
# The caption is a Typst page foreground so every page stands alone when
# enlarged; it sits in the bottom margin below the drop folio.
write_doc() {
  local dir=$1 caption=$2 body=$3 prelude=${4:-}
  {
    printf '#import "/%s/templates/series-template.typ": *\n' "$dir"
    printf '#show: book.with(title: "%s", author: "%s")\n' "$BOOK_TITLE" "$AUTHOR"
    printf '#set page(foreground: place(bottom + center, dy: -0.3in, text(font: "Source Sans 3", size: 6pt, fill: luma(45%%), tracking: 0.08em)[%s]))\n' "$caption"
    printf '#set-story-info(title: "%s", author: "%s")\n' "$STORY" "$AUTHOR"
    [ -n "$prelude" ] && printf '%s\n' "$prelude"
    cat "$body"
  } > "$dir/doc.typ"
}

compile() { # DIR PAGES OUTPDF
  typst compile --root . --font-path "$FONTS" --pages "$2" "$1/doc.typ" "$3"
}

PARTS=()

# ── Front: title page + how-to-read verso ───────────────────────────────────
cat > "$BUILD/front.typ" <<'TYP'
#set page(width: 6in, height: 9in, margin: (top: 0.8in, bottom: 1in, inside: 0.88in, outside: 0.75in))
#set text(font: "Libertinus Serif", size: 10.5pt, lang: "en")
#let kicker(s) = text(font: "Source Sans 3", size: 7.5pt, tracking: 0.12em, fill: luma(35%))[#upper(s)]
#let rule = line(length: 100%, stroke: 0.4pt + luma(60%))

#v(1.6in)
#kicker[jdbb studio · typography sampler]
#v(6pt)
#rule
#v(14pt)
#text(font: "Source Sans 3", weight: "bold", size: 26pt)[Three ways to set a book]
#v(10pt)
#text(size: 12pt, style: "italic")[The same pages, in each of the pairings offered on the transmittal form.]
#v(28pt)
#set par(leading: 6pt)
#grid(columns: (1.2in, 1fr), row-gutter: 12pt, column-gutter: 10pt,
  text(font: "Source Sans 3", weight: 600, size: 10pt)[Open classic], text(font: "Libertinus Serif")[Libertinus Serif for the text, Source Sans 3 for headings. Even, quiet, reads well at any size.],
  text(font: "Proxima Nova", weight: 600, size: 10pt)[Studio house], text(font: "Plantin MT Pro")[Plantin for the text, Proxima Nova for headings — the jdbb series look. Warm, slightly dark on the page.],
  text(font: "EB Garamond", weight: 600, size: 10.5pt)[Literary], text(font: "EB Garamond", size: 11pt)[EB Garamond for text and headings alike. Light, old-style, fiction and essays.],
)
#v(1fr)
#rule
#v(6pt)
#kicker[6 × 9 in trim · standard text size · pages shown at actual size]

#pagebreak()
#v(1.6in)
#kicker[How to read this sampler]
#v(6pt)
#rule
#v(14pt)
#set par(justify: true, leading: 6pt, spacing: 11pt)
Each pairing is shown on four consecutive pages from the same story — the chapter opener, then the next three, so you see a full spread with running heads and folios exactly as the book would carry them. The text is set at the standard size for a 6 × 9 trim; the compact and generous sizes on the form are the same setting a little smaller or larger.

Look at the pages rather than the letters: the darkness of the block of text, how the lines sit in the margins, how a heading meets the text below it. Any of the three will make a good book; they differ in tone.

The short second part shows the three section-break marks and the two paragraph styles, all in the open classic pairing. The small line at the foot of every page names its setting, so a page stands on its own when enlarged.

#v(1fr)
#text(size: 9pt, fill: luma(35%))[Sample text: “The House That Paid Its Own Bills” by Elizabeth Maher, from _Ghosts in Machines_ (Protocolized). Used here for type specimen purposes only.]
TYP
typst compile --root . --font-path "$FONTS" "$BUILD/front.typ" "$BUILD/00-front.pdf"
PARTS+=("$BUILD/00-front.pdf")

# ── Part one: three pairings × four pages ───────────────────────────────────
i=1
for pairing in classic house literary; do
  dir="$BUILD/$pairing"; mkdir -p "$dir"
  sizes=$(write_template "$dir" "$pairing" space indented)
  caption="Sampler — $(pairing_name "$pairing") · $(pairing_faces "$pairing") · $sizes · 6 × 9 in"
  write_doc "$dir" "$caption" "$BUILD/story.typ"
  compile "$dir" 1-4 "$BUILD/1$i-$pairing.pdf"
  PARTS+=("$BUILD/1$i-$pairing.pdf")
  i=$((i+1))
done

# ── Part two divider ────────────────────────────────────────────────────────
cat > "$BUILD/divider.typ" <<'TYP'
#set page(width: 6in, height: 9in, margin: (top: 0.8in, bottom: 1in, inside: 0.88in, outside: 0.75in))
#set text(font: "Libertinus Serif", size: 10.5pt, lang: "en")
#v(1.6in)
#text(font: "Source Sans 3", size: 7.5pt, tracking: 0.12em, fill: luma(35%))[PART TWO]
#v(6pt)
#line(length: 100%, stroke: 0.4pt + luma(60%))
#v(14pt)
#text(font: "Source Sans 3", weight: "bold", size: 22pt)[Section breaks and paragraphs]
#v(12pt)
#set par(justify: true, leading: 6pt, spacing: 11pt)
Three ways to mark a change of scene — white space, the studio's breve, a single ornament — followed by the same page with indented paragraphs and with block paragraphs. All in the open classic pairing at the standard size.
TYP
typst compile --root . --font-path "$FONTS" "$BUILD/divider.typ" "$BUILD/20-divider.pdf"
PARTS+=("$BUILD/20-divider.pdf")

# ── Section breaks: classic × space / breve / ornament ─────────────────────
# The running head reads the story-info state at the top of the page, before
# the body flow on page 1 has set it, so the demo docs spend page 1 on the
# state update and export page 2 — a verso mid-book (folio 30).
DEMO_PRELUDE='#counter(page).update(29)
#pagebreak()'
i=1
for brk in space breve ornament; do
  dir="$BUILD/break-$brk"; mkdir -p "$dir"
  sizes=$(write_template "$dir" classic "$brk" indented)
  caption="Sampler — Section break: $(break_name "$brk") · Open classic · $sizes"
  write_doc "$dir" "$caption" "$BUILD/breaks.typ" "$DEMO_PRELUDE"
  compile "$dir" 2 "$BUILD/2$i-break-$brk.pdf"
  PARTS+=("$BUILD/2$i-break-$brk.pdf")
  i=$((i+1))
done

# ── Paragraphs: classic × indented / block ─────────────────────────────────
i=1
for paras in indented block; do
  dir="$BUILD/paras-$paras"; mkdir -p "$dir"
  sizes=$(write_template "$dir" classic space "$paras")
  caption="Sampler — Paragraphs: ${paras^} · Open classic · $sizes"
  write_doc "$dir" "$caption" "$BUILD/paras.typ" "$DEMO_PRELUDE"
  compile "$dir" 2 "$BUILD/3$i-paras-$paras.pdf"
  PARTS+=("$BUILD/3$i-paras-$paras.pdf")
  i=$((i+1))
done

# ── Assemble ────────────────────────────────────────────────────────────────
pdfunite "${PARTS[@]}" "$OUT/typography-sampler.pdf"

# ── Chips: recto page 3 of each pairing (running head + folio), 2× thumb and
#    an enlarged view for the lightbox. Greyscale PNG keeps them small.
# render PDF PAGE DPI OUT.png — 4-bit greyscale via netpbm when available
# (half the bytes of pdftoppm's 8-bit PNG), else pdftoppm's PNG directly.
render() {
  if command -v pamdepth >/dev/null && command -v pnmtopng >/dev/null; then
    pdftoppm -r "$3" -f "$2" -l "$2" -gray "$1" | pamdepth 15 | pnmtopng -compression 9 > "$4"
  else
    pdftoppm -r "$3" -f "$2" -l "$2" -gray -png -singlefile "$1" "${4%.png}"
  fi
}
for pairing in classic house literary; do
  src=$(ls "$BUILD"/1?-"$pairing".pdf)
  render "$src" 3 96 "$OUT/chip-$pairing.png"    # 576 px wide: 2.4× a 240 px chip
  render "$src" 3 200 "$OUT/page-$pairing.png"   # 1200 px wide: lightbox
done
if command -v pngquant >/dev/null; then
  pngquant --force --ext .png --speed 1 16 "$OUT"/chip-*.png "$OUT"/page-*.png
elif command -v optipng >/dev/null; then
  optipng -quiet -o2 "$OUT"/chip-*.png "$OUT"/page-*.png
fi

echo "build-sampler: $(pdfinfo "$OUT/typography-sampler.pdf" | awk '/^Pages/{print $2}') pages"
ls -la "$OUT"
