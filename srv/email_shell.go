package srv

import (
	"fmt"
	"html"
	"strings"
)

// ─── Email shell: the studio's text-first look, in mail-safe HTML ───
//
// Every outbound HTML email wraps its body in emailShell. The look mirrors
// docs/PAGE-DESIGN-HOSTING-VISIBILITY-2026-09-11.md translated to what mail
// clients tolerate: table layout, inline styles, no gradients, no emoji, no
// web fonts, no dark-mode tricks. Light palette only (theme.css light tokens).
//
//	[jdbb] studio                       ← wordmark line, hairline below
//	KICKER · CONTEXT                    ← optional mono kicker
//	Title                               ← optional h1
//	…body…
//	— Jenna · jdbb studio               ← sign-off (bodies include their own)
//	hairline · footer note              ← optional small print
//
// Text parts remain primary and are written by each template; the shell is
// only the HTML wrap.

const (
	emailBg        = "#FCFDFD"
	emailText      = "#0E1116"
	emailSecondary = "#5D6B76"
	emailMuted     = "#8A97A1"
	emailBorder    = "#DCE3E8"
	emailRule      = "#0E1116"
	emailAccent    = "#007699"
	emailGreen     = "#2F7D4F"
	emailRed       = "#B3261E"
	emailYellow    = "#8A6A00"
	emailSans      = "-apple-system,'Segoe UI',Helvetica,Arial,sans-serif"
	emailMono      = "ui-monospace,Menlo,Consolas,'Liberation Mono',monospace"
	emailMeasure   = 560 // px; matches the prose measure
	emailBrandName = "jdbb studio"
	emailSignature = "Jenna Dixon · jdbb studio"
	emailStudioURL = "https://jdbbs.exe.xyz/"
)

// emailShellOpts controls the optional pieces around the body.
type emailShellOpts struct {
	Kicker string // mono uppercase context line, e.g. "TRANSMITTAL · FINAL" (already uppercase-safe; escaped)
	Title  string // h1 (escaped)
	Footer string // small print under the closing hairline (raw HTML, caller escapes)
}

// emailShell wraps bodyHTML (trusted, already-escaped HTML) in the studio chrome.
func emailShell(bodyHTML string, o emailShellOpts) string {
	var b strings.Builder
	b.WriteString(`<!DOCTYPE html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><meta name="color-scheme" content="light"><title>`)
	if o.Title != "" {
		b.WriteString(html.EscapeString(o.Title) + " · ")
	}
	b.WriteString(emailBrandName + `</title></head>`)
	fmt.Fprintf(&b, `<body style="margin:0;padding:0;background:%s;color:%s;font:15px/1.6 %s;-webkit-text-size-adjust:100%%">`, emailBg, emailText, emailSans)
	fmt.Fprintf(&b, `<table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="background:%s"><tr><td align="left" style="padding:28px 20px">`, emailBg)
	fmt.Fprintf(&b, `<table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="max-width:%dpx">`, emailMeasure)

	// Wordmark line
	fmt.Fprintf(&b, `<tr><td style="padding:0 0 10px;border-bottom:1px solid %s;font:600 15px/1 %s;letter-spacing:-.01em;color:%s"><a href="%s" style="color:%s;text-decoration:none"><span style="color:%s;font-weight:700">[</span>jdbb<span style="color:%s;font-weight:700">]</span>&nbsp;<span style="font-weight:400;color:%s">studio</span></a></td></tr>`,
		emailRule, emailSans, emailText, emailStudioURL, emailText, emailAccent, emailAccent, emailSecondary)

	if o.Kicker != "" || o.Title != "" {
		b.WriteString(`<tr><td style="padding:22px 0 0">`)
		if o.Kicker != "" {
			fmt.Fprintf(&b, `<div style="font:11px/1.4 %s;letter-spacing:.08em;text-transform:uppercase;color:%s">%s</div>`, emailMono, emailMuted, html.EscapeString(o.Kicker))
		}
		if o.Title != "" {
			fmt.Fprintf(&b, `<h1 style="margin:%s 0 0;font:600 22px/1.25 %s;letter-spacing:-.01em;color:%s">%s</h1>`, ternary(o.Kicker != "", "6px", "0"), emailSans, emailText, html.EscapeString(o.Title))
		}
		b.WriteString(`</td></tr>`)
	}

	// Body
	fmt.Fprintf(&b, `<tr><td style="padding:%s 0 8px;font:15px/1.6 %s;color:%s">`, ternary(o.Kicker != "" || o.Title != "", "18px", "22px"), emailSans, emailText)
	b.WriteString(bodyHTML)
	b.WriteString(`</td></tr>`)

	// Footer hairline + small print
	fmt.Fprintf(&b, `<tr><td style="padding:18px 0 0;border-top:1px solid %s;font:12px/1.5 %s;color:%s">`, emailBorder, emailSans, emailMuted)
	if o.Footer != "" {
		b.WriteString(o.Footer)
	} else {
		fmt.Fprintf(&b, `<a href="%s" style="color:%s;text-decoration:none">%s</a> &middot; Reply to this email to reach Jenna.`, emailStudioURL, emailMuted, emailStudioURL[len("https://"):len(emailStudioURL)-1])
	}
	b.WriteString(`</td></tr></table></td></tr></table></body></html>`)
	return b.String()
}

