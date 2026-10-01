<!-- exported 2026-10-01 15:01 UTC by scripts/punchlist-export.py; source scratch/run/CHECKLIST.md + notes.json -->

# Punchlist 7 · Tue 29 Sep · domains + help

Legend: **YOU** = Jenna · **ME** = Shelley · **BOTH** = together. Click a box: ☐ → ☑, click again to undo; alt- or shift-click = ◐ in progress; **note** opens a reply box (⌘↵ saves; paste screenshots in). Bottom bar adds an item to the Inbox. Archive: [list 6](/admin/runs/PUNCHLIST-2026-09-29-list6) · [list 5](/admin/runs/PUNCHLIST-2026-09-23-list5) · [all runs](/admin/runs/). Item numbers continue (next: 0.51 / 5.29 / 7.14 / 8.16) so threads stay attached.

## 0 · Inbox — new items, untriaged
- [ ] 0.49 **Bring your own agent: agent-facing manuscript prep** (skill for the author's AI → factory-ready files + change report + questions; Phase 2 over the API with per-rule rollback and a Q&A page). Next: ME draft `SKILL.md` v0 · YOU drop the HTML zip in `scratch/` and run it cold with the alternate agent. [design](https://github.com/djinna/jdbbs/blob/main/docs/reviews/AGENT-PREP-HELP-2026-09-30.md)
- [ ] 0.50 **Corrections ledger: surface to authors?** Exists and is live (migration 010, 5 API routes on project auth, admin-only “Corrections Ledger” in the Typesetting panel, re-applied to the source .docx on every PDF+EPUB build, snapshot + diff per output). Never had an author UI. Proposal: “Report a typo” after a proof (typo tier only; bigger → re-upload), match counts shown (today a 0-match correction is silent), “download corrected .docx” so the master doesn’t drift. **Stray test row:** project 7 (vgr/twitter-years) has pending `alchemy → al-TEST`, which is applied to every build; delete?

## 8 · Domains + Help — BOTH (build in order; next phase after Jenna ticks the last)
_Plan: `docs/NEXT_SESSION_PROMPT_2026-09-29.md` · `docs/reviews/CUSTOM-DOMAIN-PLAN-2026-09-29.md` · `docs/reviews/HELP-SYSTEM-SCOUT-2026-09-29.md`. Help blocks 8.1–8.10 in the handoff are 8.5–8.14 here._

- [x] 8.1 **Phase 0 · YOU — DNS + access.** `studio.jdbb.net` CNAME (DNS only) + `domain add` done; serves the app; admin login works there. `jd-blyg` http-proxy integration attached (`https://jd-blyg.int.exe.xyz/` → 200). Open: blyg-side publishing credential (studio password vs blyg API key in an integration).
- [x] 8.2 **Phase 1 · ME — Switch the app to studio.jdbb.net.** `PRODCAL_BASE_URL`; replace ~15 hard-coded `jdbbs.exe.xyz` (srv, scripts, jdbbs-public); `rel=canonical`; 301 public GETs from exe.xyz (never `/api/` or admin); footer/signature, `/factory/api`, CLI defaults, DEPLOY.md. Test: client magic link, factory upload, admin login. Restart + redirects OK’d by you 29 Sep.

  > **shelley** · 2026-09-29 16:06 UTC  
  > Live since 16:04 UTC (0929.4e024d0 + 982fd1c; jdbbs-public c72cdfb).
  > Done: PRODCAL_BASE_URL=https://studio.jdbb.net (all email links, Stripe return URLs, cookies) · footer setting updated · hard-coded hosts replaced (email shell, stylesheet, registration mail, CLI/scripts, /factory/api, /press, talk pages) · Link rel=canonical on public GETs · 301 from jdbbs.exe.xyz for public GETs (not /api/, /admin, /static/).
  > Tested on studio.jdbb.net: client magic link → portal (Secure cookie, API 401→200) · factory upload (test book 69 on mcheck) · /admin/ → exe.dev login.
  > Your turn (the VM can't reach its own exe.xyz address): open https://jdbbs.exe.xyz/factory → should land on studio.jdbb.net/factory; https://jdbbs.exe.xyz/admin/ should stay put. Then tick 8.2.
  > Yours when convenient: Stripe dashboard → business website → studio.jdbb.net (webhook on /api/ can stay; it's never redirected). Clients will sign in once more on the new host.

  > **jenna** · 2026-09-29 16:20 UTC  
  > confirmed https://jdbbs.exe.xyz/factory and https://jdbbs.exe.xyz/admin/ 👍

  > **shelley** · 2026-09-29 16:20 UTC  
  > Thanks — Phase 1 closed. Redirect + admin fallback confirmed from your browser. Next: 8.3 landing-page draft.

- [~] 8.3 **Phase 2 · BOTH — Fresh jdbb.net landing page** (replaces the Blot stub). ME: draft on a preview route (public doc, 1240 shell); YOU: facts + approve copy. Cutover is yours in Cloudflare (apex + www CNAME, flatten) + `domain add`. **Never touch MX.**

  > **jenna** · 2026-09-29 16:24 UTC  
  > go ahead

  > **shelley** · 2026-09-29 16:31 UTC  
  > Draft is up: https://studio.jdbb.net/jdbb-net (preview on the studio host; nothing on jdbb.net has changed).
  > Built from the talk, /press and field notes. The orange [ ] placeholders are facts I couldn't find a source for: the 1986 publisher's name (if you want it named), presses/imprints/roles over the years, number of books + titles you're proud of + any awards, city, whether j@djinna.com is the contact, GitHub/social. I also left out the naval-architecture work — your call.
  > Routing is already in and tested: jdbb.net/ → this page, www → apex, jdbb.net/factory etc. → studio, old Blot URLs → the landing page. Cutover (yours, after you approve the copy): Cloudflare apex CNAME @ → jdbbs.exe.xyz DNS only + 'Flatten CNAME at root', www CNAME the same, then ssh exe.dev domain add jdbbs jdbb.net and … www.jdbb.net. Never touch MX.

  > **shelley** · 2026-09-29 16:38 UTC  
  > No answers from you yet on the placeholders, so the page is waiting on you. One is partly done: I used "About 300, over 30 years" from your blyg bio for the Books line. Still open, all optional (reply here and I'll fold them in):
  > 1. The 1986 publisher: name it or leave it unnamed?
  > 2. The years in between: presses/imprints, your roles, 2–3 titles you're proud of, any awards.
  > 3. City (or leave it out).
  > 4. Is j@djinna.com the contact? Any GitHub/social links?
  > 5. Naval architecture: in or out?
  > Once you OK the copy, the cutover steps are all yours, in Cloudflare, and I'll check each one: (a) delete the Blot apex/www records, (b) CNAME @ → jdbbs.exe.xyz (DNS only) + DNS → Settings → Flatten CNAME at root, (c) CNAME www → jdbbs.exe.xyz (DNS only), (d) from your machine: ssh exe.dev domain add jdbbs jdbb.net and ssh exe.dev domain add jdbbs www.jdbb.net. Leave MX alone.

- [x] 8.4 **Phase 3 · YOU decide — blyg.jdbb.net.** Repurpose `jd-blyg` VM (CNAME `blyg` → `jd-blyg.exe.xyz` + `domain add`) or move to the Blygger Studio Worker?

  > **jenna** · 2026-09-29 16:25 UTC  
  > I don't know what the second choice is? Had thought we would continue with our exe VM but open to better ideas!

  > **shelley** · 2026-09-29 16:25 UTC  
  > Sorry, that was jargon. The second choice was to run the blyg on Cloudflare instead of on a VM. Blygger's own starter kit ("Blygger Studio") is built to deploy as a Cloudflare Worker, a small program Cloudflare hosts for you, and jdbb.net's DNS is already at Cloudflare.
  > My view: stay with the exe VM (jd-blyg). It's already running, I can already reach it through the integration, and it sits alongside everything else we run. The Worker only saves us one VM, and it would put the blog behind a second account and a second deploy process.
  > So for 8.4: keep jd-blyg. When you're ready: Cloudflare CNAME blyg → jd-blyg.exe.xyz (DNS only), then from your machine: ssh exe.dev domain add jd-blyg blyg.jdbb.net. One open question stays: how we publish to it (the blyg's studio password vs a blyg API key held in an integration).

  > **jenna** · 2026-09-29 16:29 UTC  
  > CF and exe setup steps complete, confirm? for pw vs key, I'll take your recommendation

  > **shelley** · 2026-09-29 16:31 UTC  
  > Confirmed: blyg.jdbb.net resolves to jd-blyg.exe.xyz and serves 200; /studio → blyg.jdbb.net/studio/login. Ticked.
  > Recommendation: a blyg API key, held in an exe.dev integration like the jd-blyg one. It's scoped to publishing, you can revoke it without changing your password, and it never sits on this VM. The studio password stays yours, for writing by hand. Next step (next session): brief jd-blyg's Shelley to add an API-key endpoint (or find out whether Blygger already has one), then you wrap the key in the integration.

  > **shelley** · 2026-09-29 16:38 UTC  
  > Checked. Your blyg runs Blygger Studio 0.7.0, and it has no API key. Its /api is private and works only with the studio-password cookie. A stable publishing API is still an open item on Blygger's roadmap. Another blyg owner (the blygger-desktop app) added a single "owner token", and I followed the same approach.
  > Brief for jd-blyg's Shelley: docs/briefs/JD-BLYG-PUBLISH-TOKEN-2026-09-29.md (in the prodcal repo; paste the whole file into a Shelley chat on jd-blyg). It asks for a token accepted in an X-Blyg-Token header, with rotation steps and a test. It also flags two things that need your OK: the blyg still calls itself jd-blyg.exe.xyz (feed, manifest, og:url) rather than blyg.jdbb.net, and its author link points at the old jdbbs.exe.xyz/factory.
  > Once it's built, you create the integration from your machine: ssh exe.dev integrations add http-proxy --name blyg-publish --target https://blyg.jdbb.net --header "X-Blyg-Token:<token>" --attach vm:jdbbs. It's a separate integration because the existing jd-blyg one already uses the Authorization header for the exe.dev key.

- [~] 8.5 **Help 1 · ME** — `docs/help/*.md` + frontmatter, render in shell, tiers, nav from site_pages, FTS search, `llms.txt` / `.md`.

  > **shelley** · 2026-09-29 16:52 UTC  
  > Help 1 is built and deployed. https://studio.jdbb.net/help/ is live, but all 10 pages are drafts, so only you see them (with a "draft" badge). Clients see "Help is being written".
  > Please read these (about 10 min): start-here, sign-in, transmittal, send-your-manuscript, inspect, build-and-download, factory-pass, style-sheet, factory-api. Also maintaining-help, which only you will ever see. To publish a page, change status: draft to status: live in docs/help/<page>.md, or tell me which ones and I'll flip them.
  > Things I need you to check:
  > • Copyright style: I wrote "Copyright and permissions lines". The code doesn't say what it's for.
  > • Prices: I left them out and linked /factory instead, so help doesn't go stale.
  > • Fixed a bug: the upload box said "up to 100 MB", but the server only accepts 50 MB. It now says 50.
  > • Inspect: factory.js has labels for 8 finding types the checker never reports (tracked changes, comments, low-resolution images…). The help doesn't promise those. The page copy says Inspect flags "images too small to print", but I couldn't find code that does. Worth confirming.
  > Also built: search (/help/search), /help/llms.txt, and every page as .md for agents.

- [~] 8.6 **Help 2 · ME** — “?” per page + “learn more” links; route-coverage test (+ missing transmittal site_pages row).

  > **shelley** · 2026-09-29 17:07 UTC  
  > Built and deployed. Each page now gets a "?" in the top nav that links to its help page, and the factory's Upload, Inspect, Build and Download steps each get a "learn more" link. Both only appear once the help page they point to is published. Right now only you see them, because you can see drafts. There's also a test: every client or public route must have a help page or a written reason for having none. Marketing pages, essays, decks and the DIY-hidden calendar are opted out with reasons. The site_pages row for the transmittal was already there (retired, redirects to factory#transmittal).

- [~] 8.7 **Help 3 · ME** — Report a nit (no login, honeypot + rate limit) → Section 0, tagged.

  > **shelley** · 2026-09-29 17:07 UTC  
  > Built and deployed. Every page footer (except admin) has "Report a nit". It opens a small form: what's wrong, optional email, and any text you had selected comes along. No sign-in. There's a hidden spam trap, and a limit of 5 per person per 10 minutes and 200 a day overall. Each nit is saved and lands here in Section 0, tagged [nit·help] or [nit·app]. I tested it end to end on the live site. That's the "#1 smoke test" item you saw, and I removed it. The optional email isn't used yet. A "fixed, thanks" reply comes later, through the email-consent guard.

- [~] 8.8 **Help 4 · ME** — Inspect findings content slice (one entry per finding type) + “never Word — your editor” copy lint.

  > **shelley** · 2026-09-29 17:07 UTC  
  > Drafted and deployed, still as drafts. There are 20 help pages, one per type of Inspect finding, each covering why it's flagged, how serious it is, how to fix it in any editor, and when to leave it alone. Each finding row on the factory's Inspect results gets a "learn more" link once its page is published. There's also a new copy check: help pages and client-facing screens may not say "Word" or "Microsoft". I fixed about a dozen strings to pass it. Please check one change: the "export again" hints used to name each program's menu (e.g. "Google Docs → Download → Microsoft Word"). They now say "Google Docs → File → Download → .docx". If you'd rather keep the literal menu names, I'll add them as allowed exceptions.

- [ ] 8.9 **Help 5 · ME** — self-update loop, `/admin/help/` review, approval email (EMAIL_SYSTEM.md), standing permissions.
- [ ] 8.10 **Help 6 · ME** — public What's new (drafted from commits, you approve).
- [ ] 8.11 **Help 7 · ME** — stats (views, thumbs, searches) + one daily usage email to you.
- [ ] 8.12 **Help 8 · ME** — AI Q&A A/B (50%, sticky) + kill switch + circuit breaker.
- [ ] 8.13 **Help 9 · ME** — `/factory/api` becomes the Help API section.
- [ ] 8.14 **Help 10 · ME** — 2–3 scripted Sample Press screenshots.
- [ ] 8.15 **YOU — open questions:** DIY calendar (my view: hide for DIY, no help pages v1) · landing-page facts / names to leave out.

## 9 · Factory watch — "teammate, not tool" (plan: docs/plans/FACTORY-WATCH-PLAN-2026-10-01.md)

- [ ] 9.1 **YOU — routing decisions.** P1 list (immediate email to you), daily digest time, first hot-fix categories, QC per-build email now or later. Current state assessed in the plan §1.

  > **shelley** · 2026-10-01 14:34 UTC  
  > Plan written: docs/plans/FACTORY-WATCH-PLAN-2026-10-01.md. §1 lists what alerts exist today and who gets each one; §3 has your four questions. Nothing is built yet.

- [ ] 9.2 **ME — alert hook + routing.** One `alert()` helper → punch list + my chat; P1s also email you; verify whether a store sale alerts anyone; fix the store go-live note (0.31).
- [ ] 9.3 **ME — deploy watch.** ~60 min after each restart: smoke routes + a Mike Check proof, tail journal + factory events, compare error rate, write `docs/runs/DEPLOY-…md`, one-line verdict here. Rollback only with your OK.
- [ ] 9.4 **ME — floor watch.** Rules over factory_events (repeat failures, rejected uploads, stuck builds, sign-in denials, expired pass) → `[watch]` rows + my chat, with log context; I triage and draft replies or fixes.
- [ ] 9.5 **ME — hot fixes within standing permissions** (shares the 8.9 permission store); logged in the deploy doc + digest.
- [ ] 9.6 **ME — daily digest to you** (merge with 8.11: builds, failures + what I did, nits, stuck customers); retire the per-build QC email.
- [ ] 9.7 **BOTH, optional — outside uptime ping** on studio.jdbb.net/healthz → email you.

## 10 · Policies — privacy + terms (Stripe asks for a privacy policy)

- [~] 10.1 **ME — Privacy policy `/privacy`.** Stripe: what we collect, how we use it, who we disclose it to and how, how we keep it safe. Built from a code inventory (forms, tables, logs, cookies, processors), not boilerplate. Public doc in jdbbs-public; linked from every footer, /factory, checkout and the terms.

  > **shelley** · 2026-10-01 15:01 UTC  
  > Draft is up: https://studio.jdbb.net/privacy (unlisted, marked "draft for review"; nothing links to it yet).
  > It covers Stripe's five points (what we collect, its use, who gets it, how it gets to them, security), plus how long we keep things, your rights, cookies, where the data lives (our server is in New York), and children.
  > I built it from a read of the code, not from a template, and the read turned up things the page has to admit:
  > • The index add-on sends the book's text to an AI model (Claude, through exe.dev). The page says so.
  > • 15 pages load fonts from Google, so visitors' IP addresses go to Google. The page says so; the fix is 10.5.
  > • AI assistants (me) can see account and activity records while fixing things. The page says so. Tell me if you'd rather word it differently.
  > • Nothing deletes expired passes, mail logs or sign-in records yet. The page promises only what the pass terms already promise; 10.5 builds the rest.
  > • The GitHub repo is public and contains client manuscripts (Ghosts, Twitter Years). See my note on 7.7. That needs your decision before this page goes live, because the page says we never publish anyone's work.

- [~] 10.2 **ME — Terms of Use `/terms`** (site-wide: accounts + sign-in, your work stays yours, acceptable use, purchases → Factory Pass terms, availability, liability, changes, contact). Same plain-English voice as `/factory/terms`.

  > **shelley** · 2026-10-01 15:01 UTC  
  > Draft is up: https://studio.jdbb.net/terms (unlisted, draft). It has 14 short points in the same voice as /factory/terms, and it defers to the pass terms on anything about a pass.
  > Two choices I made that you should check:
  > • Refunds: "Full refund any time before your first build, proof or final", which is what /factory says. If you meant before the first final, tell me. A proof gives you a usable EPUB, so I left it as is.
  > • "If we ever close the factory, we'll tell pass holders at least 30 days ahead." That's a promise, so it's highlighted.
  > It also covers: one pass is for one manuscript, fonts can't be extracted, nits and suggestions, you cover our costs if someone makes a rights claim against us (this matches your attestation), liability capped at what was paid, consumer law still applies, and nothing to cancel.

- [ ] 10.3 **ME — Factory Pass terms v2.** `/factory/terms` (v1, 18 Sep) still says every build uses a credit; since 0.28 proofs are free and only finals count. Point 7 says files go only to "hosting, email, payments", but the index add-on sends the book's text to an AI model (Claude via exe.dev): say so. Align with /factory; your OK before it changes (customers agreed to v1).
- [ ] 10.4 **YOU — review both + facts:** legal name of the seller (as on Stripe), governing law / city, contact address, and whether a lawyer reads them before we call them final.

  > **shelley** · 2026-10-01 15:01 UTC  
  > What I need from you (the orange bits on both pages):
  > 1. Seller's legal name, as it appears on Stripe: you personally, or a company? And a city.
  > 2. Governing law: I put Hong Kong SAR. Is that right?
  > 3. Do you want a postal address shown? Stripe recommends one but doesn't require it.
  > 4. Retention: payment records "seven years?" (the Hong Kong business-record rule) and server logs "90 days". OK?
  > 5. Backups: offsite copies already expire after 180 days, but the monthly copies on the VM are kept forever. OK to cap those at 180 days too? (10.5)
  > 6. Should a lawyer read these before they go live? Not legal advice from me. These are plain-English drafts.
  > Once you answer, I'll fill these in, remove the draft notice, add Privacy and Terms links to every footer and to /factory checkout, and you set both URLs in Stripe (10.7).

- [ ] 10.5 **ME — make the privacy policy true before it goes live.** (a) self-host the Google Fonts (15 pages load fonts.googleapis.com, so visitors' IPs go to Google); (b) cap server logs at 90 days (journald); (c) local monthly DB backups are kept forever, but the policy says deleted data leaves backups after 180 days: cap them (your OK, ties to 7.10); (d) purge job for expired passes: promised on /factory and in the terms, nothing deletes yet, first expiry March 2027.
- [ ] 10.6 **ME — security fixes the inventory found.** The `prodcal_auth_<project>` cookie holds the raw project password for a year (switch to a signed token); no HSTS or security headers; workshop registrations store plain IP addresses (nits already hash them).
- [ ] 10.7 **YOU — Stripe dashboard, after 10.4:** Settings → Public details → privacy policy URL + terms of service URL. Then ME (optional): the "I agree to the terms" box at Checkout (`consent_collection`).

## 5 · Carried from list 6 — factory build + decisions
- [~] 5.25 **Custom styles slices 4–7** — EPUB classes + Word template basedOn/indent; verse-family coalesce into one poem block; admin panel Based on/Indent; docs. [design note](https://github.com/djinna/jdbbs/blob/main/docs/reviews/CUSTOM-STYLES-MARKERS-2026-09-22.md) · **+ Face option (body/heading/code) per custom style** — Casey's sans Normal (0.46)

  > **shelley** · 2026-09-22T14:24:48Z  
  > Slices 1–3 verified on mcheck book 58 after the 0.39 hotfix: 457/457 markers resolve; verse → verse2 → verse3 step in visibly (p.12, Four Haiku). Open: verse/verse2/verse3 lines do not coalesce into one poem block (each level is its own block with a stanza gap between) — fix in slice 4 with the EPUB classes. Slices 4–7 remain for Thu 24.

- [ ] 5.26 **Inline character markers** `[[code]]…[[/code]]` (Fotis's ITAL/BOLD leak) · **+ `[[tex]]…[[/tex]]`** for LaTeX typed as text (0.43)
- [ ] 5.24 + 5.27 **Error / Inspect wording** — “our bug, email us” when the failure is in our generated file; helpful not scolding. Fotis and cblass both got “search for this in your manuscript” for a string that exists only in our template.

  > **shelley** · 2026-09-22 13:57 UTC  
  > Fotis review is in: https://github.com/djinna/jdbbs/blob/main/docs/reviews/FOTIS-HELO-WORD-REVIEW-2026-09-22.md — summary, timeline, census of his file, 15-row factory-vs-author table, rewrites of every error string he saw (incl. the two build-failed texts: 'our generated config' vs 'your paragraph'), and a short note you can send him. Root cause was ours (style named `break` → `#let break(content)` in our config; fixed in 9258e56, he can retry). Handing off to a fresh session now.

  > **jenna** · 2026-09-22 14:18 UTC  
  > x https://jdbbs.exe.xyz/hermescorp/book-001/stylesheet/
  > It looks like Fotis's book has a whole bunch of custom styles. Can you check whether the factory added those somehow during our review just now, or did Fotis misunderstand and think that she was supposed to add all of the styles that are in her book, not just the ones that are custom? Pass this to a new sub-agent in our current conversation.?
  >
  > ![screenshot](img/20260922-141747-400.png)

  > **shelley** · 2026-09-22 14:19 UTC  
  > Passed to a subagent (fotis-styles-origin, read-only). It is checking transmittal_versions for project 30 over time + every code path that writes custom_styles. Report to scratch/fotis/STYLES-ORIGIN.md; I will post the verdict here. Prior review (FOTIS-HELO-WORD-REVIEW §3) found 21 declarations, 9 duplicating factory/Word built-ins — consistent with declaring every style, but the timeline check will settle it.

  > **shelley** · 2026-09-22 14:22 UTC  
  > Verdict (b): Fotis typed them; the factory never adds declarations (only the transmittal PUT and the admin spec editor write custom_styles — upload/Inspect/build only copy a final transmittal into the spec). Her transmittal history grows 0→4→11→14→18→19→21 across client saves on Mon 21 Sep, starting 12:12 UTC right after Inspect flagged 2 Highs; 12:12:32 Inspect showed 0. So she read "declare your styles" as "declare every Word style". For Fotis: "You entered the full Word style inventory — the factory didn't add that. Declare only the ones that aren't factory styles: keep narration + in1–in6, drop verse/quote/break/part/chapter/preface/BOLD/ITAL/footnote*." Feeds 0.34 (onboarding: styles are names). Report: scratch/fotis/STYLES-ORIGIN.md

- [ ] 0.40 **`Chapter Title` factory style** — H1/H2 become real 1st/2nd-level heads; pre-pass shifts only when a file uses Chapter Title; mirrored in bookmap + chapter_detect; template/emails/#bring; then retire the title-drop rule. ~1 day. [Seapunk review §6.6](/admin/runs/)

  > **shelley** · 2026-09-22 14:43 UTC  
  > Triaged from inbox (was auto-numbered 0.37, already taken by the style-sheet link). This is item 6 of the Seapunk review — their whole book was dropped because the title-as-H1 rule swallowed everything under it. Proposal: new factory style Chapter Title; a pre-pass shifts Chapter Title→H1 and H1/2/3→H2/3/4 only when a file uses Chapter Title, so nothing already on the floor changes. Thursday, ~1 day. Today I only hotfix the symptom (title rule must never drop content; finals refuse an empty body).

- [ ] 0.34 **Onboarding: styles are names** — welcome email, /factory intro, #bring (Fotis's 21 declarations; Casey's `[[commentary]]`)

  > **shelley** · 2026-09-22 13:43 UTC  
  > Triage: After-workshop → §5 as 0.34 (parked, per your note). Scope I'll take: the welcome email, /factory intro + #bring, the transmittal's Custom Styles help text, and the Word template's own Template Guide page — one message repeated in the same words everywhere: “a style is a *name* you give a paragraph; the factory reads the names, not the look.” Also worth a 90-second screencast/GIF of applying a style in Word / Pages / Docs, since the blind spot is about the mechanics, not the concept. I'll fold in what Fotis's and Toby's files teach (both marked every paragraph by hand rather than using the style sheet). Nothing ships during the freeze.

- [ ] 0.41 **Style-sheet add-rule at the top** (decided 1.4): one prominent “+ Add a rule” by the sheet’s title, same form, section picker defaulting to *Our additions* (= the free-form “decisions we’re tracking” list, already exists as §8); no Add on the public house sheet · 0.44 Side Note style (after 1.5) · Seapunk §6.7 figure captions + §6.8 Inspect image noise

  > **shelley** · 2026-09-22 15:29 UTC  
  > Checking what exists: the book sheet (/{client}/{project}/stylesheet/) already has “+ Add a rule to <Section>” at the foot of each section. Is the ask (a) make that more visible / put one Add at the top, (b) a free-form “decisions we are tracking” list (word list: e-mail not email; Seapunk capitalised) not tied to a house section, or (c) an Add on the public house sheet /stylesheet/ when reached from a book? Parked for Thu unless you say hotfix.

- [ ] 5.28 **Log hygiene** — `style markers applied` logs every marker (tens of KB per build); log counts + unresolved only.
- [ ] 0.31 Drop `PRODCAL_BUILD_QC_EMAIL` from .env after the workshop · fix `store-go-live.sh` env export (go-live note to you never sent)

  > **shelley** · 2026-09-22 13:17 UTC  
  > Live (2b73cc8, restarted 13:14 UTC, mcheck smoke green). You got the first one: “QC · mcheck · … (proof both, book 54)”. Links, not attachments — before = the uploaded .docx (new /download/source route), after = newest PDF/EPUB of that kind, plus Inspect report, factory page, Floor. Failed builds mail too (FAILED in the subject). On the Floor every “book N” now has [docx · pdf · epub]. Remind me to switch it off after the workshop.

- [ ] 1.5 **0.44 side notes** — margin-notes page design (Tufte, whole-book switch) or is a set-off aside enough?

  > **jenna** · 2026-09-23 14:06 UTC  
  > Hmmm your "0.44 side notes" does seem to match our numbering here, I don't have the context I need to answer. Point me to a GitHub or whatever else we already have handy?

  > **shelley** · 2026-09-23 14:06 UTC  
  > Context: your inbox item 0.44 from Tue 15:52 UTC — “Can we add an affordance for side notes, and if so, how?” — archived with my full answer here: https://github.com/djinna/jdbbs/blob/main/docs/runs/PUNCHLIST-2026-09-23-list5.md (search 0.44). Short version of what I proposed: author marks a paragraph Side Note ([[sidenote]]), same “styles are names” contract. Two ways to set it: (1) MARGIN NOTES — a per-book spec switch that widens the outer margin and puts each note beside its sentence, small, in the body face (the Tufte page; it changes the text block of the whole book, so it is a title-level design decision). (2) SET-OFF ASIDE — indented, smaller, thin rule, inside the text column; this is also what the EPUB always gets because reflow has no margin. Either way ~half a day, Thu/Fri after 5.25. The one-word decision: do you want (1) margin notes built as an option, or is (2) the aside enough for the workshop books? “Both” is a fine answer too — (2) is the fallback inside (1).

- [ ] 4.4 Store live: `/api/public/store/config` enabled ✓ · first real checkout → check `/admin/store/` orders + Stripe dashboard.

## 7 · Security and recovery  — carried from list 6 — kept private, not exported

