package srv

// Back-of-book index add-on (punch list 5.13). The pieces live in
// srv/indexer (drafting, anchoring) and the series template (index-page());
// this file is the server surface: the $100 entitlement on the pass, the
// draft → review → include state machine on the book, and the hook the
// build uses to place markers. Design: docs/reviews/INDEX-ADDON-DESIGN-2026-09-19.md.
//
//	POST /api/books/{id}/index/draft   async; pass-gated (index_included or admin)
//	GET  /api/books/{id}/index         the stored document + status
//	PUT  /api/books/{id}/index         the edited document; sets reviewed
//	POST /api/books/{id}/convert {"index": true}   builds with the index
//	POST /api/admin/passes/{id}/index  admin grant of the entitlement

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
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
	"srv.exe.dev/srv/indexer"
)

// books.index_status values.
const (
	indexOff      = "off"
	indexDrafting = "drafting"
	indexDraft    = "draft"
	indexReviewed = "reviewed"
	indexError    = "error"
)

// indexDraftTimeout bounds one drafting run (Ghosts, 100 pp: ~5 min).
const indexDraftTimeout = 40 * time.Minute

// indexMaxEntries caps a reviewed document; a 400-page book indexes at ~1300 lines.
const indexMaxEntries = 4000

// passIndexIncluded: does this pass carry the add-on?
func passIndexIncluded(p *dbgen.Pass) bool {
	return p != nil && p.IndexIncluded != 0
}

// indexIncludable reports whether a build may include the index: there is a
// draft to include.
func indexIncludable(status string) bool {
	return status == indexDraft || status == indexReviewed
}

// requireIndexAccess is requirePassAccess plus the add-on entitlement: an
// admin always passes; a customer needs index_included on a live pass, else
// 402 with {"addon":"index"} so the factory page can offer the store button.
// Books without a project are admin-only, like convert.
func (s *Server) requireIndexAccess(w http.ResponseWriter, r *http.Request, book dbgen.Book) (*dbgen.Pass, bool, bool) {
	if !book.ProjectID.Valid {
		if !s.requireExeDevAdminAPI(w, r) {
			return nil, false, false
		}
		return nil, true, true
	}
	pass, isAdmin, ok := s.requirePassAccess(w, r, book.ProjectID.Int64)
	if !ok {
		return nil, false, false
	}
	if !isAdmin && !passIndexIncluded(pass) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusPaymentRequired)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": "the back-of-book index is a $100 add-on; buy it from the factory page",
			"addon": "index",
		})
		return nil, false, false
	}
	return pass, isAdmin, true
}

// ─── the document ───

// indexResponse is what GET/PUT return and what the factory page renders.
type indexResponse struct {
	BookID   int64          `json:"book_id"`
	Status   string         `json:"status"` // off | drafting | draft | reviewed | error
	Entitled bool           `json:"entitled"`
	Error    string         `json:"error,omitempty"`
	Index    *indexer.Index `json:"index,omitempty"`
}

func loadIndexDoc(book dbgen.Book) *indexer.Index {
	if !book.IndexJson.Valid || strings.TrimSpace(book.IndexJson.String) == "" {
		return nil
	}
	var idx indexer.Index
	if err := json.Unmarshal([]byte(book.IndexJson.String), &idx); err != nil {
		slog.Warn("bad index_json", "book_id", book.ID, "err", err)
		return nil
	}
	if idx.Entries == nil {
		idx.Entries = []indexer.Entry{}
	}
	return &idx
}

func (s *Server) storeIndexDoc(ctx context.Context, bid int64, idx *indexer.Index, status string) error {
	b, err := json.Marshal(idx)
	if err != nil {
		return err
	}
	return dbgen.New(s.DB).UpdateBookIndex(ctx, dbgen.UpdateBookIndexParams{
		IndexJson: sql.NullString{String: string(b), Valid: true}, IndexStatus: status, ID: bid,
	})
}

