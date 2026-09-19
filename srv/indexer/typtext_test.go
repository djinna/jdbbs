package indexer

import (
	"encoding/json"
	"strings"
	"testing"
)

const sample = `#import "/templates/series-template.typ": *

#show: book.with(
  title: "TITLE",
  author: "AUTHOR",
)

= Khlongs, Subaks, Beaings
<khlongs-subaks-beaings>

#first-para[
Rice was the region's #emph[dietary staple], key to food security.
]
#block[
See #link("https://southbeast.asia")[#underline[southbeast.asia]] for
concept art. \[ALERT: mismatch\] -- said the *machine*.

]
#section-break
= Soda Sweet as Blood
<soda>

#block[
Second chapter text, long enough to count as a chunk when repeated.
]
`

func TestExtractText(t *testing.T) {
	got := ExtractText(sample).String()
	for _, want := range []string{
		"Khlongs, Subaks, Beaings",
		"Rice was the region's dietary staple, key to food security.",
		"See southbeast.asia for\nconcept art. [ALERT: mismatch] -- said the machine.",
		"Soda Sweet as Blood",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	for _, bad := range []string{"#", "series-template", "book.with", "TITLE", "<khlongs", "https://", "first-para", "]"} {
		if strings.Contains(got, bad) && bad != "]" {
			t.Errorf("markup leaked: %q in:\n%s", bad, got)
		}
	}
	if strings.Count(got, "]") != 1 { // only the escaped one
		t.Errorf("brackets leaked:\n%s", got)
	}
}

func TestExtractOffsets(t *testing.T) {
	tx := ExtractText(sample)
	for _, r := range tx.Runes {
		if r.Off < 0 || r.Off >= len(sample) {
			t.Fatalf("offset out of range: %+v", r)
		}
		if r.R < 0x80 && sample[r.Off] != byte(r.R) && sample[r.Off] != '\\' {
			t.Fatalf("offset %d points at %q, rune %q", r.Off, sample[r.Off], r.R)
		}
	}
}

func TestFold(t *testing.T) {
	cases := map[string]string{
		"Rice was the region’s “dietary staple”": "rice was the region s dietary staple",
		"  Éminence   grise—yes ":                "eminence grise yes",
		"the pass / Factory Pass":                "the pass factory pass",
		"":                                       "",
	}
	for in, want := range cases {
		if got := Fold(in); got != want {
			t.Errorf("Fold(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFoldTextOffsets(t *testing.T) {
	tx := ExtractText(sample)
	ft := FoldText(tx)
	if len(ft.From) != len(ft.S)+1 {
		t.Fatalf("From has %d entries for %d bytes", len(ft.From), len(ft.S))
	}
	i := strings.Index(ft.S, "dietary staple")
	if i < 0 {
		t.Fatalf("folded text lacks phrase: %q", ft.S)
	}
	end := ft.From[i+len("dietary staple")-1]
	off := tx.Runes[end].Off
	if sample[off] != 'e' || !strings.HasPrefix(sample[off-13:], "dietary staple") {
		t.Errorf("end maps to %q", sample[off-13:off+1])
	}
}

func TestChunks(t *testing.T) {
	// Pad chapter 2 so it clears the 20-word floor.
	src := sample + strings.Repeat("#block[\nmore words here to make the chapter long enough to count.\n]\n", 3)
	ch := Chunks(src, false)
	if len(ch) != 2 {
		t.Fatalf("want 2 chunks, got %d: %+v", len(ch), ch)
	}
	if ch[0].Title != "Khlongs, Subaks, Beaings" || ch[1].Title != "Soda Sweet as Blood" {
		t.Errorf("titles: %q %q", ch[0].Title, ch[1].Title)
	}
	if strings.Contains(ch[0].Text, "Soda") || !strings.Contains(ch[1].Text, "Second chapter") {
		t.Errorf("bad split: %q / %q", ch[0].Text, ch[1].Text)
	}
}

func TestBudgetFor(t *testing.T) {
	if b := BudgetFor(100); b != 320 {
		t.Errorf("BudgetFor(100) = %d", b)
	}
	if b := BudgetFor(250); b != 800 {
		t.Errorf("BudgetFor(250) = %d", b)
	}
}

func TestRepairJSON(t *testing.T) {
	bad := `{"entries":[{"heading":"Ananda (parrot)","anchors":["The parrot squawked. "Attachment is the root of suffering."","he said "yes" to her"]}]}`
	var v map[string]any
	if err := json.Unmarshal([]byte(RepairJSON(bad)), &v); err != nil {
		t.Fatalf("repair failed: %v\n%s", err, RepairJSON(bad))
	}
	good := `{"a":"x, y","b":["q: r"],"c":"d\"e"}`
	if RepairJSON(good) != good {
		t.Errorf("valid JSON changed: %s", RepairJSON(good))
	}
}
