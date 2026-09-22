# Fotis (hermescorp, project 30) — *helo word*, book 57 — failed proof review

2026-09-22, read-only. Third pass of the Sam Khoo (45) / Toby Shorin (53) pattern. Sources: his two files
(`scratch/fotis/`), the failed build dir (`/tmp/prodcal-failed/book-57/`), DB (`books`, `factory_events`,
`manuscript_preflights` 58–60, `transmittals` 23), and the code. House rule applied throughout: *the factory fixes
how a thing looks; the author fixes what a thing is.*

## 1. Summary for Jenna

The build failed for a reason that was **entirely ours**: he declared a custom style called `break`, we generated
`#let break(content) = …` into `book.typ`, and `break` is a Typst keyword — the typesetter stopped on line 76 of
*our* preamble, 30 lines before his text begins. The error message then told him to search *his manuscript* for
`} break(content) = {`, a string that exists only in our file, and to look at High Inspect findings (he had zero).
That ident bug is **already fixed and live** (9258e56, restarted 13:49 UTC): `break` → `break-style`. If he presses
Build now it will compile — but the proof will look wrong in five ways, three ours (italics dropped from his
`ITAL` runs, 77 fleurons between stanzas because `[[break]]` = Section Break = ornament, indented poem lines
split out of their poem block) and two his (his `Narration` Word style is based on `Quote`, so pandoc wraps it in a
block quote; his preface is a Heading 2 with no “Preface” Heading 1, so it is typeset as a centred dedication).
He did the mise en place unusually well — 953 markers, every one resolved — but he **over-declared**: 21 custom
styles, of which 9 duplicate factory styles or Word built-ins and 3 exist only because Inspect told him to declare
them. What he should do today: prune the transmittal to ~8 declarations (§6), switch Section break to *White space*,
rebase one Word style, retitle one heading, rebuild. Fifteen minutes.

## 2. Timeline (UTC, from `factory_events` + `books` + `transmittal_versions`)

| When | What | Note |
|---|---|---|
| Mon 15:46 | Upload `helo_word.docx` → **book 44** | never Inspected or built |
| Mon 16:49 → Tue 11:58 | 12 transmittal saves | custom styles grow 0 → 4 (`center poem BOLD ITALICS`) → 11 → 19; `stanza` renamed to `break` at 11:53 (`stanza` would have compiled) |
| Tue 10:53 | Upload `helo word final.docx` → **book 56** (983 ¶, 5 `---` breaks, one image) | |
| 11:58 | Transmittal final; 11:59 template downloaded | he did not write into it |
| 12:02 | Inspect 56: **3 high** / 4 med / 88 low | highs: coloured text ¶1, undeclared `footnote reference`, undeclared `ITAL` |
| 12:11 | Upload `helo word — marked_2.docx` → **book 57** (953 ¶, title page + `---` removed, markers added) | |
| 12:11 | Inspect 57: **2 high** / 2 med / **1027 low** | 953 of the lows are one row per `[[marker]]` |
| 12:12 | Transmittal draft→final: `ITALICS`→`ITAL`, + `footnote reference`, + `footnote text` (as *paragraph* styles) | declared to silence the two highs |
| 12:12 | Inspect 57: **0 high** | green light |
| 12:13, 12:14 | Build proof ×2 → **failed in ~1 s** each | Typst syntax error in our preamble |
| 12:19 | Template downloaded again | probably wondering whether he must write into it |
| 13:45 / 13:49 | 9258e56 committed / service restarted | `break` → `break-style`; he has not retried |

The file Jenna forwarded (`marked_2`, 46 KB) is one save **later** than the book-57 upload (131 KB): it adds
`[[footnote reference]]` in ¶49, `[[footnote text]]` inside the footnote, a Word `Preface` style, and drops an
orthaned image. Otherwise identical.

## 3. What he submitted

