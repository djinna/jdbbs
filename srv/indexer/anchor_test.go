package indexer

import (
	"strings"
	"testing"
)

func TestPlaceMarkers(t *testing.T) {
	idx := &Index{Entries: []Entry{
		{Heading: "the pass", Anchors: []Anchor{{Chapter: 1, Text: "The pass is a checklist the studio runs"}}},
		// Smart quotes + markup in the source, straight quotes in the anchor; ends before punctuation.
		{Heading: "building", Subheading: "before the pass", SeeAlso: []string{"the pass", "Factory Pass"},
			Anchors: []Anchor{{Chapter: 1, Text: `"Run the pass," she said, "before you build`}}},
		// Chapter 2, straddling inline markup and a line break.
		{Heading: "transmittal", Subheading: "as specification", Anchors: []Anchor{{Chapter: 2, Text: "The transmittal is the specification the factory reads"}}},
		// Fuzzy: model paraphrased the tail.
		{Heading: "rebuilds", Anchors: []Anchor{{Chapter: 2, Text: "a rebuild any time gives the same output every time"}}},
		// Unmatched: not in the text at all.
		{Heading: "ghosts", Anchors: []Anchor{{Chapter: 2, Text: "there are no ghosts in this manuscript at all"}}},
		// Pure cross-reference.
		{Heading: "checklist", See: "the pass", Anchors: []Anchor{}},
		// Heading with a quote that needs escaping.
		{Heading: `Typst ("the typesetter")`, Anchors: []Anchor{{Chapter: 2, Text: "Typesetting is done in Typst"}}},
	}}
	out, rep := PlaceMarkers(twoChapters, idx)
	if rep.Placed != 5 || rep.Fuzzy != 1 || len(rep.Unmatched) != 1 || rep.CrossRefs != 1 {
		t.Fatalf("report: placed %d fuzzy %d unmatched %d xrefs %d", rep.Placed, rep.Fuzzy, len(rep.Unmatched), rep.CrossRefs)
	}
	if rep.Unmatched[0].Heading != "ghosts" {
		t.Errorf("unmatched: %+v", rep.Unmatched)
	}
	wants := []string{
		// after the word and its trailing space, glued to the next word
		`the studio runs #index("the pass");before every build`,
		"before you build.\"\n" + `#index("building", "before the pass", see-also: "the pass");#index("building", see-also: "Factory Pass", locator: false);The pass has`,
		`factory reads; #index("transmittal", "as specification");versions`,
		// fuzzy match ends at "same"; the last marker stays before the block's closing newline
		`gives the same #index("rebuilds");book. Typesetting is done in Typst.#index("Typst (\"the typesetter\")");` + "\n]",
		"\n" + `#index("checklist", see: "the pass", locator: false);` + "\n",
	}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Errorf("missing %q in:\n%s", w, out)
		}
	}
	// Nothing visible was added: the extracted text is unchanged.
	if got, want := strings.TrimSpace(ExtractText(out).String()), strings.TrimSpace(ExtractText(twoChapters).String()); got != want {
		t.Errorf("visible text changed:\n%s", got)
	}
}

func TestFindAnchorShortening(t *testing.T) {
	s := Fold("Rice was the region's dietary staple, key to food security, and historical enabler of abundance.")
	end, m, fuzzy := findAnchor(s, Fold("dietary staple key to food security"), 0, len(s))
	if end < 0 || fuzzy || m != "dietary staple key to food security" {
		t.Errorf("exact: %d %q %v", end, m, fuzzy)
	}
	end, m, fuzzy = findAnchor(s, Fold("key to food security and the wrong tail here"), 0, len(s))
	if end < 0 || !fuzzy || m != "key to food security and" {
		t.Errorf("trim end: %d %q %v", end, m, fuzzy)
	}
	end, m, fuzzy = findAnchor(s, Fold("wrong head words then historical enabler of abundance"), 0, len(s))
	if end < 0 || !fuzzy || m != "historical enabler of abundance" {
		t.Errorf("trim start: %d %q %v", end, m, fuzzy)
	}
	if end, _, _ = findAnchor(s, Fold("nothing like this"), 0, len(s)); end >= 0 {
		t.Errorf("should not match")
	}
	if end, _, _ = findAnchor(s, Fold("of abundance"), 0, len(s)); end < 0 {
		t.Errorf("short exact fragment should still match")
	}
}
