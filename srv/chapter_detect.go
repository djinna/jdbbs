package srv

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"srv.exe.dev/db/dbgen"
)

// TRK-DEV-012 Phase C — chapter auto-detection on upload.
//
// When a DOCX is uploaded for a project-linked book, we run pandoc once to
// emit the document AST as JSON (`--from=docx+styles -t json`) and walk it for
// level-1 headings + adjacent author bylines. Detected chapters are written as
// *suggestions* into spec.chapters_suggested[] — never into spec.chapters[],
// which is user-owned. The admin Anthology card surfaces them behind an
// "Apply" button so the cost of a wrong guess is one click, not a wrong byline
// shipped to print.
//
// Approach: pandoc → JSON AST walked in Go (not a Lua filter). Detection runs
// at upload time, where the pipeline doesn't otherwise invoke pandoc, so the
// "reuse the existing pandoc call" argument for a Lua pre-pass doesn't apply.
// The JSON AST is parsed natively in Go, is fully unit-testable with hand-built
// fixtures (no DOCX needed), and adds no new runtime dependency.

// detectedChapter is one suggested chapter. Author is "" for an h1 with no
// detectable byline (e.g. an introduction or a front-matter heading) — that's
// a title-only suggestion the user reviews, not a detection failure.
type detectedChapter struct {
	Title  string `json:"title"`
	Author string `json:"author"`
}

// --- Pandoc JSON AST (minimal) ---------------------------------------------
//
// We only model what the walk needs. Every block/inline is {t, c}; the shape
// of c depends on t. We keep c as RawMessage and decode lazily per node type.

type pandocAST struct {
	Blocks []astBlock `json:"blocks"`
}

type astBlock struct {
	T string          `json:"t"`
	C json.RawMessage `json:"c"`
}

type astInline struct {
	T string          `json:"t"`
	C json.RawMessage `json:"c"`
}

// --- Byline heuristics ------------------------------------------------------

// byPattern matches a byline line: "By <Capitalized…>" / "by <Capitalized…>".
// The capital-letter requirement is what keeps prose openers like
// "By the time he arrived…" from false-matching ("the" is lowercase).
var byPattern = regexp.MustCompile(`^[Bb]y\s+(\p{Lu}.*)$`)

// bylineMaxLen caps a byline candidate's length. Real bylines are short
// ("By Spencer Nitkey" ≈ 17 chars); chapter-opener verse epigraphs are full
// lines/sentences. This is the primary defense against an italic epigraph
// masquerading as an italic byline.
const bylineMaxLen = 60

// authorStyleNames are the normalized Word custom-style names that mark a
// paragraph as an author byline. Highest-confidence signal when present.
var authorStyleNames = map[string]bool{
	"author":        true,
	"authorname":    true,
	"byline":        true,
	"chapterauthor": true,
	"contributor":   true,
}

func normStyleName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.Map(func(r rune) rune {
		if r == ' ' || r == '-' || r == '_' {
			return -1
		}
		return r
	}, s)
	return s
}

// inlinesText flattens an inline array to plain text. Str → its text;
// Space/SoftBreak/LineBreak → a single space; styled/wrapping inlines recurse.
func inlinesText(raw json.RawMessage) string {
	var arr []astInline
	if err := json.Unmarshal(raw, &arr); err != nil {
		return ""
	}
	var b strings.Builder
	for _, in := range arr {
		b.WriteString(inlineText(in))
	}
	return b.String()
}

func inlineText(in astInline) string {
	switch in.T {
	case "Str":
		var s string
		_ = json.Unmarshal(in.C, &s)
		return s
	case "Space", "SoftBreak", "LineBreak":
		return " "
	case "Emph", "Strong", "Underline", "Strikeout", "Superscript", "Subscript", "SmallCaps":
		// c is an inline array.
		return inlinesText(in.C)
	case "Quoted":
		// c = [QuoteType, [inlines]]
		return nthInlineArray(in.C, 1)
	case "Span", "Cite":
		// Span c = [attr, [inlines]]; Cite c = [citations, [inlines]]
		return nthInlineArray(in.C, 1)
	case "Link", "Image":
		// c = [attr, [inlines], target]
		return nthInlineArray(in.C, 1)
	case "Code":
		// c = [attr, string]
		var parts []json.RawMessage
		if json.Unmarshal(in.C, &parts) == nil && len(parts) == 2 {
			var s string
			_ = json.Unmarshal(parts[1], &s)
			return s
		}
		return ""
	default:
		return ""
	}
}

