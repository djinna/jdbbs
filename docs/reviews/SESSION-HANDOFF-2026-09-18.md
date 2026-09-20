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

## Addendum 6 (2026-09-19, afternoon) — before compaction

Landed (all pushed to djinna/jdbbs main, VM rebuilt/restarted):
- **0.12** transmittal fields are warm wells: theme token `--well-bg #F6EFDD` (dark `#1C1A16`), white on focus (7a4dec1).
- **0.13 A+C** (b549648): `.fx-steps/.fx-step` moved factory.css → theme.css; transmittal.js gained `toggleFinal()`, `renderStepStrip()`, `renderHandoff()`, `renderFinish()`; intro hides once final; strip/handoff/finish hidden in print. Cache busters transmittal.css 20260919c, transmittal.js 20260919d, factory.css 20260919a.
- **0.2** factory.html Inspect line: "Python script … no LLM reads your manuscript".
- **0.8 part 2** landed by subagent typo-choices (2f205eb, 21bbdc2, 6c45b6d): `srv/typochoices.go`, EB Garamond vendored under `typesetting/fonts/ebgaramond/OTF/`, transmittal `// TYPOGRAPHY` section; docx-to-typst-enhanced.lua no longer wraps "Body Text" in `#block`; books.go uses the specialised template path. Spreads in `scratch/typo/`. Pass 12 (Ghosts) granted +11 builds via `POST /api/admin/passes/12/grant`; 6 remain for testing.
- **5.22** Rights (50187d3): `srv/rights.go` (`rightsOptions`, `rightsLine`, `rightsShort`) + test; transmittal `page_iv.rights` → spec `metadata.rights`; Typst `rights-line` field (series-template.typ copyright-page-generated); generate-word-template.py `rights_line()` mirror; epub.go `spec.Rights` → pandoc `--metadata=rights`. Keep the three wordings in step.
- **0.15** jdbbs-public/factory.html `.price .amount` clamp(16px,3vw,22px).
- **0.16** punch-list screenshots (c80ecde): runpage `POST /upload`, `GET /img/<name>`, notes `images[]`, paste handlers in page.html; export copies to `docs/runs/img/`; `/admin/runs/{kind}/{name}` (kind=img) serves them. Runpage restarted in tmux `runpage` with `RUNPAGE_CHAT_CONV=cMMJF5J` — **set the new conversation id when a fresh session starts**. AGENTS.md notes the `[screenshot: path]` convention (open with read_image).
- 0.14 domain ideas posted (no build). Comps still served on :8767 (tmux `comps`); can be killed once 0.12/0.13 are accepted.

Gotcha: Ghosts (project 14) transmittal got flipped to draft by a concurrent autosave during testing; set back to final via sqlite (`update transmittals set status='final' where project_id=14`).

Waiting on Jenna: 0.9A names/emails/titles → three passes; 0.10 four placeholders; 5.17 A/A′/B; 0.6 box, 0.7 email; §2.6–2.8; §3.2/3.4; §6.1–6.5.
Next agent work: 3.5 (push deck, link in handoff), 6.6 once §6 decided, workshop watch §4 Mon/Tue; review 0.8 spreads with Jenna (pull PNGs onto a page if she wants to compare).

## Addendum 7 (2026-09-19, ~15:10 UTC) — before compaction

Landed since addendum 6 (all pushed):
- **0.11** page is white: `--bg #FFFFFF`, `--well-bg #F5F3EC`, `--border #E6E4DC`; brand SVGs, wordmark-gen.py, press.html colour note follow.
- **0.12** follow-up: `.tx-handoff` background is a green wash (`color-mix(var(--green) 7%, var(--bg))`), not `--surface`.
- **0.13** follow-up: "Return to draft mode" (header: "Return to draft"); draft state shows only Mark Final, no Factory link.
- **0.8 part 2b** (87819e4): section break "Your own" — `typography.section_break=custom` + `section_break_text` (≤24 chars) flows to spec `elements/epub.section_break_text`, Typst `section-break-text` + `"custom"` style, EPUB `hr::after` (cssStringEscape), Word template Section Break sample, email summary. transmittal.js v20260919f, css v20260919e.
- **runpage**: `/fragment` sends `X-Page-Version` (page.html mtime); page reloads itself when it changes. Punch list exported (0.17 comps with images).
- 2.6, 2.7 ticked by Jenna.

