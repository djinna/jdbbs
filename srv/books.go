package srv

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
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
func epubFilterPath() string {
	return filepath.Join(typesettingRoot(), "filters", "docx-to-epub.lua")
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
//
// Auth (Factory Pass): with a project_id, the caller may be the project's own
// customer — requireAuth(project_id) plus a live pass. Without a project_id
// (an unlinked, admin-managed book) it stays admin-only. See
// docs/specs/FACTORY-PASS-API-2026-09-03.md.
func (s *Server) handleUploadBook(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		jsonErr(w, "file too large or bad form", 400)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	author := strings.TrimSpace(r.FormValue("author"))
	series := strings.TrimSpace(r.FormValue("series"))
	projectIDStr := strings.TrimSpace(r.FormValue("project_id"))

	var projectID sql.NullInt64
	if projectIDStr != "" {
		pid, err := strconv.ParseInt(projectIDStr, 10, 64)
		if err != nil || pid <= 0 {
			jsonErr(w, "bad project_id", 400)
			return
		}
		projectID = sql.NullInt64{Int64: pid, Valid: true}
	}
	if !projectID.Valid {
		// Non-admins must always name the project they're uploading into;
		// an unlinked book has no pass to authorize it.
		if !s.requireExeDevAdminAPI(w, r) {
			return
		}
	} else if _, _, ok := s.requirePassAccess(w, r, projectID.Int64); !ok {
		return
	}

	if title == "" || author == "" {
		jsonErr(w, "title and author required", 400)
		return
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

	// TRK-DEV-012 Phase C: auto-detect chapter title/author pairs from the
	// manuscript and stage them as spec.chapters_suggested[]. Runs in the
	// background (project-linked books only) so it never delays or fails the
	// upload — see detectChaptersAsync. The admin Anthology card surfaces the
	// result behind an Apply button.
	if book.ProjectID.Valid {
		go s.detectChaptersAsync(book)
	}
	slog.Info("manuscript uploaded", "book_id", book.ID, "project_id", projectID.Int64,
		"title", book.Title, "file", header.Filename, "bytes", len(data), "who", requestActor(r))
	if projectID.Valid {
		s.factoryEventR(r, projectID.Int64, "manuscript.uploaded",
			fmt.Sprintf("%s (%s) → book %d", header.Filename, formatBytesIEC(int64(len(data))), book.ID))
	}

	w.WriteHeader(201)
	jsonOK(w, map[string]any{
		"id":     book.ID,
		"title":  book.Title,
		"author": book.Author,
		"status": book.Status,
	})
}

// handleDetectChapters re-runs chapter auto-detection for an already-uploaded
// book and returns the staged suggestions. This is the "re-scan" path for
// DOCXs uploaded before Phase C, or after a manuscript is re-uploaded. Unlike
// the upload trigger it runs synchronously so the admin UI gets the result
// back immediately to render.
//
// TRK-DEV-012 Phase C. Auth: the project's own customer (live pass) or admin.
func (s *Server) handleDetectChapters(w http.ResponseWriter, r *http.Request) {
	bid, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}

	q := dbgen.New(s.DB)
	ref, err := q.GetBookProjectID(r.Context(), bid)
	if err != nil {
		jsonErr(w, "not found", 404)
		return
	}
	if !ref.ProjectID.Valid {
		// No project means no pass to authorize against, so this is an
		// admin-managed book either way.
		if !s.requireExeDevAdminAPI(w, r) {
			return
		}
		jsonErr(w, "book is not linked to a project; link it first so suggestions have a spec to land in", 400)
		return
	}
	if _, _, ok := s.requirePassAccess(w, r, ref.ProjectID.Int64); !ok {
		return
	}

	book, err := q.GetBook(r.Context(), bid)
	if err != nil {
		jsonErr(w, "not found", 404)
		return
	}
	if len(book.SourceData) == 0 {
		jsonErr(w, "no source file", 400)
		return
	}

	chs, err := s.detectAndStoreChapterSuggestions(r.Context(), book)
	if err != nil {
		jsonErr(w, "detection failed: "+err.Error(), 500)
		return
	}
	jsonOK(w, map[string]any{
		"chapters_suggested": chs,
		"count":              len(chs),
	})
}

// handleConvertBook runs the docx → typst → PDF pipeline. This is the metered
// step of a Factory Pass: one successful convert = one build.
//
// Gating (docs/specs/FACTORY-PASS-API-2026-09-03.md):
//   - customer: requireAuth(project) + live pass + credits_remaining > 0 (else 402)
//   - admin: gating skipped, but the ledger is still debited when the project
//     has a pass, so instructor-run builds count and the dogfood data is real
//   - one in-flight build per project (409) regardless of who asked
func (s *Server) handleConvertBook(w http.ResponseWriter, r *http.Request) {
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

	// Which deliverables to make. The default (and what the factory page
	// asks for) is "both": one build = the EPUB and the print PDF together,
	// and it costs one build credit. "epub" and "pdf" alone are accepted for
	// the API and cost the same one credit — a build is a build.
	format := "both"
	callbackURL := ""
	if r.Body != nil && r.ContentLength != 0 {
		var body struct {
			Format      string `json:"format"`
			CallbackURL string `json:"callback_url"`
		}
		if err := json.NewDecoder(io.LimitReader(r.Body, 4096)).Decode(&body); err == nil {
			if body.Format != "" {
				format = strings.ToLower(strings.TrimSpace(body.Format))
			}
			callbackURL = strings.TrimSpace(body.CallbackURL)
		}
	}
	if format != "epub" && format != "pdf" && format != "both" {
		jsonErr(w, "format must be epub, pdf or both", 400)
		return
	}
	if callbackURL != "" {
		if err := validateCallbackURL(callbackURL, s.allowLocalCallbacks); err != nil {
			jsonErr(w, "callback_url: "+err.Error(), 400)
			return
		}
	}

	var pass *dbgen.Pass
	if !book.ProjectID.Valid {
		if !s.requireExeDevAdminAPI(w, r) {
			return
		}
	} else {
		var isAdmin, ok bool
		pass, isAdmin, ok = s.requirePassAccess(w, r, book.ProjectID.Int64)
		if !ok {
			return
		}
		if pass != nil && !isAdmin && passCreditsRemaining(*pass) <= 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusPaymentRequired)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error":             "no builds remaining",
				"credits_remaining": 0,
			})
			return
		}
		// One build in flight per project: builds are minutes of CPU and the
		// status field is per-book, so a second concurrent run would race the
		// first one's status and outputs.
		inFlight, err := q.CountConvertingBooksByProject(r.Context(), book.ProjectID)
		if err != nil {
			jsonErr(w, err.Error(), 500)
			return
		}
		if inFlight > 0 {
			jsonErr(w, "a build is already running for this project; wait for it to finish", http.StatusConflict)
			return
		}
	}

	if len(book.SourceData) == 0 {
		jsonErr(w, "no source file", 400)
		return
	}

	// A final transmittal saved since the spec was written refreshes the
	// spec now, so the build uses what the author (or their machine) sent
	// without a template download in between.
	if book.ProjectID.Valid {
		if _, err := s.syncSpecFromTransmittal(r.Context(), book.ProjectID.Int64, false); err != nil {
			jsonErr(w, "spec: "+err.Error(), 500)
			return
		}
	}

	// Debit before starting so a crashed or killed build can never hand out a
	// free one; failConversion refunds. Debited for admins too when a pass
	// exists (see the contract).
	if pass != nil {
		if err := s.debitBuildCredit(r.Context(), pass.ID, bid); err != nil {
			slog.Error("build debit failed", "pass_id", pass.ID, "book_id", bid, "err", err)
			jsonErr(w, "could not reserve a build credit", 500)
			return
		}
	}

	// Mark as converting
	_ = q.UpdateBookStatus(r.Context(), dbgen.UpdateBookStatusParams{
		Status: "converting", ID: bid,
	})

	if book.ProjectID.Valid {
		left := ""
		if pass != nil {
			if p := s.passForProject(r.Context(), book.ProjectID.Int64); p != nil {
				left = fmt.Sprintf("; %d build(s) left after this", passCreditsRemaining(*p))
			}
		}
		s.factoryEventR(r, book.ProjectID.Int64, "build.started", fmt.Sprintf("book %d, %s%s", bid, format, left))
	}

	// Run conversion in background; tell the caller's machine when it lands.
	go func() {
		s.runConversion(bid, book, format)
		if callbackURL != "" {
			s.postBuildCallback(callbackURL, bid)
		}
	}()

	jsonOK(w, map[string]any{
		"status": "converting", "format": format, "book_id": bid,
		"status_url": fmt.Sprintf("/api/books/%d", bid),
	})
}