**Word styles on paragraphs (upload 57):** Normal 908 · Heading 2 14 · List Paragraph 12 · Narration 10 · Heading 1 5
· Quote 3 · part 1. Character styles on runs: ITAL 6 · BOLD 1 · footnote reference 1 (a real footnote).

**Markers (all at paragraph start unless noted):** `[[verse]]` 762 · `[[break]]` 77 · `[[in1]]` 43 · `[[heading 2]]`
13 · `[[narration]]` 10 · `[[in2]]` 9 · `[[in3]]` 8 · `[[heading 1]]` 5 · `[[bullet3]]` 5 · `[[sectionbreak]]` 5 ·
`[[bullet2]]` 4 · `[[bullet1]]` 3 · `[[quote]]` 3 · `[[in4]]` 3 · `[[preface]]` `[[in5]]` `[[in6]]` 1 each — **953,
all resolved** (762 factory, 80 alias, 111 declared). Inline, mid-paragraph: `[[ITAL]]…[[/ITAL]]` ×9,
`[[BOLD]]…[[/BOLD]]` ×2 — not a supported form yet (5.26).

He used the markers exactly as the `/factory#bring` page says, and he distinguished stanza breaks (`[[break]]` ×77)
from section breaks (`[[sectionbreak]]` ×5, where book 56 had `---`). Belt and braces: `[[heading 1]]` on
paragraphs already styled Heading 1; `[[ITAL]]` markers *and* the ITAL character style *and* direct italic on the
same runs.

**Declarations vs reality (21 declared):**

| Declared | Used? | Verdict |
|---|---|---|
| `verse`, `quote`, `break`, `breaksection` | ✓ (as markers) | **Redundant** — factory style/alias wins in the resolver (`verse`→Verse, `quote`→Block Quote, `break`→Section Break); the declared copies were never called. `breaksection` unused (he typed `[[sectionbreak]]`). |
| `part` (“heading 1”), `chapter` (“heading 2 – each poem”) | ✗ (he used `[[heading 1/2]]`) | **Redundant** — and `chapter` is a factory alias for *Heading 1*; in his parts book that would have been wrong. Parts=5 on the checklist already does this. |
| `footnote reference`, `footnote text` | ✓ (Word built-ins, 1 footnote) | **Wrong kind** — Word built-ins declared as *paragraph* styles because Inspect flagged `footnote reference` High. Nothing to declare. |
| `BOLD`, `ITAL` (character) | ✓ 1 + 6 runs | **Redundant with plain bold/italic** — but see F3: today the declaration is the *only* thing that keeps the runs from going plain, and it renders them wrong. |
| `narration` | ✓ 10 ¶ | **Real** — prose between poems. Needs Based on: Normal (and his Word style rebased, §4 A1). |
| `in1`…`in6` | ✓ 43/9/8/3/1/1 | **Real** — indent levels inside poems. 5.25 gives 2 levels on Verse; he uses 6. |
| `bullet1`…`bullet3` | ✓ 3/4/5 | **Harmless** — Word's own list levels survive pandoc→Typst as nested `- ` items; the declaration adds nothing. |
| `preface` | ✓ 1 ¶ (a Heading 2) | **Wrong tool** — a preface is a Heading 1 named “Preface”, not a style. |

## 4. Factory vs mise en place

**F = factory · A = author · B = both.** Line numbers are HEAD (9258e56).

