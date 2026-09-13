# Content review against better-documents — 2026-09-13

Standard: `docs/reference/better-documents/`. Corpus: 16 plain-text extractions (`scratch/content-review/*.txt`, gitignored) of every public and client-facing page plus the three emails a registrant receives (registration confirm, Factory Pass, build delivered). Five parallel reviewers, one group each, plus a short sixth pass on the customer factory page; findings in the skill's *location → problem → fix → severity* format, voice preserved, nothing invented.

**120 findings raw, 9 void after source check** (marked ✗ below): the `Company (leave blank)` field is a CSS-hidden honeypot (B-F8, B-F13); the two plain "Availability / What you are buying" labels are deliberate `.tag` kickers (B-F7); the `#` links on field-notes are an intentional non-link and a scroll handler (C-F11); and B-F10–F14 compared `/factory` with itself — the "in-app factory" extraction was the public page served twice, so the duplication they report does not exist. The actual customer page is group F. Confirmed against source: the `/stylesheet/` date is `new Date()` at render, not a revision date (E-F8).

**19 selected** (★) for `/admin/content-review/` (accept / edit / reject), weighted to CRITICAL/MAJOR in passes 1–2 on the pages registrants will read before Sep 21. The deck (`/exedeck`) is a past talk; its 18 findings stay here for its next revision.

| Group | Files | Findings |
|---|---|---|
| A | workshop path: /workshop, /2026-pi-symposium roster, three emails | 32 |
| B | /factory (public), /litmags | 22 |
| C | /field-notes, homepage, client portal gate, project page | 24 |
| D | /exedeck slide deck | 18 |
| E | /stylesheet/, transmittal page | 23 |
| F | customer factory page /{client}/{project}/factory/ (added after the first pass missed it) | 3 |

---

## Group A — workshop path: /workshop, /2026-pi-symposium roster, three emails

### A-F1 · workshop.txt · MAJOR · pass 1
LOCATION: para 1 (intro paragraph under the H1)
CURRENT: "a transmittal *handshake* that generates a custom authoring template, a conformance *preflight* that negotiates messy human input into machine-readable structure, and a deterministic transform"
PROBLEM: Missing context / insular opening. The audience is writers, not designers, and "transmittal" (the noun that session 1 and the Factory Pass email hinge on) is never defined anywhere on the page; the reassurance that no background is needed only arrives a section later.
PROPOSED: "a transmittal *handshake* — the transmittal is the one-page spec your book is built from — that generates a custom authoring template, a conformance *preflight* that negotiates messy human input into machine-readable structure, and a deterministic transform"

### A-F2 ★ · workshop.txt · MAJOR · pass 1
LOCATION: para 1 / dek ("Run your manuscript through a book-production factory · with Jenna Dixon, jdbb studio")
CURRENT: "Run your manuscript through a book-production factory · with Jenna Dixon, jdbb studio"
PROBLEM: Buried ask. That there are 8 curated seats, that you request one with the form below, and that confirmation comes by email is stated only in the last section; a skimming reader gets 60 lines of description before learning what the page wants from them.
PROPOSED: ADD: one sentence directly under the dek stating the ask — 8 seats, curated, request one with the form below, confirmed by email within a few days (all facts already in "Request a seat").

### A-F3 ★ · workshop.txt · MAJOR · pass 1
LOCATION: "One hard requirement"
CURRENT: "You must bring **real material** — a manuscript or text collection you'd like to turn into a book. Any size, rough is welcome."
PROBLEM: Missing context that changes what the reader must do. The Factory Pass email later says "Upload your Word manuscript", but nothing on this page or in the confirmation email says the factory takes a Word file; the roster shows registrants arriving with Obsidian .md notes and a Google Doc.
PROPOSED: ADD: state the input format the factory accepts (Word/.docx per the Factory Pass email) and whether other formats need converting first, in this section — described here, not written, because the exact accepted formats aren't stated in any of the five files.

### A-F4 · workshop.txt · MINOR · pass 1
LOCATION: "The four sessions", para 1
CURRENT: "Cumulative — each session builds on the last, so please plan to attend all four. All times UTC."
PROBLEM: Missing context. The page never says the sessions are online, or on Discord; the reader infers the venue only from "Where we send the Discord invite" in a form hint. The roster later states it plainly ("Sessions run on Discord").
PROPOSED: "Cumulative — each session builds on the last, so please plan to attend all four. Live on Discord; all times UTC."

### A-F5 · workshop.txt · MINOR · pass 1
LOCATION: "The four sessions", note below the table
CURRENT: "In Asia these land in the evening"
PROBLEM: Audience honesty. 15:00–18:30 UTC is 23:00–02:30 in Singapore/Tokyo and 20:30–00:00 in India; two of nine registrants are in Asia and "evening" understates what they are signing up for.
PROPOSED: "In Asia these land late evening to past midnight"

### A-F6 · workshop.txt · MINOR · pass 1
LOCATION: "Request a seat" form, checkbox hint
CURRENT: "I'm new to protocol studies. (Used only as a tiebreaker if we're oversubscribed — we widen the room on purpose.)"
PROBLEM: Ambiguous ask. The reader can't tell whether ticking this helps or hurts their chances; "we widen the room" is studio shorthand, not the reader's language.
PROPOSED: "I'm new to protocol studies. (If we're oversubscribed, newcomers get the tiebreak — we widen the room on purpose.)"

### A-F7 · workshop.txt · MINOR · pass 1
LOCATION: "The four sessions", table row 2
CURRENT: "Preflight — run messy input through the conformance detector; Keep / Strip / Convert"
PROBLEM: Unexplained triad. "Keep / Strip / Convert" is presented as if the reader already knows these are the three rulings you make on each flagged element.
PROPOSED: "Preflight — run messy input through the conformance detector; rule Keep / Strip / Convert on each flag"

### A-F8 · workshop.txt · MINOR · pass 2
LOCATION: "You'll leave with" (section)
CURRENT: "## You'll leave with"
PROBLEM: Order reflects effort, not importance. The outcomes list is what a prospective registrant weighs before caring about UTC times, but it sits after the schedule table and just above the form.
PROPOSED: MOVE to directly after "Who it's for" (before "The four sessions").

### A-F9 · workshop.txt · MINOR · pass 3
LOCATION: "One hard requirement"
CURRENT: "You must bring **real material** — a manuscript or text collection you'd like to turn into a book. Any size, rough is welcome. This is non-negotiable: the workshop runs on **your** files, not samples."
PROBLEM: Emphasis overload: two bolds plus "non-negotiable" plus a bolded sentence in the paragraph immediately above, in a four-line stretch. When everything is emphasised nothing is.
PROPOSED: "You must bring **real material** — a manuscript or text collection you'd like to turn into a book. Any size, rough is welcome. This is non-negotiable: the workshop runs on your files, not samples."

### A-F10 · workshop.txt · MINOR · pass 3
LOCATION: "You'll leave with", bullets
CURRENT: "A typeset **PDF + EPUB** of your own material" … "A reusable **custom authoring template**" … "A transferable **mental model**" … "A **decision rubric**" … "First-hand experience **delegating creative work to an LLM agent**"
PROBLEM: A bold phrase in every one of five bullets is decoration, not emphasis; the bullets are already one line each and skimmable.
PROPOSED: Remove the bold from all five bullets (text unchanged).

### A-F11 · workshop.txt · MINOR · pass 4
LOCATION: "Request a seat" form, "Your background" options
CURRENT: "Publishing / editorial / Design / Writing / Other"
PROBLEM: List items don't cover the stated audience. "Who it's for" names protocol-curious people and the brief names protocol researchers, yet the only option for them is "Other" — three of nine registrants on the roster ended up as OTHER, which defeats the curation the hint promises.
PROPOSED: ADD: an option for the protocol-studies / research audience (label to match whatever the registration API accepts).

### A-F12 · workshop.txt · MINOR · pass 5
LOCATION: "The four sessions", table header
CURRENT: "Local"
PROBLEM: Ambiguous label. The column shows ET / PT / CET, not the reader's local time, and the note below says UTC is canonical.
PROPOSED: "ET · PT · CET"

### A-F13 · workshop.txt · MINOR · pass 5
LOCATION: "Thanks — your request is in.", para 2
CURRENT: "we'll offer you the session recordings and a spot in the next round."
PROBLEM: Inconsistent scope of "recordings": the schedule note says only "the lecture-style portions of sessions 1 and 3" are recorded, but this line (and the form hint "Discord invite, calendar invites, and recordings", and the confirmation email) imply full session recordings.
PROPOSED: "we'll offer you the recorded portions of sessions 1 and 3 and a spot in the next round."

