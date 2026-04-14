package srv

import (
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

const (
	maxUploadSize   = 50 << 20 // 50 MB
	bookProdRoot    = "/home/exedev/book-production"
	typstFilter     = bookProdRoot + "/scripts/docx-to-typst-enhanced.lua"
	seriesTemplate  = bookProdRoot + "/templates/series-template.typ"
	fontsDir        = bookProdRoot + "/fonts"
)

// handleListBooks returns all books (without blob data).
func (s *Server) handleListBooks(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	q := dbgen.New(s.DB)
	books, err := q.ListBooks(r.Context())
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	if books == nil {
		books = []dbgen.ListBooksRow{}
	}
	jsonOK(w, books)
}

// handleUploadBook accepts multipart form: file + title + author + series.
func (s *Server) handleUploadBook(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		jsonErr(w, "file too large or bad form", 400)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	author := strings.TrimSpace(r.FormValue("author"))
	series := strings.TrimSpace(r.FormValue("series"))
	projectIDStr := strings.TrimSpace(r.FormValue("project_id"))
	if title == "" || author == "" {
		jsonErr(w, "title and author required", 400)
		return
	}

	var projectID sql.NullInt64
	if projectIDStr != "" {
		pid, err := strconv.ParseInt(projectIDStr, 10, 64)
		if err == nil {
			projectID = sql.NullInt64{Int64: pid, Valid: true}
		}
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		jsonErr(w, "file required", 400)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		jsonErr(w, "read error", 500)
		return
	}

	q := dbgen.New(s.DB)
	book, err := q.CreateBook(r.Context(), dbgen.CreateBookParams{
		Title:          title,
		Author:         author,
		Series:         series,
		SourceFilename: header.Filename,
		SourceData:     data,
		ProjectID:      projectID,
	})
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}

	w.WriteHeader(201)
	jsonOK(w, map[string]any{
		"id":     book.ID,
		"title":  book.Title,
		"author": book.Author,
		"status": book.Status,
	})
}

// handleConvertBook runs the docx → typst → PDF pipeline.
func (s *Server) handleConvertBook(w http.ResponseWriter, r *http.Request) {
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

	// Mark as converting
	_ = q.UpdateBookStatus(r.Context(), dbgen.UpdateBookStatusParams{
		Status: "converting", ID: bid,
	})

	// Run conversion in background
	go s.runConversion(bid, book)

	jsonOK(w, map[string]string{"status": "converting"})
}

func (s *Server) runConversion(bid int64, book dbgen.Book) {
	start := time.Now()
	q := dbgen.New(s.DB)

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", fmt.Sprintf("book-%d-*", bid))
	if err != nil {
		s.failConversion(bid, "create temp dir: "+err.Error())
		return
	}
	defer os.RemoveAll(tmpDir)

	// Write source docx
	docxPath := filepath.Join(tmpDir, "input.docx")
	if err := os.WriteFile(docxPath, book.SourceData, 0644); err != nil {
		s.failConversion(bid, "write docx: "+err.Error())
		return
	}

	slog.Info("book conversion starting", "id", bid, "title", book.Title)

	// Step 1: direct pandoc docx -> typst using the working book-production filter.
	typPath := filepath.Join(tmpDir, "book.typ")
	pandocCmd := exec.Command("pandoc",
		"--from=docx+styles",
		docxPath,
		"--lua-filter="+typstFilter,
		"--extract-media="+filepath.Join(tmpDir, "media"),
		"-t", "typst",
		"-o", typPath,
	)
	if out, err := pandocCmd.CombinedOutput(); err != nil {
		s.failConversion(bid, fmt.Sprintf("pandoc typst: %s\n%s", err, string(out)))
		return
	}

	// Step 2: replace placeholder header with real metadata and any spec-driven config.
	typData, err := os.ReadFile(typPath)
	if err != nil {
		s.failConversion(bid, "read generated typst: "+err.Error())
		return
	}

	configOverride := s.buildTypstConfig(bid, book)
	headerReplacement := fmt.Sprintf(`#import "%s": *
%s
#show: book.with(
  title: "%s",
  author: "%s",
`,
		seriesTemplate,
		configOverride,
		escapeTypstString(book.Title),
		escapeTypstString(book.Author),
	)
	if strings.TrimSpace(book.Series) != "" {
		headerReplacement += fmt.Sprintf("  subtitle: \"%s\",\n", escapeTypstString(book.Series))
	}
	headerReplacement += `)

`

	const generatedHeader = `#import "/templates/series-template.typ": *

#show: book.with(
  title: "TITLE",
  author: "AUTHOR",
)

`

	typText := strings.Replace(string(typData), generatedHeader, headerReplacement, 1)
	if typText == string(typData) {
		s.failConversion(bid, "direct typst header replacement failed: expected generated header not found")
		return
	}

	// Narrow cleanup for manuscript patterns that Pandoc/Typst adjacency can misparse.
	typText = literalTypstMentions(typText)
	typText = clampTypstInlineImages(typText)
	typText = strings.ReplaceAll(typText, ")#strong[", ") #strong[")
	typText = strings.ReplaceAll(typText, ")](", ")] (")
	typText = strings.ReplaceAll(typText, "\n/\n", "\n#poem[/]\n")

	if err := os.WriteFile(typPath, []byte(typText), 0644); err != nil {
		s.failConversion(bid, "write direct typst: "+err.Error())
		return
	}

	// Step 3: typst compile the generated full document.
	pdfPath := filepath.Join(tmpDir, "output.pdf")
	typstCmd := exec.Command("typst", "compile",
		"--root", "/",
		"--font-path", fontsDir,
		typPath,
		pdfPath,
	)
	typstCmd.Dir = tmpDir
	if out, err := typstCmd.CombinedOutput(); err != nil {
		s.failConversion(bid, fmt.Sprintf("typst: %s\n%s", err, string(out)))
		return
	}

	// Read generated PDF
	pdfData, err := os.ReadFile(pdfPath)
	if err != nil {
		s.failConversion(bid, "read pdf: "+err.Error())
		return
	}

	// Store PDF in DB
	ctx := context.Background()
	if _, err := q.CreateBookOutput(ctx, dbgen.CreateBookOutputParams{
		BookID:         bid,
		OutputFormat:   "pdf",
		OutputData:     pdfData,
		SourceFilename: book.SourceFilename,
	}); err != nil {
		s.failConversion(bid, "persist pdf history: "+err.Error())
		return
	}
	if err := q.UpdateBookPDF(ctx, dbgen.UpdateBookPDFParams{
		PdfData: pdfData,
		ID:      bid,
	}); err != nil {
		s.failConversion(bid, "store pdf: "+err.Error())
		return
	}

	slog.Info("book conversion complete", "id", bid, "title", book.Title,
		"pdf_size", len(pdfData),
		"elapsed", time.Since(start))
}

