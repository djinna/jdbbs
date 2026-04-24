# Typst Named Paper Sizes — Complete Reference

Source: [typst/typst `page.rs`](https://github.com/typst/typst/blob/main/crates/typst-library/src/layout/page.rs)  
Verified against Typst source, July 2025. **107 named sizes.**

Use these as the `paper` argument to `#set page(paper: "us-trade")` or as the
first positional argument: `#set page("us-trade")`.

## ISO 216 A Series

| Typst Name | Width (mm) | Height (mm) | Width (in) | Height (in) |
|------------|-----------|------------|-----------|------------|
| `a0`  | 841 | 1189 | 33.11 | 46.81 |
| `a1`  | 594 | 841 | 23.39 | 33.11 |
| `a2`  | 420 | 594 | 16.54 | 23.39 |
| `a3`  | 297 | 420 | 11.69 | 16.54 |
| `a4`  | 210 | 297 | 8.27 | 11.69 |
| `a5`  | 148 | 210 | 5.83 | 8.27 |
| `a6`  | 105 | 148 | 4.13 | 5.83 |
| `a7`  | 74 | 105 | 2.91 | 4.13 |
| `a8`  | 52 | 74 | 2.05 | 2.91 |
| `a9`  | 37 | 52 | 1.46 | 2.05 |
| `a10` | 26 | 37 | 1.02 | 1.46 |
| `a11` | 18 | 26 | 0.71 | 1.02 |

## ISO 216 B Series

| Typst Name | Width (mm) | Height (mm) | Width (in) | Height (in) |
|------------|-----------|------------|-----------|------------|
| `iso-b1` | 707 | 1000 | 27.83 | 39.37 |
| `iso-b2` | 500 | 707 | 19.69 | 27.83 |
| `iso-b3` | 353 | 500 | 13.90 | 19.69 |
| `iso-b4` | 250 | 353 | 9.84 | 13.90 |
| `iso-b5` | 176 | 250 | 6.93 | 9.84 |
| `iso-b6` | 125 | 176 | 4.92 | 6.93 |
| `iso-b7` | 88 | 125 | 3.46 | 4.92 |
| `iso-b8` | 62 | 88 | 2.44 | 3.46 |

## ISO 216 C Series (Envelopes)

| Typst Name | Width (mm) | Height (mm) | Width (in) | Height (in) |
|------------|-----------|------------|-----------|------------|
| `iso-c3` | 324 | 458 | 12.76 | 18.03 |
| `iso-c4` | 229 | 324 | 9.02 | 12.76 |
| `iso-c5` | 162 | 229 | 6.38 | 9.02 |
| `iso-c6` | 114 | 162 | 4.49 | 6.38 |
| `iso-c7` | 81 | 114 | 3.19 | 4.49 |
| `iso-c8` | 57 | 81 | 2.24 | 3.19 |

## DIN D Series

| Typst Name | Width (mm) | Height (mm) | Width (in) | Height (in) |
|------------|-----------|------------|-----------|------------|
| `din-d3` | 272 | 385 | 10.71 | 15.16 |
| `din-d4` | 192 | 272 | 7.56 | 10.71 |
| `din-d5` | 136 | 192 | 5.35 | 7.56 |
| `din-d6` | 96 | 136 | 3.78 | 5.35 |
| `din-d7` | 68 | 96 | 2.68 | 3.78 |
| `din-d8` | 48 | 68 | 1.89 | 2.68 |

## SIS (Swedish Academic)

| Typst Name | Width (mm) | Height (mm) | Width (in) | Height (in) |
|------------|-----------|------------|-----------|------------|
| `sis-g5` | 169 | 239 | 6.65 | 9.41 |
| `sis-e5` | 115 | 220 | 4.53 | 8.66 |

## ANSI Extensions

| Typst Name | Width (mm) | Height (mm) | Width (in) | Height (in) |
|------------|-----------|------------|-----------|------------|
| `ansi-a` | 216 | 279 | 8.50 | 10.98 |
| `ansi-b` | 279 | 432 | 10.98 | 17.01 |
| `ansi-c` | 432 | 559 | 17.01 | 22.01 |
| `ansi-d` | 559 | 864 | 22.01 | 34.02 |
| `ansi-e` | 864 | 1118 | 34.02 | 44.02 |

## ANSI Architectural

| Typst Name | Width (mm) | Height (mm) | Width (in) | Height (in) |
|------------|-----------|------------|-----------|------------|
| `arch-a` | 229 | 305 | 9.02 | 12.01 |
| `arch-b` | 305 | 457 | 12.01 | 17.99 |
| `arch-c` | 457 | 610 | 17.99 | 24.02 |
| `arch-d` | 610 | 914 | 24.02 | 35.98 |
| `arch-e1` | 762 | 1067 | 30.00 | 42.01 |
| `arch-e` | 914 | 1219 | 35.98 | 47.99 |

## JIS B Series (Japan)

| Typst Name | Width (mm) | Height (mm) | Width (in) | Height (in) |
|------------|-----------|------------|-----------|------------|
| `jis-b0` | 1030 | 1456 | 40.55 | 57.32 |
| `jis-b1` | 728 | 1030 | 28.66 | 40.55 |
| `jis-b2` | 515 | 728 | 20.28 | 28.66 |
| `jis-b3` | 364 | 515 | 14.33 | 20.28 |
| `jis-b4` | 257 | 364 | 10.12 | 14.33 |
| `jis-b5` | 182 | 257 | 7.17 | 10.12 |
| `jis-b6` | 128 | 182 | 5.04 | 7.17 |
| `jis-b7` | 91 | 128 | 3.58 | 5.04 |
| `jis-b8` | 64 | 91 | 2.52 | 3.58 |
| `jis-b9` | 45 | 64 | 1.77 | 2.52 |
| `jis-b10` | 32 | 45 | 1.26 | 1.77 |
| `jis-b11` | 22 | 32 | 0.87 | 1.26 |

## SAC D Series (China)

| Typst Name | Width (mm) | Height (mm) | Width (in) | Height (in) |
|------------|-----------|------------|-----------|------------|
| `sac-d0` | 764 | 1064 | 30.08 | 41.89 |
| `sac-d1` | 532 | 760 | 20.94 | 29.92 |
| `sac-d2` | 380 | 528 | 14.96 | 20.79 |
| `sac-d3` | 264 | 376 | 10.39 | 14.80 |
| `sac-d4` | 188 | 260 | 7.40 | 10.24 |
| `sac-d5` | 130 | 184 | 5.12 | 7.24 |
| `sac-d6` | 92 | 126 | 3.62 | 4.96 |

## ISO 7810 ID Cards

| Typst Name | Width (mm) | Height (mm) | Width (in) | Height (in) |
|------------|-----------|------------|-----------|------------|
| `iso-id-1` | 85.6 | 53.98 | 3.37 | 2.13 |
| `iso-id-2` | 74 | 105 | 2.91 | 4.13 |
| `iso-id-3` | 88 | 125 | 3.46 | 4.92 |

## Asia

| Typst Name | Width (mm) | Height (mm) | Width (in) | Height (in) |
|------------|-----------|------------|-----------|------------|
| `asia-f4` | 210 | 330 | 8.27 | 12.99 |

## Japan (Traditional)

| Typst Name | Width (mm) | Height (mm) | Width (in) | Height (in) |
|------------|-----------|------------|-----------|------------|
| `jp-shiroku-ban-4` | 264 | 379 | 10.39 | 14.92 |
| `jp-shiroku-ban-5` | 189 | 262 | 7.44 | 10.31 |
| `jp-shiroku-ban-6` | 127 | 188 | 5.00 | 7.40 |
| `jp-kiku-4` | 227 | 306 | 8.94 | 12.05 |
| `jp-kiku-5` | 151 | 227 | 5.94 | 8.94 |
| `jp-business-card` | 91 | 55 | 3.58 | 2.17 |

## Business Cards

| Typst Name | Width (mm) | Height (mm) | Width (in) | Height (in) |
|------------|-----------|------------|-----------|------------|
| `cn-business-card` | 90 | 54 | 3.54 | 2.13 |
| `eu-business-card` | 85 | 55 | 3.35 | 2.17 |
| `us-business-card` | 88.9 | 50.8 | 3.50 | 2.00 |
| `jp-business-card` | 91 | 55 | 3.58 | 2.17 |

## French Traditional (AFNOR)

| Typst Name | Width (mm) | Height (mm) | Width (in) | Height (in) |
|------------|-----------|------------|-----------|------------|
| `fr-tellière` | 340 | 440 | 13.39 | 17.32 |
| `fr-couronne-écriture` | 360 | 460 | 14.17 | 18.11 |
| `fr-couronne-édition` | 370 | 470 | 14.57 | 18.50 |
| `fr-raisin` | 500 | 650 | 19.69 | 25.59 |
| `fr-carré` | 450 | 560 | 17.72 | 22.05 |
| `fr-jésus` | 560 | 760 | 22.05 | 29.92 |

## United Kingdom Imperial

| Typst Name | Width (mm) | Height (mm) | Width (in) | Height (in) |
|------------|-----------|------------|-----------|------------|
| `uk-brief` | 406.4 | 342.9 | 16.00 | 13.50 |
| `uk-draft` | 254 | 406.4 | 10.00 | 16.00 |
| `uk-foolscap` | 203.2 | 330.2 | 8.00 | 13.00 |
| `uk-quarto` | 203.2 | 254 | 8.00 | 10.00 |
| `uk-crown` | 508 | 381 | 20.00 | 15.00 |
| `uk-book-a` | 111 | 178 | 4.37 | 7.01 |
| `uk-book-b` | 129 | 198 | 5.08 | 7.80 |

## United States

| Typst Name | Width (mm) | Height (mm) | Width (in) | Height (in) |
|------------|-----------|------------|-----------|------------|
| `us-letter` | 215.9 | 279.4 | 8.50 | 11.00 |
| `us-legal` | 215.9 | 355.6 | 8.50 | 14.00 |
| `us-tabloid` | 279.4 | 431.8 | 11.00 | 17.00 |
| `us-executive` | 184.15 | 266.7 | 7.25 | 10.50 |
| `us-foolscap-folio` | 215.9 | 342.9 | 8.50 | 13.50 |
| `us-statement` | 139.7 | 215.9 | 5.50 | 8.50 |
| `us-ledger` | 431.8 | 279.4 | 17.00 | 11.00 |
| `us-oficio` | 215.9 | 340.36 | 8.50 | 13.40 |
| `us-gov-letter` | 203.2 | 266.7 | 8.00 | 10.50 |
| `us-gov-legal` | 215.9 | 330.2 | 8.50 | 13.00 |
| `us-digest` | 139.7 | 215.9 | 5.50 | 8.50 |
| `us-trade` | 152.4 | 228.6 | 6.00 | 9.00 |

## Newspaper & Presentation

| Typst Name | Width (mm) | Height (mm) | Width (in) | Height (in) |
|------------|-----------|------------|-----------|------------|
| `newspaper-compact` | 280 | 430 | 11.02 | 16.93 |
| `newspaper-berliner` | 315 | 470 | 12.40 | 18.50 |
| `newspaper-broadsheet` | 381 | 578 | 15.00 | 22.76 |
| `presentation-16-9` | 297 | 167.06 | 11.69 | 6.58 |
| `presentation-4-3` | 280 | 210 | 11.02 | 8.27 |

---

## Book-Relevant Sizes (Quick Reference)

For book production, the most commonly used named sizes:

| Use Case | Typst Name | Dimensions | Notes |
|----------|------------|------------|-------|
| US Mass Market | `us-statement` | 5.5" × 8.5" | Same as `us-digest` |
| US Trade Paperback | `us-trade` | 6" × 9" | Standard trade size |
| UK Pocket | `uk-book-a` | 111 × 178 mm | A-format paperback |
| UK Trade | `uk-book-b` | 129 × 198 mm | B-format paperback |
| Academic (ISO) | `a5` | 148 × 210 mm | International standard |
| Academic (Sweden) | `sis-g5` | 169 × 239 mm | SIS thesis format |
| Textbook | `a4` | 210 × 297 mm | Large format |
| Custom (our series) | — | 4.91" × 7.59" | `353.811pt × 546.567pt` |

**For custom sizes** not matching a named paper, use explicit dimensions:
```typst
#set page(width: 4.91in, height: 7.59in)
```
