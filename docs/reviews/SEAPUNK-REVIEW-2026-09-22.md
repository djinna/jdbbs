# Seapunk Studios — *We Have Always Been Seapunks* — read-only review, 22 Sep 2026

Client `seapunkstudios`, project 32, pass 17 (Sam Chua). Books 46/47/48/49/50/55, six uploads Mon 21 Sep
16:43–17:58 UTC. Sources: `scratch/seapunk/books/*`, `factory_events` (45 rows), `manuscript_preflights`,
`book_outputs.spec_snapshot`, `pass_ledger`, and the code. House rule: *the factory fixes how a thing looks;
the author fixes what a thing is.* Nothing failed here; this is about what the factory did with a file that
was, by our own template's example, correctly structured.

## 1. Summary

The team did everything right and the factory silently threw their book away — twice. Their manuscript was the
essay title as **Heading 1** with five **Heading 2** sections under it, which is exactly the shape of the Word
template we gave them (`generate-word-template.py:577` types the book title as H1). `srv/bookmap.go:407–415`
has a "legacy template" rule: a first H1 that equals the book title is dropped **together with every paragraph
under it up to the next H1**. With no second H1, that was all 4,170 words and 10 images. Inspect *saw* this —
"dropped with the 61 paragraphs under it" — but filed it as a **low** note among 33 lows and said "ready";
the proof was 6 pages (half-title, title, copyright, empty Contents) and nothing on the factory page shows a
page or word count, so they exported a **final of an empty book at 16:50** (and, four minutes earlier, a final
of the template test, also empty). Two of three finals spent on nothing. Upload 49 fixed it by accident (first
H2 promoted to H1); 50 and 55 were busywork chasing low-severity findings (blank lines round images, direct
spacing, embedded fonts). Their latest proof (55) is a real 26-page book with three things to fix — our
footnote marks all print as ¹ (`series-template.typ:1249`, one-line hotfix), captions come adrift from
full-page images, and one empty Heading 2 they added by accident. At 19:08 they came back and downloaded the
**47 final — the empty one**; they may think that's their deliverable. Fixes are ours: reinstate both finals
(`POST /api/admin/passes/17/grant`), stop the title-heading rule from swallowing a whole book (hotfix), block or
warn on a final when the body is empty (hotfix), and — the structural fix Jenna wants — add a **Chapter Title**
style so H1/H2 can be what authors expect them to be (Thursday, ~1 day). The team should delete the empty
heading, decide whether chapter 1 should really share the book's title, and export one final of 55 once the
footnote fix is live.

## 2. Timeline (UTC, Mon 21 Sep, from `factory_events`)

| When | What | Note |
|---|---|---|
| 15:46 | Pass redeemed (coupon 11), `/seapunkstudios/book-001/` | 3 finals |
| 16:04 | Transmittal → final; template downloaded (38.5 KiB) | custom_styles: none |
| 16:43 | Upload **46** = the template itself, unedited (`SEAPUNK-TEST-BOOK-template.docx`) | Inspect ready 0/1/2 |
| 16:44–16:45 | Proof 46 (1 s) → 6 pp; both files downloaded | body empty |
| **16:46** | **Final 46** → "2 finals left"; both downloaded | template test spent a final; body empty |
| 16:48 | Upload **47** (`…-upload.docx`, 1.8 MiB, 4,170 words, 10 images) | Inspect ready 0/1/33 |
| 16:49 | Proof 47 → 6 pp; epub + pdf downloaded | body empty; same size as 46 |
| **16:50** | **Final 47** → "1 final left"; proof pdf downloaded 6 s later | **final of an empty book** |
| 16:56 | Upload **48** (−1 caption line, −2 blanks) → proof 6 pp | still empty; 0/1/33 |
| 16:59 | Upload **49** (first H2 → **H1**) → proof **26 pp**, epub downloaded | the book appears; 0/1/33 |
| 17:03 | Upload **50** (`3.docx`: 2nd H1 = book title, captions merged into image ¶, blanks removed) → proof 26 pp | 0/1/25 |
| 17:58 | Upload **55** (`4.docx`: captions un-merged, Spectral fonts removed, direct spacing stripped, +1 empty H2) → proof 26 pp; epub downloaded | 0/1/18 |
| 19:08 | Download **final pdf, book 47** | 6 pages, no body text |