### A-F14 · workshop.txt · MINOR · pass 5
LOCATION: page title (line 1)
CURRENT: "Register — Protocolize Your Book · [jdbb] studio"
PROBLEM: Naming inconsistency. The page deliberately frames the form as a curated seat *request* ("Request a seat", "your request is in", "confirm your spot") but the browser/tab title says "Register", which promises more than the form delivers.
PROPOSED: "Request a seat — Protocolize Your Book · [jdbb] studio"

### A-F15 ★ · app-cohort-roster.txt · MAJOR · pass 1
LOCATION: participant 6
CURRENT: "Mike Check … A smoke-test manuscript: three chapters of placeholder prose used to exercise every step of the factory. … Verify every page and flow the cohort will touch."
PROBLEM: Wrong audience. A QA persona is listed as a participant on a page real registrants read to find people to pair with; it also makes the summary count 9 participants for an 8-seat workshop, which registrants will notice.
PROPOSED: DELETE (exclude the smoke-test persona from the cohort-facing roster before registrants receive their Factory Pass, or mark it as a test row).

### A-F16 · app-cohort-roster.txt · MINOR · pass 2
LOCATION: para under the summary stats ("Sessions run on Discord. …")
CURRENT: "read your neighbours' and you'll see where to pair up."
PROBLEM: Unannounced ordering. Rows are numbered 1–9 and sorted alphabetically by first name, but nothing says so; a numbered list reads as rank or seat order.
PROPOSED: "read your neighbours' and you'll see where to pair up. Listed alphabetically."

### A-F17 · app-cohort-roster.txt · MINOR · pass 4
LOCATION: summary stats
CURRENT: "8 / ALL FOUR SESSIONS"
PROBLEM: Unsummarised data. "8 all four sessions" doesn't say what the number is — eight who ticked "I can attend all four sessions" on the form — and every row's session boxes are unticked, so the figure can't be reconstructed from the rows.
PROPOSED: "8 / CAN ATTEND ALL FOUR"

### A-F18 · app-cohort-roster.txt · MINOR · pass 4
LOCATION: participant rows (each card)
CURRENT: "EU / BOOK / WRITING" (and "US / OTHER / OTHER")
PROBLEM: Data without a header. The three tags per row are region, source-material type and background, but no column label says so; "OTHER OTHER" is unreadable cold.
PROPOSED: ADD: a header row or per-tag label (Region · Material · Background) above the list.

### A-F19 · app-cohort-roster.txt · MINOR · pass 4
LOCATION: participant rows, "SESSIONS" block
CURRENT: "SESSIONS ☐ S1 ☐ S2 ☐ S3 ☐ S4"
PROBLEM: Unexplained element and orphaned labels. The reader can't tell whether the boxes record intent, attendance or something to tick on the printed roster, and S1–S4 are named nowhere on this page (the workshop page calls them Handshake, Preflight, Build + delegate, Show-and-tell, with times). Registrants will use this page during the two days and it has no schedule.
PROPOSED: ADD: one line saying what the boxes are for, and the four session names with day/time UTC (or a link to the schedule on the workshop page).

### A-F20 · email-registration_confirm.txt · MINOR · pass 3
LOCATION: para 4 ("One reminder")
CURRENT: "the workshop runs on YOUR material"
PROBLEM: ALL-CAPS as emphasis; the sentence already carries the point.
PROPOSED: "the workshop runs on your material, not samples"

### A-F21 · email-registration_confirm.txt · MINOR · pass 5
LOCATION: subject line
CURRENT: "We got your Protocolize Your Book registration"
PROBLEM: Term mismatch with the page the reader just left, which calls this a seat request that may not be confirmed; "registration" in the subject reads as "you're in" one paragraph before the body says the cohort may fill.
PROPOSED: "We got your Protocolize Your Book seat request"

### A-F22 · email-registration_confirm.txt · NIT · pass 4
LOCATION: session list, parenthetical
CURRENT: "(US Eastern 11:00 / 13:00 · US Pacific 08:00 / 10:00 · Central Europe 17:00 / 19:00)"
PROBLEM: The pairs aren't tied to anything; the reader has to work out that the first time is sessions 1 and 3 and the second is 2 and 4.
PROPOSED: "(sessions 1 & 3 / 2 & 4 — US Eastern 11:00 / 13:00 · US Pacific 08:00 / 10:00 · Central Europe 17:00 / 19:00)"

### A-F23 · email-registration_confirm.txt · NIT · pass 5
LOCATION: session list
CURRENT: "Mon Sept 21" / "Tue Sept 22"
PROBLEM: Abbreviation differs from the workshop page ("Mon Sep 21") and the Factory Pass email ("Sep 21–22"); inconsistency across a sequence read by the same person. (Date itself unchanged.)
PROPOSED: "Mon Sep 21" / "Tue Sep 22"

### A-F24 ★ · email-factory_pass.txt · MAJOR · pass 1
LOCATION: para 1
CURRENT: "Your Factory Pass is live: one manuscript, all the way through the protocol."
PROBLEM: Broken sequence. The previous email promised the next message would "confirm your spot, with the Discord invite, calendar invites for all four sessions, and a short note on prepping your manuscript". This email confirms none of that and introduces "Factory Pass", a term the reader has never seen on the workshop page or in the confirmation email, so they can't tell if they have a seat or what this account is for.
PROPOSED: ADD: one sentence up top confirming the seat and saying this is the factory account you'll use in the workshop; then either the Discord invite / calendar invites / prep note, or a line saying they arrive separately and when.

### A-F25 ★ · email-factory_pass.txt · MAJOR · pass 1
LOCATION: "First steps"
CURRENT: "1. Start with the transmittal: it is the spec your book is built from. 2. Upload your Word manuscript."
PROBLEM: The reader can't tell what to do now versus in the room. The workshop page says session 1 is "fill a real transmittal" together on Mon Sep 21; this list tells them to start with the transmittal today. No deadline or "before session 1, do X" is given.
PROPOSED: ADD: state which of these steps to do before Mon Sep 21 (and by when) and which happen live in session 1.

### A-F26 ★ · email-factory_pass.txt · MAJOR · pass 4
LOCATION: unlabelled bullet list after "First steps" (lines beginning "Included: the workshop sessions themselves")
CURRENT: "- Included: the workshop sessions themselves (Sep 21–22). - After that, email support is not included — the preflight report and the docs are the self-serve path. - Live help is available at USD 100/hr, booked in advance, 30-minute minimum."
PROBLEM: A list whose items don't belong where they sit: three support-terms bullets appear under "First steps" with no heading, and the first item reads as if the Factory Pass includes the workshop rather than the reverse. Bullets expose the drift.
PROPOSED: MOVE under its own heading "Support" placed after "What's included", with the first bullet reworded: "- Live help during the workshop sessions (Sep 21–22) is included."

### A-F27 · email-factory_pass.txt · MINOR · pass 5
LOCATION: "First steps", item 4
CURRENT: "Build. Failed builds don't cost a credit."
PROBLEM: "credit" is a new term; the email has only spoken of "3 builds". The build-delivered email then counts "Builds remaining: 2 of 3".
PROPOSED: "Build. Failed builds don't count against your 3."

### A-F28 · email-factory_pass.txt · MINOR · pass 5
LOCATION: support bullets
CURRENT: "the preflight report and the docs are the self-serve path"
PROBLEM: A reference that doesn't say where it goes. "The docs" has no URL anywhere in the three emails, and the reader has just been told email support isn't included.
PROPOSED: ADD: the URL of the docs after "the docs".

### A-F29 · email-factory_pass.txt · NIT · pass 5
LOCATION: sign-off
CURRENT: "Jenna Dixon - jdbb studio"
PROBLEM: Sign-off punctuation differs from the confirmation email ("Jenna Dixon · jdbb studio") and the build email ("- Jenna / jdbb studio"); three formats in three consecutive emails.
PROPOSED: "Jenna Dixon · jdbb studio"

### A-F30 · email-build_delivered.txt · MINOR · pass 2
LOCATION: last paragraph
CURRENT: "The deliverable is a correctly typeset PDF and EPUB of the manuscript as it conforms to your transmittal. Preflight tells you what doesn't conform."
PROBLEM: The key instruction — what to do if the output looks wrong — is the last line and reads as a disclaimer; the action (fix, rebuild, two builds left) is implied, not stated.
PROPOSED: MOVE above the sign-in line, and: "The deliverable is a correctly typeset PDF and EPUB of the manuscript as it conforms to your transmittal. Preflight tells you what doesn't conform — fix that and rebuild."

