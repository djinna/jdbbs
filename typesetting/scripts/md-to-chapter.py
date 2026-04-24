#!/usr/bin/env python3
"""Convert markdown chapter (with custom styles) to Typst chapter file.

Expects markdown with fenced divs from pandoc --from=docx+styles:
  ::: {custom-style="Verse"}
  poetry here
  :::

Supported styles:
  - First Paragraph: No indent (after headings/breaks)
  - Body Text: Standard paragraph with indent  
  - Verse: Poetry/lyrics (monospace, preserve lines)
  - Code-Block: Terminal/code output (monospace)
  - Block-Quote: Extended quotations (italic, indented)
  - Epigraph: Chapter-opening quotes (italic, attributed)
  - Section-Break: Scene/section divider (breve marks)
"""

import re
import sys


def escape_typst(text):
    """Escape special Typst characters."""
    # Escape $ (math mode), # (functions), but not * or _ (formatting)
    text = text.replace('$', '\\$')
    # Don't escape # at start of lines (headers) - handled elsewhere
    return text


def convert_inline_formatting(line):
    """Convert markdown inline formatting to Typst."""
    # Italic: *text* or _text_ -> _text_
    line = re.sub(r'(?<![\\*])\*([^*\n]+)\*(?!\*)', r'_\1_', line)
    
    # Bold: **text** -> *text* (do this after italic)
    line = re.sub(r'\*\*([^*]+)\*\*', r'*\1*', line)
    
    # Links: [text](url) -> #link("url")[text]
    line = re.sub(r'\[\[([^\]]+)\]\{\.underline\}\]\(([^)]+)\)', r'#link("\2")[\1]', line)
    line = re.sub(r'\[([^\]]+)\]\(([^)]+)\)', r'#link("\2")[\1]', line)
    
    # Remove {.underline} artifacts
    line = re.sub(r'\{\.underline\}', '', line)
    
    # Em dash / en dash
    line = line.replace('---', '—')
    line = line.replace('--', '–')
    
    # Escape $ for Typst (math mode)
    line = line.replace('$', '\\$')
    
    # Remove trailing backslash escapes from pandoc
    line = line.rstrip('\\')
    
    return line


def parse_fenced_divs(content):
    """Parse content into segments with style info.
    
    Returns list of (style, content) tuples.
    style is None for unstyled content.
    """
    segments = []
    
    # Pattern for fenced divs: ::: {custom-style="StyleName"}
    div_start = re.compile(r'^:::\s*\{custom-style=["\']([^"\']+)["\']\}')
    div_end = re.compile(r'^:::$')
    
    lines = content.split('\n')
    current_style = None
    current_lines = []
    
    for line in lines:
        start_match = div_start.match(line)
        end_match = div_end.match(line)
        
        if start_match:
            # Save any accumulated unstyled content
            if current_lines:
                segments.append((current_style, '\n'.join(current_lines)))
                current_lines = []
            current_style = start_match.group(1)
        elif end_match and current_style:
            # End of styled div
            segments.append((current_style, '\n'.join(current_lines)))
            current_lines = []
            current_style = None
        else:
            current_lines.append(line)
    
    # Don't forget trailing content
    if current_lines:
        segments.append((current_style, '\n'.join(current_lines)))
    
    return segments


def render_segment(style, content, is_first_para):
    """Render a styled segment to Typst."""
    content = content.strip()
    if not content:
        return '', is_first_para
    
    output = []
    
    # Handle different styles
    if style == 'Verse':
        output.append('')
        output.append('#poem[')
        for line in content.split('\n'):
            line = convert_inline_formatting(line)
            output.append(f'  {line} \\\\')
        output.append(']')
        output.append('')
        return '\n'.join(output), False
    
    elif style == 'Code-Block':
        output.append('')
        output.append('#code-block[')
        output.append('```')
        for line in content.split('\n'):
            # Don't convert formatting in code blocks
            output.append(line)
        output.append('```')
        output.append(']')
        output.append('')
        return '\n'.join(output), False
    
    elif style == 'Block-Quote':
        output.append('')
        output.append('#blockquote[')
        for line in content.split('\n'):
            line = convert_inline_formatting(line)
            output.append(f'  {line}')
        output.append(']')
        output.append('')
        return '\n'.join(output), False
    
    elif style == 'Epigraph':
        # Try to split attribution (line starting with —)
        lines = content.split('\n')
        quote_lines = []
        attribution = None
        for line in lines:
            if line.strip().startswith('—') or line.strip().startswith('--'):
                attribution = line.strip().lstrip('—-–').strip()
            else:
                quote_lines.append(convert_inline_formatting(line))
        
        output.append('')
        if attribution:
            output.append(f'#epigraph(attribution: "{attribution}")[')
        else:
            output.append('#epigraph[')
        output.extend(f'  {l}' for l in quote_lines)
        output.append(']')
        output.append('')
        return '\n'.join(output), False
    
    elif style == 'Section-Break':
        output.append('')
        output.append('#section-break')
        output.append('')
        return '\n'.join(output), True  # Next para is first
    
    elif style == 'First Paragraph' or is_first_para:
        output.append('')
        output.append('#set par(first-line-indent: 0em)')
        for line in content.split('\n'):
            if line.strip():
                line = convert_inline_formatting(line)
                output.append(line)
            else:
                output.append('')
        return '\n'.join(output), False
    
    else:  # Body Text or unstyled
        output.append('')
        for line in content.split('\n'):
            if line.strip():
                line = convert_inline_formatting(line)
                output.append(line)
            else:
                output.append('')
        return '\n'.join(output), False


