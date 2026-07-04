package srv

import (
	"context"
	"database/sql"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"testing"
	"time"

	"srv.exe.dev/db"
	"srv.exe.dev/db/dbgen"
)

// TestBookOutputWritePathServesNewest exercises the post-018 "conversion
// complete" write path: outputs land only in book_outputs, the books row is
// flipped to ready, and reads (query + HTTP download) serve the newest row.
func TestBookOutputWritePathServesNewest(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	ctx := context.Background()
	q := dbgen.New(s.DB)

	book, err := q.CreateBook(ctx, dbgen.CreateBookParams{
		Title: "Write Path", Author: "A", SourceFilename: "wp.docx", SourceData: []byte("docx"),
	})
	if err != nil {
		t.Fatalf("create book: %v", err)
	}

	// No output yet: the row still comes back (book exists) with nil data,
	// which is what keeps the "PDF not generated yet" 404 working.
	pdf, err := q.GetBookPDF(ctx, book.ID)
	if err != nil {
		t.Fatalf("get pdf (no output): %v", err)
	}
	if pdf.PdfData != nil {
		t.Fatalf("expected nil pdf data before any output, got %d bytes", len(pdf.PdfData))
	}
	resp := apiRequestAdmin(t, ts, "GET", "/api/books/"+itoa(book.ID)+"/download/pdf", nil)
	if resp.StatusCode != 404 {
		t.Errorf("expected 404 before any output, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// The conversion-complete write path: two compiles back to back.
	for _, data := range []string{"old-pdf", "new-pdf"} {
		if _, err := q.CreateBookOutput(ctx, dbgen.CreateBookOutputParams{
			BookID: book.ID, OutputFormat: "pdf", OutputData: []byte(data), SourceFilename: "wp.docx",
		}); err != nil {
			t.Fatalf("create output %q: %v", data, err)
		}
	}
	if err := q.UpdateBookStatusReady(ctx, book.ID); err != nil {
		t.Fatalf("mark ready: %v", err)
	}

	pdf, err = q.GetBookPDF(ctx, book.ID)
	if err != nil {
		t.Fatalf("get pdf: %v", err)
	}
	if string(pdf.PdfData) != "new-pdf" {
		t.Errorf("expected newest output %q, got %q", "new-pdf", string(pdf.PdfData))
	}
	got, err := q.GetBook(ctx, book.ID)
	if err != nil {
		t.Fatalf("get book: %v", err)
	}
	if got.Status != "ready" {
		t.Errorf("expected status ready, got %q", got.Status)
	}

	resp = apiRequestAdmin(t, ts, "GET", "/api/books/"+itoa(book.ID)+"/download/pdf", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 download, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if string(body) != "new-pdf" {
		t.Errorf("download served %q, want newest %q", string(body), "new-pdf")
	}
}

// TestBlobDedupeMigrationBackfill proves the 018 upgrade path: a database on
// the pre-018 schema with blobs on the books row gets them backfilled into
// book_outputs (unless a history row already exists), and the columns are
// dropped.
func TestBlobDedupeMigrationBackfill(t *testing.T) {
	mdb, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer mdb.Close()

	files := migrationFiles(t)
	applyMigrations(t, mdb, files, 1, 16)

	// Pre-014 book: blobs only on the books row.
	if _, err := mdb.Exec(
		`INSERT INTO books (title, author, pdf_data, epub_data, status, updated_at)
		 VALUES ('Legacy', 'A', ?, ?, 'ready', '2026-01-02 03:04:05')`,
		[]byte("legacy-pdf"), []byte("legacy-epub"),
	); err != nil {
		t.Fatalf("insert legacy book: %v", err)
	}
	// Post-014 book: blob duplicated on the row AND already in book_outputs.
	if _, err := mdb.Exec(
		`INSERT INTO books (id, title, author, pdf_data, status) VALUES (2, 'Modern', 'B', ?, 'ready')`,
		[]byte("modern-pdf-stale"),
	); err != nil {
		t.Fatalf("insert modern book: %v", err)
	}
	if _, err := mdb.Exec(
		`INSERT INTO book_outputs (book_id, output_format, output_data) VALUES (2, 'pdf', ?)`,
		[]byte("modern-pdf-history"),
	); err != nil {
		t.Fatalf("insert modern output: %v", err)
	}

	applyMigrations(t, mdb, files, 17, 18)

	// Legacy blobs are backfilled, created_at inherited from updated_at.
	for _, tc := range []struct{ format, want string }{
		{"pdf", "legacy-pdf"},
		{"epub", "legacy-epub"},
	} {
		var data []byte
		var createdAt time.Time
		err := mdb.QueryRow(
			`SELECT output_data, created_at FROM book_outputs WHERE book_id = 1 AND output_format = ?`,
			tc.format,
		).Scan(&data, &createdAt)
		if err != nil {
			t.Fatalf("backfilled %s row: %v", tc.format, err)
		}
		if string(data) != tc.want {
			t.Errorf("%s backfill: got %q, want %q", tc.format, string(data), tc.want)
		}
		if want := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC); !createdAt.Equal(want) {
			t.Errorf("%s backfill created_at: got %v, want updated_at %v", tc.format, createdAt, want)
		}
	}

	// The book with existing history is NOT double-inserted, and the history
	// row (not the stale books blob) is what remains.
	var n int
	var data []byte
	if err := mdb.QueryRow(
		`SELECT count(*) FROM book_outputs WHERE book_id = 2 AND output_format = 'pdf'`,
	).Scan(&n); err != nil {
		t.Fatalf("count modern outputs: %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1 output row for book 2, got %d", n)
	}
	if err := mdb.QueryRow(
		`SELECT output_data FROM book_outputs WHERE book_id = 2 AND output_format = 'pdf'`,
	).Scan(&data); err != nil {
		t.Fatalf("modern output: %v", err)
	}
	if string(data) != "modern-pdf-history" {
		t.Errorf("book 2 output: got %q, want existing history row", string(data))
	}

	// Columns are gone.
	rows, err := mdb.Query(`SELECT name FROM pragma_table_info('books')`)
	if err != nil {
		t.Fatalf("table_info: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan column name: %v", err)
		}
		if name == "pdf_data" || name == "epub_data" {
			t.Errorf("column %s still present after 018", name)
		}
	}

	// 017 landed too.
	var idx string
	if err := mdb.QueryRow(
		`SELECT name FROM sqlite_master WHERE type='index' AND name='idx_books_project_id'`,
	).Scan(&idx); err != nil {
		t.Errorf("idx_books_project_id missing after 017: %v", err)
	}
}

// migrationFiles returns the repo's migration files sorted in apply order.
// Tests run with CWD = repo/srv, so the migrations live one level up.
func migrationFiles(t *testing.T) []string {
	t.Helper()
	dir := filepath.Join("..", "db", "migrations")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read migrations dir %s: %v", dir, err)
	}
	pat := regexp.MustCompile(`^\d{3}-.*\.sql$`)
	var files []string
	for _, e := range entries {
		if !e.IsDir() && pat.MatchString(e.Name()) {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(files)
	return files
}

// applyMigrations executes the migration files numbered from..to inclusive.
func applyMigrations(t *testing.T, mdb *sql.DB, files []string, from, to int) {
	t.Helper()
	for _, f := range files {
		n, err := strconv.Atoi(filepath.Base(f)[:3])
		if err != nil {
			t.Fatalf("parse migration number %s: %v", f, err)
		}
		if n < from || n > to {
			continue
		}
		content, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		if _, err := mdb.Exec(string(content)); err != nil {
			t.Fatalf("apply %s: %v", f, err)
		}
	}
}
