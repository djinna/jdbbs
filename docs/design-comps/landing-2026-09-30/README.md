# Minimalist landing comps (2026-09-30)

Four comps for a quieter `/`, after exe.dev's homepage, plus an overview page
(`index.html`) with the content-split table (what leaves `/` for `/factory`)
and open questions. Not deployed; nothing here is served by prodcal.

- Edit `*.src.html`, then `./build.sh` (expands `<!--HEAD-->` from `head.inc`
  and the comp switcher). `static` is a symlink to `srv/static`.
- Preview: `tmux new -d -s comps "busybox httpd -f -p 8767 -h $PWD"` →
  https://jdbbs.exe.xyz:8767/
