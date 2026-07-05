# jdbb studio design system — "Terminal Folio"

Locked 2026-07-05 after a five-round /design-shotgun exploration. Approved
mockups (portable HTML) live outside the repo at
`~/.gstack/projects/djinna-jdbbs/designs/site-redesign-20260704/` (`r3-*`,
`r4-*`). Canonical tokens: [`srv/static/theme.css`](../srv/static/theme.css).
Canonical theme/font state: [`srv/static/theme.js`](../srv/static/theme.js).

## Positioning

Book production run as an engineering discipline. The site borrows its visual
language from print production itself (job tickets, ledgers, pipeline
readouts) — engineering credibility that still says "books." The site is also
a demo of the studio's typesetting craft: the header font selector (a
csszengarden-style "same content, different setting" move) is a signature
element, not a toy.

## Hard rules (the anti-"LLM tell" canon)

1. **No cards. No boxes.** No bordered containers, no background-tinted
   rectangles wrapping content, no pills. Exceptions: real data tables
   (hairline row rules + 1px strong rules top/bottom), left-ruled indented
   callout blocks (2px accent left rule, no fill), and modal overlays
   (`--surface` + 1px `--border-strong`).
2. **Zero `border-radius`. Zero shadows. Zero gradients.** Also retired: the
   noise texture, the gradient wash, the dashed grid lines from the old
   design (neutralized in theme.css).
3. Structure comes from: full-width hairline rules (`--border`), strong 1px
   ink rules (`--border-strong`) for major breaks, whitespace, and
   shared-grid aligned columns (ledger discipline — sibling rows share one
   `grid-template-columns`).
4. **Tags are plain bracketed mono text**: `[PREFLIGHT]`, `[ACTION NEEDED]`
   (accent), `[ON TRACK]` (green), `[53/53]`. Never a pill, never a filled
   badge. Use `.tag` (+ `.ok/.warn/.err`).
5. **One filled button per page** (`.btn-fill`, `--accent-fill`). Every other
   action is an underlined text link (`.link-action`). Inputs are bare with a
   bottom rule only (`.input-bare`).
6. Section labels are uppercase mono kickers rendered as `// LABEL`
   (`.kicker`). Right-aligned meta on the same line is muted mono.
7. Progress is a bare 2px line (`.progress-line`), no track box.
8. No emoji in UI chrome. Mono glyphs (§ ▸ ▾ ✓ ● ○ →) are the icon set.

## Palette ("Process" — print CMYK heritage)

All values in theme.css. Light: paper `#F7F9FA`, ink `#0E1116`, accent text
`#007699`, accent fill `#00A8E0`, green `#2F7D4F`, red `#B3261E`. Dark: bg
`#0C0E12`, text `#E8E6E1`, accent `#3BC3F2`, fill `#00A8E0`. The accent is
process cyan — never substitute a generic SaaS blue (`#2563eb` is banned),
never warm cream/amber (reads as Claude-brand).

Semantic colors: green = done/pass, accent = active/attention, red =
deletions/errors only, muted = pending.

## Type

Mono (`--mono`) for headings, labels, data, controls, wordmark; body sans
(`--body`) for prose. Four pairings, switched by the header selector via
`html[data-font]`:

| key | mono | body |
|---|---|---|
| `jetbrains` (default) | JetBrains Mono | Inter |
| `martian` | Martian Mono | Inter |
| `plex` | IBM Plex Mono | IBM Plex Sans |
| `geist` | Geist Mono | Geist |

Google Fonts link (all pages):

```html
<link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;600;700&family=Martian+Mono:wght@400;600;700&family=IBM+Plex+Mono:wght@400;600&family=IBM+Plex+Sans:wght@400;500;600&family=Geist+Mono:wght@400;600&family=Geist:wght@400;500;600&family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
```

State persists in localStorage `prodcal-theme-v1` (`{font, dark}`); theme.js
migrates legacy font keys and keeps the random first-visit pairing pick.
Every page's `<head>` keeps the tiny inline bootstrap that applies the dark
class + `data-font` before first paint.

## Shared chrome

Masthead on every surface (`.jdbb-masthead`): wordmark left —
`<a class="jdbb-wordmark"><span class="bracket">[</span>JDBB<span class="bracket">]</span><span class="studio">studio</span></a>`
— nav/actions right, ending with the theme bar (`JdbbTheme.mount(el)`).
Below it a `.jdbb-statusline` (context left, `UTC date · state` right).
Footer: `[JDBB]` left, `Admin · © 2026 JDBB` right, above a closing hairline.

## Content facts (recurring copy)

- Trim range: `5×8 — 7×10 · print + EPUB`
- Pipeline: `docx → pandoc·Lua → Typst → PDF/X + EPUB`
- Cadence: `shared calendar · scheduled digests` (never "weekly digest")
- Contact fallback: `j@djinna.com` (pages fetch `/api/public/config`)

## Wordmark / favicon

Working wordmark is the bracketed mono lockup above. Favicon:
`/static/favicon.svg` (route `/favicon.ico` redirects there). A drawn-mark
pass is in progress; until it lands, keep the existing favicon wiring.
