<!-- exported 2026-09-19 14:51 UTC by scripts/punchlist-export.py; source scratch/run/CHECKLIST.md + notes.json -->

# Punch list 2 · Fri 18 Sep → workshop (Mon 21 / Tue 22) → talk (Wed 23)

Legend: **YOU** = Jenna's turn · **ME** = Shelley's turn · **BOTH** = look together.
Ticked = done · ◐ = in progress. Click a box to cycle ☐ → ☑ → ◐. **note** opens a reply box under any item (⌘↵ saves) — I read those back. Bottom bar adds a new item to the Inbox (now at the top). Refreshes every 15 s (pauses while you type).

Previous list (35 of 55 ticked, incl. 5.14 front matter, 5.15 images, 5.16 `@`) archived at `docs/runs/PUNCHLIST-2026-09-18.md` → readable at `/admin/runs/`. Item numbers carry over so the note threads stay attached. Checkpoint: `checkpoint-2026-09-18-pre-workshop-v2`.

**Freeze:** Sat 19 – Sun 20 fix / test / clean allowed (build → `go test ./srv/` → restart → one factory smoke as `pinstitute`). Mon 21 – Wed 23 hotfix-only. Store flips to live Stripe by timer **Wed 23 Sep 00:00 HKT**.

## 0 · Inbox — new items, untriaged (Shelley moves them into a section)
- [x] 0.1 new: let's comment out the calendar page on the client dashboard for each book. I don't think we're going to end up using that at all.  ·  _added Fri 18:20 UTC_

  > **shelley** · 2026-09-18 18:22 UTC  
  > Done and live. Commented out (not deleted) in the three places a customer met it: the top nav strip (now Your books · Transmittal · Factory), the Calendar link on each book card on the portal, and the small Calendar link in the transmittal header. The page itself still answers at /{client}/{project}/ so you can reach it from the admin side, and the client nav still says CALENDAR nowhere. Hard-reload once; theme.js is cached for up to a minute. Left as is for now: the “N/N tasks done · 0 %” progress line on each card, which is calendar data — say if you want that gone too.

- [x] 0.2 Question: can you look at what a ms inspection costs in tokens? Would be good to know if trivial or not  ·  _added Fri 18:22 UTC_

  > **shelley** · 2026-09-18 18:23 UTC  
  > Trivial — in fact zero. Inspect does not call a language model at all. It is a Python script (typesetting/scripts/detect-edge-cases.py, python-docx + ImageMagick) that reads the .docx and pattern-matches: styles used vs the template, images and their print sizes, chapter starts, special characters, odd formatting. Ghosts (24,000 words, 159 findings) takes 3.3 s and about 36 MB of RAM on the VM; the cost is a few CPU-seconds. Same for the build itself: pandoc → Typst, no model in the loop. The only things in the factory that touch a model are the ones I run by hand in a session (a review, the Obliquities read-through) and the two parked ideas that would add one — the Jev heading-classification pilot (5.12) and the LLM-drafted index (5.13). If either of those ships, a per-manuscript token cost appears for the first time; for a 24k-word book that would be roughly 35k input tokens per pass, i.e. cents, not dollars. Worth stating on the factory page, so I’ll add a line to Inspect’s blurb: “Inspect is a rules-based read of your file; no AI reads your manuscript.”

  > **jenna** · 2026-09-19 13:06 UTC  
  > 1/ For “Inspect is a rules-based read of your file; no AI reads your manuscript.” prefer “Inspect is a rules-based read of your file; no LLM reads your manuscript.” 2/ do we use the abbreviation AI anywhere else? I have a strong preference for LLM, so I guess let's go catch that if it shows up somewhere else. 3/ And let's consider actually saying "Python script" since it's a factory. Showing the working bits is kind of what we're doing. wdyt?

  > **shelley** · 2026-09-19 13:09 UTC  
  > Done, live: “It is a Python script checking the file against a list of rules; no LLM reads your manuscript.” 2/ Searched everything customers see (factory, transmittal, terms, /factory, /press, /workshop, emails): that was the only “AI”. The public field-guide page says “AI coding agent” about Shelley on the VM — that's about the agent, not the product; left it, say if you want LLM there too. 3/ Agree on “Python script” — it's literally true (typesetting/scripts/detect-edge-cases.py, 1,600 lines of rules), and naming the working bit is the whole pitch. Same move is available elsewhere if you like it: “Typst” and “pandoc” are already named on the Factory page.

