---
title: "Custom styles not in your transmittal"
summary: "Your file uses a paragraph or character style the transmittal doesn't list."
audience: authors
group: Inspect findings
status: draft
routes: ["/{client}/{project}/factory/"]
covers:
  - typesetting/scripts/detect-edge-cases.py
  - srv/preflight.go
last_verified: 1ae536e
owner: jenna
order: 10
---

**Why it's flagged.** Your file uses a custom style, such as "Letter" or "Recipe", that isn't declared in the transmittal. Inspect lists each one once. The build only knows the factory's thirteen styles, so it can't treat an unlisted one on purpose.

**How serious.** Worth fixing.

**How to fix it.** In any editor:

1. Declare the style in the transmittal's **Custom styles**, based on one of the factory styles.
2. Or apply a factory style instead, such as Block Quote or Verse.
3. No styles in your editor? Start the paragraph with a marker, like `[[quote]]` or `[[style:Letter]]`, and see [Send your manuscript](send-your-manuscript.md).

**When to leave it.** If the style is only a leftover from pasted text and you don't care how it looks, you can leave it and build a proof to see.

[All Inspect findings](inspect.md)
