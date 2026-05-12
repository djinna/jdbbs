#!/usr/bin/env python3
"""
apply-corrections-docx.py — Apply a corrections ledger to Word (.docx) files.

Corrections are small, discrete text fixes applied directly to Word document
paragraphs. This tool is for typo-level changes ONLY.

WHAT THIS TOOL HANDLES (the "corrections" tier):
  - Typo fixes:           "accomodate" → "accommodate"
  - Punctuation fixes:    "said ," → "said,"
  - Duplicate words:      "the the quick" → "the quick"
  - Case fixes:           "iphone" → "iPhone"
  - Minor wording tweaks: "very unique" → "unique"
  - Name/term fixes:      "Etherium" → "Ethereum"
  - Whitespace cleanup:   "word  word" → "word word"

WHAT THIS TOOL DOES NOT HANDLE (edit the Word doc manually instead):
  - Adding or removing sentences or paragraphs
  - Reordering content
  - Changing formatting (bold, italic, styles)
  - Modifying images, captions, or figures
  - Structural changes (chapters, headings, TOC)
  - Any replacement longer than ~200 characters

CROSS-RUN MATCHING:
  Word stores paragraph text split across multiple "runs" based on formatting
  changes (bold/italic toggles, spell-check marks, etc.). A single word like
  "accommodate" might be stored as ["accom", "odate"] in two runs.

  This tool concatenates all run text in a paragraph, finds matches against
  the joined string, maps match positions back to individual runs, and
  performs the replacement while preserving the formatting of the first
  matched run.

Usage:
  python3 apply-corrections-docx.py corrections.yaml BOOK.docx [-o OUTPUT.docx]
  python3 apply-corrections-docx.py corrections.yaml BOOK.docx --dry-run
  python3 apply-corrections-docx.py --example > corrections.yaml

Corrections file format (YAML):
  corrections:
    - find: "accomodate"
      replace: "accommodate"
      note: "spelling fix"

    - find: "the the quick"
      replace: "the quick"
      chapter: "Soda Sweet as Blood"
      note: "duplicate word, p.12"
"""

import argparse
import os
import shutil
import sys
from copy import deepcopy
from pathlib import Path

try:
    import yaml
except ImportError:
    print("PyYAML required: pip install pyyaml", file=sys.stderr)
    sys.exit(1)

try:
    from docx import Document
except ImportError:
    print("python-docx required: pip install python-docx", file=sys.stderr)
    sys.exit(1)


EXAMPLE_CORRECTIONS = """\
# Corrections ledger for: Ghosts in Machines
# Book: manuscript.docx
# Date: 2026-03-17
#
# SCOPE: Typo-level fixes only. For anything structural,
# edit the Word manuscript directly.

corrections:
  # ── Spelling ──────────────────────────────────────────
  - find: "accomodate"
    replace: "accommodate"
    note: "spelling — caught in Kindle review"

  - find: "Etherium"
    replace: "Ethereum"
    note: "project name misspelling"

  # ── Duplicate words ──────────────────────────────────
  - find: "the the quick"
    replace: "the quick"
    chapter: "Soda Sweet as Blood"
    note: "duplicate word, paragraph 3"

  # ── Punctuation ──────────────────────────────────────
  - find: "said ,"
    replace: "said,"
    note: "extra space before comma"

  - find: "Interior design.layout"
    replace: "Interior design/layout"
    note: "period should be slash in copyright page"

  # ── Minor wording ────────────────────────────────────
  - find: "very unique"
    replace: "unique"
    note: "'unique' is already absolute"

  # ── Case / proper nouns ──────────────────────────────
  - find: "iphone"
    replace: "iPhone"
    note: "brand capitalization"

  # ── Whitespace ───────────────────────────────────────
  - find: "word  word"
    replace: "word word"
    note: "double space"
"""


MAX_FIND_LEN = 200
MAX_REPLACE_LEN = 200


