package indexer

import (
	"fmt"
	"sort"
	"strings"
)

// Placement is one marker written into the source.
type Placement struct {
	Entry   int    `json:"entry"` // index into Index.Entries
	Anchor  int    `json:"anchor"`
	Offset  int    `json:"offset"`          // byte offset in the original source
	Fuzzy   bool   `json:"fuzzy,omitempty"` // matched on a trimmed fragment
	Matched string `json:"matched,omitempty"`
}

// Unmatched is an anchor that could not be found; reported, never guessed.
type Unmatched struct {
	Heading    string `json:"heading"`
	Subheading string `json:"subheading,omitempty"`
	Chapter    int    `json:"chapter"`
	Text       string `json:"text"`
}

// AnchorReport summarises a PlaceMarkers run.
type AnchorReport struct {
	Entries    int         `json:"entries"`
	Anchors    int         `json:"anchors"`
	Placed     int         `json:"placed"`
	Fuzzy      int         `json:"fuzzy"`
	CrossRefs  int         `json:"cross_refs"` // see-only entries, placed without a locator
	Unmatched  []Unmatched `json:"unmatched"`
	Placements []Placement `json:"placements,omitempty"`
}

// minAnchorWords is the shortest trimmed fragment we accept as a fuzzy match.
const minAnchorWords = 4

// PlaceMarkers writes #index(...) markers into src at each entry's anchors and
// returns the new source plus a report. Matching is on folded text (case,
// diacritics, quotes, dashes and inline markup ignored), narrowed to the
// anchor's chapter first, then the whole book; a fragment that fails is
// shortened word by word from the end, then from the start, down to
// minAnchorWords. Anything still unmatched is reported and skipped. Entries
// with no anchors (pure see-references) get a locator-less marker at the end
// of the manuscript.
func PlaceMarkers(src string, idx *Index) (string, AnchorReport) {
	tx := ExtractText(src)
	ft := FoldText(tx)
	rep := AnchorReport{Entries: len(idx.Entries)}

	// Chapter ranges in folded bytes, from the level-1 (or level-2 for parts)
	// headings. Chapter k (1-based, matching Chunks) = heads[k-1] .. heads[k].
	level := 1
	for _, h := range tx.Headings {
		if h.Level == 1 {
			level = 1
			break
		}
		level = 2
	}
	var bounds []int
	for _, h := range tx.Headings {
		if h.Level == level {
			bounds = append(bounds, runeToFolded(ft, h.RuneIdx))
		}
	}
	// Chunks() makes untitled front matter chunk 1 when it is substantial,
	// shifting chapter numbers by one; detect that the same way.
	shift := 0
	if len(bounds) > 0 && len(strings.Fields(ft.S[:bounds[0]])) >= 20 {
		shift = 1
	}
	chapterRange := func(ch int) (int, int) {
		k := ch - 1 - shift
		if k < 0 {
			if len(bounds) > 0 {
				return 0, bounds[0]
			}
			return 0, len(ft.S)
		}
		if k >= len(bounds) {
			return 0, len(ft.S)
		}
		hi := len(ft.S)
		if k+1 < len(bounds) {
			hi = bounds[k+1]
		}
		return bounds[k], hi
	}

	inserts := map[int][]string{}
	var trailing []string
	for ei, e := range idx.Entries {
		if len(e.Anchors) == 0 {
			if e.See != "" || len(e.SeeAlso) > 0 {
				trailing = append(trailing, markerFor(e, true, true))
				rep.CrossRefs++
			}
			continue
		}
		first := true
		for ai, a := range e.Anchors {
			rep.Anchors++
			lo, hi := chapterRange(a.Chapter)
			end, matched, fuzzy := findAnchor(ft.S, Fold(a.Text), lo, hi)
			if end < 0 {
				rep.Unmatched = append(rep.Unmatched, Unmatched{Heading: e.Heading, Subheading: e.Subheading, Chapter: a.Chapter, Text: a.Text})
				continue
			}
			off := sourceEnd(src, tx, ft, end)
			inserts[off] = append(inserts[off], markerFor(e, first, false))
			first = false
			rep.Placed++
			if fuzzy {
				rep.Fuzzy++
			}
			rep.Placements = append(rep.Placements, Placement{Entry: ei, Anchor: ai, Offset: off, Fuzzy: fuzzy, Matched: matched})
		}
		if first && (e.See != "" || len(e.SeeAlso) > 0) {
			// Every anchor failed but the entry carries cross-references:
			// keep those visible.
			trailing = append(trailing, markerFor(e, true, true))
		}
	}

	// Apply insertions back to front so offsets stay valid.
	offs := make([]int, 0, len(inserts))
	for o := range inserts {
		offs = append(offs, o)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(offs)))
	out := src
	for _, o := range offs {
		out = out[:o] + strings.Join(inserts[o], "") + out[o:]
	}
	if len(trailing) > 0 {
		out = strings.TrimRight(out, "\n") + "\n\n// Cross-references without locators (indexer).\n" + strings.Join(trailing, "\n") + "\n"
	}
	return out, rep
}