func (s *Server) indexResponseFor(book dbgen.Book, entitled bool) indexResponse {
	resp := indexResponse{BookID: book.ID, Status: book.IndexStatus, Entitled: entitled}
	if resp.Status == "" {
		resp.Status = indexOff
	}
	resp.Index = loadIndexDoc(book)
	if resp.Status == indexError && resp.Index != nil {
		for _, n := range resp.Index.Notes {
			if strings.HasPrefix(n, "draft failed: ") {
				resp.Error = strings.TrimPrefix(n, "draft failed: ")
			}
		}
	}
	return resp
}

// ─── handlers ───

// handleGetBookIndex — GET /api/books/{id}/index.
func (s *Server) handleGetBookIndex(w http.ResponseWriter, r *http.Request) {
	bid, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	book, err := dbgen.New(s.DB).GetBook(r.Context(), bid)
	if err != nil {
		jsonErr(w, "not found", 404)
		return
	}
	entitled := r.Header.Get("X-ExeDev-UserID") != ""
	if !book.ProjectID.Valid {
		if !s.requireExeDevAdminAPI(w, r) {
			return
		}
	} else {
		if !s.requireAuth(w, r, book.ProjectID.Int64) {
			return
		}
		if p := s.passForProject(r.Context(), book.ProjectID.Int64); passIndexIncluded(p) && passLive(*p) {
			entitled = true
		}
	}
	jsonOK(w, s.indexResponseFor(book, entitled))
}

