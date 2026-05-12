#!/usr/bin/env python3
"""
apply-corrections.py — Apply a corrections ledger to EPUB files.

Corrections are small, discrete text fixes applied directly to EPUB XHTML
content. This tool is for typo-level changes ONLY.

WHAT THIS TOOL HANDLES (the "corrections" tier):
  - Typo fixes:           "accomodate" → "accommodate"
  - Punctuation fixes:    "said ," → "said,"
  - Duplicate words:      "the the quick" → "the quick"
  - Case fixes:           "iphone" → "iPhone"
  - Minor wording tweaks: "very unique" → "unique"
  - Name/term fixes:      "Etherium" → "Ethereum"
  - Whitespace cleanup:   "word  word" → "word word"

WHAT THIS TOOL DOES NOT HANDLE (edit the Word doc instead):
  - Adding or removing sentences or paragraphs
  - Reordering content
  - Changing formatting (bold, italic, styles)
  - Modifying images, captions, or figures
  - Structural changes (chapters, headings, TOC)
  - Any change that touches HTML tags or attributes
  - Any replacement longer than ~200 characters

For anything beyond simple text substitution, edit the canonical Word
manuscript and re-run the full production pipeline.

Usage:
  python3 apply-corrections.py corrections.yaml BOOK.epub [-o OUTPUT.epub]
  python3 apply-corrections.py corrections.yaml BOOK.epub --dry-run
  python3 apply-corrections.py --example > corrections.yaml

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
import re
import shutil
import sys
import tempfile
import zipfile
from pathlib import Path

try:
    import yaml
except ImportError:
    print("PyYAML required: pip install pyyaml", file=sys.stderr)
    sys.exit(1)


EXAMPLE_CORRECTIONS = """\
# Corrections ledger for: Ghosts in Machines
# Book: GHOSTS_IN_MACHINES_ALT.epub
# Date: 2026-03-17
#
# SCOPE: Typo-level fixes only. For anything structural,
# edit the Word manuscript and rebuild via the pipeline.

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
            print(f"Error: correction #{i} 'replace' contains HTML — "
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


def is_content_xhtml(filename):
    """Check if a file inside the EPUB is an XHTML content file."""
    if not filename.endswith((".xhtml", ".html", ".htm")):
        return False
    # Skip navigation/TOC files
    basename = os.path.basename(filename).lower()
    if basename in ("toc.xhtml", "toc.html", "nav.xhtml", "cover.xhtml"):
        return False
    return True


def apply_to_epub(epub_path, corrections, output_path, dry_run=False):
    """Apply corrections to EPUB content files."""
    results = []

    with tempfile.TemporaryDirectory() as tmpdir:
        # Extract
        with zipfile.ZipFile(epub_path, "r") as zin:
            zin.extractall(tmpdir)
            namelist = zin.namelist()

        # Find content XHTML files
        content_files = [n for n in namelist if is_content_xhtml(n)]

        if not content_files:
            print("Warning: no content XHTML files found in EPUB", file=sys.stderr)
            return results

        # Apply each correction
        for corr in corrections:
            find_text = corr["find"]
            replace_text = corr["replace"]
            total_replacements = 0

            for cf in content_files:
                filepath = os.path.join(tmpdir, cf)
                with open(filepath, "r", encoding="utf-8") as f:
                    content = f.read()

                # If chapter filter specified, check if this file contains that chapter
                if corr["chapter"]:
                    if corr["chapter"] not in content:
                        continue

                # Count and apply replacements (text-only, outside of tags)
                # We use a careful approach: only replace within text nodes,
                # not inside HTML tags or attributes
                new_content = safe_text_replace(content, find_text, replace_text)
                count = content.count(find_text) - new_content.count(find_text)

                if count > 0:
                    total_replacements += count
                    if not dry_run:
                        with open(filepath, "w", encoding="utf-8") as f:
                            f.write(new_content)

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

        # Repackage EPUB (only if not dry run and changes were made)
        if not dry_run and any(r["count"] > 0 for r in results):
            repackage_epub(tmpdir, namelist, output_path)

    return results


def safe_text_replace(html_content, find, replace):
    """Replace text only within text nodes, not inside HTML tags."""
    # Split on tags, only replace in text segments
    parts = re.split(r"(<[^>]+>)", html_content)
    for i, part in enumerate(parts):
        if not part.startswith("<"):
            parts[i] = part.replace(find, replace)
    return "".join(parts)


def repackage_epub(tmpdir, namelist, output_path):
    """Repackage extracted EPUB directory back into a .epub file."""
    with zipfile.ZipFile(output_path, "w", zipfile.ZIP_DEFLATED) as zout:
        # mimetype must be first and uncompressed
        mimetype_path = os.path.join(tmpdir, "mimetype")
        if os.path.exists(mimetype_path):
            zout.write(mimetype_path, "mimetype", compress_type=zipfile.ZIP_STORED)

        for name in namelist:
            if name == "mimetype":
                continue
            filepath = os.path.join(tmpdir, name)
            if os.path.exists(filepath):
                zout.write(filepath, name)


def print_results(results):
    """Print a summary table of corrections applied."""
    print(f"\n{'#':>3}  {'Status':<13} {'Count':>5}  {'Find':<30}  Note")
    print("─" * 80)
    for r in results:
        find_display = r["find"][:28] + "…" if len(r["find"]) > 28 else r["find"]
        note = r["note"][:25] if r["note"] else ""
        print(f"{r['index']:>3}  {r['status']:<13} {r['count']:>5}  {find_display:<30}  {note}")

    applied = sum(1 for r in results if r["count"] > 0)
    total = len(results)
    print(f"\n{applied}/{total} corrections applied, "
          f"{sum(r['count'] for r in results)} total replacements")


def main():
    parser = argparse.ArgumentParser(
        description="Apply corrections ledger to EPUB files",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="For structural changes, edit the Word manuscript and rebuild.")
    parser.add_argument("corrections", nargs="?", help="YAML corrections file")
    parser.add_argument("epub", nargs="?", help="Input EPUB file")
    parser.add_argument("-o", "--output", help="Output EPUB path (default: overwrite input)")
    parser.add_argument("--dry-run", action="store_true",
                        help="Show what would change without modifying anything")
    parser.add_argument("--example", action="store_true",
                        help="Print an example corrections YAML file")
    args = parser.parse_args()

    if args.example:
        print(EXAMPLE_CORRECTIONS)
        return

    if not args.corrections or not args.epub:
        parser.print_help()
        sys.exit(1)

    if not os.path.exists(args.corrections):
        print(f"Error: corrections file not found: {args.corrections}", file=sys.stderr)
        sys.exit(1)

    if not os.path.exists(args.epub):
        print(f"Error: EPUB file not found: {args.epub}", file=sys.stderr)
        sys.exit(1)

    corrections = load_corrections(args.corrections)
    print(f"Loaded {len(corrections)} corrections from {args.corrections}")
    print(f"{'DRY RUN — ' if args.dry_run else ''}Processing: {args.epub}")

    output = args.output or args.epub
    if not args.dry_run and output == args.epub:
        # Back up before overwriting
        backup = args.epub + ".bak"
        shutil.copy2(args.epub, backup)
        print(f"Backup: {backup}")

    results = apply_to_epub(args.epub, corrections, output, dry_run=args.dry_run)
    print_results(results)

    if not args.dry_run and any(r["count"] > 0 for r in results):
        print(f"\nOutput: {output}")


if __name__ == "__main__":
    main()
