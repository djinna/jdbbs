package srv

// Book map (P4, docs/reviews/P4-FRONT-MATTER-PLAN-2026-09-18.md).
//
// One rule for every manuscript: every section head is Heading 1. We decide
// which H1s are front matter, body and back matter from the heading text
// (a closed vocabulary) plus position, and we find the untitled front-matter
// pieces (dedication, epigraph) as the page-break-separated blocks before the
// first H1. Title/Subtitle-styled paragraphs are dropped (the title page is
// generated from the transmittal) and reported.
//
// This reads word/document.xml directly rather than the pandoc AST because
// pandoc's docx reader discards page breaks and lifts Title/Subtitle into
// metadata — exactly the two signals the map depends on.

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// docxPara is one non-empty paragraph as the map sees it.
type docxPara struct {
	Style           string // normalized style name: "heading1", "title", "normal", …
	Text            string
	HasImage        bool
	PageBreakBefore bool // a hard page break (w:br type=page / pageBreakBefore / section break) precedes it
}

// headingLevel returns 1..9 for Heading N styles, 0 otherwise.
func (p docxPara) headingLevel() int {
	m := headingStyleRe.FindStringSubmatch(p.Style)
	if m == nil {
		return 0
	}
	return int(m[1][0] - '0')
}

var headingStyleRe = regexp.MustCompile(`^heading([1-9])$`)

// BookMapSection is one H1 with its classification.
type BookMapSection struct {
	Title string `json:"title"`
	Kind  string `json:"kind"`  // "front" | "body" | "back" | "title" (dropped) | "toc" (dropped)
	Paras int    `json:"paras"` // paragraphs under the heading, before the next H1
}

// BookMapUntitled is one untitled front-matter block (before the first H1).
type BookMapUntitled struct {
	Name    string `json:"name"`           // "dedication", "epigraph", "copyright", or "untitled front-matter page N"
	Paras   int    `json:"paras"`          // paragraph count (non-empty, as pandoc will see them)
	Preview string `json:"preview"`        // first ~60 chars
	Drop    bool   `json:"drop,omitempty"` // dropped by the build (byline, typed copyright)
}

// keptUntitled returns the untitled pieces that make it into the book.
func (m *BookMap) keptUntitled() []BookMapUntitled {
	out := []BookMapUntitled{}
	for _, u := range m.UntitledFront {
		if !u.Drop {
			out = append(out, u)
		}
	}
	return out
}

// BookMap is the classification of a manuscript into front / body / back.
type BookMap struct {
	Title         string            `json:"title,omitempty"`    // Title-styled text found (dropped)
	Subtitle      string            `json:"subtitle,omitempty"` // Subtitle-styled text found (dropped)
	UntitledFront []BookMapUntitled `json:"untitled_front"`
	Sections      []BookMapSection  `json:"sections"` // every H1 in order
	Warnings      []string          `json:"warnings"` // things to fix (medium)
	Notes         []string          `json:"notes"`    // things we did that the author should know (low)
	SummaryLine   string            `json:"summary"`  // Summary(), stored so JSON consumers get it
	Parts         bool              `json:"parts"`    // spec opt-in: H1 = part, H2 = chapter
}

// Front / Body / Back return the H1 titles of each kind, in order.
func (m *BookMap) Front() []string { return m.titles("front") }
func (m *BookMap) Body() []string  { return m.titles("body") }
func (m *BookMap) Back() []string  { return m.titles("back") }

func (m *BookMap) titles(kind string) []string {
	out := []string{}
	for _, s := range m.Sections {
		if s.Kind == kind {
			out = append(out, s.Title)
		}
	}
	return out
}

// Summary is the one-line book map Inspect prints, e.g.
// "Front matter: Dedication, Epigraph, Foreword · Body starts at “Chapter 1” (12 sections) · Back matter: Notes, About the Author".
func (m *BookMap) Summary() string {
	var parts []string
	front := []string{}
	for _, u := range m.keptUntitled() {
		front = append(front, titleCase(u.Name))
	}
	front = append(front, m.Front()...)
	if len(front) > 0 {
		parts = append(parts, "Front matter: "+strings.Join(front, ", "))
	} else {
		parts = append(parts, "Front matter: none")
	}
	body := m.Body()
	switch len(body) {
	case 0:
		parts = append(parts, "Body: no section headings found")
	case 1:
		parts = append(parts, fmt.Sprintf("Body starts at “%s” (1 section)", body[0]))
	default:
		parts = append(parts, fmt.Sprintf("Body starts at “%s” (%d sections)", body[0], len(body)))
	}
	if back := m.Back(); len(back) > 0 {
		parts = append(parts, "Back matter: "+strings.Join(back, ", "))
	} else {
		parts = append(parts, "Back matter: none")
	}
	return strings.Join(parts, " · ")
}

