package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// twoChapters is a manuscript small enough to read, with a synonym pair
// ("the pass" / "Factory Pass") and dialogue quotes for the JSON repair path.
const twoChapters = `#import "/templates/series-template.typ": *

#show: book.with(
  title: "TITLE",
  author: "AUTHOR",
)

= The Pass
<the-pass>

#first-para[
The pass is a checklist the studio runs before every build. It began as a
paper form and became a page. "Run the pass," she said, "before you build."
The pass has nine steps and takes an hour when nothing is wrong.
]

= The Factory
<the-factory>

#block[
The Factory Pass gates the build. The transmittal is the specification the
factory reads; versions of the transmittal are kept for every build so a
rebuild any time gives the same book. Typesetting is done in Typst.
]
`

// canned writes a fake gateway answer for the exact prompt Draft will send.
func cannedAnswer(t *testing.T, dir, model, system, user, text string) {
	t.Helper()
	b, _ := json.Marshal(canned{Model: model, System: system, User: user, Text: text,
		Usage: Usage{Calls: 1, InputTokens: 100, OutputTokens: 50}})
	if err := os.WriteFile(filepath.Join(dir, promptKey(model, system, user)+".json"), b, 0644); err != nil {
		t.Fatal(err)
	}
}

func TestDraftReplay(t *testing.T) {
	dir := t.TempDir()
	c := &Client{Model: "claude-sonnet-4-5", Replay: dir}
	opt := Options{Book: "Test Book", Pages: 10}
	chunks := Chunks(twoChapters, false)
	if len(chunks) != 2 {
		t.Fatalf("chunks: %d", len(chunks))
	}
	idx0 := &Index{Budget: BudgetFor(10)}
	words := chunks[0].Words + chunks[1].Words
	// Chapter replies (the second with an unescaped quote, pretty-printed and fenced).
	cannedAnswer(t, dir, c.Model, SystemPrompt, chapterPrompt(opt, chunks[0], shareFor(idx0.Budget, chunks[0].Words, words)),
		`{"entries":[{"heading":"the pass","subheading":"","see":"","see_also":[],"anchors":["The pass is a checklist the studio runs","The pass has nine steps"]},{"heading":"checklist","subheading":"","see":"the pass","see_also":[],"anchors":[]}]}`)
	cannedAnswer(t, dir, c.Model, SystemPrompt, chapterPrompt(opt, chunks[1], shareFor(idx0.Budget, chunks[1].Words, words)),
		"```json\n{\n \"entries\": [\n  {\"heading\": \"Factory Pass\", \"subheading\": \"\", \"see\": \"\", \"see_also\": [], \"anchors\": [\"The Factory Pass gates the build\"]},\n"+
			`  {"heading": "transmittal", "subheading": "as specification", "see": "", "see_also": [], "anchors": ["The transmittal is the specification"]},`+
			`  {"heading": "transmittal", "subheading": "versions of", "see": "", "see_also": [], "anchors": ["versions of the "transmittal" are kept"]},`+
			`  {"heading": "typesetting", "subheading": "", "see": "Typst", "see_also": [], "anchors": []},`+
			`  {"heading": "Typst", "subheading": "", "see": "", "see_also": [], "anchors": ["Typesetting is done in Typst"]}`+
			"\n ]\n}\n```")
	// The merge call: its prompt is built from the merged draft, so build it
	// the same way Draft does by running once without merge.
	idx1, err := Draft(context.Background(), c, twoChapters, Options{Book: "Test Book", Pages: 10, NoMerge: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(idx1.Entries) != 7 {
		t.Fatalf("pre-merge entries: %d", len(idx1.Entries))
	}
	mergeUser := mergePrompt(idx1)
	cannedAnswer(t, dir, c.Model, MergeSystemPrompt, mergeUser,
		`{"merge":[{"from":"the pass","to":"Factory Pass"}],"see":[],"see_also":[{"heading":"Factory Pass","also":"transmittal"}],"drop":[]}`)

	idx, err := Draft(context.Background(), c, twoChapters, Options{Book: "Test Book", Pages: 10})
	if err != nil {
		t.Fatal(err)
	}
	byKey := map[string]Entry{}
	for _, e := range idx.Entries {
		byKey[e.Heading+"|"+e.Subheading] = e
	}
	fp, ok := byKey["Factory Pass|"]
	if !ok || len(fp.Anchors) != 3 {
		t.Errorf("Factory Pass should carry the pass's anchors: %+v", fp)
	}
	if len(fp.SeeAlso) != 1 || fp.SeeAlso[0] != "transmittal" {
		t.Errorf("see also: %+v", fp.SeeAlso)
	}
	if e, ok := byKey["checklist|"]; !ok || e.See != "Factory Pass" {
		t.Errorf("see should follow the rename: %+v", e)
	}
	if e, ok := byKey["the pass|"]; ok {
		t.Errorf("'the pass' shares a word with 'Factory Pass', so no see entry expected: %+v", e)
	}
	if e := byKey["transmittal|versions of"]; len(e.Anchors) != 1 || !strings.Contains(e.Anchors[0].Text, `"transmittal"`) {
		t.Errorf("quote repair lost the anchor: %+v", e)
	}
	if idx.Usage.Calls != 3 || idx.CostUSD <= 0 {
		t.Errorf("usage: %+v cost %v", idx.Usage, idx.CostUSD)
	}
	for _, e := range idx.Entries {
		if e.Anchors == nil {
			t.Errorf("nil anchors on %q", e.Heading)
		}
		if e.See != "" && len(e.Anchors) > 0 {
			t.Errorf("see on an entry with locators: %+v", e)
		}
	}
}

func TestShareWord(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"the pass", "Factory Pass", true},
		{"beaings", "artificial ghosts", false},
		{"the store", "the house", false},
		{"AI systems", "artificial intelligence", false},
		{"afterlife, digital", "digital afterlives", true},
	}
	for _, c := range cases {
		if got := shareWord(c.a, c.b); got != c.want {
			t.Errorf("shareWord(%q, %q) = %v", c.a, c.b, got)
		}
	}
}

