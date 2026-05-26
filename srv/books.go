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

const maxUploadSize = 50 << 20 // 50 MB

// typesettingRoot returns the absolute path to the bundled typesetting/
// directory (Typst templates, conversion scripts, lua filters, fonts).
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
func typstFilterPath() string {
	return filepath.Join(typesettingRoot(), "filters", "docx-to-typst-enhanced.lua")
}
func seriesTemplatePath() string {
	return filepath.Join(typesettingRoot(), "templates", "series-template.typ")
}
func fontsDirPath() string { return filepath.Join(typesettingRoot(), "fonts") }

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

	// Step 0: apply project's pending corrections to the source docx (no-op
	// when the book isn't project-linked or has no pending corrections). The
	// resulting bytes flow through pandoc → typst like any other source.
	// Also patch in-memory metadata (Title/Author) so the typst template's
	// title page reflects corrections — the template reads book.Title and
	// book.Author, not docx content.
	ctxApply := context.Background()
	var correctionsSnapshot string
	if book.ProjectID.Valid {
		var pairs []correctionPair
		correctionsSnapshot, pairs = s.applyCorrectionsIfAny(ctxApply, bid, book.ProjectID.Int64, tmpDir, docxPath)
		if len(pairs) > 0 {
			book.Title = applyPairsToString(book.Title, pairs)
			book.Author = applyPairsToString(book.Author, pairs)
			book.Series = applyPairsToString(book.Series, pairs)
		}
	}

	// Step 1: direct pandoc docx -> typst using the bundled lua filter.
	typPath := filepath.Join(tmpDir, "book.typ")
	pandocArgs := []string{
		"--from=docx+styles",
		docxPath,
		"--lua-filter=" + typstFilterPath(),
		"--extract-media=" + filepath.Join(tmpDir, "media"),
		"-t", "typst",
		"-o", typPath,
	}

	// TRK-DEV-012 Phase B: pass anthology chapter metadata to the Lua filter
	// via --metadata-file so it can emit per-chapter #set-story-info() calls.
	// Single-author books (no chapters configured) flow through unchanged.
	if chaptersFile, ok := s.writeChaptersMetadata(book, tmpDir); ok {
		pandocArgs = append(pandocArgs, "--metadata-file="+chaptersFile)
	}

	pandocCmd := exec.Command("pandoc", pandocArgs...)
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

	configOverride, specSnapshot := s.buildTypstConfig(bid, book)
	// Pass config explicitly to book.with so the caller's merged config (above)
	// reaches the template's body styling. Without this, book()'s `config: config`
	// parameter default captures the template module's default-config rather than
	// the local override — overrides on body-font, base-size, etc. silently no-op.
	headerReplacement := fmt.Sprintf(`#import "%s": *
%s
#show: book.with(
  config: config,
  title: "%s",
  author: "%s",
`,
		seriesTemplatePath(),
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
		"--font-path", fontsDirPath(),
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
		BookID:              bid,
		OutputFormat:        "pdf",
		OutputData:          pdfData,
		SourceFilename:      book.SourceFilename,
		SpecSnapshot:        nullStringFrom(specSnapshot),
		CorrectionsSnapshot: nullStringFrom(correctionsSnapshot),
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

	// updated_at distinguishes back-to-back compiles in the download filename.
	bookMeta, err := q.GetBook(r.Context(), bid)
	if err != nil {
		jsonErr(w, "not found", 404)
		return
	}
	ts := bookMeta.UpdatedAt.UTC().Format("20060102-1504")

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
		filename := fmt.Sprintf("%s-%s.pdf", sanitizeFilename(row.Title), ts)
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
		w.Header().Set("Content-Length", strconv.Itoa(len(row.PdfData)))
		w.Header().Set("Cache-Control", "no-store")
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
		filename := fmt.Sprintf("%s-%s.epub", sanitizeFilename(row.Title), ts)
		w.Header().Set("Content-Type", "application/epub+zip")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
		w.Header().Set("Content-Length", strconv.Itoa(len(row.EpubData)))
		w.Header().Set("Cache-Control", "no-store")
		w.Write(row.EpubData)

	default:
		jsonErr(w, "format must be pdf or epub", 400)
	}
}

// bookAuth gates a book by its project's auth (or admin if unlinked).
// Returns book metadata on success so callers can reuse the lookup.
func (s *Server) bookAuth(w http.ResponseWriter, r *http.Request, bid int64) (dbgen.GetBookProjectIDRow, bool) {
	q := dbgen.New(s.DB)
	ref, err := q.GetBookProjectID(r.Context(), bid)
	if err != nil {
		jsonErr(w, "not found", 404)
		return ref, false
	}
	if ref.ProjectID.Valid {
		if !s.requireAuth(w, r, ref.ProjectID.Int64) {
			return ref, false
		}
	} else if !s.requireExeDevAdminAPI(w, r) {
		return ref, false
	}
	return ref, true
}