### A-F31 · email-build_delivered.txt · MINOR · pass 5
LOCATION: para after "Builds remaining"
CURRENT: "Sign in to your factory with the client password from your welcome email."
PROBLEM: Points to a document by a name it never had — the credentials arrived in "Your Factory Pass: Building in the Wrong Market", which called it "Password", not "client password" — and the factory URL isn't repeated here, so the reader must find the earlier email to act.
PROPOSED: "Sign in to your factory with the password from your Factory Pass email." plus ADD: the factory URL from that email on the next line.

### A-F32 · email-build_delivered.txt · NIT · pass 5
LOCATION: sign-off
CURRENT: "- Jenna
jdbb studio"
PROBLEM: Third sign-off format in the sequence (see F29).
PROPOSED: "Jenna Dixon · jdbb studio"

---

## Group B — /factory (public), /litmags

### B-F1 ★ · factory-public.txt · MAJOR · pass 2
LOCATION: Availability (unheaded paragraph after the Add-ons table)
CURRENT: "Purchasing opens after the September workshop; workshop attendees receive a pass with their seat."
PROBLEM: The single fact that changes what a visitor can do today — you cannot buy this yet — sits below the price and the add-ons table, so a buyer reads the whole offer before learning it's not on sale. Murder-mystery ordering; the ask (write to be notified) is buried.
PROPOSED: MOVE the Availability paragraph to directly under "$149 / per manuscript · USD", ahead of "Everything under “What's included” is in that number." and the Add-ons table.

### B-F2 · factory-public.txt · MAJOR · pass 1
LOCATION: para 1 (intro, before "What's included")
CURRENT: "There is no software to learn: you stay in Word."
PROBLEM: The page serves two readers — prospective buyers and workshop registrants holding a code — but code holders get no signal until the last section that the redeem form exists on this page. The reader with something to do doesn't know what to do.
PROPOSED: ADD: one line after the intro paragraph pointing code holders to the redeem form (a jump link to the "Redeem a pass" section), so registrants don't have to read the sales copy to find it.

### B-F3 · factory-public.txt · MINOR · pass 2
LOCATION: What you are buying, exactly
CURRENT: "Failed builds (the factory erroring out) are not counted against you."
PROBLEM: This is a term of the "3 builds" allowance, not a definition of the deliverable; a reader checking what 3 builds means under "What's included" won't find it three sections later.
PROPOSED: MOVE to the "3 builds" bullet under What's included: "**3 builds.** Each build produces a print PDF *and* an EPUB. Failed builds (the factory erroring out) are not counted against you."

### B-F4 · factory-public.txt · MINOR · pass 5
LOCATION: heading "How a build works"
CURRENT: "How a build works"
PROBLEM: The list covers the whole pass (Transmittal → Upload → Inspect → Build → Download), and "Build" is also the name of step 4 and of a counted allowance. The heading names a part with the name of the whole.
PROPOSED: "How a pass works"

### B-F5 · factory-public.txt · MINOR · pass 5
LOCATION: What's not · first bullet, and heading "Support edges"
CURRENT: "see Support [#support] for what a human costs"
PROBLEM: The link says "Support"; the section it lands on is headed "Support edges". A link should say where it goes.
PROPOSED: "see Support edges [#support] for what a human costs"

### B-F6 · factory-public.txt · MINOR · pass 1
LOCATION: Support edges · last paragraph
CURRENT: "Workshop attendees: the four sessions (Sep 21–22) are the included human time."
PROBLEM: "The four sessions" assumes the reader knows which workshop; the workshop isn't named or linked anywhere in the body, only in the nav.
PROPOSED: "Workshop attendees: the four Protocolize Your Book sessions [/workshop] (Sep 21–22) are the included human time."

### B-F7 ✗ void · factory-public.txt · MINOR · pass 4
LOCATION: Availability; What you are buying, exactly
CURRENT: "Availability" / "What you are buying, exactly"
PROBLEM: Every other section on the page has a heading, but these two labels are plain text, so they drop out of the page's signposting and read as a different kind of thing from their neighbours (inconsistency read as meaning). May be an extraction artefact — verify in the source.
PROPOSED: Mark both as section headings at the same level as "Storage & privacy".

### B-F8 ✗ void · factory-public.txt · MINOR · pass 1
LOCATION: Redeem a pass · form
CURRENT: "Company (leave blank)"
PROBLEM: A visible field whose only instruction is not to use it. If it's a spam trap it should be hidden from users and screen readers; if it's real, the label should say what it's for. As it stands it makes the reader wonder what they're missing.
PROPOSED: DELETE from the visible form (hide the honeypot), or label what the field is actually for.

### B-F9 · factory-public.txt · NIT · pass 4
LOCATION: Support edges · bullets
CURRENT: "**Preflight Review — $75.** One Word file plus its preflight report; one written reply listing what to fix."
PROBLEM: The three human add-ons are described twice on the page (Add-ons table and here) with slightly different wording; two descriptions of one product will drift.
PROPOSED: "If you want a human, the three paid options are in Add-ons [#add-ons] above: Preflight Review, Live help, Studio typesetting." and DELETE the three bullets.

### B-F10 ✗ void · app-factory-page.txt · MAJOR · pass 2
LOCATION: whole page — REDEEM A PASS is the last section
CURRENT: "Have a code from a workshop seat or from the studio? Redeem it here and we'll set up your project."
PROBLEM: The page's only action — the redeem form — comes after ~65 lines of sales copy duplicated verbatim from the public explainer, including a price the registrant has already covered. In-app, the reader has a code and a job to do; the structure fights them.
PROPOSED: MOVE "Redeem a pass" (heading, intro line, form) to directly under the "One manuscript through the studio's book factory" subtitle; keep the intro paragraph; replace everything from WHAT'S INCLUDED through SUPPORT EDGES with a single line linking to the public /factory page for the full terms.

### B-F11 ✗ void · app-factory-page.txt · MAJOR · pass 1
LOCATION: WHAT'S INCLUDED through SUPPORT EDGES
CURRENT: "A pass takes one manuscript through the whole protocol: a transmittal (your spec — trim size, typeface, front matter, the decisions a typesetter would ask you about)"
PROBLEM: The explainer, pricing, add-ons table, storage and support sections are a word-for-word copy of the public page. Two copies of the terms means two places to change when a price or window changes, and the app copy already lacks the public page's Support anchor link. The public page should be the single source; this page should redeem.
PROPOSED: DELETE the duplicated sections (see F10) and point to the public page for terms; leave the short intro paragraph as orientation.

### B-F12 ✗ void · app-factory-page.txt · MINOR · pass 1
LOCATION: AVAILABILITY
CURRENT: "Purchasing opens after the September workshop; workshop attendees receive a pass with their seat. If you want one before then, or want to be told when it opens, write to j@djinna.com."
PROBLEM: Addressed to prospective buyers. A registrant already logged in with a code reads "purchasing opens after the workshop" as "you can't do this yet" — the opposite of the page's purpose.
PROPOSED: DELETE from the in-app page (the public page keeps it).

### B-F13 ✗ void · app-factory-page.txt · MINOR · pass 1
LOCATION: REDEEM A PASS · form
CURRENT: "Company (leave blank)"
PROBLEM: Same as F8: a visible field instructing the user not to fill it. Also the only label on the form not styled like the others, which reads as meaningful. Hide it if it's a honeypot.
PROPOSED: DELETE from the visible form (hide the honeypot), or label what the field is actually for.

### B-F14 ✗ void · app-factory-page.txt · MINOR · pass 4
LOCATION: REDEEM A PASS · below "Redeem pass →"
CURRENT: "Redeem pass →"
PROBLEM: The public page tells the reader what happens after redeeming ("We've emailed your login details. Start with the transmittal — it generates the Word template you'll write in."); the in-app page, where the redemption actually happens, doesn't say what comes next before the button is pressed. Verify the success state isn't simply hidden in the extraction.
PROPOSED: ADD: one line before the button stating the next step after redeeming (project page opens, start with the transmittal), matching the public page's success text.

### B-F15 ★ · litmags.txt · MAJOR · pass 4
LOCATION: At a glance · Web platform / CMS · Output formats · Revenue model
CURRENT: "### Web platform / CMS"
PROBLEM: Three charts titled by category, none by finding. Chart titles should state what the data show ("Most of the 132 run on WordPress" or whatever is true), otherwise readers interpret cold. Extraction shows no captions — verify in the source.
PROPOSED: ADD: a one-line takeaway under each chart heading stating what the distribution shows, drawn from the data.

