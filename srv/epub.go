package srv

import (
	"archive/zip"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"srv.exe.dev/db/dbgen"
)

// handleGenerateEPUB converts a book's source docx to EPUB using pandoc + spec settings.
func (s *Server) handleGenerateEPUB(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	bid, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}

	q := dbgen.New(s.DB)
	book, err := q.GetBook(r.Context(), bid)
	if err != nil {
		jsonErr(w, "not found", 404)
		return
	}

	if book.SourceData == nil || len(book.SourceData) == 0 {
		jsonErr(w, "no source file", 400)
		return
	}

	go s.runEPUBGeneration(bid, book)

	jsonOK(w, map[string]string{"status": "generating_epub"})
}

func (s *Server) runEPUBGeneration(bid int64, book dbgen.Book) {
	start := time.Now()
	q := dbgen.New(s.DB)
	ctx := context.Background()

	tmpDir, err := os.MkdirTemp("", fmt.Sprintf("epub-%d-*", bid))
	if err != nil {
		slog.Error("epub: create temp dir", "id", bid, "err", err)
		return
	}
	defer os.RemoveAll(tmpDir)

	// Write source docx
	docxPath := filepath.Join(tmpDir, "input.docx")
	if err := os.WriteFile(docxPath, book.SourceData, 0644); err != nil {
		slog.Error("epub: write docx", "id", bid, "err", err)
		return
	}

	slog.Info("epub generation starting", "id", bid, "title", book.Title)

	// Apply pending corrections to the source docx before pandoc runs, so the
	// EPUB inherits the same patches the PDF pipeline does. Also patch
	// in-memory metadata — pandoc receives author/title via --metadata flags
	// pulled from the books table, NOT from docx content, so without this the
	// EPUB title page renders the unpatched values.
	var correctionsSnapshot string
	var correctionPairs []correctionPair
	if book.ProjectID.Valid {
		correctionsSnapshot, correctionPairs = s.applyCorrectionsIfAny(ctx, bid, book.ProjectID.Int64, tmpDir, docxPath)
		if len(correctionPairs) > 0 {
			book.Title = applyPairsToString(book.Title, correctionPairs)
			book.Author = applyPairsToString(book.Author, correctionPairs)
		}
	}

	// Load spec if project is linked
	var spec epubSpec
	spec.Title = book.Title
	spec.Author = book.Author
	spec.Language = "en"
	spec.TOCDepth = 2

	var specSnapshot string
	if book.ProjectID.Valid {
		dbSpec, err := q.GetBookSpec(ctx, book.ProjectID.Int64)
		if err == nil {
			spec = parseEPUBSpec(dbSpec.Data, book)
			specSnapshot = dbSpec.Data
			// Spec JSON can override Title/Author with its own (potentially
			// unpatched) values. Re-apply corrections so the EPUB metadata
			// pandoc receives is consistent with the patched docx body.
			if len(correctionPairs) > 0 {
				spec.Title = applyPairsToString(spec.Title, correctionPairs)
				spec.Author = applyPairsToString(spec.Author, correctionPairs)
				spec.Subject = applyPairsToString(spec.Subject, correctionPairs)
				spec.Description = applyPairsToString(spec.Description, correctionPairs)
			}

			// Write cover image if available
			coverRow, err := q.GetBookSpecCover(ctx, book.ProjectID.Int64)
			if err == nil && coverRow.CoverData != nil && len(coverRow.CoverData) > 0 {
				ext := ".jpg"
				if coverRow.CoverType == "image/png" {
					ext = ".png"
				}
				coverPath := filepath.Join(tmpDir, "cover"+ext)
				if err := os.WriteFile(coverPath, coverRow.CoverData, 0644); err == nil {
					spec.CoverImage = coverPath
				}
			}
		}
	}

	// Write custom CSS if provided
	var cssPath string
	css := spec.buildCSS()
	if css != "" {
		cssPath = filepath.Join(tmpDir, "custom.css")
		os.WriteFile(cssPath, []byte(css), 0644)
	}

	// Build pandoc command
	epubPath := filepath.Join(tmpDir, "output.epub")
	args := []string{
		"--from=docx+styles",
		"--to=epub3",
		fmt.Sprintf("--metadata=title:%s", spec.Title),
		fmt.Sprintf("--metadata=author:%s", spec.Author),
		fmt.Sprintf("--metadata=lang:%s", spec.Language),
		fmt.Sprintf("--toc-depth=%d", spec.TOCDepth),
		"--toc",
		"-o", epubPath,
	}

	if spec.CoverImage != "" {
		args = append(args, "--epub-cover-image="+spec.CoverImage)
	}
	if cssPath != "" {
		args = append(args, "--css="+cssPath)
	}
	if spec.Subject != "" {
		args = append(args, fmt.Sprintf("--metadata=subject:%s", spec.Subject))
	}
	if spec.Description != "" {
		args = append(args, fmt.Sprintf("--metadata=description:%s", spec.Description))
	}

	args = append(args, docxPath)

	cmd := exec.Command("pandoc", args...)
	cmd.Dir = tmpDir
	if out, err := cmd.CombinedOutput(); err != nil {
		slog.Error("epub: pandoc failed", "id", bid, "err", err, "output", string(out))
		return
	}

	// Read generated EPUB
	epubData, err := os.ReadFile(epubPath)
	if err != nil {
		slog.Error("epub: read output", "id", bid, "err", err)
		return
	}

	// TRK-DEV-009: inject per-chapter author bylines if configured.
	if len(spec.Chapters) > 0 {
		patched, perr := injectChapterAuthors(epubData, spec.Chapters)
		if perr != nil {
			slog.Warn("epub: chapter byline injection failed; shipping un-bylined EPUB",
				"id", bid, "err", perr)
		} else {
			epubData = patched
		}
	}

	// Store in DB
	if _, err := q.CreateBookOutput(ctx, dbgen.CreateBookOutputParams{
		BookID:              bid,
		OutputFormat:        "epub",
		OutputData:          epubData,
		SourceFilename:      book.SourceFilename,
		SpecSnapshot:        nullStringFrom(specSnapshot),
		CorrectionsSnapshot: nullStringFrom(correctionsSnapshot),
	}); err != nil {
		slog.Error("epub: persist history", "id", bid, "err", err)
		return
	}
	if err := q.UpdateBookEPUB(ctx, dbgen.UpdateBookEPUBParams{
		EpubData: epubData,
		ID:       bid,
	}); err != nil {
		slog.Error("epub: store", "id", bid, "err", err)
		return
	}

	slog.Info("epub generation complete", "id", bid, "title", book.Title,
		"epub_size", len(epubData), "elapsed", time.Since(start))
}

