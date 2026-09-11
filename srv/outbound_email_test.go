package srv

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestOutboundEmailLog: every send attempt — success and failure — lands in
// outbound_email, and the admin report endpoint returns it with filters.
func TestOutboundEmailLog(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	fail := false
	mailAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if fail {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer mailAPI.Close()
	s.Email = &EmailConfig{APIKey: "test", InboxID: "test@example.com", APIBase: mailAPI.URL}

	// Success.
	err := s.mail(mailMeta{Kind: mailKindAnnouncement, RefType: "registration", RefID: "7", TriggeredBy: "j@djinna.com"},
		[]string{"ada@example.com"}, []string{"cc@example.com"}, "Hello", "text", "<p>html</p>")
	if err != nil {
		t.Fatalf("mail: %v", err)
	}
	// Failure.
	fail = true
	err = s.mail(mailMeta{Kind: mailKindFactoryPass, RefType: "pass", RefID: "3"},
		[]string{"bob@example.com"}, nil, "Your pass", "text", "")
	if err == nil {
		t.Fatal("expected transport error")
	}

	// Unauthenticated → 401.
	resp, _ := http.Get(ts.URL + "/api/admin/email")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("anon: want 401, got %d", resp.StatusCode)
	}

	get := func(q string) map[string]any {
		resp := apiRequestAdmin(t, ts, "GET", "/api/admin/email"+q, nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("GET %s: %d", q, resp.StatusCode)
		}
		var m map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
			t.Fatal(err)
		}
		return m
	}

	all := get("")
	emails := all["emails"].([]any)
	if len(emails) != 2 {
		t.Fatalf("want 2 log rows, got %d", len(emails))
	}
	newest := emails[0].(map[string]any)
	if newest["kind"] != mailKindFactoryPass || newest["status_code"] != float64(502) || newest["error"] == "" {
		t.Fatalf("failure row not recorded correctly: %#v", newest)
	}
	older := emails[1].(map[string]any)
	if older["to"] != "ada@example.com" || older["cc"] != "cc@example.com" || older["status_code"] != float64(200) || older["triggered_by"] != "j@djinna.com" || older["ref_id"] != "7" {
		t.Fatalf("success row not recorded correctly: %#v", older)
	}
	kinds := all["kinds"].(map[string]any)
	if kinds[mailKindAnnouncement] != float64(1) || kinds[mailKindFactoryPass] != float64(1) {
		t.Fatalf("kind counts: %#v", kinds)
	}

	if n := len(get("?kind=announcement")["emails"].([]any)); n != 1 {
		t.Fatalf("kind filter: want 1, got %d", n)
	}
	if n := len(get("?to=bob")["emails"].([]any)); n != 1 {
		t.Fatalf("to filter: want 1, got %d", n)
	}
	if n := len(get("?failed=1")["emails"].([]any)); n != 1 {
		t.Fatalf("failed filter: want 1, got %d", n)
	}
	if n := len(get("?ref_type=registration&ref_id=7")["emails"].([]any)); n != 1 {
		t.Fatalf("ref filter: want 1, got %d", n)
	}
}

// TestRegistrationSendsAreLogged: the public registration path writes two
// log rows (organizer alert + applicant confirmation) tagged to the row id.
func TestRegistrationSendsAreLogged(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()
	mailAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))
	defer mailAPI.Close()
	s.Email = &EmailConfig{APIKey: "test", InboxID: "test@example.com", APIBase: mailAPI.URL}

	s.sendRegistrationEmails(registrationInput{Name: "Ada Lovelace", Email: "ada@example.com", Material: "a book"}, 42)

	var n int
	if err := s.DB.QueryRow(`SELECT COUNT(*) FROM outbound_email WHERE ref_type='registration' AND ref_id='42' AND status_code=200`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("want 2 logged registration sends, got %d", n)
	}
	_ = ts
}
