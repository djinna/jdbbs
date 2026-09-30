# Bring your own agent — agent-facing manuscript prep (2026-09-30)

*Idea from Jenna, 30 Sep 2026 (voice note, clarified in chat). Status:
design note; nothing built. Punch list 0.49. Related: help scout
(`HELP-SYSTEM-SCOUT-2026-09-29.md` §6), machine factory
(`MACHINE-FACTORY-SPEC-2026-09-18.md`), Word-free authoring (d2)
(`WORD-FREE-AUTHORING-2026-09-19.md`), Snowmoon ingest
(`SNOWMOON-INGEST-ASSESSMENT-2026-09-28-1230.md`).*

## The idea

Authors arrive with files that are nowhere near a book: a zip of HTML, a
folder of chapters, a doc with hand-typed bullets. None of the *mise en place*
has been done. Asking them to do it themselves goes badly. Many already have
an AI agent. So we publish instructions **written for that agent**: "give
this to your agent". What comes back should be ready for the factory, with a
report of what changed and a list of questions only a person can answer.

Phase 2: the author's agent talks to our factory directly over the API, so
the cleanup pass is a loop (fix → Inspect → fix) with change reporting,
rollback, and a Q&A page for the human.

## Current pattern (checked 30 Sep 2026)

Three layers, not rivals:

| layer | what it is | we have |
|---|---|---|
| `llms.txt` + `.md` per page | discovery: an index agents fetch | yes: `/help/llms.txt`, `/help/<slug>.md` |
| **Agent Skill** (`SKILL.md`, Anthropic open spec Dec 2025; Claude Code, Codex, Gemini CLI, Cursor read it) | a folder: `SKILL.md` (frontmatter `name` + `description`, body < ~5k tokens) + `references/` + `scripts/` + `assets/`. Teaches a procedure. | no, **this is the new piece** |
| MCP server | live, authenticated tool calls | no; the six REST calls + `factory-cli.py` do the same job for now |

For chat-only users (no skills folder), the same body is served as a single
paste-able prompt. One source, two wrappers.

## Phase 1: the skill (no server changes)

`prodcal-manuscript-prep/`

- `SKILL.md`: when to use it; the **contract** (below); the procedure:
  inventory → propose book map → convert → self-check → report.
- `references/`: the style list and `[[markers]]` (from
  `send-your-manuscript.md`), custom-style rules, the Inspect findings
  catalogue (`finding-*.md` are already written as remediation notes, so
  they double as agent references), the house style sheet as editable rules.
- `assets/`: the project template `.docx`, and the fenced-div Markdown
  convention with one worked example.
- `scripts/`: optional `md-to-template-docx.sh` (pandoc
  `--reference-doc`) so a capable agent can produce the `.docx` itself; else
  it returns Markdown.

Served at `/help/agent/` (zip + raw files) and listed in `llms.txt`. A
human page `/help/prepare-with-your-agent` explains it in studio voice, with
a copy-the-prompt button.

### The contract (what the agent must hand back)

The factory is `.docx`-in, but the input is effectively format-agnostic:
Markdown with `::: {custom-style="…"}` fenced divs → template `.docx`
gives byte-identical Typst (verified 19 Sep, path d2). So the agent may return
either.

1. **Manuscript**: one template-styled `.docx`, *or* fenced-div Markdown
   (one file per chapter is fine).
2. **`book-map.json`**: order, and front / body / back for each unit, from
   the source files (HTML filenames, `<title>`, nav/TOC files). This is the
   step these authors skipped.
3. **`CHANGES.md`**: every change, grouped by **rule** (see rollback),
   with count, 1–3 before/after samples, and source file + location.
4. **`QUESTIONS.md`**: decisions it would not make. Each item has the
   question, options, what it did meanwhile (the default), and where it
   applies.
5. **Originals untouched**, and a manifest of the input files with hashes.

Hard rules for the agent: **never rewrite prose**. Only structure,
styles, and mechanical style sheet rules. When unsure, ask rather than
decide. No new content. Keep images and their captions. Say which model did
the work.

## Rollback: per rule, not per edit

Undoing edits one at a time is too much overhead (Jenna, 30 Sep). A
change's **rule** is the unit you undo:

- *Structural rules*: "h2 → Heading 1", "`.dialogue-anna` → custom style
  Anna", "hand-typed bullets → List", "`<br><br>` → scene break".
- *Editorial rules* **are style sheet rows**: serial comma on,
  US spelling, curly quotes. The project style sheet
  (`/{client}/{project}/stylesheet/`) already holds these decisions, so
  the agent reads it (or proposes additions to *Our additions*), and
  rolling back "serial comma" means flipping that row and re-running that
  rule.

Phase 1 rollback is cheap: rules are deterministic transforms applied in
order to the untouched originals, so undoing one means re-running without
it. Phase 2 the server keeps that list per cleanup session. Individual
edits stay visible in `CHANGES.md` but can't be reverted one by one, except
as "exclude this location from rule X".

## Google Docs (in the first rollout; Jenna's main client writes there)

What we already know (`WORD-FREE-AUTHORING-2026-09-19.md` (a),
`send-your-manuscript.md`, `/word-free`):