### B-F16 · litmags.txt · MINOR · pass 5
LOCATION: title tag; H1; kicker; readout; footer nav
CURRENT: "# 132 Literary & SFF Magazines"
PROBLEM: Five names for one page: "Lit-Mag Tool Stack" (title), "132 Literary & SFF Magazines" (H1), "Lit-mag tool stacks" (kicker), "Small-press publishing stacks" (readout), "Lit-mag stack" (footer). The H1 alone doesn't say what the page is about — the stacks, not the magazines. The title tag also says "small magazines" while the page admits corporate and general-interest outliers.
PROPOSED: "# 132 Literary & SFF Magazines — how they publish" and align the title tag's wording to "Lit-mag tool stacks" to match kicker and footer.

### B-F17 · litmags.txt · MINOR · pass 2
LOCATION: The magazines · table
CURRENT: "## The magazines"
PROBLEM: 132 rows merged from two source lists and no statement of how they're ordered (alphabetical? by source? by ranking?). Readers hunt for meaning in unexplained order.
PROPOSED: ADD: one clause under the heading stating the sort order and whether the filters change it.

### B-F18 · litmags.txt · MINOR · pass 4
LOCATION: The magazines · filter chips
CURRENT: "All WordPress Substack Custom build Print edition Audio / podcast SF / fantasy Poetry Weekly Defunct / ceased ⚠ Outliers"
PROBLEM: One row mixing platform, output format, genre, cadence and status. A list whose items don't belong together — the reader can't tell whether picking two chips ANDs or ORs, or which facet each belongs to.
PROPOSED: Group the chips under short facet labels — Platform · Format · Genre · Cadence · Status — matching the table columns.

### B-F19 · litmags.txt · MINOR · pass 5
LOCATION: The magazines · column headers
CURRENT: "Description & house style" / "Editorial voice & notes"
PROBLEM: Two free-text columns whose names overlap (house style vs editorial voice); a reader can't predict which column holds what.
PROPOSED: "What it publishes & house style" / "Editorial voice & notes" — or fold the two into one column if the content overlaps in practice.

### B-F20 · litmags.txt · MINOR · pass 1
LOCATION: subtitle line under H1
CURRENT: "Schwitzgebel's SFF prestige ranking"
PROBLEM: A surname with no first name or role; workshop readers who aren't in SFF or philosophy won't know who this is or why his ranking counts as a source.
PROPOSED: "Eric Schwitzgebel's SFF prestige ranking"

### B-F21 · litmags.txt · MINOR · pass 1
LOCATION: para 1
CURRENT: "This reference ignores submission economics and prestige alike, and documents how each publication *works*"
PROBLEM: Says what the page does but not who it's for or what decision it helps with (a small press choosing a platform? a workshop exercise?). The audience is left to infer.
PROPOSED: ADD: one sentence stating who should use this reference and for what — not stated anywhere on the page.

### B-F22 · litmags.txt · NIT · pass 3
LOCATION: readout line under the nav
CURRENT: "researched Aug 2026 · Desk research, Aug 2026"
PROBLEM: The date and method are stated twice in the same line. Redundant label.
PROPOSED: "132 magazines · desk research, Aug 2026 · volunteer-run mags change often · verify before you rely on it."

---

## Group C — /field-notes, homepage, client portal gate, project page

### C-F1 ★ · field-notes.txt · MAJOR · pass 1
LOCATION: para 1 (intro under the H1)
CURRENT: "Field material for the workshop's lectures, demos, and the moment it clicks that this is everywhere."
PROBLEM: The page is public and linked from the homepage for registrants, but much of it speaks as the presenter's dossier ("the page on the projector", "participants have already made", "safe to cite & show"), so a registrant can't tell whether this is pre-reading for them or the speaker's notes. Audience unclear.
PROPOSED: "Source material for the workshop's lectures and demos — the examples we point at, collected in one place so you can come back to them."

### C-F2 ★ · field-notes.txt · MAJOR · pass 2
LOCATION: para 1 (intro) / section "01 The protocol concepts these make concrete"
CURRENT: "Here are real-world systems where language is already run as a strict protocol: *controlled vocabularies, typed interfaces, conformance regimes, versioned specs*."
PROBLEM: The four featured sections lean on these terms, but the glossary that defines them ("01 The protocol concepts…") arrives after all four, and the running order (marked only by "near pole" / "far pole" side-labels) is never stated — even though the page itself calls Zen Garden, which sits third, "the mechanism underneath everything else on this page". Unannounced ordering; definitions after use.
PROPOSED: ADD one sentence to the intro stating the order: "Four examples, from the everyday to hard mode — the styles dropdown, this site's font picker, CSS Zen Garden (the mechanism under all of them), marine classification review — then the concepts they make concrete, then a shelf of further candidates." Alternatively MOVE section "01" to directly after the intro.

### C-F3 · field-notes.txt · MAJOR · pass 3
LOCATION: section "The controlled vocabulary you already use", MS Word card
CURRENT: "**Dozens** of built-ins — Heading 1–9, Title, Subtitle, Emphasis/Strong, Quote/Intense Quote, Book Title, Caption, List Paragraph… — across **five style types** (paragraph, character, **linked**, table, list), with **inheritance**, a **“style for the following paragraph”** rule, and **unlimited custom styles**."
PROBLEM: Six bolded runs in one sentence; nothing reads as the important part. Emphasis overload (same pattern in the "Too small" card: smuggle / direct formatting / fake heading all bold).
PROPOSED: "Dozens of built-ins — Heading 1–9, Title, Subtitle, Emphasis/Strong, Quote/Intense Quote, Book Title, Caption, List Paragraph… — across five style types (paragraph, character, linked, table, list), with inheritance, a “style for the following paragraph” rule, and **unlimited custom styles**."

### C-F4 · field-notes.txt · MINOR · pass 3
LOCATION: section "The controlled vocabulary you already use", para 1
CURRENT: "*You already mark **what a thing is** and let the tool decide how it looks.*"
PROBLEM: Bold nested inside italic — two emphasis types on one element; recurs in "The resolution" card and in the whole-sentence bold at the end of the Zen Garden section ("**“what a thing is” is stored once, and “how it looks” is resolved separately and swappably.**"). Stacked emphasis.
PROPOSED: "You already mark *what a thing is* and let the tool decide how it looks." (and un-bold the Zen Garden sentence)

### C-F5 · field-notes.txt · MINOR · pass 1
LOCATION: section "The canonical proof — CSS Zen Garden", last para
CURRENT: "a book factory's stylesheet, an ABS rulebook — all of them work because"
PROBLEM: "ABS" is used before the American Bureau of Shipping is introduced in the next section; a writer reading top-to-bottom hits an unexplained acronym. Missing context.
PROPOSED: "a book factory's stylesheet, a ship-classification rulebook — all of them work because"

### C-F6 · field-notes.txt · MINOR · pass 1
LOCATION: section "Conformance on hard mode", isomorphism list; also "Why it matters for a book factory specifically"
CURRENT: "Keep / Strip / Convert (deviation that presupposes the rule)"
PROBLEM: Keep / Strip / Convert and "the real-world **Strip** step" are workshop-internal terms never defined on this page; the parenthetical explains the analogy, not the terms. Assumes knowledge the audience may not have.
PROPOSED: ADD: a short clause (or link to where it is defined) saying what the Keep / Strip / Convert triad is — the three things preflight can do with a non-conforming element — the first time it appears.

### C-F7 · field-notes.txt · MINOR · pass 4
LOCATION: section "02 The starter shelf", "Conformance / linting regimes", first item
CURRENT: "Paris MoU Port State Control ⚙⚙⚙ — public marine twin of the classification findings log; coded, convention-cited deficiencies + an explicit “action taken” status vocabulary."
PROBLEM: Repeats the "Public sources" block at the end of the marine section almost verbatim, twenty lines earlier. Redundancy on an already long page.
PROPOSED: "Paris MoU Port State Control ⚙⚙⚙ — the public twin of the classification findings log; see the featured entry above."

### C-F8 · field-notes.txt · MINOR · pass 4
LOCATION: section "02 The starter shelf", intro line and group headings
CURRENT: "Grouped by the concept each best illustrates."
PROBLEM: The groups don't use the five concepts just defined in "01": there is no "Typed interface" group, "Layered determinism" becomes "Separation of layers", and the third heading bundles three ideas ("Separation of layers · versioning · machine authors"). A list whose grouping doesn't match its own key.
PROPOSED: Rename the groups with the "01" terms (e.g. "Layered determinism · versioned protocols"), file JSON-schema generation under conformance, and either add a typed-interface group or say none is listed.

