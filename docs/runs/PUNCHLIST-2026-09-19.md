<!-- exported 2026-09-19 15:28 UTC by scripts/punchlist-export.py; source scratch/run/CHECKLIST.md + notes.json -->

# Punch list 3 · Sat 19 Sep → workshop (Mon 21 / Tue 22) → talk (Wed 23)

Legend: **YOU** = Jenna's turn · **ME** = Shelley's turn · **BOTH** = look together.
Ticked = done · ◐ = in progress. Click a box to cycle ☐ → ☑ → ◐. **note** opens a reply box under any item (⌘↵ saves; paste screenshots straight in) — I read those back. Bottom bar adds a new item to the Inbox (top). Refreshes every 15 s (pauses while you type).

Previous lists archived at `/admin/runs/` (list 2: 35 of 55 ticked → `PUNCHLIST-2026-09-18-list2.md`). Item numbers carry over so note threads stay attached. Checkpoint: `checkpoint-2026-09-18-pre-workshop-v2`.

**Freeze:** Sat 19 – Sun 20 fix / test / clean allowed (build → `go test ./srv/` → restart → one factory smoke as `pinstitute`). Mon 21 – Wed 23 hotfix-only. Store flips to live Stripe by timer **Wed 23 Sep 00:00 HKT**.

## 0 · Inbox — new items, untriaged (Shelley moves them into a section)

- [~] 0.10 new: https://desertant.com/press/ -- using this page as a model, can you draft a similar page for me in the factory? I also really like the… — DRAFT LIVE at /press (1cbe387); warm bg #FBFAF4 live app-wide; YOU fill the [bracketed] bits (bio, entity, founded, city); assets done

  > **shelley** · 2026-09-19 12:37 UTC  
  > Draft is up: https://jdbbs.exe.xyz/press — same skeleton as Desert Ant’s (lede · About · Facts · Products · Brand assets · The studio), in our chrome. Everything on it is true today except the orange [bracketed] bits, which are yours to fill: one sentence of biography if you want it, legal entity (or delete the line), founded year, city. Their “Independent research” block I left out — we don’t have third-party numbers yet; if you have a quote or two (Bringhurst, a press stat), that’s where they’d go. Wordmark SVG and a 1200×630 social card are the missing assets; I can generate both from the favicon outlines when you want them.
  >
  > The colour: done, and it’s live app-wide. Background went from our cool #FCFDFD to their #FBFAF4, and the hairlines/progress tracks moved to the same warm axis so nothing looks blue against it. Text and the cyan are unchanged. For scale: Anthropic’s cream is #F0EEE6 — ours is now about a fifth of the way there from white, which is the “hint, not tint” you asked for. Hard refresh to see it; if it reads as too much or too little on your screen, say and I nudge one value.

  > **shelley** · 2026-09-19 12:40 UTC  
  > Brand assets are no longer “to follow”: wordmark SVG (paper and ink versions, plus [jdbb] alone) and a 1200×630 social card are generated from the JetBrains Mono outlines and linked on the page; the page also carries the og:image so a pasted link shows the card. Left for you: the four orange placeholders.

- [~] 0.14 new: I'm tempted to go back to using a domain I've owned for many years, jdbb.net, as my main contact for this project. … — ideas posted; chat; no build

  > **shelley** · 2026-09-19 12:49 UTC  
  > Ideas only, as asked. Two separate things: where mail goes OUT from (the factory’s automated sends) and where people write TO you (contact address). They don’t have to be the same domain, and today they aren’t: sends go from factory@mail.jdbb.studio (Resend, set up yesterday, with SPF/DKIM so they land), and Reply-To on every one of them is j@djinna.com.
  >
  > Options for the contact side:
  > 1. Keep sending from mail.jdbb.studio; make j@jdbb.net (or hello@) the public contact and Reply-To. Cheapest: point jdbb.net’s MX at whatever already hosts your mail (Fastmail/Google/iCloud+ all take custom domains; you already own the name, so it’s a DNS change and an alias — an hour). Nothing on our side changes except one env var for Reply-To and the address on /press, /factory, the footer.
  > 2. Send from jdbb.net too (factory@jdbb.net) — add jdbb.net as a second Resend domain, three DNS records, and switch PRODCAL_MAIL_FROM. Half an hour, but only worth doing if you want one domain everywhere. Warm-up caution: a fresh sending domain right before the workshop is the one thing I’d not do this week.
  > 3. jdbb.studio as an inbox — it is only a sending subdomain today; receiving there means an actual mailbox provider or Resend’s inbound (which forwards to a webhook, not a mailbox — fine for the factory to read replies programmatically, wrong for you as a person).
  >
  > Redirects: web-side is easy — jdbb.net → jdbbs.exe.xyz is one exe.dev proxy/DNS setting, and both names can serve the site. Mail-side “redirect” is just forwarding, also easy.
  >
  > My pick: 1 now (after the workshop, as you say), and decide on 2 after you’ve seen a few weeks of real mail. Happy to chat whenever; nothing is built.


