package srv

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func postNit(t *testing.T, url string, body map[string]string) (int, string) {
	t.Helper()
	b, _ := json.Marshal(body)
	resp, err := http.Post(url+"/api/nits", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(out)
}

func TestNitsReportAndInbox(t *testing.T) {
	var mu sync.Mutex
	var inbox []string
	rp := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in struct{ Text string }
		json.NewDecoder(r.Body).Decode(&in)
		mu.Lock()
		inbox = append(inbox, in.Text)
		mu.Unlock()
		w.Write([]byte(`{"ok":true}`))
	}))
	defer rp.Close()
	t.Setenv("PRODCAL_NIT_INBOX_URL", rp.URL+"/add")

	s, ts, done := testServer(t)
	defer done()
	s.secret = []byte("test-secret")

	// Honeypot: looks accepted, stores nothing.
	if code, _ := postNit(t, ts.URL, map[string]string{"page": "/help/inspect", "comment": "buy pills", "website": "http://spam"}); code != 200 {
		t.Fatalf("honeypot answered %d", code)
	}
	// Empty and bad email are refused.
	if code, _ := postNit(t, ts.URL, map[string]string{"page": "/x"}); code != 400 {
		t.Errorf("empty nit = %d", code)
	}
	if code, _ := postNit(t, ts.URL, map[string]string{"page": "/x", "comment": "hi", "email": "nope"}); code != 400 {
		t.Errorf("bad email = %d", code)
	}
	code, body := postNit(t, ts.URL, map[string]string{"page": "/help/inspect#fixes", "selection": "the the", "comment": "doubled word\n[x] sneaky", "email": "reader@example.test", "kind": "app"})
	if code != 200 || !strings.Contains(body, `"id":`) {
		t.Fatalf("nit: %d %s", code, body)
	}
	var n int
	var kind, sel string
	s.DB.QueryRow(`SELECT COUNT(*), MAX(kind), MAX(selection) FROM nits`).Scan(&n, &kind, &sel)
	if n != 1 || kind != "help" || sel != "the the" {
		t.Errorf("stored %d nits, kind %q sel %q (want 1, help — /help pages are always help)", n, kind, sel)
	}
	deadline := time.Now().Add(2 * time.Second)
	for {
		var ok int
		s.DB.QueryRow(`SELECT inbox_ok FROM nits`).Scan(&ok)
		if ok == 1 || time.Now().After(deadline) {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	mu.Lock()
	got := strings.Join(inbox, "|")
	mu.Unlock()
	if !strings.Contains(got, "[nit·help #") || !strings.Contains(got, "“the the”") || strings.Contains(got, "\n") || strings.Contains(got, "[x]") {
		t.Errorf("inbox row: %q", got)
	}

	// Admin list is admin-only.
	if code, _, _ := helpGet(t, ts.URL+"/api/admin/nits", false); code == 200 {
		t.Errorf("nit list open to the public")
	}
	if code, body, _ := helpGet(t, ts.URL+"/api/admin/nits", true); code != 200 || !strings.Contains(body, "doubled word") {
		t.Errorf("admin nit list: %d", code)
	}

	// Per-IP rate limit: 5 per window (one already used above).
	var last int
	for i := 0; i < 6; i++ {
		last, _ = postNit(t, ts.URL, map[string]string{"page": "/", "comment": "again"})
	}
	if last != http.StatusTooManyRequests {
		t.Errorf("rate limit: last status %d, want 429", last)
	}
}
