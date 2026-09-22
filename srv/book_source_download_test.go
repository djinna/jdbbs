package srv

import (
	"context"
	"io"
	"strings"
	"testing"

	"srv.exe.dev/db/dbgen"
)

// The "before" file: /api/books/{id}/download/source serves the uploaded
// .docx under its original name (QC before/after pairs, 2026-09-22).
func TestDownloadBookSourceServesUpload(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	q := dbgen.New(s.DB)
	book, err := q.CreateBook(context.Background(), dbgen.CreateBookParams{
		Title: "Before File", Author: "A", SourceFilename: "Poems - Templated.docx", SourceData: []byte("PK-docx-bytes"),
	})
	if err != nil {
		t.Fatal(err)
	}

	resp := apiRequestAdmin(t, ts, "GET", "/api/books/"+itoa(book.ID)+"/download/source", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("status %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "PK-docx-bytes" {
		t.Errorf("body = %q", body)
	}
	if cd := resp.Header.Get("Content-Disposition"); !strings.Contains(cd, "Templated.docx") {
		t.Errorf("Content-Disposition = %q", cd)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "wordprocessingml") {
		t.Errorf("Content-Type = %q", ct)
	}

	// Anonymous: a book without a project is admin-only.
	resp = apiRequest(t, ts, "GET", "/api/books/"+itoa(book.ID)+"/download/source", nil)
	if resp.StatusCode == 200 {
		t.Errorf("anonymous source download should be refused, got 200")
	}
}