// ─── Build status: GET /api/books/{id} and the optional callback ───
//
// The machine-callable end of the factory (MACHINE-FACTORY-SPEC §2). One
// JSON shape answers "is it done, and where are the files": returned by
// GET /api/books/{id}, and POSTed to callback_url when a build finishes.

type buildStatusOutput struct {
	ID          int64     `json:"id"`
	Format      string    `json:"format"`
	SizeBytes   int64     `json:"size_bytes"`
	CreatedAt   time.Time `json:"created_at"`
	DownloadURL string    `json:"download_url"`
}

type buildStatus struct {
	BookID    int64               `json:"book_id"`
	ProjectID *int64              `json:"project_id"`
	Title     string              `json:"title"`
	Author    string              `json:"author"`
	Filename  string              `json:"source_filename"`
	Status    string              `json:"status"` // uploaded | converting | ready | error
	Error     string              `json:"error,omitempty"`
	UpdatedAt time.Time           `json:"updated_at"`
	Outputs   []buildStatusOutput `json:"outputs"`
}

func (s *Server) buildStatusFor(ctx context.Context, bid int64) (*buildStatus, error) {
	q := dbgen.New(s.DB)
	book, err := q.GetBook(ctx, bid)
	if err != nil {
		return nil, err
	}
	rows, err := q.ListBookOutputs(ctx, dbgen.ListBookOutputsParams{BookID: bid, Limit: 20})
	if err != nil {
		return nil, err
	}
	st := &buildStatus{
		BookID: book.ID, Title: book.Title, Author: book.Author, Filename: book.SourceFilename,
		Status: book.Status, Error: book.ErrorMsg, UpdatedAt: book.UpdatedAt,
		Outputs: make([]buildStatusOutput, 0, len(rows)),
	}
	if book.ProjectID.Valid {
		pid := book.ProjectID.Int64
		st.ProjectID = &pid
	}
	for _, row := range rows {
		st.Outputs = append(st.Outputs, buildStatusOutput{
			ID: row.ID, Format: row.OutputFormat, SizeBytes: row.SizeBytes.Int64, CreatedAt: row.CreatedAt,
			DownloadURL: fmt.Sprintf("/api/books/%d/outputs/%d/download", bid, row.ID),
		})
	}
	return st, nil
}

