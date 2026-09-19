package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Index is the index.json document: the reviewed/reviewable draft that the
// anchoring step and (phase 2) the factory review UI work from.
type Index struct {
	Book          string    `json:"book"`
	Generated     time.Time `json:"generated"`
	Model         string    `json:"model"`
	Words         int       `json:"words"`
	PagesEstimate int       `json:"pages_estimate"`
	Budget        int       `json:"budget_entries"` // target ceiling on entry lines
	Chapters      []string  `json:"chapters"`
	Entries       []Entry   `json:"entries"`
	Usage         Usage     `json:"usage"`
	CostUSD       float64   `json:"cost_usd"`
	Notes         []string  `json:"notes,omitempty"` // merge decisions, drops
}

// Entry is one index line: a heading, optional run-in subheading, optional
// cross-references, and the anchors (verbatim text fragments) whose pages
// become its locators. Heading+Subheading identify the entry.
type Entry struct {
	Heading    string   `json:"heading"`
	Subheading string   `json:"subheading,omitempty"`
	See        string   `json:"see,omitempty"`
	SeeAlso    []string `json:"see_also,omitempty"`
	Anchors    []Anchor `json:"anchors"`
}

// Anchor is a fragment copied from the manuscript; Chapter is the 1-based
// chunk it came from (narrows the search when anchoring).
type Anchor struct {
	Chapter int    `json:"chapter"`
	Text    string `json:"text"`
}

// Chunk is one model call's worth of manuscript: normally a chapter.
type Chunk struct {
	Index int    // 1-based
	Title string // heading text, or "" for untitled material
	Text  string // plain text
	Words int
}

var headingRe = regexp.MustCompile(`(?m)^(=+) (.+)$`)

// Chunks splits the pandoc Typst into chapters at level-1 headings (level-2
// when parts is true), then splits any chapter longer than MaxInputChars at
// paragraph breaks. Front-matter pieces before the first heading form chunk
// 1 if they carry more than a few words.
func Chunks(src string, parts bool) []Chunk {
	level := 1
	if parts {
		level = 2
	}
	text := ExtractText(src).String()
	// Headings survive extraction as plain lines; re-find them in the source
	// so titles are clean, and cut the extracted text at the same lines.
	var cuts []int
	var titles []string
	lines := strings.Split(text, "\n")
	m := headingRe.FindAllStringSubmatch(src, -1)
	want := map[string]bool{}
	for _, h := range m {
		if len(h[1]) == level {
			want[strings.TrimSpace(ExtractText(h[2]).String())] = true
		}
	}
	for i, ln := range lines {
		t := strings.TrimSpace(ln)
		if t != "" && want[t] {
			cuts = append(cuts, i)
			titles = append(titles, t)
		}
	}
	var chunks []Chunk
	add := func(title string, body []string) {
		t := strings.TrimSpace(strings.Join(body, "\n"))
		t = regexp.MustCompile(`\n{3,}`).ReplaceAllString(t, "\n\n")
		if len(strings.Fields(t)) < 20 {
			return
		}
		for _, part := range splitLong(t, MaxInputChars) {
			chunks = append(chunks, Chunk{Index: len(chunks) + 1, Title: title, Text: part, Words: len(strings.Fields(part))})
		}
	}
	if len(cuts) == 0 {
		add("", lines)
		return chunks
	}
	if cuts[0] > 0 {
		add("", lines[:cuts[0]])
	}
	for k, c := range cuts {
		end := len(lines)
		if k+1 < len(cuts) {
			end = cuts[k+1]
		}
		add(titles[k], lines[c:end])
	}
	return chunks
}

