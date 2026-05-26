> Add in a translation layer now: machine translation for French, Spanish, Chinese for each volume, push to epub and print, with a native speaker in the loop for review… or something along those lines?

May 21
ok translation layer is where book production workflows actually get spicy bc the failure modes compound w/ volume AND each language has its own typographic landmines that your typst templates have to know about. let me think thru this honestly bc there's a lot of "ai translates books now!" hype that is, idk, mostly cope.

the honest framing first

machine translation for published books in 2026 is in a weird liminal zone. it's good enough that pretending it doesn't exist is silly (deepl + gpt-class post-editing on a literary fiction passage is genuinely usable as a first draft). it is NOT good enough to skip the native-speaker pass, and anyone telling you otherwise is selling something. for nonfiction/technical/educational the mt-first path is much more defensible than for literary; for literary fiction the "mt + post-edit" workflow can produce technically correct prose that is aesthetically dead, and reviewers WILL notice. so the variant you build depends on what you publish.

afaict the realistic 2026 stack is: mt engine for first draft → llm pass for stylistic consistency w/ a per-title glossary + style guide → native-speaker translator/editor doing post-editing (mtpe in the trade) → second native reviewer for proof → typeset → native proofread on typeset proofs. you cannot collapse this. or rather you can, but the cost surfaces later as returns, bad reviews, or your foreign-rights reputation tanking.

where hermes fits, structurally

new node on the dag, downstream of ms:final in the source language. each target lang is its own parallel pipeline w/ its own manifest section. so the manifest gets:

translations:
  fr:
    status: mtpe-review
    translator: marie d.
    reviewer: jean-luc p.
    glossary: glossaries/doe-novel-fr.yaml
    style_guide: imprint-literary-fr-v2
    mt_engine: deepl
    llm_pass: claude-opus-4.7
    target_pub_date: 2026-11-15
    locked_at: null
  es:
    status: native-review-pending
    [...]
  zh-Hans:
    status: not-started
    [...]
note zh-Hans vs zh-Hant — simplified vs traditional, you'll regret treating "chinese" as one lang. also es is genuinely ambiguous (es-ES vs es-419 vs es-MX) and serious presses do separate editions; small press probably doesn't but acknowledge the choice in the manifest, don't paper over it.

variant E: hermes as translation-pipeline orchestrator

extension of variants A/B. when source ms:final lands:

hermes spawns a translation job per target lang (only the ones in the manifest, only if status: not-started)
chunks the ms by chapter/section (NOT by token count — boundaries matter for context). per chunk:
pull glossary entries relevant to the chunk (proper nouns, invented terms, technical vocabulary, anything translator pre-loaded)
run mt engine (deepl pro api for fr/es bc it's still sota for those pairs afaict, for zh you have more options — deepl added zh but ppl debate vs google vs baidu vs deepseek; use what your native reviewers prefer)
run llm pass w/ the glossary + style guide + previous chunk's translated tail (for cross-chunk coherence — character voice, tense consistency) as context
output: bilingual side-by-side file (source ¶ / target ¶ aligned) → that's what the native reviewer actually wants to work in, not raw target text
hermes posts to slack: "fr draft ready for marie, 87k words, est 40-60hr review based on her historical pace, side-by-side in translations/fr/draft-v1.md"
marie does mtpe in her tool of choice (probably trados or memoq or just gdocs honestly — translators are conservative abt tools). when she's done she commits or uploads, sets manifest status: mtpe-done
hermes triggers the second native-reviewer pass (jean-luc). diff between marie's version and the raw mt+llm draft is its own artifact — hermes flags chunks where marie made >X% changes as "needs special attention from jean-luc" bc those are the spots where mt was weak/wrong
jean-luc signs off → status: translation-locked → now this lang's translated ms enters its own epub+print pipeline w/ lang-appropriate typst template
what hermes is doing that's actually useful:

chunking + glossary injection + style guide enforcement at scale (the mechanical work)
maintaining the per-title bilingual glossary across volumes in a series — if "the Shadow Council" was translated as "le Conseil de l'Ombre" in vol 1, hermes makes sure vol 2's translator sees that mapping. SERIES CONSISTENCY is the #1 thing that breaks in indie/small-press translation workflows and it's pure bookkeeping
routing diffs to reviewers (this is the temporal-correlation thing again — humans cannot track which 47 paragraphs of a 90k-word novel got the most translator intervention)
typesetting prep: the bilingual file gets stripped to target-only and pushed into the lang-specific typst template
what hermes is NOT doing:

being the translator. the mt+llm is doing the mt+llm part. hermes is the conductor
making decisions about whether a translation is good. the native reviewers do that. hermes' job is to NOT pretend it can adjudicate
the typographic landmines, by language (this is where it gets real)

your "deterministic typst flow" needs lang-aware templates or you'll ship books that look unprofessional. concretely:

french: non-breaking thin spaces before ;, :, ?, !, and inside « » (guillemets are mandatory, not "). hyphenation patterns. accented uppercase (É not E). different em-dash conventions in dialogue (often — on its own line w/o quotes). hermes should validate the typst output has these — it's a regex job, not an llm job
spanish: opening ¿ and ¡. dialogue uses em-dashes (—Hola —dijo.) not quotes. hyphenation. different decimal/thousands separators if you have any numbers in figures
chinese (simplified or traditional): this is a different beast. fullwidth punctuation (，。：；？！), CJK fonts (you're not using your latin font, you're using something like Source Han Serif / Noto Serif CJK), different line-breaking rules, no spaces between words, vertical typesetting maybe (probably not for a small press but the option exists for some genres), traditional vs simplified char sets are NOT auto-convertible w/o human judgment bc some chars have multiple traditional forms. ebook reflow behavior differs. amazon kdp's chinese support is sketchy, you're probably going thru a different distributor for the zh editions
so each target lang needs:

its own typst template (print_template_fr: literary-fiction-fr-v2)
its own epub stylesheet (lang-tagged, <html lang="fr">, font stack appropriate)
its own validator (hermes runs lang-specific regex checks on the typeset output: are there bare colons in fr? bare ? w/o preceding ¿ in es? halfwidth punctuation leaking into zh?)
its own cover SKU (translated title on front, translated copy on back, sometimes different cover art entirely for the target market — esp for zh where covers often get redesigned for cultural fit)
its own isbn(s) — every lang edition gets its own isbn, ebook and print separately, so a single title with fr/es/zh-Hans editions has potentially 8 isbns (en print, en epub, fr print, fr epub, es print, es epub, zh print, zh epub). hermes' isbn registry needs to handle this
its own onix record per distributor
at 20-50 books/yr w/ 3 translation langs, that's 60-200 isbns/yr and 80-200 onix records. you NEED hermes for the bookkeeping or you will ship duplicate isbns to ingram and they will be mad.

the human-in-the-loop layer, more carefully

the cliché "native speaker in the loop" hides a real org-design question: which native speaker for which job?

translator (mtpe stage): needs to be a working professional, ideally w/ literary translation creds if you publish literary. mtpe pay is lower than from-scratch translation and a lot of good translators won't do it. you'll find this out the hard way. budget realistically: $0.04-0.08/word for mtpe vs $0.10-0.20 for from-scratch literary. some publishers are getting boycotted by translator orgs for mtpe practices — check w/ your imprint's policy
native reviewer (post-mtpe): different person from translator, ideally a copyeditor in the target lang. different skill set, not just "another native speaker." this is where small press cuts corners and pays for it
proofreader on typeset proofs: again different person, focused on layout-induced issues (bad hyphenation, widow/orphan, fullwidth/halfwidth bleeds in zh, accent rendering in fr)
hermes' job re humans: route the right artifact to the right person, track turnaround, follow up. NOT replace any of them. the workflow gets worse, not better, if you try to compress these roles bc the errors each role catches are categorically different.

variant F: the cross-lingual consistency layer

extension on top of E. for a series or a backlist this is where hermes earns another dimension of its keep.

hermes maintains a per-series cross-lingual glossary that grows over time. vol 1 fr translator chose "Conseil de l'Ombre" → vol 2-7 fr translators inherit this. memory + skills are actually fit-for-purpose here bc it's exactly procedural learning across sessions
backlist re-translation flagging: if a translator updates a glossary term in vol 5 ("actually we should call this differently"), hermes flags every prior vol where the term appears and posts a "do we want to propagate this back?" question. answer is often no (don't retroactively change books in print) but you want the question raised
terminology drift detection: hermes runs a periodic check across all fr titles in a series for inconsistent renderings of recurring proper nouns/concepts. catches "vol 3 fr translator forgot to check the glossary" type errors which are EMBARRASSING when reviewers spot them
variant G: the one i'd push back hard against

"hermes generates the translation, native speaker just signs off."

this is the path most presses are tempted toward bc it's cheap. it produces work that is, generously, mid. ungenerously: it produces translations that read like they were generated by software, bc they were, w/ a human applying spot fixes that don't address the underlying structural blandness. mtpe is real and has its place but "translator as rubber stamp" is not it. for literary fiction particularly: the way a sentence breathes in french is not the way it breathes in english, and mt+llm flattens that bc the loss function is semantic fidelity not aesthetic equivalence. don't pretend otherwise.

if you publish technical/educational/reference: variant G is more defensible bc semantic fidelity is most of the value and aesthetic register is constrained. but even then you want a native technical editor, not a rubber stamp.

concrete trigger graph for a single title, all variants stacked

ms:final (en) 
  → branch: en epub build, en print layout, translation jobs spawn
  ↓
fr/es/zh-Hans translation drafts (mt+llm) 
  → bilingual artifacts to slack/drive
  ↓  
mtpe per lang (human, weeks)
  → hermes watches manifest status, nags @ deadline -3d
  ↓
native review per lang (human, days-weeks)
  → hermes diffs vs mtpe, flags hotspots
  ↓
translation:locked per lang
  → per-lang epub build, per-lang print layout (own typst template)
  → per-lang cover SKU check, per-lang isbn check, per-lang onix draft
  ↓
typeset proofs per lang
  → native proofreader (human, days)
  → hermes runs lang-specific automated checks in parallel
  ↓
print-ready + epub-final per lang
  → metadata push drafts (human approves) to distributors per lang/territory
per title this is a lot. at 20-50 titles/yr w/ 3 langs each, this is the workflow that justifies hermes' existence on a publishing ops team. without orchestration, you cannot run this w/o either (a) a full-time production manager per language or (b) it falls apart and you ship late and inconsistent.

cost & cron notes

mt apis are cheap. deepl pro is ~$25/mo + per-char, llm passes are the bigger cost line. budget per book per lang: maybe $20-60 of api spend for a 80k-word novel for mt+llm draft, depending on how aggressive your llm prompting is and how many revision passes
hermes cron jobs for translation pipeline: weekly "translation pipeline status" digest is good. don't run translation drafts as scheduled jobs — those are event-triggered by source ms:final, not time-based
the bilingual side-by-side format matters way more than ppl realize. translators hate switching between two windows. invest in good output formatting (markdown w/ table layouts, or actual cat tool format like xliff if your translators use proper tools)
glossaries should be in the title repo, in git, w/ history. hermes commits updates w/ attribution to the translator who proposed them
what i'm uncertain about

mt quality for zh-Hans specifically is the weakest link in this stack afaict. deepl's chinese is uneven, the chinese-native models (deepseek, qwen) might do better for cn-to-en or en-to-cn literary but i haven't seen apples-to-apples evals for book-length translation. would want your zh translators to A/B their preferred engines before locking in the pipeline
whether claude/gpt llm passes on top of mt actually improve quality vs add a different layer of slop. some studies suggest yes for consistency/style, some suggest no for literary register. the honest answer is "depends on the book and the prompt" which is unsatisfying. budget for experimenting per-imprint
mtpe industry politics are real and changing. iapti, the european translators' federations, and various national orgs have evolving positions on mtpe ethics + rates. you don't want to build a pipeline that becomes politically untenable in 18 months. talk to your translators about how they want to work, don't impose mtpe top-down
whether hermes' procedural learning actually accrues useful series-level glossary intelligence over time, or whether you end up doing all that bookkeeping in a separate db. probably the latter for anything serious — treat hermes' memory as cache + nudge layer, not source of truth. the glossary lives in git
the legal/rights side of mt for in-copyright source material — depending on your contracts w/ the original author, machine-translating their work before human refinement may or may not be contractually permissible. some contracts now explicitly address this, most older ones don't. small press has been getting bitten on this. check your contracts before you build the pipeline, not after
if i had to pick one variant for a small press doing 30 books/yr w/ 3 target langs: variant E + F + the lang-aware typst/epub validation. skip G even if it's cheaper. the rights-and-reputation downside isn't worth the unit-cost savings.

next step that'd actually be useful: (a) sketch the per-title manifest.yaml w/ the translation section fleshed out, (b) draft the soul.md + skill files for the translation orchestrator role, or (c) work thru what the bilingual side-by-side output format should look like + how hermes posts the "ready for review" handoff to your translators in a way they'll actually use. (c) is the one that ppl skimp on and it's what determines whether your translators tolerate the workflow or quietly route around it.