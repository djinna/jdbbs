// indexanchor writes the #index(...) markers of an index.json draft into a
// pandoc-produced Typst manuscript, producing book-indexed.typ for the
// series template to compile with `index: true` (punch list 5.13,
// docs/reviews/INDEX-ADDON-DESIGN-2026-09-19.md). Anchors it cannot find are
// listed, never guessed.
//
//	go run ./cmd/indexanchor -in scratch/typo/ghosts.typ -index scratch/idx/index.json -out scratch/idx/ghosts-indexed.typ
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"srv.exe.dev/srv/indexer"
)

func main() {
	in := flag.String("in", "", "pandoc Typst manuscript (required)")
	idxPath := flag.String("index", "index.json", "draft from indexdraft")
	out := flag.String("out", "book-indexed.typ", "where to write the marked-up Typst")
	report := flag.String("report", "", "write the placement report as JSON here (optional)")
	flag.Parse()
	if *in == "" {
		flag.Usage()
		os.Exit(2)
	}
	src, err := os.ReadFile(*in)
	if err != nil {
		fail(err)
	}
	ib, err := os.ReadFile(*idxPath)
	if err != nil {
		fail(err)
	}
	var idx indexer.Index
	if err := json.Unmarshal(ib, &idx); err != nil {
		fail(fmt.Errorf("%s: %w", *idxPath, err))
	}
	marked, rep := indexer.PlaceMarkers(string(src), &idx)
	if err := os.WriteFile(*out, []byte(marked), 0644); err != nil {
		fail(err)
	}
	fmt.Fprintf(os.Stderr, "%s: %d entries, %d anchors: %d placed (%d fuzzy), %d cross-refs, %d unmatched\n",
		*out, rep.Entries, rep.Anchors, rep.Placed, rep.Fuzzy, rep.CrossRefs, len(rep.Unmatched))
	for _, u := range rep.Unmatched {
		fmt.Fprintf(os.Stderr, "  unmatched  ch%d  %s%s: %q\n", u.Chapter, u.Heading, sub(u.Subheading), u.Text)
	}
	if *report != "" {
		b, _ := json.MarshalIndent(rep, "", "  ")
		if err := os.WriteFile(*report, b, 0644); err != nil {
			fail(err)
		}
	}
}

func sub(s string) string {
	if s == "" {
		return ""
	}
	return " — " + s
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "indexanchor:", err)
	os.Exit(1)
}
