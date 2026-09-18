#!/usr/bin/env python3
"""
Edge Case Detection and Review System
Detects manual formatting in Word documents and presents them for review
"""

import argparse
import json
import re
import unicodedata
from pathlib import Path
from typing import List, Dict, Tuple
from docx import Document
from docx.enum.text import WD_COLOR_INDEX
from docx.shared import RGBColor
from datetime import datetime
import html
import importlib.util
import sys

# The [[style]] marker grammar and resolver live in apply-style-markers.py (the
# build's pre-pass); load them from the sibling file so Inspect and the build
# can never disagree. If the file is missing, marker detection is skipped.
def _load_style_markers_module():
    path = Path(__file__).resolve().parent / 'apply-style-markers.py'
    if not path.exists():
        return None
    try:
        spec = importlib.util.spec_from_file_location('apply_style_markers', path)
        mod = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(mod)
        return mod
    except Exception as exc:  # pragma: no cover - defensive
        print(f'warning: apply-style-markers.py not loadable ({exc}); skipping marker detection', file=sys.stderr)
        return None


STYLE_MARKERS = _load_style_markers_module()

# Monospace families: the one font change that usually means something (computer text).
MONOSPACE_FONT_RE = re.compile(
    r'consolas|courier|menlo|monaco|mono|source code|fira code|lucida console|inconsolata',
    re.IGNORECASE,
)