func ternary(c bool, a, b string) string {
	if c {
		return a
	}
	return b
}

// ─── Body building blocks (inline-styled, mail-safe) ───

func emailP(inner string) string {
	return fmt.Sprintf(`<p style="margin:0 0 14px">%s</p>`, inner)
}

// emailSmall is secondary running copy (13px, secondary color).
func emailSmall(inner string) string {
	return fmt.Sprintf(`<p style="margin:0 0 14px;font-size:13px;color:%s">%s</p>`, emailSecondary, inner)
}

// emailH2 is a section label: mono uppercase kicker with a hairline above.
func emailH2(text string) string {
	return fmt.Sprintf(`<div style="margin:22px 0 10px;padding-top:12px;border-top:1px solid %s;font:11px/1.4 %s;letter-spacing:.08em;text-transform:uppercase;color:%s">%s</div>`, emailBorder, emailMono, emailMuted, html.EscapeString(text))
}

// emailKV renders label/value rows as a ledger (mono labels, hairlines).
// Values are trusted HTML; labels are escaped.
func emailKV(rows [][2]string) string {
	var b strings.Builder
	b.WriteString(`<table role="presentation" cellpadding="0" cellspacing="0" style="border-collapse:collapse;width:100%;margin:0 0 14px">`)
	for i, r := range rows {
		top := emailBorder
		if i == 0 {
			top = emailRule
		}
		fmt.Fprintf(&b, `<tr><td style="padding:7px 14px 7px 0;border-top:1px solid %s;font:11px/1.5 %s;letter-spacing:.06em;text-transform:uppercase;color:%s;white-space:nowrap;vertical-align:top">%s</td><td style="padding:7px 0;border-top:1px solid %s;font:14px/1.5 %s;color:%s;vertical-align:top">%s</td></tr>`,
			top, emailMono, emailMuted, html.EscapeString(r[0]), top, emailSans, emailText, r[1])
	}
	b.WriteString(`</table>`)
	return b.String()
}

