---
title: "Unusual fonts (ignored by the factory)"
summary: "A font in your file that the factory won't use."
audience: authors
group: Inspect findings
status: draft
routes: ["/{client}/{project}/factory/"]
covers:
  - typesetting/scripts/detect-edge-cases.py
  - srv/preflight.go
last_verified: 1ae536e
owner: jenna
order: 90
---

**Why it's flagged.** Your file uses a font other than the usual body font, one entry per font. The factory sets everything in the book's face, so fonts in your file are ignored.

**How serious.** Usually Just noting; Worth a look for a monospaced font, which often means computer text.

**How to fix it.** In any editor:

1. Computer text or code: apply the Code Block style, or put `[[code block]]` at the start of the first paragraph and `[[/code block]]` at the end of the last.
2. Otherwise there's nothing to do.

**When to leave it.** If the font is a paste leftover, leave it. The book won't show it.

[All Inspect findings](inspect.md)
