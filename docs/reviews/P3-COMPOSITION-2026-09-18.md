# P3 — composition quality (H&J) — findings 2026-09-18

Punch list 5.5. Goal: fewer loose lines, runts, widows/orphans and rivers in the
print PDF, measured, not eyeballed.

## Yardstick (landed)

`typesetting/scripts/compscore.py book.pdf [--list|--json]` — reads
`pdftotext -bbox-layout`, finds the body column (modal left edge + modal line
height), and counts per book: **loose** (mean word gap > 0.5 em), **tight**
(< 0.16 em), **runt** (paragraph last line = one word or < 6 chars), **widow**,
**orphan**, **hyph3** (3 hyphenated line-ends in a row), **stack** (same word
starting/ending 3 lines), **river** (word gaps aligned ± 1.5 pt over 3 lines).
Heuristic; rivers over-report a little. Use it to compare A/B builds of the same
manuscript, not as an absolute grade.

Ghosts (book 9, Libertinus Serif 10/12, 5.5 × 8.5, 2395 body lines):

| build | loose | runt | widow+orphan | river |
|---|---|---|---|---|
| typst 0.12.0 (production) | 27 | 42 | 0 | 92 |
| typst 0.13.1, same template | 27 | 42 | 0 | 90 |
| typst 0.13.1 + `text(costs: (runt: 500%))` + `par(linebreaks: "optimized")` | 29 | **29** | 0 | 89 |

Runts −31 % for +2 loose lines. The remaining runts are lines whose previous
line is already full (the cost can't pull a word back). Widows/orphans are
already 0 in this template.

## What the lever needs

`text.costs` (hyphenation / runt / widow / orphan) exists only from **Typst
0.13**. The VM runs 0.12.0. A 0.13 upgrade compiles the series template after
one fix (landed: `outline.entry` `it.body` → `it.element.body`, output
byte-identical on 0.12) but changes layout in two places:

1. **Contents page** — `block(above: 0.9em)` in the entry show rule no longer
   spaces the entries (0.13 lays entries out with `par.spacing`). Fix: set
   spacing on the outline's paragraphs instead.
2. **Lists / block quotes** — paragraph spacing after a list is larger in
   0.13, so a paragraph that fit on p. 9 moves to p. 10 and the shift ripples
   to the end of the chapter (pages 9–11, 21–36, 49–59 of Ghosts differ by
   2–3 lines). Fix: pin `set list(spacing:)` / `set par(spacing:)` around
   lists, then re-diff page by page.

Neither is hard, but both need a page-by-page visual pass on two books. Not
the thing to do two days before a workshop where attendees are building.

## Plan

- **Now:** scorer committed; template compatible with 0.12 and 0.13.
- **Wed 24 Sep or after:** install typst 0.13.1 (`scratch/typst-x86_64-unknown-linux-musl/`
  is the tested binary), fix (1) and (2), add
  `set text(costs: (runt: 500%, widow: 200%, orphan: 200%))` +
  `set par(linebreaks: "optimized")` to `book()`, rebuild Ghosts + Twitter
  Years, compare with compscore and a page diff, then decide on hyphenation
  cost. Also try `par(justify)` micro-tuning: `text(tracking:)` is not per-line
  in Typst, so letterspacing-based H&J is not available — expect word-space
  only.
- InDesign comparison page: still wanted; Jenna to supply one page set in
  InDesign from the same text so the same scorer can run on both.

## Found on the way

- Ghosts chapter 1 has 26 stray `>` characters in the manuscript (quoted-email
  residue in the .docx, not our pipeline). Inspect should flag "stray quote
  markers" — small addition to `detect-edge-cases.py`.
