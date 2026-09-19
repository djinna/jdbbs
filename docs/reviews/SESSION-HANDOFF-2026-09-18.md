# Session handoff — Fri 18 Sep 2026 (punch-list block, pre-workshop)

Workshop Mon/Tue 21–22 Sep; talk Wed 23 Sep 17:30 UTC ("Have your factory
call my factory", deck `~/jdbbs-public/2026-pi-symposium/factory-talk.html`;
never edit `talk.html` there). Store flips to live Stripe Wed 23 00:00 HKT by timer.

## How we work
Shared punch list https://jdbbs.exe.xyz:8766/ (source `scratch/run/CHECKLIST.md`,
notes `scratch/run/notes.json`, relay tmux `runpage`, restart with
`RUNPAGE_CHAT_CONV=<conversation id>`). §0 = Jenna's inbox; answer ON the item
(`POST localhost:8766/note {"id","text","who":"shelley"}`); tick with sed;
`python3 scripts/punchlist-export.py 2026-09-18` → commit
`docs/runs/PUNCHLIST-2026-09-18.md` + `docs/runs/README.md`. **Jenna's rule
(17:45 UTC): work through the whole list — pick a smart order, do not triage
it down; watch context and compact at good stopping points.**

## State at handoff
- prodcal `main` = `7371666` + punch-list export commit, pushed to GitHub; tree clean.
- jdbbs-public clean at `26e1c77`.
- Service healthy; post-deploy free Ghosts build (book 9) OK: 61 pp, PDF + EPUB.
- Cache buster in `srv/static/transmittal.html` is now **20260918f**.

## Done today (all pushed)
5.4 C13 transmittal · 0.5 Format · 0.6 What-you-get · 5.8 machine factory
(docs/API.md, `srv/machine_factory_test.go`) · 6.4 · 0.7 Signature style ·
0.10 Glossary Entry (template = 13 styles) · "breaks belong to what follows"
rule · 0.8 drafts (`docs/brand/ABOUT-THE-BOOK-FACTORY-2026-09-18.md`, NOT
ticked) · 0.9 → 5.13 parked · **5.14 (new, this block):** checklist trap —
blank/"Coming later" rows for half-title/title/copyright/Contents switched
those generated pages off in the build; now default on, own group in the UI,
Included/Leave out only, CIP row hidden, test
`TestPullTransmittalGeneratedPagesDefaultOn`.
§2 pre-check as pinstitute (real one-time login link inserted then deleted;
`ip='smoke'`): portal, perception/transmittal, book-001/factory, free Inspect
all fine; template has 13 qFormat styles, Georgia/Arial/Courier New,
protocolized trim 4.91 × 7.59 in. Notes posted on 2.6, 2.7, 0.8, 5.14.

## Open catch
Plantin MT Pro **Bold** missing on VM (Regular + Italic only) — bold body
silently sets regular in Plantin print builds. In docs/IDEAS.md; Jenna to supply OTF.

## Waiting on Jenna
0.8 pick A/A′/B → then wire as transmittal checkbox (off by default; colophon
on © page and/or back-matter page) · 6.1–6.3, 6.5 keep/kill · 3.2 cuts ·
3.4 run-through · 2.5–2.7 eyeball · wording pass on Format / What you get /
Manuscript Checklist group labels / `/factory/terms` · one InDesign-set page for 5.5.

## Next for the agent (order)
1. Check §0 inbox first — new items take priority.
2. **5.5** on Typst 0.12: hyphenation/justification/keep-with-next levers,
   measured with `typesetting/scripts/compscore.py` on Ghosts (book 9, free)
   and Obliquities PDF (do NOT build project 22 — use its existing output).
   `costs` needs 0.13 → Wed 24+ (see `docs/reviews/P3-COMPOSITION-2026-09-18.md`).
3. **1.5 re-tag**: checkpoint tag per CHECKPOINTS.md + `go test ./srv/` +
   one pinstitute smoke (there have been ~10 deploys since `0cd67fd`).
4. 6.6 IDEAS tidy once 6.x decided; 0.8 wiring once picked; 3.5 after 3.4.
5. 5.9 only if queueing shows Mon/Tue.

## Rules carried
Never `cat` >350-line files; explicit `git add -- paths`; `make build` before
`sudo systemctl restart prodcal`; free test builds only on book 9 via
`localhost:8799/api/books/9/convert`; never burn project 22 credits; no
licensed-font embedding; no GitHub PAT; Typst 0.13 deferred to Sep 24+.

Metrics: ~45 % context at handoff; files read in full (guard bypasses): 0.

---

## Addendum — Fri 18 Sep, block 2 (after compaction)

Landed, all deployed and pushed (`checkpoint-2026-09-18-pre-workshop-v2` at
`9ac2027`, then `d7a1d76`):