type epubSpec struct {
	Title        string
	Author       string
	Language     string
	Subject      string
	Description  string
	TOCDepth     int
	CoverImage   string
	ChapterBreak string
	SectionBreak string
	BodyFontSize string
	EmbedFonts   bool
	CustomCSS    string
	Chapters     []epubChapter
}

// epubChapter is a per-chapter override read from data.epub.chapters.
// Author is the per-chapter byline injected after the <h1>; Title and File
// are metadata for the admin UI / future source-file matching.
type epubChapter struct {
	Title  string
	Author string
	File   string
}

func parseEPUBSpec(specJSON string, book dbgen.Book) epubSpec {
	var data map[string]any
	json.Unmarshal([]byte(specJSON), &data)

	spec := epubSpec{
		Title:    book.Title,
		Author:   book.Author,
		Language: "en",
		TOCDepth: 2,
	}

	// Pull from metadata section
	if meta, ok := data["metadata"].(map[string]any); ok {
		if v, ok := meta["title"].(string); ok && v != "" {
			spec.Title = v
		}
		if v, ok := meta["author"].(string); ok && v != "" {
			spec.Author = v
		}
	}

	// Pull from epub section
	if epub, ok := data["epub"].(map[string]any); ok {
		if v, ok := epub["language"].(string); ok && v != "" {
			spec.Language = v
		}
		if v, ok := epub["subject"].(string); ok {
			spec.Subject = v
		}
		if v, ok := epub["description"].(string); ok {
			spec.Description = v
		}
		if v, ok := epub["toc_depth"].(float64); ok && v > 0 {
			spec.TOCDepth = int(v)
		}
		if v, ok := epub["chapter_break"].(string); ok {
			spec.ChapterBreak = v
		}
		if v, ok := epub["section_break"].(string); ok {
			spec.SectionBreak = v
		}
		if v, ok := epub["body_font_size"].(string); ok {
			spec.BodyFontSize = v
		}
		if v, ok := epub["embed_fonts"].(bool); ok {
			spec.EmbedFonts = v
		}
		if v, ok := epub["custom_css"].(string); ok {
			spec.CustomCSS = v
		}
		if raw, ok := epub["chapters"].([]any); ok {
			for _, item := range raw {
				m, ok := item.(map[string]any)
				if !ok {
					continue
				}
				ch := epubChapter{}
				if v, ok := m["title"].(string); ok {
					ch.Title = v
				}
				if v, ok := m["author"].(string); ok {
					ch.Author = v
				}
				if v, ok := m["file"].(string); ok {
					ch.File = v
				}
				// Only keep rows that contribute something (author or title
				// — a row with neither is a stub from the UI).
				if strings.TrimSpace(ch.Author) != "" || strings.TrimSpace(ch.Title) != "" {
					spec.Chapters = append(spec.Chapters, ch)
				}
			}
		}
	}

	return spec
}

