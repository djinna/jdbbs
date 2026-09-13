# better-documents — our document-review standard

Two sources, kept together because one is a distillation of the other.

## 1. The original essay (linked, not vendored)

Anil Dash, **"Make better documents."**, 10 March 2024
<https://www.anildash.com/2024/03/10/make-better-documents/>

~3,100 words, © Anil Dash — read it at the source. Section map, for finding
your way back to a passage:

1. Know your audience and your goals
2. Stop formatting everything to death
3. White space
4. Learn your tools
5. Pay attention to sequencing and order
6. It's not a murder mystery
7. Give people wayfinding
8. Ask answerable questions
9. Close by reiterating your mission and goals
10. Bonus update: Name things right!

Points in the essay that the skill's distillation flattens, and that matter
for our pages (paraphrased):

- The audience questions go further than "who reads this": what *historical*
  context are you assuming, is the framing culturally appropriate to them,
  which other parties have a stake in the decision and have they been
  consulted, and can you state the counter-arguments to your own proposal
  honestly. (§1)
- The formatting advice is tool-level: ignore almost every button on the
  formatting bar. Bold means important, italic means emphasis, color means
  distinct — stacking them says you don't know which. Applies to charts too.
  He links Dark Horse Analytics' "remove to improve" data-table GIF as the
  canonical demo. (§2)
- Bullets are a *diagnostic*, not just a layout: a list whose items don't
  belong together exposes drift that prose hides. (§3)
- Inconsistency is read as *meaning*. Audiences are trained on produced media
  and will hunt for the significance of any change in title size or font;
  the restraint rule exists partly so those accidental changes are easy to
  catch. Do a formatting pass on anything imported from elsewhere. (§4)
- Document order tends to reflect how hard each part was to write, not how
  important it is — and readers rationally assume first = most important.
  If the order is chronological or inherited, *say so on the page*. (§5)
- "Speaking to it" is not a channel. The key fact belongs on the page, not in
  speaker notes or an appendix, because the stakeholder who matters will be
  the one who wasn't in the room. (§6)
- Wayfinding is the one place he endorses color: a limited palette, one per
  section, used with total consistency. Chart titles should state what the
  data show. (§7)
- Constrained choices unstick people ("Option A at higher cost and faster, or
  Option B?"); open questions ("how can we do better?") produce debate, not
  decisions. A third option emerging from a constrained prompt is fine. (§8)
- Naming: think about how the *recipient* will search for it later — by your
  name/org if they're outside, by project if inside; date in the filename;
  meeting invites named for the purpose, never "Meeting with Sam". (§10)

## 2. The skill (vendored, GPL-3.0)

`SKILL.md` and `document-design.md` from
<https://github.com/anildash/better-documents> (fetched 2026-09-13,
`LICENSE` alongside). Five review passes — audience & purpose · structure &
sequencing · formatting restraint · wayfinding & density · naming &
versioning — with severity scales and the report format
*location → problem → fix → severity*. Constraints we hold to: don't alter
the author's voice, don't invent content, be specific (quote the text).

Scope note: the skill is for documents, not UI, and says to follow existing
brand constraints exactly. For us that means the Terminal Folio tokens and
`/static/theme.css` are out of scope; reviews target prose, order, emphasis,
and titles. Where the skill's "neutral default" (black sans on white, no
color) conflicts with the studio's house style, house style wins.

## How we use it

- Content review of `jdbbs-public/*` and the client-facing pages in
  `srv/static/`: findings go to `/admin/content-review/` for accept / edit /
  reject before anything is changed. First run: 2026-09-13.
- Generate mode for new outward documents (workshop handouts, talk deck,
  announcement emails): conclusion and ask first, structural signposts past
  one page, dated titles.
