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
	"strconv"
	"strings"
	"time"

	"srv.exe.dev/db/dbgen"
)

const maxUploadSize = 50 << 20 // 50 MB

// typesettingRoot returns the absolute path to the bundled typesetting/
// directory (Typst templates, conversion scripts, fonts).
//
// Resolution order:
//  1. JDBBS_TYPESETTING_DIR env var (if set), resolved to absolute.
//  2. ./typesetting relative to the working directory (production layout).
//  3. Walk up parent directories looking for a sibling `typesetting/`
//     (handles `go test ./srv/...` running with CWD = repo/srv).
//  4. Fall back to "typesetting" as a last resort.
func typesettingRoot() string {
	if env := os.Getenv("JDBBS_TYPESETTING_DIR"); env != "" {
		if abs, err := filepath.Abs(env); err == nil {
			return abs
		}
		return env
	}
	if abs, err := filepath.Abs("typesetting"); err == nil {
		if _, err := os.Stat(abs); err == nil {
			return abs
		}
	}
	if cwd, err := os.Getwd(); err == nil {
		dir := cwd
		for i := 0; i < 6; i++ {
			candidate := filepath.Join(dir, "typesetting")
			if _, err := os.Stat(candidate); err == nil {
				return candidate
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	abs, err := filepath.Abs("typesetting")
	if err != nil {
		return "typesetting"
	}
	return abs
}

// Helpers for paths inside the typesetting tree.
func convertScriptPath() string  { return filepath.Join(typesettingRoot(), "scripts", "md-to-chapter.py") }
func seriesTemplatePath() string { return filepath.Join(typesettingRoot(), "templates", "series-template.typ") }
func fontsDirPath() string       { return filepath.Join(typesettingRoot(), "fonts") }

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

	// Step 1: pandoc docx → markdown with styles
	mdPath := filepath.Join(tmpDir, "chapter.md")
	pandocCmd := exec.Command("pandoc",
		"--from=docx+styles",
		"--to=markdown+fenced_divs",
		"-o", mdPath,
		docxPath,
	)
	if out, err := pandocCmd.CombinedOutput(); err != nil {
		s.failConversion(bid, fmt.Sprintf("pandoc: %s\n%s", err, string(out)))
		return
	}

	// Step 2: python md → typst chapter (script writes to stdout)
	chapterPath := filepath.Join(tmpDir, "chapter.typ")
	pyCmd := exec.Command("python3", convertScriptPath(), mdPath, book.Title, book.Author)
	chapterOut, err := pyCmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			s.failConversion(bid, fmt.Sprintf("md-to-chapter: %s\n%s", err, string(ee.Stderr)))
		} else {
			s.failConversion(bid, fmt.Sprintf("md-to-chapter: %s", err))
		}
		return
	}
	if err := os.WriteFile(chapterPath, chapterOut, 0644); err != nil {
		s.failConversion(bid, "write chapter: "+err.Error())
		return
	}

	// Step 3: Create a standalone main.typ that includes the chapter
	// Check for a book spec (project-linked config overrides)
	configOverride := s.buildTypstConfig(bid, book)

	mainTyp := fmt.Sprintf(`#import "%s": *
%s
#show: book.with(
  title: "%s",
  subtitle: "%s",
)

// Chapter content
#set page(numbering: "1")

#set-story-info(title: "%s", author: "%s")
#no-header()

#include "chapter.typ"
`,
		seriesTemplatePath(),
		configOverride,
		escapeTypstString(book.Title),
		escapeTypstString(book.Series),
		escapeTypstString(book.Title),
		escapeTypstString(book.Author),
	)

	mainPath := filepath.Join(tmpDir, "main.typ")
	if err := os.WriteFile(mainPath, []byte(mainTyp), 0644); err != nil {
		s.failConversion(bid, "write main.typ: "+err.Error())
		return
	}

	// Also symlink styles.typ and images.typ if they exist
	for _, dep := range []string{"styles.typ", "images.typ"} {
		src := filepath.Join(typesettingRoot(), "templates", dep)
		if _, err := os.Stat(src); err == nil {
			os.Symlink(src, filepath.Join(tmpDir, dep))
		}
	}

	// Step 4: typst compile (root=/ so absolute template paths work)
	pdfPath := filepath.Join(tmpDir, "output.pdf")
	typstCmd := exec.Command("typst", "compile",
		"--root", "/",
		"--font-path", fontsDirPath(),
		mainPath,
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

	format := r.PathValue("format")

	q := dbgen.New(s.DB)

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