In flight:
- **Subagent `sampler` (conv cJR7BZO)** working from `scratch/briefs/sampler-2026-09-19.md`: typography sampler PDF (3 pairings × chapter opener + 3 pp, Typst direct, no credits) + click-to-enlarge chips under the Typeface row + download link, outputs under `srv/static/samples/`, script `typesetting/scripts/build-sampler.sh`. Its untracked files so far: `srv/samplerconfig.go`, `cmd/typoconfig/`. It will commit/push and report; check `git log` / ask it for status before touching transmittal.js. Then note on 0.8 and tick.
- **0.17 — Jenna chose C**: one page. The transmittal becomes section 1 of the factory page (`/{client}/{project}/factory/`): factory header ("// FACTORY · Title · builds left"), step strip as in-page anchors, "// 1 · TRANSMITTAL" section with History/Print/Email/Word template/Return to draft actions on the section rule, the two-column form, finish block "Continue to 2 · Upload →"; steps 2–5 below. Transmittal URL redirects to the factory page (`#transmittal`). Top nav drops the separate TRANSMITTAL entry (`clientNav()` in theme.js; check `srv/nav_convergence_test.go`). Comps: `scratch/run/img/comp-017-C-*.png`. Files: factory.html/js/css, transmittal.js (1575 lines; renderForm ~l.557, renderStepStrip/renderHandoff/renderFinish after it), transmittal.css, server.go routing ~l.445–475, theme.js. Not started — do this in a fresh session; plan first (transmittal.js mounts into `#app`; factory.js is vanilla with `renderAll()`), probably load transmittal.js inside factory.html and mount it into a `#fx-transmittal` section, with factory's own step-1 stub removed.

Waiting on Jenna: 0.9A names, 0.10 placeholders, 5.17, 0.6, 0.7, 2.8, 3.2/3.4, §6.

## Addendum 8 (2026-09-19, ~16:10 UTC) — before compaction

Landed since addendum 7 (all pushed to GitHub; VM = main):
- **0.8 sampler** committed by the sampler subagent (9fc1822, fe58b62) — verified chips + lightbox in the browser; ticked.
- **0.17 (C) one-page factory** — 72c4214, deployed. Transmittal is section 1 of `/{c}/{p}/factory/`; `/transmittal/` 302s to `…/factory/#transmittal`; TRANSMITTAL out of clientNav; transmittal.html deleted; migration 045. Design as in the commit message. `go test ./srv/` green.
- **5.13 Index**: `/factory` add-ons table lists "Index · coming soon" (jdbbs-public 6b7e278). Phase 1 running in **subagent `index-addon` (conv cYBS6TS)** from `scratch/briefs/index-addon-2026-09-19.md`, in worktree `/home/exedev/prodcal-index` on branch `index-addon` — it must not touch main. When it reports: read its bullets, check `git log index-addon`, note + tick on 5.13; phase 2 (factory UI, SKU) is ours, post-workshop.
- **5.20 Typst 0.13** moved to §1, today (Jenna). 0.13.1 binary at `scratch/typst-x86_64-unknown-linux-musl/typst` compiles Ghosts identically (117 pp). 5.5 note posted: after the swap, add `par(costs:)` runt control to the series template, rebuild sampler with 0.13, compscore before/after, regenerate `srv/static/samples/`.
- runpage re-pointed at conv cF3VYRP (`RUNPAGE_CHAT_CONV`); restart tmux `runpage` with the new id when a new conversation starts.

