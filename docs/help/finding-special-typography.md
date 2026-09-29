---
title: "Spacing-sensitive or preformatted content"
summary: "Text that looks like ASCII art or a preformatted block, where the spacing matters."
audience: authors
group: Inspect findings
status: draft
routes: ["/{client}/{project}/factory/"]
covers:
  - typesetting/scripts/detect-edge-cases.py
  - srv/preflight.go
last_verified: 1ae536e
owner: jenna
order: 40
---

**Why it's flagged.** Lines look like ASCII art or preformatted text, where the exact spacing matters. Ordinary paragraph settings can squash it.

**How serious.** Worth fixing when the block isn't tied to a declared style; Just noting when it is.

**How to fix it.** In any editor:

1. Apply the **Code Block** style to the lines, or put `[[code block]]` before the first paragraph and `[[/code block]]` after the last.
2. If it's a picture of text, consider using an image instead.
3. Upload again.

**When to leave it.** If it's already in a style you declared for it, leave it.

[All Inspect findings](inspect.md)