func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// --- vocabulary --------------------------------------------------------------

// Closed vocabularies, matched against normalizeHeading(text). Decisions
// 2026-09-17: Introduction is body (page 1). Prologue is front per Jenna.
// Contents is handled separately (isContentsHeading): the section is dropped.
var frontMatterTerms = map[string]bool{
	"foreword": true, "preface": true, "prologue": true,
	"acknowledgments": true, "acknowledgements": true,
	"dedication": true, "epigraph": true,
	"list of figures": true, "list of tables": true, "list of illustrations": true,
	"list of maps": true, "list of plates": true, "illustrations": true,
	"abbreviations": true, "list of abbreviations": true,
	"authors note": true, "translators note": true, "editors note": true,
	"publishers note": true, "note on the text": true, "note on the translation": true,
	"note to the reader": true, "how to use this book": true, "how to read this book": true,
	"chronology": true, "dramatis personae": true, "cast of characters": true,
	"characters": true, "maps": true, "family tree": true,
}

var backMatterTerms = map[string]bool{
	"appendix": true, "appendices": true,
	"notes": true, "endnotes": true, "notes on sources": true, "sources": true,
	"bibliography": true, "references": true, "works cited": true,
	"further reading": true, "suggested reading": true, "selected bibliography": true,
	"glossary": true, "index": true,
	"about the author": true, "about the authors": true, "about the translator": true,
	"about the editor": true, "about the contributors": true, "contributors": true,
	"colophon": true, "about the type": true, "note on the type": true,
	"afterword": true, "acknowledgments": true, "acknowledgements": true,
	"credits": true, "permissions": true, "image credits": true, "photo credits": true,
	"also by the author": true, "also by": true,
}

// Prefix rules: "Appendix A: Tables", "A Note on the Text", "List of …".
var frontMatterPrefixes = []string{"note on ", "list of "}
var backMatterPrefixes = []string{"appendix ", "also by ", "notes to ", "notes on "}

var headingNormRe = regexp.MustCompile(`[^a-z0-9 ]+`)

// normalizeHeading lowercases, strips punctuation and a leading article
// ("A Note on the Text" → "note on the text").
func normalizeHeading(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.NewReplacer("’", "", "'", "", "—", " ", "–", " ", "-", " ", ":", " ").Replace(s)
	s = headingNormRe.ReplaceAllString(s, " ")
	s = strings.Join(strings.Fields(s), " ")
	for _, art := range []string{"a ", "an ", "the "} {
		if strings.HasPrefix(s, art) {
			s = strings.TrimPrefix(s, art)
			break
		}
	}
	return s
}

const headingVocabMaxWords = 5

func matchesVocab(text string, terms map[string]bool, prefixes []string) bool {
	n := normalizeHeading(text)
	if n == "" || len(strings.Fields(n)) > headingVocabMaxWords {
		return false
	}
	if terms[n] {
		return true
	}
	for _, p := range prefixes {
		if strings.HasPrefix(n, p) {
			return true
		}
	}
	return false
}

func isFrontMatterHeading(text string) bool {
	return matchesVocab(text, frontMatterTerms, frontMatterPrefixes)
}
func isBackMatterHeading(text string) bool {
	return matchesVocab(text, backMatterTerms, backMatterPrefixes)
}

// --- classification ------------------------------------------------------------

// untitledFrontNames is the transmittal order for untitled front-matter pages.
var untitledFrontNames = []string{"dedication", "epigraph"}

// buildBookMap classifies paragraphs. untitledNames is the transmittal's list
// of untitled front-matter pieces in order (subset of untitledFrontNames);
// nil means "not declared", in which case blocks are named positionally.
// bookTitle / bookAuthor (from the transmittal / book record) let a first H1
// that repeats the title, and a byline, be recognised and dropped (the pre-P4
// template model).
func buildBookMap(paras []docxPara, untitledNames []string, bookTitle, bookAuthor string) *BookMap {
	m := buildBookMapInner(paras, untitledNames, bookTitle, bookAuthor)
	m.SummaryLine = m.Summary()
	return m
}

