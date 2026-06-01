# Licensed fonts (print-only)

This directory is gitignored except for this README. Drop OTF/TTF files
here under per-family subdirectories, e.g.:

```
licensed/
  plantin-mt-pro/    PlantinMTPro-Regular.otf, …-Italic.otf, …-Bold.otf, …
  proxima-nova/      ProximaNova-Regular.otf, …
```

These fonts are licensed for **desktop print use only** — the same
perpetual-desktop model used in InDesign/Quark for decades. Print-PDF
embedding (subset, rendered output) is covered. EPUB redistribution is
not. They MUST NOT be:

- committed to git (the repo / GitHub is a distribution channel);
- embedded in EPUB output (the zip is a redistribution vector);
- shared with anyone outside the license holder.

## How they get used

`typst fonts --font-path typesetting/fonts` scans this tree recursively,
so once a family is dropped in (and synced to the VM) it appears in the
admin **Typesetting → PDF / Print → Body/Heading Font** menus
automatically. Selecting it sets `typography.body_font` /
`typography.heading_font`, which flow only into the Typst (PDF) compile.
The EPUB pipeline never reads those fields and a runtime guard in
`srv/epub.go` refuses any font path containing `/licensed/`.

## To populate

1. Drop the OTF/TTF files into the per-family subdirectories above.
2. Run `scripts/sync-licensed-fonts.sh` to mirror them to the VM.
3. (Optional) `fc-query <file>` to confirm the exact family name Typst
   will report, then pick that name in the admin font menu.

See `docs/TRACKER.md` → TRK-DESIGN-002 for the full rationale.
