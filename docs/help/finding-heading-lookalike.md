---
title: "Headings without a Heading style"
summary: "A line that looks like a heading but is styled as body text, so the factory can't find it."
audience: authors
group: Inspect findings
status: draft
routes: ["/{client}/{project}/factory/"]
covers:
  - typesetting/scripts/detect-edge-cases.py
  - srv/preflight.go
last_verified: 1ae536e
owner: jenna
order: 30
---

**Why it's flagged.** A paragraph looks like a chapter or section title (bold, larger, or "Chapter 3") but carries a body style. The factory finds chapters by Heading styles, so it can't see this one. It may be missing from your contents, or run into the chapter before it.

**How serious.** Worth fixing when your file uses no Heading styles at all; usually Worth a look when it does.

**How to fix it.** In any editor:

1. Select the line and apply the **Heading 1** paragraph style for a chapter title (Heading 2 or 3 for sub-heads).
2. Remove any bold or larger size you added by hand to make it look like a heading; the book sets its own.
3. Upload again and choose **Inspect again**.

**When to leave it.** If the line really is body text that happens to be bold, such as a run-in label, leave it.

[All Inspect findings](inspect.md)
