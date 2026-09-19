# Index as a factory add-on — phase 1 design note

**Date:** 2026-09-19 · punch list 5.13 · branch `index-addon` (worktree, not merged)
**Follows:** docs/reviews/INDEX-ADDON-FEASIBILITY-2026-09-18.md — design A (Typst-resolved
locators), generative model for the entries. This note records what was built, the
formats the pieces exchange, and the phase-2 surface (API, UI, SKU) left for the lead.

## What was built (pieces 1–3 of the brief)

| Piece | Where | State |
|---|---|---|
| 1 Typst index piece | `typesetting/templates/series-template.typ` ("BACK-OF-BOOK INDEX" section), fixture `typesetting/test/index-fixture.typ` | done, both typst 0.12.0 and 0.13.1 |
| 2 Entry drafting | `srv/indexer/` (`llm.go`, `draft.go`, `typtext.go`), `cmd/indexdraft` | done; Ghosts draft ≈ $0.43 |
| 3 Anchoring | `srv/indexer/anchor.go`, `cmd/indexanchor` | done; Ghosts 377/380 anchors placed |
| 4 Word `XE` passthrough | — | not started |
| Factory UI / API / SKU | — | phase 2 (designed below) |

Chain, as run on Ghosts (all paths under the worktree):

```
pandoc → book.typ                                   (existing build step 1)
go run ./cmd/indexdraft  -in book.typ -book "…" -pages 100 -out index.json   (one model read of the book)
   … Jenna reviews / edits index.json (phase-2 UI) …
go run ./cmd/indexanchor -in book.typ -index index.json -out book-indexed.typ
typst compile  (template config index: true)        (existing build step 3)
```

The feasibility doc was right that Typst has no indexer and that `in-dexter` is the
alternative; evaluated 0.7.0 and rejected it: no `see`/`see also`, subentries only as
nested lines (the series wants run-in), no article-insensitive sort, no collapsing of
consecutive pages into ranges, and a `@preview` package fetch inside the build path (the
service box would need network at compile time or a vendored copy). The in-house piece is
~170 commented lines and has no dependency.

## Piece 1 — the Typst piece

**Config.** `index: false` (default), `index-title: "Index"`, `index-columns: 2` in
`default-config`. `book()` ends with `if config.at("index", default: false) { index-page() }`,
so a book without the flag renders exactly as before (Ghosts, both binaries: identical text
and pagination; the only differing PDF bytes are Typst's per-run timestamp and `/ID`, the
same 52 bytes two consecutive compiles of unchanged source differ by).

**Marker syntax** (invisible; a labelled `metadata` element at the mention):

```typst
#index[Rice]                                   heading
#index("Rice", "as staple")                    heading + run-in subentry
#index("Typesetting", see: "Typst", locator: false)   pure cross-reference
#index("Beaings", see-also: "ghosts, artificial")    see-also (page still counts)
#index("Rice", sort: "rice")                   explicit sort key
```

A marker written in running text is followed by `;` when generated (`#index("Rice");`) so a
following `[` in the manuscript cannot be swallowed as a content argument.

**`index-page()`** (called by `book()`; also callable by hand): queries `<index-entry>`,
builds heading → subentries with (folio, physical page, location) locators, sorts by a key
that lower-cases, folds Latin diacritics, drops a leading *the/a/an* and non-alphanumerics;
groups under A–Z letter heads (`#` for digits/symbols; heads are `sticky` to their first
entry); collapses runs of consecutive folios into en-dash ranges; renders each entry as a
hanging-indent paragraph in Chicago run-in style —
`Khlongs, 5: as commons, 2; as irrigation, 1–2. See also canals` —
with locators linked to their pages; two `columns` at 0.9 em with the series heading face for
the title and letter heads. It starts on a recto via `start-back()`, takes the chapter-opener
look (no running head, drop folio), sets running heads *book title / INDEX*, and emits nothing
at all when the manuscript carries no markers. Folios come from `folio-text`, so roman
front-matter pages would show roman.