func buildBookMapInner(paras []docxPara, untitledNames []string, bookTitle, bookAuthor string) *BookMap {
	m := &BookMap{UntitledFront: []BookMapUntitled{}, Sections: []BookMapSection{}, Warnings: []string{}, Notes: []string{}}

	// Title / Subtitle: dropped, reported.
	var body []docxPara
	for _, p := range paras {
		switch p.Style {
		case "title":
			if m.Title == "" {
				m.Title = p.Text
			} else {
				m.Title += " " + p.Text
			}
			continue
		case "subtitle":
			if m.Subtitle == "" {
				m.Subtitle = p.Text
			} else {
				m.Subtitle += " " + p.Text
			}
			continue
		}
		body = append(body, p)
	}
	if m.Title != "" {
		m.Notes = append(m.Notes, fmt.Sprintf("Title-styled paragraph “%s” dropped: the title page is generated from the transmittal.", m.Title))
	}
	if m.Subtitle != "" {
		m.Notes = append(m.Notes, fmt.Sprintf("Subtitle-styled paragraph “%s” dropped: the title page is generated from the transmittal.", m.Subtitle))
	}

	// First H1.
	firstH1 := -1
	for i, p := range body {
		if p.headingLevel() == 1 {
			firstH1 = i
			break
		}
	}
	if firstH1 < 0 {
		m.Warnings = append(m.Warnings, "No Heading 1 found: the whole manuscript is treated as body text with no chapter breaks. Style every section head as Heading 1.")
		return m
	}

	// Untitled front matter: the blocks before the first H1. A block is a run
	// of paragraphs split on hard page breaks, and also on the template's
	// dedicated styles (Epigraph, Dedication, Copyright) so a style change
	// starts a new piece even without a page break.
	type preBlock struct {
		paras []docxPara
		style string // "epigraph" | "dedication" | "copyright" | "" (plain)
		drop  bool
		note  string
	}
	pieceStyle := func(p docxPara) string {
		switch p.Style {
		case "epigraph", "dedication", "copyright":
			return p.Style
		}
		return ""
	}
	var blocks []preBlock
	for _, p := range body[:firstH1] {
		st := pieceStyle(p)
		// A byline for the author ("by Jane Author" / "Jane Author") is part
		// of the generated title page; drop it.
		if st == "" && bookAuthor != "" {
			n := normalizeHeading(p.Text)
			if n == normalizeHeading(bookAuthor) || n == "by "+normalizeHeading(bookAuthor) {
				blocks = append(blocks, preBlock{paras: []docxPara{p}, drop: true,
					note: fmt.Sprintf("Byline “%s” dropped: the author goes on the generated title page.", preview(p.Text))})
				continue
			}
		}
		last := len(blocks) - 1
		if last < 0 || p.PageBreakBefore || blocks[last].drop || st != blocks[last].style {
			blocks = append(blocks, preBlock{style: st})
			last++
		}
		blocks[last].paras = append(blocks[last].paras, p)
	}
	for i := range blocks {
		if blocks[i].style == "copyright" {
			blocks[i].drop = true
			blocks[i].note = fmt.Sprintf("Copyright-styled %s dropped: the copyright page is generated from the transmittal.", plural(len(blocks[i].paras), "paragraph"))
		}
	}
	// Names for the plain blocks: the transmittal's declared pieces first,
	// then the canonical order for anything it didn't declare (a dedication
	// is still a dedication when the box wasn't ticked). Style-named blocks
	// take their style's name and use up that name.
	used := map[string]bool{}
	for _, b := range blocks {
		if !b.drop && b.style != "" {
			used[b.style] = true
		}
	}
	var names []string
	for _, n := range append(append([]string{}, untitledNames...), untitledFrontNames...) {
		if !used[n] {
			used[n] = true
			names = append(names, n)
		}
	}
	plainCount, nextName := 0, 0
	for _, b := range blocks {
		if b.drop {
			m.Notes = append(m.Notes, b.note)
			m.UntitledFront = append(m.UntitledFront, BookMapUntitled{Name: "dropped", Paras: len(b.paras), Preview: preview(b.paras[0].Text), Drop: true})
			continue
		}
		name := b.style
		if name == "" {
			plainCount++
			name = fmt.Sprintf("untitled front-matter page %d", plainCount)
			if nextName < len(names) {
				name = names[nextName]
				nextName++
			}
		}
		m.UntitledFront = append(m.UntitledFront, BookMapUntitled{Name: name, Paras: len(b.paras), Preview: preview(b.paras[0].Text)})
	}
	kept := m.keptUntitled()
	if untitledNames != nil && len(kept) > len(untitledNames) {
		listed := "none"
		if len(untitledNames) > 0 {
			listed = strings.Join(untitledNames, ", ")
		}
		var keptNames []string
		for _, u := range kept {
			keptNames = append(keptNames, u.Name)
		}
		m.Notes = append(m.Notes, fmt.Sprintf("%s before the first heading; the transmittal lists %s. All are kept, in order, as %s.", plural(len(kept), "untitled page"), listed, strings.Join(keptNames, ", ")))
	}
	// A single plain block of several paragraphs where two pieces were
	// declared is the usual mistake (no page break between dedication and epigraph).
	if len(kept) == 1 && kept[0].Paras > 1 && len(untitledNames) > 1 {
		m.Warnings = append(m.Warnings, fmt.Sprintf("One untitled block of %d paragraphs before the first heading, taken as %s. Put a page break between separate front-matter pieces.", kept[0].Paras, kept[0].Name))
	}
	for _, want := range untitledNames {
		found := false
		for _, u := range kept {
			found = found || u.Name == want
		}
		if !found {
			m.Warnings = append(m.Warnings, fmt.Sprintf("Transmittal lists %s but no untitled page was found before the first heading.", titleCase(want)))
		}
	}

	// H1 sections: heading + the paragraphs up to the next H1.
	type span struct {
		title string
		paras []docxPara
	}
	var spans []span
	for _, p := range body[firstH1:] {
		if p.headingLevel() == 1 {
			spans = append(spans, span{title: strings.TrimSpace(p.Text)})
			continue
		}
		spans[len(spans)-1].paras = append(spans[len(spans)-1].paras, p)
	}

	// Legacy template model: the book title typed as the first H1 (with a
	// byline / placeholder paragraphs under it). Dropped like a Title style.
	kinds := make([]string, len(spans))
	if len(spans) > 0 {
		first := normalizeHeading(spans[0].title)
		if first != "" && (first == normalizeHeading(bookTitle) || first == normalizeHeading(m.Title)) {
			kinds[0] = "title"
			m.Notes = append(m.Notes, fmt.Sprintf("Heading “%s” matches the book title and was dropped with the %s under it (%s): the title and copyright pages are generated from the transmittal.",
				spans[0].title, plural(len(spans[0].paras), "paragraph"), previewParas(spans[0].paras)))
		}
	}
	// A typed Contents section is dropped too: the build generates it.
	for i, sp := range spans {
		if kinds[i] == "" && isContentsHeading(sp.title) {
			kinds[i] = "toc"
			m.Notes = append(m.Notes, fmt.Sprintf("“%s” and the %s under it dropped: the contents page is generated by the build.", sp.title, plural(len(sp.paras), "paragraph")))
		}
	}

	// Classify the remaining headings: leading front run, trailing back run.
	var idx []int // indices of classifiable spans
	for i := range spans {
		if kinds[i] == "" {
			idx = append(idx, i)
		}
	}
	frontEnd := 0
	for frontEnd < len(idx) && isFrontMatterHeading(spans[idx[frontEnd]].title) {
		frontEnd++
	}
	backStart := len(idx)
	for backStart > frontEnd && isBackMatterHeading(spans[idx[backStart-1]].title) {
		backStart--
	}
	if len(idx) > 0 && frontEnd == len(idx) {
		m.Warnings = append(m.Warnings, "Every heading matches the front-matter vocabulary; there is no body. Treating them all as body chapters.")
		frontEnd, backStart = 0, len(idx)
	}
	for j, i := range idx {
		t := spans[i].title
		switch {
		case j < frontEnd:
			kinds[i] = "front"
		case j >= backStart:
			kinds[i] = "back"
		default:
			kinds[i] = "body"
			if isFrontMatterHeading(t) {
				m.Warnings = append(m.Warnings, fmt.Sprintf("“%s” looks like front matter but comes after the body starts; it is set as a body chapter. Move it before “%s” or rename it.", t, spans[idx[frontEnd]].title))
			} else if isBackMatterHeading(t) {
				m.Warnings = append(m.Warnings, fmt.Sprintf("“%s” looks like back matter but a chapter follows it; it is set as a body chapter. Move it after the last chapter or rename it.", t))
			}
		}
	}
	for i, sp := range spans {
		m.Sections = append(m.Sections, BookMapSection{Title: sp.title, Kind: kinds[i], Paras: len(sp.paras)})
	}
	return m
}

