package srv

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

// noRedirect returns a client that surfaces 302s instead of following them.
func noRedirect() *http.Client {
	return &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
}

// TestLoginLinkIssueAndRedeem: POST /api/public/login-link with the pass
// email mails a link; GET /auth/link sets the same client cookie a password
// login sets and 302s to the portal; the second use is refused.
func TestLoginLinkIssueAndRedeem(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()
	s.BaseURL = ts.URL

	var mu sync.Mutex
	var sent []string
	mailAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		mu.Lock()
		sent = append(sent, string(b))
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer mailAPI.Close()
	s.Email = &EmailConfig{APIKey: "test", InboxID: "test@example.com", APIBase: mailAPI.URL}

	pass, _, slug, _ := grantedPass(t, s, ts, "Ada Lovelace", "Notes on the Engine")
	_ = pass

	// Wrong address: 200 ok, no row, no mail.
	resp := apiRequest(t, ts, "POST", "/api/public/login-link", map[string]string{"client": slug, "email": "stranger@example.com"})
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("wrong email: want 200, got %d", resp.StatusCode)
	}
	if n := countRows(t, s, `SELECT COUNT(*) FROM login_links`); n != 0 {
		t.Fatalf("wrong email created %d rows", n)
	}
	// Unknown client: also 200, nothing created.
	resp = apiRequest(t, ts, "POST", "/api/public/login-link", map[string]string{"client": "nobody", "email": "customer@example.com"})
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("unknown client: want 200, got %d", resp.StatusCode)
	}

	// Right address (case-insensitive): a row and a mail carrying the link.
	resp = apiRequest(t, ts, "POST", "/api/public/login-link", map[string]string{"client": slug, "email": "  Customer@Example.com "})
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("issue: want 200, got %d", resp.StatusCode)
	}
	if n := countRows(t, s, `SELECT COUNT(*) FROM login_links WHERE client_slug = ? AND used_at IS NULL`, slug); n != 1 {
		t.Fatalf("want 1 live link row, got %d", n)
	}
	// The fulfilment mail from grantedPass is in the box too; find ours.
	var body string
	for i := 0; i < 60 && body == ""; i++ {
		mu.Lock()
		for _, m := range sent {
			if strings.Contains(m, "Your sign-in link for Ada Lovelace") {
				body = m
			}
		}
		mu.Unlock()
		if body == "" {
			time.Sleep(50 * time.Millisecond)
		}
	}
	if body == "" {
		t.Fatal("no sign-in mail sent for a matching address")
	}
	m := regexp.MustCompile(`/auth/link\?t=([A-Za-z0-9_-]{40,})`).FindStringSubmatch(body)
	if m == nil {
		t.Fatalf("mail carries no link: %.300s", body)
	}
	token := m[1]
	// The plaintext token must not be in the DB.
	if n := countRows(t, s, `SELECT COUNT(*) FROM login_links WHERE token_hash = ?`, token); n != 0 {
		t.Fatal("plaintext token stored")
	}
	if n := countRows(t, s, `SELECT COUNT(*) FROM login_links WHERE token_hash = ?`, loginLinkHash(token)); n != 1 {
		t.Fatal("hashed token not stored")
	}

	// Redeem: cookie + 302 to /{client}/.
	resp, err := noRedirect().Get(ts.URL + "/auth/link?t=" + token)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("redeem: want 302, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/"+slug+"/" {
		t.Fatalf("redeem: want Location /%s/, got %q", slug, loc)
	}
	var cookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "prodcal_client_"+slug {
			cookie = c
		}
	}
	if cookie == nil {
		t.Fatal("redeem set no client cookie")
	}
	if !cookie.HttpOnly || cookie.Path != "/" || cookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("cookie flags differ from password login: %+v", cookie)
	}
	// The cookie is the real thing: the client info endpoint reports authenticated.
	req, _ := http.NewRequest("GET", ts.URL+"/api/clients/"+slug, nil)
	req.AddCookie(cookie)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var info map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if info["authenticated"] != true {
		t.Fatalf("cookie from link does not authenticate: %v", info)
	}

	// Second use: 401 + the expired page pointing back at the portal.
	resp, err = noRedirect().Get(ts.URL + "/auth/link?t=" + token)
	if err != nil {
		t.Fatal(err)
	}
	page, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("second use: want 401, got %d", resp.StatusCode)
	}
	if !strings.Contains(string(page), "has expired") || !strings.Contains(string(page), `href="/`+slug+`/"`) {
		t.Errorf("expired page missing copy or portal link: %.300s", page)
	}
	if len(resp.Cookies()) != 0 {
		t.Error("second use set a cookie")
	}
}

