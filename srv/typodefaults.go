package srv

import (
	"fmt"
	"strings"
)

// Typographic defaults derived from the trim (2026-09-18, punch list 0.8).
//
// Until now every book got the same margins (0.75/0.75/0.7/0.6in) and the
// same "leading_pt: 2", whatever the trim and typeface. On a 6×9 that meant
// an 85-character measure set 10/8.6 — type tighter than solid. Bringhurst
// (2.1–2.2, 8.2): a measure of 45–75 characters, ~66 ideal; leading of
// roughly 120–135 % of the size for book text; bottom margin deeper than
// the top, inner narrower than the outer (the two inner margins read as
// one across the gutter).
//
// The transmittal exposes no margin or leading fields, so stored values
// are always our old defaults; they are treated as "unset" and replaced
// with trim-derived values. A spec that carries page.margins_custom: true
// or typography.leading_custom: true is left alone (studio-set).

// legacyMargins is the quad every spec was born with before 2026-09-18.
var legacyMargins = map[string]string{
	"margin_top": "0.75in", "margin_bottom": "0.75in",
	"margin_inside": "0.7in", "margin_outside": "0.6in",
}

// capHeightRatio approximates cap-height/size for the faces we set, because
// typst's `leading` is the gap below a line box that spans cap-height to
// baseline; baseline-to-baseline = cap-height + leading.
func capHeightRatio(bodyFont string) float64 {
	switch {
	case strings.Contains(bodyFont, "Plantin"):
		return 0.72
	case strings.Contains(bodyFont, "Libertinus"):
		return 0.66
	case strings.Contains(bodyFont, "Cardo"):
		return 0.64
	case strings.Contains(bodyFont, "EB Garamond"), strings.Contains(bodyFont, "Garamond"):
		return 0.65
	default:
		return 0.68
	}
}

// trimMargins returns margins in inches for a page of w×h inches.
// Proportions chosen so 6×9 → inside 0.875, outside 0.75, top 0.8, bottom 1.0,
// and the Protocolized trim (4.9×7.6) stays close to its measured margins.
func trimMargins(w, h float64) (top, bottom, inside, outside float64) {
	return round2(h * 0.089), round2(h * 0.111), round2(w * 0.146), round2(w * 0.125)
}

func round2(f float64) float64 { return float64(int(f*100+0.5)) / 100 }

// trimTypeSize picks the body size for the text width: ~10pt for the small
// trims, 10.5pt from 5.5in wide (6×9 and up) so the measure stays near 66–70
// characters.
func trimTypeSize(w float64) float64 {
	if w >= 5.5 {
		return 10.5
	}
	return 10
}

// leadingFor returns typst leading (pt) that gives a baseline-to-baseline of
// ratio × size for the face.
func leadingFor(size float64, bodyFont string, ratio float64) float64 {
	return round2(size*ratio - size*capHeightRatio(bodyFont))
}

// applyTypoDefaults rewrites page margins and typography size/leading in the
// spec map (in place) when they carry the legacy defaults.
func applyTypoDefaults(data map[string]any) {
	page, _ := data["page"].(map[string]any)
	if page == nil {
		return
	}
	trim, _ := page["trim"].(string)
	w, _ := page["width_in"].(float64)
	h, _ := page["height_in"].(float64)
	if p, ok := trimRegistry[trim]; ok {
		w, h = p.WidthIn, p.HeightIn
	}
	if w <= 0 || h <= 0 {
		return
	}
	custom, _ := page["margins_custom"].(bool)
	if !custom && marginsAreLegacy(page) {
		t, b, i, o := trimMargins(w, h)
		page["margin_top"] = fmt.Sprintf("%gin", t)
		page["margin_bottom"] = fmt.Sprintf("%gin", b)
		page["margin_inside"] = fmt.Sprintf("%gin", i)
		page["margin_outside"] = fmt.Sprintf("%gin", o)
	}

	typo, _ := data["typography"].(map[string]any)
	if typo == nil {
		return
	}
	if lc, _ := typo["leading_custom"].(bool); lc {
		return
	}
	lead, hasLead := typo["leading_pt"].(float64)
	base, _ := typo["base_size_pt"].(float64)
	// leading_pt ≤ 3 is the old default (2); anything larger was set on purpose.
	if hasLead && lead > 3 {
		return
	}
	if base <= 0 || base == 10 {
		base = trimTypeSize(w)
		typo["base_size_pt"] = base
	}
	body, _ := typo["body_font"].(string)
	typo["leading_pt"] = leadingFor(base, body, 1.28)
	if v, _ := typo["paragraph_indent_em"].(float64); v == 0 || v == 0.75 {
		typo["paragraph_indent_em"] = 1.25
	}
}

func marginsAreLegacy(page map[string]any) bool {
	for k, legacy := range legacyMargins {
		v, _ := page[k].(string)
		if v != "" && v != legacy {
			return false
		}
	}
	return true
}
