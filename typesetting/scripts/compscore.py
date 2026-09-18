#!/usr/bin/env python3
"""compscore.py — composition-quality scorer for a built print PDF (read-only).

Reads `pdftotext -bbox-layout` output and reports, per page and in total:
  loose      lines whose mean word gap exceeds LOOSE_EM (default 0.5 em)
  tight      lines whose mean word gap is under TIGHT_EM (default 0.16 em)
  runt       paragraph last lines that are a single word (or < RUNT_CHARS)
  widow      a paragraph's last line alone at the top of a page
  orphan     a paragraph's first line alone at the bottom of a page
  hyph3      3+ consecutive lines ending in a hyphen
  stack      the same word starting or ending 3+ consecutive lines
  river      word gaps vertically aligned (±RIVER_PT) on 3+ consecutive lines

Nothing here changes output; it is the yardstick for P3 (punch list 5.5).

  python3 typesetting/scripts/compscore.py book.pdf            # summary
  python3 typesetting/scripts/compscore.py book.pdf --list     # every hit
  python3 typesetting/scripts/compscore.py book.pdf --json
"""
import argparse, json, re, subprocess, sys, xml.etree.ElementTree as ET
from collections import Counter, defaultdict

LOOSE_EM, TIGHT_EM, RUNT_CHARS, RIVER_PT, MIN_WORDS = 0.5, 0.16, 6, 1.5, 4
NS = {"x": "http://www.w3.org/1999/xhtml"}


def load(pdf):
    xml = subprocess.run(["pdftotext", "-bbox-layout", pdf, "-"], check=True,
                         capture_output=True, text=True).stdout
    xml = re.sub(r"<!DOCTYPE[^>]*>", "", xml, count=1)
    root = ET.fromstring(xml)
    pages = []
    for pno, page in enumerate(root.iter("{%s}page" % NS["x"]), 1):
        lines = []
        for ln in page.iter("{%s}line" % NS["x"]):
            words = [(float(w.get("xMin")), float(w.get("xMax")), (w.text or "").strip())
                     for w in ln.iter("{%s}word" % NS["x"])]
            if not words:
                continue
            lines.append(dict(page=pno, x0=float(ln.get("xMin")), x1=float(ln.get("xMax")),
                              y0=float(ln.get("yMin")), y1=float(ln.get("yMax")), words=words))
        lines.sort(key=lambda l: (l["y0"], l["x0"]))
        pages.append(lines)
    return pages


def body_lines(lines):
    """Keep lines that belong to the main text column: the modal x0 (left edge)
    and the modal line height, ± tolerance. Drops folios, running heads, titles."""
    if not lines:
        return [], None
    x0s = Counter(round(l["x0"]) for l in lines)
    hs = Counter(round(l["y1"] - l["y0"], 1) for l in lines)
    col_x0 = x0s.most_common(1)[0][0]
    h = hs.most_common(1)[0][0]
    body = [l for l in lines if abs((l["y1"] - l["y0"]) - h) < 1.5 and l["x0"] >= col_x0 - 1 and l["x0"] <= col_x0 + 4 * h]
    if len(body) < 3:
        return [], None
    measure = max(l["x1"] for l in body)
    return body, dict(x0=col_x0, x1=measure, h=h, em=h / 1.2)