// handleDraftBookIndex — POST /api/books/{id}/index/draft. Async: answers
// {"status":"drafting"} and the page polls GET …/index. One drafting run
// or build per book at a time (409), because both read the source and the
// draft costs money.
func (s *Server) handleDraftBookIndex(w http.ResponseWriter, r *http.Request) {
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
	if _, _, ok := s.requireIndexAccess(w, r, book); !ok {
		return
	}
	if book.IndexStatus == indexDrafting {
		jsonErr(w, "an index draft is already running for this book; wait for it to finish", http.StatusConflict)
		return
	}
	if book.Status == "converting" {
		jsonErr(w, "a build is running for this book; wait for it to finish", http.StatusConflict)
		return
	}
	if book.ProjectID.Valid {
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
	if err := q.UpdateBookIndexStatus(r.Context(), dbgen.UpdateBookIndexStatusParams{IndexStatus: indexDrafting, ID: bid}); err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	if book.ProjectID.Valid {
		s.factoryEventR(r, book.ProjectID.Int64, "index.draft.started", fmt.Sprintf("book %d", bid))
	}
	go s.runIndexDraft(bid, book)
	jsonOK(w, map[string]any{"status": indexDrafting, "book_id": bid,
		"status_url": fmt.Sprintf("/api/books/%d/index", bid)})
}

// handlePutBookIndex — PUT /api/books/{id}/index {"entries":[…]} (or a whole
// Index document; only entries are taken from it). Validated and tidied,
// anchors re-checked against the manuscript, stored as reviewed.
func (s *Server) handlePutBookIndex(w http.ResponseWriter, r *http.Request) {
	bid, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	book, err := dbgen.New(s.DB).GetBook(r.Context(), bid)
	if err != nil {
		jsonErr(w, "not found", 404)
		return
	}
	if _, _, ok := s.requireIndexAccess(w, r, book); !ok {
		return
	}
	if book.IndexStatus == indexDrafting {
		jsonErr(w, "an index draft is running; wait for it before editing", http.StatusConflict)
		return
	}
	var in struct {
		Entries []indexer.Entry `json:"entries"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<20)).Decode(&in); err != nil {
		jsonErr(w, "invalid index document: "+err.Error(), 400)
		return
	}
	if in.Entries == nil {
		jsonErr(w, "entries required", 400)
		return
	}
	entries, err := validateIndexEntries(in.Entries)
	if err != nil {
		jsonErr(w, err.Error(), 400)
		return
	}
	idx := loadIndexDoc(book)
	if idx == nil {
		idx = &indexer.Index{Book: book.Title, Generated: time.Now().UTC(), Model: "reviewed by hand"}
	}
	idx.Notes = nil
	idx.Entries = indexer.Tidy(entries, func(n string) { idx.Notes = append(idx.Notes, n) })
	idx.Unmatched = nil
	// Re-check anchors against the manuscript so the review table's flags
	// are true for edited fragments too. Pandoc takes a second; a failure
	// here is not the reviewer's problem.
	if src, _, err := s.indexSourceForBook(book); err == nil {
		_, rep := indexer.PlaceMarkers(src, idx)
		idx.Unmatched = rep.Unmatched
	} else {
		slog.Warn("index review: anchor re-check skipped", "book_id", bid, "err", err)
	}
	if err := s.storeIndexDoc(r.Context(), bid, idx, indexReviewed); err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	if book.ProjectID.Valid {
		s.factoryEventR(r, book.ProjectID.Int64, "index.reviewed", fmt.Sprintf("book %d: %d entries, %d unmatched anchors", bid, len(idx.Entries), len(idx.Unmatched)))
	}
	book.IndexJson = sql.NullString{String: "", Valid: false}
	book.IndexStatus = indexReviewed
	resp := s.indexResponseFor(book, true)
	resp.Index = idx
	jsonOK(w, resp)
}

// validateIndexEntries is the PUT contract: headings non-empty, anchors
// non-empty text, sane sizes. See-targets are resolved by Tidy (a dangling
// one is dropped with a note rather than rejected — the reviewer sees it).
func validateIndexEntries(in []indexer.Entry) ([]indexer.Entry, error) {
	if len(in) > indexMaxEntries {
		return nil, fmt.Errorf("too many entries (%d; the limit is %d)", len(in), indexMaxEntries)
	}
	out := make([]indexer.Entry, 0, len(in))
	for i, e := range in {
		e.Heading = clip(strings.Join(strings.Fields(e.Heading), " "), 200)
		e.Subheading = clip(strings.Join(strings.Fields(e.Subheading), " "), 200)
		e.See = clip(strings.Join(strings.Fields(e.See), " "), 200)
		if e.Heading == "" {
			return nil, fmt.Errorf("entry %d has an empty heading", i+1)
		}
		var sa []string
		for _, x := range e.SeeAlso {
			if x = clip(strings.Join(strings.Fields(x), " "), 200); x != "" {
				sa = append(sa, x)
			}
		}
		e.SeeAlso = sa
		var as []indexer.Anchor
		for _, a := range e.Anchors {
			a.Text = clip(strings.TrimSpace(a.Text), 400)
			if a.Text != "" {
				as = append(as, a)
			}
		}
		if as == nil {
			as = []indexer.Anchor{}
		}
		e.Anchors = as
		if len(e.Anchors) == 0 && e.See == "" && e.Subheading == "" {
			return nil, fmt.Errorf("entry %d (%q) has no anchors and no see-reference; delete it or give it one", i+1, e.Heading)
		}
		out = append(out, e)
	}
	return out, nil
}

// handleAdminGrantPassIndex — POST /api/admin/passes/{id}/index: the admin
// grant path for the add-on (mirrors handleAdminGrantPassBuilds; the ledger
// row has delta 0 and reason "index" so the movement is on the record).
func (s *Server) handleAdminGrantPassIndex(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		jsonErr(w, "invalid pass id", http.StatusBadRequest)
		return
	}
	q := dbgen.New(s.DB)
	pass, err := q.GetPass(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		jsonErr(w, "pass not found", http.StatusNotFound)
		return
	}
	if err != nil {
		jsonErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.grantPassIndex(r.Context(), q, pass.ID, "grant"); err != nil {
		jsonErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	slog.Info("admin: index add-on granted", "pass_id", id, "by", requestActor(r))
	jsonOK(w, map[string]any{"ok": true, "id": id, "index_included": true})
}

// grantPassIndex sets the entitlement and leaves a ledger row (delta 0).
func (s *Server) grantPassIndex(ctx context.Context, q *dbgen.Queries, passID int64, reason string) error {
	if err := q.SetPassIndexIncluded(ctx, passID); err != nil {
		return err
	}
	return q.CreatePassLedgerEntry(ctx, dbgen.CreatePassLedgerEntryParams{PassID: passID, Delta: 0, Reason: "index_" + reason})
}

// ─── the drafting run ───

// indexSourceForBook runs the pandoc half of the build (docx → Typst with
// the bundled filter) on the stored source, so the draft reads and anchors
// against the same text the build will set. Returns the Typst source and
// whether the spec says level-1 headings are parts.
func (s *Server) indexSourceForBook(book dbgen.Book) (string, bool, error) {
	if len(book.SourceData) == 0 {
		return "", false, errors.New("no source file")
	}
	tmpDir, err := os.MkdirTemp("", fmt.Sprintf("index-%d-*", book.ID))
	if err != nil {
		return "", false, err
	}
	defer os.RemoveAll(tmpDir)
	docxPath := filepath.Join(tmpDir, "input.docx")
	if err := os.WriteFile(docxPath, book.SourceData, 0644); err != nil {
		return "", false, err
	}
	typPath := filepath.Join(tmpDir, "book.typ")
	cmd := exec.Command("pandoc", "--from=docx+styles", docxPath, "--lua-filter="+typstFilterPath(),
		"--extract-media="+filepath.Join(tmpDir, "media"), "-t", "typst+smart", "-o", typPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", false, fmt.Errorf("pandoc: %s: %s", err, bytes.TrimSpace(out))
	}
	b, err := os.ReadFile(typPath)
	if err != nil {
		return "", false, err
	}
	return string(b), specHasParts(s.specMapForBook(book)), nil
}

var pdfPageRe = regexp.MustCompile(`/Type\s*/Page[^s]`)

// pdfPageCount counts page objects in a PDF (good enough for Typst output;
// 0 when there is no PDF).
func pdfPageCount(pdf []byte) int {
	if len(pdf) == 0 {
		return 0
	}
	return len(pdfPageRe.FindAllIndex(pdf, -1))
}

// indexPagesForBook: the set book's page count from its latest PDF, so the
// entry budget is measured rather than estimated from words.
func (s *Server) indexPagesForBook(ctx context.Context, bid int64) int {
	row, err := dbgen.New(s.DB).GetBookPDF(ctx, bid)
	if err != nil {
		return 0
	}
	return pdfPageCount(row.PdfData)
}

// indexClient is the gateway client for drafting; tests set it to a replay client.
func (s *Server) indexClient() *indexer.Client {
	if s.IndexClient != nil {
		return s.IndexClient
	}
	return indexer.NewClientFromEnv()
}

// runIndexDraft is the background half of POST …/index/draft.
func (s *Server) runIndexDraft(bid int64, book dbgen.Book) {
	ctx, cancel := context.WithTimeout(context.Background(), indexDraftTimeout)
	defer cancel()
	start := time.Now()
	fail := func(msg string) {
		slog.Error("index draft failed", "book_id", bid, "err", msg)
		idx := loadIndexDoc(book)
		if idx == nil {
			idx = &indexer.Index{Book: book.Title, Entries: []indexer.Entry{}}
		}
		idx.Notes = append(idx.Notes, "draft failed: "+clip(msg, 300))
		if err := s.storeIndexDoc(ctx, bid, idx, indexError); err != nil {
			slog.Error("index draft: record failure", "book_id", bid, "err", err)
		}
		if book.ProjectID.Valid {
			s.factoryEvent(book.ProjectID.Int64, "", "index.draft.failed", "factory", fmt.Sprintf("book %d: %s", bid, msg))
		}
	}
	src, parts, err := s.indexSourceForBook(book)
	if err != nil {
		fail(err.Error())
		return
	}
	title, _ := s.specTitleAuthor(book)
	if title == "" {
		title = book.Title
	}
	c := s.indexClient()
	c.Log = func(format string, a ...any) { slog.Info("index draft: "+fmt.Sprintf(format, a...), "book_id", bid) }
	idx, err := indexer.Draft(ctx, c, src, indexer.Options{
		Book: title, Pages: s.indexPagesForBook(ctx, bid), Parts: parts, Log: c.Log,
	})
	if err != nil {
		fail(err.Error())
		return
	}
	_, rep := indexer.PlaceMarkers(src, idx)
	idx.Unmatched = rep.Unmatched
	if err := s.storeIndexDoc(ctx, bid, idx, indexDraft); err != nil {
		fail("store draft: " + err.Error())
		return
	}
	detail := fmt.Sprintf("book %d: %d entries, %d anchors (%d unmatched), %d calls, ≈ $%.2f, %s",
		bid, len(idx.Entries), rep.Anchors, len(rep.Unmatched), idx.Usage.Calls, idx.CostUSD, time.Since(start).Round(time.Second))
	slog.Info("index drafted", "book_id", bid, "detail", detail)
	if book.ProjectID.Valid {
		s.factoryEvent(book.ProjectID.Int64, "", "index.drafted", "factory", detail)
	}
}

// ─── the build hook ───

// indexTypstConfig adds `index: true` to the spec's merged config (or makes
// one when the book has no spec) so book() appends index-page().
func indexTypstConfig(configOverride string) string {
	const open = "#let config = merge-config(("
	if i := strings.Index(configOverride, open); i >= 0 {
		at := i + len(open)
		return configOverride[:at] + "\n  index: true," + configOverride[at:]
	}
	return "\n// Back-of-book index add-on\n#let config = merge-config((\n  index: true,\n))\n"
}

// applyIndexMarkers places the stored index's markers into the build's Typst
// source. Unmatched anchors are reported (event + log), never fatal: the
// entry still prints with the locators that did match.
func (s *Server) applyIndexMarkers(bid int64, book dbgen.Book, typText string) string {
	idx := loadIndexDoc(book)
	if idx == nil || len(idx.Entries) == 0 {
		slog.Warn("index requested but no draft stored", "book_id", bid)
		return typText
	}
	marked, rep := indexer.PlaceMarkers(typText, idx)
	detail := fmt.Sprintf("book %d: %d entries, %d anchors placed of %d (%d fuzzy), %d cross-references",
		bid, rep.Entries, rep.Placed, rep.Anchors, rep.Fuzzy, rep.CrossRefs)
	if n := len(rep.Unmatched); n > 0 {
		var ex []string
		for i, u := range rep.Unmatched {
			if i == 3 {
				ex = append(ex, "…")
				break
			}
			ex = append(ex, fmt.Sprintf("%s: %q", u.Heading, clip(u.Text, 60)))
		}
		detail += fmt.Sprintf("; %d unmatched anchor(s) left out (%s)", n, strings.Join(ex, "; "))
	}
	slog.Info("index markers placed", "book_id", bid, "detail", detail)
	if book.ProjectID.Valid {
		s.factoryEvent(book.ProjectID.Int64, "", "index.placed", "factory", detail)
	}
	return marked
}

// readIndexFlag pulls "index": true out of a convert body without consuming
// the fields handleConvertBook parses itself.
func readIndexFlag(body []byte) bool {
	var v struct {
		Index bool `json:"index"`
	}
	_ = json.Unmarshal(body, &v)
	return v.Index
}

// drainBody is a tiny helper so handleConvertBook can read the body once.
func drainBody(r io.Reader, n int64) []byte {
	b, _ := io.ReadAll(io.LimitReader(r, n))
	return b
}
