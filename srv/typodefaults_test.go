package srv

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestApplyTypoDefaults(t *testing.T) {
	var data map[string]any
	if err := json.Unmarshal([]byte(defaultSpecData()), &data); err != nil {
		t.Fatal(err)
	}
	data["page"].(map[string]any)["trim"] = "6 x 9"
	cfg := specToTypstConfig(data)
	for _, want := range []string{
		"margin-inside: 0.88in", "margin-outside: 0.75in", "margin-top: 0.8in", "margin-bottom: 1in",
		"base-size: 10.5pt", "leading: 6.51pt", "paragraph-indent: 1.25em",
	} {
		if !strings.Contains(cfg, want) {
			t.Errorf("missing %q in\n%s", want, cfg)
		}
	}

	// Studio-set values survive.
	var custom map[string]any
	json.Unmarshal([]byte(defaultSpecData()), &custom)
	custom["page"].(map[string]any)["margin_inside"] = "1.1in"
	custom["typography"].(map[string]any)["leading_pt"] = 4.8
	cfg = specToTypstConfig(custom)
	for _, want := range []string{"margin-inside: 1.1in", "margin-outside: 0.6in", "leading: 4.8pt", "base-size: 10pt"} {
		if !strings.Contains(cfg, want) {
			t.Errorf("custom: missing %q in\n%s", want, cfg)
		}
	}
}
