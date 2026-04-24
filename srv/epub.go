package srv

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

	// Load spec if project is linked
	var spec epubSpec
	spec.Title = book.Title
	spec.Author = book.Author
	spec.Language = "en"
	spec.TOCDepth = 2

	if book.ProjectID.Valid {
		dbSpec, err := q.GetBookSpec(ctx, book.ProjectID.Int64)
		if err == nil {
			spec = parseEPUBSpec(dbSpec.Data, book)

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

	// Store in DB
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
	}

	return spec
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