- [x] 0.3 Google Docs authors have no custom paragraph styles — adopt `[[style name]]` … `[[/style name]]` markers (Jenna trialled `[[style:computer text]]` in the Obliquities rollup); Inspect: “Marked styles” section + one row per unusual font with meaning (Cardo paste flag was noise, not an error)  ·  _from chat Fri 18:40 UTC_

  > **shelley** · 2026-09-18 18:30 UTC  
  > Looked at the file. Google Docs gives you no paragraph styles beyond Normal/Heading/Title, so every one of your 681 paragraphs arrives as “normal” and the intent lives only in direct formatting (Consolas for the TCP dialogue) and your markers. Your instinct is right and I want to make it the official convention rather than a workaround:
  >
  > MARKERS. A paragraph that begins with [[name]] takes the factory style of that name — [[code block]], [[block quote]], [[epigraph]], [[verse]], [[first paragraph]], [[section break]], or any custom style you declared on the transmittal. On its own it applies to that paragraph only; to cover a run, put [[/name]] (or [[end]]) at the end of the last paragraph — so your TCP joke is [[code block]] on the first line and [[/code block]] on the last. I’ll also accept your [[style:name]] spelling, and a few natural aliases (computer text / terminal / code → Code Block; quote / extract → Block Quote). Case and spaces don’t matter. The factory strips the markers on the way in and sets the real style before pandoc sees the file, so the Word template path and the Google Docs path produce the same book. Nobody has to hunt for anything in Word afterwards.
  >
  > INSPECT gets two changes. (1) A new “Marked styles” section: every marker found, how many paragraphs it covers, and whether the name resolved — [[computer text]] → Code Block (alias), [[caption]] → unknown, with the hint to declare it or pick a factory style. (2) “Unusual fonts” stops being 11 red items. One row per font, with the paragraph range and a meaning: “Consolas, 9 paragraphs (¶118–126), monospace — computer text? mark it [[code block]] … [[/code block]]”, and “Cardo, one run in ¶71 — looks like a paste; harmless, the factory sets everything in the book face.” Low severity unless it looks like meaning.
  >
  > Building it now — pre-pass script + Inspect changes, then a test on this file (free slot). About an hour. I’ll say when it’s live so your afternoon run can use [[code block]] … [[/code block]] on the TCP joke. Until then [[style:computer text]] will be printed literally, so hold the build.

  > **shelley** · 2026-09-18 18:47 UTC  
  > Live (216dfeb). Rules as built:
  > • Start a paragraph with [[code block]] — or [[style:code block]], case/spaces don’t matter. That paragraph takes the style.
  > • To cover a run, end the LAST paragraph with [[/code block]] (or [[end]]). So for the TCP joke: keep your [[style:computer text]] on the first line and add [[/computer text]] at the end of “…Hello, would you like to hear a TCP joke?”. Without the closer only the first line is code — Inspect will tell you exactly that.
  > • Names: the 13 factory styles, your transmittal custom styles, and aliases — computer text / code / terminal → Code Block; quote / extract → Block Quote; poem → Verse.
  > • Recognised markers are stripped and the real style applied before pandoc, for print and EPUB alike. Unrecognised ones ([[caption]]) stay in the text so you see them in the proof, and Inspect flags them in a new “Marked styles” section. Caption isn’t a factory style — pandoc already treats the paragraph after an image as its caption, so you can just drop that marker.
  > • Unusual fonts: now one row per font with a meaning. On your file: “Consolas, 9 paragraphs ¶118–126, monospace — computer text? mark it…” (medium) and “Cardo, one run ¶71 — looks like a paste, harmless” (low). Fonts are ignored by the factory either way; that Cardo phrase in ¶71 really is Cardo in the file, so the old report wasn’t wrong, just unhelpful.
  > Ran your rollup through a free slot: no “[[style” left in the PDF, TCP line set in the mono face, EPUB matches. The rule is also on the factory page under Upload (“Drafting in Google Docs?”). Go ahead with the build.

- [x] 0.4 New: Ms Transmittal has indented front matter items that shouldn't be, let's fix  ·  _added Fri 19:10 UTC_

  > **shelley** · 2026-09-18 19:11 UTC  
  > Fixed and live. Epigraph, List of Figures/Tables, Foreword, Preface and Acknowledgments were carrying an “indent” flag from the old Chicago-style transmittal (sub-items under Dedication/Contents). Gone from the defaults, and the page now ignores the flag on existing transmittals too, so every front-matter row is flush left. Reload once.

- [x] 0.5 new: on ms transmittal, delete "Email me a copy"  ·  _added Fri 19:13 UTC_

  > **shelley** · 2026-09-18 19:14 UTC  
  > Gone for customers (History · Duplicate · Print · Word template remain). You still see “Email” on the admin side. Live; reload.

- [x] 0.6 New: in the work session we just had, both of the people I was teaching to use the factory had trouble with Word. Neither of them has a current license, and in one case, we were finally barely able to make the Claude Word work. In the other case, we tried using Pages, but couldn't get it to honor Word's style sheets. I don't know, maybe we could have fought with it some more. That person also tried giving the template and the exported Google Doc in a docx format to Claude to say, "Just do it for me," and that document also failed. We were able to upload it, but it didn't pass the inspection. It could be inspected, but it didn't pass the build, so something inside of it was just still a mess.  See if you can find out some resources I can give our participants in the workshop on Monday for how to use Word better, or how to get access. Surely, this is going to be a real pain point for a lot of people, and we should think about it.  ·  _added Fri 20:06 UTC_

  > **shelley** · 2026-09-18 20:12 UTC  
  > First, the failed build: I pulled the two files that failed this afternoon (books 22–25 on Obliquities) and the problem was ours, not the documents. Two pipeline bugs, both fixed and live:
  >
  > 1. The Foreword ended with a Signature-styled paragraph; the body-start hook got dropped *inside* that block, and Typst refuses a page break inside a container. (Your Obliquities_1_rollup…templated.docx — rebuilt fine just now, 77 pp, credit refunded on the failures so net one build.)
  > 2. An email address inside a block quote (production@protocol-institute.org) was read by Typst as a label reference. Quotes are now rendered properly — that also fixes italics being lost inside quotes, which was a quiet bug.
  >
  > Also: failed builds now keep their working files on the server so I can diagnose in a minute instead of guessing.
  >
  > So “Claude, just do it for me” actually produced a usable file. Resources for the Word-access question coming in the next note.

  > **shelley** · 2026-09-18 20:13 UTC  
  > Resources for Monday — “I don’t have Word.” Four honest routes, best first:
  >
  > 1. **Word for the web (free).** office.com → sign in with any free Microsoft account → upload the template .docx to OneDrive and open it. The Home tab’s Styles gallery shows our template styles (I checked: the template flags them as quick styles, which is exactly what Word online lists). Apply styles, download as .docx. Limits: no custom-style *creation*, no macros — neither matters for us. This is the one to put on the slide. Microsoft’s own how-to: support.microsoft.com → “Use styles in Word for the web”.
  >
  > 2. **LibreOffice Writer (free, offline, Mac/Windows/Linux).** libreoffice.org. Opens the template with its styles intact; Styles sidebar is F11 (or View → Styles); File → Save As → Word 2007–365 (.docx). Honors Word paragraph styles far better than Pages. Skip Pages for this — it renames styles on export and that’s what bit us today.
  >
  > 3. **Google Docs + style markers (free, shipped this afternoon).** Write with Docs’ Heading 1/2/3 as usual, and where a paragraph needs a template style type a marker at its start: [[quote]], [[code]], [[verse]], [[caption]]… ([[/quote]] at the end of a run). File → Download → .docx, upload to the factory; Inspect lists them under “Marked styles” and the build applies them. No Word at all. The one-line explainer is already on the factory page under “Drafting in Google Docs?”.
  >
  > 4. **Pay once, briefly.** Microsoft 365 Personal has a one-month free trial (cancel before it bills); Word on iPad is free on screens under 10.1″; anyone with a .edu address usually has Microsoft 365 free through their school. Buying Word outright (Office Home 2024, ~$150) is the no-subscription option.
  >
  > And the “Claude, just do it” route: it worked — the file was fine, the build bug was ours (see note above). Worth saying out loud on Monday: “hand the template + your Google Doc export to Claude, ask it to move the text into the template’s styles” is a legitimate path now, with Inspect as the check.
  >
  > Suggestion: I can add a short “No Word? Three ways in” box to the factory page and the workshop page with exactly the above — say yes and it’s twenty minutes.

  > **shelley** · 2026-09-18 20:14 UTC  
  > One more from the rebuilt PDF: chapter titles were hyphenating (“Intelli-gence Media”, “Have Your Factory Call My Fac-tory”). Headings never hyphenate now; live for your next build.

  > **shelley** · 2026-09-19 12:26 UTC  
  > Ticking this one: build bugs fixed, resources posted. The offered “No Word? Three ways in” box on the factory/workshop pages is a yes/no from you — if yes I’ll queue it as its own item.

- [x] 0.7 For failed builds, let's think about some improvements to this message. We saw it yesterday during our testing, and it was not very helpful. The inspect report maybe could update to show the actual failure. Open to your other ideas too!  ·  _added Sat 12:17 UTC_

  > **shelley** · 2026-09-19 12:20 UTC  
  > Live. Failed builds now say what actually went wrong, in three layers:
  >
  > 1. Plain explanation + what to do (“A chapter start landed inside a styled block… add an empty Body paragraph before the heading” / “Your file uses a Word style (tweet-p) that isn’t in your template…” / “An image couldn’t be read — re-insert it in Word” / “Your template asks for a font we don’t have” / “We couldn’t read this Word file — re-save as .docx from Word or LibreOffice, especially if it came from Pages”). When it’s our bug, the message says so.
  > 2. **Near: “…the text where it broke…”** — quoted from the document so they can Cmd-F it in Word.
  > 3. A folded “Technical detail” line with the typesetter’s own error, for when they email you.
  >
  > The refund line stays underneath. I rewrote the three failure messages from this afternoon on Obliquities so the page shows the new wording (open the factory and you’ll see it).
  >
  > I put it on the build line rather than in Inspect: Inspect describes the file, the build line describes the build, and that’s where they’re looking when it fails. If you’d rather it also appear at the top of Inspect, say so.
  >
  > One more idea, not built: email you automatically whenever a *customer’s* build fails (message + who + which file), so on Monday you hear about it before they give up. Easy to add — yes/no?

- [~] 0.10 new: https://desertant.com/press/ -- using this page as a model, can you draft a similar page for me in the factory? I also really like the… — DRAFT LIVE at /press (1cbe387); warm bg #FBFAF4 live app-wide; YOU fill the [bracketed] bits (bio, entity, founded, city); assets done

  > **shelley** · 2026-09-19 12:37 UTC  
  > Draft is up: https://jdbbs.exe.xyz/press — same skeleton as Desert Ant’s (lede · About · Facts · Products · Brand assets · The studio), in our chrome. Everything on it is true today except the orange [bracketed] bits, which are yours to fill: one sentence of biography if you want it, legal entity (or delete the line), founded year, city. Their “Independent research” block I left out — we don’t have third-party numbers yet; if you have a quote or two (Bringhurst, a press stat), that’s where they’d go. Wordmark SVG and a 1200×630 social card are the missing assets; I can generate both from the favicon outlines when you want them.
  >
  > The colour: done, and it’s live app-wide. Background went from our cool #FCFDFD to their #FBFAF4, and the hairlines/progress tracks moved to the same warm axis so nothing looks blue against it. Text and the cyan are unchanged. For scale: Anthropic’s cream is #F0EEE6 — ours is now about a fifth of the way there from white, which is the “hint, not tint” you asked for. Hard refresh to see it; if it reads as too much or too little on your screen, say and I nudge one value.

  > **shelley** · 2026-09-19 12:40 UTC  
  > Brand assets are no longer “to follow”: wordmark SVG (paper and ink versions, plus [jdbb] alone) and a 1200×630 social card are generated from the JetBrains Mono outlines and linked on the page; the page also carries the og:image so a pasted link shows the card. Left for you: the four orange placeholders.

- [x] 0.11 new: I did some more testing about the background color, and let's go with this hex for now. #FFFEF6  ·  _added Sat 12:42 UTC_

  > **shelley** · 2026-09-19 12:42 UTC  
  > Done and live: --bg is #FFFEF6 everywhere (theme.css is the single source, so admin, portals, factory, public pages all follow). Hairlines and progress tracks nudged a step lighter to match (#E4E2D8 / #F3F1E8); the press page colour note and the social card now say #FFFEF6 too. Hard refresh to see it.

  > **jenna** · 2026-09-19 14:31 UTC  
  > Sorry, but I still find the background color too dark in the context of the rest of our page. I'm open to suggestions, like maybe we just go to a white page altogether?

  > **shelley** · 2026-09-19 14:33 UTC  
  > Suggestion: go white (#FFFFFF). #FFFEF6 is close enough to white that it reads as “white, slightly off” rather than paper. With a white ground the warmth comes from the field wells, the green hand-off wash and the type. If yes I’d also nudge wells a touch more neutral (#F5F3EC) and hairlines to #E6E4DC so nothing looks yellow against white. Say go and it’s a one-line change, live in a minute.

  > **jenna** · 2026-09-19 14:34 UTC  
  > Sure let's try that now

  > **shelley** · 2026-09-19 14:35 UTC  
  > Live: page is #FFFFFF everywhere (theme.css is the one source — admin, portals, transmittal, factory, public pages). Wells #F5F3EC, hairlines #E6E4DC. Social card, wordmark-on-paper SVG and the press page colour note updated to match. Hard reload; theme.css has no cache-buster so it may take one extra refresh.

- [x] 0.12 new: I got to watch two people work through the transmittal form yesterday, and I saw that they had trouble distinguishi… — COMPS at :8767 (A boxed/B tinted/C paper) — YOU pick

  > **shelley** · 2026-09-19 12:49 UTC  
  > Comps are up (live markup, you can type in them): https://jdbbs.exe.xyz:8767/ — top half. What I think the trouble is: a field today is a bottom rule with mono text on it, in the same mono at nearly the same size as the label above and the help below, so nothing says “type here” and filled values read as printed text. Three directions beside the current one:
  > A · Boxed wells — white well, hairline frame, your typed values in the sans. Plainest “this is an input”; bends our no-boxes rule for inputs only.
  > B · Tinted wells — no frame, a slightly darker warm panel under the rule, white when focused. Gentlest; keeps the ledger look.
  > C · Form-on-paper — heavier grey rule, dashed when empty (like a paper form), values in the sans, labels a shade darker.
  > My lean: B if smallest change, A if the goal is nobody hesitates again. Each is ~10 lines of CSS once you point.

  > **jenna** · 2026-09-19 13:00 UTC  
  > B - but let's make the tint a little warmer than the current tint?

  > **shelley** · 2026-09-19 13:02 UTC  
  > Done, live on the transmittal: fields are now warm wells (#F6EFDD — a step warmer/yellower than the #F3F1E8 tint the comp used), white when you click in, still the bottom rule. Textareas match. Reload with shift if you still see the old rules. Say if you want it warmer still; it's one token (--well-bg).

  > **jenna** · 2026-09-19 14:27 UTC  
  > On https://jdbbs.exe.xyz/pinstitute/book-001/transmittal/ the fields still seem a little dark, but they definitely clash with the new white block up at the top. Take another swing at that?

  > **shelley** · 2026-09-19 14:28 UTC  
  > Second swing, live now (hard reload): the white block was the new “Transmittal is final” panel using the modal-white — theme rules say that colour is for overlays only, so it now sits on a faint green wash of the page paper instead, and the fields went a shade lighter (#F9F4E6). No pure white anywhere on the page now; the panel, wells and paper are all the same family.

- [x] 0.13 new: during the testing session yesterday, I saw people have trouble with the Final button up at the top because you are… — COMPS at :8767 (A strip+finish / B sticky / C hand-off panel; rec A+C) — YOU pick

  > **shelley** · 2026-09-19 12:49 UTC  
  > Same comps page, bottom half: https://jdbbs.exe.xyz:8767/ . Two problems, three directions that combine:
  > A · Step strip + finish block — put the Factory’s 1–5 strip on the transmittal (step 1 current), and end the form with a strong rule, one sentence, the same Mark Final button, and “Continue to the Factory →”. The page ends with an action, not a field.
  > B · Sticky status bar — a one-line bar pinned to the bottom of the window (status · saved · Mark Final), which after Final becomes “✓ Final · Word template ↓ · Next: upload your manuscript →”. Always in reach; costs ~40 px.
  > C · Hand-off panel on Final — clicking Final scrolls to the top and shows a green-ruled panel: what just happened, then 1) download the template 2) go to the Factory. Fixes the “now what?” but not the tall page on its own.
  > Recommendation: A + C. No pinned chrome, and the transmittal reads as step one of five, which is what it is. Building A+C is maybe an hour; B another hour. Say which.

  > **jenna** · 2026-09-19 13:02 UTC  
  > Agree. A + C

  > **shelley** · 2026-09-19 13:09 UTC  
  > Built and live (A + C). On the transmittal now: the factory's 1–5 strip under the header (1 Transmittal current; ✓ once final), and the form ends with a finish block — rule, one sentence, Mark Final + “Continue to the Factory →”. On Final the page scrolls up and a green-ruled hand-off panel replaces the intro: 1) download the Word template 2) go to the Factory, with a filled Continue button; the finish block flips to “Switch to Draft · Continue to the Factory”. Hidden in print. Check on Ghosts: https://jdbbs.exe.xyz/prot/zoo/transmittal/

  > **jenna** · 2026-09-19 14:29 UTC  
  > 1/ "Switch to Draft" Should maybe read "Return to draft mode"?
  > 2/ what does it say when they are still in draft mode? Maybe that line just isn't there at all?

  > **shelley** · 2026-09-19 14:30 UTC  
  > 1/ Done — “Return to draft mode” at the bottom, “Return to draft” in the header (short slot). Copy in both sentences says “return to draft mode” too.
  > 2/ In draft mode there was no “Switch to Draft” line, but there was a “Continue to the Factory →” link beside Mark Final. Gone now: while in draft the bottom block is just “That’s the whole form. Mark it final and the factory generates your Word template…” + the Mark Final button. The Factory link only appears once it’s final. Live (hard reload).

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

- [x] 0.15 new: on https://jdbbs.exe.xyz/factory let's reduce the size of the price by about half.  ·  _added Sat 13:14 UTC_

  > **shelley** · 2026-09-19 13:21 UTC  
  > Done: $549 on /factory is now 22 px (was 44), same weight and mono; “per manuscript · USD” unchanged beside it. Reload to see.

- [x] 0.16 new: can we edit the standing punch list pages so that I can paste screenshots in like I do in a normal session?  ·  _added Sat 13:14 UTC_

  > **shelley** · 2026-09-19 13:21 UTC  
  > Done. Paste a screenshot into any note box (or the New item bar) — it uploads on paste, shows as a thumbnail, and saves with the note. I get it in chat as a file I can open, and the export copies it into the archive at /admin/runs/. The red square is my test paste.
  >
  > ![screenshot](img/20260919-132039-645.png)

- [~] 0.17 new: the page width shift from  1 Transmittal to the rest of the flow (2 Upload 3 Inspect 4 Build 5 Download) is quite jarring the first page is standalone, and then the rest are stacked on one much narrower page. Take a look, and again, without actually changing anything, show me comps for a couple of ideas for how we can pull these parts of the process back together?  ·  _added Sat 14:37 UTC_

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


## 2 · Your side — YOU

- [x] 2.6 Unlock the `pinstitute` portal and eyeball the new cards (Transmittal · Factory → · Calendar, `[FACTORY PASS]`)

  > **shelley** · 2026-09-18 17:24 UTC  
  > Pre-checked as pinstitute just now (signed in with a real one-time link, minus the email): portal shows both projects with the Transmittal · Factory → · Calendar cards and [FACTORY PASS] on Obliquities; perception/transmittal loads with the new Format + What-you-get sections; book-001/factory shows 2 of 3 builds left, and Inspect ran free in 1.5 s (1156 findings on the test rollup, book map printed, no build spent). Nothing broke. Your turn is only the eyeball.

  > **jenna** · 2026-09-19 14:40 UTC  
  > done, though somewhat superceded now by our ongoing work. fine to mark 2.6 complete

- [x] 2.7 Re-download the Word template on a final transmittal, confirm in Word — Styles pane shows the **13** factory styles, fonts Georgia/Arial/Courier New, page = trim (server side pre-checked; the Word look is yours)

  > **shelley** · 2026-09-18 17:24 UTC  
  > Pre-checked the file itself: the template downloaded from book-001/factory has 13 qFormat styles (Normal, Heading 1–3, First Paragraph, Block Quote, Code Block, Section Break, Verse, Copyright, Epigraph, Signature, Glossary Entry), fonts Georgia / Arial / Courier New, page = the protocolized trim (4.91 × 7.59 in). Item text says 11 — read it as 13. What only Word can confirm is the Styles pane.

  > **shelley** · 2026-09-18 18:07 UTC  
  > Pre-checked the plumbing: the template download correctly refuses until the transmittal is marked final (409, ‘fill in the transmittal and mark it final first’). The Word-side look (Styles pane, fonts, page size) is yours — it is thirteen styles now, not eleven; wording fixed on the new list.

  > **jenna** · 2026-09-19 14:40 UTC  
  > done

- [ ] 2.8 Today's book with colour images: run Inspect, tell me what it says about the images and how the print PDF's grey looks (5.15 landed; Ghosts had nothing to test on)

## 3 · P2 — talk deck from Venkat's "Have your factory call my factory" — BOTH

Source: https://protocolized.summerofprotocols.com/p/have-your-factory-call-my-factory · deck is `factory-talk.html`; never edit `/2026-pi-symposium/talk`. Talk: Wed 23 Sep 17:30 UTC, SIGP4B.

- [ ] 3.2 YOU — cut / reorder / add beats
- [ ] 3.4 BOTH — run-through in the browser; timing
- [ ] 3.5 ME — push; link in the handoff

## 4 · Workshop watch (Mon 21 / Tue 22) — BOTH

- [ ] 4.1 `/admin/factory/` (Floor) on a second screen; Shelley session open with `factory-tail`
- [ ] 4.2 Attendees free via hand-granted passes (New pass form) — no coupons, no $0 checkouts (form + attendee login verified Fri, 2.5)
- [ ] 4.3 Watch for `build failed`, `401` bursts (password confusion), builds "queued" (> 2 at once)
- [ ] 4.4 Tue night: confirm `systemctl list-timers prodcal-store-live.timer` still armed; Wed 00:00 HKT check the go-live email arrived and `/api/public/store/config` says live

## 5 · Queued — ME unless marked

- [x] 0.8 New: the default typography for our print PDF is pretty bad. Can you review our typography.md and perhaps look at the Bringhurst references and go ahead and bui… — part 1 (trim-derived defaults) LIVE d187fa5; part 2 = transmittal typography choices (typeface pairing · text size · section break · paragraph style), IN PROGRESS (subagent typo-choices, conv cDXESK6)

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

- [x] 5.18 New pass → fresh transmittal starts with empty title/author although the pass form asked for both; prefill `book.title` / `book.author` from the pass (found Fri during 2.5)

  > **shelley** · 2026-09-18 18:10 UTC  
  > Fixed and live (all three pass paths: admin form, coupon, store). The transmittal draft is now seeded with the title and author at the moment the pass is created, so an attendee opens it Monday and sees their book named. Verified with a throwaway pass, then revoked/archived it.

- [x] 5.22 **Creative Commons copyright notices** (from 0.4) — transmittal copyright section: a “Rights” choice: All rights reserved (default) · CC BY · CC BY-SA · CC BY-NC · CC BY-NC-SA · CC BY-ND · CC BY-NC-ND · CC0 / public domain; generated copyright page prints the proper notice (“This work is licensed under a Creative Commons Attribution 4.0 International License. To view a copy, visit creativecommons.org/licenses/by/4.0/”) instead of “All rights reserved”, EPUB `dc:rights` matches. Template: `series-template.typ` l.604. Post-workshop unless a CC book shows up first.

  > **shelley** · 2026-09-18 18:59 UTC  
  > Queued as 5.22. Today the copyright page hard-codes “All rights reserved” (series-template.typ l.604); the fix is a Rights choice on the transmittal’s copyright section (All rights reserved · the six CC licences · CC0) that swaps in the proper CC notice on the generated page and sets the EPUB’s dc:rights to match. Half a day. Post-workshop unless a CC book turns up first — say if the Obliquities rollup is one.

  > **shelley** · 2026-09-19 13:21 UTC  
  > Built (50187d3), ahead of schedule because it was cheap once I was in the copyright page: a “Rights” select in the transmittal's Copyright page section — All rights reserved (default) · CC BY · BY-SA · BY-NC · BY-NC-SA · BY-ND · BY-NC-ND · CC0. The © line becomes e.g. “Copyright © 2026 Name. This work is licensed under a Creative Commons Attribution 4.0 International License. To view a copy of this license, visit https://creativecommons.org/licenses/by/4.0/” (CC's recommended wording), CC0 gets the dedication sentence. Print PDF, Word template and the EPUB's dc:rights all take it from one place (srv/rights.go). Preview updates live. Existing transmittals keep All rights reserved.

- [~] 5.5 P3 **H&J / composition quality** in the print PDF — Typst 0.12 levers checked, nothing to change before Monday; scorer in `typesetting/scripts/compscore.py`; runts need Typst 0.13 `costs` (Wed 24+); still to do: compare against an InDesign-set page (YOU supply)

  > **shelley** · 2026-09-18 14:11 UTC  
  > Scorer landed: `python3 typesetting/scripts/compscore.py book.pdf --list` counts loose lines, runts, widows/orphans, hyphen runs, stacks and rivers per book. Ghosts today: 27 loose, 42 runts, 0 widows/orphans, ~90 rivers over 2395 lines.
  >
  > The lever that actually fixes runts (Typst’s `costs`) needs Typst 0.13; we run 0.12. I tried 0.13 in scratch: runts 42 → 29 (−31%), and the template compiles after a one-line fix (landed, no output change on 0.12). But 0.13 also changes spacing around lists and on the Contents page, which shifts lines on ~30 pages of Ghosts — a page-by-page check on two books before it can go live. Recommendation: do the upgrade Wed 24 or after, not before attendees build. Full note: docs/reviews/P3-COMPOSITION-2026-09-18.md. Still want one InDesign-set page from you to run the same scorer on. Side-find: Ghosts ch.1 has 26 stray `>` from a quoted email in the .docx — Inspect should flag that (small, will add).

  > **shelley** · 2026-09-18 17:37 UTC  
  > Checked what else 0.12 can do, so 5.5 is not just waiting on the upgrade. Answer: nothing worth touching before Monday. Ghosts’ 27 “loose” lines are all 0.50–0.58 em — a hair over the yardstick, 1.1 per 100 lines, no hyphen stacks, no widows/orphans. That is clean composition; a compositor would leave it. Obliquities (same scorer, your existing PDF, no build spent) is looser: 3.5 per 100 lines, median 0.53 em, max 0.79, because the protocolized trim gives a ~3.4 in measure at the same 10 pt. The honest fix there is design — 9.5 pt or a hair less margin — a choice for the series spec, not a factory default to flip two days out. The 6 “orphans” it reports are numbered-list first lines (scorer sees the hanging number as a new column) — false positives, I’ll tighten the scorer. Runts remain the only real defect and that is the 0.13 `costs` lever, Wed 24+. Leaving 5.5 in progress for the InDesign page whenever you have one; otherwise it is done for the workshop.

- [ ] 5.12 Jev (typesafe.ai) pilot — heading classification behind a flag, per `docs/reviews/TYPESAFE-REVIEW-2026-09-18.md` (agreed 2026-09-18; post-workshop)
- [ ] 5.13 Index as a factory add-on (LLM-drafted conceptual index, Typst-resolved locators) — per `docs/reviews/INDEX-ADDON-FEASIBILITY-2026-09-18.md`; ~5 days; post-workshop
- [ ] 5.19 Tuned per-image grey conversion + proof sheet as a ~$100 add-on (from 5.15; post-workshop)
- [ ] 5.20 Typst 0.13 upgrade on the VM (Wed 24+; unlocks `par(costs:)` for 5.5)
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

