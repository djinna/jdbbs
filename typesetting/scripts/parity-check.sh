#!/usr/bin/env bash
# parity-check.sh — Page-aligned parity check of the local Typst build against
# the InDesign golden (reference/GHOSTS.pdf). Emits document facts, the embedded
# font list, a page-geometry table, and side-by-side + overlay renders into
# scratch/parity/.
#
# Usage:
#   typesetting/scripts/parity-check.sh <reference-pdf-page> [candidate-page]
#
# Env overrides: REF= SRC= OUT= DPI=
#
# Alignment is by text: the script reads the reference page's first substantial
# line and finds the candidate page that contains it, so differing front-matter
# / chapter-opener page counts don't throw the comparison off. Pass an explicit
# candidate page as the 2nd arg to skip auto-matching.
#
# Geometry (margins, measure) is font-independent and should match NOW. Word /
# line counts depend on font metrics and will only converge once the licensed
# fonts (Plantin MT Pro / Proxima Nova) replace the open substitutes.
set -eu

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

REF="${REF:-reference/GHOSTS.pdf}"
SRC="${SRC:-manuscripts/ghosts/main.typ}"
OUT="${OUT:-scratch/parity}"
DPI="${DPI:-200}"
REFPAGE="${1:?usage: parity-check.sh <reference-pdf-page> [candidate-page]}"
CANDPAGE="${2:-}"

for tool in typst pdfinfo pdffonts pdftotext pdftoppm magick perl; do
  command -v "$tool" >/dev/null || { echo "parity-check: missing '$tool'"; exit 1; }
done
[ -f "$REF" ] || { echo "parity-check: reference not found: $REF"; exit 1; }

mkdir -p "$OUT"
CAND="$OUT/candidate.pdf"

echo "==> Compiling candidate ($SRC)"
typst compile --root . --font-path typesetting/fonts "$SRC" "$CAND"

echo
echo "==> Document facts"
printf "  %-5s %s\n" "REF"  "$(pdfinfo "$REF"  | awk -F': +' '/Pages|Page size/{printf "%s | ",$0}')"
printf "  %-5s %s\n" "CAND" "$(pdfinfo "$CAND" | awk -F': +' '/Pages|Page size/{printf "%s | ",$0}')"

echo
echo "==> Candidate embedded fonts (want Plantin/Proxima; subs = Libertinus/SourceSans)"
pdffonts "$CAND" | sed 's/^/  /'

# ---- align candidate page to the reference page by text ----
# Normalize a page to a lowercase, letters-only, single-spaced word stream so
# differing line breaks / punctuation / hyphenation between InDesign and Typst
# don't defeat the match. Then probe with several short word-windows.
norm() { pdftotext -f "$2" -l "$2" "$1" - 2>/dev/null | tr '[:upper:]' '[:lower:]' | tr -cs '[:lower:]' ' ' | tr -s ' '; }

read -ra rw <<< "$(norm "$REF" "$REFPAGE")"
n="${#rw[@]}"
echo
[ "$n" -ge 10 ] || { echo "==> ref p${REFPAGE}: too little text (${n} words) — image/opener page? pick another"; exit 1; }
probes=()
for frac in 20 40 60; do
  i=$(( n * frac / 100 ))
  probes+=("${rw[i]} ${rw[i+1]} ${rw[i+2]} ${rw[i+3]} ${rw[i+4]}")
done
echo "==> Reference p${REFPAGE} anchor: \"${probes[0]} ...\""

if [ -z "$CANDPAGE" ]; then
  npages="$(pdfinfo "$CAND" | awk '/Pages/{print $2}')"
  p=1
  while [ "$p" -le "$npages" ]; do
    cnorm=" $(norm "$CAND" "$p") "
    for probe in "${probes[@]}"; do
      case "$cnorm" in *" $probe "*) CANDPAGE="$p"; break 2;; esac
    done
    p=$((p + 1))
  done
