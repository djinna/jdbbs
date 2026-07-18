#!/usr/bin/env python3
"""Parse the Zoothesia editorial STYLESHEET.md into a JSON seed for stylesheet_items.

Usage: python3 parse_stylesheet.py [SOURCE.md] [OUTPUT.json]
Defaults:
  SOURCE = /home/exedev/book-production/editorial/Zoothesia/STYLESHEET.md
  OUTPUT = /home/exedev/prodcal/db/seed/stylesheet_seed.json
"""
import json
import re
import sys

SRC = sys.argv[1] if len(sys.argv) > 1 else \
    "/home/exedev/book-production/editorial/Zoothesia/STYLESHEET.md"
OUT = sys.argv[2] if len(sys.argv) > 2 else \
    "/home/exedev/prodcal/db/seed/stylesheet_seed.json"


def is_separator(row):
    s = row.strip()
    if not s.startswith("|"):
        return False
    inner = s.strip("|")
    return "-" in inner and all(c in " -:|" for c in inner)


def split_pipe_cells(row):
    # A markdown table row: strip the leading/trailing pipe, split on |.
    s = row.strip()
    if s.startswith("|"):
        s = s[1:]
    if s.endswith("|"):
        s = s[:-1]
    return [c.strip() for c in s.split("|")]


def is_bullet_start(l):
    return l.startswith("- ")


def is_num_start(l):
    return re.match(r"^\d+\.\s", l) is not None


def is_heading(l):
    return l.startswith("#")


def is_hr(l):
    return l.strip() in ("---", "***", "___")


def starts_new_block(l):
    return (l.startswith("|") or l.startswith(">") or is_bullet_start(l)
            or is_num_start(l) or is_heading(l) or is_hr(l))


def parse_blocks(body):
    """Return list of ('rule', [cells]) or ('prose', text) in document order."""
    lines = body.split("\n")
    blocks = []
    i = 0
    n = len(lines)
    while i < n:
        line = lines[i]
        if line.strip() == "" or is_hr(line):
            i += 1
            continue
        if line.startswith("|"):
            # gather a contiguous run of pipe lines
            rows = []
            while i < n and lines[i].startswith("|"):
                rows.append(lines[i])
                i += 1
            for r in rows:
                if is_separator(r):
                    # the immediately-preceding emitted rule was the header row
                    if blocks and blocks[-1][0] == "rule":
                        blocks.pop()
                    continue
                cells = split_pipe_cells(r)
                blocks.append(("rule", cells))
            continue
        if line.startswith(">"):
            buf = []
            while i < n and lines[i].startswith(">"):
                buf.append(lines[i])
                i += 1
            blocks.append(("prose", "\n".join(buf)))
            continue
        if is_heading(line):
            blocks.append(("prose", line.rstrip()))
            i += 1
            continue
        # bullet, numbered item, or paragraph: consume with indented continuations
        buf = [line]
        i += 1
        while i < n and lines[i].strip() != "" and not starts_new_block(lines[i]):
            buf.append(lines[i])
            i += 1
        blocks.append(("prose", "\n".join(buf)))
    return blocks


def main():
    with open(SRC, encoding="utf-8") as f:
        text = f.read()

    lines = text.split("\n")
    # Locate section (## ) heading line indices.
    heads = [idx for idx, l in enumerate(lines) if l.startswith("## ")]

    items = []

    # ---- Overview: everything before the first '## ' heading ----
    overview_raw = "\n".join(lines[: heads[0]])
    # drop standalone hr lines but keep the h1 + intro paragraphs as one block
    overview_lines = [l for l in overview_raw.split("\n") if not is_hr(l)]
    overview = "\n".join(overview_lines).strip()
    items.append({
        "section_ord": -1,
        "section": "Overview",
        "item_ord": 1,
        "kind": "prose",
        "col1": "", "col2": "", "col3": "",
        "body": overview,
    })

    # ---- Each ## section ----
    seq_fallback = 100  # for headings with no leading number
    for hi, start in enumerate(heads):
        end = heads[hi + 1] if hi + 1 < len(heads) else len(lines)
        heading = lines[start][3:].strip()  # drop '## '
        body = "\n".join(lines[start + 1:end])
        m = re.match(r"^(\d+)\.", heading)
        if m:
            section_ord = int(m.group(1))
        else:
            section_ord = seq_fallback
            seq_fallback += 1
        blocks = parse_blocks(body)
        item_ord = 0
        for kind, payload in blocks:
            item_ord += 1
            if kind == "rule":
                cells = payload
                col1 = cells[0] if len(cells) > 0 else ""
                col2 = cells[1] if len(cells) > 1 else ""
                col3 = cells[2] if len(cells) > 2 else ""
                items.append({
                    "section_ord": section_ord,
                    "section": heading,
                    "item_ord": item_ord,
                    "kind": "rule",
                    "col1": col1, "col2": col2, "col3": col3,
                    "body": "",
                })
            else:
                items.append({
                    "section_ord": section_ord,
                    "section": heading,
                    "item_ord": item_ord,
                    "kind": "prose",
                    "col1": "", "col2": "", "col3": "",
                    "body": payload,
                })

    with open(OUT, "w", encoding="utf-8") as f:
        json.dump(items, f, ensure_ascii=False, indent=2)
        f.write("\n")

    # ---- Validation report ----
    with open(OUT, encoding="utf-8") as f:
        reloaded = json.load(f)
    assert reloaded == items, "round-trip mismatch"

    from collections import Counter, OrderedDict
    kinds = Counter(it["kind"] for it in items)
    print("OUTPUT:", OUT)
    print("total items:", len(items))
    print("by kind:", dict(kinds))
    print("sections (section_ord, section, #items):")
    seen = OrderedDict()
    for it in items:
        key = (it["section_ord"], it["section"])
        seen[key] = seen.get(key, 0) + 1
    for (ord_, name), cnt in seen.items():
        print(f"  {ord_:>3}  {name!r}  -> {cnt} items")


if __name__ == "__main__":
    main()