func splitLong(t string, max int) []string {
	if len(t) <= max {
		return []string{t}
	}
	paras := strings.Split(t, "\n\n")
	var out []string
	var cur strings.Builder
	for _, p := range paras {
		if cur.Len() > 0 && cur.Len()+len(p) > max {
			out = append(out, cur.String())
			cur.Reset()
		}
		if cur.Len() > 0 {
			cur.WriteString("\n\n")
		}
		cur.WriteString(p)
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}

// WordsPerPage is the rule of thumb for the page estimate when the caller
// has no build to measure (A-format/6×9 at 10–10.5pt sets 330–400 words).
const WordsPerPage = 380

// EntriesPerIndexPage is roughly what two 9pt columns hold.
const EntriesPerIndexPage = 80

// MaxEntriesPerChunk caps what one chapter may contribute before the merge.
const MaxEntriesPerChunk = 60

// BudgetFor returns the ceiling on entry lines for a book of the given pages:
// an index of about 4 % of the book's length (the brief's 3–5 %).
func BudgetFor(pages int) int {
	if pages < 1 {
		pages = 1
	}
	return int(math.Ceil(float64(pages) * 0.04 * EntriesPerIndexPage))
}

// Options for Draft.
type Options struct {
	Book    string // title, for the prompt
	About   string // one line on the book's kind/audience (optional)
	Pages   int    // 0 → estimate from words
	Parts   bool   // level-1 headings are parts, chapters are level 2
	Log     func(format string, a ...any)
	Merge   bool // run the cross-chapter consolidation call (default true via Draft)
	NoMerge bool
}

// SystemPrompt is the indexer persona for the per-chapter call.
const SystemPrompt = `You are a professional back-of-book indexer working to the Chicago Manual of Style (chapter 16). You index one chapter at a time and return JSON only.

What to index: the subjects a reader would look up — concepts, themes, technologies, institutions, named people, places, works, and recurring motifs that the chapter actually discusses. For fiction and essays, index the ideas, settings, institutions and characters that carry the piece, not every noun.
What not to index: passing mentions, generic words, the chapter's own title as a heading (unless it is a topic), anything discussed for less than a sentence. An index is not a concordance; fewer, better entries win.

Headings: noun or noun phrase, lower-case unless a proper noun, singular unless the plural is the natural term. Personal names inverted ("Wiener, Norbert"). Keep the wording a reader would look up first; if a synonym is equally likely, give a "see" entry (heading = synonym, see = the heading you used, anchors = []).
Subheadings: only when a heading is discussed under clearly different aspects; short (2–5 words), starting lower-case, no leading "the"/"and". Otherwise leave subheading empty.
Anchors: each entry lists 1–4 anchors, each an EXACT contiguous fragment of 5–12 words copied verbatim from the chapter text (same spelling, no ellipsis, no paraphrase). Pick fragments that contain no quotation marks. One anchor per distinct discussion, placed where the discussion starts; if a discussion runs for many paragraphs, add a second anchor near its end so the index can show a page range. Do not anchor every mention.

Return compact JSON (no indentation, no markdown fence, no commentary) in exactly this shape:
{"entries":[{"heading":"…","subheading":"","see":"","see_also":[],"anchors":["…","…"]}]}`

// MergeSystemPrompt is the cross-chapter consolidation persona.
const MergeSystemPrompt = `You are the same indexer, now reading your draft headings from every chapter together to produce one consistent index. Return JSON only.

Tasks: (1) merge synonyms and variant wordings into one canonical heading ("the pass"/"Factory Pass"; "AI"/"artificial intelligence"; singular/plural; name spelled two ways) — choose the wording a reader looks up first; (2) where a merged-away wording is one a reader might also try, keep it as a "see" cross-reference; (3) add "see also" links between related headings that are both kept; (4) if the draft is over budget, drop the weakest headings — thin ones with a single anchor and no subheadings, generic ones — until it fits. Never invent headings that are not in the draft.

Return compact JSON (no indentation, no fence, no commentary), exactly:
{"merge":[{"from":"heading as drafted","to":"canonical heading"}],"see":[{"heading":"synonym a reader might try","to":"canonical heading"}],"see_also":[{"heading":"…","also":"…"}],"drop":["heading","…"]}`

type chapterReply struct {
	Entries []struct {
		Heading    string   `json:"heading"`
		Subheading string   `json:"subheading"`
		See        string   `json:"see"`
		SeeAlso    []string `json:"see_also"`
		Anchors    []string `json:"anchors"`
	} `json:"entries"`
}

type mergeReply struct {
	Merge   []struct{ From, To string }      `json:"merge"`
	See     []struct{ Heading, To string }   `json:"see"`
	SeeAlso []struct{ Heading, Also string } `json:"see_also"`
	Drop    []string                         `json:"drop"`
}

// Draft runs the whole pass: chunk, one call per chunk, merge, budget.
func Draft(ctx context.Context, c *Client, src string, opt Options) (*Index, error) {
	logf := opt.Log
	if logf == nil {
		logf = func(string, ...any) {}
	}
	if c.Log == nil {
		c.Log = logf
	}
	chunks := Chunks(src, opt.Parts)
	if len(chunks) == 0 {
		return nil, fmt.Errorf("no text found")
	}
	words := 0
	for _, ch := range chunks {
		words += ch.Words
	}
	pages := opt.Pages
	if pages == 0 {
		pages = int(math.Ceil(float64(words) / WordsPerPage))
	}
	idx := &Index{Book: opt.Book, Generated: time.Now().UTC(), Model: c.model(), Words: words,
		PagesEstimate: pages, Budget: BudgetFor(pages)}
	for _, ch := range chunks {
		idx.Chapters = append(idx.Chapters, ch.Title)
	}
	logf("%d chunks, %d words, ~%d pages, budget %d entry lines", len(chunks), words, pages, idx.Budget)

	// Per-chapter calls.
	var raw []Entry
	for _, ch := range chunks {
		share := shareFor(idx.Budget, ch.Words, words)
		user := chapterPrompt(opt, ch, share)
		text, u, err := c.Complete(ctx, SystemPrompt, user)
		idx.Usage.Add(u)
		if err != nil {
			return idx, fmt.Errorf("chunk %d (%s): %w", ch.Index, ch.Title, err)
		}
		var rep chapterReply
		if err := json.Unmarshal([]byte(RepairJSON(stripFences(text))), &rep); err != nil {
			return idx, fmt.Errorf("chunk %d (%s): bad JSON from model: %w\n%.300s", ch.Index, ch.Title, err, text)
		}
		n := 0
		for _, e := range rep.Entries {
			h := cleanHeading(e.Heading)
			if h == "" {
				continue
			}
			en := Entry{Heading: h, Subheading: cleanHeading(e.Subheading), See: cleanHeading(e.See)}
			for _, sa := range e.SeeAlso {
				if s := cleanHeading(sa); s != "" {
					en.SeeAlso = append(en.SeeAlso, s)
				}
			}
			for _, a := range e.Anchors {
				a = strings.TrimSpace(a)
				if a != "" {
					en.Anchors = append(en.Anchors, Anchor{Chapter: ch.Index, Text: a})
				}
			}
			raw = append(raw, en)
			n++
		}
		logf("chunk %d %q: %d entries (asked ≤ %d)", ch.Index, ch.Title, n, share)
	}
	idx.Entries = mergeIdentical(raw)

	if !opt.NoMerge && len(chunks) > 1 {
		if err := consolidate(ctx, c, idx, logf); err != nil {
			idx.Notes = append(idx.Notes, "consolidation skipped: "+err.Error())
			logf("consolidation failed, keeping per-chapter draft: %v", err)
		}
	}
	idx.Entries = tidy(idx.Entries)
	sortEntries(idx.Entries)
	idx.CostUSD = idx.Usage.CostUSD(idx.Model)
	return idx, nil
}

// tidy enforces the invariants the Typst piece and the review UI rely on:
// no self-references, a "see" only on an entry without locators (otherwise it
// becomes "see also"), non-nil anchors, no duplicate see-alsos.
func tidy(es []Entry) []Entry {
	for i := range es {
		e := &es[i]
		if e.Anchors == nil {
			e.Anchors = []Anchor{}
		}
		if e.See != "" && Fold(e.See) == Fold(e.Heading) {
			e.See = ""
		}
		if e.See != "" && (len(e.Anchors) > 0 || e.Subheading != "") {
			e.SeeAlso = appendUnique(e.SeeAlso, e.See)
			e.See = ""
		}
		var sa []string
		for _, s := range e.SeeAlso {
			if Fold(s) != Fold(e.Heading) {
				sa = appendUnique(sa, s)
			}
		}
		e.SeeAlso = sa
	}
	return es
}

var stopword = map[string]bool{"the": true, "and": true, "for": true, "with": true, "from": true}

// shareWord reports whether two headings have a folded word in common
// (ignoring articles and very short words).
func shareWord(a, b string) bool {
	set := map[string]bool{}
	for _, w := range strings.Fields(Fold(a)) {
		if len(w) > 2 && !stopword[w] {
			set[w] = true
		}
	}
	for _, w := range strings.Fields(Fold(b)) {
		if set[w] {
			return true
		}
	}
	return false
}

// shareFor is a chapter's slice of the budget: ∝ words, +15 % slack for the
// merge to trim, never more than a chapter can sensibly carry.
func shareFor(budget, chapterWords, totalWords int) int {
	share := int(math.Ceil(float64(budget) * 1.15 * float64(chapterWords) / float64(totalWords)))
	if share < 6 {
		share = 6
	}
	if share > MaxEntriesPerChunk {
		share = MaxEntriesPerChunk
	}
	return share
}

func chapterPrompt(opt Options, ch Chunk, share int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Book: %s\n", opt.Book)
	if opt.About != "" {
		fmt.Fprintf(&b, "About the book: %s\n", opt.About)
	}
	title := ch.Title
	if title == "" {
		title = "(untitled front matter)"
	}
	fmt.Fprintf(&b, "Chapter %d: %s\n", ch.Index, title)
	fmt.Fprintf(&b, "Entry budget for this chapter: at most %d entries (heading+subheading lines). Choose well.\n\n", share)
	b.WriteString("=== CHAPTER TEXT ===\n")
	b.WriteString(ch.Text)
	b.WriteString("\n=== END ===\n")
	return b.String()
}

// mergePrompt presents the per-chapter draft compactly for the
// consolidation call: heading (n anchors) [→ see x]: subheadings.
func mergePrompt(idx *Index) string {
	type hs struct {
		subs    []string
		anchors int
		see     string
	}
	byHead := map[string]*hs{}
	var order []string
	for _, e := range idx.Entries {
		k := e.Heading
		h, ok := byHead[k]
		if !ok {
			h = &hs{}
			byHead[k] = h
			order = append(order, k)
		}
		if e.Subheading != "" {
			h.subs = append(h.subs, e.Subheading)
		}
		h.anchors += len(e.Anchors)
		if e.See != "" {
			h.see = e.See
		}
	}
	sort.Strings(order)
	var b strings.Builder
	fmt.Fprintf(&b, "Book: %s. Budget: %d entry lines; the draft has %d headings and %d lines.\n\n=== DRAFT HEADINGS ===\n", idx.Book, idx.Budget, len(order), len(idx.Entries))
	for _, k := range order {
		h := byHead[k]
		fmt.Fprintf(&b, "%s (%d anchors)", k, h.anchors)
		if h.see != "" {
			fmt.Fprintf(&b, " → see %s", h.see)
		}
		if len(h.subs) > 0 {
			fmt.Fprintf(&b, ": %s", strings.Join(h.subs, "; "))
		}
		b.WriteString("\n")
	}
	b.WriteString("=== END ===\n")
	return b.String()
}

// consolidate runs the merge call and applies it.
func consolidate(ctx context.Context, c *Client, idx *Index, logf func(string, ...any)) error {
	text, u, err := c.Complete(ctx, MergeSystemPrompt, mergePrompt(idx))
	idx.Usage.Add(u)
	if err != nil {
		return err
	}
	var rep mergeReply
	if err := json.Unmarshal([]byte(RepairJSON(stripFences(text))), &rep); err != nil {
		return fmt.Errorf("bad merge JSON: %w", err)
	}
	// Apply: rename, then drop, then add see / see-also, then re-merge.
	rename := map[string]string{}
	var autoSee [][2]string
	for _, m := range rep.Merge {
		from, to := cleanHeading(m.From), cleanHeading(m.To)
		if from != "" && to != "" && from != to {
			rename[Fold(from)] = to
			idx.Notes = append(idx.Notes, fmt.Sprintf("merge %q → %q", from, to))
			// A merged-away wording that shares no word with its target is
			// one a reader might still look up: keep it as a see-reference
			// ("beaings. See artificial ghosts").
			if !shareWord(from, to) {
				autoSee = append(autoSee, [2]string{from, to})
			}
		}
	}
	// Chase chains a→b→c.
	resolve := func(h string) string {
		for i := 0; i < 5; i++ {
			to, ok := rename[Fold(h)]
			if !ok || Fold(to) == Fold(h) {
				return h
			}
			h = to
		}
		return h
	}
	drop := map[string]bool{}
	for _, d := range rep.Drop {
		drop[Fold(cleanHeading(d))] = true
	}
	var out []Entry
	dropped := 0
	for _, e := range idx.Entries {
		e.Heading = resolve(e.Heading)
		if e.See != "" {
			e.See = resolve(e.See)
		}
		for i := range e.SeeAlso {
			e.SeeAlso[i] = resolve(e.SeeAlso[i])
		}
		if drop[Fold(e.Heading)] {
			dropped++
			continue
		}
		out = append(out, e)
	}
	if dropped > 0 {
		idx.Notes = append(idx.Notes, fmt.Sprintf("dropped %d entries under %d headings to fit the budget", dropped, len(rep.Drop)))
	}
	have := map[string]bool{}
	for _, e := range out {
		have[Fold(e.Heading)] = true
	}
	sees := autoSee
	for _, s := range rep.See {
		sees = append(sees, [2]string{cleanHeading(s.Heading), cleanHeading(s.To)})
	}
	for _, s := range sees {
		h, to := s[0], resolve(s[1])
		if h == "" || to == "" || have[Fold(h)] || !have[Fold(to)] || Fold(h) == Fold(to) {
			continue
		}
		out = append(out, Entry{Heading: h, See: to})
		have[Fold(h)] = true
	}
	for _, s := range rep.SeeAlso {
		h, also := resolve(cleanHeading(s.Heading)), resolve(cleanHeading(s.Also))
		if !have[Fold(h)] || !have[Fold(also)] || Fold(h) == Fold(also) {
			continue
		}
		for i := range out {
			if Fold(out[i].Heading) == Fold(h) && out[i].Subheading == "" {
				out[i].SeeAlso = appendUnique(out[i].SeeAlso, also)
				break
			}
		}
	}
	idx.Entries = mergeIdentical(out)
	logf("consolidated: %d merges, %d see, %d see-also, %d drops → %d entries", len(rep.Merge), len(rep.See), len(rep.SeeAlso), len(rep.Drop), len(idx.Entries))
	return nil
}

// mergeIdentical joins entries with the same folded heading+subheading,
// keeping the first-seen wording and the union of anchors and refs.
func mergeIdentical(in []Entry) []Entry {
	pos := map[string]int{}
	var out []Entry
	for _, e := range in {
		k := Fold(e.Heading) + "|" + Fold(e.Subheading)
		if i, ok := pos[k]; ok {
			out[i].Anchors = append(out[i].Anchors, e.Anchors...)
			if out[i].See == "" {
				out[i].See = e.See
			}
			for _, s := range e.SeeAlso {
				out[i].SeeAlso = appendUnique(out[i].SeeAlso, s)
			}
			continue
		}
		pos[k] = len(out)
		out = append(out, e)
	}
	// Dedupe anchors within an entry.
	for i := range out {
		seen := map[string]bool{}
		var as []Anchor
		for _, a := range out[i].Anchors {
			k := Fold(a.Text)
			if k == "" || seen[k] {
				continue
			}
			seen[k] = true
			as = append(as, a)
		}
		out[i].Anchors = as
	}
	return out
}

func appendUnique(xs []string, s string) []string {
	for _, x := range xs {
		if Fold(x) == Fold(s) {
			return xs
		}
	}
	return append(xs, s)
}

func sortEntries(es []Entry) {
	sort.SliceStable(es, func(i, j int) bool {
		a, b := Fold(es[i].Heading), Fold(es[j].Heading)
		if a != b {
			return a < b
		}
		return Fold(es[i].Subheading) < Fold(es[j].Subheading)
	})
}

// cleanHeading trims quotes, trailing punctuation and doubled spaces.
func cleanHeading(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, `"“”'`)
	s = strings.TrimRight(s, ".,;:")
	s = strings.Join(strings.Fields(s), " ")
	return s
}