// nthInlineArray decodes c as a heterogeneous tuple and flattens the inline
// array at index idx.
func nthInlineArray(raw json.RawMessage, idx int) string {
	var parts []json.RawMessage
	if err := json.Unmarshal(raw, &parts); err != nil || idx >= len(parts) {
		return ""
	}
	return inlinesText(parts[idx])
}

// hasLineBreak reports whether an inline tree contains a soft/hard line break,
// recursing through wrapping inlines (Emph, Span, …). Multi-line candidates
// (e.g. a verse epigraph, which is typically one Emph wrapping SoftBreaks) are
// rejected as bylines. The length cap is the primary defense; this catches the
// short-but-multi-line case the cap would miss.
func hasLineBreak(raw json.RawMessage) bool {
	var arr []astInline
	if json.Unmarshal(raw, &arr) != nil {
		return false
	}
	for _, in := range arr {
		switch in.T {
		case "SoftBreak", "LineBreak":
			return true
		case "Emph", "Strong", "Underline", "Strikeout", "Superscript", "Subscript", "SmallCaps":
			if hasLineBreak(in.C) {
				return true
			}
		case "Quoted", "Span", "Cite", "Link", "Image":
			if nthHasLineBreak(in.C, 1) {
				return true
			}
		}
	}
	return false
}

func nthHasLineBreak(raw json.RawMessage, idx int) bool {
	var parts []json.RawMessage
	if err := json.Unmarshal(raw, &parts); err != nil || idx >= len(parts) {
		return false
	}
	return hasLineBreak(parts[idx])
}

// headerLevelTitle returns (level, title-text, true) for a Header block.
func headerLevelTitle(b astBlock) (int, string, bool) {
	if b.T != "Header" {
		return 0, "", false
	}
	// c = [level, [id, classes, kvs], [inlines]]
	var parts []json.RawMessage
	if json.Unmarshal(b.C, &parts) != nil || len(parts) != 3 {
		return 0, "", false
	}
	var level int
	if json.Unmarshal(parts[0], &level) != nil {
		return 0, "", false
	}
	return level, inlinesText(parts[2]), true
}

// bylineFromBy extracts the name from a "By <Name>" line, applying the length
// cap and an em-/en-dash rejection (a dash signals an epigraph attribution,
// not a byline). Returns ("", false) when the text is not a byline.
func bylineFromBy(text string) (string, bool) {
	t := strings.TrimSpace(text)
	if t == "" || len(t) > bylineMaxLen || strings.ContainsAny(t, "—–") {
		return "", false
	}
	m := byPattern.FindStringSubmatch(t)
	if m == nil {
		return "", false
	}
	return strings.TrimSpace(m[1]), true
}

// emphByline returns the text of a paragraph that is *entirely* italic (a
// run of Emph inlines plus connective whitespace), subject to the byline
// guards: short, single-line, no dash. This is the lowest-confidence signal.
func emphByline(b astBlock) (string, bool) {
	if b.T != "Para" && b.T != "Plain" {
		return "", false
	}
	if hasLineBreak(b.C) {
		return "", false
	}
	var arr []astInline
	if json.Unmarshal(b.C, &arr) != nil {
		return "", false
	}
	sawEmph := false
	for _, in := range arr {
		switch in.T {
		case "Emph":
			sawEmph = true
		case "Space", "SoftBreak", "LineBreak":
			// connective whitespace is allowed
		default:
			// any non-italic, non-space content → not a pure-italic byline
			return "", false
		}
	}
	if !sawEmph {
		return "", false
	}
	text := strings.TrimSpace(inlinesText(b.C))
	if text == "" || len(text) > bylineMaxLen || strings.ContainsAny(text, "—–") {
		return "", false
	}
	return text, true
}

// bylineCandidate is a Para/Plain block surfaced for byline inspection, tagged
// with the Word custom-style of its nearest wrapping Div (if any).
type bylineCandidate struct {
	style string
	block astBlock
}