| # | Symptom | Cause | Owner | Fix |
|---|---|---|---|---|
| F1 | Build fails: `expected pattern, found keyword break` at `book.typ:76` | `typstStyleIdent` emitted a Typst keyword as a function name | F | **Done** — `srv/bookspecs.go:809`, `srv/customstyles.go:50`. Extend the reserved set to Typst built-ins pandoc's writer emits (`quote text par heading footnote link emph strong block box image figure table list enum sub super strike underline`) **and** every function `series-template.typ`/`styles.typ` export (`poem epigraph signature sc ital bold bold-ital …`). Simplest: prefix every declared ident (`cs-break`) — the Lua filter already reads the ident from `pandoc-meta.json`, so one change. His `quote` shadowed Typst's `quote` (harmless here because the filter emits `#blockquote`); his `ital`/`bold` shadowed ours (F3). |
| F2 | “Near: `} break(content) = {` (search for this in your manuscript)” | `diagnoseBuildFailure` quotes any `book.typ` line, including our preamble; `typstLineText` strips `#let`, prepends the previous `}` | F | `srv/builderr.go:63-67`: if the failing line is before `#show: book.with(` (or in `templates/*.typ`), it is **our** error — no *Near*, different message (§5 text A). `srv/builderr.go:117-140`: never quote lines that contain `#let`/`=`. `srv/static/factory.js:410` drop “search for this in your manuscript” when *Near* is absent. Also surface Typst's `hint:` line in the detail. |
| F3 | His 6 ITAL runs and 1 BOLD run will print as **roman 0.9 em** (or plain, post-9258e56) | pandoc `+styles` replaces native Emph/Strong with a `Span custom-style=ITAL` (verified: `-t native`); the filter's `char_style_map` sets `ital`/`bold` to `nil` “let pandoc handle” — pandoc doesn't. Declared char style → `#let ital(content) = content` overrides factory `ital` | F | `typesetting/filters/docx-to-typst-enhanced.lua:219-224`: map `italic/ital/emphasis` → `ital`, `bold/strong` → `bold`. `srv/customstyles.go:33`: add `bold` to `customStyleCharParents`; default `based_on` from the name (`ital*`→italic, `bold`→bold). Never emit a `#let` whose ident equals a factory function (F1). |
| F4 | 77 stanza breaks will print as **77 fleurons ❧** | `[[break]]` is an alias for Section Break (`apply-style-markers.py:67`, lua `:316`); his transmittal chose Section break = Ornament. The `˘` in every Section Break ¶ is **ours** (`SECTION_BREAK_PLACEHOLDER`, `apply-style-markers.py:73`), inserted because pandoc drops empty paragraphs; the filter ignores it. | B | Today: he switches Section break to *White space*. Factory: a poetry book needs an ornament-free **stanza break** — a Section Break *between two Verse blocks* should be a gap, or add alias `stanza` → gap. `section-break-gap` (`series-template.typ:330`) always prints the mark unless config is `blank`. `/factory#bring` says “`[[break]]` for a scene break” — poets read it as stanza. |
| F5 | Indented lines (`in1`…) will sit in their **own poem block** with 0.5 em padding above and below, breaking the stanza | `Blocks()` coalesces only paragraphs with the *same* func (lua `:363-386`); a Verse-based `in1` is func `in1`, not `poem` | F | Coalesce Verse-derived styles **into** the neighbouring `poem` block and indent per line (`#h(1.5em)` / `pad(left:)` on the line), so `#poem[…]` stays one block per stanza. Blocks Toby's `verse2` verification too. |
| F6 | 6 indent levels needed, 2 offered | 5.25 caps `indent` at 2 (`customstyles.go:161`) | B | Ask: is 3 levels enough for him (in4–in6 are 5 lines total)? Or lift the cap to 4 for Verse only. Today: in3–in6 → level 2. |
| A1 | His 10 `Narration` ¶ come out as `#narration[#blockquote[…]]` — italic, indented, plus a spurious “manual list item” comment | His Word `Narration` style is *based on Quote*; pandoc treats the Quote ancestry as a block quote. The pre-pass keeps an existing Word style as-is (`apply-style-markers.py:232 ensure_style`) | A (F polish) | Author: Modify Style → Narration → *Style based on: Normal*. Factory polish: when a marker resolves to a style that already exists in the file, rebase it on Normal and clear its indent (we own the look); or in `Div` (lua `:407`) unwrap a lone BlockQuote inside a declared style. `manual_lettered` regex `^[a-zA-Z][%.%)%s]+.+` (lua `:473`) fires on every sentence beginning “I …” / “A …” — comment-only noise, but it is in every poetry book. |
| A2 | Preface (11 ¶) typeset as a **centred dedication** before the Contents | “Traumagnetic” is a Heading 2 + `[[preface]]`; the pre-pass turned it into a body style, leaving 11 untitled ¶ before the first Heading 1 → `front-piece(kind: "dedication")`. The book map only knows a preface by a Heading 1 reading “Preface” (`bookmap.go:153`) | A (F wording) | Author: Heading 1 “Preface” (keep “Traumagnetic” as a Heading 2 under it). Factory: connect the two Inspect rows (§5 I4) — the warning said “no Preface heading”, the note said “kept as dedication”, neither said “these are the same 11 paragraphs”. |
| F7 | Inspect made him declare `footnote reference` and `ITAL` as custom styles (2 Highs) | `BUILTIN_CHARACTER_STYLES` (`detect-edge-cases.py:63`) lacks Word's built-ins (`footnote reference`, `hyperlink`, `* Char`); a char style that is only bold/italic is flagged as if it were a design decision | F | Add the Word built-ins; for a character style whose definition is only `w:i`/`w:b`, say “this is plain italic — nothing to declare” (Low) once F3 lands. |
| F8 | Inspect: **953 Low rows**, one per marker; the inline `[[ITAL]]`/`[[BOLD]]` (11 markers that *will* print) not reported at all | `detect_style_markers` (`detect-edge-cases.py:155-190`) emits per marker and only sees openers at paragraph start | F | Aggregate per name (“`[[verse]]` ×762 → Verse”); add a finding for `[[…]]` seen mid-paragraph: “inline markers arrive in 5.26; until then use Ctrl+I”. |
| F9 | Inline `[[/ITAL]]` closers at paragraph end were **silently eaten**, openers kept (`[[ITAL]]Justin Case`) | `CLOSE_RE` matches any `[[/x]]` at paragraph end; an unmatched name falls back to `open_stack[-1]` — the `[[verse]]` on the same line (`apply-style-markers.py:165-178`) | F | Until 5.26: a closer whose name matches no open opener should be left alone (and reported). 4 of 9 `[[/ITAL]]` and 1 of 2 `[[/BOLD]]` vanished in `input.docx`. |
| F10 | Transmittal let him declare `verse`, `quote`, `break`, `chapter`, `footnote reference` without comment | No check against factory names, aliases, Word built-ins, or reserved words (`srv/static/transmittal.js:1409`) | F | Inline note on the row: “`verse` is already a factory style — you can use `[[verse]]` without declaring it”; block reserved/keyword names outright with a friendly reason. |
| F11 | Template guide still says Verse is “one paragraph, Shift+Enter for soft returns” | `generate-word-template.py:712` predates 3411728 (one paragraph per line, merged by the filter) | F | Update the sentence; his file follows the new contract, the template contradicts it. Also: 19 custom-style samples that look identical to Normal (`:772-800`) — with 5.25 they can carry `basedOn` + indent. |
| F12 | “Transmittal lists Notes but no Notes heading was found” (Medium) | He ticked Notes = *ft/end* meaning footnotes; `bookmap.go:576` expects a heading | F | Skip the warning when the Notes row's subtype is footnotes. |

