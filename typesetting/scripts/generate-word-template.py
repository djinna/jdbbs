#!/usr/bin/env python3
"""Generate a styled Word (.docx) template from a book spec JSON.

Usage:
    echo '{...}' | python3 generate-word-template.py > template.docx
    python3 generate-word-template.py --spec-file spec.json --output template.docx

Reads a book_specs.data JSON on stdin (or --spec-file) and produces a .docx
with paragraph/character styles matching the spec. The .docx serves as a
starting template for copyeditors working in Word.
"""

import argparse
import io
import datetime
import json
import sys

from docx import Document
from docx.enum.style import WD_STYLE_TYPE
from docx.enum.text import WD_ALIGN_PARAGRAPH
from docx.shared import Pt, Emu
from docx.oxml.ns import qn
from docx.oxml import OxmlElement


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

SECTION_BREAK_CHARS = {
    "breve":    "\u02D8",       # ˘
    "asterism":  "\u2042",      # ⁂
    "dinkus":   "* * *",
    "fleuron":  "\u2767",       # ❧
    "rule":     "\u2014" * 3,   # ———
}


def _weight_to_bold(weight) -> bool | None:
    """Convert a heading weight spec to a python-docx bold flag.

    Returns True for bold/semibold/600, False for medium/500/normal, None for
    anything else (inherit).
    """
    if weight is None:
        return None
    w = str(weight).strip().lower()
    if w in ("bold", "700"):
        return True
    if w in ("semibold", "600"):
        return True
    if w in ("medium", "500", "normal", "400", "regular"):
        return False
    # numeric fallback
    try:
        return int(w) >= 600
    except ValueError:
        return None


# Stand-in fonts. Clients don't have the book's typefaces installed (and
# Plantin / Proxima are licensed, so we never embed them). Word silently
# substitutes Calibri for an unknown face, which looks nothing like the book.
# Instead we name a face everyone has and say so in the Template Guide.
STANDIN_FONTS = {
    "body":    "Georgia",
    "heading": "Arial",
    "code":    "Courier New",
}

# The styles a manuscript should use. Everything else python-docx's default
# template ships (Body Text 2, List Continue 3, Macro Text…) gets hidden from
# the Styles pane. Order here is the order in Word's "Recommended" view.
FACTORY_STYLES = [
    "Normal", "First Paragraph", "Heading 1", "Heading 2", "Heading 3",
    "Block Quote", "Epigraph", "Verse", "Code Block", "Section Break",
    "Copyright", "Signature", "Glossary Entry",
]


def _ensure_font_exists(run_font, font_name: str):
    """Set the font name on a run's font object (all four slots), and drop
    any theme-font attributes — Word prefers w:asciiTheme over w:ascii, which
    is how headings kept coming out in Calibri Light."""
    run_font.name = font_name
    # Font.element is the owning w:style / w:r, not its rPr.
    rpr = run_font.element.find(qn("w:rPr"))
    rfonts = rpr.find(qn("w:rFonts")) if rpr is not None else None
    if rfonts is not None:
        for attr in ("asciiTheme", "hAnsiTheme", "eastAsiaTheme", "cstheme"):
            rfonts.attrib.pop(qn(f"w:{attr}"), None)
        for attr in ("ascii", "hAnsi", "eastAsia", "cs"):
            rfonts.set(qn(f"w:{attr}"), font_name)


def _set_black(style):
    """Force a style's text to plain black (python-docx's default theme paints
    headings blue via w:themeColor)."""
    rpr = style.element.get_or_add_rPr()
    color = rpr.find(qn("w:color"))
    if color is None:
        color = rpr.makeelement(qn("w:color"), {})
        rpr.append(color)
    for attr in ("themeColor", "themeShade", "themeTint"):
        color.attrib.pop(qn(f"w:{attr}"), None)
    color.set(qn("w:val"), "000000")


def _parse_length_in(value, default_in: float) -> float:
    """'0.75in' / '18mm' / '54pt' / bare number (inches) → inches."""
    if value is None:
        return default_in
    if isinstance(value, (int, float)):
        return float(value)
    v = str(value).strip().lower()
    try:
        if v.endswith("in"):
            return float(v[:-2])
        if v.endswith("mm"):
            return float(v[:-2]) / 25.4
        if v.endswith("cm"):
            return float(v[:-2]) / 2.54
        if v.endswith("pt"):
            return float(v[:-2]) / 72
        return float(v)
    except ValueError:
        return default_in