// TestLoginLinkExpiredAndRateLimit: an expired link is refused; more than
// three live links per 15 minutes are silently not issued.
func TestLoginLinkExpiredAndRateLimit(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()
	s.BaseURL = ts.URL
	_, _, slug, _ := grantedPass(t, s, ts, "Bob Builder", "Bricks")

	token, ok, err := s.issueLoginLink(t.Context(), slug, "customer@example.com", "127.0.0.1")
	if err != nil || !ok {
		t.Fatalf("issue: ok=%v err=%v", ok, err)
	}
	// Age it past the TTL.
	if _, err := s.DB.Exec(`UPDATE login_links SET expires_at = ? WHERE token_hash = ?`,
		time.Now().UTC().Add(-time.Minute).Format(loginLinkTimeLayout), loginLinkHash(token)); err != nil {
		t.Fatal(err)
	}
	resp, err := noRedirect().Get(ts.URL + "/auth/link?t=" + token)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expired: want 401, got %d", resp.StatusCode)
	}
	if n := countRows(t, s, `SELECT COUNT(*) FROM login_links WHERE used_at IS NOT NULL`); n != 0 {
		t.Fatal("expired link was marked used")
	}
	// Garbage token → 401, no crash.
	resp, _ = noRedirect().Get(ts.URL + "/auth/link?t=not-a-token")
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("garbage: want 401, got %d", resp.StatusCode)
	}

	// Rate limit: the expired one still counts as unused-in-window; two more
	// succeed, the fourth is refused.
	for i := 0; i < 2; i++ {
		if _, ok, err := s.issueLoginLink(t.Context(), slug, "customer@example.com", ""); err != nil || !ok {
			t.Fatalf("issue %d: ok=%v err=%v", i+2, ok, err)
		}
	}
	if _, ok, err := s.issueLoginLink(t.Context(), slug, "customer@example.com", ""); err != nil || ok {
		t.Fatalf("fourth issue should be rate-limited: ok=%v err=%v", ok, err)
	}
	// Through the endpoint it is still a 200 and no new row.
	before := countRows(t, s, `SELECT COUNT(*) FROM login_links`)
	resp = apiRequest(t, ts, "POST", "/api/public/login-link", map[string]string{"client": slug, "email": "customer@example.com"})
	resp.Body.Close()
	if resp.StatusCode != 200 || countRows(t, s, `SELECT COUNT(*) FROM login_links`) != before {
		t.Fatal("rate-limited request should be 200 with no new row")
	}
}

// TestLoginLinkClientEmailColumn: a legacy client with no pass but an email
// on the client row can also get a link.
func TestLoginLinkClientEmailColumn(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()
	hash, _ := hashPassword("pw")
	if _, err := s.DB.Exec(`INSERT INTO clients (slug, name, password_hash, email) VALUES ('vgr', 'VGR', ?, 'V@Example.com')`, hash); err != nil {
		t.Fatal(err)
	}
	ok, err := s.loginLinkEmailMatches(t.Context(), "vgr", "v@example.com")
	if err != nil || !ok {
		t.Fatalf("client email should match: ok=%v err=%v", ok, err)
	}
	ok, _ = s.loginLinkEmailMatches(t.Context(), "vgr", "other@example.com")
	if ok {
		t.Fatal("unrelated address matched")
	}
	_ = ts
}

// A defensive password reset must also kill unredeemed sign-in links;
// otherwise a link captured before the reset restores access afterwards.
func TestLoginLinkInvalidatedByPasswordReset(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()
	s.BaseURL = ts.URL
	mailAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))
	defer mailAPI.Close()
	s.Email = &EmailConfig{APIKey: "test", InboxID: "test@example.com", APIBase: mailAPI.URL}
	_, _, slug, _ := grantedPass(t, s, ts, "Ada Lovelace", "Notes on the Engine")

	token, ok, err := s.issueLoginLink(t.Context(), slug, "customer@example.com", "127.0.0.1")
	if err != nil || !ok {
		t.Fatalf("issue link: ok=%v err=%v", ok, err)
	}
	resp := apiRequestAdmin(t, ts, "POST", "/api/admin/clients/"+slug+"/password", nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("reset: %d", resp.StatusCode)
	}
	resp, err = noRedirect().Get(ts.URL + "/auth/link?t=" + token)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("pre-reset link redeemed after reset: %d", resp.StatusCode)
	}
	if len(resp.Cookies()) != 0 {
		t.Fatal("pre-reset link issued a cookie after reset")
	}
}
