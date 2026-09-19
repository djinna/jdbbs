"""Generate the [jdbb] studio wordmark SVGs and the social card.
Outlines from JetBrains Mono (bundled, OFL) so the files need no font.
Run from repo root: python3 docs/brand/wordmark-gen.py  -> srv/static/brand/
"""
import subprocess, pathlib
from fontTools.ttLib import TTFont
from fontTools.pens.svgPathPen import SVGPathPen
from fontTools.pens.transformPen import TransformPen

FONT = "typesetting/fonts/jetbrainsmono/fonts/ttf/JetBrainsMono-%s.ttf"
INK, ACC, PAPER, SEC, SEC_DARK = "#0E1116", "#007699", "#FFFFFF", "#5D6B76", "#8A97A1"
OUT = pathlib.Path("srv/static/brand"); OUT.mkdir(exist_ok=True)
_fonts = {}
def font(w):
    if w not in _fonts:
        f = TTFont(FONT % w); _fonts[w] = (f.getGlyphSet(), f.getBestCmap(), f["head"].unitsPerEm)
    return _fonts[w]

def run(weight, text, size, x, y, track=0):
    gs, cmap, upm = font(weight); sc = size / upm; d = []; cx = x
    for ch in text:
        g = gs[cmap[ord(ch)]]; pen = SVGPathPen(gs)
        g.draw(TransformPen(pen, (sc, 0, 0, -sc, cx, y))); d.append(pen.getCommands())
        cx += g.width * sc + track * size
    return " ".join(d), cx - x

def wordmark(size, ink, acc, sec, studio=True, x=0, y=0):
    """Returns (svg paths, total width). Baseline at y."""
    parts = []
    p, w = run("Bold", "[", size, x, y); parts.append(f'<path d="{p}" fill="{acc}"/>'); x += w
    p, w = run("Bold", "jdbb", size, x, y, track=-0.02); parts.append(f'<path d="{p}" fill="{ink}"/>'); x += w
    p, w = run("Bold", "]", size, x, y); parts.append(f'<path d="{p}" fill="{acc}"/>'); x += w
    if studio:
        x += size * 0.45
        p, w = run("Regular", "studio", size, x, y); parts.append(f'<path d="{p}" fill="{sec}"/>'); x += w
    return "".join(parts), x

def svg(w, h, body, bg=None):
    rect = f'<rect width="{w}" height="{h}" fill="{bg}"/>' if bg else ""
    return f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {w} {h}" width="{w}" height="{h}">{rect}{body}</svg>'

S = 100
for name, ink, sec, bg in (("jdbb-wordmark", INK, SEC, None), ("jdbb-wordmark-paper", PAPER, SEC_DARK, None)):
    body, w = wordmark(S, ink, ACC, sec, x=S * 0.1, y=S * 0.95)
    (OUT / f"{name}.svg").write_text(svg(round(w + S * 0.1), round(S * 1.3), body))
body, w = wordmark(S, INK, ACC, SEC, studio=False, x=S * 0.1, y=S * 0.95)
(OUT / "jdbb-wordmark-short.svg").write_text(svg(round(w + S * 0.1), round(S * 1.3), body))

# Social card 1200x630: paper ground, wordmark, one line, url. Rule at top.
W, H = 1200, 630
wm, ww = wordmark(96, INK, ACC, SEC, x=96, y=330)
line, _ = run("Regular", "Books from Word files.", 34, 96, 420)
url, _ = run("Regular", "jdbbs.exe.xyz", 26, 96, 540)
card = (f'<rect x="96" y="96" width="{W-192}" height="2" fill="{INK}"/>' + wm +
        f'<path d="{line}" fill="{SEC}"/><path d="{url}" fill="{ACC}"/>')
(OUT / "jdbb-social-card.svg").write_text(svg(W, H, card, bg=PAPER))
subprocess.run(["convert", "-density", "144", str(OUT / "jdbb-social-card.svg"), "-resize", f"{W}x{H}",
                "-quality", "92", str(OUT / "jdbb-social-card.jpg")], check=True)
print(sorted(p.name for p in OUT.iterdir()))