def analyse(pages, opts):
    hits, per_page = [], defaultdict(Counter)
    prev_page_last_para_open = False  # previous page ended mid-paragraph
    for pno, lines in enumerate(pages, 1):
        body, col = body_lines(lines)
        if not body:
            continue
        em, x0, x1 = col["em"], col["x0"], col["x1"]
        n = len(body)
        info = []
        for i, l in enumerate(body):
            ws = l["words"]
            text = " ".join(w[2] for w in ws)
            indented = l["x0"] > x0 + 0.4 * em
            full = l["x1"] >= x1 - 0.6 * em
            gaps = [ws[k + 1][0] - ws[k][1] for k in range(len(ws) - 1)]
            mean_gap = sum(gaps) / len(gaps) if gaps else 0
            info.append(dict(text=text, indented=indented, full=full, gaps=gaps, x0=l["x0"],
                             gap_centres=[(ws[k][1] + ws[k + 1][0]) / 2 for k in range(len(ws) - 1)],
                             mean_gap=mean_gap, nwords=len(ws)))

        def hit(kind, i, extra=""):
            per_page[pno][kind] += 1
            hits.append((pno, kind, info[i]["text"][:70], extra))

        for i, li in enumerate(info):
            # paragraph starts: indented lines (or first line after a short line)
            # An indented line directly under another line with the same left
            # edge is a hanging block (list item, block quote), not a new paragraph.
            same_edge = i > 0 and info[i - 1]["indented"] and abs(info[i - 1]["x0"] - li["x0"]) < 1.0
            li["para_start"] = (li["indented"] and not same_edge) or (i > 0 and not info[i - 1]["full"])
            li["para_end"] = not li["full"] or (i + 1 < n and info[i + 1]["indented"])
        for i, li in enumerate(info):
            if li["full"] and li["nwords"] >= MIN_WORDS:
                if li["mean_gap"] > opts.loose * em:
                    hit("loose", i, "gap %.2f em" % (li["mean_gap"] / em))
                elif li["mean_gap"] < opts.tight * em:
                    hit("tight", i, "gap %.2f em" % (li["mean_gap"] / em))
            if li["para_end"] and not li["full"] and not li["para_start"]:
                last = li["text"]
                if li["nwords"] == 1 or len(last) < opts.runt_chars:
                    hit("runt", i, repr(last))
            if i == 0 and li["para_end"] and not li["full"] and not li["para_start"] and prev_page_last_para_open:
                hit("widow", i)
            if i == n - 1 and li["para_start"] and not li["para_end"]:
                hit("orphan", i)
        prev_page_last_para_open = not info[-1]["para_end"]
        # runs: hyphen endings and stacks
        run = 0
        for i, li in enumerate(info):
            run = run + 1 if li["text"].endswith("-") else 0
            if run == 3:
                hit("hyph3", i)
        for side in (0, -1):
            run = 1
            for i in range(1, n):
                a = re.sub(r"\W", "", info[i]["text"].split()[side].lower()) if info[i]["text"].split() else ""
                b = re.sub(r"\W", "", info[i - 1]["text"].split()[side].lower()) if info[i - 1]["text"].split() else ""
                run = run + 1 if a and a == b and len(a) > 2 else 1
                if run == 3:
                    hit("stack", i, ("start" if side == 0 else "end") + " " + repr(a))
        # rivers: gap centres aligned across 3+ consecutive full lines
        for i in range(2, n):
            if not all(info[j]["full"] for j in (i - 2, i - 1, i)):
                continue
            for c in info[i]["gap_centres"]:
                if any(abs(c - d) <= opts.river for d in info[i - 1]["gap_centres"]) and \
                   any(abs(c - d) <= opts.river for d in info[i - 2]["gap_centres"]):
                    hit("river", i, "x=%.0f" % c)
                    break
    return hits, per_page


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("pdf")
    ap.add_argument("--list", action="store_true", help="print every hit")
    ap.add_argument("--json", action="store_true")
    ap.add_argument("--loose", type=float, default=LOOSE_EM)
    ap.add_argument("--tight", type=float, default=TIGHT_EM)
    ap.add_argument("--runt-chars", type=int, default=RUNT_CHARS)
    ap.add_argument("--river", type=float, default=RIVER_PT)
    o = ap.parse_args()
    pages = load(o.pdf)
    hits, per_page = analyse(pages, o)
    kinds = ["loose", "tight", "runt", "widow", "orphan", "hyph3", "stack", "river"]
    tot = Counter(k for _, k, _, _ in hits)
    body_pages = sum(1 for p in pages if body_lines(p)[0])
    body_lines_n = sum(len(body_lines(p)[0]) for p in pages)
    if o.json:
        print(json.dumps(dict(pages=len(pages), body_pages=body_pages, body_lines=body_lines_n,
                              totals={k: tot[k] for k in kinds},
                              per_page={p: dict(c) for p, c in sorted(per_page.items())},
                              hits=[dict(page=p, kind=k, text=t, note=x) for p, k, t, x in hits]), indent=1))
        return
    print(f"{o.pdf}: {len(pages)} pages, {body_pages} with body text, {body_lines_n} body lines")
    print("  " + "  ".join(f"{k}={tot[k]}" for k in kinds))
    per100 = {k: 100 * tot[k] / max(body_lines_n, 1) for k in ("loose", "tight", "river")}
    print("  per 100 body lines: " + "  ".join(f"{k}={v:.1f}" for k, v in per100.items()))
    print(f"  per body page: runt={tot['runt']/max(body_pages,1):.2f}  widow+orphan={(tot['widow']+tot['orphan'])/max(body_pages,1):.2f}")
    if o.list:
        for p, k, t, x in hits:
            print(f"  p{p:<4} {k:<7} {x:<14} {t}")


if __name__ == "__main__":
    main()
