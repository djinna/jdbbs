# Typst Consistency Dial-In Workflow

Goal: drive the stochasticity out of Typst testing by dialing in one variable at a time, each with a pass/fail gate against a fixed reference.

## Why this exists

Previous attempts stalled because too many things were changing at once. You'd nudge margins, then nudge line height, then realize you couldn't tell which change caused the page to look wrong. This workflow eliminates that by:

1. Locking a **single reference** (GHOSTS PDF) as the unambiguous truth.
2. Moving **one variable at a time** through a fixed sequence, with a visual diff gate at each step.
3. Producing **named artifacts** per iteration so you can always roll back to "step 3, v2".

## Two-part structure

### Part 1 — Reproduce GHOSTS (Protocolized series dial-in)

Take an existing published Protocolized book, lock its PDF as truth, and rebuild it from the Typst pipeline until the output is visually indistinguishable. This exercises *all* the infrastructure without introducing new style demands. When it's green, the pipeline is trustworthy.

### Part 2 — Extend for TY (The Twitter Years) custom styles

Once Part 1 is done, TY is the first book with design demands beyond the Protocolized template — custom paragraph and character styles. Build the custom-styles extension on top of a proven pipeline, not during it.

---

## Part 1 — Reproduce GHOSTS

### Setup (one-time)

- **Reference**: `reference/GHOSTS.pdf`. Don't touch it.
- **Source**: `reference/ghosts_epub/` already has the extracted manuscript. Use it as the input rather than re-converting from DOCX to eliminate pandoc variance.
- **Install Typst** pinned to a specific version (`typst 0.12.0` or whatever's current). Record the version in `dial-in/VERSION.txt`.
- **Install fonts**: Plantin MT Pro, Proxima Nova, Menlo, Hiragino, Thonburi. These are the fonts in `reference/SERIES_DESIGN_SPEC.md`. If any are licensed and unavailable, substitute open equivalents *and document the substitution* — it'll show up in diffs and you'll know why.
- **Test chapter**: pick **one chapter** of GHOSTS, ~15–25 pages, that contains at least one of each tricky element (blockquote, verse, multilingual block, section break, terminal/code block, chapter opener). Don't dial in against the full book until the test chapter is green.
- **Build target**: add `make dial-in STEP=n` to the Makefile that builds the test chapter under a frozen config and writes to `dial-in/step-N-<description>/chapter.pdf`.
- **Diff harness**: a `make dial-in-diff STEP=n` target that renders both the reference chapter (from GHOSTS.pdf) and the built chapter side-by-side into a single HTML page, 50% opacity overlaid, page-by-page. `pdftoppm` + a scrap HTML generator is plenty.

### The 8 dial-in steps (in order)

Each step: (1) change one thing, (2) build, (3) visually diff, (4) lock in if it matches the reference. If it doesn't match, iterate *on that step only* (`step-3-v1`, `step-3-v2`) until it does. Do not move to the next step until the current one is green.

| # | Variable | Measurement / target |
|---|---|---|
| 1 | **Trim + gutter** | 353.811 × 546.567 pt = 124.8 × 192.8 mm. Already done via Phase 3.1 `protocolized` preset. |
| 2 | **Page margins** | Top 0.75" · Bottom 0.75" · Inside 0.7" · Outside 0.6". Measure from `SERIES_DESIGN_SPEC.md` and verify by overlaying rules on the reference. |
| 3 | **Body text** | Plantin MT Pro Regular, ~10pt, line-height 1.2, justified + hyphens, first paragraph no indent, subsequent 0.75em. |
| 4 | **Paragraph rules** | First-paragraph (post chapter-opener, post section-break) no-indent rule. Widow/orphan control (Typst default is usually fine but confirm). |
| 5 | **Heading scale** | Chapter titles: Proxima Nova Bold ~20pt title case, upper left. Author names ~16pt Medium. TOC entries 0.833em. |
| 6 | **Running heads** | Proxima Nova Medium 0.75em small-caps. Verso: page # + author. Recto: story title + page #. |
| 7 | **Section break glyphs** | Three spaced breves `˘ ˘ ˘` centered, 0.5em above/below. |
| 8 | **Chapter opener geometry** | Full-bleed faded background image, title+author upper portion, no page # on opener, forced page break before. |

Other elements — terminal/code, verse, multilingual stacks, TOC layout, front matter flow — fall out of the above once fonts + margins + body are pinned. If any of them don't match after step 8, spin up step 9+ named artifacts for that specific element.

### Pass/fail gate per step

A step is "done" when the side-by-side diff page shows the built chapter and the reference chapter to be **visually indistinguishable at page-level zoom**. Minor rendering differences (font hinting, kerning drift by <0.5pt) are acceptable; line-break or layout drift is not.

If you cannot achieve a match because Typst lacks a feature, write it up in `dial-in/step-N-NOTES.md` with what you tried, what Typst does differently, and whether it's a blocker or acceptable drift. Don't silently give up on a step.

### Anti-patterns to avoid

- "It looks wrong." — **Be specific**: "line 14 of page 42 has an extra space after the em-dash compared to the reference."
- Changing two variables at once to "save time" — this is how the last attempt got stuck. Always one change, always named.
- Deleting old dial-in artifacts. They're cheap and rolling back by reading an old HTML diff is invaluable.
- Starting the full book before the test chapter is green.

---

## Part 2 — Extend for TY custom styles

Once Part 1 is green, TY is the first book whose manuscript has paragraph and character styles beyond what the Protocolized template defines. The app schema already carries a `custom_styles[]` array (Typesetting tab → "Custom Styles" section, currently empty on most projects). This is the hook.

### Step TY-1: Audit TY's manuscript

Pull TY's DOCX (whatever you have in hand — likely `manuscripts/...` or the project's attached files). Extract all unique paragraph and character styles. A 10-line Python script with `python-docx` can list them. Output a table of:

- Style name
- Style kind (paragraph vs character)
- Frequency (how many runs/paragraphs use it)
- Sample passage
- Proposed Typst equivalent

Decide which need distinct Typst treatment and which can collapse into existing template styles (e.g., a "Quote" paragraph style that's really just the existing blockquote can be collapsed).

### Step TY-2: Extend the custom_styles schema

Current schema carries `{name, description, body}`. For real use it probably needs:

```
{
  name: "TweetQuote",
  kind: "paragraph" | "character",
  typst: {
    show_rule: "...",                   // Typst show-rule snippet
    or_function: "...",                 // or a custom element function
  },
  word: {
    style_name: "TweetQuote",           // label in the DOCX
    font: "Proxima Nova",
    size_pt: 10,
    ...
  },
  description: "...",
}
```

Decide the exact shape *against one real TY style* first, not in the abstract. Iterate.

### Step TY-3: Wire Typst side

In `typesetting/templates/series-template.typ` (or a new `ty-template.typ` if divergence gets big), extend the template to read `config.custom_styles` and emit a show-rule per entry. Test against one TY chapter.

### Step TY-4: Wire Word-template side

`typesetting/scripts/generate-word-template.py` already emits styled `.docx`. Extend it to emit paragraph/character styles from `custom_styles[]` so the editor opens Word and sees the TY-specific styles in the gallery.

### Step TY-5: Admin UI

The admin panel already has a "Custom Styles + Add" section. Flesh out its form to match the extended schema so editors can add a new style per book without touching code.

---

## Success criteria

**Part 1 green**: Building the GHOSTS test chapter produces a PDF visually indistinguishable from the reference at page-level zoom, using only the admin UI's spec editor as input.

**Part 2 green**: Adding a new custom paragraph style via the admin UI makes it appear in both the Typst output and the Word template, and a TY test chapter renders with the custom style applied where the manuscript uses it.

Once both are green, the pipeline is trustworthy enough for the real three-book production run.

---

## What this workflow is *not*

- Not a commitment to visually match GHOSTS at the **pixel** level — font hinting alone makes that impossible across machines. Page-level zoom is the gate.
- Not a replacement for the normal PR workflow. Each dial-in step can be its own PR or commit; the `dial-in/` directory is the working log, not the final home of the config.
- Not a full Part 3 for `jdbbs.exe.xyz` client-portal presentation of the built PDFs. That's downstream.