// findAnchor looks for folded in s[lo:hi], then in all of s, shortening the
// fragment when needed. Returns the folded end offset (exclusive), the
// fragment that matched, and whether it was shortened. -1 when not found.
func findAnchor(s, folded string, lo, hi int) (int, string, bool) {
	if folded == "" {
		return -1, "", false
	}
	try := func(f string) int {
		if i := strings.Index(s[lo:hi], f); i >= 0 {
			return lo + i + len(f)
		}
		if i := strings.Index(s, f); i >= 0 {
			return i + len(f)
		}
		return -1
	}
	if e := try(folded); e >= 0 {
		return e, folded, false
	}
	words := strings.Fields(folded)
	// Trim from the end, then from the start.
	for n := len(words) - 1; n >= minAnchorWords; n-- {
		f := strings.Join(words[:n], " ")
		if e := try(f); e >= 0 {
			return e, f, true
		}
	}
	for st := 1; len(words)-st >= minAnchorWords; st++ {
		f := strings.Join(words[st:], " ")
		if e := try(f); e >= 0 {
			return e, f, true
		}
	}
	return -1, "", false
}

// runeToFolded maps a rune index to the first folded byte at or after it.
func runeToFolded(ft FoldedText, ri int) int {
	return sort.SearchInts(ft.From, ri)
}

// sourceEnd maps a folded end offset (exclusive) to the byte offset in src
// just past the word that contains the last matched character.
func sourceEnd(src string, tx Text, ft FoldedText, end int) int {
	ri := ft.From[end-1]
	off := tx.Runes[ri].Off
	if src[off] == '\\' {
		off++
	}
	_, size := decodeRune(src[off:])
	off += size
	// Run to the end of the word so a marker never splits one, then over any
	// closing punctuation so the text item still ends at a natural break
	// (a marker between a word and its comma reflowed three paragraphs of
	// Ghosts; after the comma it reflowed none).
	for off < len(src) && isWordByte(src[off]) {
		_, size := decodeRune(src[off:])
		off += size
	}
	for off < len(src) {
		r, size := decodeRune(src[off:])
		if !strings.ContainsRune(".,;:!?\"'”’)»", r) {
			break
		}
		off += size
	}
	// And over one following space or newline when a word follows, so the
	// marker is glued to the start of the next word: a zero-width element
	// between a word and its trailing space cost Typst the break opportunity
	// at that space and reflowed whole paragraphs.
	if off+1 < len(src) && (src[off] == ' ' || src[off] == '\n') && isWordByte(src[off+1]) {
		off++
	}
	return off
}

// markerFor renders the Typst marker for an entry. Cross-references are
// attached to the entry's first marker only; noLocator marks a pure
// cross-reference that must not contribute a page.
func markerFor(e Entry, withRefs, noLocator bool) string {
	var b strings.Builder
	b.WriteString(`#index(`)
	b.WriteString(typstString(e.Heading))
	if e.Subheading != "" {
		b.WriteString(", ")
		b.WriteString(typstString(e.Subheading))
	}
	if withRefs {
		if e.See != "" {
			b.WriteString(", see: ")
			b.WriteString(typstString(e.See))
		}
		if len(e.SeeAlso) > 0 {
			// The Typst piece takes one see-also per marker and dedupes;
			// emit the rest as extra locator-less markers.
			b.WriteString(", see-also: ")
			b.WriteString(typstString(e.SeeAlso[0]))
		}
	}
	if noLocator {
		b.WriteString(", locator: false")
	}
	b.WriteString(");")
	if withRefs && len(e.SeeAlso) > 1 {
		for _, sa := range e.SeeAlso[1:] {
			fmt.Fprintf(&b, "#index(%s, see-also: %s, locator: false);", typstString(e.Heading), typstString(sa))
		}
	}
	return b.String()
}

func typstString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
}