Ledger: `pass_ledger` rows 42 (book 46) and 43 (book 47), −1 each. `builds_used = 2`, 1 left.

## 3. Iteration table

Source = python-docx over `word/document.xml`; proof = `pdfinfo`/`pdftotext`/`pdfimages`. Inspect = high/med/low.

| Book | Source: words · non-empty ¶ · images · styles | Proof: pages · words · images | Inspect | Changed vs previous upload | What the proof did in response |
|---|---|---|---|---|---|
| 46 | 900 · 47 · 0 · full template set (H1 ×1, H2 ×3, H3 ×10 …) | 6 · 0 body · 0 | 0/1/2 | — (the template, unedited) | H1 = book title → whole guide dropped. **Final spent.** |
| 47 | 4,170 · 52 · 10 · Normal 46, **H1 ×1**, H2 ×5, `[[quote]]` ×1, 4 footnotes | 6 · 0 body · 0 | 0/1/33 | real manuscript | same 6 empty pages. **Final spent.** |
| 48 | 4,156 · 51 · 10 · same | 6 · 0 body · 0 | 0/1/33 | −"Above: Blade Runner…" caption, −2 blank ¶ | nothing (still no 2nd H1) |
| 49 | 4,156 · 51 · 10 · **H1 ×2**, H2 ×4 | **26 · 4,931 · 10** | 0/1/33 | "How I got enamoured…" H2 → H1 | full book; chapter 1 titled with the essay title |
| 50 | 4,161 · 52 · 10 · H1 ×2, H2 ×5 | 26 · 4,931 · 10 | 0/1/25 | essay title back to H2; new H1 "We Have Always Been Seapunks" as chapter head; 7 captions merged into image ¶; blank ¶ removed (manual_break 9→1) | identical layout; chapter 1 now shares the book title |
| 55 | 4,161 · 52 · 10 · H1 ×2, H2 ×**6 (one empty)** | 26 · 4,931 · 10 | 0/1/18 | captions back to own ¶; embedded Spectral fonts removed (1.90→1.57 MB; `unusual_font` gone); direct spacing stripped (11→1); +1 empty H2 | identical layout; `<h2></h2>` in EPUB, gap in PDF |

Reading: 48 → 49 is the only change that mattered. 50 and 55 were ~50 minutes of chasing lows that do not
affect the build (blank paragraphs are dropped anyway; Word spacing is ignored; embedded fonts are ignored).

## 4. The 47/48 vs 49 anomaly

**Cause (a-variant): the manuscript text lived under a heading the factory drops.** Not text boxes, not tracked
changes, not the spec.

| Candidate | Evidence | Verdict |
|---|---|---|
| (b) spec snapshot changed 48→49 | `diff` of `book_outputs.spec_snapshot` 110 vs 112: only `chapters_suggested` gains "How I got enamoured…" — that field is *derived from* the manuscript's H1s | not a cause |
| (c) source differs materially | 48 vs 49 `document.xml`: 66,440 vs 66,243 bytes; same 10 media files; same footnotes; paragraph-level diff = **one style change** (¶3 Heading 2 → Heading 1) and one blank ¶ removed | the trigger, not a defect in the file |
| (a) text ignored by pandoc | 0 text boxes, 0 content controls, 0 tracked changes, 0 `w:pict` in every source | no |
| **(a′) text dropped by our book map** | Inspect 47 `book_map`: `sections=[{title:"We Have Always Been Seapunks", kind:"title", paras:61, words:4170}]`, summary "Body: no section headings found"; note "dropped with the 61 paragraphs under it" | **yes** |