func isContentsHeading(text string) bool {
	n := normalizeHeading(text)
	return n == "contents" || n == "table of contents"
}

func plural(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return fmt.Sprintf("%d %ss", n, noun)
}

func previewParas(ps []docxPara) string {
	if len(ps) == 0 {
		return "nothing"
	}
	var out []string
	for i, p := range ps {
		if i == 3 {
			out = append(out, "…")
			break
		}
		out = append(out, "“"+preview(p.Text)+"”")
	}
	return strings.Join(out, ", ")
}

func preview(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	r := []rune(s)
	if len(r) > 60 {
		return string(r[:57]) + "…"
	}
	return s
}

// --- cross-check against the transmittal ---------------------------------------

// specFrontMatterKeys maps transmittal booleans to the heading they imply.
// Generated pieces (half_title, title_page, copyright_page, toc) and body
// pieces (introduction) are not headings the author types, so they are absent.
var specFrontMatterKeys = map[string]string{
	"foreword": "Foreword", "preface": "Preface", "acknowledgments": "Acknowledgments",
}
var specBackMatterKeys = map[string]string{
	"notes": "Notes", "appendix": "Appendix", "glossary": "Glossary",
	"bibliography": "Bibliography", "index": "Index",
}

// untitledNamesFromSpec returns the untitled pieces the transmittal declares,
// in canonical order; nil when the spec has no front_matter block.
func untitledNamesFromSpec(spec map[string]any) []string {
	fm, ok := spec["front_matter"].(map[string]any)
	if !ok {
		return nil
	}
	out := []string{}
	for _, n := range untitledFrontNames {
		if v, _ := fm[n].(bool); v {
			out = append(out, n)
		}
	}
	return out
}