class EdgeCaseDetector:
    """Detects manual/local formatting that might need review"""
    
    ASCII_DENSE_RE = re.compile(r"[\\/_|()\[\]{}<>-]{3,}")
    CJK_RE = re.compile(r"[\u4E00-\u9FFF\u3040-\u30FF\u3400-\u4DBF]")
    THAI_RE = re.compile(r"[\u0E00-\u0E7F]")
    # Word defaults plus every house style generate-word-template.py ships in
    # the client's template. Anything the template hands the author must not
    # come back from Inspect as an "undeclared custom style".
    BUILTIN_PARAGRAPH_STYLES = {
        'normal', 'body text', 'first paragraph', 'heading 1', 'heading 2', 'heading 3',
        'heading 4', 'heading 5', 'heading 6', 'title', 'subtitle', 'quote', 'block quote',
        'list paragraph', 'list bullet', 'list number', 'caption',
        'code block', 'section break', 'verse', 'copyright', 'epigraph', 'signature',
        'glossary entry',
    }
    BUILTIN_CHARACTER_STYLES = {
        'default paragraph font', 'strong', 'emphasis', 'subtle emphasis', 'intense emphasis'
    }

    def __init__(self, doc_path: str, declared_styles: List[Dict] | None = None):
        self.doc = Document(doc_path)
        self.edge_cases = []
        self.doc_path = doc_path
        self.declared_styles = declared_styles or []
        self.declared_style_names = {
            self._normalize_style_name((item.get('word_style') or item.get('name') or ''))
            for item in self.declared_styles
            if self._normalize_style_name((item.get('word_style') or item.get('name') or ''))
        }
        self.template_guide_cutoff = self._find_template_guide_cutoff()

    def _paragraph_context(self, index: int) -> Dict:
        context_before = []
        context_after = []
        for offset in (-1, -2):
            pos = index + offset
            if 0 <= pos < len(self.doc.paragraphs):
                text = (self.doc.paragraphs[pos].text or '').strip()
                if text:
                    context_before.append(text[:120])
        context_before.reverse()
        for offset in (1, 2):
            pos = index + offset
            if 0 <= pos < len(self.doc.paragraphs):
                text = (self.doc.paragraphs[pos].text or '').strip()
                if text:
                    context_after.append(text[:120])
        return {
            'context_before': context_before,
            'context_after': context_after,
        }
        
    def detect_all(self) -> List[Dict]:
        """Run all detection methods"""
        self.detect_manual_formatting()
        self.detect_style_markers()
        self.detect_unusual_fonts()
        self.detect_colored_text()
        self.detect_manual_lists()
        self.detect_manual_breaks()
        self.detect_stray_quote_markers()
        self.detect_blank_line_breaks()
        self.detect_heading_lookalikes()
        self.detect_mixed_styles()
        self.detect_direct_formatting()
        self.detect_image_inventory()
        self.detect_observed_styles()
        self.detect_special_typography()
        self.detect_language_scripts()
        return self.edge_cases
    
    def detect_manual_formatting(self):
        """Detect bold/italic applied directly (not via character style)"""
        for i, para in enumerate(self.doc.paragraphs):
            if self._should_skip_paragraph(para):
                continue
            for run in para.runs:
                issues = []
                
                # Check for direct bold (not from style)
                if run.bold and not self._is_from_style(run, 'bold'):
                    issues.append("manual bold")
                    
                # Check for direct italic (not from style)  
                if run.italic and not self._is_from_style(run, 'italic'):
                    issues.append("manual italic")
                    
                # Check for direct underline
                if run.underline:
                    issues.append("manual underline")
                    
                if issues:
                    self.edge_cases.append({
                        'type': 'manual_formatting',
                        'location': f'Paragraph {i+1}',
                        'text': run.text[:100],
                        'issues': issues,
                        'severity': 'medium',
                        'auto_decision': 'preserve',
                        'suggestion': 'Manual emphasis detected — preserve it as intentional bold/italic formatting in EPUB and Typst'
                    })
    
    def detect_style_markers(self):
        """[[style:name]] markers — the Google-Docs stand-in for custom paragraph
        styles. Same grammar and resolver as apply-style-markers.py (imported
        above); one finding per marker occurrence."""
        if STYLE_MARKERS is None:
            return
        resolver = STYLE_MARKERS.StyleResolver(self.declared_styles)
        found = STYLE_MARKERS.scan_markers(self.doc.paragraphs, resolver)
        for e in found:
            start, end = e['start'] + 1, e['end'] + 1
            para = self.doc.paragraphs[e['start']]
            text = STYLE_MARKERS.paragraph_runs_text(para)[e['open_len']:]
            if e['close_index'] == e['start'] and e.get('close_strip'):
                text = text[:len(text) - e['close_strip']]
            text = text.strip()[:100]
            resolved = e['resolved']
            count = e['paragraphs']
            location = f'Paragraph {start}' if start == end else f'Paragraphs {start}-{end}'
            if resolved:
                closer = resolved.lower()
                suggestion = (
                    f"Becomes {resolved} (via {e['via']} \u2018{e['marker']}\u2019), "
                    f"{count} {'paragraph' if count == 1 else 'paragraphs'}."
                )
                if count == 1:
                    suggestion += f" To cover a run, put [[/{closer}]] at the end of the last paragraph."
                severity = 'low'
            else:
                suggestion = (
                    "Not a factory style and not declared on the transmittal \u2014 will print as written. "
                    f"Use one of: {STYLE_MARKERS.SUGGESTED_NAMES}, or declare \u2018{e['marker']}\u2019 "
                    "as a custom style on the transmittal."
                )
                severity = 'medium'
            self.edge_cases.append({
                'type': 'style_marker',
                'location': location,
                'text': text,
                'marker': e['marker'],
                'resolved': resolved,
                'paragraphs': count,
                'severity': severity,
                'suggestion': suggestion,
            })

    def detect_unusual_fonts(self):
        """One finding per non-standard font: where, how much, and what (if
        anything) to do about it. The factory sets everything in the book face,
        so a font change only matters if it carries meaning (computer text).
        Emoji/symbol fonts are aggregated as before."""
        standard_fonts = {
            'Calibri', 'Times New Roman', 'Arial', 'Libertinus Serif',
            'Source Sans 3', 'JetBrains Mono', 'Cambria', 'Georgia'
        }
        unusual_runs = []
        for i, para in enumerate(self.doc.paragraphs):
            if self._should_skip_paragraph(para):
                continue
            for run in para.runs:
                if run.font.name and run.font.name not in standard_fonts:
                    unusual_runs.append((i + 1, run))

        grouped = {}
        for location, run in unusual_runs:
            font_name = run.font.name
            text = (run.text or '').strip()
            key = font_name
            grouped.setdefault(key, []).append((location, text))

        for font_name, entries in grouped.items():
            nonempty_samples = [text for _, text in entries if text]
            locations = sorted({loc for loc, _ in entries})
            if self._font_should_be_aggregated(font_name, nonempty_samples):
                self.edge_cases.append({
                    'type': 'font_treatment',
                    'subtype': 'emoji_font_usage',
                    'location': f'Paragraphs {locations[0]}-{locations[-1]}' if len(locations) > 1 else f'Paragraph {locations[0]}',
                    'text': '; '.join(nonempty_samples[:5]) if nonempty_samples else font_name,
                    'font': font_name,
                    'severity': 'low',
                    'suggestion': 'Emoji/special font usage detected — absorb as character-style treatment and verify font survival in EPUB and PDF'
                })
                continue
            n_paras = len(locations)
            n_runs = len(entries)
            para_range = (f'\u00b6{locations[0]}\u2013{locations[-1]}' if n_paras > 1 else f'\u00b6{locations[0]}')
            location = f'Paragraphs {locations[0]}-{locations[-1]}' if n_paras > 1 else f'Paragraph {locations[0]}'
            sample = next((t for t in nonempty_samples if len(t) > 3 and not t.startswith('[[')),
                          nonempty_samples[0] if nonempty_samples else font_name)
            if MONOSPACE_FONT_RE.search(font_name):
                severity = 'medium'
                suggestion = (
                    f"Monospace, {n_paras} {'paragraph' if n_paras == 1 else 'paragraphs'} ({para_range}). "
                    "Computer text? Put [[code block]] at the start of the first paragraph and [[/code block]] "
                    "at the end of the last, or use the Code Block style in Word. Otherwise ignore: the factory "
                    "sets everything in the book face."
                )
            elif n_runs <= 2:
                severity = 'low'
                suggestion = 'Looks like a paste. Harmless: the factory sets everything in the book face.'
            else:
                severity = 'low'
                suggestion = (
                    'The factory sets everything in the book face; a font change only matters if it means '
                    'something (computer text, a foreign script). Say so with a [[style]] marker or on the transmittal.'
                )
            self.edge_cases.append({
                'type': 'unusual_font',
                'location': location,
                'text': sample[:100],
                'font': font_name,
                'paragraphs': n_paras,
                'runs': n_runs,
                'severity': severity,
                'suggestion': suggestion,
            })
    
    def detect_colored_text(self):
        """Detect colored text (non-black)"""
        for i, para in enumerate(self.doc.paragraphs):
            if self._should_skip_paragraph(para):
                continue
            for run in para.runs:
                # Check font color
                if run.font.color and run.font.color.rgb:
                    rgb = run.font.color.rgb
                    if rgb != RGBColor(0, 0, 0):  # Not black
                        # Near-black greys (Google Docs exports paint body
                        # text RGB(68,68,68); Word "Text 1 lighter" is
                        # similar) are not a colour the author chose — the
                        # build sets them black anyway. Reserve high for
                        # actual colours that would vanish in print.
                        lum = 0.2126 * rgb[0] + 0.7152 * rgb[1] + 0.0722 * rgb[2]
                        spread = max(rgb[0], rgb[1], rgb[2]) - min(rgb[0], rgb[1], rgb[2])
                        # 90/255 ≈ 35 % keeps Google Docs' RGB(68,68,68) and Word's
                        # "Text 1, lighter 25%" (RGB(64,64,64)) in; real greys stay high.
                        near_black = lum < 90 and spread < 24
                        self.edge_cases.append({
                            'type': 'colored_text',
                            'location': f'Paragraph {i+1}',
                            'text': run.text[:100],
                            'color': f'RGB({rgb[0]},{rgb[1]},{rgb[2]})',
                            'severity': 'low' if near_black else 'high',
                            'suggestion': ('Dark grey text (near black) - the build sets it black; nothing to do'
                                           if near_black else
                                           'Colored text detected - will be removed for print')
                        })
                
                # Check highlight color
                if run.font.highlight_color and run.font.highlight_color != WD_COLOR_INDEX.AUTO:
                    self.edge_cases.append({
                        'type': 'highlighted_text',
                        'location': f'Paragraph {i+1}',
                        'text': run.text[:100],
                        'severity': 'medium',
                        'suggestion': 'Highlighted text detected - review purpose'
                    })
    
    def detect_manual_lists(self):
        """Detect manually formatted lists (using -, *, •, numbers)"""
        list_patterns = [
            (r'^\s*[-*•]\s+', 'bullet'),
            (r'^\s*\d+[\.)]\s+', 'numbered'),
            (r'^\s*[a-zA-Z][\.)]\s+', 'lettered')
        ]
        
        import re
        for i, para in enumerate(self.doc.paragraphs):
            if self._should_skip_paragraph(para):
                continue
            text = para.text.strip()
            for pattern, list_type in list_patterns:
                if re.match(pattern, text):
                    self.edge_cases.append({
                        'type': 'manual_list',
                        'location': f'Paragraph {i+1}',
                        'text': text[:100],
                        'list_type': list_type,
                        'severity': 'low',
                        'suggestion': f'Manual {list_type} list - convert to proper list style'
                    })
                    break
    
    def detect_manual_breaks(self):
        """Detect manual section breaks (*, * * *, ---, ___, etc.)"""
        break_patterns = [
            r'^\s*\*\s*\*\s*\*\s*$',
            r'^\s*[-_]{3,}\s*$',
            r'^\s*[•·∗]{3,}\s*$',
            r'^\s*\*\s*$',
            r'^#{3,}\s*$'
        ]
        
        import re
        for i, para in enumerate(self.doc.paragraphs):
            if self._should_skip_paragraph(para):
                continue
            text = para.text.strip()
            for pattern in break_patterns:
                if re.match(pattern, text):
                    self.edge_cases.append({
                        'type': 'manual_break',
                        'location': f'Paragraph {i+1}',
                        'text': text,
                        'severity': 'low',
                        'suggestion': 'Manual section break - convert to proper style or spacing'
                    })
                    break
    
    def detect_stray_quote_markers(self):
        """Detect '>' quote markers inside running prose (email / Markdown
        residue pasted into the manuscript). One finding per paragraph."""
        import re
        pat = re.compile(r'(?:^|\s)>\s')
        for i, para in enumerate(self.doc.paragraphs):
            if self._should_skip_paragraph(para):
                continue
            text = para.text
            n = len(pat.findall(text))
            if n == 0 or len(text.strip()) < 20:
                continue
            m = pat.search(text)
            start = max(0, m.start() - 30)
            self.edge_cases.append({
                'type': 'stray_quote_marker',
                'location': f'Paragraph {i+1}',
                'text': text[start:start + 80].strip(),
                'severity': 'medium',
                'count': n,
                'suggestion': f'{n} stray ">" marker(s) in running text — quoted-email or Markdown residue; delete them'
            })

    def detect_blank_line_breaks(self):
        """Two or more empty paragraphs in a row between text: a scene break
        faked with the Return key. Blank paragraphs vanish in the build (and
        never survive a page break), so the break is lost unless it becomes
        a 'Section Break' paragraph."""
        paras = self.doc.paragraphs
        i = 0
        seen_text = False
        while i < len(paras):
            if self._should_skip_paragraph(paras[i]) or paras[i].text.strip():
                seen_text = seen_text or bool(paras[i].text.strip())
                i += 1
                continue
            j = i
            while j < len(paras) and not paras[j].text.strip():
                j += 1
            run_len = j - i
            if run_len >= 2 and seen_text and j < len(paras):
                self.edge_cases.append({
                    'type': 'manual_break',
                    'location': f'Paragraph {i+1}',
                    'text': f'({run_len} empty paragraphs)',
                    'severity': 'low',
                    'suggestion': 'Blank lines used as a scene break - empty paragraphs are dropped in the build; use the Section Break style instead',
                    **self._paragraph_context(i),
                })
            i = j

    def detect_heading_lookalikes(self):
        """Body-styled paragraphs that look like headings: short, bold
        throughout or set larger than the body text, or starting 'Chapter N' /
        'Part N'. The factory only finds chapter openers through Heading
        styles, so these are the single most damaging thing a
        written-outside-the-template manuscript can carry."""
        import re
        heading_words = re.compile(r'^(chapter|part|book|prologue|epilogue|interlude)\b', re.I)
        try:
            base_size = self.doc.styles['Normal'].font.size
        except Exception:
            base_size = None
        base_pt = base_size.pt if base_size else 12.0

        uses_heading_styles = False
        candidates = []
        for i, para in enumerate(self.doc.paragraphs):
            if self._should_skip_paragraph(para):
                continue
            style = getattr(getattr(para, 'style', None), 'name', '') or ''
            snorm = self._normalize_style_name(style)
            if snorm.startswith('heading'):
                uses_heading_styles = True
                continue
            if snorm in ('title', 'subtitle'):
                continue
            text = para.text.strip()
            if not text or len(text) > 90 or text.endswith(('.', ',', ';')) and not heading_words.match(text):
                continue
            runs = [r for r in para.runs if r.text.strip()]
            if not runs:
                continue
            all_bold = all(r.bold for r in runs)
            max_pt = max((r.font.size.pt for r in runs if r.font.size), default=0)
            larger = max_pt >= base_pt + 3
            named = bool(heading_words.match(text))
            if not (all_bold or larger or named):
                continue
            why = []
            if all_bold:
                why.append('bold')
            if larger:
                why.append(f'{max_pt:g}pt vs {base_pt:g}pt body')
            if named:
                why.append('starts like a chapter title')
            candidates.append((i, text, style, why))

        for i, text, style, why in candidates:
            self.edge_cases.append({
                'type': 'heading_lookalike',
                'location': f'Paragraph {i+1}',
                'text': text[:100],
                'issues': why,
                'style_name': style,
                'severity': 'high' if not uses_heading_styles else 'medium',
                'suggestion': (
                    f"Looks like a heading but is styled '{style}' ({', '.join(why)}). "
                    + ('The manuscript uses no Heading styles at all, so the factory will find no chapters. ' if not uses_heading_styles else '')
                    + 'Apply Heading 1 (chapter) or Heading 2 (section) so the build can start chapters on a new page and list them in the EPUB contents.'
                ),
            })

    def detect_mixed_styles(self):
        """Detect paragraphs with multiple fonts/sizes"""
        for i, para in enumerate(self.doc.paragraphs):
            if self._should_skip_paragraph(para):
                continue
            fonts = set()
            sizes = set()
            
            for run in para.runs:
                if run.font.name:
                    fonts.add(run.font.name)
                if run.font.size:
                    sizes.add(run.font.size)
            
            if len(fonts) > 1 or len(sizes) > 1:
                self.edge_cases.append({
                    'type': 'mixed_formatting',
                    'location': f'Paragraph {i+1}',
                    'text': para.text[:100],
                    'fonts': list(fonts),
                    'sizes': [str(s) for s in sizes],
                    'severity': 'medium',
                    'suggestion': 'Multiple fonts/sizes in one paragraph - review intent'
                })
    
    def detect_direct_formatting(self):
        """Detect spacing, indentation, alignment issues"""
        for i, para in enumerate(self.doc.paragraphs):
            if self._should_skip_paragraph(para):
                continue
            issues = []
            
            # Check for manual spacing
            if para.paragraph_format.space_before:
                issues.append(f"space before: {para.paragraph_format.space_before}")
            if para.paragraph_format.space_after:
                issues.append(f"space after: {para.paragraph_format.space_after}")
                
            # Check for manual indentation (not from style)
            if para.paragraph_format.first_line_indent:
                issues.append(f"first line indent: {para.paragraph_format.first_line_indent}")
            if para.paragraph_format.left_indent:
                issues.append(f"left indent: {para.paragraph_format.left_indent}")
                
            # Check for center/right alignment
            if para.alignment and str(para.alignment) != "LEFT":
                issues.append(f"alignment: {para.alignment}")
                
            if issues:
                self.edge_cases.append({
                    'type': 'direct_spacing',
                    'location': f'Paragraph {i+1}',
                    'text': para.text[:100],
                    'issues': issues,
                    'severity': 'low',
                    'suggestion': 'Direct paragraph formatting applied (spacing, indent, or alignment) - review whether this should become a named paragraph style instead'
                })
    
    def detect_image_inventory(self):
        """Emit one finding per inline image so preflight can surface image counts."""
        for i, shape in enumerate(self.doc.inline_shapes):
            width_in = round(shape.width / 914400, 2) if getattr(shape, 'width', None) else None
            height_in = round(shape.height / 914400, 2) if getattr(shape, 'height', None) else None
            colour = self._image_is_colour(shape)
            finding = {
                'type': 'image_inventory',
                'location': f'Image {i+1}',
                'text': f'Inline image {i+1}' + (' (colour)' if colour else ' (black and white)' if colour is False else ''),
                'image_index': i + 1,
                'width_in': width_in,
                'height_in': height_in,
                'severity': 'low',
                'suggestion': ('Colour: kept in the EPUB, converted to grey for the print PDF (auto-level, gentle S-curve). '
                               'Place your own grey version in Word if this one needs a careful hand.'
                               if colour else 'Review caption / alt text / placement for this image')
            }
            if colour is not None:
                finding['colour'] = colour
            self.edge_cases.append(finding)

    def _image_is_colour(self, shape):
        """True/False from mean HSL saturation via ImageMagick; None if unknown."""
        try:
            rid = shape._inline.graphic.graphicData.pic.blipFill.blip.embed
            blob = self.doc.part.related_parts[rid].blob
        except Exception:
            return None
        try:
            import subprocess
            # Same recipe as srv/images.go colourArgs: share of pixels with clear chroma.
            out = subprocess.run(['convert', '-[0]', '-resize', '400x400>', '-colorspace', 'sRGB',
                                  '-fx', 'max(r,g,b)-min(r,g,b)', '-threshold', '10%',
                                  '-format', '%[fx:mean]', 'info:'], input=blob, capture_output=True, timeout=60)
            return float(out.stdout.decode().strip()) >= 0.005
        except Exception:
            return None

    def detect_observed_styles(self):
        """Inventory observed manuscript styles, but stay quiet for obvious built-ins."""
        seen = set()
        emoji_runs = []
        for i, para in enumerate(self.doc.paragraphs):
            if self._should_skip_paragraph(para):
                continue
            para_text = para.text.strip()
            para_style = getattr(getattr(para, 'style', None), 'name', None)
            para_norm = self._normalize_style_name(para_style or '')
            if para_style and para_norm not in self.BUILTIN_PARAGRAPH_STYLES:
                key = ('paragraph', para_style)
                if key not in seen:
                    seen.add(key)
                    auto_decision = 'preserve' if para_norm in self.declared_style_names else None
                    suggestion = (
                        'Declared paragraph style is already present in the manuscript — preserve it unless treatment needs to change'
                        if auto_decision == 'preserve'
                        else 'Observed paragraph style in manuscript; confirm whether it should be declared intentionally in the project spec'
                    )
                    self.edge_cases.append({
                        'type': 'observed_style',
                        'style_kind': 'paragraph',
                        'style_name': para_style,
                        'location': f'Paragraph {i+1}',
                        'text': para_text[:100] if para_text else para_style,
                        'severity': 'low',
                        'auto_decision': auto_decision,
                        'suggestion': suggestion,
                        **self._paragraph_context(i),
                    })
                    self._append_declaration_finding('paragraph', para_style, i, para_text, auto_decision == 'preserve')
            for run in para.runs:
                run_text = run.text.strip()
                run_style = getattr(getattr(run, 'style', None), 'name', None)
                run_norm = self._normalize_style_name(run_style or '')
                if run_style and run_norm not in self.BUILTIN_CHARACTER_STYLES:
                    key = ('character', run_style)
                    if key not in seen:
                        seen.add(key)
                        auto_decision = 'preserve' if run_norm in self.declared_style_names else None
                        suggestion = (
                            'Declared character style is already present in the manuscript — preserve it unless treatment needs to change'
                            if auto_decision == 'preserve'
                            else 'Observed character style in manuscript; confirm whether it should be declared intentionally in the project spec'
                        )
                        self.edge_cases.append({
                            'type': 'observed_style',
                            'style_kind': 'character',
                            'style_name': run_style,
                            'location': f'Paragraph {i+1}',
                            'text': run_text[:100] if run_text else run_style,
                            'severity': 'low',
                            'auto_decision': auto_decision,
                            'suggestion': suggestion,
                            **self._paragraph_context(i),
                        })
                        self._append_declaration_finding('character', run_style, i, run_text, auto_decision == 'preserve')
                if self._looks_like_emoji_font(run):
                    emoji_runs.append((i + 1, run.font.name or 'emoji font', run_text[:40]))

    def _append_declaration_finding(self, kind: str, style_name: str, index: int, text: str, declared: bool):
        """Mirror of srv/preflight.go appendUndeclaredStyleWarnings, so the HTML
        report the author reads carries the same high-severity 'undeclared
        custom style' card the JSON summary counts (the Go side dedupes on
        style name and adds nothing when the finding is already here)."""
        norm = self._normalize_style_name(style_name)
        if norm in ('normal', 'default paragraph font', ''):
            return
        if declared:
            self.edge_cases.append({
                'type': 'declared_custom_style_used',
                'style_name': style_name,
                'style_kind': kind,
                'location': f'Paragraph {index+1}',
                'text': text[:100] if text else style_name,
                'severity': 'low',
                'suggestion': 'Declared custom style is present in the manuscript and ready for intentional EPUB/Typst handling.',
            })
        else:
            self.edge_cases.append({
                'type': 'undeclared_custom_style',
                'style_name': style_name,
                'style_kind': kind,
                'location': f'Paragraph {index+1}',
                'text': text[:100] if text else style_name,
                'severity': 'high',
                'suggestion': 'Custom style is used in the manuscript but not declared in the project spec/transmittal. Add it before production so EPUB and Typst can treat it intentionally.',
            })

    def detect_special_typography(self):
        """Detect likely spacing-sensitive or preformatted blocks such as ASCII art."""
        current_block = []
        current_start = None

        def flush_block():
            nonlocal current_block, current_start
            if not current_block:
                return
            block_text = '\n'.join(current_block)
            match = self._best_declared_style_match(block_text)
            auto_decision = 'preserve' if match else None
            severity = 'low' if match else 'high'
            if match:
                suggestion = (
                    f"Likely spacing-sensitive ASCII / preformatted block is already assigned to declared style "
                    f"\"{match}\" — preserve as intentional special typography. Only review further if spacing is still local formatting."
                )
            else:
                suggestion = 'Likely spacing-sensitive ASCII / preformatted block — preserve as intentional special typography'
            self.edge_cases.append({
                'type': 'special_typography',
                'subtype': 'ascii_art_candidate',
                'location': f'Paragraph {current_start}' if len(current_block) == 1 else f'Paragraphs {current_start}-{current_start + len(current_block) - 1}',
                'text': block_text[:200],
                'line_count': len(current_block),
                'declared_style_match': match,
                'severity': severity,
                'auto_decision': auto_decision,
                'suggestion': suggestion,
            })
            current_block = []
            current_start = None

        for i, para in enumerate(self.doc.paragraphs, start=1):
            if self._should_skip_paragraph(para):
                flush_block()
                continue
            text = para.text.rstrip('\n')
            if text.strip() and self._looks_like_ascii_art(text):
                if current_start is None:
                    current_start = i
                current_block.append(text)
            else:
                flush_block()

        flush_block()

    def detect_language_scripts(self):
        """Report non-Latin scripts so production can confirm language handling choices."""
        seen = set()
        for i, para in enumerate(self.doc.paragraphs):
            if self._should_skip_paragraph(para):
                continue
            text = para.text.strip()
            if not text:
                continue
            if self._looks_like_ascii_art(text) or self._is_box_drawing_heavy(text):
                continue
            script = self._detect_script(text)
            if not script:
                continue
            key = (script, text[:40])
            if key in seen:
                continue
            seen.add(key)
            self.edge_cases.append({
                'type': 'language_script',
                'location': f'Paragraph {i+1}',
                'text': text[:100],
                'script': script,
                'severity': 'medium',
                'suggestion': 'Non-Latin script detected — preserve text and verify language-specific EPUB/Typst handling'
            })

    def _looks_like_ascii_art(self, text: str) -> bool:
        stripped = text.rstrip()
        if len(stripped) < 6:
            return False
        if '  ' not in stripped and not stripped.startswith(' '):
            return False
        symbol_chars = sum(1 for ch in stripped if not ch.isalnum() and not ch.isspace())
        if symbol_chars < 2:
            return False
        if self.ASCII_DENSE_RE.search(stripped):
            return True
        unique_symbols = {ch for ch in stripped if not ch.isalnum() and not ch.isspace()}
        if len(unique_symbols) >= 2 and symbol_chars >= max(2, len(stripped) // 5):
            return True
        compact = stripped.strip()
        return any(marker in compact for marker in ('||', '//', '\\\\', '/_/', '(_', '_/')) and symbol_chars >= 2

    def _detect_script(self, text: str):
        if self.THAI_RE.search(text):
            return 'thai'
        if self.CJK_RE.search(text):
            return 'cjk'
        return None

    def _is_box_drawing_heavy(self, text: str) -> bool:
        significant = [ch for ch in text if ch.strip()]
        if not significant:
            return False
        decorative = 0
        for ch in significant:
            codepoint = ord(ch)
            if (
                0x2500 <= codepoint <= 0x257F or
                0x2580 <= codepoint <= 0x259F or
                0xFFE0 <= codepoint <= 0xFFEE or
                codepoint == 0xFF3F or
                unicodedata.name(ch, '').startswith('BOX DRAWINGS') or
                unicodedata.name(ch, '').startswith('FULLWIDTH')
            ):
                decorative += 1
        return decorative >= max(2, len(significant) // 4)

    def _normalize_style_name(self, name: str) -> str:
        return (name or '').strip().lower()

    def _find_template_guide_cutoff(self):
        start = None
        for idx, para in enumerate(self.doc.paragraphs):
            text = (para.text or '').strip()
            if start is None and text == 'Template Guide':
                start = idx
            if start is not None and text == 'Style Summary':
                return idx + 3
        return None

    def _is_template_guide_document(self) -> bool:
        texts = [p.text.strip() for p in self.doc.paragraphs[:40] if p.text.strip()]
        joined = '\n'.join(texts)
        markers = [
            'Template Guide',
            'This document is a styled Word template generated from your book specification.',
            'Custom Styles',
            'Body Text (Normal)',
            'Use \'Section Break\' to mark scene or section divisions.'
        ]
        hits = sum(1 for marker in markers if marker in joined)
        return hits >= 2

    def _should_skip_paragraph(self, para) -> bool:
        text = (para.text or '').strip()
        if self.template_guide_cutoff is not None:
            try:
                idx = self.doc.paragraphs.index(para)
                if idx <= self.template_guide_cutoff:
                    return True
            except ValueError:
                pass
        if not text:
            return False
        extra_guide_markers = (
            "This paragraph uses the 'tweet-p' style.",
            "This paragraph uses the 'metadata-p' style.",
            'metadata-c (character style):',
            'metadata-p: tweet metadata if standalone',
            '$ echo \'Hello, World!\'',
            'Shall I compare thee to a summer\'s day?',
            'In the beginning was the Word.',
        )
        return any(marker in text for marker in extra_guide_markers)

    def _best_declared_style_match(self, text: str):
        for candidate in sorted(self.declared_style_names):
            if 'ascii' in candidate:
                return candidate
        return None

    def _font_should_be_aggregated(self, font_name: str, samples: List[str]) -> bool:
        lowered = (font_name or '').lower()
        if 'emoji' in lowered or 'symbol' in lowered:
            return True
        if not samples:
            return False
        emoji_like = 0
        for sample in samples:
            for ch in sample:
                if not ch.strip():
                    continue
                if ord(ch) > 0x2600 or unicodedata.category(ch).startswith('So'):
                    emoji_like += 1
        return emoji_like >= max(1, len(''.join(samples)) // 4)

    def _looks_like_emoji_font(self, run) -> bool:
        font_name = (run.font.name or '').lower()
        if not font_name:
            return False
        return 'emoji' in font_name or 'symbol' in font_name

    def _is_from_style(self, run, attr: str) -> bool:
        """Check if formatting comes from a style (simplified)"""
        # This is a simplified check - would need more logic for full detection
        return False


class EdgeCaseReviewer:
    """Review-only interface for edge cases"""

    # Section metadata: (label, description, render_order_hint)
    SECTION_META = {
        'undeclared_custom_style': ('Undeclared Custom Styles', 'Custom styles used in manuscript but not declared in production spec'),
        'heading_lookalike': ('Headings Without Heading Styles', 'Paragraphs that look like chapter or section headings (bold, larger, "Chapter N") but carry a body style — the factory cannot find them'),
        'language_script': ('Language / Script', 'Non-Latin script content requiring special font or language handling'),
        'special_typography': ('Special Typography', 'Spacing-sensitive or preformatted content (ASCII art, code blocks)'),
        'observed_style': ('Observed Styles', 'Non-built-in Word styles found in the manuscript'),
        'declared_custom_style_used': ('Declared Custom Styles', 'Declared custom styles confirmed present in the manuscript'),
        'font_treatment': ('Font Treatment', 'Emoji or special font usage detected'),
        'manual_formatting': ('Manual Formatting', 'Bold, italic, or underline applied directly to runs rather than via character styles'),
        'manual_list': ('Manual Lists', 'Text formatted as lists using characters (-, *, 1.) rather than Word list styles'),
        'direct_spacing': ('Direct Paragraph Formatting', 'Paragraph-level spacing, indentation, or alignment applied directly rather than via styles'),
        'image_inventory': ('Image Inventory', 'Inline images found in the manuscript'),
        'colored_text': ('Colored Text', 'Non-black text color detected'),
        'highlighted_text': ('Highlighted Text', 'Highlighted text detected'),
        'unusual_font': ('Unusual Fonts', 'Fonts are ignored by the factory — everything is set in the book face. One entry per font: worth a look only if the change means something (computer text, a foreign script)'),
        'style_marker': ('Marked styles', '[[style:name]] markers at the start of a paragraph — the Google-Docs way to ask for a factory style (code block, block quote, verse…). Resolved markers are applied and removed by the build; unresolved ones print as written'),
        'manual_break': ('Manual Breaks', 'Manual section break characters detected'),
        'stray_quote_marker': ('Stray Quote Markers', '">" characters inside running text (quoted-email or Markdown residue)'),
        'mixed_formatting': ('Mixed Formatting', 'Multiple fonts or sizes within a single paragraph'),
    }

    # Preferred display order (high-severity / small-count first)
    SECTION_ORDER_PRIORITY = [
        'undeclared_custom_style', 'style_marker', 'heading_lookalike', 'language_script', 'special_typography',
        'colored_text', 'highlighted_text', 'unusual_font',
        'observed_style', 'declared_custom_style_used', 'font_treatment',
        'manual_formatting', 'manual_list', 'direct_spacing',
        'manual_break', 'stray_quote_marker', 'mixed_formatting', 'image_inventory',
    ]

    def __init__(self, edge_cases: List[Dict], doc_path: str):
        self.edge_cases = edge_cases
        self.doc_path = doc_path

    # ── Helpers ────────────────────────────────────────────────────────────

    @staticmethod
    def _nice_type_label(case_type: str) -> str:
        meta = EdgeCaseReviewer.SECTION_META.get(case_type)
        return meta[0] if meta else case_type.replace('_', ' ').title()

    @staticmethod
    def _section_description(case_type: str) -> str:
        meta = EdgeCaseReviewer.SECTION_META.get(case_type)
        return meta[1] if meta else ''

    @staticmethod
    def _severity_label(severity: str) -> str:
        return {'high': 'Worth fixing', 'medium': 'Worth a look', 'low': 'Just noting', 'auto-preserved': 'Carried through automatically'}.get(severity, severity.title())

    @staticmethod
    def _severity_counts(cases: List[Dict]) -> Dict[str, int]:
        counts = {'high': 0, 'medium': 0, 'low': 0}
        for c in cases:
            sev = c.get('severity', 'low')
            counts[sev] = counts.get(sev, 0) + 1
        return counts

    @staticmethod
    def _dominant_severity(cases: List[Dict]) -> str:
        for sev in ('high', 'medium', 'low'):
            if any(c.get('severity') == sev for c in cases):
                return sev
        return 'low'

    @staticmethod
    def _esc(text) -> str:
        return html.escape(str(text)) if text else ''

    @staticmethod
    def _truncate(text: str, length: int = 80) -> str:
        text = (text or '').strip()
        return text[:length] + '…' if len(text) > length else text

    # ── Section rendering ─────────────────────────────────────────────────

    def _render_severity_badges(self, cases: List[Dict]) -> str:
        counts = self._severity_counts(cases)
        parts = []
        for sev in ('high', 'medium', 'low'):
            if counts[sev]:
                parts.append(f'<span class="badge badge-{sev}">{counts[sev]} {self._severity_label(sev)}</span>')
        return ' '.join(parts)

    def _render_bulk_table(self, section_id: str, cases: List[Dict], case_type: str) -> str:
        """Render a compact table for bulk sections (>20 items)."""
        # Build summary line
        summary = self._build_bulk_summary(cases, case_type)
        rows = []
        for c in cases:
            loc = self._esc(c.get('location', ''))
            text = self._esc(self._truncate(c.get('text', ''), 90))
            # Determine subtype column
            subtype = ''
            if case_type == 'manual_list':
                subtype = self._esc(c.get('list_type', ''))
            elif case_type == 'manual_formatting':
                subtype = self._esc(', '.join(c.get('issues', [])))
            elif case_type == 'direct_spacing':
                subtype = self._esc(', '.join(c.get('issues', []))[:60])
            elif c.get('subtype'):
                subtype = self._esc(c.get('subtype', ''))
            sev_class = c.get('severity', 'low')
            rows.append(
                f'<tr class="bulk-row" data-section="{self._esc(section_id)}">'
                f'<td class="col-loc">{loc}</td>'
                f'<td class="col-text">{text}</td>'
                f'<td class="col-sub">{subtype}</td>'
                f'<td class="col-sev"><span class="sev-dot sev-dot-{sev_class}"></span></td>'
                f'</tr>'
            )
        visible_rows = '\n'.join(rows[:10])
        hidden_rows = '\n'.join(rows[10:])
        show_more = ''
        if len(rows) > 10:
            show_more = '\n'.join(
                r.replace('class="bulk-row"', f'class="bulk-row bulk-row-hidden" style="display:none"')
                for r in rows[10:]
            )

        btn = ''
        if len(rows) > 10:
            btn = (
                f'<button class="btn-show-all" data-section="{self._esc(section_id)}" '
                f'onclick="toggleBulkRows(this, \'{self._esc(section_id)}\')" '
                f'type="button">Show all {len(rows)}</button>'
            )

        return (
            f'<p class="bulk-summary">{summary}</p>'
            f'<div class="table-wrap">'
            f'<table class="bulk-table">'
            f'<thead><tr><th>Location</th><th>Text</th><th>Detail</th><th></th></tr></thead>'
            f'<tbody>{visible_rows}\n{show_more}</tbody>'
            f'</table>'
            f'{btn}'
            f'</div>'
        )

    def _build_bulk_summary(self, cases: List[Dict], case_type: str) -> str:
        n = len(cases)
        label = self._nice_type_label(case_type).lower()
        if case_type == 'manual_list':
            by_lt = {}
            for c in cases:
                lt = c.get('list_type', 'unknown')
                by_lt[lt] = by_lt.get(lt, 0) + 1
            parts = ', '.join(f'{v} {k}' for k, v in sorted(by_lt.items(), key=lambda x: -x[1]))
            return f'{n} {label}: {parts}'
        elif case_type == 'manual_formatting':
            issue_counts = {}
            for c in cases:
                for iss in c.get('issues', []):
                    issue_counts[iss] = issue_counts.get(iss, 0) + 1
            parts = ', '.join(f'{v} {k}' for k, v in sorted(issue_counts.items(), key=lambda x: -x[1]))
            return f'{n} {label} runs: {parts}'
        elif case_type == 'direct_spacing':
            return f'{n} paragraphs with direct spacing, indentation, or alignment'
        return f'{n} {label} items'

    def _render_card(self, case: Dict) -> str:
        """Render a single finding as a compact card."""
        sev = case.get('severity', 'low')
        parts = []
        parts.append(f'<article class="finding severity-{sev}">')
        parts.append(
            f'<div class="finding-top">'
            f'<div class="location">{self._esc(case.get("location", ""))}</div>'
            f'<div class="severity-note">{self._severity_label(sev)}</div>'
            f'</div>'
        )
        parts.append(f'<div class="text-sample">{self._esc(case.get("text", ""))}</div>')

        if case.get('script'):
            parts.append(f'<span class="script-tag">{self._esc(case["script"])}</span>')
        if case.get('subtype'):
            parts.append(f'<div class="detail"><strong>Subtype:</strong> {self._esc(case["subtype"])}</div>')
        if case.get('declared_style_match'):
            parts.append(f'<div class="detail"><strong>Declared style match:</strong> {self._esc(case["declared_style_match"])}</div>')
        if case.get('style_name'):
            kind = self._esc(case.get('style_kind', 'style'))
            parts.append(f'<div class="detail"><strong>Observed {kind} style:</strong> {self._esc(case["style_name"])}</div>')
        if case.get('context_before'):
            ctx = ' / '.join(self._esc(x) for x in case['context_before'])
            parts.append(f'<div class="detail"><strong>Before:</strong> {ctx}</div>')
        if case.get('context_after'):
            ctx = ' / '.join(self._esc(x) for x in case['context_after'])
            parts.append(f'<div class="detail"><strong>After:</strong> {ctx}</div>')
        if case.get('issues'):
            parts.append(f'<div class="detail"><strong>Issues:</strong> {self._esc(", ".join(case["issues"]))}</div>')
        if case.get('font'):
            parts.append(f'<div class="detail"><strong>Font:</strong> {self._esc(case["font"])}</div>')
        if case.get('color'):
            parts.append(f'<div class="detail"><strong>Color:</strong> {self._esc(case["color"])}</div>')
        if case.get('suggestion'):
            parts.append(f'<div class="suggestion">{self._esc(case["suggestion"])}</div>')
        parts.append('</article>')
        return '\n'.join(parts)

    def _render_card_section(self, cases: List[Dict]) -> str:
        """Render ≤20 items as compact cards inside a finding-list container."""
        inner = '\n'.join(self._render_card(c) for c in cases)
        return f'<div class="finding-list">{inner}</div>'

    def _render_image_grid(self, cases: List[Dict]) -> str:
        """Render image inventory as a compact grid."""
        if not cases:
            return (
                '<div class="empty-note">'
                '<p>No inline images detected.</p>'
                '<p class="detail">If you expected images, check whether they are floating/anchored shapes rather than inline Word images.</p>'
                '</div>'
            )
        cells = []
        for c in cases:
            idx = c.get('image_index', '?')
            w = c.get('width_in')
            h = c.get('height_in')
            dims = f'{w}″ × {h}″' if w and h else 'unknown size'
            cells.append(
                f'<div class="img-cell">'
                f'<div class="img-cell-icon">🖼</div>'
                f'<div class="img-cell-label">Image {self._esc(str(idx))}</div>'
                f'<div class="img-cell-dims">{self._esc(dims)}</div>'
                f'</div>'
            )
        joined_cells = '\n'.join(cells)
        return f'<div class="img-grid">{joined_cells}</div>'

    def _render_auto_preserved(self, cases: List[Dict]) -> str:
        """Render the auto-preserved section as grouped compact tables."""
        if not cases:
            return (
                '<div class="empty-note">'
                'No auto-preserved formatting items were detected on this run.'
                '</div>'
            )

        # Group by issue type
        groups: Dict[str, List[Dict]] = {}
        for c in cases:
            issues = c.get('issues', [])
            key = ', '.join(issues) if issues else c.get('type', 'other')
            groups.setdefault(key, []).append(c)

        # Sort groups: largest first
        sorted_groups = sorted(groups.items(), key=lambda x: -len(x[1]))

        parts = []
        parts.append(
            '<p class="bulk-summary" style="margin-bottom:0.75rem">'
            f'{len(cases)} items detected but treated as intentional and preserved.</p>'
        )

        for group_label, group_cases in sorted_groups:
            nice_label = self._esc(group_label.replace('_', ' '))
            group_id = f'ap-{hash(group_label) % 99999}'
            parts.append(
                f'<div class="ap-group" style="margin-bottom:0.75rem">'
                f'<div class="ap-group-header" style="font-size:0.78rem;font-family:var(--mono);'
                f'color:var(--text-secondary);margin-bottom:0.35rem;">'
                f'{nice_label} <span style="color:var(--text-muted)">({len(group_cases)})</span></div>'
            )
            rows = []
            has_detail = any(c.get('style_name') or c.get('declared_style_match') for c in group_cases)
            for c in group_cases:
                loc = self._esc(c.get('location', ''))
                text = self._esc(self._truncate(c.get('text', ''), 80))
                detail_parts = []
                if c.get('declared_style_match'):
                    detail_parts.append(self._esc(c['declared_style_match']))
                elif c.get('style_name'):
                    detail_parts.append(self._esc(c['style_name']))
                detail_td = f'<td class="col-sub">{", ".join(detail_parts)}</td>' if has_detail else ''
                rows.append(
                    f'<tr class="bulk-row" data-section="{self._esc(group_id)}">'
                    f'<td class="col-loc">{loc}</td>'
                    f'<td class="col-text">{text}</td>'
                    f'{detail_td}'
                    f'</tr>'
                )
            # Show first 5 rows, hide rest
            visible = '\n'.join(rows[:5])
            hidden = ''
            if len(rows) > 5:
                hidden = '\n'.join(
                    r.replace('class="bulk-row"',
                              f'class="bulk-row bulk-row-hidden" style="display:none"')
                    for r in rows[5:]
                )
            btn = ''
            if len(rows) > 5:
                btn = (
                    f'<button class="btn-show-all" data-section="{self._esc(group_id)}" '
                    f'onclick="toggleBulkRows(this, \'{self._esc(group_id)}\')" '
                    f'type="button">Show all {len(rows)}</button>'
                )
            detail_th = '<th>Detail</th>' if has_detail else ''
            parts.append(
                f'<div class="table-wrap">'
                f'<table class="bulk-table">'
                f'<thead><tr><th>Location</th><th>Text</th>{detail_th}</tr></thead>'
                f'<tbody>{visible}\n{hidden}</tbody>'
                f'</table>'
                f'{btn}'
                f'</div>'
                f'</div>'
            )

        return '\n'.join(parts)

    # ── Main report generation ────────────────────────────────────────────

    def generate_html_report(self, output_path: str):
        """Generate an HTML report for review"""
        auto_preserve = [c for c in self.edge_cases if c.get('auto_decision') == 'preserve']
        actionable = [c for c in self.edge_cases if c.get('auto_decision') != 'preserve']

        report_generated_at = datetime.now().astimezone().strftime('%Y-%m-%d %H:%M %Z')
        report_title = Path(self.doc_path).name

        # Partition actionable findings by type
        by_type: Dict[str, List[Dict]] = {}
        for c in actionable:
            by_type.setdefault(c['type'], []).append(c)
        # Ensure image_inventory section always exists
        if 'image_inventory' not in by_type:
            by_type['image_inventory'] = []

        # Order sections: priority list first, then any remaining
        seen = set()
        section_order = []
        for t in self.SECTION_ORDER_PRIORITY:
            if t in by_type:
                section_order.append(t)
                seen.add(t)
        for t in by_type:
            if t not in seen:
                section_order.append(t)

        # Severity totals for overview
        sev_totals = self._severity_counts(actionable)

        # Build stacked bar data
        bar_segments = []
        total_actionable = len(actionable)
        for t in section_order:
            cases = by_type[t]
            if cases:
                bar_segments.append((t, len(cases), self._dominant_severity(cases)))

        # ── Build HTML ─────────────────────────────────────────────────
        sections_html = []
        nav_pills = []

        for case_type in section_order:
            cases = by_type[case_type]
            count = len(cases)
            label = self._nice_type_label(case_type)
            desc = self._section_description(case_type)
            sid = f'sec-{case_type}'
            sev_badges = self._render_severity_badges(cases)
            dom_sev = self._dominant_severity(cases)

            # Decide if section starts expanded
            has_high = any(c.get('severity') == 'high' for c in cases)
            starts_open = (count <= 3 and count > 0) or has_high
            open_attr = ' open' if starts_open else ''

            # Build inner content
            if case_type == 'image_inventory':
                inner = self._render_image_grid(cases)
            elif count > 20 and case_type in ('manual_formatting', 'manual_list', 'direct_spacing', 'mixed_formatting'):
                inner = self._render_bulk_table(sid, cases, case_type)
            elif count > 0:
                inner = self._render_card_section(cases)
            else:
                inner = '<div class="empty-note">No items detected.</div>'

            # Nav pill
            nav_pills.append(
                f'<a href="#{sid}" class="nav-pill" data-section="{self._esc(sid)}">'
                f'{self._esc(label)} <span class="nav-count">{count}</span></a>'
            )

            sections_html.append(
                f'<section id="{self._esc(sid)}" class="report-section" data-print-label="{self._esc(label)}">'
                f'<details{open_attr}>'
                f'<summary class="section-summary">'
                f'<div class="section-summary-left">'
                f'<span class="section-chevron"></span>'
                f'<span class="section-label-inline">{self._esc(label)}</span>'
                f'<span class="badge badge-dim">{count}</span>'
                f'{sev_badges}'
                f'</div>'
                f'<div class="section-summary-desc">{self._esc(desc)}</div>'
                f'</summary>'
                f'<div class="section-body">{inner}</div>'
                f'</details>'
                f'</section>'
            )

        # Auto-preserved section
        ap_count = len(auto_preserve)
        ap_open = ' open' if ap_count <= 3 and ap_count > 0 else ''
        ap_inner = self._render_auto_preserved(auto_preserve)
        nav_pills.append(
            f'<a href="#sec-auto-preserved" class="nav-pill" data-section="sec-auto-preserved">'
            f'Carried through automatically <span class="nav-count">{ap_count}</span></a>'
        )
        sections_html.append(
            f'<section id="sec-auto-preserved" class="report-section" data-print-label="Carried through automatically">'
            f'<details{ap_open}>'
            f'<summary class="section-summary">'
            f'<div class="section-summary-left">'
            f'<span class="section-chevron"></span>'
            f'<span class="section-label-inline">Carried through automatically</span>'
            f'<span class="badge badge-dim">{ap_count}</span>'
            f'</div>'
            f'<div class="section-summary-desc">Items detected but treated as intentional — no action needed</div>'
            f'</summary>'
            f'<div class="section-body">{ap_inner}</div>'
            f'</details>'
            f'</section>'
        )

        # Stacked bar HTML
        bar_html = ''
        if total_actionable > 0:
            max_n = max(n for (_, n, _) in bar_segments) if bar_segments else 1
            bar_rows = []
            for t, n, _sev in bar_segments:
                pct = (n / max_n) * 100
                lbl = self._nice_type_label(t)
                bar_rows.append(
                    f'<div class="section-bar-row">'
                    f'<span class="section-bar-label">{self._esc(lbl)}</span>'
                    f'<span class="section-bar-track">'
                    f'<span class="section-bar-fill" style="width:{pct:.1f}%"></span>'
                    f'</span>'
                    f'<span class="section-bar-count">{n}</span>'
                    f'</div>'
                )
            bar_html = f'<div class="section-bars">{"" .join(bar_rows)}</div>'

        # ── Final assembly ─────────────────────────────────────────────
        html_out = f"""<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Preflight — {self._esc(report_title)} · jdbb studio</title>
<link rel="stylesheet" href="/static/theme.css">
<script src="/static/theme.js" defer></script>
<style>
* {{ box-sizing: border-box; }}
html {{ background: var(--bg); }}
body {{ margin: 0; background: var(--bg); color: var(--text); font-family: var(--body); line-height: 1.55; }}
body, body * {{ font-weight: 400; }}

.pf-back {{
  display: inline-block; margin: 13px 0 0; font-family: var(--mono); font-size: 11px;
  letter-spacing: .06em; color: var(--text-secondary); text-decoration: none;
}}
.pf-back:hover {{ color: var(--accent); text-decoration: underline; text-underline-offset: 3px; }}
.pf-report {{ max-width: 880px; padding: 30px 0 72px; }}
.eyebrow {{
  margin: 0 0 7px; font-family: var(--mono); font-size: 11px; font-weight: 600;
  letter-spacing: .14em; text-transform: uppercase; color: var(--accent);
}}
.eyebrow::before {{ content: "// "; }}
h1, h2, h3 {{ margin: 0; color: var(--text); font-weight: 600; }}
.header {{ padding-bottom: 20px; border-bottom: 1px solid var(--border-strong); }}
.header h1 {{ font-size: clamp(22px, 3vw, 30px); line-height: 1.18; overflow-wrap: anywhere; }}
.meta {{ margin-top: 7px; font-family: var(--mono); font-size: 12px; color: var(--text-secondary); }}

.sticky-nav {{
  position: sticky; top: 0; z-index: 20; display: flex; gap: 8px 18px; overflow-x: auto;
  margin: 0; padding: 10px 0; border-bottom: 1px solid var(--border);
  background: var(--bg); font-family: var(--mono); white-space: nowrap;
  opacity: 0; pointer-events: none; max-height: 0; overflow-y: hidden;
  transition: opacity .2s, max-height .2s;
}}
.sticky-nav.visible {{ opacity: 1; pointer-events: auto; max-height: 48px; overflow-x: auto; }}
.nav-pill {{
  flex: none; color: var(--text-secondary); font-size: 10.5px; letter-spacing: .08em;
  text-transform: uppercase; text-decoration: none; border-bottom: 1px solid transparent; padding-bottom: 2px;
}}
.nav-pill:hover {{ color: var(--text); }}
.nav-pill.active {{ color: var(--accent); border-bottom-color: var(--accent); }}
.nav-count {{ color: var(--text-muted); font-variant-numeric: tabular-nums; }}

.overview {{ padding: 24px 0; border-bottom: 1px solid var(--border-strong); }}
.overview-label {{
  margin-bottom: 7px; font-family: var(--mono); font-size: 10.5px; letter-spacing: .12em;
  text-transform: uppercase; color: var(--text-muted);
}}
.overview h2 {{ font-size: 19px; }}
.overview-copy {{ margin: 5px 0 16px; color: var(--text-secondary); font-size: 14px; }}
.summary-grid {{ display: flex; flex-wrap: wrap; gap: 7px 18px; margin-bottom: 18px; }}
.badge {{
  font-family: var(--mono); font-size: 10.5px; letter-spacing: .06em; color: var(--text-secondary);
  border-bottom: 1px solid var(--border); padding-bottom: 2px; white-space: nowrap;
}}
.badge-high {{ color: var(--red); border-color: var(--red); }}
.badge-medium {{ color: var(--orange); border-color: var(--orange); }}
.badge-low {{ color: var(--green); border-color: var(--green); }}
.badge-dim {{ color: var(--text-muted); }}
.section-bars {{ display: flex; flex-direction: column; gap: 7px; }}
.section-bar-row {{ display: grid; grid-template-columns: minmax(120px, 160px) minmax(80px, 1fr) 28px; gap: 10px; align-items: center; font-family: var(--mono); font-size: 10.5px; }}
.section-bar-label {{ overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: var(--text-secondary); }}
.section-bar-track {{ height: 4px; background: var(--progress-bg); overflow: hidden; }}
.section-bar-fill {{ display: block; height: 100%; background: var(--accent); }}
.section-bar-count {{ color: var(--text-muted); text-align: right; font-variant-numeric: tabular-nums; }}

.report-section {{ border-bottom: 1px solid var(--border); }}
.report-section details {{ border: 0; }}
.section-summary {{
  display: flex; align-items: baseline; justify-content: space-between; gap: 12px;
  padding: 16px 0; cursor: pointer; list-style: none;
}}
.section-summary::-webkit-details-marker {{ display: none; }}
.section-summary:hover .section-label-inline {{ color: var(--accent); }}
.section-summary-left {{ display: flex; flex-wrap: wrap; align-items: baseline; gap: 7px 12px; min-width: 0; }}
.section-chevron {{ width: 11px; height: 11px; position: relative; flex: none; }}
.section-chevron::after {{
  content: ""; position: absolute; top: 2px; left: 1px; width: 6px; height: 6px;
  border-right: 1px solid var(--text-muted); border-bottom: 1px solid var(--text-muted); transform: rotate(-45deg);
}}
details[open] > .section-summary .section-chevron::after {{ top: 1px; transform: rotate(45deg); }}
.section-label-inline {{ font-size: 17px; font-weight: 600; }}
.section-summary-desc {{ max-width: 37ch; color: var(--text-muted); font-size: 12px; text-align: right; }}
.section-body {{ padding: 0 0 20px; }}

.finding-list {{ border-top: 1px solid var(--border); }}
.finding {{ padding: 14px 0; border-bottom: 1px solid var(--border); }}
.finding-top {{ display: flex; justify-content: space-between; align-items: baseline; gap: 12px; margin-bottom: 7px; }}
.location, .severity-note {{ font-family: var(--mono); font-size: 10.5px; letter-spacing: .07em; text-transform: uppercase; }}
.location {{ color: var(--text-secondary); }}
.severity-note {{ color: var(--text-muted); white-space: nowrap; }}
.severity-high .severity-note {{ color: var(--red); }}
.severity-medium .severity-note {{ color: var(--orange); }}
.severity-low .severity-note {{ color: var(--green); }}
.text-sample {{
  margin: 7px 0; padding: 8px 10px; border-left: 1px solid var(--border-strong);
  font-family: var(--mono); font-size: 12px; white-space: pre-wrap; overflow-wrap: anywhere;
}}
.detail {{ margin-top: 4px; color: var(--text-secondary); font-size: 12.5px; }}
.detail strong {{ color: var(--text); font-weight: 600; }}
.suggestion {{ margin-top: 8px; padding-left: 10px; border-left: 1px solid var(--accent); color: var(--text-secondary); font-size: 12.5px; }}
.script-tag {{ display: inline-block; margin-bottom: 3px; font-family: var(--mono); font-size: 10.5px; color: var(--accent); }}

.bulk-summary {{ margin: 0 0 12px; color: var(--text-secondary); font-size: 13px; }}
.table-wrap {{ overflow-x: auto; }}
.bulk-table {{ width: 100%; border-collapse: collapse; font-size: 12px; }}
.bulk-table th {{ padding: 7px 8px; border-bottom: 1px solid var(--border-strong); color: var(--text-muted); font-family: var(--mono); font-size: 10px; font-weight: 400; letter-spacing: .1em; text-align: left; text-transform: uppercase; }}
.bulk-table td {{ padding: 8px; border-bottom: 1px solid var(--border); color: var(--text-secondary); vertical-align: top; }}
.col-loc {{ width: 100px; color: var(--text-muted); font-family: var(--mono); font-size: 10.5px; white-space: nowrap; }}
.col-text {{ max-width: 500px; font-family: var(--mono); font-size: 11px; overflow-wrap: anywhere; }}
.col-sub {{ max-width: 180px; color: var(--text-muted); font-size: 11px; overflow-wrap: anywhere; }}
.col-sev {{ width: 20px; text-align: center; }}
.sev-dot {{ display: inline-block; width: 6px; height: 6px; }}
.sev-dot-high {{ background: var(--red); }} .sev-dot-medium {{ background: var(--orange); }} .sev-dot-low {{ background: var(--green); }}
.btn-show-all {{ margin-top: 10px; padding: 0; border: 0; background: none; color: var(--accent); cursor: pointer; font-family: var(--mono); font-size: 11px; text-decoration: underline; text-underline-offset: 3px; }}

.img-grid {{ display: grid; grid-template-columns: repeat(auto-fill, minmax(140px, 1fr)); border-top: 1px solid var(--border); border-left: 1px solid var(--border); }}
.img-cell {{ padding: 12px; border-right: 1px solid var(--border); border-bottom: 1px solid var(--border); text-align: center; }}
.img-cell-icon {{ font-size: 18px; }} .img-cell-label {{ margin-top: 3px; font-size: 12px; }} .img-cell-dims {{ font-family: var(--mono); font-size: 10.5px; color: var(--text-muted); }}
.empty-note {{ padding: 12px 0; border-top: 1px solid var(--border); color: var(--text-secondary); font-size: 13px; }}
.ap-group {{ margin-bottom: 16px !important; }}
.ap-group-header {{ margin-bottom: 6px !important; font-family: var(--mono) !important; font-size: 10.5px !important; letter-spacing: .08em; text-transform: uppercase; color: var(--text-secondary) !important; }}

@media (max-width: 640px) {{
  .pf-report {{ padding-top: 22px; }}
  .section-summary {{ align-items: flex-start; flex-direction: column; gap: 5px; }}
  .section-summary-desc {{ max-width: none; text-align: left; }}
  .finding-top {{ align-items: flex-start; flex-direction: column; gap: 3px; }}
  .section-bar-row {{ grid-template-columns: minmax(90px, 120px) minmax(60px, 1fr) 22px; gap: 7px; }}
}}
@media print {{
  html, html.dark {{
    --bg: #fff; --surface: #fff; --border: #ccc; --border-strong: #000;
    --text: #000; --text-secondary: #333; --text-muted: #555; --accent: #000;
    --green: #000; --red: #000; --orange: #000; --progress-bg: #ddd;
  }}
  .jdbb-masthead, .pf-back, .sticky-nav, #theme-bar {{ display: none !important; }}
  body {{ background: #fff; color: #000; }}
  .jdbb-shell {{ max-width: none; padding: 0; }} .pf-report {{ max-width: none; padding: 0; }}
  .report-section details, .report-section details > .section-body {{ display: block !important; }}
  .report-section details > summary {{ display: none !important; }}
  .report-section::before {{ content: attr(data-print-label); display: block; padding: 12px 0 6px; border-top: 1px solid #000; font-family: monospace; font-size: 10pt; text-transform: uppercase; }}
  .bulk-row-hidden {{ display: table-row !important; }} .btn-show-all {{ display: none !important; }}
  .finding, .report-section {{ break-inside: avoid; }}
}}
</style>
</head>
<body>
<div class="jdbb-shell">
  <header class="jdbb-masthead">
    <a class="jdbb-wordmark" href="/"><span class="bracket">[</span><span class="kj">j</span>dbb<span class="bracket">]</span><span class="studio">studio</span></a>
    <nav><div id="theme-bar"></div></nav>
  </header>

  <a class="pf-back" href="{{{{FACTORY_URL}}}}">← Back to the factory</a>

  <main class="pf-report">
    <header class="header">
      <div class="eyebrow">Inspect</div>
      <h1>{self._esc(report_title)}</h1>
      <div class="meta">Manuscript preflight check before generating EPUB or PDF.</div>
      <div class="meta">Generated {report_generated_at}</div>
    </header>

    <nav class="sticky-nav" id="stickyNav" aria-label="Preflight sections">
      {''.join(nav_pills)}
    </nav>

    <section class="overview">
      <div class="overview-label">Overview</div>
      <h2>Preflight summary</h2>
      <p class="overview-copy">Review findings, make fixes in Word if needed, then rerun inspection.</p>
      <div class="summary-grid">
        <span class="badge">{len(self.edge_cases)} total</span>
        <span class="badge">{len(actionable)} actionable</span>
        <span class="badge">{ap_count} carried through automatically</span>
        <span class="badge badge-high">{sev_totals['high']} Worth fixing</span>
        <span class="badge badge-medium">{sev_totals['medium']} Worth a look</span>
        <span class="badge badge-low">{sev_totals['low']} Just noting</span>
      </div>
      {bar_html}
    </section>

    {''.join(sections_html)}
  </main>
</div>

<script>
/* ── Sticky nav visibility ───────────────────────────────── */
(function() {{
  var nav = document.getElementById('stickyNav');
  var header = document.querySelector('.header');
  if (!nav || !header) return;
  var io = new IntersectionObserver(function(entries) {{
    nav.classList.toggle('visible', !entries[0].isIntersecting);
  }}, {{ threshold: 0 }});
  io.observe(header);
}})();

/* ── Active nav pill on scroll ───────────────────────────── */
(function() {{
  var pills = document.querySelectorAll('.nav-pill');
  var sections = document.querySelectorAll('.report-section');
  if (!pills.length || !sections.length) return;
  var sectionMap = {{}};
  pills.forEach(function(p) {{ sectionMap[p.getAttribute('data-section')] = p; }});

  function update() {{
    var current = null;
    sections.forEach(function(s) {{
      if (s.getBoundingClientRect().top <= 120) current = s.id;
    }});
    pills.forEach(function(p) {{ p.classList.remove('active'); }});
    if (current && sectionMap[current]) sectionMap[current].classList.add('active');
  }}
  var ticking = false;
  window.addEventListener('scroll', function() {{
    if (!ticking) {{ requestAnimationFrame(function() {{ update(); ticking = false; }}); ticking = true; }}
  }});
  update();
}})();

/* ── Smooth scroll for nav pills ─────────────────────────── */
(function() {{
  document.querySelectorAll('.nav-pill').forEach(function(a) {{
    a.addEventListener('click', function(e) {{
      e.preventDefault();
      var target = document.querySelector(a.getAttribute('href'));
      if (!target) return;
      var details = target.querySelector('details');
      if (details && !details.open) details.open = true;
      target.scrollIntoView({{ behavior: 'smooth', block: 'start' }});
    }});
  }});
}})();

/* ── Bulk table show-all toggle ──────────────────────────── */
function toggleBulkRows(btn, sectionId) {{
  var rows = document.querySelectorAll('.bulk-row-hidden[data-section="' + sectionId + '"]');
  var showing = btn.getAttribute('data-expanded') === '1';
  rows.forEach(function(r) {{ r.style.display = showing ? 'none' : ''; }});
  btn.setAttribute('data-expanded', showing ? '0' : '1');
  var table = btn.closest('.table-wrap').querySelector('table');
  var allRows = table ? table.querySelectorAll('tbody tr') : [];
  var total = allRows.length;
  btn.textContent = showing ? 'Show all ' + total : 'Show fewer';
}}

/* ── Hash navigation: auto-open targeted section ─────────── */
(function() {{
  function openHash() {{
    var hash = window.location.hash;
    if (!hash) return;
    var target = document.querySelector(hash);
    if (!target) return;
    var details = target.querySelector('details');
    if (details && !details.open) details.open = true;
    setTimeout(function() {{ target.scrollIntoView({{ behavior: 'smooth', block: 'start' }}); }}, 100);
  }}
  openHash();
  window.addEventListener('hashchange', openHash);
}})();
</script>
</body>
</html>"""

        with open(output_path, 'w', encoding='utf-8') as f:
            f.write(html_out)


def main():
    parser = argparse.ArgumentParser(description='Detect edge cases in Word documents')
    parser.add_argument('input_file', help='Input Word document')
    parser.add_argument('-o', '--output', default='edge_case_review.html',
                       help='Output HTML report (default: edge_case_review.html)')
    parser.add_argument('--json', action='store_true',
                       help='Also output raw JSON data')
    parser.add_argument('--declared-styles', default='',
                       help='Optional JSON file of declared custom styles from transmittal/spec')
    
    args = parser.parse_args()
    
    declared_styles = []
    if args.declared_styles:
        declared_path = Path(args.declared_styles)
        if declared_path.exists():
            declared_styles = json.loads(declared_path.read_text(encoding='utf-8'))
    
    print(f"Analyzing {args.input_file}...")
    detector = EdgeCaseDetector(args.input_file, declared_styles=declared_styles)
    edge_cases = detector.detect_all()
    
    print(f"Found {len(edge_cases)} edge cases")
    
    # Generate HTML review interface
    reviewer = EdgeCaseReviewer(edge_cases, args.input_file)
    reviewer.generate_html_report(args.output)
    print(f"HTML report generated: {args.output}")
    
    # Optionally save JSON
    if args.json:
        json_output = args.output.replace('.html', '.json')
        with open(json_output, 'w') as f:
            json.dump(edge_cases, f, indent=2)
        print(f"JSON data saved: {json_output}")


if __name__ == '__main__':
    main()