// stripFences removes ```json fences and any prose around the outermost JSON
// object, which models add despite instructions.
func stripFences(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "{"); i > 0 {
		s = s[i:]
	}
	if j := strings.LastIndex(s, "}"); j >= 0 && j < len(s)-1 {
		s = s[:j+1]
	}
	return s
}

// RepairJSON escapes double quotes that a model left unescaped inside JSON
// strings (anchors copied from dialogue). Inside a string, a `"` counts as
// the closing quote only if the next non-space character can follow a string
// (`,` `]` `}` `:`) or the input ends; otherwise it is content and gets
// escaped. Valid JSON passes through unchanged.
func RepairJSON(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	in := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if in {
			switch c {
			case '\\':
				b.WriteByte(c)
				if i+1 < len(s) {
					i++
					b.WriteByte(s[i])
				}
				continue
			case '"':
				j := i + 1
				for j < len(s) && (s[j] == ' ' || s[j] == '\n' || s[j] == '\t' || s[j] == '\r') {
					j++
				}
				if j >= len(s) || s[j] == ',' || s[j] == ']' || s[j] == '}' || s[j] == ':' {
					in = false
					b.WriteByte(c)
				} else {
					b.WriteString(`\"`)
				}
				continue
			}
			b.WriteByte(c)
			continue
		}
		if c == '"' {
			in = true
		}
		b.WriteByte(c)
	}
	return b.String()
}
