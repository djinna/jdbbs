# Protocol Institute — first client through the whole factory · **Test 2**

**2026-09-17.** Test 1 (14:46–15:10 UTC) started, then aborted after signup — artefacts kept below. Sandbox Stripe (`stripe-test`). Client `production@protocol-institute.org`.
Legend: **YOU** = your turn · **ME** = Shelley's turn · **BOTH** = look together.
Ticked = done. This page refreshes itself every 20 s.

## 0 · Instrumentation — ME

- [x] 0.1 Request log (non-GET + every 4xx/5xx, with who) + slog for login / upload / Inspect / transmittal
- [x] 0.2 `scripts/factory-tail.sh`; running in tmux `factory-tail`
- [x] 0.3 Build, tests green, restart, commit, push
- [x] 0.4 This page live

## 1 · Signup from the sales page — YOU

- [x] 1.1 Open `https://jdbbs.exe.xyz/factory` → **Buy a Factory Pass**. Use a **private window** so nothing from the admin session leaks in
- [x] 1.2 Stripe Checkout: email `production@protocol-institute.org`, **cardholder name is what the sign-in slug comes from** (C1) — `jdixon` is taken by test 1, so use e.g. "Protocol Institute" → `/protocol-institute/`; card `4242 4242 4242 4242`, any future date, any CVC. **No promo code** this time (full-price path)
- [x] 1.3 Land on the success page — note what it tells you to do next
- [x] 1.4 ME — watch: `checkout session created` → `store: fulfilled` → `factory pass fulfilled` → `Your Factory Pass` email; check `/admin/store/` row, pass + client + project created
- [x] 1.5 YOU — open the inbox for `production@…`: pass email arrived? Does it read right? (audit copy BCC'd to j@)

## 2 · Portal + transmittal — YOU

- [x] 2.1 Follow the link in the pass email, log in with the password it gives
- [x] 2.2 ME — `client login` logged; no 4xx on the way in
- [x] 2.3 YOU — Transmittal: title, author, trim, formats, any custom styles. **Save draft** first
- [x] 2.4 YOU — **Mark Final**
- [x] 2.5 ME — `transmittal saved status=final`, template built, `template_ready` email sent
- [x] 2.6 YOU — download the Word template (portal or email link); open it; styles list matches what you declared?

## 3 · Manuscript + Inspect — YOU

- [x] 3.1 Upload the manuscript **as it is today** (before applying the template) → **Inspect**. This is the real-world Inspect data point
- [x] 3.2 ME — `manuscript uploaded` (size) → `inspect run` counts; compare with what the report page shows you
- [~] 3.3 BOTH — read the findings: false positives? unclear labels? anything you'd want it to have caught? — **skipped today** (call at 16:30; manuscript used as-is, throwaway)
- [ ] 3.4 YOU — apply the template (Organizer or paste into template), re-upload, re-Inspect
- [ ] 3.5 ME — counts dropped as expected; Keep / Strip / Convert decisions recorded

## 4 · Build — YOU

- [x] 4.1 **Build**
- [x] 4.2 ME — `book conversion starting` → `complete` (elapsed, sizes), credit 3→2 in the ledger, `Build ready` email
- [x] 4.3 YOU — download PDF + EPUB. Check: trim, title page, chapter openers, section breaks, running heads, EPUB nav
- [ ] 4.4 BOTH — catches → list below

## 5 · Second title (optional today) — BOTH

- [x] 5.1 New project under the same client from the portal
- [x] 5.2 ME — attach a pass to it (API; admin button is on the pre-freeze list)
- [x] 5.3 Repeat 2–4 (second EPUB build 12:30 local, free — confirmed by user)

## 6 · Wrap — ME

- [x] 6.1 Fix small catches inline; park big ones
- [x] 6.2 Handoff addendum; commit, push (`docs/reviews/SESSION-HANDOFF-2026-09-13.md`)
- [ ] 6.3 BOTH — freeze / hotfix policy for Sep 19–23

## Test 1 — aborted (artefacts kept)

Client `jdixon`, project 21 `jdixon/book-001` "Obliquities", **pass 6 revoked** (bundle $427: pass + 3 builds + 6 mo storage), store order 3, session `cs_test_…E46W`, pass email received 14:48 UTC (old wording).

## Catches

- **C1 (design question)** Client slug comes from the Stripe cardholder name → `jdixon`, portal at `/jdixon/`. An organisation buying for an author gets a URL named after whoever paid. Options: ask for "press / author name" on the sales form, or let the client rename in the portal.
- **C1b** Slugger is *first initial + surname*: "Protocol Institute" → `pinstitute`. Fine for people, odd for presses/orgs. Same fix as C1.
- **C1c** "Hi Protocol," — `firstName()` on an org name. Part of the C1 fix.
- **C4 (yours, 1 min)** Email footer saved value reads "Replied to this email go to Jenna." — fix in `/admin/docs/` → Email strings (or clear to fall back to the default "Reply to this email to reach Jenna.").
- **Obs.** Pass email "Support" (3 bullets) and the sales page "Support edges" (Preflight Review $75 / Live help / Studio typesetting $800) don't quite match. Decide which is canonical at wrap.
- [x] **C23 (UX, small) — header link done `b676a61`; build actions always buttons `102079a`; report crumb → C18** No way back: factory page wordmark links to `/` (public home), not the client's own home (`/pinstitute/`); and the Preflight report page has no link back to `/pinstitute/book-001/factory/`. Add a "← Your books" / "← Back to the factory" crumb (report part folds into C18).
- [x] **C22 — DONE `b676a61`** (EPUB free/unlimited, PDF counted; verified live on book 17: EPUB build 1.3 s, pass 7 still 2/3, "EPUB ready" email). Split the build. Today one build = PDF + EPUB, 3 per pass. Proposal: **EPUB builds unlimited** (like Inspect — 0.8 s, no typst, and it's the cleanup loop for the PDF), **PDF builds counted** (3 per pass, +3 add-on). Two buttons: "Build EPUB — free, as often as you like" / "Build print PDF — uses 1 of 3". Touches: `POST /api/books/{id}/convert` gets `format` (epub|pdf|both), debit only on pdf; factory page step 4 copy + buttons; sales page "3 builds" → "3 print builds, unlimited EPUB"; Build-ready email per format; keep last ~10 EPUB outputs per book. Alternative (two separate pools 3+3) rejected as more schema/UI for less clarity.
- **C21 (small)** PDF `Author` metadata = "Protocol Institute" (books.author, i.e. Stripe cardholder) while EPUB `dc:creator` = "Venkatesh Rao" (transmittal). Build should take title/author from the book spec for the PDF too. Same root as C1.
- **C20 (done `81eebd3`, threshold fixed)** Near-black greys (luminance < 90/255, spread < 24) → low with "the build sets it black; nothing to do". Applies to the next Inspect run.
- **C20-orig** Colored Text: 51 findings all **high** for RGB(68,68,68) — near-black grey from a Google Docs export. Treat dark greys (luminance below ~25 %) as low / "auto-normalised" (the build sets black anyway); reserve high for actual colours that would vanish in print.
- **C19 (done `81eebd3`)** Summary now excludes `auto_decision=preserve` from high/medium/low and reports `preserved`; factory shows "Carried through automatically: 446". Summary recomputed from stored findings so old runs show the new buckets (project 22 now 61 / 2 / 642 + 446).
- **C19-orig** Factory page and report disagree on the same run: factory shows 61 / **448** / 642 (worth fixing / worth a look / just noting); report shows 61 high / **2 medium** / 642 low + **446 auto-preserved**. The report is right — manual bold/italic is carried through by the build. Make the factory summary use the same buckets (add "carried through automatically: 446"), otherwise 448 "worth a look" scares people off a clean-enough file.
- **C18 (done `8093d1d` + `07c4d28`)** Report re-skinned to theme.css (masthead, `// INSPECT`, factory vocabulary, printable); "← Back to the factory" crumb injected at serve time (`{{FACTORY_URL}}`); 401 is now an HTML "Sign in to your factory first" page linking to the factory. Project 22 report regenerated.
- **C18-orig** `/api/projects/{id}/preflight/report` still wears the old "Production · Typesetting" look (serif, card grid, pills) — not the theme.css / masthead / mono design language of the factory page. Re-skin only: same data, `/static/theme.css` chrome, wordmark, mono labels; keep it printable. Also: opening the report URL in a fresh tab gave one **401** (16:05:27) before it loaded — probably a second browser profile without the client cookie; confirm, and make the 401 page say "sign in to your factory first" with a link rather than a bare 401.
- **C17 (done `4ff3b0b`)** Stand-ins Georgia / Arial / Courier New, black; page = trim + mirrored margins; guide names the real typefaces. Theme-font attrs were the Calibri culprit. Visual check via LibreOffice (`scripts/docx-preview.sh`, installed on VM with MS core fonts) looks right; **Jenna to confirm in Word**. Typeface picker (c) still parked.
- **C17-orig** Word template fonts + page. Template already asks for Libertinus Serif / Source Sans 3 (`generate-word-template.py:121`) but clients don't have them installed → Word substitutes Calibri, and the python-docx default theme paints headings blue. Fix: (a) use honest stand-ins everyone has (Georgia body, Arial headings, black), and say in the guide "these are stand-ins; the book is set in <typeface>"; (b) set the section page size to the trim + book margins so line length feels right; (c) later: typeface picker on the transmittal (Libertinus / Plantin / Source Sans / Proxima) → template + build both honour it. Don't embed fonts (Plantin/Proxima are licensed).
- **C16 (done `4ff3b0b`)** 160 built-ins hidden, 11 factory styles qFormat in order, latent styles hidden, pane filter = Recommended (schema-ordered settings.xml). **Jenna to confirm the Styles pane in Word.**
- **C16-orig** Word Styles pane shows ~40 python-docx built-ins (Body Text 2, List Continue 3, Macro Text…). Fix in `generate-word-template.py`: mark every style not in the factory set `semiHidden` + `unhideWhenUsed`, drop `qFormat`; set `qFormat` + `uiPriority` on ours so **Recommended** (Word's default view) shows only the factory styles in a sensible order; strip `latentStyles` exceptions so latent built-ins stay hidden; also write `stylePaneFormatFilter` = "In current document" into settings.xml. Add one line to the Template Guide: "Styles pane → Options → 'In current document' if you see more than these."
- **C15 (UX, post-run)** Client-side "Email" button on the transmittal sends the transmittal to the **studio inbox** (jdbb@agentmail.to, cc j@djinna.com) — the client gets nothing, so from their side the button does nothing visible; JD clicked it thinking it was how you submit. For factory clients either drop it or make it "Email me a copy" (to the pass email). Mark Final is the real submit and already sends the template email.
- **C14 (fixed `bf2004b`)** Book title saved with trailing whitespace ("Obliquities 1 ") → shows in email subjects `Transmittal [FINAL]: Obliquities 1 `. Trim title/author/subtitle on transmittal save (server side, `srv/transmittal.go`).
- **C13 (parked → post-run; the big one)** Rest of transmittal for the factory:
  - **Book Design → keep, trim first.** Trim radio shows 3 presets + Other while `trimRegistry` (`bookspecs.go:1027`) knows ~15 (US digest/trade, UK A/B, A5, JIS…) and the Protocolized 124.8×192.8. Make trim a proper picker with all presets + a live page-shape preview; drop Est. pp / PPI / Spine width (printer's numbers, and pp comes from the build), Text Complexity, Outside Designer, Trim Guidance (free text nobody acts on). **Reuse previous book** → a real dropdown of this client's earlier factory projects (copies the book spec: trim, typeface, custom styles, front matter); today it's a free-text field the pipeline never reads. Keep Design Notes as the one free-text box.
  - **Page Proofs → delete.** No reviewer routing; the client downloads the PDF and reads it.
  - **Deliverables → replace checkboxes with a fixed statement.** Every pass yields: print-interior PDF + EPUB (always both), the Word template, and the Inspect report. **Typst source stays with the factory** (it's the house template + our filters; it isn't useful without the toolchain and giving it away invites "why doesn't it compile"). Fonts never (licensed). Cover files never (theirs). Printer delivery radio (PDF/X) → delete; say plainly it's an RGB PDF and the printer converts.
  - **Subrights → delete** (copub/marketing-pages — press-only).
  - **Design guidance we should give instead**: one paragraph per trim on what it suits; typeface choice (the real design lever the factory offers); front-matter menu (half-title, title, copyright, dedication, epigraph, contents) as checkboxes that drive the template.
- **C12 (parked → post-run)** Transmittal Pub Info & © → keep and **make it the copyright-page builder**. Today `generate-word-template.py:378` already pre-writes a 3-line Copyright paragraph from year/holder/publisher; credit fields are unused. Restructure fields to what a good page needs: © year, holder, publisher (+ city), edition/printing line, ISBNs (from Book), cover design credit, interior credit (auto: "Typeset by jdbb studio in <typeface>"), acknowledgements for reprinted material/photos (the old Credit/Other/Photo fields), optional LoC/CIP line. Show a **live preview** of the assembled page in the transmittal; write the same block into the Word template so they edit it in place.
- **C11 (parked → post-run; terms before go-live)** Transmittal Permissions & Consents: drop the status/date tracking fields (press-era). Replace with one courtesy line + a single attestation checkbox ("Everything in this manuscript is mine or I have permission to reprint it. The factory typesets what I send; clearing rights is my responsibility."). The real home for the liability statement is a **Terms page** (`/factory/terms`) linked from the sales page and — after Sep 23 — required at Stripe Checkout via `consent_collection.terms_of_service` (needs `terms_of_service_url` set in the Stripe dashboard). No terms page exists today. Not legal advice — have a lawyer read the terms text.
- **C10 (parked → post-run)** Transmittal Editing section: drop Developmental Edit (level + instructions) and Level of Copyediting (+ instructions) — the factory doesn't sell editing; the only human add-on is Preflight Review $75. **Keep** Special Characters, Mathematical Formulas, Custom Styles — these drive the template and typography. Rename the section "Typography notes" (or fold into Book Design). Possible later add-on: "process against your stylesheet" — today what exists is the public house stylesheet (`/stylesheet/`), the per-project corrections list (find→replace applied at build), and Inspect (Word-style conformance); an editorial-style checker would be new work, not a toggle.
- **C24 (parked → month one)** `.mobi` output. Don't build it: KDP stopped taking .mobi uploads (2022), Send-to-Kindle dropped it (2023); Kindle takes EPUB. If ever asked, calibre `ebook-convert` (apt, ~1 GB) does EPUB→AZW3 in one shell-out. Do add a sales-page line: "EPUB is what Kindle wants."
- **C25 (pricing, done `7371bcb` + jdbbs-public)** Team feedback: pass **$549**, every build counts (EPUB + print PDF together, 3 per pass), +3 builds **$99**; free-EPUB half of C22 reverted (API still accepts `format`, but every format costs a credit and refunds on failure). Promos in Stripe (test mode, recreated on live): **WORKSHOP49** $500 off, expires Tue 22 Sep 23:59 HKT; **PROTOCOL50** 50% off, first 3 redemptions, through 31 Dec HKT. PYB149 left in Stripe test mode, no longer created by code.
- **C26 (load, done)** Load test `docs/reviews/LOADTEST-2026-09-17.md`: one build ≈ 9 s / 360 MiB peak (139k-word DOCX); N=8 parallel = 50 s each, 2.9 GiB, no OOM; CPU-bound. **Global build semaphore of 2** added (`Server.buildSem`, `maxConcurrentBuilds`); rest queue as "converting". 2 vCPU / 7 GB is enough for the workshop with the gate; **4 vCPU** is the upgrade if queue time matters, RAM bump not needed, disk fine (~1.8 GiB for 90 output pairs).
- **C27 (open → month one)** 20 MB EPUB in the load test was **not a regression** of the fallback-font fix (`78fd9ee` still embeds Noto CJK only when text nodes contain CJK): the Twitter Years manuscript has one ASCII-art bunny tweet with fullwidth `＿￣`, `ㅅ`, `づ` and an ideographic space → rule fires, 16 MB of Noto Serif TC for one joke. Refinement: raise the trigger to a real CJK *run* (say ≥ 20 consecutive ideographs) and/or subset the font to used glyphs (`pyftsubset`). Latin-only EPUBs are ~100 KB as intended.
- **P3 (parked → after basic building is finished; typography)** H&J quality in the print PDF. The Word template render alone showed loose lines and rivers Jenna would hand-tune in InDesign. Question: how close can we get to InDesign-grade composition *without* page-by-page tuning? Levers in typst: `#set par(justify: true, linebreaks: "optimized")` (Knuth–Plass style, already?), `#set text(hyphenate: true, lang: "en")` + hyphenation costs, `par.spacing`/`leading`, `text(tracking:)` micro-adjust, `costs: (hyphenation:, runt:, widow:, orphan:)` (typst ≥0.12 `par(costs:)`), min-hyphen chars, discretionary tracking per paragraph. Then a *measurement* pass: script that scores loose lines (word-space stretch), rivers, stacks, runts, widows/orphans on the PDF (pdftotext -bbox / typst query) and reports a "composition score" per chapter so we tune globally against numbers, not eyes. Also font-level: Libertinus has OpenType `kern`, `liga`; consider optical margin alignment (hanging punctuation, `par(hanging-indent)` no — typst lacks optical margins; could fake via `text(overhang: true)` for `-`/`,`/`.`). Deliverable: a "composition settings" section in the transmittal spec + the scorer.
- **P4 (parked → after current work; front matter)** How the factory ingests front matter. Today the Word template treats the book title as H1 and the author as First Paragraph — wrong model. Need Word styles (and typst/EPUB handling) for: **i** half-title (title only, form of the title); **ii** frontispiece (or blank); **iii** title page — title, author, publisher name + logo; **iv** copyright page, clear page break before it; TOC placeholder (or tell them not to include one — the build generates it); front-matter section titles (Foreword, Preface, Acknowledgments) — decide H1 vs a dedicated "Front Matter Title" style so they get roman folios and don't count as chapters; the roman → arabic switch at the first chapter or Introduction; and **arabic page 1 must land on a recto** (insert blank verso if needed). Also decide what's typed in Word vs generated from the transmittal (title page + copyright page could both be generated from spec; then the author supplies only Dedication/Epigraph/Foreword/Preface text). Deliverable: template section "Front matter" with the styles + guide text, lua filter + series-template support, EPUB nav landmarks.
  **Decisions 2026-09-17 (with Jenna):**
  - **Google Docs is the constraint** — no custom styles; export keeps only Title, Subtitle, Heading 1–6, Normal. So: *don't* ask people to retag in Word. One rule for everyone: **every section head is Heading 1** (Foreword, Preface, Chapter 1, Appendix, Acknowledgments alike).
  - **We classify by heading text**, closed vocabulary + position: an H1 is front matter if it matches ~15 terms (Foreword, Preface, Acknowledgments, Prologue, Note on the Text, List of Figures, …; case-insensitive, ≤4 words) *and* comes before the first non-matching H1; back matter mirrors it (Appendix, Notes, Bibliography, Glossary, About the Author, Colophon) after the last chapter. **Introduction = body** (page 1, arabic) — Jenna: almost always. Never silent: Inspect reports the book map ("Front matter: Foreword, Preface · Body starts at 'Chapter 1' · Back matter: Notes, About the Author"); override = rename the heading.
  - **Untitled front matter** (dedication, epigraph have no head, so no H1 to catch): everything before the first H1 — after dropping Title/Subtitle-styled paragraphs — is untitled front matter, split on page breaks, named in the transmittal's order (dedication, then epigraph). Inspect: "2 untitled front-matter pages: Dedication, Epigraph." Guidance: page break between them.
  - **Cross-check the transmittal**: it already lists the front-matter pieces; Inspect reports matches and "listed but not found".
  - **Title page + copyright page are generated from the transmittal, not typed** (title, subtitle, author, publisher, logo, ISBN, cover credit — `cover.credit` is captured but not yet placed on p. iv). Guidance: "Don't type a title page or copyright page; start your file with the dedication or foreword." Title/Subtitle-styled paragraphs that do appear are recognised and dropped (reported).
  - **Parts = transmittal opt-in** "This book has parts": H1 = part title, H2 = chapter, H3 = section; front/back-matter heads stay H1 (correct hierarchically — a Foreword sits at part level). Docs export keeps the levels, nothing to retag. Pipeline: pandoc `--top-level-division=part` when on.
  - Folios: roman through front matter, arabic 1 at the first body H1, **on a recto** (blank verso inserted if needed). Half-title (i), frontispiece/blank (ii), title (iii), copyright (iv) generated; dedication/epigraph follow on v+.
- **P1 (parked → after current work)** Review https://docs.typesafe.ai/introduction — anything useful for the factory?
- **P2 (parked → after current work)** New talk deck from Venkat's post https://protocolized.summerofprotocols.com/p/have-your-factory-call-my-factory ("Have your factory call my factory" — the origin of the factory idea). Use https://jdbbs.exe.xyz/2026-pi-symposium/talk as the *template only* (leave existing decks alone). **First step: pick ≤10 beats and show them to Jenna before doing anything else.**
- **C9 (done `3022078`)** Cover upload on the factory page under step 2 (add/replace/remove, thumbnail, EPUB-only note); transmittal Cover section → what-you-get / what-you-bring + credit + notes. Was: Transmittal Cover section is the press-era version (JDBB/Publisher front-spine-back checkboxes, paper/cloth, colours, credit, budget). Factory makes the interior only; the client does their own cover and printer submission. Replace with: (a) a short "what you get / what you bring" note — interior print PDF (trim, bleed-free, RGB — printer does CMYK) + EPUB; cover is theirs; (b) **one upload: front-cover comp for the EPUB** (JPEG/PNG, portrait, ≥1600 px tall). Plumbing exists — `POST /api/projects/{id}/book-spec/cover` (`srv/bookspecs.go:1073`, jpeg/png, 10 MB) and `epub.go:133` already embeds it as `--epub-cover-image` — but **no client UI calls it**. Put the upload on the factory page next to the manuscript upload, and show a thumbnail once set.
- **C8 (done)** Verified: inline PNG (1200 px, 3.4 in) in a DOCX → `--extract-media` → typst `#box(image(...))` → PDF at 353 dpi, and → EPUB `media/file0.png` (scratch/c8). Transmittal Illustrations table replaced with guidance (inline, one per paragraph, caption next paragraph, ≥ 1100 px for full width, PNG/JPEG, RGB) + "Art notes" textarea; old `illustrations.*_no` keys still load, just not shown. Small follow-up for month one: keep image + caption paragraph together (`block(breakable: false)` in the lua filter) — my test had the caption fall to the next page.
- **C8-orig** Transmittal Illustrations table (type / no. / here / to come) is hard for clients. Inspect already emits `image_inventory` (one finding per inline image, placed w×h in inches) — replace the table with guidance text + a read-only "Inspect found N images" line after upload. Guidance to write: embed images inline in the Word doc where they belong, one per paragraph, caption in the next paragraph; no floating/text-wrapped pictures; ≥ ~990 px wide for a full-measure figure; no separate upload path yet. **Must verify**: an inline image in a DOCX actually survives `pandoc --extract-media` → typst → PDF and → EPUB at a sane size. Not yet tested.
- **C7 (parked → post-run)** Transmittal Parts / Chapters / Words / MS pp / Est. book pp: (a) Chapters and Words/Chars fields lack the input underline (CSS); (b) we shouldn't ask at all — chapter detection runs at upload, Inspect can count words and images, Est. book pp comes from the build. Make these read-only "from your manuscript" values once a file is uploaded.
- **C6 (parked → post-run)** Transmittal Production section is press-era (Mechs Delivery, Weeks in Prod., Bound Book Date, duplicate Transmittal Date). Plan: drop the section; move Print Run into Book; one optional "Target date" field; keep old JSON keys so saved transmittals still load.
- **C5 (fixed `6486a72`)** Factory link in the pass email not obviously clickable (code style, no underline) → underlined.
- **C2 (fixed `61e1499`)** Pass email quoted the base pass — "3 builds … (6 months)" — on a 6-build, 12-month pass. Now computes from the pass and lists add-ons explicitly. Stray "the" before Discord removed. *Next fulfilment gets the new text; yours is the old one.*
- **C3 (open)** Sales page never mentions the add-ons; they appear only as Stripe `optional_items` on the checkout page. Jenna added storage; Stripe shows **both** add-ons in the session — check the sandbox Payments view for how +3 builds got in.
- **Obs.** Bundle at checkout ($349 + $49 + $29 = $427) fulfilled correctly: pass 6, 6 builds, 12 mo storage, no promo.

## Log excerpts

### 2.4–2.5 Mark Final → template (15:47 UTC)
```
15:47:36 email sent to=[jdbb@agentmail.to] cc=[j@djinna.com] subject="Transmittal [DRAFT]: Obliquities 1 "   (client POST /transmittal/email)
15:47:39 transmittal saved project_id=22 status=final was=draft who=client:pinstitute
15:47:40 email sent to=[production@protocol-institute.org] subject="Your Word template is ready: Obliquities 1"   ← 1 s after final
15:47:43 email sent to=[jdbb@agentmail.to] cc=[j@djinna.com] subject="Transmittal [FINAL]: Obliquities 1 "   (client POST /transmittal/email)
```

### 3.1–3.2 Upload → Inspect (16:03 UTC)
```
16:03:21 manuscript uploaded book_id=17 project_id=22 title="Obliquities 1" file="Obliquities rollup 1 - CANONICAL 2026-08-24 tss.docx" bytes=667892 who=client:pinstitute
16:03:27 inspect run book_id=17 status=ready findings=1151 high=61 medium=448 low=642 by_type="colored_text:51 direct_spacing:642 heading_lookalike:1 manual_formatting:446 mixed_formatting:1 unusual_font:10"   (POST /preflight 1414 ms)
16:05:27 WARN GET /api/projects/22/preflight/report status=401   ← then loaded OK; see C18
```
Real-world data point: 652 KB Google-Docs-exported manuscript, 6 s to inspect, 1 real structural finding (1 heading lookalike: "PROTOCOL INSTITUTE" bold Normal), everything else is direct formatting the build normalises.


### 4.x Build (16:09 UTC, admin-triggered via `POST /api/books/17/convert`)
```
16:09:57 book conversion starting id=17                         pass_ledger: -1 build (3→2)
16:09:59 epub generation starting → 16:10:00 complete epub_size=102777 elapsed=836ms
16:10:01 email sent to=[production@protocol-institute.org] subject="Build ready: Obliquities 1"
16:10:01 book conversion complete pdf_size=403340 elapsed=3.31s
```
**EPUB**: epubcheck 3.2 → 0 errors / 0 warnings. 24 chapter files + title page, nav 35 entries, NCX 33. No cover (none uploaded — C9). dc:title/creator had trailing spaces → fixed `b007f28`. Nav includes the manuscript's own planning headings ("[YES | half-title aka p i]") and deck paragraphs as H2 — that's the as-is manuscript, not a factory fault; the template pass (3.3) is what fixes it.
**PDF**: 95 pp at 353.8×546.6 pt (Protocolized trim), metadata Author wrong (C21). Not reviewed further today by agreement.

Also: one `Transmittal Updated` admin notification at 15:26:42 on the first draft save (debounced — 12 saves, 1 email). No 4xx/5xx.

**Test 2**
```
15:17:59 store: checkout session created kind=pass          POST /api/public/store/checkout 200 361ms
15:18:59 factory pass fulfilled pass_id=7 project=pinstitute/book-001
15:18:59 store: fulfilled amount=34900 promo=""
15:19:00 email sent → production@protocol-institute.org "Your Factory Pass: Obliquities"
```
**Test 1**

```
14:46:08 store: checkout session created kind=pass          POST /api/public/store/checkout 200 448ms
14:48:03 factory pass fulfilled pass_id=6 project=jdixon/book-001
14:48:03 store: fulfilled amount=42700 promo=""
14:48:03 email sent → production@protocol-institute.org "Your Factory Pass: Obliquities"
```