// handleGetBook is the completion signal for a build: poll it after convert
// until status is "ready" or "error". Same auth as the outputs list.
func (s *Server) handleGetBook(w http.ResponseWriter, r *http.Request) {
	bid, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	if _, ok := s.bookAuth(w, r, bid); !ok {
		return
	}
	st, err := s.buildStatusFor(r.Context(), bid)
	if err != nil {
		jsonErr(w, "not found", 404)
		return
	}
	jsonOK(w, st)
}

// validateCallbackURL keeps callback_url from being used to make the server
// poke at itself or its neighbours (SSRF): http(s) only, a hostname, and no
// loopback / private / link-local address — checked on the literal and on
// what the name resolves to.
func validateCallbackURL(raw string, allowLocal bool) error {
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("must be http or https")
	}
	host := u.Hostname()
	if host == "" {
		return errors.New("missing host")
	}
	if u.User != nil {
		return errors.New("credentials in the URL are not allowed")
	}
	if allowLocal {
		return nil
	}
	if strings.EqualFold(host, "localhost") || strings.HasSuffix(strings.ToLower(host), ".localhost") {
		return errors.New("localhost is not allowed")
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return fmt.Errorf("cannot resolve %s", host)
	}
	for _, ip := range ips {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast() {
			return fmt.Errorf("%s resolves to a private or local address", host)
		}
	}
	return nil
}

// postBuildCallback POSTs the build status JSON to the caller's URL once,
// after the build has finished. Best effort: a failure is logged and noted on
// the Floor, never retried — the caller can always GET /api/books/{id}.
func (s *Server) postBuildCallback(callbackURL string, bid int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	st, err := s.buildStatusFor(ctx, bid)
	if err != nil {
		slog.Error("build callback: status lookup failed", "book_id", bid, "err", err)
		return
	}
	body, _ := json.Marshal(st)
	req, err := http.NewRequestWithContext(ctx, "POST", callbackURL, bytes.NewReader(body))
	if err != nil {
		slog.Error("build callback: bad request", "book_id", bid, "err", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "jdbb-factory/1.0 (+https://jdbbs.exe.xyz/factory)")
	client := &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse // no following redirects into places we did not vet
		},
	}
	resp, err := client.Do(req)
	detail := fmt.Sprintf("book %d → %s", bid, redactURL(callbackURL))
	if err != nil {
		slog.Warn("build callback failed", "book_id", bid, "url", redactURL(callbackURL), "err", err)
		detail += ": " + clip(err.Error(), 200)
	} else {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		resp.Body.Close()
		detail += fmt.Sprintf(": HTTP %d", resp.StatusCode)
		if resp.StatusCode >= 300 {
			slog.Warn("build callback rejected", "book_id", bid, "url", redactURL(callbackURL), "status", resp.StatusCode)
		}
	}
	if st.ProjectID != nil {
		s.factoryEvent(*st.ProjectID, "", "build.callback", "factory", detail)
	}
}

