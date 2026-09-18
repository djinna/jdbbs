# Index as a factory add-on — first-pass feasibility

**Date:** 2026-09-18 · punch-list 0.9 · **Verdict: feasible, not quick. Hold until after the workshop.**

## The question

Could the factory produce a real back-of-book index — a *conceptual* index, the kind an indexer
writes, not a sorted word list — as a paid add-on? And is Jev (typesafe.ai) the tool?

## Short answers

1. **An LLM can write a meaningful index today.** Not a concordance: a generative model reading
   the book chapter by chapter can pick topics, merge synonyms ("the pass" / "Factory Pass"),
   subdivide ("transmittal — as specification; — versions of; — vs. book spec"), add
   cross-references ("typesetting, *see* Typst"), and skip passing mentions. Indexers' own
   professional bodies (ASI, SI) have been testing this since 2024; the consensus is "a strong
   first draft that a human should edit", which is exactly the tier we'd sell.

2. **Jev is the wrong tool for this.** Jev classifies: pick-one, score, yes/no, over ≤32k tokens
   of state. Indexing is generation (inventing headings, choosing wording, structuring
   subentries) over a whole book. Jev could play a *supporting* role — "is this candidate
   term index-worthy for this book? (Noul)", "which of these three headings is the canonical
   one? (Choice)" — as a cheap second opinion on a list produced elsewhere. The core needs a
   generative model (Claude, via the API key this VM already has). So: not "a good use of Jev";
   Jev stays on 5.12 (heading classification).

3. **The hard part is not the entries; it is the locators.** An index is useless without page
   numbers, and page numbers exist only after Typst has set the book. Three designs:
   - **(A) Typst-native, correct.** Insert `#index[term]` markers into the Typst source at the
     mentions, let Typst resolve page numbers at compile time and emit the index. This is how a
     LaTeX index works and is the only design that stays right after every rebuild. Work: the
     LLM proposes *terms + the sentences that mention them*; we anchor markers by text match in
     the pandoc-produced Typst; the series template gains an `index` back-matter piece (two
     columns, run-in subentries, letter groups). Typst has no built-in indexer — we write one
     (~150 lines: collect `metadata` + `locate`, sort, group, dedupe page ranges) or use the
     `in-dexter` package. Corrections and rebuilds keep the index true.
   - **(B) Post-hoc from the PDF.** Build once, extract text per page, have the model index
     from the paginated text, then set the index and rebuild. But adding the index changes
     nothing before it (back matter), so pages hold — *unless* the author corrects the body,
     which reflows and silently breaks every locator. Cheaper to build; breaks the factory's
     "rebuild any time" promise. Not acceptable as shipped.
   - **(C) EPUB only.** Entries link to anchors, no page numbers. Trivial by comparison, but
     nobody pays for an EPUB index.
   Design A is the one worth doing.

4. **The manuscript side.** Word has index fields (`XE`). We could accept an author's own
   `XE` marks and carry them through — pandoc drops them today, so the Lua filter would need
   to read `w:fldSimple`/`w:instrText` — which gives a "bring your own index marks" tier for
   authors who do it themselves, and a starting list the LLM refines.

## Effort (honest)

| Piece | Effort |
|---|---|
| Typst index piece in the series template (collect, sort, group, ranges, two-column, letter heads) | 1 day |
| LLM pass: chapter-chunked prompt → structured entries (heading, subheading, see/see-also, anchor sentences); merge across chapters; cap by trim/pp (≈ 3–5 % of pages) | 1–2 days incl. prompt iteration on Ghosts + Obliquities |
| Anchoring markers into the pandoc Typst output (fuzzy match, ignore inline markup) | 1 day |
| Factory UI: add-on SKU in the store, "Index" step after Inspect, review the draft entries before build, on/off per build | 1 day |
| Word `XE` field passthrough (optional tier) | ½ day |
| **Total** | **~5 days**, plus Jenna's editorial time judging the drafts |

Cost per book: one full read of the manuscript by a frontier model, ~80–120k input tokens for a
250-page book, well under a dollar; cents with caching. Price as an add-on ($40–80?) against
a human indexer at $3–5/page.

## Recommendation

Feasible and genuinely valuable — it turns a tier of publishing labour that authors skip
because of cost into a checkbox — but it is a five-day build with a new Typst subsystem and a
new model dependency inside the build path. Nothing here is safe to start the Friday before a
workshop. Park as **5.13**, revisit Wed Sep 24+. When we do it: design A, generative model for
entries, Jev optional as a confidence gate on candidate terms, Jenna edits the draft index in a
review step before it goes to build. First test book: Obliquities (real client, essays with
recurring named concepts — a good index-shaped book).
