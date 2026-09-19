package srv

import (
	"encoding/json"
	"strings"
	"testing"
)

// The four transmittal rows: pairing → fonts, size → pt on 6×9, section
// break → template style, paragraphs → indent/spacing in config.typ.
func TestApplyTypoChoices(t *testing.T) {
	spec := func() map[string]any {
		var d map[string]any
		json.Unmarshal([]byte(defaultSpecData()), &d)
		d["page"].(map[string]any)["trim"] = "6 x 9"
		return d
	}
	t.Run("literary pairing, generous, ornament, block", func(t *testing.T) {
		d := spec()
		applyTypoChoices(map[string]any{
			"pairing": "literary", "size": "generous", "section_break": "ornament", "paragraphs": "block",
		}, d)
		typo := d["typography"].(map[string]any)
		if typo["body_font"] != "EB Garamond" || typo["heading_font"] != "EB Garamond" {
			t.Fatalf("fonts: %v / %v", typo["body_font"], typo["heading_font"])
		}
		if d["elements"].(map[string]any)["section_break"] != "fleuron" {
			t.Fatalf("section_break: %v", d["elements"].(map[string]any)["section_break"])
		}
		cfg := specToTypstConfig(d)
		for _, want := range []string{
			`body-font: "EB Garamond"`, "base-size: 11.25pt,", "leading: 7.09pt,",
			"paragraph-indent: 0em,", "paragraph-spacing: 12.72pt,", `section-break: "fleuron"`,
		} {
			if !strings.Contains(cfg, want) {
				t.Errorf("config missing %q:\n%s", want, cfg)
			}
		}
	})
	t.Run("defaults: studio, standard, space, indented", func(t *testing.T) {
		d := spec()
		applyTypoChoices(map[string]any{}, d)
		cfg := specToTypstConfig(d)
		for _, want := range []string{
			`body-font: "Libertinus Serif"`, `heading-font: "Source Sans 3"`, "base-size: 10.5pt,",
			"leading: 6.51pt,", "paragraph-spacing: 6.51pt,", "paragraph-indent: 1.25em,", `section-break: "blank"`,
		} {
			if !strings.Contains(cfg, want) {
				t.Errorf("config missing %q:\n%s", want, cfg)
			}
		}
	})
	t.Run("compact on 6x9 is 10pt", func(t *testing.T) {
		d := spec()
		applyTypoChoices(map[string]any{"size": "compact", "pairing": "house"}, d)
		cfg := specToTypstConfig(d)
		if !strings.Contains(cfg, "base-size: 10pt,") || !strings.Contains(cfg, `body-font: "Plantin MT Pro"`) {
			t.Errorf("config:\n%s", cfg)
		}
	})
	t.Run("back to studio resets a named pairing", func(t *testing.T) {
		d := spec()
		applyTypoChoices(map[string]any{"pairing": "house"}, d)
		applyTypoChoices(map[string]any{"pairing": "studio"}, d)
		if bf := d["typography"].(map[string]any)["body_font"]; bf != "Libertinus Serif" {
			t.Errorf("body_font after studio: %v", bf)
		}
	})
}
