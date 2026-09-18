# "About the Book Factory" — reader-facing text for books built by the factory

Drafted 2026-09-18 for punch-list item 0.8. Two lengths: a colophon for the
copyright page (small type, replaces or follows the interior credit), and a
back-matter page headed **About the Book Factory**. Both are offered to the
author, never imposed: a checkbox on the transmittal (copyright page → colophon;
back matter → the page), off by default. Wiring is a separate item.

Voice: plain, third person, no marketing adjectives; says what happened to the
book, not what the studio believes. Numbers only where they are true of every
build. The URL is the one line that sells.

---

## A. Colophon (copyright page, small type) — 78 words

> This book was made in the jdbb studio book factory. The author wrote it in
> Word, in a template of thirteen named styles generated from a one-page
> transmittal; the factory read the file, reported what it found, and set the
> print interior and the EPUB from that single source, in licensed type, with
> no hand-composed pages. Corrections are rebuilt the same way, so every copy
> of this edition is set from the same file. jdbbs.exe.xyz/factory

## A′. Colophon, shortest (one line after the typesetting credit) — 26 words

> Set from the author's Word file by the jdbb studio book factory, print and
> EPUB from one source, no hand-composed pages. jdbbs.exe.xyz/factory

---

## B. Back-matter page — "About the Book Factory" — 486 words

> **About the Book Factory**
>
> This book was not typeset in the usual sense. No one placed its pages by
> hand. It was built by the jdbb studio book factory, a small machine for
> turning a Word file into a book, and this page says how, because the method
> is part of what you are holding.
>
> The author wrote in Word, in a template the factory generated for this
> title. The template has thirteen named paragraph styles and nothing else:
> Normal, First Paragraph, three levels of heading, Block Quote, Epigraph,
> Verse, Code Block, Section Break, Copyright, Signature and Glossary Entry. Every paragraph
> in the manuscript carries one of those names. That is the whole contract
> between writer and factory. There is no software to learn and nothing to
> install; the work of authorship stays in the tool the author already used.
>
> Before the writing began, the author filled in a transmittal: a short form
> that records what the book is and how it should be set. The page size. The
> front matter — a dedication, an epigraph, a foreword — and its order. The
> copyright page, which the factory composes from the transmittal rather than
> asking anyone to type it. Special characters, mathematics, custom styles the
> book needs. The transmittal is the specification; the template is generated
> from it; the finished book follows it.
>
> When a draft was ready, the author uploaded it and the factory inspected it:
> a machine read of the file that reports which styles were used, where the
> file departs from the template, how many images it found and how large each
> will print, where the chapters begin. The report is the factory's reply. The
> author fixes what it flags and uploads again, for as long as it takes.
>
> Then the build. From that one Word file the factory produces two things at
> once: a print-ready interior PDF at the trim size on the transmittal, and an
> EPUB with the author's cover embedded. The interior is set in the studio's
> house design — a typographic system of margins, type sizes, running heads
> and spacing worked out once and applied to every book that passes through —
> in fonts the studio licenses for print. The two editions come from the same
> source, in the same pass, so they cannot drift apart.
>
> This matters after publication. A book set by hand is finished when the
> typesetter stops; a correction means opening the files again and hoping the
> lines still fall where they did. A book from the factory is set from its
> source every time. Fix the Word file, rebuild, and every page is composed
> afresh, in seconds. The edition you hold is one build of a file that can be
> built again.
>
> The factory is the work of jdbb studio, a one-person book studio that
> published books the ordinary way for thirty years before deciding that the
> ordinary way was mostly waiting. It is offered to anyone with a manuscript
> and a few hours. If you have written something and would like it to become
> a book like this one, the factory is at **jdbbs.exe.xyz/factory**.

---

### Notes for Jenna

- "thirty years" and "one-person" in the last paragraph are placeholders for
  whatever is true; cut the sentence if you'd rather not do biography here.
- "thirteen named paragraph styles" is true as of today (Signature and
  Glossary Entry added 2026-09-18). If the count changes, this line and the colophon change with it.
- "in seconds" — a full build of a 250-page book is ~9 s on the VM; true.
- The page is ~490 words; at 6 × 9 it fills one recto with air. If you want
  it on one page in Small trim, cut paragraph 4 (Inspect) to two sentences.
- Suggested placement: last page of back matter, after About the Author, so
  the book ends on the author and then the maker.