// crossCheckSpec appends "listed but not found" / "found but not listed"
// warnings for the titled front/back pieces the transmittal tracks.
func (m *BookMap) crossCheckSpec(spec map[string]any) {
	check := func(block string, keys map[string]string, found []string) {
		sec, _ := spec[block].(map[string]any)
		have := map[string]bool{}
		for _, t := range found {
			have[normalizeHeading(t)] = true
		}
		names := make([]string, 0, len(keys))
		for k := range keys {
			names = append(names, k)
		}
		sort.Strings(names)
		for _, k := range names {
			label := keys[k]
			listed, _ := sec[k].(bool)
			present := have[normalizeHeading(label)]
			if k == "acknowledgments" && !present {
				present = have["acknowledgements"]
			}
			if k == "appendix" && !present {
				for h := range have {
					if strings.HasPrefix(h, "appendi") {
						present = true
					}
				}
			}
			if k == "bibliography" && !present {
				present = have["references"] || have["works cited"] || have["selected bibliography"]
			}
			if k == "notes" && !present {
				present = have["endnotes"]
			}
			switch {
			case listed && !present && sec != nil:
				m.Warnings = append(m.Warnings, fmt.Sprintf("Transmittal lists %s but no “%s” heading was found.", label, label))
			case !listed && present && sec != nil:
				m.Notes = append(m.Notes, fmt.Sprintf("“%s” found in the manuscript but not ticked on the transmittal; it is included anyway.", label))
			}
		}
	}
	check("front_matter", specFrontMatterKeys, m.Front())
	check("back_matter", specBackMatterKeys, m.Back())
}

// bookMapFromDOCX reads the DOCX and builds the map; spec may be nil.
func bookMapFromDOCX(docxPath string, spec map[string]any, bookTitle, bookAuthor string) (*BookMap, error) {
	paras, err := readDOCXParagraphs(docxPath)
	if err != nil {
		return nil, err
	}
	var names []string
	if spec != nil {
		names = untitledNamesFromSpec(spec)
	}
	m := buildBookMap(paras, names, bookTitle, bookAuthor)
	if spec != nil {
		m.crossCheckSpec(spec)
	}
	return m, nil
}

// --- DOCX reading ----------------------------------------------------------------