func TestDraftReplayMissing(t *testing.T) {
	c := &Client{Model: "m", Replay: t.TempDir()}
	if _, err := Draft(context.Background(), c, twoChapters, Options{Book: "x", Pages: 10}); err == nil {
		t.Fatal("expected replay miss error")
	}
}

func TestTidyCrossRefs(t *testing.T) {
	seq := 0
	a := func(n int) []Anchor {
		var as []Anchor
		for i := 0; i < n; i++ {
			seq++
			as = append(as, Anchor{Chapter: 1, Text: fmt.Sprintf("anchor %d", seq)})
		}
		return as
	}
	var notes []string
	es := Tidy([]Entry{
		// reciprocal pair, one side thin → folded into the other, see left behind
		{Heading: "memory optimization", SeeAlso: []string{"emotions, compression of"}, Anchors: a(2)},
		{Heading: "emotions, compression of", SeeAlso: []string{"memory optimization"}, Anchors: a(1)},
		// reciprocal pair, both substantial → kept (legitimate Chicago)
		{Heading: "cognitive science", SeeAlso: []string{"neuroscience"}, Anchors: a(2)},
		{Heading: "neuroscience", SeeAlso: []string{"cognitive science"}, Anchors: a(2)},
		// reciprocal pair, thin side is a proper name → kept
		{Heading: "Ting Ting", SeeAlso: []string{"artificial intelligence"}, Anchors: a(1)},
		{Heading: "artificial intelligence", SeeAlso: []string{"Ting Ting"}, Anchors: a(3)},
		// see-also to "heading, subheading" → heading; to nothing → dropped; to self → dropped
		{Heading: "loneliness", SeeAlso: []string{"Beeman, Darius, loneliness and isolation", "smart speakers", "Loneliness"}, Anchors: a(1)},
		{Heading: "Beeman, Darius", Subheading: "loneliness and isolation", Anchors: a(1)},
		// see to a missing heading → dropped
		{Heading: "klon", See: "khlong"},
	}, func(n string) { notes = append(notes, n) })
	get := func(h string) *Entry {
		for i := range es {
			if es[i].Heading == h && es[i].Subheading == "" {
				return &es[i]
			}
		}
		return nil
	}
	if e := get("memory optimization"); e == nil || len(e.Anchors) != 3 || len(e.SeeAlso) != 0 {
		t.Errorf("memory optimization should have absorbed the thin side: %+v", e)
	}
	if e := get("emotions, compression of"); e == nil || e.See != "memory optimization" || len(e.Anchors) != 0 {
		t.Errorf("thin side should be a see entry: %+v", e)
	}
	if e := get("cognitive science"); e == nil || len(e.SeeAlso) != 1 || e.SeeAlso[0] != "neuroscience" {
		t.Errorf("substantial reciprocal pair should be kept: %+v", e)
	}
	if e := get("Ting Ting"); e == nil || e.See != "" || len(e.SeeAlso) != 1 {
		t.Errorf("proper name must not be folded: %+v", e)
	}
	if e := get("loneliness"); e == nil || len(e.SeeAlso) != 1 || e.SeeAlso[0] != "Beeman, Darius" {
		t.Errorf("see-also targets: %+v", e)
	}
	if e := get("klon"); e == nil || e.See != "" {
		t.Errorf("dangling see should be dropped: %+v", e)
	}
	if len(notes) != 3 {
		t.Errorf("notes: %q", notes)
	}
}
