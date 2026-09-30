#!/bin/sh
# Expand <!--HEAD--> and <!--COMPNAV--> in *.src.html -> *.html
cd "$(dirname "$0")"
for f in *.src.html; do
  out="${f%.src.html}.html"
  python3 - "$f" "$out" <<'PY'
import sys,re
src=open(sys.argv[1]).read(); name=sys.argv[2]
head=open('head.inc').read()
comps=[('index.html','overview'),('a-quiet.html','A quiet'),('b-page.html','B page'),('c-terminal.html','C terminal'),('d-card.html','D card')]
nav='<div class="compnav">'+''.join(f'<a href="{h}"'+(' aria-current="page"' if h==name else '')+f'>{t}</a>' for h,t in comps)+'</div>'
open(name,'w').write(src.replace('<!--HEAD-->',head).replace('<!--COMPNAV-->',nav))
PY
done