Mechanism, with lines:
- `srv/bookmap.go:407–415` — if the first H1 normalises to the book title → `kinds[0] = "title"` and a *Note*
  (low) is emitted. Everything up to the next H1 is that span.
- `typesetting/filters/docx-to-typst-enhanced.lua:716–720` — `dropping = (kind == "title" or kind == "toc")`,
  all blocks skipped until the next level-1 Header. `docx-to-epub.lua:128–133` does the same for the EPUB.
- `srv/bookmap_inspect.go:43–51` — notes are hard-wired `severity: "low"`, suggestion "Nothing to do unless
  this is not what you meant." Inspect status stays "ready" with 0 highs.
- `typesetting/scripts/generate-word-template.py:577–580` — the template opens with `add_heading(title, level=1)`
  + "by {author}" and then **Heading 2** "Template Guide" — i.e. the template *demonstrates* book-title-as-H1
  with H2 sections. The guide text says "Every chapter starts with a Heading 1" but the sample contradicts it.
  The team copied the sample's shape, replaced the guide with their essay, and got an empty book. Book 46 (the
  unedited template) also builds empty, which is how the first final was lost.
- `srv/static/factory.js:1597–1603` — the "Export final" confirm says how many finals it costs; it says nothing
  about what the last proof contained. No page count is stored or shown for any output (`book_outputs` has no
  pages column; the proof card at `factory.js:1091–1095` shows only download links and dates).

**Was the 47 final effectively empty? Yes.** `final.pdf`: 6 pages, 59 words, 0 images: half-title, title page,
copyright, "Contents" with no entries. `final.epub`: 4.5 KB, `ch001.xhtml` is 563 bytes. Same for 46. **The
team spent two of three finals on nothing.** Owner: factory. Recommend reinstating both (Jenna decides).

Why 49 worked: with a second H1 the title span shrank to the byline (1 paragraph, dropped correctly) and
"How I got enamoured…" became body chapter 1.

## 5. Source vs proof fidelity on book 55 (their latest)

Front matter (half-title, title, copyright, contents) generated correctly; byline "by Seapunk Studios" dropped
with a note (correct); `[[quote]]` → Block Quote, italic, inline attribution "Mark Fisher" kept; 4 footnotes
present and numbered 1–4 at the foot; 10 images all present, converted to grey for print, colour in EPUB;
running heads recto title / verso author; no italics or bold in the source, so none to lose.