# CT_Settings child order (ECMA-376 17.15.1.78), as far as we need it. Word
# rejects settings.xml with children out of sequence, so insert by schema.
_SETTINGS_ORDER = [
    "writeProtection", "view", "zoom", "removePersonalInformation",
    "removeDateAndTime", "doNotDisplayPageBoundaries", "displayBackgroundShape",
    "printPostScriptOverText", "printFractionalCharacterWidth", "printFormsData",
    "embedTrueTypeFonts", "embedSystemFonts", "saveSubsetFonts", "saveFormsData",
    "mirrorMargins", "alignBordersAndEdges", "bordersDoNotSurroundHeader",
    "bordersDoNotSurroundFooter", "gutterAtTop", "hideSpellingErrors",
    "hideGrammaticalErrors", "activeWritingStyle", "proofState", "formsDesign",
    "attachedTemplate", "linkStyles", "stylePaneFormatFilter",
    "stylePaneSortMethod", "documentType", "mailMerge", "revisionView",
    "trackRevisions", "doNotTrackMoves", "doNotTrackFormatting",
    "documentProtection", "autoFormatOverride", "styleLockTheme",
    "styleLockQFSet", "defaultTabStop",
]


def _insert_setting(settings, el):
    """Insert *el* into w:settings at its schema position."""
    local = el.tag.split("}")[1]
    idx = _SETTINGS_ORDER.index(local)
    later = {qn(f"w:{n}") for n in _SETTINGS_ORDER[idx + 1:]}
    for i, child in enumerate(list(settings)):
        if child.tag in later:
            settings.insert(i, el)
            return
    # Everything present precedes us in the schema, or is beyond our table
    # (defaultTabStop is always there in python-docx's default, so we rarely
    # get here).
    settings.append(el)


def _set_page_to_trim(doc, page: dict):
    """Section page size = trim, margins = the book's, mirrored for facing
    pages, so line length and page depth in Word feel like the book."""
    from docx.shared import Inches
    w = _parse_length_in(page.get("width_in"), 5.5)
    h = _parse_length_in(page.get("height_in"), 8.5)
    sec = doc.sections[0]
    sec.page_width = Inches(w)
    sec.page_height = Inches(h)
    sec.top_margin = Inches(_parse_length_in(page.get("margin_top"), 0.75))
    sec.bottom_margin = Inches(_parse_length_in(page.get("margin_bottom"), 0.75))
    # With mirrorMargins on, Word reads left as inside and right as outside.
    sec.left_margin = Inches(_parse_length_in(page.get("margin_inside"), 0.7))
    sec.right_margin = Inches(_parse_length_in(page.get("margin_outside"), 0.6))
    settings = doc.settings.element
    if settings.find(qn("w:mirrorMargins")) is None:
        _insert_setting(settings, settings.makeelement(qn("w:mirrorMargins"), {}))


def _tidy_styles_pane(doc, visible: list[str]):
    """Show only the factory styles in Word's Styles pane.

    Every style not in *visible* is marked semiHidden + unhideWhenUsed and
    loses qFormat; ours get qFormat + a uiPriority in list order. Latent
    styles (built-ins Word knows about but the file doesn't define) default
    to hidden and the per-style exceptions are dropped. Finally the pane
    filter is set to "Recommended", which is exactly the qFormat set.
    """
    W = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
    order = {name: i for i, name in enumerate(visible)}
    for style in doc.styles:
        el = style.element
        for tag in ("w:semiHidden", "w:unhideWhenUsed", "w:qFormat", "w:uiPriority"):
            for old in el.findall(qn(tag)):
                el.remove(old)
        # These children belong right after name/aliases/basedOn/next/link.
        anchor = 0
        for i, child in enumerate(list(el)):
            if child.tag in (qn("w:name"), qn("w:aliases"), qn("w:basedOn"),
                             qn("w:next"), qn("w:link"), qn("w:autoRedefine"),
                             qn("w:hidden")):
                anchor = i + 1
        new = []
        if style.name in order:
            new.append(el.makeelement(qn("w:uiPriority"), {qn("w:val"): str(order[style.name] + 1)}))
            new.append(el.makeelement(qn("w:qFormat"), {}))
        else:
            new.append(el.makeelement(qn("w:uiPriority"), {qn("w:val"): "99"}))
            new.append(el.makeelement(qn("w:semiHidden"), {}))
            new.append(el.makeelement(qn("w:unhideWhenUsed"), {}))
        for j, n in enumerate(new):
            el.insert(anchor + j, n)

    latent = doc.styles.element.find(qn("w:latentStyles"))
    if latent is not None:
        for ex in list(latent):
            latent.remove(ex)
        latent.set(qn("w:defSemiHidden"), "1")
        latent.set(qn("w:defUnhideWhenUsed"), "1")
        latent.set(qn("w:defQFormat"), "0")
        latent.set(qn("w:defUIPriority"), "99")

    settings = doc.settings.element
    for old in settings.findall(qn("w:stylePaneFormatFilter")):
        settings.remove(old)
    flt = settings.makeelement(qn("w:stylePaneFormatFilter"), {
        qn("w:val"): "2801",  # visibleStyles + clearFormatting + top3HeadingStyles… → "Recommended"
        qn("w:allStyles"): "0", qn("w:customStyles"): "0", qn("w:latentStyles"): "0",
        qn("w:stylesInUse"): "0", qn("w:headingStyles"): "0", qn("w:numberingStyles"): "0",
        qn("w:tableStyles"): "0", qn("w:directFormattingOnRuns"): "0",
        qn("w:directFormattingOnParagraphs"): "0", qn("w:directFormattingOnNumbering"): "0",
        qn("w:directFormattingOnTables"): "0", qn("w:clearFormatting"): "1",
        qn("w:top3HeadingStyles"): "0", qn("w:visibleStyles"): "1",
        qn("w:alternateStyleNames"): "0",
    })
    _insert_setting(settings, flt)


