"""Book 2 smoke — the messy path.

Produce a Google-Docs-export-style manuscript written OUTSIDE the jdbb template:
no Heading styles (headings are bold/larger Normal paragraphs), scene breaks
faked with blank lines and "***", a pasted colored run, a manually italicized
epigraph, a hand-typed numbered list, mixed straight/curly quotes, a
tab-indented paragraph, double spaces, a "Normal (Web)" stray style, and a
"14 Sept — …" dateline (Book 1 false-positive check).

    python3 scratch/make-ms-messy.py            # -> scratch/mcheck-book2-messy.docx
    python3 scratch/make-ms-messy.py --clean T  # -> also scratch/mcheck-book2-clean.docx,
                                                #    restyled with template T's styles
                                                #    (simulates Word's Organizer import)
"""
import sys, copy
from docx import Document
from docx.shared import Pt, RGBColor
from docx.enum.text import WD_ALIGN_PARAGRAPH

OUT = "scratch/mcheck-book2-messy.docx"

d = Document()
# Google Docs export: Normal is Arial 11, single spaced, 0 indent.
n = d.styles["Normal"]
n.font.name = "Arial"; n.font.size = Pt(11)
n.paragraph_format.space_after = Pt(0); n.paragraph_format.first_line_indent = None
# A stray "Normal (Web)" style, as pasted from a browser.
web = d.styles.add_style("Normal (Web)", 1)
web.base_style = n; web.font.name = "Times New Roman"; web.font.size = Pt(12)

def para(text, **fmt):
    p = d.add_paragraph()
    r = p.add_run(text)
    for k, v in fmt.items(): setattr(r.font, k, v)
    return p

def heading(text, size=16):
    """Manual heading: bold, larger, centered, Normal style."""
    p = para(text, bold=True, size=Pt(size))
    p.alignment = WD_ALIGN_PARAGRAPH.CENTER
    return p

def blank(k=1):
    for _ in range(k): d.add_paragraph()

def scene_break_blank():
    blank(2)

def scene_break_stars():
    p = para("***"); p.alignment = WD_ALIGN_PARAGRAPH.CENTER

def body(text): return para(text)

# ── front matter, all manual ────────────────────────────────────────────────
heading("Messy Draft: A Field Guide", 22)
p = para("What Happens When You Skip the Template", italic=True); p.alignment = WD_ALIGN_PARAGRAPH.CENTER
blank()
p = para("Mike Check"); p.alignment = WD_ALIGN_PARAGRAPH.CENTER
blank(3)
body("Copyright © 2026 Mike Check. All rights reserved.  Published by jdbb studio.")
blank(2)

# manually italicized epigraph, right-aligned source
p = para("\u201cFormat nothing. Style everything.\u201d", italic=True)
p.paragraph_format.left_indent = Pt(72)
p = para("\u2014 a production editor", italic=True); p.alignment = WD_ALIGN_PARAGRAPH.RIGHT
blank(2)

chapters = [
 ("Chapter 1: The Export", [
   "It started, like most of these things do, with a Google Doc.  The manuscript had lived there for two years, shared with three readers and one very patient editor.",
   'Mike hit "File", then "Download", then "Microsoft Word (.docx)". The file was 212 kilobytes. It felt finished.',
   "\u201cIt\u2019s done,\u201d he wrote in the chat. \u201cUploading now.\u201d",
 ]),
 ("Chapter 2: Bold Is Not a Heading", [
   "The first thing Inspect said was that the book had no chapters.  Mike scrolled back up. There they were, in 16-point bold, centered, exactly where chapters go.",
   "But bold is a look, not a name.  The factory reads names.",
   "\tHe had typed a tab to indent this paragraph, because that is what the keyboard is for.",
 ]),
 ("Chapter 3: The Space Between", [
   "Scene breaks were two blank lines. Sometimes three. Once, memorably, a row of asterisks that had drifted left when the font changed.",
   "In the PDF, blank lines are just blank lines. They do not survive a page break, and they do not tell the EPUB anything at all.",
 ]),
 ("Chapter 4: Pasted", [
   None,  # colored run inserted below
   "The green had come from a website. The website had come from a search. Nobody remembered pasting it, which is how pasting works.",
 ]),
 ("Chapter 5: The List", [
   "Here is what Mike had learned, in the order he learned it:",
   "1. Bold is not a heading.",
   "2. Blank lines are not a scene break.",
   "3. A tab is not an indent.",
   "4.  Two spaces after a period is a habit, not a rule.",
 ]),
 ("Chapter 6: Quotes", [
   "Some quotes were \"straight\" and some were \u201ccurly\u201d and a few were 'single' and \u2018also single\u2019, depending on which app he had been typing in that week.",
   "The apostrophes were worse. Mike's, Mike\u2019s, and once, inexplicably, Mike`s.",
 ]),
 ("Chapter 7: Dateline", [
   "14 Sept \u2014 Ran the Inspect for the fourth time. It flagged the date on this line as a list item. It is not a list item. It is a date.",
   "15 Sept \u2014 It still thinks so.",
 ]),
 ("Chapter 8: Normal (Web)", [
   "WEB",  # marker for Normal (Web) styled paragraph
   "That paragraph arrived in a style called Normal (Web), which nobody chose and nobody can see.",
 ]),
 ("Chapter 9: Import", [
   "The prep email had said: import the template's styles into your working document.  Word: Manage Styles, Import/Export, copy everything across.",
   "He did. The bold headings were still bold. But now there was a Heading 1 to apply, and an Epigraph, and a Section Break with a little breve in it.",
 ]),
 ("Chapter 10: Build", [
   "The second build took forty seconds. The PDF was six by nine. The chapters started on fresh pages.",
   "\u201cHuh,\u201d Mike wrote. \u201cThat\u2019s all it wanted.\u201d",
 ]),
]