**0.17 remaining (small):**
1. Walk the **draft** state (project 23 `pinstitute/perception` is draft): section rule shows `[DRAFT]` + Mark Final accent link; intro paragraph; finish block "Mark Final"; step strip has 1 current. Then Mark Final → handoff panel appears, strip flips 1 ✓, `tx:status` mirrors; Return to draft. Do NOT toggle Obliquities (22) — Mark Final emails the customer.
2. Walk the **customer sign-in** path (no admin header: `localhost:8000/pinstitute/book-001/factory/`) — gate shows, transmittal section stays empty until unlock (tx:unauthorized → showAuth), then mounts.
3. Print preview via the section's Print action (only section 1 should print).
4. Mobile width (≤ 800 px): `.tx-columns` collapses; check the section rule wraps cleanly.
5. Export punch list (`python3 scripts/punchlist-export.py`), note on 0.17 with a screenshot, tick, commit.
6. `srv/email_test.go:278` still uses a `/transmittal/` URL literal — harmless; update when touching.

Then: **5.20 Typst swap** (keep 0.12 as `/usr/local/bin/typst-0.12`; `make build` not needed; factory smoke as `pinstitute` after), then 5.5 costs + sampler rebuild.

Waiting on Jenna: 0.9A names, 0.10 placeholders, 5.17, 0.6, 0.7, 2.8, 3.2/3.4, §6.

Metrics: ~60 % context at handoff; files read in full (guard bypass): 1 (INDEX-ADDON-FEASIBILITY, 74 lines).

## Addendum 9 (2026-09-19, ~16:05 UTC) — before compaction

Landed since addendum 8 (all pushed; VM = main = 7f196b3):
- **0.17 checks complete, ticked** (draft state, Mark Final ↔ draft on mcheck, Print, History, redirect, customer magic-link → mount, mobile). The "tattendee didn't mount" scare was project 26 being archived.
- **5.20 done, ticked**: `/usr/local/bin/typst` = 0.13.1, 0.12 kept as `typst-0.12`. Live rebuild of mcheck book 14 OK (mcheck pass now 2/3 used). DEPLOY.md updated.
- **5.5**: `costs: (runt: 200%, hyphenation: 70%)` in series-template (0e36abc, requires Typst ≥ 0.13); sampler rebuilt + deployed; scores posted on the note. Stays `[~]` for Jenna's InDesign page.
- Jenna's jdbbs-public commit 07ecd3f (doc editor, /factory "How it runs" links) reviewed + pushed; 0.18 ticked.
- Answered the three §6 follow-ups that were missed on 18 Sep (6.1, 6.2, 6.3) and 6.6/0.9/5.13 questions.

**Running subagents (check with `subagent` tool, slug + short timeout):**
- `index-addon` (cYBS6TS) — worktree `/home/exedev/prodcal-index`, branch `index-addon`. Piece 1 committed (b67e4b0: `#index[]` markers + `index-page()` + fixture). Piece 2 (Go drafter/anchorer `srv/indexer/`, `cmd/indexdraft`, `cmd/indexanchor`) in progress; then Ghosts sample PDF + `docs/reviews/INDEX-ADDON-DESIGN-2026-09-19.md` + REPORT. Jenna wants the Ghosts index PDF today to judge; factory wiring is post-workshop. Post on 5.13 when it lands.
- `stylesheet-61` (cTYSBXT) — **6.1 B chosen by Jenna**: per-project interactive style sheet. Brief `scratch/briefs/stylesheet-6.1-2026-09-19.md`. Works on main (migration 046, `/api/projects/{id}/stylesheet…`, page `/{client}/{project}/stylesheet/`, portal + admin links, tests). Commits but does not push — review `git log`, run `go test ./srv/`, browser-check on mcheck/book-001, then push and post screenshots on 6.1. **Don't build/restart in parallel with it.**

**Decisions / promises made on the punch list today:**
- 0.9: proposed a short "Have your own factory? It's an API" section at the foot of `/factory` + a separate public page `/factory/api` rendering `docs/API-CLI-RECIPE-2026-09-19.md` (jdbbs-public file + one route), plus a 6-line terminal demo script for the workshop — **Sunday morning if Jenna says yes**.
- 6.2: **Sunday night**: archive list 3 (`scripts/punchlist-export.py`, already writes docs/runs), start Punch list 4 — Workshop at the same runpage URL, §0 Inbox only. During Mon/Tue triage each item Hotfix-now vs After-workshop.
- 6.6: mine once 6.1 lands and Jenna nods on 6.3 (park): tidy docs/IDEAS.md (move 6.1 to §5 dated, keep 6.3/6.5 parked), tick 6.6.
- 5.13: told Jenna today = Ghosts sample index PDF + design doc to judge; SKU/UI post-workshop.