// redactURL drops the query string (where tokens tend to live) for logs.
func redactURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return "(unparseable url)"
	}
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}

// epubOutputsKept is how many EPUB outputs a book keeps; older ones are
// pruned so repeated builds don't fill the outputs table.
const epubOutputsKept = 10

// maxConcurrentBuilds is the global build cap; see Server.buildSem.
const maxConcurrentBuilds = 2

// acquireBuildSlot blocks until one of the maxConcurrentBuilds slots is free
// and returns the release func. Queue time is logged so a workshop-day
// backlog shows up in the journal.
func (s *Server) acquireBuildSlot(bid int64) func() {
	s.buildSemOnce.Do(func() { s.buildSem = make(chan struct{}, maxConcurrentBuilds) })
	start := time.Now()
	s.buildSem <- struct{}{}
	if waited := time.Since(start); waited > time.Second {
		slog.Info("build queued", "id", bid, "waited", waited.Round(time.Millisecond))
	}
	return func() { <-s.buildSem }
}

// runConversion is the build. format is "pdf", "epub" or "both".
func (s *Server) runConversion(bid int64, book dbgen.Book, format string) {
	defer s.acquireBuildSlot(bid)()
	if format == "epub" {
		s.runEPUBBuild(bid, book)
		return
	}
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
		// The transmittal (via the book spec) is the record for title and
		// author; the upload form's values are a fallback for unlinked
		// books. The EPUB already reads spec metadata — the PDF must agree
		// (C21: PDF Author was the Stripe cardholder, EPUB's the author).
		if t, a := s.specTitleAuthor(book); t != "" || a != "" {
			if t != "" {
				book.Title = t
			}
			if a != "" {
				book.Author = a
			}
			if len(pairs) > 0 {
				book.Title = applyPairsToString(book.Title, pairs)
				book.Author = applyPairsToString(book.Author, pairs)
			}
		}
	}

	// Book map (P4): front / body / back from heading text + position, plus
	// the untitled pieces before the first heading. The lua filter turns it
	// into template calls; the spec decides which generated pages i–iv exist.
	specMap := s.specMapForBook(book)
	specTitle, specAuthor := s.specTitleAuthor(book)
	if specTitle == "" {
		specTitle = book.Title
	}
	if specAuthor == "" {
		specAuthor = book.Author
	}
	bookMap, bmErr := bookMapFromDOCX(docxPath, specMap, specTitle, specAuthor)
	if bmErr != nil {
		slog.Warn("book map failed; building without front-matter structure", "book_id", bid, "err", bmErr)
		bookMap = nil
	} else {
		bookMap.Parts = specHasParts(specMap)
		slog.Info("book map", "book_id", bid, "map", bookMap.Summary())
	}

	// Step 1: direct pandoc docx -> typst using the bundled lua filter.
	typPath := filepath.Join(tmpDir, "book.typ")
	pandocArgs := []string{
		"--from=docx+styles",
		docxPath,
		"--lua-filter=" + typstFilterPath(),
		"--extract-media=" + filepath.Join(tmpDir, "media"),
		"-t", "typst+smart",
		"-o", typPath,
	}

	// Spec-derived metadata for the Lua filter (anthology chapters, declared
	// custom styles) via --metadata-file. Books with neither flow through unchanged.
	if metaFile, ok := s.writePandocMetadata(book, tmpDir, bookMap, specFrontMatterTOC(specMap)); ok {
		pandocArgs = append(pandocArgs, "--metadata-file="+metaFile)
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
	if bookMap != nil {
		headerReplacement += "  front-matter: " + frontMatterTypst(specMap, book) + ",\n"
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

	// Store PDF in DB (book_outputs is the single artifact store; 018 dropped
	// the duplicate books.pdf_data column).
	ctx := context.Background()
	if _, err := q.CreateBookOutput(ctx, dbgen.CreateBookOutputParams{
		BookID:              bid,
		OutputFormat:        "pdf",
		OutputData:          pdfData,
		SourceFilename:      book.SourceFilename,
		SpecSnapshot:        nullStringFrom(specSnapshot),
		CorrectionsSnapshot: nullStringFrom(correctionsSnapshot),
	}); err != nil {
		s.failConversion(bid, "store pdf: "+err.Error())
		return
	}
	// Step 4: the EPUB, when this is a "both" build. It runs inline — the
	// book stays "converting" until every artifact exists, which is what the
	// customer page polls on. A "pdf" build just marks ready.
	if err := s.finalizeBuild(ctx, bid, book, format == "both"); err != nil {
		s.failConversion(bid, err.Error())
		return
	}

	// Factory Pass: the build succeeded, so the credit debited at request time
	// stands. Send the customer their receipt (EMAIL_SYSTEM.md pathway #7).
	// Re-read the pass so credits_remaining reflects this build.
	if book.ProjectID.Valid {
		if pass := s.passForProject(ctx, book.ProjectID.Int64); pass != nil {
			s.sendBuildDeliveredEmail(*pass, book, format)
		}
	}

	if book.ProjectID.Valid {
		s.factoryEvent(book.ProjectID.Int64, "", "build.done", "factory",
			fmt.Sprintf("book %d, %s, %s", bid, format, time.Since(start).Round(time.Second)))
	}
	slog.Info("book conversion complete", "id", bid, "title", book.Title,
		"pdf_size", len(pdfData),
		"elapsed", time.Since(start))
}

// finalizeBuild completes a build once the PDF is stored: it generates the
// EPUB and only then marks the book ready, so anything polling status sees
// "converting" until both deliverables exist.
//
// An EPUB failure does NOT fail the build. The customer has a correctly
// typeset print PDF — that is most of what they bought — so the status still
// goes to "ready", the reason is left in error_msg as a non-fatal note, and
// the build credit is not refunded. (A refund here would also be wrong the
// other way round: the PDF is downloadable, so the build was delivered.)
//
// A returned error means the status write itself failed, i.e. the build
// genuinely cannot be recorded; the caller treats that as a failed build.
func (s *Server) finalizeBuild(ctx context.Context, bid int64, book dbgen.Book, withEPUB bool) error {
	q := dbgen.New(s.DB)
	if !withEPUB {
		if err := q.UpdateBookStatus(ctx, dbgen.UpdateBookStatusParams{
			Status: "ready", ErrorMsg: "", ID: bid,
		}); err != nil {
			return fmt.Errorf("mark ready: %w", err)
		}
		return nil
	}
	if epubErr := s.getEPUBRunner()(bid, book); epubErr != nil {
		slog.Error("build: epub stage failed but pdf succeeded; delivering pdf-only build",
			"id", bid, "title", book.Title, "raw_error", epubErr)
		note := "Your print PDF is ready, but we couldn't generate the EPUB. Re-save the Word file as .docx and try another build, or email j@djinna.com."
		if err := q.UpdateBookStatus(ctx, dbgen.UpdateBookStatusParams{
			Status: "ready", ErrorMsg: note, ID: bid,
		}); err != nil {
			return fmt.Errorf("mark ready (epub failed): %w", err)
		}
		return nil
	}
	s.pruneEPUBOutputs(ctx, bid)
	// Clear any error_msg left by an earlier attempt: "ready" plus a stale
	// error reads as a broken build in the UI.
	if err := q.UpdateBookStatus(ctx, dbgen.UpdateBookStatusParams{
		Status: "ready", ErrorMsg: "", ID: bid,
	}); err != nil {
		return fmt.Errorf("mark ready: %w", err)
	}
	return nil
}

// runEPUBBuild is the EPUB-only build (format "epub"): no pandoc→typst, no
// typst compile, about a second. It costs a build credit like any other
// build, so a failure goes through failConversion and is refunded.
func (s *Server) runEPUBBuild(bid int64, book dbgen.Book) {
	start := time.Now()
	q := dbgen.New(s.DB)
	ctx := context.Background()
	if err := s.getEPUBRunner()(bid, book); err != nil {
		s.failConversion(bid, "epub: "+err.Error())
		return
	}
	s.pruneEPUBOutputs(ctx, bid)
	if err := q.UpdateBookStatus(ctx, dbgen.UpdateBookStatusParams{
		Status: "ready", ErrorMsg: "", ID: bid,
	}); err != nil {
		slog.Error("epub build: mark ready failed", "id", bid, "err", err)
		return
	}
	if book.ProjectID.Valid {
		s.factoryEvent(book.ProjectID.Int64, "", "build.done", "factory", fmt.Sprintf("book %d, epub only", bid))
		if pass := s.passForProject(ctx, book.ProjectID.Int64); pass != nil {
			s.sendBuildDeliveredEmail(*pass, book, "epub")
		}
	}
	slog.Info("epub build complete", "id", bid, "title", book.Title, "elapsed", time.Since(start))
}

func (s *Server) pruneEPUBOutputs(ctx context.Context, bid int64) {
	if err := dbgen.New(s.DB).PruneBookOutputs(ctx, dbgen.PruneBookOutputsParams{
		BookID: bid, OutputFormat: "epub", Limit: epubOutputsKept,
	}); err != nil {
		slog.Warn("prune epub outputs failed", "book_id", bid, "err", err)
	}
}

var unknownTypstVariableRE = regexp.MustCompile(`(?mi)unknown variable:\s*([A-Za-z0-9_-]+)`)

// customerBuildError turns pipeline stderr into a stable message suitable for
// a workshop customer. The raw trace remains in structured server logs only.
func customerBuildError(raw string) string {
	if match := unknownTypstVariableRE.FindStringSubmatch(raw); len(match) == 2 {
		return fmt.Sprintf("Your file uses a Word style (%s) that isn't in your template. Inspect lists styles not in your transmittal — remove or remap it, or ask us to add it.", match[1])
	}
	if strings.Contains(strings.ToLower(raw), "pandoc") {
		return "We couldn't read this Word file. Re-save it as .docx from Word and try again."
	}
	return "We couldn't build this file. Run Inspect for clues, then email j@djinna.com if it keeps happening."
}

// failConversion marks a build as failed and, when the project holds a Factory
// Pass, refunds the credit debited at request time — a build the customer
// can't download was never a build (+1 build_failed_refund in the ledger).
func (s *Server) failConversion(bid int64, msg string) {
	customerMsg := customerBuildError(msg)
	slog.Error("book conversion failed", "id", bid, "raw_error", msg, "customer_message", customerMsg)
	q := dbgen.New(s.DB)
	ctx := context.Background()
	_ = q.UpdateBookStatus(ctx, dbgen.UpdateBookStatusParams{
		Status:   "error",
		ErrorMsg: clip(customerMsg, 2000),
		ID:       bid,
	})
	ref, err := q.GetBookProjectID(ctx, bid)
	if err != nil || !ref.ProjectID.Valid {
		return
	}
	s.factoryEvent(ref.ProjectID.Int64, "", "build.failed", "factory", fmt.Sprintf("book %d: %s", bid, customerMsg))
	pass := s.passForProject(ctx, ref.ProjectID.Int64)
	if pass == nil {
		return
	}
	if err := s.refundBuildCredit(ctx, pass.ID, bid); err != nil {
		slog.Error("build refund failed", "pass_id", pass.ID, "book_id", bid, "err", err)
	}
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
		if bookRef.ProjectID.Valid {
			s.factoryEventR(r, bookRef.ProjectID.Int64, "download", fmt.Sprintf("pdf, book %d", bid))
		}

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
		if bookRef.ProjectID.Valid {
			s.factoryEventR(r, bookRef.ProjectID.Int64, "download", fmt.Sprintf("epub, book %d", bid))
		}

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
			SizeBytes:      row.SizeBytes.Int64, // length() of a NOT NULL blob; sqlc types it nullable
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

// specTitleAuthor returns metadata.title / metadata.author from the
// project's book spec (populated from the transmittal), or empty strings.
func (s *Server) specTitleAuthor(book dbgen.Book) (string, string) {
	if !book.ProjectID.Valid {
		return "", ""
	}
	q := dbgen.New(s.DB)
	spec, err := q.GetBookSpec(context.Background(), book.ProjectID.Int64)
	if err != nil {
		return "", ""
	}
	var data map[string]any
	if err := json.Unmarshal([]byte(spec.Data), &data); err != nil {
		return "", ""
	}
	meta, _ := data["metadata"].(map[string]any)
	title, _ := meta["title"].(string)
	author, _ := meta["author"].(string)
	return strings.TrimSpace(title), strings.TrimSpace(author)
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
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\n", `\n`) // a literal newline would end the string; \n is a typst escape
	return s
}

// typstMentionRe finds bare @handle mentions that typst would read as a
// citation. Pandoc already emits `\@` for @ in markup, and @ inside a quoted
// string (`#link("…/@user")`) is code mode; both are excluded — rewriting them
// printed a literal "#sym.at" in Obliquities' email and footnote URLs.
var typstMentionRe = regexp.MustCompile(`(^|[^[:alnum:]_/\\"])@([A-Za-z0-9_]+)`)
var typstImageClampRe = regexp.MustCompile(`#image\("([^"]+)"(?:,\s*width:\s*[^,\)]+)?(?:,\s*height:\s*[^,\)]+)?([^\)]*)\)`)

func literalTypstMentions(s string) string {
	return typstMentionRe.ReplaceAllString(s, `${1}#sym.at#h(0em)${2}`)
}

func clampTypstInlineImages(s string) string {
	return typstImageClampRe.ReplaceAllString(s, `#image("$1", width: 100%, fit: "contain"$2)`)
}

// writePandocMetadata looks up the book's spec and writes what the Lua filter
// needs from it to a JSON metadata file passed via --metadata-file:
//
//   - chapters: anthology chapters, so the filter emits per-chapter
//     #set-story-info() calls (TRK-DEV-012 Phase B);
//   - custom_styles: declared custom styles as {word_style, ident, type}, so a
//     Word paragraph/character style the spec declares becomes #ident[...]
//     instead of falling through to a bare #block. ident is typstStyleIdent of
//     the style's name — the same function names the #let in buildTypstConfig.
//
// Returns ("", false) if there is no spec or nothing to pass — the caller then
// runs pandoc without the flag.
func (s *Server) writePandocMetadata(book dbgen.Book, tmpDir string, bookMap *BookMap, toc bool) (string, bool) {
	payload := map[string]any{}
	if bookMap != nil {
		payload["book_map"] = map[string]any{
			"toc":            toc,
			"parts":          bookMap.Parts,
			"sections":       bookMap.Sections,
			"untitled_front": bookMap.UntitledFront,
		}
	}

	specData := ""
	if book.ProjectID.Valid {
		q := dbgen.New(s.DB)
		if spec, err := q.GetBookSpec(context.Background(), book.ProjectID.Int64); err == nil {
			specData = spec.Data
		}
	}
	if specData == "" {
		return writePandocMetadataFile(book, tmpDir, payload)
	}

	parsed := parseEPUBSpec(specData, book)
	if len(parsed.Chapters) > 0 {
		type metaChapter struct {
			Title  string `json:"title"`
			Author string `json:"author"`
			File   string `json:"file,omitempty"`
		}
		chs := make([]metaChapter, 0, len(parsed.Chapters))
		for _, c := range parsed.Chapters {
			chs = append(chs, metaChapter{Title: c.Title, Author: c.Author, File: c.File})
		}
		payload["chapters"] = chs
	}

	styles := declaredStylesForPandoc(specData)
	if len(styles) > 0 {
		payload["custom_styles"] = styles
	}
	return writePandocMetadataFile(book, tmpDir, payload)
}

func writePandocMetadataFile(book dbgen.Book, tmpDir string, payload map[string]any) (string, bool) {
	if len(payload) == 0 {
		return "", false
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", false
	}
	path := filepath.Join(tmpDir, "pandoc-meta.json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		slog.Warn("pandoc metadata: write failed; book map, custom styles and chapters will not reach the filter",
			"book_id", book.ID, "err", err)
		return "", false
	}
	styles, _ := payload["custom_styles"].([]pandocStyle)
	slog.Info("pandoc metadata written for Lua filter",
		"book_id", book.ID, "has_chapters", payload["chapters"] != nil, "has_book_map", payload["book_map"] != nil,
		"custom_styles", len(styles))
	return path, true
}

// specMapForBook returns the linked project's spec as a map, or nil.
func (s *Server) specMapForBook(book dbgen.Book) map[string]any {
	if !book.ProjectID.Valid {
		return nil
	}
	q := dbgen.New(s.DB)
	spec, err := q.GetBookSpec(context.Background(), book.ProjectID.Int64)
	if err != nil {
		return nil
	}
	var data map[string]any
	if err := json.Unmarshal([]byte(spec.Data), &data); err != nil {
		return nil
	}
	return data
}

// specFrontMatterTOC: does the build generate a contents page? Default yes.
func specFrontMatterTOC(spec map[string]any) bool {
	fm, _ := spec["front_matter"].(map[string]any)
	if v, ok := fm["toc"].(bool); ok {
		return v
	}
	return true
}

// specHasParts reports whether the book is divided into parts (Heading 1 =
// part, Heading 2 = chapter). Opt-in from the transmittal: the "Parts" count
// in checklist_stats (any number ≥ 1), or an explicit structure.parts flag.
func specHasParts(spec map[string]any) bool {
	if st, ok := spec["structure"].(map[string]any); ok {
		if v, ok := st["parts"].(bool); ok {
			return v
		}
	}
	cs, _ := spec["checklist_stats"].(map[string]any)
	switch v := cs["parts"].(type) {
	case float64:
		return v >= 1
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(v))
		return err == nil && n >= 1
	}
	return false
}