| Symptom | Cause | Owner | Fix |
|---|---|---|---|
| Every in-text footnote mark prints **¹** (p.5, 8, 9, 15) while the notes are ¹ ² ³ ⁴ | `series-template.typ:1249` `show footnote: it => super[#it.numbering]` — `it.numbering` is the pattern `"1"`, not the counter | factory | `#super[#context counter(footnote).at(it.location()).first()]` or just `show footnote: set text(...)`. One line, add to typst test fixture. **Hotfix.** |
| Captions stranded: image fills p.3 / p.7, caption "A punk in Camden…" / "Some excellent examples…" opens the next page | image-only ¶ set at text width (4.2 in) → 526×819 px image is 6.5 in tall; following caption ¶ is unrelated body text | factory | treat the ¶ after an image-only ¶ as a caption (or `[[caption]]`), wrap as `figure(image, caption)` and cap image height at ~65 % of text block so image+caption stay together. `images.typ`, `docx-to-typst-enhanced.lua`. Thursday. |
| Captions set as indented body text, justified and hyphenated ("Chi-ang Mai", "Yan-gon") | no Caption style in the factory set (`apply-style-markers.py:53`, `customstyles.go:27`) | factory | add **Caption** (small, ragged, no hyphenation, no indent) + `[[caption]]` alias. Thursday, same slice. |
| Blank gap before "A hole in the grey curtain" (p.4); `<h2></h2>` in EPUB | ¶11 in 55 is an empty Heading 2 (added between 50 and 55) | both | author: delete it. factory: drop empty headings in the marker pre-pass; Inspect medium "empty heading at ¶11". |
| Chapter 1 is titled "We Have Always Been Seapunks" — identical to the book; Contents has one entry that repeats the title; opener reads title/title/essay-title | author's 50→55 choice (workaround for the drop) | author (with our guidance) | make the essay title the chapter (as in 49) *or* keep it and accept the repetition. Root cause is our H1 rule — see §6 #1. |
| Print images 125–350 effective dpi (4.2 in wide from 526–1456 px) | inventory reports colour only (`image_inventory` finding has `width_in` but no dpi) | factory → author decides | add effective dpi to the finding and a medium at < 200 dpi. Thursday. |
| Fleuron ❧ used as the footnote separator — same ornament the spec picks for section breaks | design choice in `series-template.typ` (footnote rule) | factory (question for Jenna) | use a short rule for the footnote separator, keep the fleuron for section breaks. |
| EPUB images `style="width:4.2in;height:…in"` | pandoc carries Word's physical size | factory | `max-width:100%; height:auto` in `stylesheet1.css`; low. |
| 11 `direct_spacing` lows (47–50) and 9 `manual_break` lows were all the padding around figures; the team spent two uploads removing them | Inspect treats blank ¶ next to an image as a "scene break" and Word spacing as a finding | factory | suppress `manual_break` when a neighbour ¶ is image-only; say "ignored by the build" not "review". Low. |

## 6. Suggestions for the factory (ranked)

1. **Stop the title-heading rule from swallowing a manuscript** — `srv/bookmap.go:407–415`. Only classify the
   first H1 as `"title"` when its span is a byline-sized stub (say ≤ 3 ¶ / ≤ 60 words). Otherwise keep it as
   body chapter 1 and emit a **medium** warning: "Heading matches the book title; kept as chapter 1 — if your
   chapters are Heading 2, make them Heading 1." Add a case to `bookmap_test.go`. Would have given Seapunk a
   full book on upload 47. **Hotfix.** (~15 lines.)
2. **Never let a final build an empty body** — `srv/books.go:343` (before the credit is debited): for
   `kind == final`, read the stored book map (`parseStoredBookMap`) and refuse with 422 when body words == 0
   ("The factory reads no chapter text in this file — see the book map in Inspect"). Client: `factory.js:1597`
   confirm text adds "Your last proof: N pages, body ≈ M words" from the same map (no schema change needed
   today; a `pages` column on `book_outputs` is the proper Thursday version, shown on the proof card at
   `factory.js:1091`). **Hotfix** (server guard + confirm text); page-count column Thursday.
3. **Escalate the Inspect note** — `srv/bookmap_inspect.go:43–51`: a note that drops > N paragraphs is a
   `book_map_warning` (medium), and if the map's body word count is 0 while `m.Words > 0` emit a **high**
   "Nothing will be typeset" so status is not "ready". **Hotfix**, pairs with #1.
4. **Footnote marks** — `typesetting/templates/series-template.typ:1249`, one line (see §5). **Hotfix.**
5. **Fix the template's own example** — `generate-word-template.py:577–585`: don't type the book title as H1;
   open with a sample chapter heading (H1 "Chapter One"), byline gone (generated), "Template Guide" as an H1
   (detect-edge-cases already strips it). Also fix the copy at `/home/exedev/jdbbs-public/factory.html:203`.
   **Hotfix-safe** (Python + copy; no pipeline change) — but see #6 before touching it twice.