for i, (title, paras) in enumerate(chapters):
    if i % 2 == 0: scene_break_blank()
    else: blank()
    heading(title)
    blank()
    for t in paras:
        if t is None:
            p = d.add_paragraph()
            p.add_run("The fourth chapter had a paragraph that was ")
            r = p.add_run("bright green, in Verdana, 13 point"); r.font.color.rgb = RGBColor(0x00, 0x99, 0x00); r.font.name = "Verdana"; r.font.size = Pt(13)
            p.add_run(", and a highlight that nobody had asked for.")
            continue
        if t == "WEB":
            d.add_paragraph("This paragraph was pasted from a browser and kept its clothes on.", style="Normal (Web)")
            continue
        body(t)
    # a scene break inside every third chapter
    if i % 3 == 1:
        scene_break_stars()
        body("After the break, the paragraph started flush left, as if nothing had happened.")
    elif i % 3 == 2:
        scene_break_blank()
        body("Two blank lines later, another scene.")

d.save(OUT)
print(OUT, len(d.paragraphs), "paragraphs;", sum(1 for p in d.paragraphs if not p.text.strip()), "empty")

# ── optional: simulate the Organizer "import template styles" + restyle ─────
if "--clean" in sys.argv:
    tpl_path = sys.argv[sys.argv.index("--clean") + 1]
    CLEAN = "scratch/mcheck-book2-clean.docx"
    # Open the template, empty it, and re-emit the manuscript into it using the
    # template's styles. This is what the Organizer import + "apply styles"
    # pass in Word leaves behind; the manual formatting is kept on purpose
    # where an author would plausibly leave it (bold runs on headings).
    t = Document(tpl_path)
    for p in list(t.paragraphs): p._element.getparent().remove(p._element)
    for tb in list(t.tables): tb._element.getparent().remove(tb._element)
    src = Document(OUT)
    HEADS = {c[0] for c in chapters}
    prev_break = True
    for p in src.paragraphs:
        txt = p.text
        if not txt.strip():
            continue                                    # blank-line scene breaks → dropped
        if txt.strip() == "***":
            t.add_paragraph("\u02d8", style="Section Break"); prev_break = True; continue
        if txt in HEADS:
            t.add_paragraph(txt, style="Heading 1"); prev_break = True; continue
        if txt.startswith("Messy Draft"):
            t.add_paragraph(txt, style="Title"); continue
        if txt.startswith("What Happens"):
            t.add_paragraph(txt, style="Subtitle"); continue
        if txt.startswith("Copyright"):
            t.add_paragraph(txt.replace("  ", "\n"), style="Copyright"); continue
        if txt.startswith("\u201cFormat nothing"):
            t.add_paragraph(txt, style="Epigraph"); continue
        if txt.startswith("\u2014 a production editor"):
            t.paragraphs[-1].add_run("\n" + txt); continue
        if txt == "Mike Check":
            t.add_paragraph(txt, style="First Paragraph"); continue
        style = "First Paragraph" if prev_break else "Normal"
        np_ = t.add_paragraph(style=style)
        for r in p.runs:
            nr = np_.add_run(r.text.lstrip("\t").replace("  ", " "))
            nr.italic = r.italic   # keep italics; drop bold/size/color/font
        prev_break = False
    t.save(CLEAN)
    print(CLEAN, len(t.paragraphs), "paragraphs; styles:", sorted({p.style.name for p in t.paragraphs}))