// emailTable renders a ledger with a header row. Cells are trusted HTML.
// align is per-column "l"/"r"; nil = all left.
func emailTable(head []string, rows [][]string, align []string) string {
	al := func(i int) string {
		if i < len(align) && align[i] == "r" {
			return "right"
		}
		return "left"
	}
	var b strings.Builder
	b.WriteString(`<table role="presentation" cellpadding="0" cellspacing="0" style="border-collapse:collapse;width:100%;margin:0 0 14px">`)
	b.WriteString(`<tr>`)
	for i, h := range head {
		fmt.Fprintf(&b, `<th align="%s" style="padding:0 10px 7px 0;border-bottom:1px solid %s;font:11px/1.4 %s;letter-spacing:.06em;text-transform:uppercase;color:%s;text-align:%s">%s</th>`, al(i), emailRule, emailMono, emailMuted, al(i), html.EscapeString(h))
	}
	b.WriteString(`</tr>`)
	for _, r := range rows {
		b.WriteString(`<tr>`)
		for i, c := range r {
			fmt.Fprintf(&b, `<td align="%s" style="padding:7px 10px 7px 0;border-bottom:1px solid %s;font:13.5px/1.5 %s;color:%s;text-align:%s;vertical-align:top">%s</td>`, al(i), emailBorder, emailSans, emailText, al(i), c)
		}
		b.WriteString(`</tr>`)
	}
	b.WriteString(`</table>`)
	return b.String()
}

// emailStat is one big-number tile, laid inline (use several in a row via emailStats).
func emailStats(items [][2]string) string {
	var b strings.Builder
	b.WriteString(`<table role="presentation" cellpadding="0" cellspacing="0" style="border-collapse:collapse;margin:0 0 14px"><tr>`)
	for _, it := range items {
		fmt.Fprintf(&b, `<td style="padding:0 28px 0 0;vertical-align:top"><div style="font:600 24px/1.1 %s;letter-spacing:-.02em;color:%s">%s</div><div style="font:11px/1.6 %s;letter-spacing:.06em;text-transform:uppercase;color:%s">%s</div></td>`, emailSans, emailText, it[1], emailMono, emailMuted, html.EscapeString(it[0]))
	}
	b.WriteString(`</tr></table>`)
	return b.String()
}

// emailLink is an accent-colored inline link (escaped href + text).
func emailLink(href, text string) string {
	return fmt.Sprintf(`<a href="%s" style="color:%s;text-decoration:underline">%s</a>`, html.EscapeString(href), emailAccent, html.EscapeString(text))
}

// emailButton is the one "call to action": a bordered text link, not a pill.
func emailButton(href, text string) string {
	return fmt.Sprintf(`<p style="margin:18px 0"><a href="%s" style="display:inline-block;padding:9px 16px;border:1px solid %s;color:%s;font:600 13px/1 %s;letter-spacing:.02em;text-decoration:none">%s &rarr;</a></p>`, html.EscapeString(href), emailRule, emailText, emailSans, html.EscapeString(text))
}

// emailCode is inline mono for codes, slugs, passwords, URLs-as-text.
func emailCode(text string) string {
	return fmt.Sprintf(`<span style="font:13px/1.4 %s;background:#F1F4F6;padding:1px 6px;color:%s">%s</span>`, emailMono, emailText, html.EscapeString(text))
}

// emailStatus colors a status word: done/final/green, active/yellow, blocked/red.
func emailStatus(text string) string {
	c := emailSecondary
	switch strings.ToLower(text) {
	case "done", "final", "complete", "completed", "ok", "pass", "passed", "ready":
		c = emailGreen
	case "active", "in progress", "in-progress", "draft", "working", "pending", "warn", "warning":
		c = emailYellow
	case "blocked", "failed", "fail", "error", "overdue", "late":
		c = emailRed
	}
	return fmt.Sprintf(`<span style="color:%s;font:12px/1.4 %s;letter-spacing:.04em;text-transform:uppercase">%s</span>`, c, emailMono, html.EscapeString(text))
}

// emailSignoff is the standard closing.
func emailSignoff() string {
	return fmt.Sprintf(`<p style="margin:22px 0 0">&mdash; Jenna<br><span style="color:%s">[jdbb] studio</span></p>`, emailSecondary)
}

// emailList renders <ul> with studio spacing. Items are trusted HTML.
func emailList(items []string, ordered bool) string {
	tag := "ul"
	if ordered {
		tag = "ol"
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<%s style="margin:0 0 14px;padding-left:22px">`, tag)
	for _, it := range items {
		fmt.Fprintf(&b, `<li style="margin:0 0 5px">%s</li>`, it)
	}
	fmt.Fprintf(&b, `</%s>`, tag)
	return b.String()
}
