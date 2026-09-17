# Session handoff — 2026-09-13 (workshop materials block)

Workshop: **Protocolize Your Book**, Protocol Symposium 2026, Sept 21–22
(four sessions; talk slot 25+5 somewhere Sept 21–25 — confirm day/time).
Deploy freeze Sep 19–22. Today is Sep 13.

## State at handoff

Shipped this session (all on GitHub + deployed): outbound-email log + Mail
tab; email restyle + `/admin/email-preview/`; BCC audit copy + permanent
smoke persona; stable slugs (`mcasey` / `book-001`) + admin Rename URL +
old-URL redirects; `pi-public` → **`jdbbs-public`** with its own GitHub repo
(`git@github.com:djinna/jdbbs-public.git`), planning material folded in at
`jdbbs-public/2026-pi-symposium/source/` (old `~/pi-book-workshop` renamed
`…MERGED-into-jdbbs-public`, safe to delete).

**Push both repos at the end of every work block:** prodcal →
`git push git@github.com:djinna/jdbbs.git main`; jdbbs-public →
`git -C ~/jdbbs-public push`.

## Standing reminders
- Mike Check (`bookiq@gmail.com`, reg #10, `/mcheck/book-001/`) is a
  permanent fixture: consent ON, auto-added to every announcement, never
  delete. j@djinna.com is BCC'd on every send.
- Tell Mike Casey his sign-in is now **jdbbs.exe.xyz/mcasey/** (same
  password); old links redirect.
- Toby Shorin (#9) has no Factory code. Andrea Leiter (#2) has a code but
  consent=false, so never emailed.

## URL plan for event pages (decided in principle, not built)

Namespace = the cohort slug that already exists: **`/2026-pi-symposium/`**
(roster lives at the root of it today, `handleCohortPage`). Add a directory
mount `GET /2026-pi-symposium/{page}` → `jdbbs-public/2026-pi-symposium/{page}.html`
so handouts publish by saving a file (roster route must keep precedence for
the bare path). Evergreen pages stay flat (`/field-notes`, `/factory`,
`/stylesheet/`). `/workshop` stays where it is (in nine emails); optionally
301 it later. Reserve `2026-pi-symposium` as a client slug.

Register each new page in `site_pages` (admin Pages registry) and follow
`docs/PAGE-DESIGN-HOSTING-VISIBILITY-2026-09-11.md` (1240 shell, theme.css,
canonical masthead/footer). `exedeck.html` is the documented deck exception
(own 960px stage) — reuse its stage for the talk deck.

## Work items — first-order instructions

### 1. Talk deck + handout → `/2026-pi-symposium/talk`
Source: `2026-pi-symposium/source/talk-proposal-2026-07-27.md` (locked title,
25-min outline in six beats, artifact list, "New Nature" hook) and
`/field-notes` for the examples. Build as ONE HTML file with two renderings:
slide stage (clone `exedeck.html` mechanics: keyboard nav, 960 stage) and a
print/handout stylesheet that lays the same sections out as prose. ~12
slides: problem → transmittal handshake → styles as controlled vocabulary +
preflight (the ABS "hard mode" analog from field-notes) → deterministic stack
(Ghosts as the worked example) → calendar + LLM agents as participants →
paid off / leaked → New Nature close. Open decision from the proposal still
open: live demo vs. 60–90 s pre-recorded clips — recommend clips (Factory
build of the smoke manuscript, preflight report). Numbers on slides come from
`jdbbs-public/notes/research-prodcal.md`; re-check before presenting.

### 2. Participant handout: session structure → `/2026-pi-symposium/workshop`
Sources: `workshop-proposal-2026-07-27.md` §"Session plan" (4 sessions:
handshake / conformance & negotiation / deterministic stack + LLM / payoff &
leaks) and `RUN-OF-SHOW-2026-07-27.md` (beat sheets, homework between
sessions, Discord coworking model). **Conflict to resolve first:** proposal
says 4 × 75–90 min, run-of-show plans ±3 h — get the confirmed schedule and
Zoom/Discord links from PI, then write. One screen / one printed page: dates
+ times per session, what each session does, homework after S1–S3, what to
bring, what you leave with, Factory link (`/factory` + their code), where to
get help. Same tone as `/workshop` registration page. Link it from the
cohort roster page.

### 3. `EXAMPLES-IN-THE-WILD-2026-07-27.html` — already live
It is the July 27 version of **`/field-notes`** (same four starred examples +
two shelves; field-notes.html is the maintained copy, restyled to the shared
theme on 09-11). Nothing to remount. If a side-by-side is still wanted, the
old file is at `jdbbs-public/2026-pi-symposium/source/EXAMPLES-IN-THE-WILD-2026-07-27.html`
and can be served temporarily by dropping a one-line `servePublicDoc` route.
The `.md` twin has the NDA-safe notes and verification statuses (✅/🟡/💡)
that the HTML dropped — useful when scripting the talk.

### 4. Other work products worth revisiting (quick sweep)
- `/exedeck` — Aug 24 SIGPfB deck about exe.dev; the *mechanics* and the
  studio voice are the template for item 1. `jdbbs-public/notes/research-*.md`
  = sourced facts behind it.
- `RUN-OF-SHOW-2026-07-27.md` — beat sheets with a tech-readiness checklist
  (hero manuscript timed run, messy sample manuscript, Zen Garden flip
  214↔215). The "messy sample" manuscript was never made; Session 2 needs it.
- `SUBMIT-paste-ready-2026-07-27.md` — final abstract text as accepted; the
  handout should not contradict it.
- `/stylesheet/` — house editorial stylesheet (62 items); cite it in the prep
  guidance (item 5) as "what the copyedit checks", distinct from Word styles.
- `docs/reviews/FACTORY-PASS-SESSION-A-DRY-RUN-2026-09-04.md` — the timed
  dry run of the participant surface; source for realistic timings in item 2.
- `typesetting/scripts/generate-word-template.py` and
  `detect-edge-cases.py` — the authoritative list of style names the Factory
  understands (Normal, Heading 1–3, Block Quote, Section Break, custom
  `tweet-p`/`metadata-p`, …). Item 5 must quote these, not invent them.
- `/2026-pi-symposium` roster page — the natural home for links to items 1, 2
  and the prep page.
- `docs/book-production-deepdive.html` — older architecture explainer;
  probably superseded by field-notes, skim once and retire or link.

### 5. Announcement email TODAY: Word-doc prep guidance
Send from the tracker outbox (`/admin/registrations`) to all consenting
registrants (Mike Check auto-added; you're BCC'd). Content, in this order:
dates/times reminder (once confirmed); "have your Factory Pass redeemed
before Session 1" with `/factory` + note that codes are on the card; the
prep rules — use **paragraph styles, not manual formatting** (Normal /
Heading 1–2 / Block Quote…), no faked section breaks, no pasted colors;
**Google Docs is fine** if they File → Download → .docx and the Word styles
pane shows only the styles they meant to use (no `Normal (Web)`, `Heading 1
Char`, `Style1` outliers — that is the tell); rough content is welcome,
rough *formatting* is what Session 2 fixes; upload to their Factory and run
Inspect before the session if they want a head start. Keep to ~250 words;
plain text is primary. Ideally the same text becomes
`/2026-pi-symposium/prep` so the email can link it instead of carrying it.
Before sending: issue Toby Shorin a code; decide about Andrea Leiter.

## Suggested order
5 (today, time-sensitive) → 2 (needs PI schedule; ask now) → 1 (largest) →
the directory mount + `site_pages` rows when the first page exists → 4 as
you go. All before Sep 19.

---

## Addendum (later on 2026-09-13) — two items added, decisions taken

### 6. Content review against better-documents → `/admin/content-review/`
Standard: `docs/reference/better-documents/` (skill vendored GPL-3.0 +
README note on Anil Dash's 2024 essay). Five passes; report format
*location → problem → fix → severity*; don't alter voice, don't invent
content; house style (Terminal Folio) is out of scope — prose, order,
emphasis, titles only.

Corpus (decided): jdbbs-public `field-notes, workshop, factory, exedeck,
litmags`; app pages `srv/static/landing, index, client, cohort, factory,
housestyle, transmittal`; **plus the two emails** (registration confirm,
Factory Pass). Skip the three anonymized client artifacts and admin pages.

Plan: fan out ~4 pages per subagent on extracted text (not HTML), subagents
return findings only, don't commit. Save **all** findings to
`docs/reviews/CONTENT-REVIEW-2026-09-13.md`; dedupe/rank to **≤20** (cap per
page, weight CRITICAL/MAJOR in passes 1–2). Build admin-only page
`/admin/content-review/` (register in `site_pages`): card = page · location ·
current text · proposed text · reason · severity, **Accept / Edit / Reject**,
decisions persisted (`content_review_items` table) so it survives two
sittings; "Export accepted" emits (file, old, new) patch list that I apply by
hand. Page proposes, never writes to jdbbs-public. Nothing changes until
user decides.

Also: register `/admin/email-preview/` in `site_pages` while in there
(still missing, verified today).

### 7. "Why book" → short page `/2026-pi-symposium/why-book` + weave
Decided: ~400-word sourced page, linked from the deck (beat 1 "the
problem"), the `/2026-pi-symposium/` overview, roster, and the prep email;
a 3-sentence weave in deck + overview pointing at it.

Thesis (agreed): demand is shrinking and readers trust the object less;
supply is four million titles/yr, mostly machine-made and unvetted; the
book's remaining edge is that it was *finished* — chosen, edited, checked,
fixed in form. A factory makes the finishing legible and repeatable for a
human author, with the machine doing conformant work inside the rules.
Protocol is how you tell a finished book from a generated one. Lands on
New Nature + "LLM as conformant participant".

Sources gathered (verify primaries before slides):
- Sean deLone, *Dear Head of Mine*, "The Nonfiction Book Market is
  Collapsing", 2026-09-08 — fiction/nonfiction inversion in top-25
  hardcovers 2015→2025; reads it as a trust-in-authority collapse, not a
  reading crisis. <https://dearheadofmine.substack.com/p/the-nonfiction-book-market-is-collapsing>
- Rose Horowitch, *The Atlantic*, Aug 2026 cover "The End of Reading Is
  Here" — 51% of US adults read zero books; "postliterate". Anchor of the
  season; NPR (Jul 11), Ringer (Jul 28), Defector booksellers' rebuttal
  (Sep 1). Paywalled — user should read it before we quote.
  <https://www.theatlantic.com/magazine/2026/08/reading-crisis-postliterate-age/687618/>
- Supply side: >4M new US titles in 2025 (+⅓ y/y), ~3.5M self-published;
  e-books/week tripled since Nov 2022; "Aug 2026 study" on slop diluting
  revenue for all books (second-hand via Kaspersky blog 08-13 / Baptist
  News — find the primary).
- Guardrails: Authors Guild "human authored" attestation; Hachette pulled
  *Shy Girl*; $2.5M debut deal withdrawn Jul 2026 over suspected AI
  (WSJ-derived, Fudzilla 08-18; Jane Friedman FAQ).
- Counter-signal: ALLi mid-2026 — indie authors out-earn trad (Author Media
  07-21; second-hand).

## Revised order
6 (content review, ~half session) and 7 (why-book page) **before** the deck
and workshop handout, since both change what those pages say. Then:
5 prep email (today-ish) → 2 workshop handout (needs PI schedule) → 1 deck
→ directory mount + `site_pages` rows → 4 sweep.

---

## Addendum 2 (2026-09-13, late) — content-review decisions in, nothing applied yet

All 19 cards decided at `/admin/content-review/`: 15 accepted, 4 edited,
0 rejected. Export saved at `scratch/content-review/export-1.md` (gitignored;
regenerate with `GET /api/admin/content-review/export`). **Nothing has been
applied to any file yet.** Mark each card "applied" on the page as it lands.

Edits and notes from the user that change the work:

- **A-F25 (Factory Pass email, First steps).** Real flow is *transmittal →
  generated Word template → upload*. Put the transmittal at step 1 ("it is
  the spec your book is built from and generates your Word template"),
  upload at step 2 ("workshop attendees: be ready to do this in session 2,
  Mon Sep 21"). User asked that this flow be stated consistently everywhere
  — emails (`srv/passes.go` ~1030 text, ~1063 HTML), customer factory page
  step 1 (`srv/static/factory.html`), public `/factory` steps, `/workshop`.
- **A-F26.** Live-help minimum changed **30 minutes → 1 hour** in the email
  edit. Same string lives at `jdbbs-public/factory.html:166` and `:204`,
  `srv/static/factory.html:214`, `srv/passes.go:977`. Change all four so
  they agree (user has been told; confirm it was intentional if in doubt).
  Also give the three support bullets a "Support" heading.
- **E-F1 (stylesheet intro).** Reorder the sections in
  `srv/static/housestyle.html` so the intro's references come up in order
  (author-facing §1 and §5 first, then "what we do" §2), renumber, and
  update the intro's § numbers to match.
- **E-F11 (transmittal intro).** Work the phrase *mise en place* in a
  couple of times on the transmittal page (`srv/static/transmittal.js`) —
  the user finds it helps clients grasp why the transmittal comes first.

Where each accepted change lands:
- `jdbbs-public/workshop.html` — A-F2 (line under `.subtitle`, l.115), A-F3 (l.126)
- `jdbbs-public/factory.html` — B-F1 move Availability block (l.172–173) above l.157; 1-hr minimum
- `jdbbs-public/litmags.html` — B-F15 compute takeaways from the table for the three charts
- `jdbbs-public/field-notes.html` — C-F1 intro, C-F2 order sentence, C-F10 kicker
- `srv/static/landing.html` — C-F14 new-author line after hero (link /factory)
- `srv/static/client.html` / `app.js` — C-F20 portal gate line
- `srv/static/housestyle.html` — E-F1 (edited), E-F2 §2 rewording
- `srv/static/transmittal.js` — E-F11 (edited), E-F12 Mark Final help, E-F13 five "Priority field" strings
- `srv/static/factory.html` — F-F1 "Inspect — the preflight —"; 1-hr minimum; step-1 template sentence
- `srv/passes.go` — A-F24 para 1, A-F25 steps, A-F26 support heading (text + HTML variants)
- roster data — A-F15 hide Mike Check (reg #10) from `/2026-pi-symposium`; keep in `/admin/registrations`

jdbbs-public edits publish instantly (read at request time); `srv/` edits need
`make build && sudo systemctl restart prodcal`. Run `go test ./srv/` — the
email tests may assert on old strings.

**Next review block:** agreed to run another ~20 after the new pages exist
(why-book, handout, deck), ~Sep 17–18, drawing from the remaining ~90
findings in `CONTENT-REVIEW-2026-09-13.md` plus the new pages.

## Addendum 3 (2026-09-14) — smoke 1 done, handout live, Book 2 smoke next

State at handoff: prodcal `fe28e81`, jdbbs-public `70b78a2`, both pushed, both clean. Context ~60 % (post-compaction session); files read in full: 3 (all under threshold).

**Done today.** S2/S4 times fixed everywhere (`1b58cb7`); favicon (`25eb9ea`); Mike Check Book 1 smoke end-to-end → 4 fixes (`eb590c0`: custom styles reach the docx build, shared `typstStyleIdent`, detector whitelist, template copyright year); template_ready email pathway confirmed live (#27), build-ready #28. Context hygiene rules + `scripts/readguard.sh` (`3375831`, see `docs/CONTEXT-HYGIENE.md`). **Participant session guide live at `/2026-pi-symposium/workshop`** (jdbbs-public, registry migration 034, linked from roster + prep email). Prep email `scratch/prep-email.txt` complete incl. Discord server/call/chat links — **user sends from `/admin/registrations`**; user is editing the handout HTML locally and will upload — diff text changes in, keep markup.

**First, short: factory map (~30 min).** User asked where the LLM sits; answer: nowhere in the pipeline — Inspect is rule-based (S2), the LLM appears once in S3 as a *conformant author* demo whose output goes through the same Inspect + build (RUN-OF-SHOW S3 beat 5; also names dev-time + copyedit as the real touchpoints). Build `/2026-pi-symposium/map` in jdbbs-public (handout shell, Mermaid via CDN for now): stages S1→S4 as nodes, each labelled with **actor** (author / Inspect rules / human Keep-Strip-Convert / deterministic Word→Typst→PDF+EPUB / LLM-as-author) and **route** (`/client/...`, `/admin/...`). Register in site_pages, `unlisted` until user decides to share; link from the admin home as a clickable route map (nodes are routes). Hand-render to SVG later only if it goes in the deck.

**Then: Mike Check Book 2 smoke — the messy path** (Book 1 was the happy path). Mike Check is a test persona (project 17, `/mcheck/`), not a person. Fresh session; the run is output-heavy.

- [x] Book 1 state: transmittal `final`, spec pulled, 2 credits left. Leave as is; Book 2 is a **new book** on the same project (check `/mcheck/` shows both).
- [x] Transmittal: different trim (**6×9**), different formats mix, no or different custom styles (e.g. one paragraph "Epigraph", no character style). Mark Final → confirm template_ready email fires again.
- [x] Manuscript written **outside** the template: Google-Docs-style export — `scratch/make-ms-messy.py` (python-docx): manual bold/italic for headings, blank-line faked scene breaks, a pasted colored run, mixed straight/curly quotes, a "Normal (Web)"-style stray style, a hand-numbered list, a "14 Sept — …" dateline (Book 1 false-positive check).
- [x] Inspect → expect high findings; record counts. Keep/Strip/Convert.
- [x] Import template styles into the messy doc (Word Organizer path the prep email promises) — simulate via python-docx or apply template as base; re-Inspect → should drop.
- [x] Build 1 (dirty) → does it fail or produce ugly output? Build 2 (clean) → **second build on same book**: credit decrement to 1, artifacts superseded, build-ready email again.
- [x] Book 1 soft findings: chapter-opener page break (`scratch/mcheck-build.pdf` p.3–4) — check in Book 2 PDF too; dateline false "manual list item".
- [x] Log every catch; fix small ones inline, park big ones here.

*(All eight ticked 2026-09-16 — the run happened in Addendum 5 below; the boxes were never updated, which sent one later session back here by mistake. Results + parked items live in Addendum 5.)*

**After smoke 2:** admin doc editor (`docs/IDEAS.md` 2026-09-14) → talk deck `/2026-pi-symposium/talk` → content-review round 2 (~Sep 17–18, `/admin/content-review/`, 19/19 decided in round 1) → freeze Sep 19.

**Exact next action:** new session; `git status` + `git log -1` in both repos; open `/mcheck/` as admin and start the Book 2 transmittal.

## Addendum 4 (2026-09-14, evening) — map live, announce tooling, Book 2 smoke still next

State: prodcal `d2b1888`, jdbbs-public `8fe3201`, both pushed, both clean. Context ~55 % at handoff (post-compaction session); no files read in full over threshold.

**Done.** Session guide: author's local copy edits merged (`6216334`). **Factory map live at `/2026-pi-symposium/map`** (Mermaid via CDN, actor-coloured S1–S4, route ledger; registry `unlisted`, migration 035; linked from guide continue-nav + admin masthead "Map"). Admin registrations: **Reuse** button on past announcements (reloads subject+body; strips `[TEST]`); **minimal Markdown** in announcement bodies (`[text](url)`, `**bold**`, `*italic*`, bare URLs autolinked; text part gets `text (url)`; `srv/announcement_md_test.go`); Preview merge now renders the HTML letter in a sandboxed iframe (srcdoc set via property — `esc()` doesn't escape quotes) with plain-text toggle. Sign-offs now `[jdbb] studio` in all emails. Prep email `scratch/prep-email.txt` synced to the guide edits; `[TEST]` #29 sent 21:29 UTC to Mike Check (BCC j@). **User sends the real one** via Reuse → drop `[TEST]` → tick cohort → Send.

**Parked for the admin doc editor:** editable email signature + footer line (currently constants in `srv/email_shell.go`).

**Next: Book 2 smoke** — checklist unchanged in Addendum 3. Then doc editor → deck → content-review round 2 → freeze Sep 19.

## Addendum 4b (2026-09-14, late) — talk deck rev 1 up; rev 2 decisions taken

**Talk:** *Book as Protocol Stack: Building a protocolized book-production factory* — **Wed Sep 23, 17:30 UTC** (1:30 PM EDT), 25+5, SIGP4B. PI listing: https://protocol-institute.org/events/protocol-symposium-2026/#p-77 (use that title, not the proposal's).

**Rev 1 live** at `/2026-pi-symposium/talk` (jdbbs-public `c4ffa03`; registry migration 036 in prodcal `c1d7682`, applies on next restart). Cloned from `exedeck.html` stage (one tall page, arrow keys, ToC ≥1340 px / jump menu, print = slide per page) but with the legacy inlined tokens/masthead/theme-bar CSS removed — theme.css + canonical footer only.

**Rev 2 — decided with the author, not yet done:**
1. Drop the New Nature slide. Keep the observation (rules we chose vs. rules that hardened by accident); say "New Nature" once in the close because the accepted abstract ends on it.
2. Drop *The Twitter Years* as an example. Ghosts only.
3. **Rewrite the prose.** Rev 1 is proposal-speak cut into slides (150–250 words each). Rev 2: slides go sparse (plain first-person sentence as title, ≤ ~50 words, at most one short table/list); the talking moves into `<aside class="notes">` per slide, shown only in the print/handout rendering. Then run the better-documents five passes (`docs/reference/better-documents/SKILL.md`) and report what changed.
4. Clips still to record: Inspect on the messy manuscript (`scratch/mcheck-book2-messy.docx` from smoke 2), and a Factory build.

**Admin doc editor (next after smoke 2, not built):** `/admin/docs/` lists jdbbs-public files from the Pages registry, textarea + save + view link; publishes instantly (servePublicDoc reads disk). Include editable email signature + footer line. No preview in v1.

**Smoke 2** running in subagent `book2-smoke` (owns prodcal working tree; has landed `5cf3793`, `2f240eb`, working in `srv/static/factory.js`). It will append Addendum 5 itself.

## Addendum 5 (2026-09-14/15) — Book 2 smoke: messy path

Run by subagent `book2-smoke`. Test persona Mike Check, client `/mcheck/`. Book 1 (project 17) untouched: transmittal `final`, pass 2 of 3 credits left.

**Per step**

1. **Book 2 = new project** (a "book" is a project here: one transmittal + one pass per project). Created from the client portal as Mike (`POST /api/clients/mcheck/projects`) → project 19 `/mcheck/book-002/`; `/mcheck/` lists both. **Catch (fixed, `5cf3793`):** a portal-created second project has no pass → factory read-only, and the only admin pass path (`fulfillPass`) always makes a *new* client + password. `POST /api/admin/passes {project_id}` now attaches a pass to an existing project, inheriting customer name/email from the sibling pass (test `TestAdminAttachPassToExistingProject`). API only — no admin UI button yet.
2. **Transmittal** as Mike: 6×9, printer format "Other", deliverables PDF+EPUB+DOCX, one custom paragraph style `Epigraph`, no character style. Draft→final → `template_ready` #45 (to bookiq@, BCC j@). Template pulled to `scratch/mcheck-b2-template.docx`.
3. **Messy manuscript**: `scripts/make-ms-messy.py` (committed; copy in `scratch/`) → `scratch/mcheck-book2-messy.docx`: 10 chapters, headings as bold 16pt centered Normal, blank-line + `***` scene breaks, green Verdana run, manual-italic epigraph, hand-typed `1.`–`4.` list, mixed quotes, tab indent, double spaces, `Normal (Web)` stray style, `14 Sept —` dateline. `--clean TEMPLATE` mode rewrites it into the template's styles (simulates Organizer import + restyle).
4. **Inspect (dirty, book 15)**, before fixes: 43 findings (3 high / 14 med / 26 low): colored 1, unusual_font 1, manual_list 4, manual_break 3 (`***` only), direct_spacing 18, manual_formatting 14 (*the 11 bold chapter headings — "auto-preserved, no action needed"*), observed_style 1, undeclared_custom_style 1. **Three catches, fixed in `2f240eb`:** (a) no finding at all for headings that carry no Heading style — added `heading_lookalike` (bold-throughout / ≥ body+3pt / "Chapter N…"; **high** when the doc uses no Heading styles, else medium); (b) blank-line scene breaks invisible — runs of ≥2 empty paragraphs now a `manual_break`; (c) HTML report showed 2 high / 42 total while the JSON summary said 3 / 43, because `undeclared_custom_style` was appended only in Go — the detector now emits it (Go dedupes). After: **64 findings, 14 high / 14 med / 36 low** (heading_lookalike 11, manual_break 13). Dateline **not** falsely flagged as a list (Book 1 soft finding not reproduced). Book 1 manuscript + the template re-checked: no new false positives. `factory.js` type labels for the new/unlabelled types (`3931b43`). Keep/Strip/Convert: there is **no API or UI for the ruling** — it is a human reading of the report; noted, not exercised.
5. **Import template styles** (simulated): clean docx = 49 paragraphs, styles Title/Subtitle/Copyright/Epigraph/Heading 1/First Paragraph/Normal/Section Break. Re-Inspect (book 16): **64 → 4** (the 4 hand-typed list lines, correct).
6. **Build 1 (dirty)**: did *not* fail — produced ugly output as predicted: 2 pages, no chapter breaks, `***` came out as `˘ ˘ ˘` (each `*` mapped to the breve), EPUB one `ch001`. Trim right (432×648 pt). Credit 3→2, `build_delivered` #46. **Build 2 (clean)**: 12 pages, 6×9, **every chapter opens on a fresh page** (p3 ch1, p4 ch2 …), EPUB 11 chapter files + nav, colour/Verdana not carried. Credit **2→1** (ledger rows 7, 8), both books' outputs listed newest-first, `build_delivered` #49. Files: `scratch/mcheck-b2-{dirty,clean}.{pdf,epub}`.

**Parked (owner decision / bigger):**
- Chapter openers start on the next page, not the next *recto* (p4 is a verso) — same as Book 1's soft finding; a spec/Typst option.
- Word `Title`/`Subtitle` paragraphs vanish from the print PDF (p1 shows only "Mike Check"); the half-title presumably comes from the spec. Decide: map Title/Subtitle → title page, or have Inspect say they're ignored.
- Dirty build "succeeds": should Build warn/confirm when Inspect found `heading_lookalike` high findings (no chapters will be found)? UI copy today: "you can build anyway".
- ~~`***` → `˘ ˘ ˘`: the section-break glyph substitution applies per `*`~~ **Not a bug** (checked 2026-09-16): `˘ ˘ ˘` *is* the `breve` ornament (`series-template.typ:235`); the clean build renders identically. Pandoc reads a lone `***` as a horizontal rule → `#horizontalrule` → `section-break`, so a hand-typed `***` already collapses to one Section Break.
- Transmittal data-model drift: default JSON has `files.archives`, the form writes `files.deliverables`; nothing server-side reads either.
- Admin UI for "attach pass to existing project" (API exists now).

**Workshop-ready?** Yes for the messy path as designed: Inspect now names the real problem (headings without Heading styles) as high, the template-import route drops findings 64→4, and the second build meters correctly. Repo pushed; `prodcal.service` rebuilt and restarted.

Metrics: context ~45 % at handoff; files read in full over threshold: 0.

## Addendum 6 (2026-09-15) — talk deck rev 2 live; admin doc editor built

**Deck rev 2** at `/2026-pi-symposium/talk` (jdbbs-public `7968fc4`): 12 sparse slides (≤ ~50 words on screen), speaker script in `<aside class="notes">` per slide — hidden on screen, shown by the `n` key or in print (= handout; clip cues `.note` hidden there). Ghosts only; "New Nature" once in the close. Better-documents five passes run and reported to the author, who is reviewing. Clips still to record.

**Admin doc editor `/admin/docs/`** (this commit; migration 037 `studio_settings` + registry row):
- Lists every `site_pages` row with owner `jdbbs-public` (the registry is the allow-list — unregistered paths are 400), with size, mtime, and git status from one `git status --porcelain` in `publicDocsDir()`.
- Open → textarea; **Save** writes temp+rename, normalises CRLF, refuses empty, and returns 409 if the file's mtime moved since load. Live on the next request (servePublicDoc reads disk). Ctrl/Cmd-S saves.
- **Commit to git** stages and commits that one file (no push — shell step). "Nothing to commit" when clean.
- **Email strings**: `email_signature` (default `— Jenna\n[jdbb] studio`) and `email_footer` (default `https://jdbbs.exe.xyz/ · Reply to this email to reach Jenna.`), stored in `studio_settings`, cached in-process, loaded at startup and on save; empty value = revert to default. `emailSignoff()`, new `emailSignoffText()` and `emailFooterHTML()` in `srv/email_shell.go` read them; the plain-text sign-offs in `announcementText` and the two pass emails use `emailSignoffText()`. The formal `Jenna Dixon · [jdbb] studio` lines (welcome/receipt) are unchanged. No HTML allowed in values.
- Admin masthead gets a **Docs** link. API: `GET /api/admin/docs`, `GET|PUT /api/admin/docs/file`, `POST /api/admin/docs/commit`, `PUT /api/admin/settings/{key}`.
- Not in v1: preview, Markdown files under `notes/`, push, multi-file commits.

**Next:** author review of deck rev 2 → record the two clips → content-review round 2 → freeze Sep 19.

## Addendum 2026-09-16 — Factory Pass store (Stripe Checkout) live in sandbox

**State.** Store code is complete, deployed, and switched ON against the Stripe
**sandbox** `jdbbs`. Two exe.dev Stripe integrations, both restricted keys, both
attached to `vm:jdbbs` (exe.dev's "verify" fails on them because we don't grant
Balance — "Store anyway" is correct):

| integration | host | key | scopes |
|---|---|---|---|
| `stripe-test` | `https://stripe-test.int.exe.xyz` | sandbox `rk_test_…oxlO` | Checkout Sessions, Products, Prices, Coupons, Promotion Codes **Write**; Events **Read** |
| `stripe` | `https://stripe.int.exe.xyz` (code default) | live `rk_live_…xd9V` | same |

`.env` has `PRODCAL_STORE=on` and `PRODCAL_STRIPE_URL=https://stripe-test.int.exe.xyz`
— that second line is what keeps the server on the sandbox. **Going live is:
delete that line, restart, check the log for `store: catalog ready`.** Nothing has
been created in the live account yet (verified 2026-09-16).

**Proven end to end on 2026-09-16** (pass id 5, `mcheck2/book-001`, bookiq@gmail.com):
pass + `+3 builds` add-on + `PYB149` → $198 (promo hit the pass only) → thanks page
→ 6 builds, fulfilment email; second thanks-page hit = no second pass. Then a
`+6 months storage` add-on bought from the API and *never returned to* — the
60 s event poller fulfilled it (expiry Mar→Sep 2027, add-on email sent).
`optional_items`, `custom_fields`, `custom_text`, `expand[]=discounts.promotion_code`
all accepted by the real API.

**Pieces.** `db/migrations/038-store.sql`, `039-store-admin-page.sql`;
`srv/stripe.go` (client), `srv/store.go` (catalog by lookup_key, checkout,
`fulfillStoreSession`, poller, admin handlers), `srv/store_email.go`,
`srv/store_test.go` (6 tests, fake Stripe via httptest); pages
`srv/static/store-thanks.html` (`/factory/thanks`), `srv/static/store-admin.html`
(`/admin/store/`, linked from admin masthead); customer add-on buttons in
`srv/static/factory.js` footer (`?order=` confirm); `~/jdbbs-public/factory.html`
$349 + `#buy-on`/`#buy-off` toggled by `/api/public/store/config`.
Routes: `POST /api/public/store/checkout`, `GET /api/public/store/session`,
`GET /api/public/store/config`, `GET /api/admin/store/orders`,
`POST /api/admin/passes/{id}/status`.

**Freeze decision (user, 2026-09-16): leave `PRODCAL_STORE=on` against the
sandbox through the workshop** — the Buy flow is demoed live in the sessions with
card 4242. Do not turn it off. Attendees use the code form, not Buy.

**Go-live checklist (after Sep 23).**
1. Dashboard (live): Settings → Emails → turn on customer receipts; Branding
   (logo, colour) if wanted. Statement descriptor is set per-product as `JDBB STUDIO`.
2. Remove the `PRODCAL_STRIPE_URL` line from `.env`, `sudo systemctl restart prodcal`,
   confirm `journalctl -u prodcal | grep "store: catalog ready"` — this bootstraps
   the three products, coupon `pyb149-200-off`, promo `PYB149` (exp. 31 Oct 2026) in live.
3. One real $349 purchase with own card, refund from dashboard, Revoke on `/admin/store/`.
4. Update `/factory` copy (`/admin/docs/`) — "Purchasing opens after the workshop"
   is the off-state text; on-state text is in `#buy-on`.

**Known gaps / ideas.** No admin "refund" button (dashboard does it; then Revoke).
Poller looks back 72 h on cold start — fine. Test-mode dashboard links in
`/admin/store/` orders table point at live `dashboard.stripe.com/payments/…`;
sandbox needs `/test/` — cosmetic. `store_orders.note` unused. Legacy "Test mode"
sandbox in the Stripe account can be deleted.