### C-F9 · field-notes.txt · MINOR · pass 1
LOCATION: section "02 The starter shelf", intro line
CURRENT: "Hardness: ⚙ light · ⚙⚙ real · ⚙⚙⚙ hard mode."
PROBLEM: The key doesn't say what is being rated — hard to implement, hard to read, or how unforgiving the regime is. The marine section defines "hard mode" as the last; the key should say so.
PROPOSED: "Hardness — how unforgiving the regime is: ⚙ light · ⚙⚙ real · ⚙⚙⚙ hard mode."

### C-F10 ★ · field-notes.txt · MINOR · pass 5
LOCATION: kicker above the H1
CURRENT: "Book Factory Workshop · Source material"
PROBLEM: The workshop is called "Protocolize Your Book" everywhere else, including this page's own footer and the homepage link; a registrant may think this is material for a different event.
PROPOSED: "Protocolize Your Book · Source material"

### C-F11 ✗ void · field-notes.txt · MINOR · pass 5
LOCATION: section "The controlled vocabulary you already use", last two paras; section "02", "Controlled vocabularies", first item
CURRENT: "That is **CSS Zen Garden** [#] — separation of layers"
PROBLEM: This link and "Google Docs vs. MS Word styles [#]" both point to "#" — they go nowhere, and the text doesn't say they refer to sections on this page. Links that don't say where they go.
PROPOSED: Point both at the relevant section anchors on this page; text "CSS Zen Garden (below)" and "Google Docs vs. MS Word styles (featured above)".

### C-F12 · field-notes.txt · NIT · pass 1
LOCATION: section "Conformance on hard mode", last block
CURRENT: "**Public sources (safe to cite & show):**"
PROBLEM: Presenter-facing aside on a public page; a registrant doesn't cite or show anything. Author-centred phrasing.
PROPOSED: "**Public sources:**"

### C-F13 · field-notes.txt · NIT · pass 5
LOCATION: browser title / H1 / homepage link
CURRENT: "Language as Protocol — Examples in the Wild · [jdbb] studio"
PROBLEM: Three casings of the same title across the browser tab ("Examples in the Wild"), the H1 ("examples in the wild."), and the homepage link ("Field notes: language as protocol"). Readers read inconsistency as meaning.
PROPOSED: "Language as protocol — examples in the wild · [jdbb] studio"

### C-F14 ★ · app-home.txt · MAJOR · pass 1
LOCATION: hero, para under "Book production, end‑to‑end."
CURRENT: "A working studio for trade-paperback typesetting, manuscript transmittals, and shared production calendars — for authors, editors, and small presses."
PROBLEM: The page names its audiences but the only action on it is for existing clients (the portal). A new author or press is not told how to start, and the Factory Pass — the thing they would buy — isn't mentioned. Buried/absent ask.
PROPOSED: ADD: one sentence after the hero stating how a new author or press starts (what to send, to j@djinna.com which is already on the page) and a link to the Factory Pass.

### C-F15 · app-home.txt · MINOR · pass 4
LOCATION: "PIPELINE READOUT" block
CURRENT: "job/ghosts‑tp‑01  ·  rev 14 · typ 0.13"
PROBLEM: A block of job data (title, word count, "galley v2 out · awaiting author sign‑off") with no caption saying whether it is a live client job or an illustration; a prospective client may read it as someone's confidential project shown publicly. Unsummarised data.
PROPOSED: ADD: a qualifier in the block title stating whether the job is live or a sample.

### C-F16 · app-home.txt · MINOR · pass 1
LOCATION: CAPABILITIES, item 05
CURRENT: "Shared production schedules with scheduled digests."
PROBLEM: "Digests" is undefined for authors and presses and is repeated in the CADENCE line ("shared calendar · scheduled digests") without ever saying what a digest is or who receives it. Missing context.
PROPOSED: "Shared production schedules, plus scheduled digests of what changed."

### C-F17 · app-home.txt · MINOR · pass 4
LOCATION: "RESOURCES" list
CURRENT: "House stylesheet / Field notes: language as protocol / Workshop: Protocolize Your Book / Lit-mag tool stack"
PROBLEM: Four links for three different readers (a client document, workshop material, a tool list) under one heading with no signal of which is for whom. A list whose items don't belong together.
PROPOSED: ADD: split into "For clients" and "Workshop" sub-labels, or add a three-word qualifier to each item.

### C-F18 · app-home.txt · MINOR · pass 5
LOCATION: "RESOURCES" list, item 3
CURRENT: "Workshop: Protocolize Your Book"
PROBLEM: An event link with no date; the reader can't tell if it is upcoming or past.
PROPOSED: "Workshop: Protocolize Your Book · Sep 21–22"

### C-F19 · app-home.txt · NIT · pass 5
LOCATION: "RESOURCES" list, item 4
CURRENT: "Lit-mag tool stack"
PROBLEM: The same page is called "Lit-mag stack" in this page's footer and in the field-notes footer. Naming inconsistency.
PROPOSED: "Lit-mag stack"

### C-F20 ★ · app-client-portal.txt · MAJOR · pass 1
LOCATION: heading block
CURRENT: "Password required"
PROBLEM: The homepage told the client to "Enter your client code"; this page asks for a "Password", and the client name "Mike Check" sits between heading and button with no label saying whose portal this is or what to type. Either the credential is named two ways or there are two credentials and neither page says so.
PROPOSED: ADD: one line under the heading naming the portal ("Mike Check") and saying which credential is wanted here and how it relates to the client code from the welcome email.

### C-F21 · app-client-portal.txt · MINOR · pass 5
LOCATION: link under the UNLOCK button
CURRENT: "Forgot or need a reset?"
PROBLEM: Doesn't say what happens when clicked — a form, an email, a phone call. The homepage already gives the answer (ask j@djinna.com); this page should too.
PROPOSED: "Forgot it? Ask j@djinna.com for a reset."

### C-F22 · app-client-portal.txt · NIT · pass 5
LOCATION: link under the UNLOCK button
CURRENT: "← Back"
PROBLEM: Destination unnamed.
PROPOSED: "← Back to studio home"

### C-F23 · app-project.txt · MINOR · pass 1
LOCATION: empty state under the TIMELINE tab
CURRENT: "No tasks yet"
PROBLEM: The only prose on the page tells the client nothing about what happens next or whether they must do anything for tasks to appear. Missing next step.
PROPOSED: ADD: one clause saying who adds tasks and what triggers it (e.g. the studio, once the transmittal is in).

### C-F24 · app-project.txt · NIT · pass 5
LOCATION: back link above the project title
CURRENT: "← MCHECK"
PROBLEM: Shows the client slug in caps rather than the client's name ("Mike Check" on the gate page); a client may not recognise their own code as a place to go back to.
PROPOSED: "← Mike Check"

---

## Group D — /exedeck slide deck

### D-F1 · exedeck.txt · MAJOR · pass 2
LOCATION: slide 03 / 10 "auth as a header", para beginning "There is a third door"
CURRENT: "**There is a third door, and it faces inward.** The machine can ask what it is."
PROBLEM: Slide 03 already carries two ideas (inbound, outbound) under a title that promises exactly two obligations; a third door plus "Two honest limits" makes four blocks on one slide and quietly amends the deck's own "two doors" framing from slide 01. One idea per slide; the title no longer describes the slide.
PROPOSED: MOVE the "third door" paragraph to its own slide between 03 and 04 (kicker: "the door that faces inward"), keeping "Two honest limits" on 03 as the close of the two-obligations argument. Adjust the "10 slides" count and slide numbers accordingly.

### D-F2 · exedeck.txt · MAJOR · pass 2
LOCATION: slide 07 / 10 "case three · shared machines", sub-block "Where the industry is heading"
CURRENT: "This morning another vendor announced letting a teammate join a running agent session, hand off a task, and recover work after a crash."
PROBLEM: Slide 07 is the longest slide and stacks three ideas — the share mechanism, the firm's "trust comes from review" conclusion, and an industry-trend aside — so the takeaway in the title is diluted by the time the reader reaches "Watch this area." The trend paragraph does not support the slide's claim about sharing a machine.
PROPOSED: MOVE the "Where the industry is heading" block to slide 10 as a short aside after the fourth conclusion ("A machine is a place…"), where it extends a stated conclusion instead of interrupting a case.

### D-F3 · exedeck.txt · MINOR · pass 2
LOCATION: slide 08 / 10 "how the work happens", final para
CURRENT: "The same header contract makes the application portable. Offline, the identical unmodified server runs behind a 127.0.0.1 proxy that injects the same X-ExeDev-UserID header the cloud proxy does."
PROBLEM: This is a second idea on a slide whose title is "The agent runs on the machine, not on my laptop." Portability of the header contract belongs with the header mechanism (slide 03), not with agent practice; a reader looking for it later will not find it under "how the work happens".
PROPOSED: MOVE to slide 03, after "My authorization logic is one header check, and I did not write an authentication system."