def load_corrections(path):
    """Load and validate corrections from YAML file."""
    with open(path) as f:
        data = yaml.safe_load(f)

    if not data or "corrections" not in data:
        print("Error: YAML must have a top-level 'corrections' list", file=sys.stderr)
        sys.exit(1)

    corrections = []
    for i, c in enumerate(data["corrections"], 1):
        if "find" not in c or "replace" not in c:
            print(f"Error: correction #{i} missing 'find' or 'replace'", file=sys.stderr)
            sys.exit(1)

        find = c["find"]
        replace = c["replace"]

        if len(find) > MAX_FIND_LEN:
            print(f"Error: correction #{i} 'find' exceeds {MAX_FIND_LEN} chars — "
                  "this is too large for a correction. Edit the Word doc instead.",
                  file=sys.stderr)
            sys.exit(1)

        if len(replace) > MAX_REPLACE_LEN:
            print(f"Error: correction #{i} 'replace' exceeds {MAX_REPLACE_LEN} chars — "
                  "this is too large for a correction. Edit the Word doc instead.",
                  file=sys.stderr)
            sys.exit(1)

        if "<" in replace or ">" in replace:
            print(f"Error: correction #{i} 'replace' contains HTML/XML — "
                  "corrections are text-only. Edit the Word doc for formatting changes.",
                  file=sys.stderr)
            sys.exit(1)

        corrections.append({
            "find": find,
            "replace": replace,
            "chapter": c.get("chapter"),
            "note": c.get("note", ""),
            "index": i,
        })

    return corrections


def replace_in_paragraph(paragraph, find_text, replace_text):
    """
    Find and replace text in a paragraph, handling cross-run matches.

    Word splits paragraph text into runs based on formatting boundaries.
    A single word can span multiple runs, e.g.:
      Run 0: "The quick brown fox accom"
      Run 1: "odate"
      Run 2: "s change well."

    Strategy:
      1. Concatenate all run texts to get the full paragraph string.
      2. Find all occurrences of find_text in the concatenated string.
      3. For each match, map the character positions back to runs.
      4. Replace the matched text across those runs, preserving the
         formatting of the first run that contains match characters.

    Returns the number of replacements made.
    """
    runs = paragraph.runs
    if not runs:
        return 0

    # Build the concatenated text and a map of (run_index, char_offset_in_run)
    # for each character position in the full string.
    full_text = ""
    char_map = []  # char_map[i] = (run_index, offset_within_run)
    for ri, run in enumerate(runs):
        run_text = run.text or ""
        for ci in range(len(run_text)):
            char_map.append((ri, ci))
        full_text += run_text

    if find_text not in full_text:
        return 0

    # Find all non-overlapping match positions
    matches = []
    start = 0
    while True:
        pos = full_text.find(find_text, start)
        if pos == -1:
            break
        matches.append(pos)
        start = pos + len(find_text)

    if not matches:
        return 0

    # Process matches in reverse order so earlier character positions
    # remain valid after we modify later ones.
    for match_pos in reversed(matches):
        match_end = match_pos + len(find_text)

        # Determine which runs are affected
        first_run_idx, first_char_offset = char_map[match_pos]
        last_run_idx, last_char_offset = char_map[match_end - 1]

        if first_run_idx == last_run_idx:
            # Simple case: match is entirely within one run
            run = runs[first_run_idx]
            run_text = run.text
            run.text = (run_text[:first_char_offset] +
                        replace_text +
                        run_text[last_char_offset + 1:])
        else:
            # Cross-run match: the match spans from first_run_idx to last_run_idx.
            #
            # Approach:
            #   - In the first run, replace from first_char_offset onward
            #     with the replacement text.
            #   - In the last run, remove characters up to and including
            #     last_char_offset.
            #   - Clear all intermediate runs entirely.

            # First run: keep text before match, append replacement
            first_run = runs[first_run_idx]
            first_run.text = first_run.text[:first_char_offset] + replace_text

            # Intermediate runs: clear them
            for ri in range(first_run_idx + 1, last_run_idx):
                runs[ri].text = ""

            # Last run: remove the matched portion from the beginning
            last_run = runs[last_run_idx]
            last_run.text = last_run.text[last_char_offset + 1:]

        # Rebuild char_map and full_text for subsequent (earlier) matches
        full_text = ""
        char_map = []
        for ri, run in enumerate(runs):
            run_text = run.text or ""
            for ci in range(len(run_text)):
                char_map.append((ri, ci))
            full_text += run_text

    return len(matches)