func (s *Server) failConversion(bid int64, msg string) {
	slog.Error("book conversion failed", "id", bid, "error", msg)
	q := dbgen.New(s.DB)
	ctx := context.Background()
	_ = q.UpdateBookStatus(ctx, dbgen.UpdateBookStatusParams{
		Status:   "error",
		ErrorMsg: msg,
		ID:       bid,
	})
}

// handleDownloadBook serves the PDF or EPUB for a book.
func (s *Server) handleDownloadBook(w http.ResponseWriter, r *http.Request) {
	bid, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}

	q := dbgen.New(s.DB)

	// Auth: look up book's project and require auth if set
	bookRef, err := q.GetBookProjectID(r.Context(), bid)
	if err != nil {
		jsonErr(w, "not found", 404)
		return
	}
	if bookRef.ProjectID.Valid {
		if !s.requireAuth(w, r, bookRef.ProjectID.Int64) {
			return
		}
	} else if !s.requireExeDevAdminAPI(w, r) {
		// Books without a project require admin access
		return
	}

	format := r.PathValue("format")

	switch format {
	case "pdf":
		row, err := q.GetBookPDF(r.Context(), bid)
		if err != nil {
			jsonErr(w, "not found", 404)
			return
		}
		if row.PdfData == nil {
			jsonErr(w, "PDF not generated yet", 404)
			return
		}
		filename := sanitizeFilename(row.Title) + ".pdf"
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
		w.Header().Set("Content-Length", strconv.Itoa(len(row.PdfData)))
		w.Write(row.PdfData)

	case "epub":
		row, err := q.GetBookEPUB(r.Context(), bid)
		if err != nil {
			jsonErr(w, "not found", 404)
			return
		}
		if row.EpubData == nil {
			jsonErr(w, "EPUB not generated yet", 404)
			return
		}
		filename := sanitizeFilename(row.Title) + ".epub"
		w.Header().Set("Content-Type", "application/epub+zip")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
		w.Header().Set("Content-Length", strconv.Itoa(len(row.EpubData)))
		w.Write(row.EpubData)

	default:
		jsonErr(w, "format must be pdf or epub", 400)
	}
}

// handleDeleteBook removes a book.
func (s *Server) handleDeleteBook(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	bid, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	q := dbgen.New(s.DB)
	if err := q.DeleteBook(r.Context(), bid); err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	jsonOK(w, map[string]string{"ok": "true"})
}

// handleLinkBookProject links/unlinks a book to a project.
func (s *Server) handleLinkBookProject(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	bid, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}

	var body struct {
		ProjectID *int64 `json:"project_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, "bad request", 400)
		return
	}

	var pid sql.NullInt64
	if body.ProjectID != nil {
		pid = sql.NullInt64{Int64: *body.ProjectID, Valid: true}
	}

	q := dbgen.New(s.DB)
	if err := q.UpdateBookProject(r.Context(), dbgen.UpdateBookProjectParams{
		ProjectID: pid,
		ID:        bid,
	}); err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}

	jsonOK(w, map[string]string{"ok": "true"})
}

// buildTypstConfig looks up the book_spec for the linked project
// and generates Typst config override lines. Returns empty string if no spec found.
func (s *Server) buildTypstConfig(bid int64, book dbgen.Book) string {
	if !book.ProjectID.Valid {
		return ""
	}

	q := dbgen.New(s.DB)
	spec, err := q.GetBookSpec(context.Background(), book.ProjectID.Int64)
	if err != nil {
		slog.Debug("no book spec for project", "project_id", book.ProjectID.Int64, "err", err)
		return ""
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(spec.Data), &data); err != nil {
		slog.Warn("bad spec JSON", "project_id", book.ProjectID.Int64, "err", err)
		return ""
	}

	return specToTypstConfig(data)
}

func escapeTypstString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

var typstMentionRe = regexp.MustCompile(`(^|[^[:alnum:]_/])@([A-Za-z0-9_]+)`)
var typstImageClampRe = regexp.MustCompile(`#image\("([^"]+)"(?:,\s*width:\s*[^,\)]+)?(?:,\s*height:\s*[^,\)]+)?([^\)]*)\)`)

func literalTypstMentions(s string) string {
	return typstMentionRe.ReplaceAllString(s, `${1}#sym.at#h(0em)${2}`)
}

func clampTypstInlineImages(s string) string {
	return typstImageClampRe.ReplaceAllString(s, `#image("$1", width: 100%, fit: "contain"$2)`)
}

func sanitizeFilename(s string) string {
	s = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9',
			r == '-', r == '_', r == '.':
			return r
		case r == ' ':
			return '-'
		default:
			return -1
		}
	}, s)
	return s
}