Waiting on Jenna: 0.9 (yes/no on page + the three customer names for passes), 0.10, 5.17, 0.6, 0.7, 2.8, 3.2/3.4, 5.5 InDesign page, 6.3 nod.

Metrics: ~55 % context at handoff; files read in full (guard bypass): 0.

## Addendum 10 (2026-09-19, ~16:45 UTC) — before compaction #2

Landed since addendum 9 (all pushed; VM = main = 980090a + export commit):
- **6.1 B live** (subagent `stylesheet-61`, done): `/{client}/{project}/stylesheet/`, migration 046, `srv/project_stylesheet.go`, static `project-stylesheet.*`, 5 tests. Report `scratch/briefs/stylesheet-6.1-2026-09-19.REPORT.md`. mcheck/book-001 carries demo state (1 edit, 1 reject, 2 added).
- **5.13 index add-on** (subagent `index-addon`, done): branch `index-addon` pushed, 5 commits, **not merged**. Ghosts sample PDF posted on the punch list (`scratch/run/img/513-*`), design note `docs/reviews/INDEX-ADDON-DESIGN-2026-09-19.md` on the branch. Awaiting Jenna's verdict (kind of index right? SKU or hand-run?). Post-workshop either way.
- **0.19** transmittal columns rebalanced; **0.20 A** tick grid for the checklist; **0.21/0.22** admin console de-calendared (stats, filters, row links → Factory · Style sheet · Set password · Rename URL · Archive). **6.3/6.6** IDEAS.md tidied.
- Runpage: light/dark theme + toggle; serves `.pdf` from `/img/`; restarted in tmux `runpage` with `RUNPAGE_CHAT_CONV=cF3VYRP` (it had died — check `tmux ls` at session start and restart with the current conv id).

**Next, in order:**
1. **0.9 — Jenna said yes, go now**: (a) short "Have your own factory? It's an API" section at the foot of `~/jdbbs-public/factory.html` (above "Redeem a pass"), (b) public page `/factory/api` — new HTML in jdbbs-public rendering `docs/API-CLI-RECIPE-2026-09-19.md` content (+ link to `scripts/factory-cli.py`), route added next to the other `servePublicDoc` routes in `srv/server.go` + `site_pages` migration row (047) + nav convergence test, (c) a 6-line terminal demo script for the workshop. Build/restart needed only for the route. Post screenshots on 0.9.
2. 0.22 follow-ups if Jenna answers (rename "Set password" → "Password gate"; project-path line → portal URL).
3. Sunday night: archive punch list 3, start Punch list 4 — Workshop.
Waiting on Jenna: 0.9A names, 0.10, 5.17, 0.6, 0.7, 2.8, 3.2/3.4, 5.5 InDesign page, 5.13 verdict, 0.22 follow-ups.

## Addendum 11 (2026-09-19, ~17:25 UTC) — before compaction #3

Landed since addendum 10 (all pushed; VM = main):
- **0.9 done-ish** (`[~]` until Jenna reads): `/factory/api` public recipe page (jdbbs-public `factory-api.html`), `/factory/api/factory-cli.py` (symlink → `scripts/factory-cli.py`; `servePublicDocIn` now serves `.py/.sh/.txt/.md` as text/plain), "Have your own factory?" section on `/factory`, migration 047, `scripts/factory-demo.sh` (six-call terminal demo; tested inspect-only on prot/zoo id 14, token in `scratch/zoo-token.txt`, 6 credits). 0.9A still needs names.
- **5.5 ticked** (InDesign comparison struck by Jenna; sampler links posted).
- **0.23 done**: landing page — Factory Pass offer card (`.offer`) above the pipeline readout, price 23 px, both hero buttons and the "New here?" line removed, Factory Pass first in nav, lede says "book factory" not "production calendars", hero→card spacing tightened. Commits `b1adbfe` + spacing.
- **Runpage**: notes now render inline `![](/img/x.png)` as images and URLs as links (`scripts/runpage/page.html`); when posting, prefer `"images":[...]` in the POST body. Jenna wants comps *on the item*, clickable — never only in chat.