### D-F4 · exedeck.txt · MINOR · pass 5
LOCATION: slide 01 / 10 "summary", para 2
CURRENT: "**a fleet of meeting-recording bots**"
PROBLEM: Slide 06 describes one bot for one recurring study group, and slide 10 calls it "A recorder for one recurring call." "Fleet" sets up a different case than the one delivered; readers will hunt for the other bots.
PROPOSED: "**a meeting-recording bot for a recurring call**"

### D-F5 · exedeck.txt · MINOR · pass 1
LOCATION: slide 02 / 10 "what it is", para 3
CURRENT: "One user processed 40 TB of public LIDAR data across 100 machines at once"
PROBLEM: Two sentences earlier the allowance is "spread across up to 50 machines". A reader who was just told the cap is 50 cannot reconcile 100 without context that is not on the slide.
PROPOSED: ADD: one clause explaining how 100 concurrent machines squares with the 50-machine figure (a different tier, a raised cap, or that 50 is the default) — or drop the machine count and keep the 40 TB.

### D-F6 · exedeck.txt · MINOR · pass 5
LOCATION: slide 05 / 10 "why a whole machine", title
CURRENT: "## The pipeline is five separate programs."
PROBLEM: The table beneath has six rows, and "licensed fonts" is not a program; readers will count and lose confidence in the number. The title also states a fact rather than the slide's takeaway, which is in the last paragraph ("A real machine is … the shortest accurate description of what the work requires").
PROPOSED: "## Five programs, licensed fonts, one disk: this needs a real machine."

### D-F7 · exedeck.txt · MINOR · pass 2
LOCATION: slide 04 / 10 "case one · document pipeline", para after the four steps
CURRENT: "**Preflight is the step relevant to this group.**"
PROBLEM: The sentence that tells a protocols audience which of the four steps to read for arrives after they have read all four. Front-load: say which step matters before listing the steps.
PROPOSED: MOVE to directly under the title, before "An editor completes a **transmittal**", so the four-step list is read with step 02 already flagged.

### D-F8 · exedeck.txt · MINOR · pass 1
LOCATION: slide 04 / 10, sub-block "An ossified standard, up close" (also slide 10, conclusion 1: "the sufficiency argument from the QWERTY session")
CURRENT: "This is the argument from the water-rate session: standardizing at the extraction step stalls, but lowering the cost of compliance is tractable."
PROBLEM: "The water-rate session" and "the QWERTY session" are references to other talks at the same event. A public web reader (the deck is at a public path and linked from the workshop pages) has no way to know what those sessions argued; the point survives only because the colon restates it.
PROPOSED: "This is the argument from the water-rate session earlier in this SIG: standardizing at the extraction step stalls, but lowering the cost of compliance is tractable." (and on slide 10: "the sufficiency argument from the QWERTY session earlier in this SIG"). ADD: links or one-line pointers to those sessions if they are published.

### D-F9 · exedeck.txt · MINOR · pass 2
LOCATION: slide 06 / 10 "case two · the recorder", title
CURRENT: "## A voice call becomes a transcript and a recap."
PROBLEM: The title describes what the bot does; the slide's point for this talk is the sentence buried at the end of para 2 — "Here exe.dev is the trusted intermediary rather than the host." Title should state the takeaway.
PROPOSED: "## The recorder lives elsewhere; exe.dev holds only the relay."

### D-F10 · exedeck.txt · MINOR · pass 2
LOCATION: slide 10 / 10 "conclusions", sub-block "The caveat"
CURRENT: "That gap deserves attention more than anything else in the talk."
PROBLEM: The thing the author says matters most appears only in the last paragraph of the last slide, and the summary on slide 01 does not mention it. A reader who stops after the summary — the stated purpose of a summary slide — misses it.
PROPOSED: ADD: one sentence at the end of slide 01 para 3 naming the caveat (the approach lowers the cost of building, not the cost of knowing a Unix machine), so the summary previews the qualification as well as the claim.

### D-F11 · exedeck.txt · MINOR · pass 4
LOCATION: slide 08 / 10 "how the work happens", data block "The press system, measured"
CURRENT: "The press system, measuredFeb 23 → Aug 24, 2026"
PROBLEM: The block has a title and a date range but no line stating what the numbers show or why they are on this slide; the reader has to infer that 226 commits / 16,522 lines / 71 sessions are evidence for "the agent runs on the machine".
PROPOSED: ADD: a one-line caption under the block title stating what the six figures demonstrate (what one operator plus an on-machine agent produced in six months), so the table is read as evidence rather than inventory.

### D-F12 · exedeck.txt · MINOR · pass 3
LOCATION: slide 09 / 10 "tempo", para 3
CURRENT: "**The benefit is that problems I reported have quietly stopped existing.** **The cost is that documentation ages faster than it can be written, and advice has a short shelf life.**"
PROBLEM: Two consecutive full sentences in bold; when both halves of an "even-handed" comparison are emphasised, neither is. The title already names the cost as the point.
PROPOSED: "The benefit is that problems I reported have quietly stopped existing. **The cost is that documentation ages faster than it can be written, and advice has a short shelf life.**"

### D-F13 · exedeck.txt · MINOR · pass 2
LOCATION: slide 02 / 10 "what it is", title
CURRENT: "## A VM, a disk that persists, and a proxy."
PROBLEM: The title is an inventory; the slide's takeaway is in para 3 — the pooled allowance makes a machine "a cheap unit of organization", one per concern. Title should carry that.
PROPOSED: "## A VM, a persistent disk, a proxy — and one machine per concern."

### D-F14 · exedeck.txt · NIT · pass 5
LOCATION: slide 07 / 10, sub-block "Where the industry is heading"
CURRENT: "This morning another vendor announced"
PROBLEM: "This morning" is spoken-word deixis on a public page dated only "August 2026"; a September reader cannot place it.
PROPOSED: "On the day of this talk another vendor announced"

### D-F15 · exedeck.txt · NIT · pass 5
LOCATION: slide 09 / 10 "tempo", para 1 and the version readout
CURRENT: "The agent it shipped with was 68 days old." … "→ 304 increments in 67 days"
PROBLEM: Two adjacent figures for the same interval (68 vs 67 days); the reader will assume one is wrong.
PROPOSED: "The agent it shipped with was 67 days old." (or make the readout "68 days" — whichever matches Jun 17 → the day the VM was opened; keep both the same)

### D-F16 · exedeck.txt · NIT · pass 5
LOCATION: slide 09 / 10 "tempo", title
CURRENT: "## The tools are changing weekly, and that is a real cost."
PROBLEM: The body measures "roughly 4.5 per day" and spends half the slide on the benefit and the mitigation; "weekly" undersells the data and "cost" states half the slide.
PROPOSED: "## The tools change daily. That is a benefit and a real cost."

### D-F17 · exedeck.txt · NIT · pass 3
LOCATION: slide 01 / 10 "summary", para 2
CURRENT: "**a production system for a small press**, **a fleet of meeting-recording bots**, and **shared machines at the naval architecture firm I work for**"
PROBLEM: All three items of the list are bold, so the bold marks nothing; the sentence is already the list.
PROPOSED: "a production system for a small press, a meeting-recording bot for a recurring call, and shared machines at the naval architecture firm I work for"

### D-F18 · exedeck.txt · NIT · pass 1
LOCATION: slide 03 / 10, "third door" para
CURRENT: "It is normally attached to every machine by default"
PROBLEM: "Normally" and "by default" hedge the same claim twice, and leave the reader unsure whether the endpoint can be absent.
PROPOSED: "It is attached to every machine by default"

---

## Group E — /stylesheet/, transmittal page

### E-F1 ★ · app-stylesheet.txt · MAJOR · pass 1
LOCATION: intro para (under "House editorial stylesheet")
CURRENT: "What we copyedit to unless a book's own stylesheet says otherwise. Mechanics apply everywhere; voice and dialogue rules are for fiction; the apparatus section is for nonfiction."
PROBLEM: Written from the studio's side ("what we copyedit to"). Clients and registrants are sent here to prepare a Word manuscript, but the page never says which sections they should apply themselves and which describe what the studio will do — the ask is missing (audience & purpose).
PROPOSED: "What we copyedit to unless a book's own stylesheet says otherwise. If you're preparing a manuscript for us, follow §1 and §5; §2 is what we do with the file once it arrives. Mechanics apply everywhere; voice and dialogue rules are for fiction; the notes-and-citations section (§6) is for nonfiction."