// divStyleAndInner pulls a Div's custom-style and its inner blocks.
// c = [attr, [blocks]] ; attr = [id, [classes], [[k,v]...]].
func divStyleAndInner(b astBlock) (string, []astBlock) {
	var parts []json.RawMessage
	if json.Unmarshal(b.C, &parts) != nil || len(parts) != 2 {
		return "", nil
	}
	style := ""
	var attr []json.RawMessage
	if json.Unmarshal(parts[0], &attr) == nil && len(attr) == 3 {
		var kvs [][]string
		_ = json.Unmarshal(attr[2], &kvs)
		for _, kv := range kvs {
			if len(kv) == 2 && strings.EqualFold(kv[0], "custom-style") {
				style = kv[1]
			}
		}
	}
	var inner []astBlock
	_ = json.Unmarshal(parts[1], &inner)
	return style, inner
}

func isAuthorStyle(style string) bool {
	return authorStyleNames[normStyleName(style)]
}

// collectCandidates appends Para/Plain byline candidates from b, descending
// through Divs and carrying the nearest custom-style down. This is what makes
// detection work on real DOCX input: `pandoc --from=docx+styles` wraps every
// paragraph in a Div tagged with its Word paragraph style ("First Paragraph",
// "Body Text", "Author", …), so byline paragraphs are never top-level Paras.
// Blank paragraphs are skipped so they don't consume the window.
func collectCandidates(b astBlock, style string, out *[]bylineCandidate) {
	switch b.T {
	case "Para", "Plain":
		if strings.TrimSpace(inlinesText(b.C)) == "" {
			return
		}
		*out = append(*out, bylineCandidate{style: style, block: b})
	case "Div":
		s, inner := divStyleAndInner(b)
		ds := style
		if s != "" {
			ds = s
		}
		for _, ib := range inner {
			collectCandidates(ib, ds, out)
		}
	}
	// Other block types (Header, BlockQuote, lists, tables, …) are not byline
	// candidates and are ignored here.
}

// bylineWindow is how many content paragraphs after an h1 we inspect for a
// byline. The byline (or a verse epigraph that precedes it) sits right under
// the heading; a small window keeps us anchored to the chapter opener and
// avoids matching "By <Name>" deep in body prose.
const bylineWindow = 3

// scanByline returns the best byline among the first few content paragraphs
// after an h1 (start = h1 index + 1), stopping at the next h1. Selection is by
// *confidence*, not position: Word custom-style (Author/Byline/…) first, then
// "By <Name>", then short all-italic. Confidence-ordering is what lets a real
// "By Spencer Nitkey" line win over a preceding italic verse epigraph in the
// same opener.
func scanByline(blocks []astBlock, start int) string {
	var cands []bylineCandidate
	for j := start; j < len(blocks) && len(cands) < bylineWindow; j++ {
		b := blocks[j]
		if lvl, _, ok := headerLevelTitle(b); ok && lvl == 1 {
			break // next chapter — stop
		}
		collectCandidates(b, "", &cands)
	}
	if len(cands) > bylineWindow {
		cands = cands[:bylineWindow]
	}

	// 1) Word custom-style byline (highest confidence).
	for _, c := range cands {
		if isAuthorStyle(c.style) {
			text := strings.TrimSpace(inlinesText(c.block.C))
			if text == "" {
				continue
			}
			if name, ok := bylineFromBy(text); ok {
				return name
			}
			return text
		}
	}
	// 2) "By <Name>" prefix.
	for _, c := range cands {
		if name, ok := bylineFromBy(strings.TrimSpace(inlinesText(c.block.C))); ok {
			return name
		}
	}
	// 3) Short all-italic paragraph (lowest confidence).
	for _, c := range cands {
		if name, ok := emphByline(c.block); ok {
			return name
		}
	}
	return ""
}

// detectChaptersFromAST walks a pandoc JSON AST and returns one suggestion per
// level-1 heading, in source order. h1s without a detectable byline yield a
// title-only suggestion (author=""). A document with no level-1 headings (e.g.
// Twitter Years) returns an empty slice — no suggestions, no UI noise.
func detectChaptersFromAST(astJSON []byte) ([]detectedChapter, error) {
	var doc pandocAST
	if err := json.Unmarshal(astJSON, &doc); err != nil {
		return nil, fmt.Errorf("parse pandoc AST: %w", err)
	}
	out := []detectedChapter{}
	for i := range doc.Blocks {
		level, title, ok := headerLevelTitle(doc.Blocks[i])
		if !ok || level != 1 {
			continue
		}
		title = strings.TrimSpace(title)
		if title == "" {
			continue
		}
		out = append(out, detectedChapter{
			Title:  title,
			Author: scanByline(doc.Blocks, i+1),
		})
	}
	return out, nil
}

