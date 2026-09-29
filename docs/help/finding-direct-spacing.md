---
title: "Spacing/indent/alignment applied by hand"
summary: "Spacing, indent or alignment set on a paragraph directly, not by a style."
audience: authors
group: Inspect findings
status: draft
routes: ["/{client}/{project}/factory/"]
covers:
  - typesetting/scripts/detect-edge-cases.py
  - srv/preflight.go
last_verified: 1ae536e
owner: jenna
order: 130
---

**Why it's flagged.** A paragraph has its own spacing, indent or alignment instead of getting it from a style. The book sets its own spacing, so hand-set values may not carry over as you expect.

**How serious.** Just noting.

**How to fix it.** In any editor:

1. Clear the hand-set spacing or indent (your editor's "clear formatting" or paragraph reset).
2. Apply the paragraph style that matches what the text is: Block Quote, Verse, Epigraph and so on.
3. Don't use extra Enter presses or spaces to push text around.

**When to leave it.** If you can't tell any difference on the page, leave it. Build a proof to check.

[All Inspect findings](inspect.md)
