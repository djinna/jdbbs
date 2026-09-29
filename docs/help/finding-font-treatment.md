---
title: "Emoji or special font use"
summary: "Emoji or a special font in your text."
audience: authors
group: Inspect findings
status: draft
routes: ["/{client}/{project}/factory/"]
covers:
  - typesetting/scripts/detect-edge-cases.py
  - srv/preflight.go
last_verified: 1ae536e
owner: jenna
order: 150
---

**Why it's flagged.** Inspect saw emoji or a special-purpose font. These may not print the way they look on screen, because the book uses its own fonts.

**How serious.** Just noting.

**How to fix it.** In any editor:

1. Check the marked paragraphs in a proof.
2. If you don't need the emoji or symbol, replace it with plain text.
3. Tell the studio in the transmittal's Special Characters box if it matters.

**When to leave it.** If you can see it in the proof and it looks right, leave it.

[All Inspect findings](inspect.md)