def convert_md_to_typst(md_content, title, author):
    """Convert markdown to Typst."""
    output = []
    
    # Chapter header - use direct styling since config isn't available in includes
    output.append('#set text(font: "Libertinus Serif", size: 10pt)')
    output.append('#set par(justify: true, leading: 0.6em, first-line-indent: 0.75em)')
    output.append('')
    output.append(f'#text(font: "Source Sans 3", size: 1.667em, weight: "bold")[{title}]')
    output.append('')
    output.append(f'#text(font: "Source Sans 3", size: 1.333em, weight: "medium")[{author}]')
    output.append('')
    output.append('#v(2em)')
    output.append('')
    
    # Remove YAML frontmatter if present
    if md_content.startswith('---'):
        end_idx = md_content.find('---', 3)
        if end_idx != -1:
            md_content = md_content[end_idx + 3:].strip()
    
    # Remove markdown chapter header if present (## Title at start)
    lines = md_content.split('\n')
    while lines and (lines[0].startswith('## ') or lines[0].startswith('# ') or not lines[0].strip()):
        lines.pop(0)
    md_content = '\n'.join(lines)
    
    # Parse into styled segments
    segments = parse_fenced_divs(md_content)
    
    is_first_para = True
    for style, content in segments:
        rendered, is_first_para = render_segment(style, content, is_first_para)
        if rendered.strip():
            output.append(rendered)
    
    return '\n'.join(output)


def convert_md_to_typst_legacy(md_content, title, author):
    """Legacy converter for markdown without custom styles."""
    lines = md_content.strip().split('\n')
    output = []
    
    # Chapter header
    output.append('#set text(font: "Libertinus Serif", size: 10pt)')
    output.append('#set par(justify: true, leading: 0.6em, first-line-indent: 0.75em)')
    output.append('')
    output.append(f'#text(font: "Source Sans 3", size: 1.667em, weight: "bold")[{title}]')
    output.append('')
    output.append(f'#text(font: "Source Sans 3", size: 1.333em, weight: "medium")[{author}]')
    output.append('')
    output.append('#v(2em)')
    output.append('')
    output.append('#set par(first-line-indent: 0em)')  # First para no indent
    
    in_code_block = False
    para_count = 0
    
    i = 0
    while i < len(lines):
        line = lines[i]
        
        # Skip markdown chapter header
        if i < 3 and line.startswith('## '):
            i += 1
            continue
        
        # Section headers
        if line.startswith('### '):
            section_title = line[4:]
            output.append('')
            output.append('#v(1em)')
            output.append(f'#text(font: "Source Sans 3", size: 1em, weight: "semibold")[{section_title}]')
            output.append('#v(0.5em)')
            output.append('#set par(first-line-indent: 0em)')
            para_count = 0
            i += 1
            continue
        elif line.startswith('## '):
            section_title = line[3:]
            output.append('')
            output.append('#v(1.5em)')
            output.append(f'#text(font: "Source Sans 3", size: 1.2em, weight: "semibold")[{section_title}]')
            output.append('#v(0.5em)')
            output.append('#set par(first-line-indent: 0em)')
            para_count = 0
            i += 1
            continue
            
        # Code blocks
        if line.startswith('```'):
            in_code_block = not in_code_block
            if in_code_block:
                output.append('')
                output.append('#set par(first-line-indent: 0em)')
            output.append('```' if in_code_block else '```')
            if not in_code_block:
                output.append('')
            i += 1
            continue
            
        if in_code_block:
            output.append(line)
            i += 1
            continue
        
        # Blank lines
        if not line.strip():
            if para_count == 1:
                output.append('')
                output.append('#set par(first-line-indent: 0.75em)')
            output.append('')
            i += 1
            continue
        
        # Regular content
        line = convert_inline_formatting(line)
        
        if line.strip():
            para_count += 1
        output.append(line)
        i += 1
    
    return '\n'.join(output)


if __name__ == '__main__':
    if len(sys.argv) < 4:
        print("Usage: md-to-chapter.py input.md title author > output.typ")
        sys.exit(1)
    
    input_file = sys.argv[1]
    title = sys.argv[2]
    author = sys.argv[3]
    
    with open(input_file, 'r') as f:
        md_content = f.read()
    
    # Detect if content has custom styles (fenced divs)
    if '::: {custom-style=' in md_content:
        typst_content = convert_md_to_typst(md_content, title, author)
    else:
        # Fall back to legacy processing
        typst_content = convert_md_to_typst_legacy(md_content, title, author)
    
    print(typst_content)