### E-F2 ★ · app-stylesheet.txt · MAJOR · pass 1
LOCATION: 2. Manuscript handling, paras 1–4
CURRENT: "Request manuscripts as .docx, not plain text or pasted text." / "Edit in a lossless round-trip (.docx → markdown → .docx or equivalent)." / "Deliver a clean edited file plus a changelog … The changelog is the review surface"
PROBLEM: This section is addressed to the copyeditor ("request", "edit", "deliver"), not to the author reading it. An author has to invert every sentence to find the one instruction that applies to them (send .docx) and the one thing they'll receive (a changelog to review). Audience mismatch.
PROPOSED: Para 1: "Send manuscripts as .docx, not plain text or pasted text. Word files preserve italics, smart quotes, em dashes, and scene breaks; plain-text paste loses them." Para 4: "You get back a clean edited file plus a changelog (before → after list and author queries). Review the changelog, not tracked changes — they don't reliably survive conversion." Paras 2–3 are internal process; keep, but open them with "On our side:" or move them below the author-facing paragraphs.

### E-F3 · app-stylesheet.txt · MINOR · pass 2
LOCATION: 8. What not to change
CURRENT: "Authorial fragments and run-ons used stylistically." / "Slang, vulgarity, and dialect in dialogue." / "Invented terminology and neologisms (style them per the book's word list)."
PROBLEM: Three of the five items restate the "Keep:" list in §3 (fragments, profanity/slang, coined words) with slightly different wording ("do not soften" vs. nothing; "narration" dropped). Two lists for one rule invite drift and leave the reader unsure which is authoritative (structure; bullets as diagnostic).
PROPOSED: MOVE the three fiction items into the §3 "Keep:" list (or replace them in §8 with one line: "Everything under §3 Keep."), leaving §8 with the two items that are genuinely new: comma splices and an author's consistent, defensible departure from house style.

### E-F4 · app-stylesheet.txt · MINOR · pass 4
LOCATION: 7. Recurring mechanical fixes
CURRENT: "Frequent misspellings: angrier not \"angier\"; desiccated not \"dessicated\"; putrefying not \"putrifying\"; hall pass two words." / "Missing prepositions (\"step up to the corner\")." / "Dropped/duplicated words: delete false starts (\"I was the I couldn't…\")."
PROBLEM: The list mixes house-level rules (hyphenated compound modifiers, its/it's) with one manuscript's specific typos and sentences. On a page presented as the studio default, the reader can't tell which items are policy and which are leftovers from a project sheet (list whose items don't belong together; imported content needs a pass).
PROPOSED: ADD: either state on the page where the examples come from (one clause after "Watch-list of corrections that recur across manuscripts:") or replace the manuscript-specific examples (angier, hall pass, step up to the corner, "I was the I couldn't") with generic ones. Same applies to §1 examples "color, fueled, desiccated" and §5 "Quiet Harbor: Calm Waters for Your Feed; Amber Mornings".

### E-F5 · app-stylesheet.txt · MINOR · pass 1
LOCATION: 6. Nonfiction apparatus → Heading hierarchy
CURRENT: "Levels must nest (no A-head jumping to C-head); a level needs at least two siblings or none"
PROBLEM: "A-head / C-head" is typesetter jargon; the authors and small presses this sheet is for mostly won't know it (missing context).
PROPOSED: "Levels must nest (no first-level heading jumping straight to a third-level one); a level needs at least two siblings or none"

### E-F6 · app-stylesheet.txt · MINOR · pass 4
LOCATION: 1. House style basics → Ellipsis; 8. What not to change, last item
CURRENT: "Single glyph … or spaced periods, but be consistent within a book" vs. "An author's consistent, defensible choice that differs from house style (e.g. spaced ellipses)"
PROBLEM: §1 says either ellipsis form is house style; §8 cites spaced ellipses as the example of departing from house style. The two rows contradict each other on the sheet's own test case.
PROPOSED: §1 Ellipsis: "Single glyph …; an author's consistent spaced periods are acceptable — note it on the project sheet"

### E-F7 · app-stylesheet.txt · MINOR · pass 4
LOCATION: 4. Dialogue, items 3–4
CURRENT: "Dialogue tags lowercase after a comma: she said, \"Let's go.\" and \"Never,\" he said."
PROBLEM: The first example (she said, "Let's go.") doesn't show a tag after a comma — it shows the rule in the next line (capitalize dialogue introduced by a comma), which reuses the same example. The reader is left working out which example illustrates which rule.
PROPOSED: "Dialogue tags lowercase after a comma: \"Never,\" he said."

### E-F8 · app-stylesheet.txt · MINOR · pass 5
LOCATION: kicker above title
CURRENT: "EDITORIAL · HOUSE STANDARD · 2026-09-13"
PROBLEM: The date is unlabelled and equals today's date, so a reader can't tell whether it is the revision date of the standard or just the render date. For a reference document people are told to follow, the version date is the information that matters (naming & versioning).
PROPOSED: "EDITORIAL · HOUSE STANDARD · REV. 2026-09-13" (and, if the value is currently the render date, make it the last-revised date instead)

### E-F9 · app-stylesheet.txt · NIT · pass 3
LOCATION: 5. Typography & formatting, item 2; 8. What not to change, last item
CURRENT: "(e.g. *gag*)" / "(e.g. (… *gag*.) with the period inside the parenthesis)" / "(e.g. spaced ellipses)"
PROBLEM: §1 Latin abbreviations mandates "comma after e.g./i.e."; the sheet breaks its own rule three times (§3 gets it right: "e.g., \"The same…\""). Inconsistency on a stylesheet reads as a deliberate exception.
PROPOSED: "(e.g., *gag*)" / "(e.g., (… *gag*.) with the period inside the parenthesis)" / "(e.g., spaced ellipses)"

### E-F10 · app-stylesheet.txt · NIT · pass 4
LOCATION: 1. House style basics → Numbers, example cell
CURRENT: "\"seventy-two-degree day\"; \"five in the morning\"; 1998; 40 °C"
PROBLEM: "five in the morning" is repeated verbatim as the example in the next row (Time of day), where it belongs. Duplicate examples make the reader look for a distinction that isn't there.
PROPOSED: "\"seventy-two-degree day\"; 1998; 40 °C"

### E-F11 ★ · app-transmittal-page.txt · CRITICAL · pass 1
LOCATION: page head (between "Smoke Test Manuscript" / "DRAFT / Autosaves as you edit" and "BOOK INFORMATION")
CURRENT: "DRAFT" / "Autosaves as you edit"
PROBLEM: The page opens straight into fields. Nowhere does it say what a transmittal is, who fills it in (client or studio), which sections are required before the studio can start, or what happens when it is marked final. A client landing here cannot tell what the form is for or what "done" looks like (buried/missing ask).
PROPOSED: ADD: one or two lines of intro under the title: what the transmittal is (the handoff record for this manuscript), who completes which parts, which fields are needed before production starts, and what Mark final does (see F12). Voice of the existing checklist help copy ("For each component, choose whether…") is the model.

### E-F12 ★ · app-transmittal-page.txt · MAJOR · pass 1
LOCATION: header actions
CURRENT: "MARK FINAL"
PROBLEM: The one consequential action on the page has no help copy: does it lock the form, notify the studio, start the production clock? Without stakes or an outcome, clients will either avoid it or press it prematurely.
PROPOSED: ADD: a short line adjacent to the button (or in the intro from F11) stating what marking final does, whether it can be undone, and what the client should expect next.

### E-F13 ★ · app-transmittal-page.txt · MAJOR · pass 3
LOCATION: help copy under ART & PRODUCTION PLAN / BUDGET, PRODUCTION PLAN / BUDGET, INSTRUCTIONS FOR DEVELOPMENTAL EDITOR, INSTRUCTIONS FOR COPYEDITOR, TRIM GUIDANCE
CURRENT: "Priority field: include art plan expectations, budget notes, and constraints." / "Priority field: key production-plan and budget context for cover + print timing." / "Priority field: use this for anything the developmental editor must not miss." / "Priority field: use this for anything the copyeditor must not miss." / "Priority field: use this for trim intent, flexibility, and format direction."
PROBLEM: Five fields are labelled "Priority field", so none reads as priority; the prefix is a redundant label doing the work emphasis should do. Two of the five then restate the field name instead of telling the client what to write (emphasis overload; redundant label).
PROPOSED: Drop the prefix everywhere and keep the concrete instruction: "Anything the developmental editor must not miss." / "Anything the copyeditor must not miss." / "Trim intent, how flexible it is, and format direction." For the two plan/budget fields see F14. If one field truly is the priority, say so on that one only.