// runPandocToAST converts DOCX bytes to a pandoc JSON AST. Stdout carries the
// JSON; stderr carries warnings (which we surface only on failure).
func runPandocToAST(ctx context.Context, docx []byte) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "detect-chapters-*")
	if err != nil {
		return nil, fmt.Errorf("temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	docxPath := filepath.Join(tmpDir, "input.docx")
	if err := os.WriteFile(docxPath, docx, 0644); err != nil {
		return nil, fmt.Errorf("write docx: %w", err)
	}

	cmd := exec.CommandContext(ctx, "pandoc", "--from=docx+styles", docxPath, "-t", "json")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("pandoc -t json: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}

// detectAndStoreChapterSuggestions runs the full pass for a book: pandoc → AST
// → detect → persist into the project's spec.chapters_suggested[]. Returns the
// detected chapters. It is a no-op (nil, nil) for books with no project, since
// suggestions live on the per-project spec and have nowhere else to go.
func (s *Server) detectAndStoreChapterSuggestions(ctx context.Context, book dbgen.Book) ([]detectedChapter, error) {
	if !book.ProjectID.Valid {
		return nil, nil
	}
	if len(book.SourceData) == 0 {
		return nil, fmt.Errorf("book %d has no source data", book.ID)
	}
	ast, err := runPandocToAST(ctx, book.SourceData)
	if err != nil {
		return nil, err
	}
	chs, err := detectChaptersFromAST(ast)
	if err != nil {
		return nil, err
	}
	if err := s.storeChapterSuggestions(ctx, book.ProjectID.Int64, chs); err != nil {
		return nil, err
	}
	return chs, nil
}

// storeChapterSuggestions writes chs into spec.chapters_suggested[] using an
// atomic SQLite json_set so it never read-modify-writes the surrounding spec
// (and so it never touches the user-owned spec.chapters[]). When there are no
// detected chapters we only clear a *pre-existing* suggestion list — we don't
// create a spec row just to store an empty array (a single-author upload to a
// fresh project shouldn't materialize a spec early).
func (s *Server) storeChapterSuggestions(ctx context.Context, projectID int64, chs []detectedChapter) error {
	q := dbgen.New(s.DB)
	_, specErr := q.GetBookSpec(ctx, projectID)
	specMissing := specErr == sql.ErrNoRows
	if specErr != nil && !specMissing {
		return specErr
	}

	if len(chs) == 0 {
		if specMissing {
			return nil // nothing to clear, nothing to create
		}
	} else if specMissing {
		if _, err := q.UpsertBookSpec(ctx, dbgen.UpsertBookSpecParams{
			ProjectID: projectID,
			Data:      defaultSpecData(),
		}); err != nil {
			return fmt.Errorf("create spec for suggestions: %w", err)
		}
	}

	if chs == nil {
		chs = []detectedChapter{}
	}
	payload, err := json.Marshal(chs)
	if err != nil {
		return err
	}
	_, err = s.DB.ExecContext(ctx,
		`UPDATE book_specs
		    SET data = json_set(data, '$.chapters_suggested', json(?)),
		        updated_at = CURRENT_TIMESTAMP
		  WHERE project_id = ?`,
		string(payload), projectID)
	return err
}

// detectChaptersAsync runs detection in the background after an upload. It must
// never fail the upload, so errors are logged and swallowed; the book status is
// left untouched. A 90s ceiling guards against a pandoc hang on a pathological
// DOCX.
//
// Failures go to slog ONLY — the prodcal journalctl stream we use for
// diagnosis — and deliberately not to the project journal: that log feeds
// client digest emails (srv/client_digest_email.go pulls all entry types with
// no filter), and a parse failure is operator info, not client-facing project
// activity. The user-facing signal is the empty Anthology card plus the
// "Re-scan manuscript" button, which already prompt manual entry.
func (s *Server) detectChaptersAsync(book dbgen.Book) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	chs, err := s.detectAndStoreChapterSuggestions(ctx, book)
	if err != nil {
		slog.Warn("chapter auto-detect failed; upload unaffected",
			"book_id", book.ID, "project_id", book.ProjectID.Int64,
			"source", book.SourceFilename, "err", err)
		return
	}
	slog.Info("chapter auto-detect complete",
		"book_id", book.ID, "project_id", book.ProjectID.Int64, "detected", len(chs))
}
