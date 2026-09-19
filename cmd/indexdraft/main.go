// indexdraft asks a generative model for a back-of-book index draft of a
// pandoc-produced Typst manuscript and writes index.json (punch list 5.13,
// docs/reviews/INDEX-ADDON-DESIGN-2026-09-19.md). It never touches the
// database or the metered API.
//
//	go run ./cmd/indexdraft -in scratch/typo/ghosts.typ -book "Ghosts in Machines" -out scratch/idx/index.json
//
// Offline: -replay DIR answers every prompt from canned JSON recorded by an
// earlier -record DIR run (tests use srv/indexer/testdata).
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"srv.exe.dev/srv/indexer"
)

func main() {
	in := flag.String("in", "", "pandoc Typst manuscript (required)")
	out := flag.String("out", "index.json", "where to write the draft")
	book := flag.String("book", "", "book title for the prompt")
	about := flag.String("about", "", "one line on the book's kind/audience (optional)")
	pages := flag.Int("pages", 0, "page count of the set book (0 = estimate from words)")
	parts := flag.Bool("parts", false, "level-1 headings are parts; chapters are level 2")
	model := flag.String("model", "", "gateway model (default $INDEXER_LLM_MODEL or "+indexer.DefaultModel+")")
	replay := flag.String("replay", "", "dry run: answer prompts from this directory of canned responses")
	record := flag.String("record", "", "save live responses here for later -replay")
	noMerge := flag.Bool("no-merge", false, "skip the cross-chapter consolidation call")
	chunksOnly := flag.Bool("chunks", false, "print the chunk table and exit (no model calls)")
	flag.Parse()
	if *in == "" {
		flag.Usage()
		os.Exit(2)
	}
	src, err := os.ReadFile(*in)
	if err != nil {
		fail(err)
	}
	if *chunksOnly {
		for _, ch := range indexer.Chunks(string(src), *parts) {
			fmt.Printf("%2d  %6d words  %s\n", ch.Index, ch.Words, ch.Title)
		}
		return
	}
	c := indexer.NewClientFromEnv()
	if *model != "" {
		c.Model = *model
	}
	if *replay != "" {
		c.Replay = *replay
	}
	if *record != "" {
		c.Record = *record
	}
	logf := func(f string, a ...any) { fmt.Fprintf(os.Stderr, f+"\n", a...) }
	start := time.Now()
	idx, err := indexer.Draft(context.Background(), c, string(src), indexer.Options{
		Book: *book, About: *about, Pages: *pages, Parts: *parts, Log: logf, NoMerge: *noMerge,
	})
	if idx != nil {
		b, _ := json.MarshalIndent(idx, "", "  ")
		if werr := os.WriteFile(*out, b, 0644); werr != nil {
			fail(werr)
		}
	}
	if err != nil {
		fail(err)
	}
	fmt.Fprintf(os.Stderr, "%s: %d entries, %d calls, %d in / %d out tokens, ≈ $%.3f, %s\n",
		*out, len(idx.Entries), idx.Usage.Calls, idx.Usage.InputTokens, idx.Usage.OutputTokens, idx.CostUSD, time.Since(start).Round(time.Second))
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "indexdraft:", err)
	os.Exit(1)
}
