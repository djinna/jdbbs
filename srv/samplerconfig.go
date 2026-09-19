package srv

import (
	"encoding/json"
	"fmt"
)

// SamplerTemplate writes into dir the specialised series template a real
// build would use for a transmittal carrying the given typography choices
// on the named trim, and returns the template path plus the config lines.
// Used by cmd/typoconfig for the customer typography sampler
// (typesetting/scripts/build-sampler.sh), so the sampler pages and the
// production PDF cannot drift apart: both go through applyTypoChoices →
// specToTypstConfig → writeSpecialisedTemplate.
//
//	pairing       studio | classic | house | literary
//	size          compact | standard | generous
//	sectionBreak  space | breve | ornament   (transmittal keys, not template names)
//	paragraphs    indented | block
func SamplerTemplate(dir, trim, pairing, size, sectionBreak, paragraphs string) (path, config string, err error) {
	config, err = SamplerTypstConfig(trim, pairing, size, sectionBreak, paragraphs)
	if err != nil {
		return "", "", err
	}
	path, err = writeSpecialisedTemplate(dir, config)
	return path, config, err
}

// SamplerTypstConfig returns the Typst `#let config = merge-config((…))`
// block for the default spec with the given trim and typography choices.
func SamplerTypstConfig(trim, pairing, size, sectionBreak, paragraphs string) (string, error) {
	p, ok := trimRegistry[trim]
	if !ok {
		return "", fmt.Errorf("unknown trim %q", trim)
	}
	var data map[string]any
	if err := json.Unmarshal([]byte(defaultSpecData()), &data); err != nil {
		return "", err
	}
	page := ensureMap(data, "page")
	page["trim"] = trim
	page["width_in"] = p.WidthIn
	page["height_in"] = p.HeightIn
	applyTypoChoices(map[string]any{
		"pairing":       pairing,
		"size":          size,
		"section_break": sectionBreak,
		"paragraphs":    paragraphs,
	}, data)
	return specToTypstConfig(data), nil
}