# Mirrors srv/rights.go (5.22): keep the wording in step.
CC_LICENCES = {
    "cc_by": ("Attribution 4.0 International", "licenses/by/4.0/"),
    "cc_by_sa": ("Attribution-ShareAlike 4.0 International", "licenses/by-sa/4.0/"),
    "cc_by_nc": ("Attribution-NonCommercial 4.0 International", "licenses/by-nc/4.0/"),
    "cc_by_nc_sa": ("Attribution-NonCommercial-ShareAlike 4.0 International", "licenses/by-nc-sa/4.0/"),
    "cc_by_nd": ("Attribution-NoDerivatives 4.0 International", "licenses/by-nd/4.0/"),
    "cc_by_nc_nd": ("Attribution-NonCommercial-NoDerivatives 4.0 International", "licenses/by-nc-nd/4.0/"),
}


def rights_line(code, year, holder):
    """© line + licence sentence for p. iv, as srv/rights.go prints it."""
    code = (code or "").strip()
    who = " ".join(x for x in ((year or "").strip(), (holder or "").strip()) if x)
    cline = f"Copyright © {who}." if who else ""
    if code == "cc0":
        h = (holder or "").strip() or "The author"
        return (f"{h} has dedicated this work to the public domain under the Creative Commons "
                "CC0 1.0 Universal dedication. To view a copy, visit https://creativecommons.org/publicdomain/zero/1.0/")
    if code in CC_LICENCES:
        name, path = CC_LICENCES[code]
        lic = (f"This work is licensed under a Creative Commons {name} License. "
               f"To view a copy of this license, visit https://creativecommons.org/{path}")
        return f"{cline} {lic}" if cline else lic
    return f"{cline} All rights reserved." if cline else "All rights reserved."


def copyright_page_text(spec, author, body_font):
    """The copyright page (p. iv) as the build will set it, one line per
    element, from the transmittal's copyright-page builder (C12). Mirrors
    copyright-page-generated in series-template.typ; keep the two in step."""
    meta = spec.get("metadata", {}) or {}
    cover = spec.get("cover", {}) or {}

    def g(d, k):
        return str(d.get(k) or "").strip()

    lines = []
    title = g(meta, "title")
    if title:
        lines.append(title)
    year = g(meta, "copyright_year") or str(datetime.date.today().year)
    holder = g(meta, "copyright_holder") or author
    lines.append(rights_line(g(meta, "rights"), year, holder))
    publisher = g(meta, "publisher")
    city = g(meta, "publisher_city")
    if publisher:
        lines.append(f"Published by {publisher}, {city}." if city else f"Published by {publisher}.")
    for key in ("edition_line",):
        if g(meta, key):
            lines.append(g(meta, key))
    if g(cover, "credit"):
        lines.append(g(cover, "credit"))
    interior = g(meta, "interior_credit") or "Typeset by jdbb studio in {typeface}"
    lines.append(interior.replace("{typeface}", body_font or "Libertinus Serif"))
    for key in ("credit_lines", "loc_line"):
        if g(meta, key):
            lines.append(g(meta, key))
    if g(meta, "isbn_paper"):
        lines.append(f"ISBN {g(meta, 'isbn_paper')} (paperback)")
    if g(meta, "isbn_epub"):
        lines.append(f"ISBN {g(meta, 'isbn_epub')} (ebook)")
    if g(meta, "additional_notices"):
        lines.append(g(meta, "additional_notices"))
    if g(meta, "printed_in"):
        lines.append(g(meta, "printed_in"))
    return "\n".join(lines)


def _set_paragraph_style_font(style, font_name: str, size_pt: float,
                                bold: bool | None = None,
                                italic: bool | None = None):
    """Configure the font for a paragraph style."""
    font = style.font
    _ensure_font_exists(font, font_name)
    font.size = Pt(size_pt)
    if bold is not None:
        font.bold = bold
    if italic is not None:
        font.italic = italic