// frontMatterTypst builds the `front-matter:` dict for series-template.typ's
// book(): which generated pages exist (spec.front_matter, default on) and the
// text they carry (spec.metadata, spec.cover). Never the manuscript's own
// title page — that is dropped by the book map.
func frontMatterTypst(spec map[string]any, book dbgen.Book) string {
	fm, _ := spec["front_matter"].(map[string]any)
	meta, _ := spec["metadata"].(map[string]any)
	cover, _ := spec["cover"].(map[string]any)
	on := func(k string) bool {
		if v, ok := fm[k].(bool); ok {
			return v
		}
		return true
	}
	str := func(m map[string]any, k string) string {
		v, _ := m[k].(string)
		return strings.TrimSpace(v)
	}
	q := func(v string) string {
		if v == "" {
			return "none"
		}
		return `"` + escapeTypstString(v) + `"`
	}
	fields := []string{
		fmt.Sprintf("half-title: %t", on("half_title")),
		fmt.Sprintf("title-page: %t", on("title_page")),
		fmt.Sprintf("copyright-page: %t", on("copyright_page")),
		"subtitle: " + q(str(meta, "subtitle")),
		"publisher: " + q(str(meta, "publisher")),
		"isbn-paper: " + q(str(meta, "isbn_paper")),
		"isbn-epub: " + q(str(meta, "isbn_epub")),
		"copyright-year: " + q(str(meta, "copyright_year")),
		"copyright-holder: " + q(str(meta, "copyright_holder")),
		"credit-lines: " + q(str(meta, "credit_lines")),
		"cover-credit: " + q(str(cover, "credit")),
		// C12 copyright-page builder fields.
		"publisher-city: " + q(str(meta, "publisher_city")),
		"edition-line: " + q(str(meta, "edition_line")),
		"interior-credit: " + q(interiorCredit(spec)),
		"loc-line: " + q(str(meta, "loc_line")),
		"printed-in: " + q(str(meta, "printed_in")),
		"notices: " + q(str(meta, "additional_notices")),
		"logo: none",
	}
	return "(" + strings.Join(fields, ", ") + ")"
}