## 5. Error-message review

### 5a. The build-failed text he saw

> That proof failed: The typesetter stopped on this file. Run Inspect on it and look at the High findings first; if
> the text quoted below is the problem, edit that paragraph and rebuild. Otherwise email us the message — failed
> builds aren't counted. Near: `} break(content) = {` (search for this in your manuscript) ▸ Technical detail:
> expected pattern, found keyword `break`. Proofs are free, so nothing was used. Stuck? Email …

What went wrong for him: (1) it was our bug, said as if it were his; (2) “look at the High findings” — he had none;
(3) the quoted string is not in his manuscript; (4) “failed builds aren't counted” + “nothing was used” say the
same thing twice; (5) the useful hint from Typst (“try `break_`”) was dropped. Source: `srv/builderr.go:85`,
`factory.js:410,651`.

**Proposed text A — error in our generated template/config** (failing line before `#show: book.with(`, or in
`templates/*.typ`, or a `#let`/`import` line):

> **That proof stopped in our typesetting setup, not in your text.** Nothing to change on your side — we've been
> emailed the details and will fix it and rebuild for you. Proofs are free, so nothing was used. If you're at the
> workshop, wave; otherwise we'll reply to {email}.
> ▸ Technical detail (for us): `book.typ:76 — expected pattern, found keyword \`break\``