// h1OpenRE matches the first <h1...>...</h1> tag in an XHTML chapter file.
// Pandoc emits chapter headings as <h1 class="...">Title</h1>; we insert
// the byline as a sibling immediately after the closing </h1>.
var h1OpenRE = regexp.MustCompile(`(?s)(<h1\b[^>]*>.*?</h1>)`)

// injectChapterAuthors rewrites the EPUB ZIP to add a <p class="chapter-author">
// paragraph after the first <h1> of each chapter XHTML file, matched in spine
// order. Files without an <h1> (nav, title page, cover) are passed through
// untouched. If there are more chapter files than configured authors, the
// extras are left alone; if fewer, the surplus authors are silently dropped.
func injectChapterAuthors(epubData []byte, chapters []epubChapter) ([]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(epubData), int64(len(epubData)))
	if err != nil {
		return nil, fmt.Errorf("open epub: %w", err)
	}

	// First pass: collect entry order and identify chapter files (xhtml with <h1>).
	type entry struct {
		f       *zip.File
		body    []byte
		isChap  bool
	}
	entries := make([]*entry, 0, len(zr.File))
	chapIdxs := make([]int, 0, len(chapters))
	for i, f := range zr.File {
		e := &entry{f: f}
		entries = append(entries, e)
		name := strings.ToLower(f.Name)
		if !strings.HasSuffix(name, ".xhtml") && !strings.HasSuffix(name, ".html") {
			continue
		}
		// Skip nav / TOC files — pandoc names them nav.xhtml.
		base := filepath.Base(name)
		if base == "nav.xhtml" || base == "toc.xhtml" || base == "title_page.xhtml" || base == "cover.xhtml" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", f.Name, err)
		}
		buf, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", f.Name, err)
		}
		e.body = buf
		if h1OpenRE.Match(buf) {
			e.isChap = true
			chapIdxs = append(chapIdxs, i)
		}
	}

	// Inject bylines into chapter entries, matched by order.
	n := len(chapters)
	if len(chapIdxs) < n {
		n = len(chapIdxs)
	}
	for k := 0; k < n; k++ {
		author := strings.TrimSpace(chapters[k].Author)
		if author == "" {
			continue
		}
		e := entries[chapIdxs[k]]
		loc := h1OpenRE.FindIndex(e.body)
		if loc == nil {
			continue
		}
		byline := fmt.Sprintf(`<p class="chapter-author">%s</p>`, escapeXMLText(author))
		// Splice the byline immediately after the first </h1>.
		var nb bytes.Buffer
		nb.Grow(len(e.body) + len(byline))
		nb.Write(e.body[:loc[1]])
		nb.WriteString(byline)
		nb.Write(e.body[loc[1]:])
		e.body = nb.Bytes()
	}

	// Re-zip. Preserve original file mode / order; for "mimetype" (which must
	// be the first entry and stored, not deflated), keep Store method.
	var out bytes.Buffer
	zw := zip.NewWriter(&out)
	for _, e := range entries {
		fh := &zip.FileHeader{
			Name:     e.f.Name,
			Method:   e.f.Method,
			Modified: e.f.Modified,
		}
		// EPUB mimetype must be stored uncompressed.
		if e.f.Name == "mimetype" {
			fh.Method = zip.Store
		}
		w, err := zw.CreateHeader(fh)
		if err != nil {
			return nil, fmt.Errorf("create %s: %w", e.f.Name, err)
		}
		if e.body != nil {
			if _, err := w.Write(e.body); err != nil {
				return nil, fmt.Errorf("write %s: %w", e.f.Name, err)
			}
		} else {
			rc, err := e.f.Open()
			if err != nil {
				return nil, fmt.Errorf("reopen %s: %w", e.f.Name, err)
			}
			if _, err := io.Copy(w, rc); err != nil {
				rc.Close()
				return nil, fmt.Errorf("copy %s: %w", e.f.Name, err)
			}
			rc.Close()
		}
	}
	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("close zip: %w", err)
	}
	return out.Bytes(), nil
}