def is_heading(paragraph):
    """Check if a paragraph is a heading style."""
    style_name = (paragraph.style.name or "").lower()
    return style_name.startswith("heading")


def apply_to_docx(docx_path, corrections, output_path, dry_run=False):
    """Apply corrections to a Word document."""
    doc = Document(docx_path)
    results = []

    # Build a chapter context map: for each paragraph index, track
    # the most recent heading text seen above it.
    paragraphs = doc.paragraphs
    current_chapter = None
    chapter_context = []  # parallel to paragraphs
    for para in paragraphs:
        if is_heading(para):
            current_chapter = para.text.strip()
        chapter_context.append(current_chapter)

    for corr in corrections:
        find_text = corr["find"]
        replace_text = corr["replace"]
        total_replacements = 0

        for pi, para in enumerate(paragraphs):
            # If chapter filter specified, check the chapter context
            if corr["chapter"]:
                ctx = chapter_context[pi]
                if ctx is None or corr["chapter"] not in ctx:
                    continue

            if dry_run:
                # For dry run, count matches without modifying
                full_text = "".join(run.text or "" for run in para.runs)
                count = full_text.count(find_text)
                total_replacements += count
            else:
                count = replace_in_paragraph(para, find_text, replace_text)
                total_replacements += count

        status = "APPLIED" if total_replacements > 0 else "NOT FOUND"
        if dry_run and total_replacements > 0:
            status = "WOULD APPLY"

        results.append({
            "index": corr["index"],
            "find": find_text,
            "replace": replace_text,
            "count": total_replacements,
            "status": status,
            "note": corr["note"],
        })

    # Save (only if not dry run and changes were made)
    if not dry_run and any(r["count"] > 0 for r in results):
        doc.save(output_path)

    return results


def print_results(results):
    """Print a summary table of corrections applied."""
    print(f"\n{'#':>3}  {'Status':<13} {'Count':>5}  {'Find':<30}  Note")
    print("\u2500" * 80)
    for r in results:
        find_display = r["find"][:28] + "\u2026" if len(r["find"]) > 28 else r["find"]
        note = r["note"][:25] if r["note"] else ""
        print(f"{r['index']:>3}  {r['status']:<13} {r['count']:>5}  {find_display:<30}  {note}")

    applied = sum(1 for r in results if r["count"] > 0)
    total = len(results)
    print(f"\n{applied}/{total} corrections applied, "
          f"{sum(r['count'] for r in results)} total replacements")


def main():
    parser = argparse.ArgumentParser(
        description="Apply corrections ledger to Word (.docx) files",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="For structural changes, edit the Word manuscript directly.")
    parser.add_argument("corrections", nargs="?", help="YAML corrections file")
    parser.add_argument("docx", nargs="?", help="Input Word (.docx) file")
    parser.add_argument("-o", "--output", help="Output .docx path (default: overwrite input)")
    parser.add_argument("--dry-run", action="store_true",
                        help="Show what would change without modifying anything")
    parser.add_argument("--example", action="store_true",
                        help="Print an example corrections YAML file")
    args = parser.parse_args()

    if args.example:
        print(EXAMPLE_CORRECTIONS)
        return

    if not args.corrections or not args.docx:
        parser.print_help()
        sys.exit(1)

    if not os.path.exists(args.corrections):
        print(f"Error: corrections file not found: {args.corrections}", file=sys.stderr)
        sys.exit(1)

    if not os.path.exists(args.docx):
        print(f"Error: Word file not found: {args.docx}", file=sys.stderr)
        sys.exit(1)

    corrections = load_corrections(args.corrections)
    print(f"Loaded {len(corrections)} corrections from {args.corrections}")
    print(f"{'DRY RUN — ' if args.dry_run else ''}Processing: {args.docx}")

    output = args.output or args.docx
    if not args.dry_run and output == args.docx:
        # Back up before overwriting
        backup = args.docx + ".bak"
        shutil.copy2(args.docx, backup)
        print(f"Backup: {backup}")

    results = apply_to_docx(args.docx, corrections, output, dry_run=args.dry_run)
    print_results(results)

    if not args.dry_run and any(r["count"] > 0 for r in results):
        print(f"\nOutput: {output}")


if __name__ == "__main__":
    main()