// handleListBookOutputs returns recent compile artifacts for a book (metadata only).
// Pass ?include=spec,corrections (comma-separated) to also return the matching
// snapshots per row.
func (s *Server) handleListBookOutputs(w http.ResponseWriter, r *http.Request) {
	bid, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	if _, ok := s.bookAuth(w, r, bid); !ok {
		return
	}

	var includeSpec, includeCorrections bool
	for _, tok := range strings.Split(r.URL.Query().Get("include"), ",") {
		switch strings.TrimSpace(tok) {
		case "spec":
			includeSpec = true
		case "corrections":
			includeCorrections = true
		}
	}
	q := dbgen.New(s.DB)
	rows, err := q.ListBookOutputs(r.Context(), dbgen.ListBookOutputsParams{
		BookID: bid,
		Limit:  20,
	})
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}

	type outRow struct {
		ID                  int64     `json:"id"`
		BookID              int64     `json:"book_id"`
		OutputFormat        string    `json:"output_format"`
		SourceFilename      string    `json:"source_filename"`
		SizeBytes           int64     `json:"size_bytes"`
		CreatedAt           time.Time `json:"created_at"`
		SpecSnapshot        *string   `json:"spec_snapshot,omitempty"`
		CorrectionsSnapshot *string   `json:"corrections_snapshot,omitempty"`
	}
	out := make([]outRow, 0, len(rows))
	for _, row := range rows {
		item := outRow{
			ID:             row.ID,
			BookID:         row.BookID,
			OutputFormat:   row.OutputFormat,
			SourceFilename: row.SourceFilename,
			SizeBytes:      row.SizeBytes,
			CreatedAt:      row.CreatedAt,
		}
		if includeSpec && row.SpecSnapshot.Valid {
			s := row.SpecSnapshot.String
			item.SpecSnapshot = &s
		}
		if includeCorrections && row.CorrectionsSnapshot.Valid {
			c := row.CorrectionsSnapshot.String
			item.CorrectionsSnapshot = &c
		}
		out = append(out, item)
	}
	jsonOK(w, out)
}

// handleDownloadBookOutput streams a specific historical compile artifact.
func (s *Server) handleDownloadBookOutput(w http.ResponseWriter, r *http.Request) {
	bid, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	oid, err := strconv.ParseInt(r.PathValue("output_id"), 10, 64)
	if err != nil {
		jsonErr(w, "bad output_id", 400)
		return
	}
	if _, ok := s.bookAuth(w, r, bid); !ok {
		return
	}

	q := dbgen.New(s.DB)
	row, err := q.GetBookOutput(r.Context(), dbgen.GetBookOutputParams{ID: oid, BookID: bid})
	if err != nil {
		jsonErr(w, "not found", 404)
		return
	}

	bookMeta, err := q.GetBook(r.Context(), bid)
	if err != nil {
		jsonErr(w, "not found", 404)
		return
	}

	ts := row.CreatedAt.UTC().Format("20060102-1504")
	ext := row.OutputFormat
	ct := "application/octet-stream"
	switch row.OutputFormat {
	case "pdf":
		ct = "application/pdf"
	case "epub":
		ct = "application/epub+zip"
	}
	filename := fmt.Sprintf("%s-%s.%s", sanitizeFilename(bookMeta.Title), ts, ext)
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Header().Set("Content-Length", strconv.Itoa(len(row.OutputData)))
	w.Header().Set("Cache-Control", "no-store")
	w.Write(row.OutputData)
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

// buildTypstConfig looks up the book_spec for the linked project and returns
// (Typst config override lines, raw spec JSON for snapshotting). Both empty if no spec.
func (s *Server) buildTypstConfig(bid int64, book dbgen.Book) (string, string) {
	if !book.ProjectID.Valid {
		return "", ""
	}

	q := dbgen.New(s.DB)
	spec, err := q.GetBookSpec(context.Background(), book.ProjectID.Int64)
	if err != nil {
		slog.Debug("no book spec for project", "project_id", book.ProjectID.Int64, "err", err)
		return "", ""
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(spec.Data), &data); err != nil {
		slog.Warn("bad spec JSON", "project_id", book.ProjectID.Int64, "err", err)
		return "", ""
	}

	return specToTypstConfig(data), spec.Data
}

// nullStringFrom maps "" to NULL so legacy/no-spec rows stay NULL instead of empty.
func nullStringFrom(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
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

// writeChaptersMetadata looks up the book's spec, extracts anthology chapters,
// and writes them to a JSON metadata file that pandoc can pass to the Lua
// filter via --metadata-file. Returns the file path and true on success.
// Returns ("", false) if the book has no spec, no chapters, or the write
// fails — caller should treat these as a single-author book.
//
// TRK-DEV-012 Phase B.
func (s *Server) writeChaptersMetadata(book dbgen.Book, tmpDir string) (string, bool) {
	if !book.ProjectID.Valid {
		return "", false
	}
	q := dbgen.New(s.DB)
	spec, err := q.GetBookSpec(context.Background(), book.ProjectID.Int64)
	if err != nil {
		return "", false
	}
	parsed := parseEPUBSpec(spec.Data, book)
	if len(parsed.Chapters) == 0 {
		return "", false
	}
	type metaChapter struct {
		Title  string `json:"title"`
		Author string `json:"author"`
		File   string `json:"file,omitempty"`
	}
	chs := make([]metaChapter, 0, len(parsed.Chapters))
	for _, c := range parsed.Chapters {
		chs = append(chs, metaChapter{Title: c.Title, Author: c.Author, File: c.File})
	}
	payload, err := json.Marshal(map[string]any{"chapters": chs})
	if err != nil {
		return "", false
	}
	path := filepath.Join(tmpDir, "chapters.json")
	if err := os.WriteFile(path, payload, 0644); err != nil {
		slog.Warn("chapters metadata: write failed; PDF will lack per-chapter set-story-info()",
			"book_id", book.ID, "err", err)
		return "", false
	}
	slog.Info("chapters metadata written for Lua filter",
		"book_id", book.ID, "chapters", len(chs))
	return path, true
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