## 1 · This weekend — ME (from list 2, decided)

- [~] 0.17 **One-page factory (C)** — transmittal becomes section 1 of `/{client}/{project}/factory/`; step strip = in-page anchors; finish button “Continue to 2 · Upload →”; transmittal URL redirects; TRANSMITTAL leaves the top nav. Comps `comp-017-C-*`. Fresh session, plan first.

  > **jenna** · 2026-09-19 14:39 UTC  
  > "CONTINUE TO THE FACTORY →" on our ms transmittal is part of what's confusing -- the transmittal is part of the factory. It's the first part of the factory flow.

  > **shelley** · 2026-09-19 14:43 UTC  
  > Three comps, all mock-ups on the live pages (nothing changed). Agree the “Continue to the Factory” label is the tell — in every option below the transmittal is step 1 *of* the factory and the button says “Continue to 2 · Upload →”, and the top nav loses its separate TRANSMITTAL entry (just YOUR BOOKS · FACTORY).
  >
  > **A — Factory widens to match (image 1).** The factory takes the same 1240 shell and header as the transmittal; steps 2–5 stay stacked in a ~720 reading column on the left, and a right-hand rail carries what the transmittal decided (status, trim, type, front matter, rights, edit link), the latest build and the pass. The transmittal keeps its own page but both pages now share the same width, masthead and step strip, so the jump is gone. Medium effort; also gives the factory a use for its blank right half.
  >
  > **B — Transmittal narrows to match (image 2).** The transmittal becomes one ~720 column like the factory: same header with a “// TRANSMITTAL” kicker, sections stacked instead of two-up (the two-field rows stay two-up inside). Smallest change; the form gets about twice as long to scroll.
  >
  > **C — One page (images 3–4).** The transmittal *is* section 1 of the factory page: factory header (“// FACTORY · Obliquities · 1 of 3 builds left”), the strip, then “// 1 · TRANSMITTAL” with its History/Print/Email/Word template actions on the section rule, the two-column form, and its finish block closing with “Continue to 2 · Upload →”. Steps 2–5 follow below; the strip becomes in-page anchors. One URL for the whole flow; the transmittal URL redirects there. Biggest change, most coherent.
  >
  > My pick: A now (ready for Monday), C as where this goes after the workshop. Which?
  >
  > ![screenshot](img/comp-017-A-wide-factory-rail.png)
  >
  > ![screenshot](img/comp-017-B-narrow-transmittal.png)
  >
  > ![screenshot](img/comp-017-C-one-factory-top.png)
  >
  > ![screenshot](img/comp-017-C-one-factory-bottom.png)

  > **jenna** · 2026-09-19 14:55 UTC  
  > Good notes, I think we just bite the bullet and do C now.

- [x] 0.8 **Typography sampler** (part 1 of 0.8) — 3-pairing sample PDF built straight from Typst + click-to-enlarge chips under the Typeface row; helper session building; then note + tick here. Part 2 (defaults, “Your own” break mark) is live.

  > **shelley** · 2026-09-19 12:25 UTC  
  > Part 1 done and live — the defaults. Diagnosis from your Obliquities PDF: every book was getting the same margins (¾″ all round, whatever the trim) and a stored leading of 2pt, which on a 6×9 gave an 85-character line set 10 on 8.6 — tighter than solid. That’s the “pretty bad”. Bringhurst: 45–75 characters, 66 ideal; leading ~120–135 %; bottom margin deeper than top, inner narrower than outer.
  >
  > Now derived from the trim: 6×9 → inside 0.88″, outside 0.75″, top 0.8″, bottom 1.0″; 10.5pt on 13.4 (computed from the typeface’s cap height, so it holds when the face changes); 1.25em indent; small trims keep 10pt. Obliquities lands at ~70 characters a line and reads like a book. Your next build picks it up; page count goes 77 → ~110, which is the honest number.
  >
  > Part 2 — putting choices back on the transmittal. Proposal, deliberately few, each a radio row with a one-line description:
  > • **Typeface**: 3 pairings from TYPOGRAPHY_PAIRINGS.md (e.g. Plantin + Proxima “studio house”, Libertinus + Source Sans “open classic”, EB Garamond + … “literary”) + “studio’s choice”.
  > • **Text size**: Compact · Standard · Generous (10/12.5, 10.5/13.5, 11/14.5 — one knob, shown as “roughly N words a page”).
  > • **Section breaks**: white space · breve · ornament.
  > • **Paragraphs**: indented (default) · block with space between (for essays/nonfiction).
  > Alignment stays justified+hyphenated (a real choice only for poetry, which the Verse style already handles). Anything else stays studio-side in the spec. Then a review pass over a real build for each pairing. I’d hand this to a fresh session/subagent — half a day.

  > **shelley** · 2026-09-19 12:42 UTC  
  > Part 2 is now in progress (a subagent is building it while I keep the list moving): four radio rows on the transmittal — Typeface (Open classic = Libertinus/Source Sans · Studio house = Plantin/Proxima · Literary = EB Garamond · studio’s choice), Text size (Compact/Standard/Generous), Section breaks (space/breve/ornament), Paragraphs (indented/block). Report here when it lands; I’ll want your eye on the three sample spreads.

  > **shelley** · 2026-09-19 13:09 UTC  
  > Part 2 landed (2f205eb, 21bbdc2, 6c45b6d): the transmittal has a // TYPOGRAPHY section with four rows — Typeface (Studio's choice / Open classic Libertinus+Source Sans / Studio house Plantin+Proxima / Literary EB Garamond, now vendored, OFL), Text size (Compact/Standard/Generous), Section breaks (space/breve/ornament), Paragraphs (indented/block). First answer in each row is the default. Two pipeline fixes came out of it: Word “Body Text” paragraphs no longer become hundreds of #block wrappers, and the specialised template path is actually used (stray Proxima in the open-classic pairing gone). Spreads for the three pairings are in scratch/typo/ — I'll pull them onto a page for you next if you want to compare before Monday. Ghosts' free builds are used up (3/3) doing this; further test builds need a refund or another test project.

  > **jenna** · 2026-09-19 14:47 UTC  
  > 1/Feels like it would be pretty helpful to roll out a sampler PDF that had, let's say, four pages showing each of the main styles. Chap opener + 3 pp so they can see a full spread with running heads. I suppose no reason not to build out a pretty full sampler, take a first swing at that? And let's include small chips that people can click on to enlarge, in addition to giving them a link to download the PDF with the full sampler. 
  > 2/ for "Section breaks" let's add a field for their own break character choice?

  > **shelley** · 2026-09-19 14:55 UTC  
  > 2/ Done, live: Section breaks has a fourth choice, “Your own”, which opens a small centred field (up to 24 characters, placeholder * * *). Goes through everywhere — print PDF, EPUB, the Word template’s Section Break sample, the transmittal email summary. Empty falls back to the breve.
  > 1/ Sampler PDF + click-to-enlarge chips under the Typeface row: in progress with a helper (built straight from Typst, no credits used). Will report here when it lands.

  > **shelley** · 2026-09-19 15:13 UTC  
  > 1/ Sampler live. On the transmittal, under Typeface: three small page previews (Open Classic · Studio House · Literary), same chapter set three ways; click one to enlarge, or “Download the full sampler (PDF)” — 20 pages, 6×9, built straight from Typst (typesetting/scripts/build-sampler.sh, fonts subset-embedded only). All three parts of 0.8 done → ticked.

- [~] 5.20 **Typst 0.13 upgrade on the VM** — today (Jenna, 19 Sep) after 0.17: side-by-side build, swap binary keeping 0.12 for rollback, factory smoke; unlocks `par(costs:)` for 5.5

  > **jenna** · 2026-09-19 15:12 UTC  
  > why post workshop? let's do it today?

  > **shelley** · 2026-09-19 15:13 UTC  
  > No technical reason — it was parked for the freeze (swapping the typesetting engine two days before attendees build). Just checked: the 0.13.1 binary already on the VM compiles Ghosts with no warnings, same 117 pages as 0.12. So yes, today: after 0.17 lands I'll (1) build the last pinstitute book with 0.13 side by side and diff page counts + text, (2) swap the binary keeping 0.12 as `typst-0.12` for a one-line rollback, (3) run the factory smoke. If the diff shows anything odd I'll stop and show you. Moving 5.20 to §1.


## 2 · Your side — YOU

- [ ] 2.8 Today's book with colour images: run Inspect, tell me what it says about the images and how the print PDF's grey looks (5.15 landed; Ghosts had nothing to test on)

## 3 · P2 — talk deck from Venkat's "Have your factory call my factory" — BOTH

- [ ] 3.2 YOU — cut / reorder / add beats
- [ ] 3.4 BOTH — run-through in the browser; timing
- [ ] 3.5 ME — push; link in the handoff

## 4 · Workshop watch (Mon 21 / Tue 22) — BOTH

- [ ] 4.1 `/admin/factory/` (Floor) on a second screen; Shelley session open with `factory-tail`
- [ ] 4.2 Attendees free via hand-granted passes (New pass form) — no coupons, no $0 checkouts (form + attendee login verified Fri, 2.5)
- [ ] 4.3 Watch for `build failed`, `401` bursts (password confusion), builds "queued" (> 2 at once)
- [ ] 4.4 Tue night: confirm `systemctl list-timers prodcal-store-live.timer` still armed; Wed 00:00 HKT check the go-live email arrived and `/api/public/store/config` says live

## 5 · Queued — ME unless marked

- [~] 0.9 new: I'd like to at least start building out how someone else can hit the book factory from their own factory via our API. I have, today, three customers, and i… — plan on the item: A) three passes on this VM (15 min, need names); B) bearer token + CLI/Python recipe, fresh session

  > **shelley** · 2026-09-19 12:25 UTC  
  > First think, in plain terms. Two very different things are hiding in this item, and I’d separate them:
  >
  > **A. Portals for your three customers (Monday-showable).** This is *not* new code. ProdCal is already multi-tenant: each customer is a client slug with a portal at /{client}/ and a Factory Pass per project. So “a portal for each of the three” = three passes on this VM, today: Admin → Store → New pass (name, email, project). Each gets /{client}/{project}/factory/ with transmittal, Inspect, builds. One VM, one codebase, always in sync because it *is* the app. Handing it to them = sending the login. I can set all three up in fifteen minutes if you give me names + emails + working titles; you did one earlier today (snitkey). Separate VMs would mean three copies to keep in sync and three sets of licensed fonts — worse, not better.
  >
  > **B. “Hit the book factory from their own factory” — the API.** The Factory Pass API already exists and is documented (docs/API.md, docs/specs/FACTORY-PASS-API-2026-09-03.md): upload a .docx, POST /api/books/{id}/convert, poll, download PDF/EPUB, all against a project the pass owns. What’s missing for an outside program (or an outside agent) to use it:
  >  1. **A token instead of a browser login.** Passes authenticate via a cookie; a program needs `Authorization: Bearer <pass token>`. Small server change; the pass already has a secret.
  >  2. **A one-page “build a book from the command line” recipe**: curl upload → convert → poll → download, plus the same as a 30-line Python script. This is the demo for Monday: “here is a book being built from a terminal, no browser.”
  >  3. **Machine-readable Inspect and build errors** (both exist as JSON already; document them).
  >  4. Later: webhooks when a build finishes; a spec/transmittal PUT so the caller can set trim/title without the form.
  >
  > Plan: fresh session, starts from a copy of the API doc, does 1+2 today (2–3 h), demo = the Python script building Ghosts on project 14. It touches srv/auth for the bearer path and adds docs; it won’t collide with what I’m doing here (transmittal, template, filters). A is a 15-minute task in either session — send me the three names.

  > **shelley** · 2026-09-19 12:42 UTC  
  > B is in progress (subagent): Bearer-token alias, a one-page curl recipe, and a stdlib-only Python script that uploads → builds → downloads. A still needs the three names/emails/working titles from you.

  > **shelley** · 2026-09-19 12:49 UTC  
  > B landed (ba069b6): `Authorization: Bearer <project token>` works on every factory endpoint; docs/API-CLI-RECIPE-2026-09-19.md is the one-page curl recipe; scripts/factory-cli.py (stdlib Python, ~60 lines) does upload → build → poll → download, or inspect. Tested end to end on Ghosts. This is your “their factory calls our factory” demo for Monday: a book built from a terminal, no browser. A (the three passes) still needs names.

- [ ] 5.17 **About the Book Factory** text (was 0.8) — YOU pick A / A′ / B and edit the placeholders (drafts are on this item's notes); then ME: transmittal checkbox → colophon on the copyright page and/or back-matter page, off by default

  > **shelley** · 2026-09-18 17:19 UTC  
  > Drafts are ready for your read (not ticking until you have): docs/brand/ABOUT-THE-BOOK-FACTORY-2026-09-18.md — also in the repo on GitHub. Three lengths: A colophon for the copyright page (78 w), A′ one line after the typesetting credit (26 w), B a back-matter page headed “About the Book Factory” (~490 w). Voice: plain third person, says what happened to the book, no adjectives; the URL is the only sell. Once you pick and mark up the wording I wire it as a transmittal checkbox (off by default): copyright-page colophon, or back-matter page, or both. The one-liner, so you can judge the voice here: “Set from the author's Word file by the jdbb studio book factory, print and EPUB from one source, no hand-composed pages. jdbbs.exe.xyz/factory”

  > **shelley** · 2026-09-18 17:59 UTC  
  > Sorry — I left these in a repo file, which you can’t open from this page. Here they are in full. Three lengths; the numbers are true of every build today (thirteen styles, ~9 s).
  > A. Colophon (copyright page, small type) — 78 words
  >  This book was made in the jdbb studio book factory. The author wrote it in Word, in a template of thirteen named styles generated from a one-page transmittal; the factory read the file, reported what it found, and set the print interior and the EPUB from that single source, in licensed type, with no hand-composed pages. Corrections are rebuilt the same way, so every copy of this edition is set from the same file. jdbbs.exe.xyz/factory
  >
  > A′. Colophon, shortest (one line after the typesetting credit) — 26 words
  >  Set from the author's Word file by the jdbb studio book factory, print and EPUB from one source, no hand-composed pages. jdbbs.exe.xyz/factory
  >  ---
  >
  > B. Back-matter page — "About the Book Factory" — 486 words
  >  **About the Book Factory**
  >  This book was not typeset in the usual sense. No one placed its pages by hand. It was built by the jdbb studio book factory, a small machine for turning a Word file into a book, and this page says how, because the method is part of what you are holding.
  >  The author wrote in Word, in a template the factory generated for this title. The template has thirteen named paragraph styles and nothing else: Normal, First Paragraph, three levels of heading, Block Quote, Epigraph, Verse, Code Block, Section Break, Copyright, Signature and Glossary Entry. Every paragraph in the manuscript carries one of those names. That is the whole contract between writer and factory. There is no software to learn and nothing to install; the work of authorship stays in the tool the author already used.
  >  Before the writing began, the author filled in a transmittal: a short form that records what the book is and how it should be set. The page size. The front matter — a dedication, an epigraph, a foreword — and its order. The copyright page, which the factory composes from the transmittal rather than asking anyone to type it. Special characters, mathematics, custom styles the book needs. The transmittal is the specification; the template is generated from it; the finished book follows it.
  >  When a draft was ready, the author uploaded it and the factory inspected it: a machine read of the file that reports which styles were used, where the file departs from the template, how many images it found and how large each will print, where the chapters begin. The report is the factory's reply. The author fixes what it flags and uploads again, for as long as it takes.
  >  Then the build. From that one Word file the factory produces two things at once: a print-ready interior PDF at the trim size on the transmittal, and an EPUB with the author's cover embedded. The interior is set in the studio's house design — a typographic system of margins, type sizes, running heads and spacing worked out once and applied to every book that passes through — in fonts the studio licenses for print. The two editions come from the same source, in the same pass, so they cannot drift apart.
  >  This matters after publication. A book set by hand is finished when the typesetter stops; a correction means opening the files again and hoping the lines still fall where they did. A book from the factory is set from its source every time. Fix the Word file, rebuild, and every page is composed afresh, in seconds. The edition you hold is one build of a file that can be built again.
  >  The factory is the work of jdbb studio, a one-person book studio that published books the ordinary way for thirty years before deciding that the ordinary way was mostly waiting. It is offered to anyone with a manuscript and a few hours. If you have written something and would like it to become a book like this one, the factory is at **jdbbs.exe.xyz/factory**.
  >  ---
  >  #
  > Notes for Jenna
  >  - "thirty years" and "one-person" in the last paragraph are placeholders for whatever is true; cut the sentence if you'd rather not do biography here. - "thirteen named paragraph styles" is true as of today (Signature and Glossary Entry added 2026-09-18). If the count changes, this line and the colophon change with it. - "in seconds" — a full build of a 250-page book is ~9 s on the VM; true. - The page is ~490 words; at 6 × 9 it fills one recto with air. If you want it on one page in Small trim, cut paragraph 4 (Inspect) to two sentences. - Suggested placement: last page of back matter, after About the Author, so the book ends on the author and then the maker.

- [~] 5.5 P3 **H&J / composition quality** in the print PDF — Typst 0.12 levers checked, nothing to change before Monday; scorer in `typesetting/scripts/compscore.py`; runts need Typst 0.13 `costs` (Wed 24+); still to do: compare against an InDesign-set page (YOU supply)

  > **shelley** · 2026-09-18 14:11 UTC  
  > Scorer landed: `python3 typesetting/scripts/compscore.py book.pdf --list` counts loose lines, runts, widows/orphans, hyphen runs, stacks and rivers per book. Ghosts today: 27 loose, 42 runts, 0 widows/orphans, ~90 rivers over 2395 lines.
  >
  > The lever that actually fixes runts (Typst’s `costs`) needs Typst 0.13; we run 0.12. I tried 0.13 in scratch: runts 42 → 29 (−31%), and the template compiles after a one-line fix (landed, no output change on 0.12). But 0.13 also changes spacing around lists and on the Contents page, which shifts lines on ~30 pages of Ghosts — a page-by-page check on two books before it can go live. Recommendation: do the upgrade Wed 24 or after, not before attendees build. Full note: docs/reviews/P3-COMPOSITION-2026-09-18.md. Still want one InDesign-set page from you to run the same scorer on. Side-find: Ghosts ch.1 has 26 stray `>` from a quoted email in the .docx — Inspect should flag that (small, will add).

  > **shelley** · 2026-09-18 17:37 UTC  
  > Checked what else 0.12 can do, so 5.5 is not just waiting on the upgrade. Answer: nothing worth touching before Monday. Ghosts’ 27 “loose” lines are all 0.50–0.58 em — a hair over the yardstick, 1.1 per 100 lines, no hyphen stacks, no widows/orphans. That is clean composition; a compositor would leave it. Obliquities (same scorer, your existing PDF, no build spent) is looser: 3.5 per 100 lines, median 0.53 em, max 0.79, because the protocolized trim gives a ~3.4 in measure at the same 10 pt. The honest fix there is design — 9.5 pt or a hair less margin — a choice for the series spec, not a factory default to flip two days out. The 6 “orphans” it reports are numbered-list first lines (scorer sees the hanging number as a new column) — false positives, I’ll tighten the scorer. Runts remain the only real defect and that is the 0.13 `costs` lever, Wed 24+. Leaving 5.5 in progress for the InDesign page whenever you have one; otherwise it is done for the workshop.

  > **jenna** · 2026-09-19 15:16 UTC  
  > good notes here -- review these notes and compare with the sampler you're making for us? None of our existing projects are "real" yet so zero cost to us if text shifts from Typst 0.13

  > **shelley** · 2026-09-19 15:17 UTC  
  > Agreed — with no real projects yet the reflow is free, so 0.13 goes in today (5.20) and 5.5 gets its lever. Plan when I get there (after 0.17): swap to 0.13.1; add the runt/widow `costs` setting to the series template (0.13-only syntax, so it lands with the swap); rebuild the sampler with 0.13 and run compscore on both versions of the sampler pages + Ghosts so you can see the before/after numbers on the same pages; sampler PNGs/PDF on the transmittal get regenerated from the 0.13 build so what customers see is what the factory now sets. I'll post the scores here.

- [ ] 5.12 Jev (typesafe.ai) pilot — heading classification behind a flag, per `docs/reviews/TYPESAFE-REVIEW-2026-09-18.md` (agreed 2026-09-18; post-workshop)
- [~] 5.13 Index as a factory add-on — listed as coming soon on /factory; phase 1 (Typst index piece, LLM drafting, anchoring) running in parallel on branch `index-addon` (brief `scratch/briefs/index-addon-2026-09-19.md`); phase 2 UI after 0.17 — per `docs/reviews/INDEX-ADDON-FEASIBILITY-2026-09-18.md`

  > **jenna** · 2026-09-19 15:13 UTC  
  > 1/ add Index as a coming-soon add-on 2/ let's start it today: can you draft a prompt I can pass to it's own session or you give to a subagent so that it can proceed in parallel safely?

  > **shelley** · 2026-09-19 15:18 UTC  
  > 1/ Done — /factory lists 'Index · coming soon' in the Add-ons table (live now; price 'tbd'). 2/ Started: brief at scratch/briefs/index-addon-2026-09-19.md, subagent `index-addon` (conv cYBS6TS) running in parallel. Safety: it works in a git worktree on branch `index-addon` (/home/exedev/prodcal-index), never touches main, the live service, the DB or the API; no factory UI (that's phase 2, mine, after 0.17); develops on Ghosts only — it will ask here before touching Obliquities. Deliverables: Typst index piece (flagged), LLM drafting pass via the VM's keyless LLM gateway, marker anchoring, a design note docs/reviews/INDEX-ADDON-DESIGN-2026-09-19.md, and a report. I'll post its summary here when it lands. If you'd rather drive it from your own session instead, the brief file is the prompt — paste it in.

- [ ] 5.19 Tuned per-image grey conversion + proof sheet as a ~$100 add-on (from 5.15; post-workshop)
- [ ] 5.9 4 vCPU bump — only if the workshop shows build queueing (load test: CPU-bound)
- [ ] 5.21 Plantin MT Pro **Bold** missing on the VM (see `docs/IDEAS.md`) — YOU supply the OTF; never committed

## 6 · Ideas log (`docs/IDEAS.md`) — BOTH decide, then queue or kill

- [ ] 6.1 Per-client interactive stylesheet instances (accept/reject seeded from the universal fiction/nonfiction stylesheet)

  > **jenna** · 2026-09-18 00:19 UTC  
  > Seems like this shouldn't be too bad to do since we trimmed down our universal fiction/non-fiction style sheets. Do you agree?

  > **shelley** · 2026-09-18 00:33 UTC  
  > Agree — the trimmed universal sheets make this ~half a day: a per-project table of accept/reject rows seeded from the universal sheet + the existing /stylesheet UI pointed at it. Queue post-workshop (5.10?).

  > **jenna** · 2026-09-18 01:08 UTC  
  > I'm going to make a big deal about style sheets in both senses (for the workshop and in my presentation to the symposium), and disambiguating style sheets (like in Microsoft Word) and the style sheets we mean here (lke whether to use the Oxford comma or not) feels worth doing now and handling exposing the UX in the client dashboard. In the workshop, I think I'd encourage them to accept all of our defaults for efficiency, but also to allow them to add or edit. Am I crazy?

- [ ] 6.2 Reusable "review table" pattern from the sitemap-review tool (JSON table + per-row reply box)

  > **jenna** · 2026-09-18 00:19 UTC  
  > I don't know enough about what you mean for this one.

  > **shelley** · 2026-09-18 00:33 UTC  
  > Means: the sitemap-review tool we used earlier (table of items, a reply box on each row — like this punch list) could be turned into a reusable component for future review rounds. My call: kill until a second concrete review round needs it.

  > **jenna** · 2026-09-18 01:06 UTC  
  > Hmm feels like, through the course of the workshop, I might very quickly end up with a bunch of small borks/nits/qols that we could fix in near real time, and a clean punch list to keep track of them instead of throwing things at you in a session would feel much safer. Maybe we make a blank page shaped like this that has a starter data entry field that I could use to kick off things as they come up on Monday and Tuesday.

- [ ] 6.3 Client-visible on-disk documents (`pi-client/{slug}/` + `serveClientDoc`) — trigger: first real client doc

  > **jenna** · 2026-09-18 00:20 UTC  
  > I don't know enough about what you mean for this one.

  > **shelley** · 2026-09-18 00:33 UTC  
  > Means: like /exedeck or /workshop, but per client — drop an HTML file in pi-client/{slug}/ and it shows only to that client, behind their portal login. Nothing needs it yet; keep on the ideas list until a real client doc exists.

  > **jenna** · 2026-09-18 01:04 UTC  
  > Hmmm we'll have a bunch of clients once the workshop participants start Monday, so this is something that needs to be ready before the workshop, unless I'm still not quite understanding what you mean here.

- [ ] 6.5 Print cover / spine calc — separate product question (cover is EPUB-only today)

  > **jenna** · 2026-09-18 00:22 UTC  
  > Yeah, let's not lose track of this, but the whole "how to get it printed" question is not built into the factory at all yet. I think this is a good one to keep on the idea list for now. I just did some good research, and the two print companies we could point people towards are Bookmobile and Accutrack. Add those to our research if we add this to the factory after the workshop.

  > **shelley** · 2026-09-18 00:33 UTC  
  > Added to docs/IDEAS.md with Bookmobile + Accutrack named. Stays on the ideas list.

- [~] 6.6 Strike the done rows (Admin doc editor) and move the kept ones to §5 with a date — ME, once 6.1–6.5 are decided

  > **jenna** · 2026-09-18 00:22 UTC  
  > I don't understand what this one means.

  > **shelley** · 2026-09-18 00:33 UTC  
  > Means: housekeeping of docs/IDEAS.md itself — strike rows that are done (did the Admin doc editor one tonight) and move the ones we keep into §5 with a date. It is just “tidy the list once 6.1–6.5 are decided”; I will do it when you have answered 6.1–6.4.

  > **jenna** · 2026-09-18 01:09 UTC  
  > Ah gotcha, yes let's keep Punch lists are up to date, and let's also figure out a way to keep a master list of punch lists. Maybe even add these links to our admin dashboard so that I can go back and double-check what we decided, in case I see something wrong or need to change my mind.

  > **jenna** · 2026-09-18 12:13 UTC  
  > didn't see a reply to this admin improvement tracker idea?

  > **shelley** · 2026-09-18 12:13 UTC  
  > Sorry, missed it. Yes. Plan (added as 5.10, Saturday, ~1 h): (1) the export of each punch list to docs/runs/PUNCHLIST-YYYY-MM-DD.md includes every note thread under its item, so the decisions travel with the list (today’s copies only have the ticks); (2) docs/runs/README.md is the master index; (3) a “Runs” card on the admin dashboard → /admin/runs/ that renders those files from disk, read-only, so you can re-read what we decided from any device and change your mind on the live list. Admin-only and read-only, so safe to land before the freeze.

  > **shelley** · 2026-09-18 14:03 UTC  
  > Done row struck (Admin doc editor → a Done section at the foot of docs/IDEAS.md). Already-decided ideas moved to §5 with dates: 5.7 magic-link (was there), 5.12 Jev pilot (new). 6.1–6.5 are still yours to keep or kill — tick or note each and I’ll move the keepers to §5 and drop the rest from IDEAS.