// interiorCreditDefault is what the copyright page says about the typesetting
// when the transmittal leaves the line blank. {typeface} is filled from the
// spec's body font at build time (here and in generate-word-template.py).
const interiorCreditDefault = "Typeset by jdbb studio in {typeface}"

// interiorCredit returns the resolved interior/typesetting credit line for the
// copyright page: metadata.interior_credit or the default, with {typeface}
// replaced by typography.body_font.
func interiorCredit(spec map[string]any) string {
	meta, _ := spec["metadata"].(map[string]any)
	typo, _ := spec["typography"].(map[string]any)
	line, _ := meta["interior_credit"].(string)
	line = strings.TrimSpace(line)
	if line == "" {
		line = interiorCreditDefault
	}
	font, _ := typo["body_font"].(string)
	font = strings.TrimSpace(font)
	if font == "" {
		font = "Libertinus Serif"
	}
	return strings.ReplaceAll(line, "{typeface}", font)
}

// pandocStyle is one declared custom style as the Lua filter wants it.
type pandocStyle struct {
	WordStyle string `json:"word_style"`
	Ident     string `json:"ident"`
	Type      string `json:"type"`
}

// declaredStylesForPandoc extracts the spec's custom_styles into the
// {word_style, ident, type} triples the filter maps on.
func declaredStylesForPandoc(specJSON string) []pandocStyle {
	var data map[string]any
	if err := json.Unmarshal([]byte(specJSON), &data); err != nil {
		return nil
	}
	raw, ok := data["custom_styles"].([]any)
	if !ok {
		return nil
	}
	out := []pandocStyle{}
	for _, item := range raw {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		name, _ := m["name"].(string)
		word, _ := m["word_style"].(string)
		typ, _ := m["type"].(string)
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if strings.TrimSpace(word) == "" {
			word = name
		}
		if typ != "character" {
			typ = "paragraph"
		}
		out = append(out, pandocStyle{WordStyle: strings.TrimSpace(word), Ident: typstStyleIdent(name), Type: typ})
	}
	return out
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
