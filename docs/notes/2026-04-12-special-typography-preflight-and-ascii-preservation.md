# Special typography, preflight review, and ASCII-art preservation — 2026-04-12

Purpose
- Capture the development decision reached after the EPUB-first QA checkpoint.
- Treat the current ASCII-art problem as the first concrete example of a broader class: special typography / layout-sensitive manuscript content.
- Record the recommended workflow now so the same thinking can guide both EPUB and Typst/PDF work.

## Context
During the real-project EPUB-first testing, custom-style identity survived into the EPUB artifact, but a follow-up issue remained: some ASCII art in the manuscript was broken. Rather than treating ASCII as a one-off curiosity, the better framing is that it stands in for any manuscript content whose meaning depends on exact spacing, line breaks, or nonstandard presentation.

Examples in this class:
- ASCII art
- layout-sensitive social-post blocks
- poem-like / preformatted text
- terminal transcripts
- handwritten-style reproductions converted to monospace blocks
- any intentional spacing-dependent arrangement that would be damaged by normalization

## Product/workflow decision
The right long-term workflow is:

1. Development editor flags special typography in the manuscript transmittal.
2. Copyeditor also looks for and flags any such elements during manuscript prep.
3. Each approved special case gets represented explicitly in the production spec as a named custom style / treatment, rather than relying on accidental local formatting.
4. Before either EPUB generation or print/PDF generation, run a manuscript preflight step that outputs:
   - a local-formatting report
   - an image inventory
   - a list of declared/observed special-typography blocks
   - unresolved items needing editorial review
5. Conversion should consume this reviewed information and preserve semantics for both output paths.

This keeps the system review-first rather than silently normalizing away meaning.

## Key principle
ASCII art should not be modeled as "bad spacing".
It should be modeled as intentional preformatted content.

That same principle extends to other special typography: if meaning depends on spacing or block structure, the pipeline should preserve it as a semantic block, not treat it as generic prose that happens to contain extra spaces.

## Recommended data model / process change
### Manuscript transmittal
Add or formalize a section for special typography / nonstandard manuscript elements.

Suggested fields per item:
- label / short name
- manuscript location or chapter
- description of intent
- expected treatment
- Word style to apply
- EPUB treatment
- print treatment
- notes / unresolved questions

This can begin as a simple list field if needed, but the long-term shape should be structured enough to drive template generation and review.

Important distinction:
- "Known manual formatting to review" remains useful for accidental local formatting.
- "Special typography" should be a positive declaration of intentional exceptions.

## Preflight step before any output generation
Add a preflight review step before EPUB and before Typst/PDF (or one shared step required by both).

Recommended product placement:
- Do not run this automatically on upload.
- Upload should remain upload-only.
- The inspection/preflight action belongs in Typesetting, because it is part of output preparation rather than file intake.
- Files should remain the place to upload, link, replace, and manage source files.

Recommended UI pattern:
1. Typesetting gets an explicit `Inspect Manuscript` / `Run Preflight` button near the existing output actions.
2. The generated report should persist and remain revisit-able after the run.
3. The most useful persistent home is a shared Typesetting panel (not the Files tab) because the report informs both EPUB and PDF/print decisions.
4. Files may show a lightweight status hint such as `Preflight: warnings found` or `Last inspected <date>`, but not be the primary review surface.

Reasoning:
- Running automatically on upload would blur the intentionally separate workflow the project has already moved toward.
- A user may upload an early or rough manuscript just to attach/link it; preflight should happen when the team is actually preparing outputs.
- The report is not just about the file as a blob; it is about production decisions affecting both output paths.

Minimum useful outputs:
1. Local formatting report
   - manual bold/italic/underline
   - colored/highlighted text
   - manual lists
   - manual section breaks
   - mixed fonts/sizes
   - direct spacing/indent/alignment overrides
2. Image inventory
   - image count
   - approximate location/order
   - caption presence/absence if detectable
   - alt-text status / placeholder follow-up
3. Special typography report
   - declared items from transmittal/spec
   - detected likely spacing-sensitive blocks
   - unresolved blocks requiring keep/strip/convert/preserve-as-verbatim decisions
4. Pass/warn status
   - initial recommendation: warning-first, not hard-blocking
   - unresolved high-risk items should be very visible before generation proceeds

Recommended persistence model:
- Store the latest preflight report as project/book-associated review data.
- Keep a persistent link from Typesetting to the latest report/output.
- Show summary chips inline, for example:
  - `Preflight passed`
  - `3 warnings`
  - `2 unresolved special typography items`
  - `5 images detected`