### E-F14 · app-transmittal-page.txt · MAJOR · pass 5
LOCATION: headings "ART & PRODUCTION PLAN / BUDGET" (after ILLUSTRATIONS) and "PRODUCTION PLAN / BUDGET" (under COVER)
CURRENT: "Priority field: include art plan expectations, budget notes, and constraints." / "Priority field: key production-plan and budget context for cover + print timing."
PROBLEM: Two free-text fields with near-identical names and overlapping help copy ("production plan / budget" in both). A client can't tell which one to put a budget figure in, and the studio will get it split across both (ambiguous labels).
PROPOSED: First: heading "ART PLAN / BUDGET", help "Interior art: what is expected, budget notes, and constraints." Second: heading "COVER & PRINT PLAN / BUDGET", help "Cover and print-timing plan, with budget context."

### E-F15 · app-transmittal-page.txt · MAJOR · pass 1
LOCATION: SUBRIGHTS section
CURRENT: "SUBRIGHTS" / "✕ COPUB WITH" / "✕ TITLE PAGE" / "✕ PAGE IV" / "✕ COVER" / "✕ REMOVE MKTG PGS?"
PROBLEM: No help copy explains that these fields concern a co-publisher's edition, or what "TITLE PAGE", "PAGE IV" and "COVER" ask for (changes for that edition?). "PAGE IV" and "MKTG PGS" are in-house shorthand. A client meets five toggles with no way to know whether they apply (missing context).
PROPOSED: ADD: one line under "SUBRIGHTS" saying these apply only when the book has a co-publisher and that each field records what changes for the co-publisher's edition. Spell out the last item: "✕ REMOVE MARKETING PAGES?"

### E-F16 · app-transmittal-page.txt · MINOR · pass 1
LOCATION: MANUSCRIPT CHECKLIST help copy
CURRENT: "For each component, choose whether it is in the manuscript now, coming later, or not included in this book."
PROBLEM: The table has an EXPECTED DATE column that the help copy never mentions, so the client doesn't know it is needed (and only needed) for "Coming later".
PROPOSED: "For each component, choose whether it is in the manuscript now, coming later, or not included in this book. If coming later, give the expected date."

### E-F17 · app-transmittal-page.txt · MINOR · pass 4
LOCATION: ILLUSTRATIONS table
CURRENT: "TYPE	NO.	HERE	TO COME"
PROBLEM: The only table on the page without a line saying what to enter. "NO." vs. "HERE" vs. "TO COME" is decodable but the reader has to do it cold (unsummarised data).
PROPOSED: ADD, above the table, in the checklist's voice: "For each type, give the total count, how many are in the manuscript now, and how many are still to come."

### E-F18 · app-transmittal-page.txt · MINOR · pass 1
LOCATION: CUSTOM STYLES help copy
CURRENT: "Add any project-specific Word styles needed for this manuscript. These will be copied into the book spec and used for Word template generation."
PROBLEM: "Book spec" and "Word template generation" are studio-side terms; an author doesn't know what counts as a custom style or why it matters to them (missing context).
PROPOSED: "List any Word paragraph or character styles this manuscript uses beyond the standard ones. We copy them into the book spec and build the Word template from them."

### E-F19 · app-transmittal-page.txt · MINOR · pass 1
LOCATION: EDITING → LEVEL OF COPYEDITING / DEVELOPMENTAL EDIT NEEDED
CURRENT: "LEVEL OF COPYEDITING" (options Light / Medium / Heavy) and "DEVELOPMENTAL EDIT NEEDED" (options No / Light pass / Standard developmental edit / Heavy / substantive)
PROBLEM: The client is asked to pick a level with no definition of what light, medium or heavy means here. The house stylesheet describes the light regime ("fix mechanics, don't rewrite"), but the form doesn't point to it.
PROPOSED: ADD: one line of help copy under each select giving a one-phrase definition per level, or a pointer to the /stylesheet/ page's §3 for what "light" covers.

### E-F20 · app-transmittal-page.txt · MINOR · pass 1
LOCATION: COVER → COLORS
CURRENT: "COLORS" / "JDBB" / "FRONT" "SPINE" "BACK" / "PUBLISHER" / "FRONT" "SPINE" "BACK"
PROBLEM: Two identical rows of colour fields, one labelled JDBB and one PUBLISHER, with nothing saying which party fills which (studio proposal vs. publisher's request? number of ink colours vs. named colours?).
PROPOSED: ADD: a line under "COLORS" saying what goes in each row and who fills it.

### E-F21 · app-transmittal-page.txt · MINOR · pass 4
LOCATION: DELIVERABLES
CURRENT: "CLIENT RECEIVES AT PROJECT END" … "FONTS USED (IF LICENSABLE)" / "PRINTER DELIVERY" / "PDF/X" / "OTHER"
PROBLEM: "Printer delivery" sits inside the list headed "Client receives at project end" but describes what goes to the printer, not the client — the list's items don't belong together.
PROPOSED: Make "PRINTER DELIVERY" a sibling heading to "CLIENT RECEIVES AT PROJECT END" (e.g. "PRINTER RECEIVES"), not an item under it.

### E-F22 · app-transmittal-page.txt · MINOR · pass 5
LOCATION: field/section labels
CURRENT: "TRANSMITTAL DATE" (appears under BOOK INFORMATION and again under PRODUCTION); "EST. BOOK PP" (under the checklist counts and again under BOOK DESIGN); "MECHS DELIVERY"; "Other BM"; "Series title/Frontis."; "PPI"; "← MCHECK"
PROBLEM: Two labels appear twice on the same page with no indication whether they are the same value or different ones; several others are in-house abbreviations an author won't expand ("mechs", "BM", "Frontis.", "PPI", "MCHECK"), and the back link doesn't say where it goes (ambiguous labels; links that don't say their destination).
PROPOSED: Remove one of each duplicate or qualify it (e.g. "EST. BOOK PP (FROM MS COUNTS)" vs. "EST. BOOK PP (DESIGN)"); expand "Other BM" → "Other back matter", "Series title/Frontis." → "Series title / frontispiece", "MECHS DELIVERY" → "Mechanicals delivery", "PPI" → "PPI (pages per inch)". ADD: spell out what MCHECK is in the back link.

### E-F23 · app-transmittal-page.txt · NIT · pass 5
LOCATION: page title
CURRENT: "Smoke Test Manuscript"
PROBLEM: The page (and its History / Print / Email outputs) is titled with the manuscript name alone; nothing in the title says it is a transmittal, so a client filing the printed or emailed copy has to open it to know what it is (recipient will search by document type).
PROPOSED: "Transmittal · Smoke Test Manuscript"

---

## Group F — customer factory page /{client}/{project}/factory/ (added after the first pass missed it)

### F-F1 ★ · app-customer-factory.txt · MAJOR · pass 5
LOCATION: step "3 · INSPECT", help copy and button
CURRENT: "Inspect reads your Word file and tells you what won’t come through the way you meant"
PROBLEM: The page names step 3 "Inspect" and the button "Inspect manuscript", but the workshop page, field-notes, the Factory Pass email ("Run a preflight", "the preflight report") and the build email ("Preflight tells you what doesn't conform") all call it preflight. Two names for the thing the reader will do most often; they will look for a "preflight" button and not find it.
PROPOSED: "Inspect — the preflight — reads your Word file and tells you what won’t come through the way you meant"  (one alias on first use; button and step title unchanged)

### F-F2 · app-customer-factory.txt · MINOR · pass 5
LOCATION: step "4 · BUILD", last line
CURRENT: "A build that fails is not counted — you keep the credit."
PROBLEM: "Credit" appears here and in the Factory Pass email but nowhere else; the counter on the page says "3 of 3 builds left". One unit name.
PROPOSED: "A build that fails is not counted — you keep the build."

### F-F3 · app-customer-factory.txt · MINOR · pass 4
LOCATION: SUPPORT block
CURRENT: "After: email support is not included. Self-serve = preflight report + docs. Live support: USD 100/hr, booked in advance, 30-min minimum."
PROBLEM: Telegraphic to the point of shorthand ("Self-serve = …"), and "docs" is not a link anywhere on the page or in the emails; the reader is told support isn't included and then not told where the self-serve path is.
PROPOSED: "After that, email support is not included; the preflight report and the docs are the self-serve path. Live help: USD 100/hr, booked in advance, 30-minute minimum."  plus a link on "the docs".