6. **Add a `Chapter Title` factory style so H1/H2 are first/second-level heads** (Jenna's ask; Thursday, ~1 day).
   Design that keeps every existing manuscript unchanged:
   - New paragraph style **Chapter Title** in the factory set (`apply-style-markers.py:53`,
     `customstyles.go:27`, `generate-word-template.py:76`), alias `[[chapter]]` → Chapter Title (today it maps
     to Heading 1, `apply-style-markers.py:68`).
   - **Normalisation pre-pass**, not a pipeline rewrite: in `apply-style-markers.py`, if the file has ≥ 1
     Chapter Title paragraph → rewrite Chapter Title → Heading 1 and Heading 1/2/3 → Heading 2/3/4 before
     pandoc. Lua filters, EPUB split (`--epub-chapter-level`), Typst heading styles and TOC all keep working
     untouched. Files with no Chapter Title behave exactly as today (pinstitute, Fotis, Toby untouched).
   - Mirror the same rule in `srv/bookmap.go` (`parseDOCXStyleNames` / `headingLevel`, ~20 lines) since the
     map reads `document.xml` itself; and in `srv/chapter_detect.go:434` (`runPandocToAST` on the raw
     `SourceData`, so it needs the pre-pass or the same shift on the AST).
   - Decide H4: either add an `h4` size/weight to the spec (`headings` block) or fold shifted H3 into H3.
   - Template sample becomes: Chapter Title "Chapter One" → Heading 1 "A-head" → Heading 2 "B-head"; guide
     text and emails (`passes.go:1332/1384/1484/1498`), `factory.html #bring`, `WORD-FREE-AUTHORING` doc updated.
   - Front/back matter (Foreword, Notes…) are Chapter Title too; the vocabulary logic is unchanged.
   - Once live, retire the "legacy template" title-drop rule entirely (drop only the heading + byline, never
     content) — #1 becomes the permanent behaviour.
7. **Figures: caption binding + Caption style + height cap** — `images.typ`, `docx-to-typst-enhanced.lua`,
   EPUB css. Thursday.
8. **Inspect noise around images** — suppress `manual_break`/`direct_spacing` for image-adjacent ¶ and say
   "ignored by the build". Thursday. Add effective dpi to `image_inventory`.
9. **Reinstate the two finals** — `POST /api/admin/passes/17/grant` (+2, note "empty-body finals 46/47,
   factory fault"). Jenna decides; not code.

## 7. Note Jenna can forward to the Seapunk team

Hi Sam — thank you for the six uploads on Monday; watching you iterate that fast taught us more than any test
file. Your latest proof (4.docx, the 26-page one) is the real book: all ten images, the Mark Fisher quote as a
block quote, the four footnotes, the contents page. Two things on our side that you'll see fixed shortly: the
little footnote numbers in the text all print as "1" (ours, one line), and the captions under the two full-page
images slip onto the next page (ours too). On your side: there's an empty Heading 2 just before "A hole in the
grey curtain" — delete it — and decide whether chapter 1 should really carry the same title as the book, or
whether the essay title ("How I got enamoured…") is the chapter, as it was in your third upload. Either is fine.
The two finals you exported early came out with no chapter text; that was our fault, not yours — the factory
threw your text away because your first heading matched the book title. We're changing that, and we're putting
those two finals back on your pass, so you have three again. Wait for the footnote fix, upload once more, read
the proof, then export one final. — Jenna

## 8. Open questions for the lead

- Reinstate 2 finals on pass 17 via `/api/admin/passes/17/grant`? (My recommendation: yes, note the reason.)
- Should the 19:08 download of the empty 47 final trigger an email today, before they print or share it?
- #1 threshold: byline stub = ≤ 3 ¶ / ≤ 60 words? Or simply "never drop content, only the heading"?
- #6 (Chapter Title): does Jenna want the shifted H3 to become an H4 (new spec knob) or merge into H3?
- Footnote separator: keep the fleuron, or a short rule? (Same glyph as the section break in this spec.)
- Do we want the unedited template (book 46) to be un-buildable as a final? It is currently a valid,
  credit-consuming build of nothing.
- Freeze: #1–#5 are all small; are all four hotfix-worthy before Wed 23, or only #2 and #4?
