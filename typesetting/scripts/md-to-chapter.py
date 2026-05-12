#!/usr/bin/env python3
"""Convert markdown chapter (with custom styles) to a Typst chapter file.

Input is expected to come from:
  pandoc --from=docx+styles --to=markdown+fenced_divs

Supported paragraph/custom styles:
  - First Paragraph
  - Body Text / unstyled paragraphs
  - Verse
  - Code Block / Code-Block
  - Block Quote / Block-Quote
  - Epigraph
  - Section Break / Section-Break

Other custom styles fall back to normal paragraph rendering so the pipeline
stays resilient for manuscript-specific styles like tweet-p / metadata-p.
"""

from __future__ import annotations

import re
import sys
from typing import Dict, List, Tuple


HEADING_RE = re.compile(r"^(#{1,6})\s+(.*)$")
FOOTNOTE_DEF_RE = re.compile(r"^\[\^([^\]]+)\]:\s*(.*)$")
FOOTNOTE_REF_RE = re.compile(r"\[\^([^\]]+)\]")
DIV_START_RE = re.compile(r'^:::\s*\{custom-style=["\']([^"\']+)["\']\}')
DIV_END_RE = re.compile(r'^:::$')


def normalize_style(style: str | None) -> str | None:
    if style is None:
        return None
    return style.strip().lower().replace("_", "-")

# Built-in styles that have explicit handlers — skip these in the custom fallback
BUILTIN_STYLES = {
    'verse', 'code-block', 'code block', 'block-quote', 'block quote',
    'epigraph', 'section-break', 'section break', 'first paragraph',
    'quote', 'list bullet', 'list number', 'normal',
}


def style_to_typst_func(style: str) -> str | None:
    """Convert a Word custom-style name to a valid Typst function name.

    Returns None if the style is a built-in or can't be converted.
    Examples:
        "tweet"      → "tweet"
        "metadata-p" → "metadata-p"  (hyphens are valid in Typst identifiers)
        "My Style"   → "my-style"
    """
    normalized = normalize_style(style)
    if normalized is None or normalized in BUILTIN_STYLES:
        return None
    # Convert spaces/underscores to hyphens, strip non-alphanumeric except hyphens
    import re as _re
    name = _re.sub(r'[^a-z0-9-]', '', normalized)
    if not name or name[0].isdigit():
        return None
    return name



def convert_markdown_heading(line: str) -> str | None:
    m = HEADING_RE.match(line.strip())
    if not m:
        return None
    level = len(m.group(1))
    title = m.group(2).strip()
    return f"{'=' * level} {title}"


def escape_typst_text(text: str) -> str:
    """Escape characters that Typst interprets specially.

    Handles #, $, and @ while avoiding double-escaping characters
    that pandoc already escaped with backslashes.
    """
    text = re.sub(r'(?<!\\)#', r'\#', text)
    text = re.sub(r'(?<!\\)\$', r'\$', text)
    text = re.sub(r'(?<!\\)@', r'\@', text)
    return text


def extract_footnotes(md_content: str) -> Tuple[str, Dict[str, str]]:
    lines = md_content.splitlines()
    body: List[str] = []
    notes: Dict[str, str] = {}

    i = 0
    while i < len(lines):
        line = lines[i]
        m = FOOTNOTE_DEF_RE.match(line)
        if not m:
            body.append(line)
            i += 1
            continue

        note_id = m.group(1)
        parts = [m.group(2).strip()]
        i += 1
        while i < len(lines):
            continuation = lines[i]
            if FOOTNOTE_DEF_RE.match(continuation):
                break
            if continuation.startswith("    ") or continuation.startswith("\t"):
                parts.append(continuation.strip())
                i += 1
                continue
            if continuation.strip() == "":
                break
            break
        note_text = " ".join(p for p in parts if p).strip()
        # Strip any custom-style fenced div wrappers from footnote content
        note_text = re.sub(r'::: \{custom-style="[^"]*"\}\s*', '', note_text)
        note_text = re.sub(r'\s*:::', '', note_text)
        notes[note_id] = note_text.strip()

    return "\n".join(body), notes


