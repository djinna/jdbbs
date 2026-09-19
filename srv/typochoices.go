package srv

import "strings"

// Customer typographic choices on the transmittal (2026-09-19, punch list
// 0.8 part 2). Deliberately few: four radio rows, each stored under the
// transmittal's `typography` key and mirrored into the book spec. Anything
// not offered here stays a studio decision (spec edits in admin).
//
//   typography.pairing        studio | classic | house | literary
//   typography.size           compact | standard | generous
//   typography.section_break  space | breve | ornament
//   typography.paragraphs     indented | block
//
// Unset or unknown values resolve to the first (default) of each row, so
// transmittals saved before these rows existed build exactly as before —
// except that the default section break is now white space, which is what
// a manuscript with no stated preference reads best with.

// typoPairing is one offered body/heading pairing. Family names must match
// what `typst fonts --font-path typesetting/fonts` reports.
type typoPairing struct {
	Key, Name, Body, Heading string
}

// typoPairings are the three offered pairings, in transmittal order. "studio"
// (studio's choice) resolves to the first.
var typoPairings = []typoPairing{
	{"classic", "Open classic", "Libertinus Serif", "Source Sans 3"},
	{"house", "Studio house", "Plantin MT Pro", "Proxima Nova"},
	{"literary", "Literary", "EB Garamond", "EB Garamond"},
}

// resolvePairing returns the pairing for a transmittal key; studio/unknown →
// the first (classic), which is also the spec default.
func resolvePairing(key string) typoPairing {
	for _, p := range typoPairings {
		if p.Key == key {
			return p
		}
	}
	return typoPairings[0]
}

// typoSizeFactor is the multiplier on the trim-derived body size.
func typoSizeFactor(size string) float64 {
	switch size {
	case "compact":
		return 0.95
	case "generous":
		return 1.06
	}
	return 1
}

// roundQuarterPt rounds a point size to the nearest 0.25pt.
func roundQuarterPt(pt float64) float64 { return float64(int(pt*4+0.5)) / 4 }

// typoSectionBreakStyle maps the transmittal's section-break choice onto the
// series template's `section-break` styles (blank / breve / fleuron).
func typoSectionBreakStyle(choice string) string {
	switch choice {
	case "breve":
		return "breve"
	case "ornament":
		return "fleuron"
	}
	return "blank"
}

// typoParagraphsBlock reports whether the transmittal asked for block
// paragraphs (no first-line indent, space between).
func typoParagraphsBlock(choice string) bool { return strings.EqualFold(choice, "block") }

// applyTypoChoices mirrors the transmittal's `typography` map into the spec
// (in place): the resolved pairing's fonts, and the raw choice keys so the
// build (applyTypoDefaults / specToTypstConfig) and the EPUB can honour them.
func applyTypoChoices(txTypo map[string]any, specData map[string]any) {
	typo := ensureMap(specData, "typography")
	get := func(k string) string {
		v, _ := txTypo[k].(string)
		return strings.TrimSpace(v)
	}
	pairing := get("pairing")
	prev, _ := typo["pairing"].(string)
	typo["pairing"] = pairing
	bf, _ := typo["body_font"].(string)
	switch {
	case pairing != "" && pairing != "studio":
		p := resolvePairing(pairing)
		typo["body_font"] = p.Body
		typo["heading_font"] = p.Heading
	case bf == "" || (prev != "" && prev != "studio"):
		// Studio's choice: the default pairing — also when the customer has
		// just switched back from a named pairing. Otherwise the fonts the
		// studio set in the spec stand.
		typo["body_font"] = typoPairings[0].Body
		typo["heading_font"] = typoPairings[0].Heading
	}
	typo["size"] = get("size")
	typo["paragraphs"] = get("paragraphs")
	sb := typoSectionBreakStyle(get("section_break"))
	ensureMap(specData, "elements")["section_break"] = sb
	ensureMap(specData, "epub")["section_break"] = sb
}
