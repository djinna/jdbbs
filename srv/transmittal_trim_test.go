package srv

import (
	"encoding/json"
	"testing"
)

func TestTrimTransmittalIdentity(t *testing.T) {
	in := `{"book":{"title":"Obliquities 1 ","author":"  A. Writer","isbn_paper":" keep "},"design":{"trim":"6 x 9"}}`
	out := trimTransmittalIdentity(in)
	var d map[string]map[string]any
	if err := json.Unmarshal([]byte(out), &d); err != nil {
		t.Fatal(err)
	}
	if d["book"]["title"] != "Obliquities 1" || d["book"]["author"] != "A. Writer" {
		t.Fatalf("not trimmed: %v", d["book"])
	}
	if d["book"]["isbn_paper"] != " keep " {
		t.Fatalf("non-identity field touched: %q", d["book"]["isbn_paper"])
	}
	if d["design"]["trim"] != "6 x 9" {
		t.Fatalf("other sections lost")
	}
	if got := trimTransmittalIdentity("not json"); got != "not json" {
		t.Fatalf("garbage should pass through, got %q", got)
	}
	clean := `{"book":{"title":"X"}}`
	if got := trimTransmittalIdentity(clean); got != clean {
		t.Fatalf("clean input should be returned byte-for-byte, got %q", got)
	}
}