fi
[ -n "$CANDPAGE" ] || { echo "  !! anchor not found in candidate — pass the candidate page explicitly"; exit 1; }
echo "==> Matched candidate page: ${CANDPAGE}"

# ---- page geometry via -bbox (perl: macOS awk lacks capture-group match()) ----
geom() { # <pdf> <page>  ->  "left right top bottom measure words lines"
  # Restrict to the body band [52,494]pt so the running header and page folio
  # (which live in the top/bottom margins) don't pollute the body-block metrics.
  pdftotext -bbox -f "$2" -l "$2" "$1" - 2>/dev/null | perl -ne '
    if (/<word xMin="([\d.]+)" yMin="([\d.]+)" xMax="([\d.]+)" yMax="([\d.]+)"/) {
      my ($xn,$yn,$xx,$yx)=($1,$2,$3,$4);
      next if $yn < 52 || $yn > 494;
      $L=$xn if !defined($L)||$xn<$L; $R=$xx if !defined($R)||$xx>$R;
      $T=$yn if !defined($T)||$yn<$T; $B=$yx if !defined($B)||$yx>$B;
      $n++; $seen{int($yn/2)}=1;
    }
    END{ printf "%.1f %.1f %.1f %.1f %.1f %d %d",
      $L//0,$R//0,$T//0,$B//0,($R//0)-($L//0),$n//0,scalar(keys %seen) }'
}

read -r rL rR rT rB rM rW rN <<EOF
$(geom "$REF" "$REFPAGE")
EOF
read -r cL cR cT cB cM cW cN <<EOF
$(geom "$CAND" "$CANDPAGE")
EOF

echo
echo "==> Page geometry (points, origin top-left)"
echo "    Margins + measure are font-independent → should match now."
echo "    words/lines depend on font metrics → expected to differ until real fonts land."
printf "  %-9s %9s %9s %9s\n" "metric" "REF" "CAND" "delta"
prow() { printf "  %-9s %9s %9s %9s\n" "$1" "$2" "$3" "$(perl -e "printf '%.1f', $3 - $2")"; }
prow "left"    "$rL" "$cL"
prow "right"   "$rR" "$cR"
prow "top"     "$rT" "$cT"
prow "bottom"  "$rB" "$cB"
prow "measure" "$rM" "$cM"
prow "words"   "$rW" "$cW"
prow "lines"   "$rN" "$cN"

# ---- visual: side-by-side + overlay ----
rm -f "$OUT"/ref-*.png "$OUT"/cand-*.png
pdftoppm -png -r "$DPI" -f "$REFPAGE"  -l "$REFPAGE"  "$REF"  "$OUT/ref"  >/dev/null
pdftoppm -png -r "$DPI" -f "$CANDPAGE" -l "$CANDPAGE" "$CAND" "$OUT/cand" >/dev/null
refpng=""; for f in "$OUT"/ref-*.png;  do refpng="$f";  break; done
candpng="";for f in "$OUT"/cand-*.png; do candpng="$f"; break; done

# Horizontal append (not montage — montage auto-labels tiles with the filename,
# which needs an X font ImageMagick can't reliably find on macOS). A gray border
# on each page separates them. LEFT = reference, RIGHT = candidate.
magick \( "$refpng"  -bordercolor "#999999" -border 6 \) \
       \( "$candpng" -bordercolor "#999999" -border 6 \) \
  +append "$OUT/side-by-side.png"

# overlay: difference composite (candidate resized to ref dims). Where the two
# text blocks occupy the same rectangle the edges line up; margin/measure drift
# shows as offset bands. Glyph shapes differ until real fonts land, so expect
# the whole text block to light up — read it for block *position*, not glyphs.
dims="$(magick identify -format '%wx%h' "$refpng")"
magick "$refpng" \( "$candpng" -resize "${dims}!" \) \
  -compose difference -composite "$OUT/overlay.png"

echo
echo "==> Renders written:"
echo "    $OUT/side-by-side.png   (primary — read this)"
echo "    $OUT/overlay.png        (difference — block position alignment)"
