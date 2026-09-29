---
title: "Stray \">\" marks from a pasted email"
summary: "\">\" characters inside your running text, left over from a quoted email or Markdown."
audience: authors
group: Inspect findings
status: draft
routes: ["/{client}/{project}/factory/"]
covers:
  - typesetting/scripts/detect-edge-cases.py
  - srv/preflight.go
last_verified: 1ae536e
owner: jenna
order: 50
---

**Why it's flagged.** A ">" appears inside a paragraph, usually left behind when text was pasted from an email reply or a Markdown file. It will print as a ">" in the book.

**How serious.** Worth a look.

**How to fix it.** In any editor:

1. Find the ">" in your editor and delete it.
2. If it marked a quotation, apply the Block Quote style (or start the paragraph with `[[quote]]`).
3. Upload again.

**When to leave it.** If the ">" is meant, for example an arrow in a code sample or a comparison, leave it.

[All Inspect findings](inspect.md)