// readDOCXParagraphs returns the non-empty paragraphs of word/document.xml
// with normalized style names (resolved through word/styles.xml) and hard
// page-break flags. Empty paragraphs are skipped (as pandoc skips them) but a
// page break they carry is passed on to the next non-empty paragraph.
func readDOCXParagraphs(docxPath string) ([]docxPara, error) {
	zr, err := zip.OpenReader(docxPath)
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	var docFile, stylesFile *zip.File
	for _, f := range zr.File {
		switch filepath.ToSlash(f.Name) {
		case "word/document.xml":
			docFile = f
		case "word/styles.xml":
			stylesFile = f
		}
	}
	if docFile == nil {
		return nil, fmt.Errorf("word/document.xml not found in %s", filepath.Base(docxPath))
	}
	styleNames := map[string]string{}
	if stylesFile != nil {
		rc, err := stylesFile.Open()
		if err == nil {
			styleNames, _ = parseDOCXStyleNames(rc)
			_ = rc.Close()
		}
	}
	rc, err := docFile.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return parseDOCXParagraphs(rc, styleNames)
}

// parseDOCXStyleNames maps w:styleId → normalized w:name ("Heading1" → "heading1").
func parseDOCXStyleNames(r io.Reader) (map[string]string, error) {
	out := map[string]string{}
	dec := xml.NewDecoder(r)
	curID := ""
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return out, err
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch se.Name.Local {
		case "style":
			curID = xmlAttr(se, "styleId")
		case "name":
			if curID != "" {
				out[curID] = normStyleName(xmlAttr(se, "val"))
			}
		}
	}
	return out, nil
}

func xmlAttr(se xml.StartElement, local string) string {
	for _, a := range se.Attr {
		if a.Name.Local == local {
			return a.Value
		}
	}
	return ""
}

// parseDOCXParagraphs walks document.xml. It ignores headers/footers (not in
// this part) and treats table-cell paragraphs like any other.
func parseDOCXParagraphs(r io.Reader, styleNames map[string]string) ([]docxPara, error) {
	dec := xml.NewDecoder(r)
	var out []docxPara
	var cur *docxPara
	var text strings.Builder
	pendingBreak := false // page break carried into the next paragraph
	breakAfter := false   // a page break inside the current paragraph, after its text
	inPPr := false
	inT := false // inside w:t (the only element whose character data is text)
	depth := 0   // nesting of w:p (textboxes can nest paragraphs; we only take the outermost)

	flush := func() {
		if cur == nil {
			return
		}
		cur.Text = strings.TrimSpace(text.String())
		if cur.Text != "" || cur.HasImage {
			cur.PageBreakBefore = cur.PageBreakBefore || pendingBreak
			pendingBreak = breakAfter
			out = append(out, *cur)
		} else if cur.PageBreakBefore || breakAfter {
			// Empty paragraph that only carries a break: pass it on.
			pendingBreak = true
		}
		breakAfter = false
		cur = nil
		text.Reset()
	}

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("parse document.xml: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "p":
				depth++
				if depth == 1 {
					cur = &docxPara{Style: "normal"}
				}
			case "pPr":
				inPPr = depth == 1
			case "pStyle":
				if inPPr && cur != nil {
					id := xmlAttr(t, "val")
					if n, ok := styleNames[id]; ok && n != "" {
						cur.Style = n
					} else {
						cur.Style = normStyleName(id)
					}
				}
			case "pageBreakBefore":
				if inPPr && cur != nil && xmlAttr(t, "val") != "0" && xmlAttr(t, "val") != "false" {
					cur.PageBreakBefore = true
				}
			case "sectPr":
				// A section break inside a paragraph's properties ends a section
				// at this paragraph; the next paragraph starts on a new page
				// (nextPage/oddPage/evenPage; "continuous" is checked below).
				if inPPr {
					breakAfter = true
				}
			case "type":
				if inPPr && xmlAttr(t, "val") == "continuous" {
					breakAfter = false
				}
			case "br":
				if depth == 1 && cur != nil {
					switch xmlAttr(t, "type") {
					case "page":
						if strings.TrimSpace(text.String()) == "" && !cur.HasImage {
							cur.PageBreakBefore = true
						} else {
							breakAfter = true
						}
					case "", "textWrapping":
						text.WriteByte(' ') // manual line break inside a paragraph
					}
				}
			case "lastRenderedPageBreak":
				// Word's soft cache of where a page happened to end; not a hard break.
			case "drawing", "pict":
				if depth == 1 && cur != nil {
					cur.HasImage = true
				}
			case "tab":
				if depth == 1 && cur != nil {
					text.WriteByte(' ')
				}
			case "t":
				inT = true
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "t":
				inT = false
			case "pPr":
				inPPr = false
			case "p":
				if depth == 1 {
					flush()
				}
				depth--
			}
		case xml.CharData:
			if depth == 1 && cur != nil && inT {
				text.Write(t)
			}
		}
	}
	flush()
	return out, nil
}