**Known limits.** Sort key is ASCII after folding (CJK/Thai headings group under `#`).
An unbreakable token longer than a column (`AmaStore_L47_HeartVariant1.0` in the Ghosts
draft) overflows; the review step should reword such headings. Locators are page folios only
(no `n` for notes, no bold for main discussion).

## Piece 2 — drafting

`srv/indexer` calls the VM's keyless gateway (`https://llm.int.exe.xyz/v1/messages`,
Anthropic Messages API; `INDEXER_LLM_URL`, `INDEXER_LLM_MODEL` env; default
`claude-sonnet-4-5`). Hard caps: 8 000 output tokens per call, 120 000 input chars per chunk
(longer chapters split at paragraph breaks); 3 retries on 5xx/429. `-record DIR` saves every
answer keyed by a hash of the prompt and doubles as a cache; `-replay DIR` runs offline and
fails on a miss — tests never touch the network.

**Flow.** `Chunks()` splits the pandoc Typst at level-1 headings (level-2 with `-parts`)
using the shared text extractor (`ExtractText`: function calls, brackets, labels, comments
and `#import/#show/#set/#let` statements removed, offsets kept). One call per chapter with
the indexer persona (`SystemPrompt`) and a per-chapter entry budget; replies are parsed with a
fence stripper and `RepairJSON` (models copy dialogue with unescaped quotes). Identical
heading+subheading pairs are merged across chapters. Then one consolidation call
(`MergeSystemPrompt`) sees every heading with its subheadings and anchor counts and returns
`merge / see / see_also / drop`; renames are chased, a merged-away wording that shares no word
with its target becomes a *see* entry automatically (`beaings. See artificial ghosts`), drops
are applied, `tidy()` enforces the invariants (no self-references; a *see* only on an entry
without locators, otherwise it becomes *see also*; anchors never null).

**Budget.** `pages × 0.04 × 80` entry lines (≈ 4 % of the book as index, ~80 lines per
two-column page): 320 for Ghosts. Pages come from `-pages` (the set book) or words/380.
Chapters get a share ∝ words, +15 %, capped at 60; the merge trims to budget.

**`index.json` schema**

```json
{
  "book": "Ghosts in Machines", "generated": "2026-09-19T16:02:11Z",
  "model": "claude-sonnet-4-5", "words": 23889, "pages_estimate": 100,
  "budget_entries": 320,
  "chapters": ["Khlongs, Subaks, Beaings", "Soda Sweet as Blood", "…"],
  "entries": [
    {"heading": "artificial ghosts", "subheading": "containment protocols",
     "see": "", "see_also": ["uploaded consciousness"],
     "anchors": [{"chapter": 2, "text": "the containment field flickered once and held"}]},
    {"heading": "beaings", "see": "artificial ghosts", "anchors": []}
  ],
  "usage": {"calls": 10, "input_tokens": 46086, "output_tokens": 19162},
  "cost_usd": 0.426,
  "notes": ["merge \"beaings\" → \"artificial ghosts\"", "dropped 86 entries under 86 headings to fit the budget"]
}
```

`heading` + `subheading` identify an entry; `anchors[].text` is a verbatim fragment,
`anchors[].chapter` the 1-based chunk (narrows the search). An entry with `see` and no
anchors is a pure cross-reference. This is the document the phase-2 review UI edits.

**Ghosts result** (`scratch/idx/ghosts-index.json` in the worktree): 209 entries, 155
distinct headings, 64 subentries, 21 *see*, 24 *see also*, 380 anchors; 10 calls,
46 086 in / 19 162 out tokens, ≈ $0.43 at list price, 5 min wall-clock. Reads like a first
draft an indexer would edit: good headings (Ama, Airloom, Checkpoints, khlongs, spirit
houses, uploaded consciousness), sensible subentries, a few merges to undo
("distributed AI" → "uploaded consciousness") and a couple of over-literal headings.

## Piece 3 — anchoring