**Proposed text B — error in the author's text** (failing line inside the body, prose quoted):

> **That proof stopped at one paragraph of your text.** Search your manuscript for the words below, look at that
> paragraph for anything unusual (a stray `[[marker]]`, an `@`, a pasted symbol), edit it, and build again — proofs
> are free. If it looks fine to you, email us this message and we'll take it from there.
> Near: “{prose}”
> ▸ Technical detail: `{first error line}` — {typst hint, if any}

Rule for choosing: *Near* is only ever prose; if we cannot quote prose, use text A.

### 5b. Inspect findings for his file — current text → proposed

| # | Current (severity) | Problem in his case | Proposed |
|---|---|---|---|
| I1 | `undeclared_custom_style` **High**: “Custom style is used in the manuscript but not declared in the project spec/transmittal. Add it before production so EPUB and Typst can treat it intentionally.” (`detect-edge-cases.py:656`, `preflight.go:284`) — for `footnote reference` and `ITAL` | Made him declare a Word built-in and a plain italic. “project spec”, “treat it intentionally” is our jargon. | “Your file uses a style we don't know yet: **ITAL** (6 places, e.g. ¶439). If it's just italic or bold, nothing to do — we keep it. If it means something more (a term, a title, a command), add it under **Custom styles** on your transmittal and tell us what it's based on.” Word built-ins: no finding. |
| I2 | `style_marker` Low ×953: “Becomes Verse (via factory ‘verse’), 1 paragraph. To cover a run, put [[/verse]] at the end of the last paragraph.” | 762 identical rows bury the two rows that matter. | One row per name: “**[[verse]]** ×762 → Verse. **[[break]]** ×77 → Section Break — in a poem you may want *White space* as your Section break, or these will print as ornaments. **[[in1]]** ×43 → in1 (your custom style, based on Verse, one level in).” |
| I3 | *(none)* — inline `[[ITAL]]…[[/ITAL]]` not detected | 11 markers will print in the proof, unannounced. | “**[[ITAL]]** appears in the middle of 9 lines (e.g. ¶334 *Justin Case*). Markers work at the start of a paragraph today; inside a line, use plain italic (Ctrl+I) for now — inline markers are coming this week. Until then these print as written so you can find them.” |
| I4 | `book_map_warning` Medium: “Transmittal lists Preface but no “Preface” heading was found.” + `book_map_note` Low: “1 untitled page before the first heading; the transmittal lists none. All are kept, in order, as dedication.” | Two rows, one problem; neither says what will happen to his 11 paragraphs. | One row, Medium: “You ticked **Preface**, and there are 11 paragraphs before your first Heading 1 that read like one (“Traumagnetic…”). We'd set them as a dedication page. If they're your preface, give them a Heading 1 that says **Preface** and they'll get their own pages and a Contents line.” |
| I5 | `book_map_warning` Medium: “Transmittal lists Notes but no “Notes” heading was found.” | His notes are footnotes (he said so: *ft/end*). | Drop when subtype is footnotes; else “You ticked Notes as end-notes but we found no **Notes** heading — footnotes are fine as they are; for end-notes, add a Heading 1 “Notes” at the back.” |
| I6 | `manual_formatting` Medium ×3: “Manual emphasis detected — preserve it as intentional bold/italic formatting in EPUB and Typst” | Reads like a to-do; it is fine. | Low: “Bold/italic typed directly (Ctrl+B / Ctrl+I) in 3 places — that carries straight through. Nothing to do.” |
| I7 | `declared_custom_style_used` Low: “Declared custom style is present in the manuscript and ready for intentional EPUB/Typst handling.” | Jargon; and for `footnote reference` it confirmed a mistake. | “**narration** (declared, based on Normal) — found in 10 paragraphs. ✓” |
| I8 | Unresolved marker (he had none; he will if he deletes a declaration and keeps its marker): “Not a factory style and not declared on the transmittal — will print as written. Use one of: …” | Already reworded in the design note. | Use the design-note text: “You've used `[[in3]]` in 8 paragraphs — we don't know that one yet. Add it under **Custom styles**…”. |

