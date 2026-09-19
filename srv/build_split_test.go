package srv

import (
	"net/http"
	"strings"
	"testing"
	"time"

	dbgen "srv.exe.dev/db/dbgen"
)

// TestEPUBOnlyBuildIsAFreeProof: an EPUB-only build (format=epub) is a
// proof by construction (0.28) — free, never debited — and the book keeps
// only the newest epubOutputsKept EPUB outputs.
func TestEPUBOnlyBuildIsAFreeProof(t *testing.T) {
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

	// Plenty of credits so the pruning loop below can run past the cap.
	if _, err := s.DB.Exec(`UPDATE passes SET builds_included = ? WHERE id = ?`, epubOutputsKept+2, pass.ID); err != nil {
		t.Fatalf("set credits: %v", err)
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

	// Since 0.28 an EPUB-only build is a proof by construction: free, never
	// debited. The pruning cap still applies.
	if n := countLedger(t, s, pass.ID, "build"); n != 0 {
		t.Fatalf("epub builds wrote %d debit rows, want 0 (epub-only builds are free proofs)", n)
	}
	var kept int
	if err := s.DB.QueryRow(`SELECT COUNT(*) FROM book_outputs WHERE book_id = ? AND output_format = 'epub'`, bookID).Scan(&kept); err != nil {
		t.Fatal(err)
	}
	if kept != epubOutputsKept {
		t.Fatalf("kept %d epub outputs, want %d", kept, epubOutputsKept)
	}

	// With every credit spent, an EPUB-only build still runs (it is free);
	// a final is the thing that gets a 402.
	if _, err := s.DB.Exec(`UPDATE passes SET builds_used = builds_included WHERE id = ?`, pass.ID); err != nil {
		t.Fatal(err)
	}
	post := func(body string) int {
		req, _ := http.NewRequest("POST", ts.URL+"/api/books/"+itoa(bookID)+"/convert", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(cookie)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		return resp.StatusCode
	}
	if code := post(`{"format":"both","kind":"final"}`); code != http.StatusPaymentRequired {
		t.Fatalf("final with no credits: expected 402, got %d", code)
	}
	if code := post(`{"format":"epub"}`); code != 200 {
		t.Fatalf("epub build with no credits: expected 200 (free proof), got %d", code)
	}
}