def convert_inline_formatting(line: str, footnotes: Dict[str, str]) -> str:
    def repl_footnote(match: re.Match[str]) -> str:
        note_id = match.group(1)
        note = footnotes.get(note_id)
        if not note:
            return f"[{note_id}]"
        rendered = convert_inline_formatting(note, {})
        return f"#footnote[{rendered}]"

    line = FOOTNOTE_REF_RE.sub(repl_footnote, line)

    # Custom character styles: [text]{custom-style="name"} → #name[text]
    def repl_custom_style(match):
        text = match.group(1)
        style_name = match.group(2)
        func = style_to_typst_func(style_name)
        if func:
            return f'#{func}[{text}]'
        return text  # fallback: just the text content
    line = re.sub(r'\[([^\]]+)\]\{custom-style=["\'](.*?)["\']\}', repl_custom_style, line)

    # Links: [text](url) -> #link("url")[text]
    line = re.sub(r'\[\[([^\]]+)\]\{\.underline\}\]\(([^)]+)\)', r'#link("\2")[\1]', line)
    line = re.sub(r'\[([^\]]+)\]\(([^)]+)\)', r'#link("\2")[\1]', line)

    # Protect URLs from being mangled by italic/bold conversion
    url_placeholders = []
    def save_url(m):
        url_placeholders.append(m.group(0))
        return f'\x00URL{len(url_placeholders)-1}\x00'
    line = re.sub(r'https?://[^\s)\]>]+', save_url, line)

    # Bold-italic (***text***) must be handled before bold or italic separately
    line = re.sub(r'\*\*\*([^*\n]+)\*\*\*', r'*_\1_*', line)
    # Bold (**text**) → *text* (Typst bold)
    line = re.sub(r'\*\*([^*\n]+)\*\*', r'*\1*', line)
    # Italic (*text*) → _text_ (Typst italic) — but not inside bold markers
    line = re.sub(r'(?<![\\*])\*([^*\n]+)\*(?!\*)', r'_\1_', line)

    # Restore URLs
    for i, url in enumerate(url_placeholders):
        line = line.replace(f'\x00URL{i}\x00', url)

    line = re.sub(r'\{\.underline\}', '', line)
    line = line.replace('---', '—')
    line = line.replace('--', '–')
    line = escape_typst_text(line)
    line = line.rstrip('\\')

    # Images: handle AFTER escaping so #image/#block don't get \# escaped
    # Pattern matches both escaped and unescaped markdown image syntax
    def repl_image(m):
        path = m.group(1)
        return f'#block(breakable: false, width: 100%)[#image("{path}", width: 100%, fit: "contain")]'
    line = re.sub(r'!\[[^\]]*\]\(([^)]+)\)(?:\{[^}]*\})?', repl_image, line)

    return line


def parse_fenced_divs(content: str) -> List[Tuple[str | None, str]]:
    segments: List[Tuple[str | None, str]] = []
    current_style: str | None = None
    current_lines: List[str] = []

    for line in content.splitlines():
        start_match = DIV_START_RE.match(line)
        end_match = DIV_END_RE.match(line)

        if start_match:
            if current_lines:
                segments.append((current_style, '\n'.join(current_lines)))
                current_lines = []
            current_style = start_match.group(1)
            continue

        if end_match and current_style is not None:
            segments.append((current_style, '\n'.join(current_lines)))
            current_lines = []
            current_style = None
            continue

        current_lines.append(line)

    if current_lines:
        segments.append((current_style, '\n'.join(current_lines)))

    return segments


def render_plain_content(content: str, is_first_para: bool, footnotes: Dict[str, str]) -> Tuple[str, bool]:
    output: List[str] = []
    first_para_lines: List[str] = []  # accumulate first-para lines
    blockquote_lines: List[str] = []  # accumulate bare blockquote lines

    def flush_first_para():
        nonlocal is_first_para
        if first_para_lines:
            body = ' '.join(first_para_lines)
            output.append(f'#par(first-line-indent: 0em)[{body}]')
            first_para_lines.clear()
            is_first_para = False

    def flush_blockquote():
        if blockquote_lines:
            joined = ' '.join(blockquote_lines)
            # If entire blockquote is wrapped in *...* (italic), strip and use #emph
            is_italic = False
            inner = joined
            if (joined.startswith('*') and joined.endswith('*')
                    and not joined.startswith('**')):
                inner = joined[1:-1]
                is_italic = True
            elif joined.startswith('_') and joined.endswith('_'):
                inner = joined[1:-1]
                is_italic = True
            rendered = convert_inline_formatting(inner, footnotes)
            if is_italic:
                output.append(f'#blockquote[#emph[{rendered}]]')
            else:
                output.append(f'#blockquote[{rendered}]')
            blockquote_lines.clear()

    for raw_line in content.splitlines():
        line = raw_line.strip()

        # Bare blockquote line: accumulate
        if line.startswith('> '):
            flush_first_para()
            blockquote_lines.append(line[2:].strip())
            continue
        elif blockquote_lines:
            # Non-blockquote line after blockquote: flush the blockquote
            flush_blockquote()

        # Blank line
        if not line:
            flush_first_para()
            output.append('')
            continue

        heading = convert_markdown_heading(line)
        if heading is not None:
            flush_first_para()
            output.append('')
            output.append(heading)
            output.append('')
            is_first_para = True
            continue

        rendered = convert_inline_formatting(raw_line, footnotes)
        if is_first_para:
            first_para_lines.append(rendered)
        else:
            output.append(rendered)

    # Flush any remaining
    flush_first_para()
    flush_blockquote()

    return '\n'.join(output).strip(), is_first_para


