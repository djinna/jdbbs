---
title: The factory from your own desk
summary: The same six steps as the Factory page, as API calls or a small Python script.
audience: api
status: draft
routes: ["/factory/api"]
covers:
  - jdbbs-public/factory-api.html
  - jdbbs-public/factory-cli.py
last_verified: efed34a
owner: jenna
order: 10
---

Everything the Factory page does, you can do from a script: check the pass,
upload, inspect, build, wait, download. Nothing more. The page and the API
are the same factory.

- **Reference:** [Factory API](/factory/api), six calls with examples and an
  error table.
- **Script:** `factory-cli.py` (Python, standard library only), linked from
  the reference. `inspect ms.docx` and
  `build ms.docx --out ./out --kind proof`.
- **Access:** a per-book token the studio issues, and a live Factory Pass.

This page will become the full API section of Help.