- **5.15 images** (`41351b2`): `srv/images.go` — after pandoc extracts
  media, every colour raster (share of clearly-coloured pixels ≥ 0.5 %; a
  mean-saturation test missed a white chart with two thin coloured lines) is
  rewritten in place: luminance grey, `-auto-level`, `-sigmoidal-contrast
  3,50%`. Already-grey untouched; EPUB reads the .docx so keeps colour. Spec
  `images.print_colour` ← transmittal `illustrations.print_colour` checkbox
  (Illustrations section, `tx-attest` box, off by default). Inspect
  `image_inventory` findings carry `colour: true/false` (same recipe in
  `detect-edge-cases.py _image_is_colour`); factory page counts them.
  Verified with a 3-image test book through the real pipeline (book 18 on
  project 14, deleted).
- **5.16** `@`-in-link fix; **5.14** front matter default on; compscore fix
  (see block 1 above).
- **5.18** (`passes.go` 4b): pass fulfilment seeds the transmittal draft with
  title/author — `seededTransmittalData` in `transmittal.go`.
- **2.5** verified server-side: admin New pass → password → client verify
  200 / wrong 401 / no cookie 401. Throwaways: pass 9 + project 26
  (`tattendee`), pass 10 + project 27 (`sauthor`) — revoked + archived
  (project DELETE is disabled by design).
- **Punch list rolled over**: list 1 archived `docs/runs/PUNCHLIST-2026-09-18.md`
  (35/55); list 2 is `scratch/run/CHECKLIST.md`, exported as
  `docs/runs/PUNCHLIST-2026-09-18-list2.md` (export arg `2026-09-18-list2`).
  Item numbers kept so notes stay attached; 0.8 → 5.17 (notes migrated).
  Inbox §0 now at the **top**; `server.py /add` inserts at end of §0 section.
  Archive copies of the raw files in `scratch/run/archive/` (gitignored).
- **0.8 drafts** (About the Book Factory A/A′/B) were only in
  `docs/brand/ABOUT-THE-BOOK-FACTORY-2026-09-18.md`; Jenna couldn't see them
  → posted in full as a note on 5.17. Rule: anything Jenna must read goes on
  the item, never as a repo path.

Open on list 2 (all waiting on Jenna or post-workshop): 2.6–2.8, 3.2/3.4/3.5,
4.x (workshop days), 5.17 pick → wire, 5.5 InDesign compare, 5.12/5.13/5.19/
5.20/5.21, 6.1–6.6. Nothing for the agent to start unprompted except 3.5 after
3.4 and 6.6 after 6.x decisions.

### Addendum 2 — Fri 18 Sep, late block

- **Calendar hidden from customers** (commented out in `theme.js clientNav`,
  `client.html` card, `transmittal.js` header). Page still serves. Jenna may
  also want the "N/N tasks done" line on the card gone — asked, no answer yet.
- **Inspect costs no tokens** (0.2) — rules-based; line added to the factory
  page's Inspect blurb.
- **`[[style]]` markers** (0.3, `216dfeb`, built by subagent `style-markers`):
  `typesetting/scripts/apply-style-markers.py` pre-pass, hooked via
  `srv/stylemarkers.go` into `runDirectBuild` and `generateEPUB`; Inspect has
  `detect_style_markers` ("Marked styles") and `detect_unusual_fonts` is one
  row per font with meaning. Rules on item 0.3 and on the factory page under
  Upload. Spec kept at `scratch/SPEC-style-markers.md` (gitignored).
- **5.22 CC copyright notices** queued (from inbox 0.4): Rights choice on the
  transmittal → `series-template.typ` l.604 notice + EPUB `dc:rights`.
- Jenna is running the Obliquities rollup (`scratch/obliq-jrd.docx`,
  Google-Docs export, all-"normal" styles) this afternoon on pinstitute —
  watch `journalctl -u prodcal` for `style markers applied` and
  `images converted to grey`.

## Addendum 3 (evening)

- 0.4 transmittal front-matter indents removed (all rows flush left; renderer ignores stored `indent`).
- 0.5 "Email me a copy" removed from customer transmittal (admin keeps Email). transmittal.js buster 20260918k.
- 0.6 **Today's Obliquities build failures were pipeline bugs, fixed (7aae7ee):**
  - `apply_book_map` held the `]` closer of a `#signature[` wrapper in `pending`, so `#start-body()` landed inside the container → typst "pagebreaks are not allowed inside of containers". Closers now flush before hooks.
  - `BlockQuote` used `stringify` → unescaped `@` (email) read as label ref; also lost italics. Now `pandoc.write(..., 'typst')`.
  - Failed typst builds keep their work dir at `$TMPDIR/prodcal-failed/book-N` (log line "failed build kept").
  - Book 23 (Obliquities rollup templated) rebuilt OK, 77 pp. Failed builds refund credits.
