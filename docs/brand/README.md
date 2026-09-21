# Brand bits

- `srv/static/favicon.svg` — the live favicon: `jd` in JetBrains Mono ExtraBold,
  42px on a 64 tile, −4 tracking, accent `j` (#007699) + paper `d` on ink (#0E1116).
  Glyphs are real outlines from `typesetting/fonts/jetbrainsmono`, so no font dependency.
- `favicon-bb-bookbuilder.svg` — the `bb` sibling. jdbb once stood for
  *Jenna Dixon, Bookbuilder*; kept for that reason.
- `favicon-gen.py` — regenerates the variants into `scratch/favicon/`
  (needs `fonttools`). Run from repo root.
- `wordmark-gen.py` — generates `srv/static/brand/` (2026-09-19): `jdbb-wordmark.svg`,
  `-paper.svg` (for ink grounds), `-short.svg` (`[jdbb]` only), and the 1200×630
  `jdbb-social-card.svg/.jpg`. Outlines from JetBrains Mono Bold/Regular; brackets
  accent, `jdbb` ink, `studio` secondary. Linked from `/press` (jdbbs-public). Run
  from repo root; needs `fonttools` + ImageMagick.

## Writing the name

- In running text and captions the studio is **`[jdbb] studio`** — square brackets
  always, lowercase, a space before *studio* (Jenna, 2026-09-20, talk-deck nit).
  Not "jdbb studio", not "JDBB". The brackets are part of the name, not decoration;
  the wordmark markup (`.jdbb-wordmark`, PAGE-DESIGN §3) is the same thing set in type.
- Short form where space is tight: `[jdbb]`.
- Bylines: `Jenna Dixon, [jdbb] studio`.