- **No custom paragraph styles in Docs.** It has a fixed set only:
  Normal text, Title, Subtitle, Heading 1–6. Those export to Word's built-in
  styles, which the factory reads natively. Everything else (quote, verse,
  epigraph, code, signature, custom styles) goes through **`[[style]]`
  markers** typed in the text: `[[quote]]` at the start of a paragraph,
  `[[verse]]…[[/verse]]` for ranges, `[[style:Letter]]` for a declared custom
  style. Shipped 18 Sep (`apply-style-markers.py`).
- **Verified end to end** on 19 Sep: a Docs-shaped `.docx` with 12 markers
  gave byte-identical Typst to the Word baseline. Inspect: 0 high, markers
  listed under "Marked styles", book map correct.
- **Inline:** italic and bold come through. There are no character styles, so
  small caps and the like need direct formatting or (later) inline markers
  (5.26).

What the skill tells the author's agent for a Docs manuscript:

1. **Get the file.** Take the `.docx` export, never copy-paste. Three ways,
   easiest first:
   (a) the author does File → Download → Microsoft Word and hands it over;
   (b) a link-shared doc: `https://docs.google.com/document/d/<ID>/export?format=docx`
   (no OAuth; "anyone with the link can view");
   (c) Drive API `files.export` with the Word MIME type (OAuth or a service
   account the doc is shared with; export cap ~10 MB, so a heavily illustrated
   book may need (a)).
2. **Clean in the `.docx` copy by default, or in a *copy* of the Google Doc.**
   The Docs API (`documents.batchUpdate`) can set `namedStyleType` (turn
   bolded Normal text into Heading 1) and insert `[[markers]]` as text, so a
   capable agent can prep the Doc itself. The author keeps working in Docs,
   and their master and the book stay the same file. Rules: work on a
   `files.copy`, never the original; the API can't create *suggested* edits,
   so changes land as direct edits, and rollback is "the original is
   untouched" plus `CHANGES.md`.
3. **Before export:** accept or reject every suggestion and resolve comments
   (to check: whether the export carries suggestions as tracked changes;
   Inspect should flag them if so). Expect these to lose content or flatten to
   text on export: smart chips, dropdowns, building blocks, document *tabs*
   (to check how multi-tab docs export), drawings (become images).
4. **Typical Docs mess** the agent should expect, each with an Inspect
   finding and a help page already: headings faked with bold/size
   (`heading-lookalike`), web fonts in runs (`unusual-font`), hand-typed
   bullets (`manual-list`), blank-line spacing (`manual-break`),
   highlight/colour used as notes to self (`highlighted-text`,
   `colored-text`).

**Factory API access for the client** is the same as for anyone else: a
project password (`POST /api/projects/{id}/auth`, studio-minted) plus a live
Factory Pass; then the six calls or `factory-cli.py`. A Docs-reading agent
loops: export → upload → Inspect → fix (in the Doc copy or the `.docx`) →
repeat.

## Phase 2: factory-to-factory (API)

Built on the machine factory (token + Factory Pass, six calls). What's
missing:

- **Accept `.md` / zip at upload** and convert server-side (the d2 "M"
  effort). The zip carries manuscript + `book-map.json` + `CHANGES.md` +
  `QUESTIONS.md`.
- **Cleanup session** on a book: the ordered rule list with per-rule
  on/off → rebuild. Inspect counts before and after each pass, so the
  report shows "Inspect: 212 → 9 findings".
- **Inspect as the loop's oracle**: findings already carry a type and a
  help slug, so the agent reads `/help/finding-<type>.md` and fixes. Free
  calls, so iterating costs no builds.
- **Q&A page** `/{client}/{project}/questions/` (customer) with an admin
  mirror. Both ends, per AGENTS.md. Answers flow back: a style sheet answer
  becomes a style sheet row; a structural one toggles a rule. Doubles as a
  paid service: Jenna drives the session with the author.
- **MCP later**: a thin wrapper over the same calls plus the skill as
  its prompt. Only if API testers ask.

## Test plan

- **Fixture**: an author's zip of HTML files (not Snowmoon). Jenna holds
  it; drop into `scratch/<name>/` (gitignored) when we commit to it. It
  tests the book-map step, which is exactly what's missing.
- **Cold run**: Jenna's alternate agent gets *only* the published skill,
  and neither of us helps it. Same week as the alternate API-client test.
- **Score**: returned the contract (5 parts)? Inspect findings before and
  after; book map right versus Jenna's judgement; questions raised
  (did the real decisions come up, with little noise?); changes Jenna
  would reject; any prose touched (must be 0).
- Second fixture, later: Snowmoon (colour-coded speakers → custom styles),
  which also checks this agrees with the studio-side importer idea.

## Order

1. Draft `SKILL.md` + references from the existing help pages (help page
   status `draft`, admin-only). ~½ day.
2. Cold run on the HTML zip; revise from the score.
3. Human page + `/help/agent/` serving + `llms.txt` entry.
4. Phase 2 pieces, in order: `.md`/zip upload → Q&A page → cleanup session
   with per-rule toggles.

## Open

- Who answers questions by default, author or Jenna? (Both see the page;
  a "studio-assisted" flag could route them to Jenna first.)
- Does a skill-prepared manuscript need its own Factory Pass tier, or is it
  just an input path?