## 6. What to tell Fotis today (note Jenna can send)

> Fotis — that failed proof was our bug, not yours: you named a style `break`, which is a reserved word in our
> typesetter, and our error message wrongly pointed you at your manuscript. Sorry. It's fixed; you can build now.
> Your markers were spot on — all 953 resolved. Five quick things before you press Build so the proof looks right:
> 1. **Transmittal → Custom styles:** delete `verse`, `quote`, `break`, `breaksection`, `part`, `chapter`,
>    `preface`, `footnote reference`, `footnote text`, `BOLD`, `ITAL` — the factory already has those (`[[verse]]`,
>    `[[quote]]`, `[[break]]` work without declaring). Keep `narration` (based on Normal), `in1` (based on Verse,
>    indent one level), `in2` (Verse, two levels), and for now set `in3`–`in6` to Verse, two levels too. `bullet1–3`
>    can stay.
> 2. **Transmittal → Typography → Section break: White space.** Your 77 `[[break]]` stanza gaps would otherwise
>    each print a ❧.
> 3. **In Word:** select the ITAL/BOLD runs, Ctrl+Space (clears the character style), then Ctrl+I / Ctrl+B; delete the
>    inline `[[ITAL]]`…`[[/ITAL]]` markers — plain italic carries through. Inline markers arrive later this week.
> 4. **In Word:** Modify Style → *Narration* → “Style based on: Normal” (it's based on Quote today, which makes
>    it an indented italic quotation).
> 5. **In Word:** make “Traumagnetic” a Heading 1 that reads **Preface** (keep “Traumagnetic” as a Heading 2 under
>    it) and remove `[[preface]]`. Then Inspect (free) and Build a proof (free).

## 7. Open questions for the lead

- **F5 is the big one for today's 5.25 verification:** a Verse-based declared style breaks the poem block. Toby's
  `verse2` will show the same padding gaps. Merge into the parent block with per-line indent?
- Indent cap: 2 levels vs his 6 (`in4`–`in6` = 5 lines total). Lift to 3–4 for Verse, or tell poets to flatten?
- Stanza break as a first-class thing: alias `stanza` → gap only, or “Section Break between two Verse blocks = gap”?
  His transmittal even has `section_break_text: "*.*"` typed under *Ornament* — he wanted his own mark for real
  section breaks and a plain gap for stanzas; today those are one control.
- F1 scope: reserved-word list vs. prefixing every declared ident (`cs-…`). Prefixing also ends the `ital`/`bold`
  collision for good; `md-to-chapter.py style_to_typst_func` must follow.
- Should the pre-pass rebase an *existing* Word style it resolves to (A1) — “we own the look” — or is touching
  the author's style definitions a step too far?
- With Parts on, a front-matter Heading 1 “Preface” renders through `part-opener`. Is that acceptable for a proof,
  or should `kind: front` sections use `chapter-opener` regardless of `config.parts`?
- Fotis downloaded the template twice and never used it. Worth a line on the Floor/transmittal: “you don't need
  to write into the template if your markers resolve”?
- Un-Inspected book 44 and book 56 (image, `---` breaks) are dead uploads — fine to leave; noting so the QC mail
  isn't confusing.
