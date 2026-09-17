package srv

import (
	"net/http"
	"strings"
	"testing"
	"time"

	dbgen "srv.exe.dev/db/dbgen"
)

// TestEPUBBuildIsFreeAndUnlimited: an EPUB-only build (format=epub) never
// debits a credit, runs even when the pass has no print builds left, and
// keeps only the newest epubOutputsKept EPUB outputs.
func TestEPUBBuildIsFreeAndUnlimited(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	pass, password, clientSlug, _ := grantedPass(t, s, ts, "Ebook Author", "Free EPUBs")
	cookie := clientCookie(t, ts, clientSlug, password)

	resp := uploadBook(t, ts, itoa(pass.ProjectID), cookie, false)
	if resp.StatusCode != 201 {
		t.Fatalf("upload: expected 201, got %d", resp.StatusCode)
	}
	var created map[string]any
	decodeJSON(t, resp, &created)
	bookID := int64(created["id"].(float64))

	// No print builds left — EPUB must still be allowed.
	if _, err := s.DB.Exec(`UPDATE passes SET builds_used = builds_included WHERE id = ?`, pass.ID); err != nil {
		t.Fatalf("exhaust credits: %v", err)
	}

	// Fake EPUB stage: store a tiny output like the real one does.
	q := dbgen.New(s.DB)
	s.epubRunner = func(bid int64, book dbgen.Book) error {
		_, err := q.CreateBookOutput(t.Context(), dbgen.CreateBookOutputParams{
			BookID: bid, OutputFormat: "epub", OutputData: []byte("PK"), SourceFilename: book.SourceFilename,
		})
		return err
	}

	for i := 0; i < epubOutputsKept+2; i++ {
		req, err := http.NewRequest("POST", ts.URL+"/api/books/"+itoa(bookID)+"/convert", strings.NewReader(`{"format":"epub"}`))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(cookie)
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != 200 {
			t.Fatalf("epub build %d: expected 200, got %d", i, resp.StatusCode)
		}
		resp.Body.Close()
		// Wait for the background build to land.
		deadline := time.Now().Add(5 * time.Second)
		for {
			var status string
			if err := s.DB.QueryRow(`SELECT status FROM books WHERE id = ?`, bookID).Scan(&status); err != nil {
				t.Fatal(err)
			}
			if status == "ready" {
				break
			}
			if time.Now().After(deadline) {
				t.Fatalf("epub build %d never became ready (status %q)", i, status)
			}
			time.Sleep(20 * time.Millisecond)
		}
		_, _ = s.DB.Exec(`UPDATE books SET status = 'uploaded' WHERE id = ?`, bookID)
	}

	if n := countLedger(t, s, pass.ID, "build"); n != 0 {
		t.Fatalf("epub builds wrote %d debit rows, want 0", n)
	}
	var kept int
	if err := s.DB.QueryRow(`SELECT COUNT(*) FROM book_outputs WHERE book_id = ? AND output_format = 'epub'`, bookID).Scan(&kept); err != nil {
		t.Fatal(err)
	}
	if kept != epubOutputsKept {
		t.Fatalf("kept %d epub outputs, want %d", kept, epubOutputsKept)
	}

	// A print build with no credits is still a 402.
	req, _ := http.NewRequest("POST", ts.URL+"/api/books/"+itoa(bookID)+"/convert", strings.NewReader(`{"format":"pdf"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusPaymentRequired {
		t.Fatalf("pdf build with no credits: expected 402, got %d", resp.StatusCode)
	}
}
