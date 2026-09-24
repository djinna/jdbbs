#!/usr/bin/env python3
"""Export the live punchlist (checklist + every note thread) to docs/runs/.

    python3 scripts/punchlist-export.py            # -> docs/runs/PUNCHLIST-<today>.md
    python3 scripts/punchlist-export.py 2026-09-18 # explicit date

Writes the checklist with each item's notes nested beneath it as a blockquote
(who · when), then rebuilds docs/runs/README.md as an index of every file in
the directory. The admin page /admin/runs/ renders these files from disk.
"""
import json, os, re, shutil, sys, time

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
MD = os.path.join(ROOT, "scratch/run/CHECKLIST.md")
NOTES = os.path.join(ROOT, "scratch/run/notes.json")
OUT = os.path.join(ROOT, "docs/runs")
ITEM = re.compile(r"^(\s*)- \[[ x~]\] (\d+\.\d+[a-z]?)\b")

IMG_SRC = os.path.join(ROOT, "scratch/run/img")
IMG_OUT = os.path.join(OUT, "img")

def archive_img(url):
    """Copy a runpage image (/img/name) into docs/runs/img/ and return the
    relative markdown path. Screenshots pasted on the punchlist travel with
    the archive; the scratch dir is gitignored."""
    name = os.path.basename(url)
    src = os.path.join(IMG_SRC, name)
    if os.path.isfile(src):
        os.makedirs(IMG_OUT, exist_ok=True)
        dst = os.path.join(IMG_OUT, name)
        if not os.path.exists(dst): shutil.copyfile(src, dst)
    return "img/" + name

def export(date):
    md = open(MD).read().rstrip("\n").split("\n")
    md = [re.sub(r"\]\(/img/([^)]+)\)", lambda m: "](" + archive_img(m.group(1)) + ")", ln) for ln in md]
    try: notes = json.load(open(NOTES))
    except FileNotFoundError: notes = {}
    out = []
    private = False
    for i, ln in enumerate(md):
        # Sections whose heading carries "(private)" stay in scratch: they hold
        # operational security detail that must not reach the public repo.
        if ln.startswith("## "):
            private = "(private)" in ln.lower()
            if private:
                out.append(ln.replace("(private)", "").rstrip() + " — kept private, not exported")
                out.append("")
                continue
        if private:
            continue
        out.append(ln)
        m = ITEM.match(ln)
        if not m or m.group(2) not in notes: continue
        ind = m.group(1) + "  "
        # a nested list follows this item? then the notes must come before it
        # to stay attached; pandoc keeps both in the same <li>.
        for n in notes[m.group(2)]:
            out.append("")
            out.append(f"{ind}> **{n.get('who','?')}** · {n.get('ts','')}  ")
            for t in (n.get("text") or "").split("\n"):
                out.append(f"{ind}> {t}" if t.strip() else f"{ind}>")
            for u in n.get("images") or []:
                out.append(f"{ind}>")
                out.append(f"{ind}> ![screenshot]({archive_img(u)})")
        out.append("")
    stamp = time.strftime("%Y-%m-%d %H:%M UTC", time.gmtime())
    head = [f"<!-- exported {stamp} by scripts/punchlist-export.py; source scratch/run/CHECKLIST.md + notes.json -->", ""]
    path = os.path.join(OUT, f"PUNCHLIST-{date}.md")
    open(path, "w").write("\n".join(head + out) + "\n")
    return path

def index():
    files = sorted(f for f in os.listdir(OUT) if f.endswith(".md") and f != "README.md")
    lines = ["# Runs", "", "Punchlists and run logs, one file per day or event. Exported by",
             "`scripts/punchlist-export.py`; rendered read-only at `/admin/runs/`.", ""]
    for f in reversed(files):
        title = f[:-3]
        with open(os.path.join(OUT, f)) as fh:
            for ln in fh:
                if ln.startswith("# "): title = ln[2:].strip(); break
        lines.append(f"- [{title}]({f})")
    open(os.path.join(OUT, "README.md"), "w").write("\n".join(lines) + "\n")

if __name__ == "__main__":
    date = sys.argv[1] if len(sys.argv) > 1 else time.strftime("%Y-%m-%d", time.gmtime())
    p = export(date); index(); print(p)
