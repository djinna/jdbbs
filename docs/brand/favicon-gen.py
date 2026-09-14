from fontTools.ttLib import TTFont
from fontTools.pens.svgPathPen import SVGPathPen
from fontTools.pens.transformPen import TransformPen
import sys
FONT="typesetting/fonts/jetbrainsmono/fonts/ttf/JetBrainsMono-%s.ttf"
def glyphs(weight):
    f=TTFont(FONT%weight); gs=f.getGlyphSet(); cmap=f.getBestCmap(); upm=f['head'].unitsPerEm
    return f,gs,cmap,upm
def text_path(weight, text, size, x, y):
    """Return SVG path d for text at baseline (x,y) with font-size size, plus advance width."""
    f,gs,cmap,upm=glyphs(weight); scale=size/upm; d=[]; cx=x
    for ch in text:
        g=gs[cmap[ord(ch)]]; pen=SVGPathPen(gs)
        tp=TransformPen(pen,(scale,0,0,-scale,cx,y)); g.draw(tp); d.append(pen.getCommands()); cx+=g.width*scale
    return " ".join(d), cx-x
INK="#0E1116"; ACC="#007699"; PAPER="#F6F4EF"
def tile(fill,rx=12): return f'<rect width="64" height="64" rx="{rx}" fill="{fill}"/>'
def svg(body): return f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64">{body}</svg>'
out={}
# A: [jd] light tile, brackets accent
p,w=text_path("Bold","[jd]",30,0,0); x=(64-w)/2
pb,_=text_path("ExtraBold","[",30,x,44); pj,wj=text_path("Bold","jd",30,x+w/4,44); pb2,_=text_path("ExtraBold","]",30,x+3*w/4,44)
out["A-brackets-light"]=svg(tile(PAPER)+f'<path d="{pb}" fill="{ACC}"/><path d="{pj}" fill="{INK}"/><path d="{pb2}" fill="{ACC}"/>')
# B: jd big, dark tile, paper text, accent j
p,w=text_path("ExtraBold","jd",40,0,0); x=(64-w)/2
pj,wj=text_path("ExtraBold","j",40,x,48); pd,_=text_path("ExtraBold","d",40,x+wj,48)
out["B-jd-dark"]=svg(tile(INK)+f'<path d="{pj}" fill="{ACC}"/><path d="{pd}" fill="{PAPER}"/>')
# C: jd big, light tile, ink text, accent j
out["C-jd-light"]=svg(tile(PAPER)+f'<path d="{pj}" fill="{ACC}"/><path d="{pd}" fill="{INK}"/>')
# D: jd big, accent tile, paper text
out["D-jd-accent"]=svg(tile(ACC)+f'<path d="{pj}" fill="{PAPER}"/><path d="{pd}" fill="{PAPER}"/>')
# E: [jd] dark tile
p,w=text_path("Bold","[jd]",30,0,0); x=(64-w)/2
pb,_=text_path("ExtraBold","[",30,x,44); pj2,_=text_path("Bold","jd",30,x+w/4,44); pb2,_=text_path("ExtraBold","]",30,x+3*w/4,44)
out["E-brackets-dark"]=svg(tile(INK)+f'<path d="{pb}" fill="{ACC}"/><path d="{pj2}" fill="{PAPER}"/><path d="{pb2}" fill="{ACC}"/>')
for k,v in out.items(): open(f"scratch/favicon/{k}.svg","w").write(v)
html='<body style="background:#ddd;font:12px sans-serif;padding:20px">'
for bg in ["#fff","#202124"]:
    html+=f'<div style="background:{bg};padding:16px;margin-bottom:12px;display:flex;gap:30px;align-items:flex-end">'
    for k,v in out.items():
        html+=f'<div style="text-align:center;color:#888"><div style="display:flex;gap:10px;align-items:flex-end;justify-content:center">'
        for s in (16,32,64): html+=f'<img src="{k}.svg" width="{s}" height="{s}">'
        html+=f'</div>{k}</div>'
    html+='</div>'
open("scratch/favicon/index.html","w").write(html+"</body>")

# Variant round 2: tighter letter-fit (negative tracking), plus "bb"
def pair(weight, a, b, size, track):
    pa,wa=text_path(weight,a,size,0,0); pb_,wb=text_path(weight,b,size,0,0)
    w=wa+wb+track; x=(64-w)/2; y=32+size*0.36
    pa,_=text_path(weight,a,size,x,y); pb_,_=text_path(weight,b,size,x+wa+track,y)
    return pa,pb_
out2={}
for name,(a,b) in {"jd":("j","d"),"bb":("b","b")}.items():
    for track in (0,-4,-7):
        pa,pb_=pair("ExtraBold",a,b,42,track)
        out2[f"{name}-t{track}"]=svg(tile(INK)+f'<path d="{pa}" fill="{ACC}"/><path d="{pb_}" fill="{PAPER}"/>')
for k,v in out2.items(): open(f"scratch/favicon/{k}.svg","w").write(v)
html='<body style="background:#ddd;font:12px sans-serif;padding:10px">'
for bg in ["#fff","#202124"]:
    html+=f'<div style="background:{bg};padding:16px;margin-bottom:12px;display:flex;gap:22px;align-items:flex-end;flex-wrap:wrap">'
    for k,v in out2.items():
        html+=f'<div style="text-align:center;color:#888"><div style="display:flex;gap:10px;align-items:flex-end;justify-content:center">'
        for s in (16,32,64): html+=f'<img src="{k}.svg" width="{s}" height="{s}">'
        html+=f'</div>{k}</div>'
    html+='</div>'
open("scratch/favicon/round2.html","w").write(html+"</body>")
