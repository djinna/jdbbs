---
title: "Marked styles ([[style]] markers)"
summary: "The [[markers]] Inspect found at the start of paragraphs, and what each becomes."
audience: authors
group: Inspect findings
status: draft
routes: ["/{client}/{project}/factory/"]
covers:
  - typesetting/scripts/detect-edge-cases.py
  - srv/preflight.go
  - docs/help/send-your-manuscript.md
last_verified: 1ae536e
owner: jenna
order: 20
---

**Why it's flagged.** Inspect found `[[marker]]` text at the start of a paragraph, the way to ask for a style when your editor can't hold one. A marker the factory recognises is applied and removed in the build. A marker it doesn't recognise, and that isn't declared in your transmittal, prints exactly as you typed it.

**How serious.** Carried through automatically or Just noting when the marker is recognised (it shows what it becomes); Worth a look when it isn't.

**How to fix it.** In any editor:

1. Recognised markers: nothing to do. Check the finding says the style you meant.
2. Unrecognised: fix the spelling, using a name from [Send your manuscript](send-your-manuscript.md), or declare the name as a custom style in the transmittal.
3. To cover several paragraphs, close the run with the matching end marker, like `[[/code block]]`.

[All Inspect findings](inspect.md)
