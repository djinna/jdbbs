---
title: "Book map: front / body / back matter"
summary: "How the build reads your file, and any warnings or notes about it."
audience: authors
group: Inspect findings
status: draft
routes: ["/{client}/{project}/factory/"]
covers:
  - typesetting/scripts/detect-edge-cases.py
  - srv/bookmap_inspect.go
  - srv/preflight.go
last_verified: 1ae536e
owner: jenna
order: 15
---

**Why it's flagged.** Inspect shows how the build will sort your book into front matter, body and back matter, using your Heading 1 titles and where they fall. Warnings and notes describe what it will drop or rename, such as your own title page or contents, which the factory generates.

**How serious.** Usually Worth a look for a warning, Just noting for a note; a problem the map calls an error can show as Worth fixing.

**How to fix it.** In any editor:

1. Check the map matches your book.
2. To move a section, rename or move its Heading 1.
3. Make every section title Heading 1.

**When to leave it.** If the map is what you meant, leave it. The build uses the map as shown.

[All Inspect findings](inspect.md)