def render_segment(style: str | None, content: str, is_first_para: bool, footnotes: Dict[str, str]) -> Tuple[str, bool]:
    content = content.strip()
    if not content:
        return '', is_first_para

    normalized = normalize_style(style)
    lines = content.splitlines()

    if normalized == 'verse':
        output = ['', '#poem[']
        for line in lines:
            output.append(f"  {convert_inline_formatting(line, footnotes)} \\\\ ")
        output.extend([']', ''])
        return '\n'.join(output), False

    if normalized in {'code-block', 'code block'}:
        output = ['', '#code-block[', '```']
        output.extend(lines)
        output.extend(['```', ']', ''])
        return '\n'.join(output), False

    if normalized in {'block-quote', 'block quote', 'quote'}:
        # Strip '> ' prefixes and join lines before formatting,
        # so italic/bold markers that span across continuation lines work correctly.
        stripped = []
        for line in lines:
            if line.startswith('> '):
                line = line[2:]
            stripped.append(line.strip())
        joined = ' '.join(s for s in stripped if s)
        rendered = convert_inline_formatting(joined, footnotes)
        return f'\n#blockquote[{rendered}]\n', False

    if normalized == 'epigraph':
        quote_lines: List[str] = []
        attribution: str | None = None
        for line in lines:
            stripped = line.strip()
            if stripped.startswith('—') or stripped.startswith('--'):
                attribution = stripped.lstrip('—-–').strip()
            else:
                quote_lines.append(convert_inline_formatting(line, footnotes))
        quote_body = '\n'.join(quote_lines)
        if attribution:
            return f'#epigraph([{quote_body}], attribution: [{convert_inline_formatting(attribution, footnotes)}])', False
        return f'#epigraph([{quote_body}])', False

    if normalized in {'section-break', 'section break'}:
        return '\n'.join(['', '#section-break', '']), True

    if normalized == 'first paragraph':
        output = ['']
        for line in lines:
            output.append(convert_inline_formatting(line, footnotes) if line.strip() else '')
        return '\n'.join(output), False

    # Custom paragraph style: emit a Typst function call if the style name is usable
    if style is not None:
        func_name = style_to_typst_func(style)
        if func_name is not None:
            rendered_lines = []
            for line in lines:
                rendered_lines.append(convert_inline_formatting(line, footnotes) if line.strip() else '')
            body = '\n'.join(rendered_lines).strip()
            return f'\n#{func_name}[{body}]\n', False
        else:
            # Unrecognized style name — render as plain text with a comment
            print(f"WARNING: unrecognized custom style: {style!r}", file=__import__('sys').stderr)

    # Default: normal paragraph rendering (no style or unrecognizable style name)
    return render_plain_content(content, is_first_para, footnotes)


def convert_md_to_typst(md_content: str, title: str, author: str) -> str:
    md_content, footnotes = extract_footnotes(md_content)

    if md_content.startswith('---'):
        end_idx = md_content.find('---', 3)
        if end_idx != -1:
            md_content = md_content[end_idx + 3 :].strip()

    lines = md_content.splitlines()
    while lines and (lines[0].startswith('# ') or lines[0].startswith('## ') or not lines[0].strip()):
        lines.pop(0)
    md_content = '\n'.join(lines)

    output: List[str] = [
        '// Chapter content — typography is controlled by the template + spec config.',
        '// This file is #include\'d into main.typ which sets up the book() show rule.',
        '#import "config.typ": *',
        '#import "custom-styles.typ": *',
        '',
        f'= {escape_typst_text(title)}',
        '',
        f'#v(0.25em)',
        f'#text(size: 1.333em, weight: 600)[{escape_typst_text(author)}]',
        '',
        '#v(2em)',
        '',
    ]

    segments = parse_fenced_divs(md_content)
    is_first_para = True
    for style, content in segments:
        rendered, is_first_para = render_segment(style, content, is_first_para, footnotes)
        if rendered.strip():
            output.append(rendered)

    return '\n'.join(output).rstrip() + '\n'


def main() -> int:
    if len(sys.argv) < 4:
        print('Usage: md-to-chapter.py input.md title author > output.typ', file=sys.stderr)
        return 1

    input_file = sys.argv[1]
    title = sys.argv[2]
    author = sys.argv[3]

    with open(input_file, 'r', encoding='utf-8') as f:
        md_content = f.read()

    print(convert_md_to_typst(md_content, title, author), end='')
    return 0


if __name__ == '__main__':
    raise SystemExit(main())