**Open / next, in order:**
1. **0.25 Word-free authoring** (Jenna's question, 17:20): is Word still the only way to kick off the factory? Friday testers had no Word; Pages renamed styles on export; Word for the web felt janky. 0.6 only listed workarounds. Do an honest survey (Docs + `[[style]]` markers as first-class path — already shipped and Inspect lists "Marked styles"; Markdown/plain-text template; a Pages-export style-name fixer in the pipeline; LibreOffice template; own browser editor), costs in fidelity/support, recommendation. Deliverable: `docs/reviews/WORD-FREE-AUTHORING-2026-09-19.md` + summary note on 0.25. Think first, patch later; nothing lands in the pipeline during the freeze without Jenna's yes.
2. **0.24 client portal off the landing page** — Jenna said YES to `/portal` (17:30), build it first thing after compaction (static page in `srv/static/`, route, `site_pages` row 048, `PUBLIC_NAV` `/#portal` → `/portal`, footer link, remove `.portal` section from landing.html; portal-form JS lives at landing.html ~569).
3. 0.22 follow-ups (rename "Set password"; project-path line) if she answers.
4. Sunday night: archive punch list 3, start Punch list 4 — Workshop.
Waiting on Jenna: 0.9 read, 0.9A names, 0.10, 5.17, 0.7, 2.8, 3.2/3.4, 5.13 verdict.

## Addendum 12 (2026-09-19, ~23:40 UTC) — before compaction #4

Landed since addendum 11 (all pushed): 0.24 `/portal` (done). 0.25 Word-free intake: decision note `docs/reviews/WORD-FREE-AUTHORING-2026-09-19.md`; `/factory#bring` + `/workshop#bring` copy (jdbbs-public, pushed); transmittal handoff copy; `?format=odt` on `GET /api/projects/{id}/word-template` (soffice headless, `docxToODT` in `srv/bookspecs.go`). 0.26 portal card = Factory → · Style sheet; factory step strip sticky + scroll-spy (`.here`). 0.27 factory header: pass line under ← Your books; section-1 head rule removed; completion bar → "84% filled" figure. Landing/portal/handoff buttons ↓.

**NEXT: 0.28 — Proof vs. Final builds (Jenna: "great idea", go; freeze is NOT on today/Sunday).**
Model: **Build proof** = free, unlimited (one in flight, ~30/day/project): clean EPUB + print PDF with a small mono footer on every page ("PROOF · {title} · built {date} UTC · not for print") and a matching line on the copyright page. **Export final** = uses 1 of 3: same build, flag off, clean PDF (+EPUB). EPUB is never watermarked. Credits → "finals"; store `builds-3` → "+3 finals".
Where things are:
- Build handler `srv/books.go` ~L268–360: `format` parsing, credit check (`passCreditsRemaining`), `debitBuildCredit` (refund in `failConversion`); `runConversion(format)` L572; `runEPUBBuild` L885. Add `kind: proof|final` (default final for API back-compat? — decide: default **proof** on the factory page, API default final to keep the recipe honest; document).
- Typst template: `typesetting/` (series template; find where inputs like trim/typeface are passed — `--input` from Go). Add `proof` input → footer + copyright line.
- Books table: add `kind` column (migration 049) so Download lists finals first; pass line "N of 3 finals left".
- Factory UI `srv/static/factory.js` ~L581–603 (`fx-build-btn` text), L1170 (`convert` POST `{format:'both'}`), step 5 list; `factory.html` step 4/5 markup.
- Copy: `~/jdbbs-public/factory.html` (What's included, Price, How a build works step 4), `factory-api.html`, `scripts/factory-cli.py` / `factory-demo.sh`, email `buildDeliveredText/HTML` in `srv/passes.go` L1375+, `srv/store.go` L53 pack description.
- Tests: `srv/passes_test.go` (debit paths ~L784, 826, 1374). Smoke on mcheck (id 17) — never pinstitute (22).
Open decisions still with Jenna: 0.9A names, 0.22 follow-ups, 5.13, 5.17, 0.7, 2.8, 3.2/3.4, 0.10. Sunday night: archive punch list 3 → Punch list 4 (Workshop).

## Addendum 13 (2026-09-20, ~00:20 UTC) — parked: 5.21 typeface choice / MyFonts

Jenna (5.21): Plantin MT Pro has more weights on MyFonts (link on the item). She has kept page-design choices simple on purpose, but would like people to be able to choose their own typefaces and buy licences themselves. **Ask: first scout only — does MyFonts.com have an API we could connect to so customers choose + purchase font licences for themselves? Discuss feasibility before building anything.** Do this after the next compaction. Scope of the scout: (1) MyFonts / Monotype public APIs (purchase, catalogue, affiliate/referral), (2) licence types vs. our server-side Typst rendering (desktop vs. app/server licence — the build runs on the VM, not on the customer's machine), (3) the realistic shapes: referral-out + upload-back with a licence attestation; a curated menu of families we hold server licences for (per-book surcharge); Google-Fonts/OFL menu now as the free tier; (4) how `typesetting/fonts/licensed/README.md` and the spec's `body-font`/`heading-font` already handle a custom family. Post findings on 5.21 as a short options memo, then wait for her call.

## Addendum 14 (2026-09-20, ~00:30 UTC) — compaction #4 state

**0.28 Proof vs. Final — server + factory UI LANDED and live** (commits "Proof vs. Final builds (0.28), server half" and "Factory page: Build proof…"; both pushed). Smoke: mcheck book 14 proof → outputs 84/85 kind=proof, footer on every page, filename `-PROOF.pdf`, no ledger row.
Remaining for 0.28 (do first next session):
1. Copy: `~/jdbbs-public/factory.html` (What's included / Price / How a build works step 4 / add-ons "+3 finals"), `factory-api.html` + `scripts/factory-cli.py`/`factory-demo.sh` (`kind: proof|final`, default final, epub-only = free proof), `srv/store.go` builds-3 → "+3 finals" name/description, `srv/EMAIL_SYSTEM.md` (pathway #7: finals only), `docs/IDEAS.md`/DEPLOY if relevant. Then `git -C ~/jdbbs-public push origin main`.
2. Final export smoke on mcheck (book 14) — confirm clean PDF, ledger debit, email text.
3. Tick 0.28 on the punch list; export; commit docs/runs.

**5.13 Index add-on — subagent `index-phase2` (conv cNVICG3) running in worktree `/home/exedev/prodcal-index`, branch `index-addon`**, brief `scratch/briefs/index-phase2-2026-09-20.md`. It will write `scratch/briefs/index-phase2-2026-09-20.REPORT.md` and edit `~/jdbbs-public/factory.html` (NOT pushed) when done. Lead's job after: review report, rebase/merge branch to main, `make build && go test ./srv/ && sudo systemctl restart prodcal`, smoke an index draft + build on mcheck, push jdbbs-public, post v2 Ghosts index images (`scratch/run/img/513-v2-*.png`) on 5.13. Jenna's asks: no letter heads, $100 SKU, free-in-test-Stripe for the workshop.

**5.21 MyFonts scout** parked (addendum 13). Other open: 0.9A, 0.22, 5.17, 0.7, 2.8, 3.2/3.4, 0.10; Sunday night archive punch list 3 → 4.

## Addendum 15 (2026-09-20, ~00:15 UTC) — compaction #5 state

**Done this stretch:** 0.28 fully landed + ticked (final smoke OK; copy on /factory, /factory/api, CLI, Stripe product "+3 finals" via new `syncProductCopy`, EMAIL_SYSTEM.md). **Punch list cycled → list 4 (Workshop)**: list 3 archived `docs/runs/PUNCHLIST-2026-09-19-list3.md` (+ raw in `scratch/run/archive/`), rows shortened with links to the archive/GitHub per Jenna; new items number from **0.30 / 5.23**. Runpage numbering collided once (auto gave 0.26) — renumber inbox items by hand if it happens again.

**Next, in order:**
1. **4.5 — merge `index-addon` (5.13).** Subagent `index-phase2` finished: report `scratch/briefs/index-phase2-2026-09-20.REPORT.md` (read it first; commits c0952b2…1281b77 on top of phase 1; rebased on main 8cc6fc4 — main has moved 3 small commits since, so `git -C /home/exedev/prodcal-index rebase main` then `git merge --ff-only index-addon` in main). Then `go test ./srv/ ./srv/indexer/`, `make build && sudo systemctl restart prodcal`, smoke on mcheck (grant index via `/admin/store/` "+ index", draft, review, build with index), `git -C ~/jdbbs-public diff` → review → push, post `scratch/run/img/513-v2-101…105.png` on 5.13, tick 5.13/4.5.
2. **0.29 head spacing** — TYPOGRAPHY.md now has the rule ("Space Around Subheads", added 20 Sep); tune the template TO that rule, Jenna to approve the rule first. (Jenna's screenshot: space above/below H2 reads equal). Diagnosis: `sub-head`/`sub-sub-head` in `series-template.typ` (~L944–966) add `#box(height: 1em)` + empty `par[]` under the head, so below ≈ 1.5× above despite config `h2-above 1em / h2-below 0.25em`. First attempt (plain `block(sticky: true, above/below)`) was **too tight both sides — Jenna rejected, reverted**. Proper fix: sticky block, then re-tune above ≈ 1.5–2 lines / below ≈ 0.5 line in body-leading multiples (TYPOGRAPHY.md L134: heads occupy whole multiples of leading), also check head size vs spec and that the first paragraph after a head is unindented; test file `scratch/proof/h.typ`, compile `typst compile --root . --font-path typesetting/fonts scratch/proof/h.typ out.pdf`, post before/after crops on 0.29 BEFORE deploying. After-workshop unless Jenna says hotfix.
3. **4.6** pre-workshop pass + checkpoint tag `checkpoint-2026-09-20-pre-workshop-v3`.
4. 5.22 MyFonts scout (options memo), 0.25 Word-free note — after the above.

## Addendum 16 (2026-09-20, ~00:40 UTC) — Sunday block, conv cRARHD4

Done + pushed: **4.5/5.13** index add-on merged (ff), migration 050 live, smoke on mcheck book 14 (draft → review → proof with index), jdbbs-public copy pushed. **4.6** attendee pass on mcheck (book 39: upload → Inspect → proof → final, ledger + email + Floor), tag `checkpoint-2026-09-20-pre-workshop-v3`. mcheck client password was reset for the test (`scratch/idx-smoke/pw`); one real sign-in-link email went to bookiq@gmail.com at 00:17 UTC.

**0.29 — waiting on Jenna: A or B.** Rule approved ("lgtm, update the Sampler too"). Template change parked at `docs/typesetting/patches/029-subhead-spacing-A-2026-09-20.patch` (variant A: `h2-above 18pt / h2-below 10.5pt / h3-above 14pt / h3-below 7.5pt`, sticky blocks, no hidden box/empty par). Variant B = same patch with `h2-above 24pt / h2-below 16.5pt`. **The template is read from disk at build time — never leave an untested edit in the working tree.** Measuring tool: `python3 scratch/proof/baselines.py FILE.pdf [page]` (ink-band baselines at 288 dpi, no PIL needed); body grid measures 11.625 pt. Fixed parts: H2 above +10.75, below +6.75; H3 above +6.5, below +7.0 (add to the block gap to get baseline Δ). After her pick: `git apply`, adjust, `go test ./srv/`, rebuild `typesetting/scripts/build-sampler.sh` (needs `scratch/typo/ghosts2.typ`), one Ghosts proof, then tick. Gap: consecutive A→B head gets full B above (1¾ L), rule says half — needs context; after workshop.

**Browser gotcha:** the factory's Export-final `window.confirm` wedges the headless tab (CDP says "no dialog"); fix = `kill` the renderer pid, then navigate again, and stub `window.confirm=()=>true` before clicking. Admin proxy for the browser is `http://127.0.0.1:8799` (tmux `adminproxy`); `:8000` direct = attendee view.

Next: 5.22 MyFonts scout memo → 0.25 Word-free note.