`PlaceMarkers(src, idx)` folds the source (`FoldText`: lower-case, diacritics, quotes,
dashes and markup gone, offset map kept) and each anchor the same way, searches the anchor's
chapter first and then the whole book, and if the fragment fails, shortens it a word at a
time from the end, then from the start, down to four words (`fuzzy` in the report).
Unmatched anchors are listed, never guessed. The marker is written after the matched word's
closing punctuation and one following space, glued to the next word — a zero-width element
between a word and its space removes Typst's break opportunity there and reflowed whole
paragraphs (Ghosts went from dozens of reflowed lines to one paragraph). Cross-references
attach to the entry's first marker; see-only entries get `locator: false` markers at the end
of the file. Output is a report (`-report x.json`) plus `book-indexed.typ`.

Ghosts: 380 anchors, 377 placed (9 fuzzy), 3 unmatched (all model paraphrases of dialogue),
21 cross-references; compiled to 105 pages (100 body + 5 index) on both binaries with
identical text. Sample: `scratch/idx/ghosts-indexed-typst0.13.pdf` pp. 101–105.

## Phase 2 — proposed surface (not built)

**API** (all under the existing book auth; drafting costs money, so admin- or pass-gated):

- `POST /api/books/{id}/index/draft` `{pages?: int, about?: string}` → runs pandoc (step 1
  of the build, already there) + `indexer.Draft`, stores `index.json` on the book
  (new column `books.index_json TEXT` + `index_status` draft|reviewed|off), returns it with
  `usage`/`cost_usd`. Async like convert; ~5 min for a 100-page book.
- `GET /api/books/{id}/index` → the stored draft. `PUT /api/books/{id}/index` → the edited
  document (validated: headings non-empty, `see` targets exist, anchors verbatim or flagged);
  sets `index_status = reviewed`.
- `POST /api/books/{id}/convert` gains `"index": true|false`; when true the build runs
  `PlaceMarkers` between steps 2 and 3, writes the anchoring report into the build log,
  and `specToTypstConfig` emits `index: true`. Unmatched anchors surface as build warnings,
  not failures.
- Spec: `index: {enabled: bool, title: "Index", columns: 2}` in the book spec so the
  transmittal can carry the choice.

**Factory UI** — a step between Inspect and Build, "Index — draft, review, include":
draft button (shows estimated cost and that it takes minutes), then a two-pane review:
left the entries as they will print (heading, subentries, see/see also, anchor count),
right the anchors of the selected entry with their sentence in context, unmatched ones
flagged; inline edit/merge/delete/add; "Include the index in this build" toggle on the
Build card. The draft is also downloadable as JSON for editing outside.

**Store SKU** — `{LookupKey: "index", Name: "Back-of-book index", Description: "A drafted,
reviewable index set into the print PDF — Typst-resolved page numbers, true after every
rebuild.", Amount: 4900–7900}` in `storeCatalog` (srv/store.go); entitlement flag on the
pass like `builds`. Model cost is well under a dollar a book.

## Open questions for Jenna

1. Price point and whether the index is included in a higher pass tier or sold alone.
2. Model: Sonnet 4.5 at ≈ $0.43/100 pp read well; try Opus on Obliquities for comparison?
3. Obliquities as the first real test — needs her go-ahead (files not pulled).
4. Index for fiction/anthologies (Ghosts): keep as a demo only, or offer it?
5. Should the review step be mandatory before an index can go into a build (I think yes)?

## What remains, honest effort

- Piece 4, Word `XE` passthrough in the Lua filter: ½ day; not started.
- Server integration of the chain (columns, endpoints, build flag, `specToTypstConfig`):
  1 day. Factory UI step: 1 day (lead, after 0.17).
- Prompt tuning on a non-fiction book (Obliquities) and a page-measured budget (read the
  page count from a first build instead of `-pages`): ½ day.
- Nice-to-haves: `n` locators for footnotes; bold main-discussion locators; CJK sort.