func escapeXMLText(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

// buildCSS generates the EPUB custom stylesheet from spec settings.
func (s *epubSpec) buildCSS() string {
	var parts []string

	if s.BodyFontSize != "" && s.BodyFontSize != "inherit" {
		parts = append(parts, fmt.Sprintf("body { font-size: %s; }", s.BodyFontSize))
	}

	// Chapter break styling
	switch s.ChapterBreak {
	case "rule":
		parts = append(parts, "h1 { border-bottom: 1px solid #ccc; padding-bottom: 0.5em; }")
	case "ornament":
		parts = append(parts, "h1::before { content: '\\2766'; display: block; text-align: center; margin-bottom: 1em; font-size: 1.5em; }")
	}

	// Section break ornaments
	switch s.SectionBreak {
	case "breve":
		parts = append(parts, "hr { border: none; text-align: center; } hr::after { content: '\\02D8'; font-size: 1.5em; }")
	case "asterism":
		parts = append(parts, "hr { border: none; text-align: center; } hr::after { content: '\\2042'; font-size: 1.5em; }")
	case "dinkus":
		parts = append(parts, "hr { border: none; text-align: center; } hr::after { content: '* * *'; letter-spacing: 0.5em; }")
	case "blank":
		parts = append(parts, "hr { border: none; margin: 1.5em 0; }")
	}

	// Per-chapter author byline (TRK-DEV-009). Only emit when configured
	// so single-author EPUBs are byte-identical to today's output.
	if len(s.Chapters) > 0 {
		parts = append(parts, ".chapter-author { text-align: center; font-style: italic; font-size: 0.95em; margin: 0.25em 0 1.25em; color: #555; }")
	}

	// User's custom CSS
	if s.CustomCSS != "" {
		parts = append(parts, "\n/* User custom CSS */")
		parts = append(parts, s.CustomCSS)
	}

	if len(parts) == 0 {
		return ""
	}

	return "/* Generated EPUB styles */\n" + joinLines(parts)
}

func joinLines(ss []string) string {
	out := ""
	for _, s := range ss {
		out += s + "\n"
	}
	return out
}

// Ensure sql import is used
var _ = sql.ErrNoRows