def _set_line_spacing_pt(style, spacing_pt: float):
    """Set exact line spacing in points on a paragraph style."""
    pf = style.paragraph_format
    pf.line_spacing = Pt(spacing_pt)
    # python-docx sets line_spacing_rule automatically when given Pt


def _set_first_line_indent(style, indent_pt: float):
    """Set first-line indent in points."""
    style.paragraph_format.first_line_indent = Pt(indent_pt)


def _set_left_indent(style, indent_pt: float):
    """Set left indent in points."""
    style.paragraph_format.left_indent = Pt(indent_pt)


def _set_contextual_spacing(style):
    """Word's 'Don't add space between paragraphs of the same style'."""
    ppr = style.element.get_or_add_pPr()
    if ppr.find(qn("w:contextualSpacing")) is None:
        ppr.append(OxmlElement("w:contextualSpacing"))


def _set_alignment(style, justify: bool):
    """Set alignment to JUSTIFY or LEFT."""
    style.paragraph_format.alignment = (
        WD_ALIGN_PARAGRAPH.JUSTIFY if justify else WD_ALIGN_PARAGRAPH.LEFT
    )


# ---------------------------------------------------------------------------
# Document builder
# ---------------------------------------------------------------------------

def build_template(spec: dict) -> Document:
    """Build and return a Document configured from *spec*."""
    doc = Document()

    meta = spec.get("metadata", {})
    typo = spec.get("typography", {})
    hdgs = spec.get("headings", {})
    elms = spec.get("elements", {})
    customs = spec.get("custom_styles") or []

    # Derived values
    # The book's real typefaces (named in the guide) vs the stand-ins the
    # template actually uses (see STANDIN_FONTS).
    book_body_font    = typo.get("body_font", "Libertinus Serif")
    book_heading_font = typo.get("heading_font", "Source Sans 3")
    book_code_font    = typo.get("code_font", "JetBrains Mono")
    body_font    = STANDIN_FONTS["body"]
    heading_font = STANDIN_FONTS["heading"]
    code_font    = STANDIN_FONTS["code"]
    base_size    = float(typo.get("base_size_pt", 10))
    leading      = float(typo.get("leading_pt", 2))
    line_sp      = base_size + leading
    indent_em    = float(typo.get("paragraph_indent_em", 0.75))
    indent_pt    = indent_em * base_size
    justify      = bool(typo.get("justify", True))

    # ------------------------------------------------------------------
    # 1. Normal (base body text) style
    # ------------------------------------------------------------------
    _set_page_to_trim(doc, spec.get("page") or {})

    normal = doc.styles["Normal"]
    _set_paragraph_style_font(normal, body_font, base_size)
    _set_black(normal)
    _set_line_spacing_pt(normal, line_sp)
    _set_first_line_indent(normal, indent_pt)
    _set_alignment(normal, justify)
    normal.paragraph_format.space_before = Pt(0)
    normal.paragraph_format.space_after = Pt(0)

    # ------------------------------------------------------------------
    # 2. First Paragraph (no indent, used after headings / breaks)
    # ------------------------------------------------------------------
    first_para = doc.styles.add_style("First Paragraph", WD_STYLE_TYPE.PARAGRAPH)
    first_para.base_style = normal
    _set_first_line_indent(first_para, 0)

    # ------------------------------------------------------------------
    # 3. Heading styles
    # ------------------------------------------------------------------
    heading_specs = [
        ("Heading 1", hdgs.get("h1_size_em", 1.667), hdgs.get("h1_weight", "bold")),
        ("Heading 2", hdgs.get("h2_size_em", 1.333), hdgs.get("h2_weight", 600)),
        ("Heading 3", hdgs.get("h3_size_em", 1.0),   hdgs.get("h3_weight", "medium")),
    ]
    for style_name, size_em, weight in heading_specs:
        style = doc.styles[style_name]
        size_pt = float(size_em) * base_size
        bold = _weight_to_bold(weight)
        _set_paragraph_style_font(style, heading_font, size_pt, bold=bold)
        _set_black(style)
        style.paragraph_format.space_before = Pt(base_size * 1.5)
        style.paragraph_format.space_after = Pt(base_size * 0.5)
        style.paragraph_format.first_line_indent = Pt(0)
        style.paragraph_format.alignment = WD_ALIGN_PARAGRAPH.LEFT
        # Keep with next
        style.paragraph_format.keep_with_next = True
        # The break belongs to what follows: a chapter starts on a new page
        # because its heading says so, never because the previous piece
        # padded its end with empty paragraphs (Inspect flags those).
        if style_name == "Heading 1":
            style.paragraph_format.page_break_before = True

    # ------------------------------------------------------------------
    # 4. Block Quote
    # ------------------------------------------------------------------
    bq_style = doc.styles.add_style("Block Quote", WD_STYLE_TYPE.PARAGRAPH)
    bq_style.base_style = normal
    bq_italic = elms.get("blockquote_style", "italic") == "italic"
    _set_paragraph_style_font(bq_style, body_font, base_size, italic=bq_italic)
    _set_left_indent(bq_style, base_size * 2)  # ~2em left indent
    _set_first_line_indent(bq_style, 0)
    bq_style.paragraph_format.space_before = Pt(base_size * 0.5)
    bq_style.paragraph_format.space_after = Pt(base_size * 0.5)

    # ------------------------------------------------------------------
    # 5. Code Block
    # ------------------------------------------------------------------
    code_size_em = float(elms.get("code_block_size_em", 0.8))
    code_size = code_size_em * base_size

    cb_style = doc.styles.add_style("Code Block", WD_STYLE_TYPE.PARAGRAPH)
    cb_style.base_style = normal
    _set_paragraph_style_font(cb_style, code_font, code_size)
    _set_left_indent(cb_style, base_size * 1.5)
    _set_first_line_indent(cb_style, 0)
    _set_alignment(cb_style, False)  # left-aligned
    _set_line_spacing_pt(cb_style, code_size + leading)
    cb_style.paragraph_format.space_before = Pt(base_size * 0.5)
    cb_style.paragraph_format.space_after = Pt(base_size * 0.5)

    # ------------------------------------------------------------------
    # 6. Section Break
    # ------------------------------------------------------------------
    sb_char_key = elms.get("section_break", "breve")
    sb_char = SECTION_BREAK_CHARS.get(sb_char_key, sb_char_key)

    sb_style = doc.styles.add_style("Section Break", WD_STYLE_TYPE.PARAGRAPH)
    sb_style.base_style = normal
    sb_style.paragraph_format.alignment = WD_ALIGN_PARAGRAPH.CENTER
    _set_first_line_indent(sb_style, 0)
    sb_style.paragraph_format.space_before = Pt(base_size)
    sb_style.paragraph_format.space_after = Pt(base_size)

    # ------------------------------------------------------------------
    # 7. Poem / Verse
    # ------------------------------------------------------------------
    poem_size_em = float(elms.get("poem_size_em", 0.75))
    poem_size = poem_size_em * base_size

    poem_style = doc.styles.add_style("Verse", WD_STYLE_TYPE.PARAGRAPH)
    poem_style.base_style = normal
    _set_paragraph_style_font(poem_style, body_font, poem_size, italic=True)
    _set_left_indent(poem_style, base_size * 2)
    _set_first_line_indent(poem_style, 0)
    _set_alignment(poem_style, False)
    _set_line_spacing_pt(poem_style, poem_size + leading)
    poem_style.paragraph_format.space_before = Pt(base_size * 0.25)
    poem_style.paragraph_format.space_after = Pt(base_size * 0.25)

    # ------------------------------------------------------------------
    # 8. Copyright
    # ------------------------------------------------------------------
    cr_style = doc.styles.add_style("Copyright", WD_STYLE_TYPE.PARAGRAPH)
    cr_style.base_style = normal
    cr_size = base_size * 0.8
    _set_paragraph_style_font(cr_style, body_font, cr_size)
    _set_first_line_indent(cr_style, 0)
    _set_alignment(cr_style, False)
    _set_line_spacing_pt(cr_style, cr_size + leading)

    # ------------------------------------------------------------------
    # 9. Epigraph
    # ------------------------------------------------------------------
    ep_style = doc.styles.add_style("Epigraph", WD_STYLE_TYPE.PARAGRAPH)
    ep_style.base_style = normal
    _set_paragraph_style_font(ep_style, body_font, base_size, italic=True)
    _set_left_indent(ep_style, base_size * 3)
    _set_first_line_indent(ep_style, 0)
    _set_alignment(ep_style, False)
    ep_style.paragraph_format.space_before = Pt(base_size * 0.5)
    ep_style.paragraph_format.space_after = Pt(base_size * 0.5)

    # ------------------------------------------------------------------
    # 9b. Signature — closes a foreword / afterword by another hand:
    #     name, title, place, one line each. No indent, flush left, a line
    #     of air above the block; contextual spacing keeps the lines tight.
    # ------------------------------------------------------------------
    sig_style = doc.styles.add_style("Signature", WD_STYLE_TYPE.PARAGRAPH)
    sig_style.base_style = normal
    _set_paragraph_style_font(sig_style, body_font, base_size)
    _set_first_line_indent(sig_style, 0)
    _set_alignment(sig_style, False)
    sig_style.paragraph_format.space_before = Pt(base_size)
    sig_style.paragraph_format.space_after = Pt(0)
    sig_style.paragraph_format.keep_together = True
    _set_contextual_spacing(sig_style)

    # ------------------------------------------------------------------
    # 9c. Glossary Entry — one paragraph per term: the term in bold, then
    #     its definition. Hanging indent so wrapped lines sit under the
    #     definition, not the term; a little space between entries.
    # ------------------------------------------------------------------
    gl_style = doc.styles.add_style("Glossary Entry", WD_STYLE_TYPE.PARAGRAPH)
    gl_style.base_style = normal
    _set_paragraph_style_font(gl_style, body_font, base_size)
    _set_left_indent(gl_style, base_size * 1.5)
    _set_first_line_indent(gl_style, -base_size * 1.5)
    _set_alignment(gl_style, False)
    gl_style.paragraph_format.space_before = Pt(0)
    gl_style.paragraph_format.space_after = Pt(base_size * 0.5)
    gl_style.paragraph_format.keep_together = True

    # ------------------------------------------------------------------
    # 10. Custom styles from spec
    # ------------------------------------------------------------------
    for cs in customs:
        cs_name = cs.get("word_style") or cs.get("name", "Custom")
        cs_type_str = cs.get("type", "paragraph").lower()
        # Skip if this style already exists (e.g. built-in or already created)
        if cs_name in [s.name for s in doc.styles]:
            continue
        if cs_type_str == "character":
            cs_style = doc.styles.add_style(cs_name, WD_STYLE_TYPE.CHARACTER)
            cs_style.base_style = doc.styles["Default Paragraph Font"]
        else:
            cs_style = doc.styles.add_style(cs_name, WD_STYLE_TYPE.PARAGRAPH)
            cs_style.base_style = normal
        # Set description if present (stored as style's element name hint)
        # python-docx doesn't expose style description directly, but we note
        # it in the sample content below.

    visible_styles = FACTORY_STYLES + [
        (cs.get("word_style") or cs.get("name", "Custom")) for cs in customs
    ]
    _tidy_styles_pane(doc, visible_styles)

    # ------------------------------------------------------------------
    # Sample content demonstrating each style
    # ------------------------------------------------------------------
    title = meta.get("title", "Book Title")
    author = meta.get("author", "Author Name")

    # Title page
    p = doc.add_heading(title, level=1)
    p = doc.add_paragraph(f"by {author}", style="First Paragraph")
    p.alignment = WD_ALIGN_PARAGRAPH.CENTER

    doc.add_paragraph()  # blank line

    # Instructions
    p = doc.add_heading("Template Guide", level=2)

    p = doc.add_paragraph(
        "This document is a styled Word template generated from your book "
        "specification. Each paragraph style below matches a Typst style used "
        "in final production. Use these styles consistently so the manuscript "
        "converts cleanly.",
        style="First Paragraph"
    )
    doc.add_paragraph(
        f"About the fonts: this template is set in {body_font}, {heading_font} and "
        f"{code_font} because every computer has them. They are stand-ins. "
        f"The book itself will be set in {book_body_font} (text), "
        f"{book_heading_font} (headings) and {book_code_font} (code) when it is built."
    )
    doc.add_paragraph(
        "The page is the book's trim size with the book's margins, so the line "
        "length you see is close to what the printed page will carry. "
        "The Styles pane (Home \u2192 Styles) lists only the styles below; if you "
        "see more, open its Options and choose \u201cRecommended\u201d."
    )

    # --- Front matter / back matter (P4 book map) ---
    p = doc.add_heading("How the book is assembled", level=2)
    doc.add_paragraph(
        "The factory reads your file in three zones. Everything before the first "
        "Heading 1 is front matter you typed yourself; each Heading 1 is a section; "
        "sections named like back matter at the end of the file are back matter. "
        "Page numbers, running heads and the Contents come from that reading.",
        style="First Paragraph"
    )
    doc.add_paragraph(
        "Do not type these — they are generated from your transmittal: the half-title, "
        "the title page, the copyright page and the table of contents. If you type "
        "a title, a \u201cby\u201d line, a Copyright-styled block or a Contents heading, "
        "the factory drops it and tells you so in Inspect. (The title and byline at the "
        "top of this template are dropped the same way.)"
    )
    doc.add_paragraph(
        "Do type these before your first Heading 1, each on its own page (Insert \u2192 "
        "Page Break): a dedication, then an epigraph if you have one. Use the "
        "\u2018Epigraph\u2019 style for the epigraph; the dedication is plain Normal text. "
        "They are set centred on their own pages, after the copyright page and before "
        "the Contents."
    )
    doc.add_paragraph(
        "Front matter with a heading uses Heading 1 like any chapter and is recognised "
        "by its name: Foreword, Preface, Prologue, Acknowledgments, Author\u2019s Note, "
        "Note on the Text, List of Illustrations, Chronology, Cast of Characters. "
        "These get roman page numbers. An Introduction is treated as body (page 1)."
    )
    doc.add_paragraph(
        "Back matter is recognised by name at the end of the file: Afterword, Appendix, "
        "Notes, Bibliography, Further Reading, Glossary, Index, About the Author, "
        "Acknowledgments, Credits, Colophon. An Epilogue is body. Anything else with a "
        "Heading 1 is a chapter."
    )
    doc.add_paragraph(
        "Every chapter starts with a Heading 1 \u2014 that is what makes a new page and a "
        "Contents entry. Heading 2 and Heading 3 divide a chapter; they never start a page."
    )

    # --- Normal / Body Text ---
    p = doc.add_heading("Body Text (Normal)", level=3)
    doc.add_paragraph(
        "This paragraph uses the Normal style — the default for body text. "
        f"Font: {body_font} at {base_size}pt with {line_sp}pt line spacing. "
        f"First-line indent: {indent_pt:.1f}pt. "
        f"{'Justified' if justify else 'Left-aligned'}.",
        style="First Paragraph"
    )
    doc.add_paragraph(
        "Subsequent paragraphs in the same section use Normal style with the "
        "first-line indent. Only the first paragraph after a heading or break "
        "should use 'First Paragraph' (no indent)."
    )

    # --- First Paragraph ---
    p = doc.add_heading("First Paragraph", level=3)
    doc.add_paragraph(
        "Apply 'First Paragraph' to the first paragraph after any heading or "
        "section break. It is identical to Normal but has no first-line indent.",
        style="First Paragraph"
    )

    # --- Block Quote ---
    p = doc.add_heading("Block Quote", level=3)
    doc.add_paragraph(
        "Use 'Block Quote' for extended quotations:",
        style="First Paragraph"
    )
    doc.add_paragraph(
        "This is a sample block quotation. It is indented from the left margin "
        "and set in italic (per your spec). Use this style for any quotation "
        "that runs longer than a few words.",
        style="Block Quote"
    )

    # --- Code Block ---
    p = doc.add_heading("Code Block", level=3)
    doc.add_paragraph(
        "Use 'Code Block' for terminal output, code listings, or monospaced content:",
        style="First Paragraph"
    )
    doc.add_paragraph(
        f"$ echo 'Hello, World!'\n"
        f"Hello, World!\n"
        f"# Font: {code_font} at {code_size:.1f}pt",
        style="Code Block"
    )

    # --- Section Break ---
    p = doc.add_heading("Section Break", level=3)
    doc.add_paragraph(
        "Use 'Section Break' to mark scene or section divisions. "
        f"Your spec uses the '{sb_char_key}' symbol:",
        style="First Paragraph"
    )
    doc.add_paragraph(sb_char, style="Section Break")
    doc.add_paragraph(
        "The paragraph after a section break should use 'First Paragraph'.",
        style="First Paragraph"
    )

    # --- Verse ---
    p = doc.add_heading("Verse / Poem", level=3)
    doc.add_paragraph(
        "Use 'Verse' for poetry or lyrics. Each line is a separate line within "
        "one paragraph (Shift+Enter for soft returns):",
        style="First Paragraph"
    )
    doc.add_paragraph(
        "Shall I compare thee to a summer's day?\n"
        "Thou art more lovely and more temperate.",
        style="Verse"
    )

    # --- Epigraph ---
    p = doc.add_heading("Epigraph", level=3)
    doc.add_paragraph(
        "Use 'Epigraph' for chapter-opening quotations:",
        style="First Paragraph"
    )
    doc.add_paragraph(
        "In the beginning was the Word.\n— John 1:1",
        style="Epigraph"
    )

    # --- Signature ---
    p = doc.add_heading("Signature", level=3)
    doc.add_paragraph(
        "A foreword or afterword by someone other than the author ends with a "
        "signature: their name, then title or affiliation, then place, one line "
        "each. Put each line in 'Signature' style, straight after the last "
        "paragraph. The factory sets the block flush left with a line of space "
        "above it, and keeps it together on one page:",
        style="First Paragraph"
    )
    for line in ("Ada Reader", "Founding Director, The Institute", "Whitehorse, Yukon"):
        doc.add_paragraph(line, style="Signature")

    # --- Glossary Entry ---
    p = doc.add_heading("Glossary Entry", level=3)
    doc.add_paragraph(
        "For a glossary, put each term and its definition in one paragraph in "
        "'Glossary Entry' style, with the term in bold at the start. Entries "
        "hang: wrapped lines tuck under the definition. Head the section "
        "'Glossary' in Heading 1 so the factory files it as back matter:",
        style="First Paragraph"
    )
    for term, defn in (("Transmittal", "The one-page record of what a book is and how it should be set; the factory generates the Word template from it."),
                       ("Trim", "The finished page size of a printed book, width by height, in inches.")):
        gp = doc.add_paragraph(style="Glossary Entry")
        gp.add_run(term).bold = True
        gp.add_run("  " + defn)

    # --- Copyright ---
    p = doc.add_heading("Copyright", level=3)
    doc.add_paragraph(
        "The copyright page is generated from your transmittal, so you do not need to "
        "type one. If you do, put it in 'Copyright' style: the factory recognises it "
        "and drops it rather than printing two. The style exists so a typed page still "
        "looks right in Word:",
        style="First Paragraph"
    )
    doc.add_paragraph(copyright_page_text(spec, author, book_body_font), style="Copyright")

    # --- Custom Styles ---
    if customs:
        p = doc.add_heading("Custom Styles", level=2)
        for cs in customs:
            cs_name = cs.get("word_style") or cs.get("name", "Custom")
            desc = cs.get("description", "No description provided.")
            cs_type_str = cs.get("type", "paragraph").lower()
            if cs_type_str == "character":
                p = doc.add_paragraph(style="First Paragraph")
                p.add_run(f"{cs_name} (character style): ").bold = True
                # Try to apply the character style to a sample run
                try:
                    r = p.add_run("sample text")
                    r.style = doc.styles[cs_name]
                except KeyError:
                    p.add_run("sample text")
                p.add_run(f" — {desc}")
            else:
                p = doc.add_paragraph(style="First Paragraph")
                p.add_run(f"{cs_name}: ").bold = True
                p.add_run(desc)
                # Add a sample paragraph in the style
                try:
                    doc.add_paragraph(
                        f"This paragraph uses the '{cs_name}' style.",
                        style=cs_name
                    )
                except KeyError:
                    pass

    # --- Style summary table ---
    p = doc.add_heading("Style Summary", level=2)

    style_rows = [
        ("Normal",          f"{body_font}, {base_size}pt",    "Default body text with indent"),
        ("First Paragraph", f"{body_font}, {base_size}pt",    "After headings/breaks (no indent)"),
        ("Heading 1",       f"{heading_font}, {float(hdgs.get('h1_size_em', 1.667)) * base_size:.1f}pt", "Chapter titles"),
        ("Heading 2",       f"{heading_font}, {float(hdgs.get('h2_size_em', 1.333)) * base_size:.1f}pt", "Sub-sections"),
        ("Heading 3",       f"{heading_font}, {float(hdgs.get('h3_size_em', 1.0)) * base_size:.1f}pt",   "Sub-sub-sections"),
        ("Block Quote",     f"{body_font}, {base_size}pt italic", "Extended quotations"),
        ("Code Block",      f"{code_font}, {code_size:.1f}pt",    "Code / terminal output"),
        ("Section Break",   f"Centered, '{sb_char_key}'",         "Scene / section divider"),
        ("Verse",           f"{body_font}, {poem_size:.1f}pt italic", "Poetry / lyrics"),
        ("Epigraph",        f"{body_font}, {base_size}pt italic", "Book or chapter epigraph"),
        ("Copyright",       f"{body_font}, {cr_size:.1f}pt",      "Typed copyright page (dropped; generated from transmittal)"),
        ("Signature",       f"{body_font}, {base_size}pt",        "Name / title / place closing a foreword or afterword"),
        ("Glossary Entry",  f"{body_font}, {base_size}pt, hanging", "One term (bold) + definition per paragraph"),
    ]
    for cs in customs:
        cs_name = cs.get("word_style") or cs.get("name", "Custom")
        cs_desc = cs.get("description", "")
        style_rows.append((cs_name, cs.get("type", "paragraph"), cs_desc))

    table = doc.add_table(rows=1, cols=3)
    table.style = "Table Grid"
    hdr = table.rows[0].cells
    hdr[0].text = "Style Name"
    hdr[1].text = "Font / Size"
    hdr[2].text = "Usage"
    for name, font_info, usage in style_rows:
        row = table.add_row().cells
        row[0].text = name
        row[1].text = font_info
        row[2].text = usage

    return doc


# ---------------------------------------------------------------------------
# CLI
# ---------------------------------------------------------------------------

def main():
    parser = argparse.ArgumentParser(
        description="Generate a Word template from a book spec JSON."
    )
    parser.add_argument(
        "--spec-file", "-s",
        help="Path to spec JSON file (reads stdin if omitted)",
    )
    parser.add_argument(
        "--output", "-o",
        help="Output .docx path (writes to stdout if omitted)",
    )
    args = parser.parse_args()

    # Read spec
    if args.spec_file:
        with open(args.spec_file, "r", encoding="utf-8") as f:
            spec = json.load(f)
    else:
        spec = json.load(sys.stdin)

    # Build document
    doc = build_template(spec)

    # Write output
    if args.output:
        doc.save(args.output)
    else:
        buf = io.BytesIO()
        doc.save(buf)
        sys.stdout.buffer.write(buf.getvalue())


if __name__ == "__main__":
    main()