- Also surface notable finding-specific chips directly in Typesetting when present, especially:
  - declared custom style in use
  - undeclared custom style
  - ASCII / preformatted block detected
  - script/language flag (for example CJK or Thai)
  - manual bold/italic/underline findings
- EPUB and PDF actions should read from the same latest report status.

Current implementation note:
- `book-production/scripts/detect-edge-cases.py` already covers several local-formatting categories.
- It does not yet emit an image inventory.
- It does not yet treat ASCII art / preformatted blocks as a first-class category.

## EPUB direction
For EPUB, the preservation target for ASCII art and similar blocks should be a dedicated semantic wrapper that renders as preformatted monospace content.

Practical direction:
- represent the block explicitly, not as ordinary paragraphs
- preserve internal spaces and line breaks
- apply monospace styling
- avoid generic prose normalization inside the block

Important nuance:
- For small spacing-sensitive blocks, a preformatted text treatment is appropriate.
- For very wide ASCII art, exact fidelity on narrow EPUB readers may still be difficult. Wrapping destroys the art; forced non-wrapping may overflow small screens.
- Therefore some cases may need a fallback decision: preserve as text when feasible, otherwise treat as an image-like artifact or redesign the presentation intentionally.

This means the pipeline should distinguish:
- ordinary styled text
- preformatted text art
- display elements whose fidelity is too width-sensitive for reliable text reflow

## Typst / print direction
The same semantic decision should feed the Typst path.

Do not create one solution for EPUB and a separate ad hoc solution for print.
Instead:
- identify the block once during preflight/review
- mark it as intentional preformatted/special typography
- emit a dedicated representation for EPUB
- emit a dedicated representation for Typst

That shared semantic layer should make the print fix faster once the EPUB side is solved.

## Guidance on double-space cleanup in Word
Problem:
A global search-and-replace of double spaces to single spaces will damage ASCII art and any other spacing-dependent block.

Best practical solution found:
- put ASCII art (and any similar spacing-sensitive content) in its own dedicated paragraph style
- then run Find/Replace only within body-text styles, not globally across the manuscript

Reasoning:
- Word supports style-scoped Find/Replace via the Find and Replace dialog using format/style restrictions
- this is safer than a document-wide Replace All when the manuscript contains intentional preformatted blocks

Operational recommendation:
1. Create/use a dedicated style such as `ascii-art` / `verbatim-block` / equivalent project style.
2. Apply that style to every spacing-sensitive block before cleanup.
3. Run double-space cleanup only on prose styles (for example Normal/body styles), style by style if necessary.
4. Do not rely on memory or visual scanning to avoid the ASCII sections.

Important caution:
- Using nonbreaking spaces as a protection trick is possible in emergencies, but it is not the preferred editorial workflow.
- It pollutes the source, obscures intent, and can create downstream cleanup problems.
- Prefer explicit styles over invisible character hacks.

## External research note captured for future reference
Two practical references were checked during this inquiry:
- Microsoft Support: general Find/Replace guidance emphasizes reviewing replacements rather than blindly replacing all.
- A Super User answer (captured via StackPrinter) confirms that Word Find/Replace can be constrained by style from the Find and Replace window; this supports the style-scoped cleanup strategy above.

## Strong recommendation
Treat ASCII art as the proving case for a broader "special typography" workflow.

That means:
- transmittal declaration
- copyeditor flagging
- explicit custom-style/spec treatment
- shared preflight report before any output generation
- semantic preservation across both EPUB and Typst
- style-scoped editorial cleanup in Word instead of global blind normalization

## Suggested next implementation steps
1. Add a structured special-typography field to the manuscript transmittal / spec bridge.
2. Extend `detect-edge-cases.py` to:
   - emit image inventory
   - flag likely preformatted / ASCII-art-like blocks
   - surface unresolved special-typography items distinctly from generic local formatting
3. Decide the canonical semantic name(s) for spacing-sensitive blocks in the spec.
4. Implement one preservation path in EPUB for verified preformatted blocks.
5. Reuse the same semantic decision in the Typst path.
6. Update author/editor guidance so cleanup instructions explicitly warn against global whitespace normalization across documents containing special typography.

## Why this note matters
This is not only about one broken ASCII-art block.
It marks a design shift from:
- "clean up odd formatting later"

to:
- "treat intentional layout-sensitive manuscript content as declared, reviewable, semantic production data."