- 0.6 Word-access resources posted as note (Word for the web / LibreOffice / Docs + `[[style]]` markers / trial). Offered a "No Word? Three ways in" box on factory + workshop pages — awaiting yes.
- Chapter-title hyphenation fixed (e088520): `show heading: set text(hyphenate: false)`.

## Addendum 4 (late evening) — stopping point + order for what's left

Landed since addendum 3:
- 0.7 **Failed-build messages** (d125b3e): `srv/builderr.go` `diagnoseBuildFailure` → explanation + `Near: “…”` (text quoted from the book.typ line typst points at) + folded `Technical detail`. factory.js `errDetailHTML`; css 20260918b, js 20260918e. Books 22/24/25 error_msg rewritten by hand.
- 0.8 part 1 **Trim-derived typography defaults** (d187fa5): `srv/typodefaults.go` `applyTypoDefaults` called at top of `specToTypstConfig`. Legacy margins quad / `leading_pt ≤ 3` / `base 10` / `indent 0.75` are treated as unset. Escape hatches: `page.margins_custom`, `typography.leading_custom`. Cap-height table per face. Test `TestApplyTypoDefaults`.
- Headings never hyphenate (e088520).
- 0.9 plan posted on the item (A: three passes here; B: bearer token + CLI recipe).

### Suggested order for the next session(s)
1. **Jenna answers** (5 min each, unblock everything below): 0.6 "No Word?" box yes/no · 0.7 email-on-customer-failure yes/no · 0.9A three names/emails/titles · 5.17 A/A′/B · 5.22 does Obliquities need CC.
2. **0.9A** three passes (15 min) — most demo value per minute for Monday.
3. **0.8 part 2** transmittal typography choices (half day; needs a build-review pass with a real book per pairing). Own session/subagent; touches transmittal.js/go, bookspecs.go sync, TYPOGRAPHY_PAIRINGS.md.
4. **0.9B** bearer token + CLI/Python recipe (2–3 h). Own session; touches auth + docs/API.md only.
5. **0.6 box** + **0.7 email** if yes (20 min each).
6. 5.22 CC notices, 3.x deck once Jenna has done 3.2/3.4, 6.6 after 6.x decisions.
7. Workshop watch §4 Mon/Tue. `journalctl -u prodcal -f | rg 'conversion failed|failed build kept|style markers'`.

### Gotchas
- Failed builds keep work dir at `/tmp/prodcal-failed/book-N` — compile by hand with `typst compile --root / --font-path typesetting/fonts …/book.typ out.pdf` (media paths inside point at the original tmp dir; sed them).
- Build credits: admin builds on passed projects debit; failures refund. Free tests only on project 14 / book 9.
- Obliquities (project 22) now builds; next build will be ~110 pp at 10.5/13.4 instead of 77 pp.

## Addendum 5 (2026-09-19, midday) — winding down for compaction

Landed:
- 0.10 **/press** (jdbbs-public/press.html, route in server.go, migration 044, status draft) — modelled on desertant.com/press; four orange `.fill` placeholders for Jenna (bio, entity, founded, city). Brand assets generated by `docs/brand/wordmark-gen.py` → `srv/static/brand/` (wordmark SVGs, social card); og:image set.
- 0.11 **Warm paper**: `--bg #FFFEF6` (Jenna's pick), `--border #E4E2D8`, dim/progress/toggle `#F3F1E8`. docs-editor.html hardcode replaced with tokens.
- 0.9B (subagent, ba069b6): Bearer auth alias in `checkAuth`, `docs/API-CLI-RECIPE-2026-09-19.md`, `scripts/factory-cli.py`. 0.9A (three passes) still needs names.
- 0.12 / 0.13 **comps** at scratch/comps/index.html, served by busybox httpd in tmux `comps` on :8767 (https://jdbbs.exe.xyz:8767/). Recommendation posted: 0.12 B (tinted) or A (boxed); 0.13 A+C (step strip + finish block + hand-off panel). Waiting on Jenna's pick; then ~1–2 h build in transmittal.js/.css (+ fx-steps markup borrowed from factory).
- 0.14 domain/email ideas posted on the item; no build.

In flight: subagent **typo-choices** (conv cDXESK6) — 0.8 part 2 transmittal typography rows (pairing/size/section break/paragraphs), vendoring EB Garamond, new `srv/typochoices.go`; edits bookspecs.go, typodefaults.go, epub.go, transmittal.js. If it hasn't reported when the next session starts: `git status`, read its report via previous-conversations skill, review its spreads in scratch/typo/, and post on 0.8.

Next: Jenna's picks (0.12, 0.13, 0.6 box, 0.7 email, 5.17, 5.22, 0.9A names) → build 0.13 A+C and 0.12 → review 0.8 part 2 → workshop watch.
