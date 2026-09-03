package srv

import (
	"bytes"
	"database/sql"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"srv.exe.dev/db/dbgen"
)

// ─── helpers ───

// seedCoupon inserts a coupon directly. expiresAt of the zero time means
// "no expiry".
func seedCoupon(t *testing.T, s *Server, code string, expiresAt time.Time) int64 {
	t.Helper()
	var exp sql.NullTime
	if !expiresAt.IsZero() {
		exp = sql.NullTime{Time: expiresAt, Valid: true}
	}
	q := dbgen.New(s.DB)
	c, err := q.CreateCoupon(t.Context(), dbgen.CreateCouponParams{
		Code:           code,
		Sku:            passSKU,
		MaxRedemptions: 1,
		ExpiresAt:      exp,
	})
	if err != nil {
		t.Fatalf("seed coupon %s: %v", code, err)
	}
	return c.ID
}

// redeem posts the public redemption form.
func redeem(t *testing.T, ts *httptest.Server, body any) *http.Response {
	t.Helper()
	return apiRequest(t, ts, "POST", "/api/public/redeem", body)
}

// uploadBook posts a multipart upload, optionally with a client cookie and/or
// the admin header, and returns the response.
func uploadBook(t *testing.T, ts *httptest.Server, projectID string, cookie *http.Cookie, admin bool) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("title", "Test Manuscript")
	_ = mw.WriteField("author", "A. Tester")
	if projectID != "" {
		_ = mw.WriteField("project_id", projectID)
	}
	fw, err := mw.CreateFormFile("file", "manuscript.docx")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := fw.Write([]byte("not-a-real-docx")); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	_ = mw.Close()

	req, err := http.NewRequest("POST", ts.URL+"/api/books/upload", &buf)
	if err != nil {
		t.Fatalf("create upload request: %v", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	if cookie != nil {
		req.AddCookie(cookie)
	}
	if admin {
		req.Header.Set("X-ExeDev-UserID", "test-admin")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do upload: %v", err)
	}
	return resp
}

// clientCookie signs in as a client with the given password and returns the
// session cookie.
func clientCookie(t *testing.T, ts *httptest.Server, slug, password string) *http.Cookie {
	t.Helper()
	resp := apiRequest(t, ts, "POST", "/api/clients/"+slug+"/verify",
		map[string]string{"password": password})
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("client verify %s: expected 200, got %d", slug, resp.StatusCode)
	}
	for _, c := range resp.Cookies() {
		if c.Name == "prodcal_client_"+slug {
			return c
		}
	}
	t.Fatalf("no client cookie for %s", slug)
	return nil
}

// grantedPass fulfils a pass through the admin endpoint (same fulfillPass as
// redemption) and returns the pass row plus the generated client password.
func grantedPass(t *testing.T, s *Server, ts *httptest.Server, name, title string) (dbgen.Pass, string, string, string) {
	t.Helper()
	resp := apiRequestAdmin(t, ts, "POST", "/api/admin/passes", map[string]any{
		"name": name, "email": "customer@example.com", "title": title, "author": name,
	})
	if resp.StatusCode != 201 {
		t.Fatalf("admin create pass: expected 201, got %d", resp.StatusCode)
	}
	var body map[string]any
	decodeJSON(t, resp, &body)
	pid := int64(body["project_id"].(float64))
	q := dbgen.New(s.DB)
	pass, err := q.GetPassByProject(t.Context(), pid)
	if err != nil {
		t.Fatalf("load granted pass: %v", err)
	}
	return pass, body["password"].(string), body["client_slug"].(string), body["project_slug"].(string)
}

// countLedger counts ledger rows for a pass with a given reason.
func countLedger(t *testing.T, s *Server, passID int64, reason string) int {
	t.Helper()
	return countRows(t, s, `SELECT COUNT(*) FROM pass_ledger WHERE pass_id = ? AND reason = ?`, passID, reason)
}

// waitForLedger waits (briefly) for a ledger row with the given reason to
// appear — the refund is written by the background conversion goroutine.
func waitForLedger(t *testing.T, s *Server, passID int64, reason string) {
	t.Helper()
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		if countLedger(t, s, passID, reason) > 0 {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("no %q ledger row for pass %d within the deadline", reason, passID)
}

// ─── redemption ───

// TestRedeemCreatesClientProjectAndPass: the happy path builds the whole
// customer world in one shot — client (with a password), project, pass with
// the included builds — and burns the coupon.
func TestRedeemCreatesClientProjectAndPass(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	couponID := seedCoupon(t, s, "PYB-AAAA-BBBB", time.Now().Add(24*time.Hour))

	resp := redeem(t, ts, map[string]string{
		"code": "pyb-aaaa-bbbb", // case-insensitive on purpose
		"name": "Ada Lovelace", "email": "ada@example.com",
		"title": "Notes on the Analytical Engine", "author": "Ada Lovelace",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("redeem: expected 200, got %d", resp.StatusCode)
	}
	var body map[string]any
	decodeJSON(t, resp, &body)
	if body["ok"] != true {
		t.Fatalf("redeem: expected ok=true, got %+v", body)
	}
	clientSlug, _ := body["client_slug"].(string)
	projectSlug, _ := body["project_slug"].(string)
	if clientSlug != "ada-lovelace" {
		t.Errorf("client_slug = %q, want \"ada-lovelace\"", clientSlug)
	}
	if projectSlug != "notes-on-the-analytical" {
		t.Errorf("project_slug = %q, want \"notes-on-the-analytical\" (24-char cap)", projectSlug)
	}
	portal, _ := body["portal_url"].(string)
	if !strings.HasSuffix(portal, "/"+clientSlug+"/"+projectSlug+"/factory/") {
		t.Errorf("portal_url = %q, want it to end in the factory page path", portal)
	}

	// Client exists and is password-protected (the password was emailed, not
	// stored in plaintext).
	var passwordHash string
	if err := s.DB.QueryRow(`SELECT password_hash FROM clients WHERE slug = ?`, clientSlug).Scan(&passwordHash); err != nil {
		t.Fatalf("client row missing: %v", err)
	}
	if !strings.HasPrefix(passwordHash, "$2") {
		t.Errorf("client password_hash = %q, want a bcrypt hash", passwordHash)
	}

	// Project + pass.
	q := dbgen.New(s.DB)
	project, err := q.GetProjectByPath(t.Context(), dbgen.GetProjectByPathParams{
		ClientSlug: clientSlug, ProjectSlug: projectSlug,
	})
	if err != nil {
		t.Fatalf("project not created: %v", err)
	}
	if project.Name != "Notes on the Analytical Engine" {
		t.Errorf("project name = %q, want the manuscript title", project.Name)
	}
	pass, err := q.GetPassByProject(t.Context(), project.ID)
	if err != nil {
		t.Fatalf("pass not created: %v", err)
	}
	if pass.Source != "coupon" || pass.CouponID.Int64 != couponID {
		t.Errorf("pass source/coupon = %q/%d, want coupon/%d", pass.Source, pass.CouponID.Int64, couponID)
	}
	if pass.BuildsIncluded != passBuildsIncluded || passCreditsRemaining(pass) != passBuildsIncluded {
		t.Errorf("pass credits = %d of %d, want %d included", passCreditsRemaining(pass), pass.BuildsIncluded, passBuildsIncluded)
	}
	if !passLive(pass) {
		t.Errorf("a freshly fulfilled pass must be live (status %q, expires %v)", pass.Status, pass.ExpiresAt)
	}
	if want := pass.FulfilledAt.AddDate(0, passStorageMonths, 0); !pass.ExpiresAt.Equal(want) {
		t.Errorf("expires_at = %v, want fulfilled_at + %d months (%v)", pass.ExpiresAt, passStorageMonths, want)
	}

	// Coupon burned.
	coupon, err := q.GetCouponByCode(t.Context(), "PYB-AAAA-BBBB")
	if err != nil {
		t.Fatalf("reload coupon: %v", err)
	}
	if coupon.RedeemedCount != 1 {
		t.Errorf("coupon redeemed_count = %d, want 1", coupon.RedeemedCount)
	}
}

// TestRedeemLinksMatchingRegistration: redeeming with an email that matches a
// cohort registration attributes the code to that registration, so the tracker
// can show issued/redeemed per attendee.
func TestRedeemLinksMatchingRegistration(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	resp := apiRequest(t, ts, "POST", "/api/public/register", map[string]any{
		"name": "Grace Hopper", "email": "grace@example.com", "material": "a memoir",
	})
	resp.Body.Close()
	var regID int64
	if err := s.DB.QueryRow(`SELECT id FROM event_registrations WHERE email = ?`, "grace@example.com").Scan(&regID); err != nil {
		t.Fatalf("registration not seeded: %v", err)
	}
	seedCoupon(t, s, "PYB-CCCC-DDDD", time.Now().Add(24*time.Hour))

	resp = redeem(t, ts, map[string]string{
		"code": "PYB-CCCC-DDDD", "name": "Grace Hopper",
		"email": "grace@example.com", "title": "Compiling", "author": "Grace Hopper",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("redeem: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	q := dbgen.New(s.DB)
	coupon, err := q.GetCouponByCode(t.Context(), "PYB-CCCC-DDDD")
	if err != nil {
		t.Fatalf("reload coupon: %v", err)
	}
	if coupon.RegistrationID.Int64 != regID {
		t.Errorf("coupon registration_id = %d, want %d", coupon.RegistrationID.Int64, regID)
	}

	// And the admin tracker surfaces it.
	resp = apiRequestAdmin(t, ts, "GET", "/api/admin/registrations", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("admin registrations: expected 200, got %d", resp.StatusCode)
	}
	var listing map[string]any
	decodeJSON(t, resp, &listing)
	rows := listing["registrations"].([]any)
	if len(rows) != 1 {
		t.Fatalf("expected 1 registration row, got %d", len(rows))
	}
	row := rows[0].(map[string]any)
	if row["coupon_code"] != "PYB-CCCC-DDDD" {
		t.Errorf("tracker coupon_code = %v, want PYB-CCCC-DDDD", row["coupon_code"])
	}
	if row["coupon_redeemed_at"] == "" {
		t.Errorf("tracker coupon_redeemed_at should be set once redeemed, got %v", row["coupon_redeemed_at"])
	}
}

// TestRedeemSameCodeTwiceRejected: a used code is refused with 400 and never
// silently re-fulfilled (a second pass would double the storage bill and
// confuse the customer about which portal is theirs).
func TestRedeemSameCodeTwiceRejected(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	seedCoupon(t, s, "PYB-EEEE-FFFF", time.Now().Add(24*time.Hour))
	first := redeem(t, ts, map[string]string{
		"code": "PYB-EEEE-FFFF", "name": "First Redeemer",
		"email": "first@example.com", "title": "Book One",
	})
	if first.StatusCode != 200 {
		t.Fatalf("first redeem: expected 200, got %d", first.StatusCode)
	}
	first.Body.Close()

	passesBefore := countRows(t, s, `SELECT COUNT(*) FROM passes`)
	clientsBefore := countRows(t, s, `SELECT COUNT(*) FROM clients`)

	second := redeem(t, ts, map[string]string{
		"code": "PYB-EEEE-FFFF", "name": "Second Redeemer",
		"email": "second@example.com", "title": "Book Two",
	})
	if second.StatusCode != http.StatusBadRequest {
		t.Fatalf("second redeem of the same code: expected 400, got %d", second.StatusCode)
	}
	var errBody map[string]string
	decodeJSON(t, second, &errBody)
	if !strings.Contains(strings.ToLower(errBody["error"]), "already been redeemed") {
		t.Errorf("unexpected error message: %q", errBody["error"])
	}

	if got := countRows(t, s, `SELECT COUNT(*) FROM passes`); got != passesBefore {
		t.Errorf("rejected redemption created a pass: %d -> %d", passesBefore, got)
	}
	if got := countRows(t, s, `SELECT COUNT(*) FROM clients`); got != clientsBefore {
		t.Errorf("rejected redemption created a client: %d -> %d", clientsBefore, got)
	}
}

// TestRedeemExpiredCodeRejected: an expired code is a 400 with nothing created.
func TestRedeemExpiredCodeRejected(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	seedCoupon(t, s, "PYB-GGGG-HHHH", time.Now().Add(-1*time.Hour))
	resp := redeem(t, ts, map[string]string{
		"code": "PYB-GGGG-HHHH", "name": "Late Redeemer",
		"email": "late@example.com", "title": "Too Late",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expired code: expected 400, got %d", resp.StatusCode)
	}
	var errBody map[string]string
	decodeJSON(t, resp, &errBody)
	if !strings.Contains(strings.ToLower(errBody["error"]), "expired") {
		t.Errorf("unexpected error message: %q", errBody["error"])
	}
	if n := countRows(t, s, `SELECT COUNT(*) FROM passes`); n != 0 {
		t.Errorf("expired code must not fulfil anything, found %d passes", n)
	}
	if n := countRows(t, s, `SELECT COUNT(*) FROM clients`); n != 0 {
		t.Errorf("expired code must not create a client, found %d", n)
	}
}

// TestRedeemUnknownCodeRejected: a typo'd code is a 400, not a 500.
func TestRedeemUnknownCodeRejected(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	resp := redeem(t, ts, map[string]string{
		"code": "PYB-ZZZZ-ZZZZ", "name": "Nobody",
		"email": "nobody@example.com", "title": "No Book",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("unknown code: expected 400, got %d", resp.StatusCode)
	}
	resp.Body.Close()
	if n := countRows(t, s, `SELECT COUNT(*) FROM passes`); n != 0 {
		t.Errorf("unknown code must not fulfil anything, found %d passes", n)
	}
}

// TestRedeemHoneypotLooksSuccessfulButDoesNothing: the hidden "company" field
// is bot bait — answer 200 so the bot doesn't learn it was caught, while
// leaving the coupon unredeemed and creating nothing.
func TestRedeemHoneypotLooksSuccessfulButDoesNothing(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	seedCoupon(t, s, "PYB-JJJJ-KKKK", time.Now().Add(24*time.Hour))
	resp := redeem(t, ts, map[string]string{
		"code": "PYB-JJJJ-KKKK", "name": "Bot", "email": "bot@example.com",
		"title": "Spam", "company": "Acme Spam Co",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("honeypot: expected 200, got %d", resp.StatusCode)
	}
	var body map[string]any
	decodeJSON(t, resp, &body)
	if body["ok"] != true {
		t.Errorf("honeypot response should look successful, got %+v", body)
	}
	if _, ok := body["portal_url"]; ok {
		t.Errorf("honeypot response must not hand out a portal URL: %+v", body)
	}
	if n := countRows(t, s, `SELECT COUNT(*) FROM passes`); n != 0 {
		t.Errorf("honeypot created %d passes, want 0", n)
	}
	if n := countRows(t, s, `SELECT COUNT(*) FROM clients`); n != 0 {
		t.Errorf("honeypot created %d clients, want 0", n)
	}
	if n := countRows(t, s, `SELECT COUNT(*) FROM coupons WHERE redeemed_count > 0`); n != 0 {
		t.Errorf("honeypot burned a coupon")
	}
}

// TestRedeemSlugCollisionsGetSuffixed: two customers with the same name and
// title still get distinct portals.
func TestRedeemSlugCollisionsGetSuffixed(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	seedCoupon(t, s, "PYB-1111-2222", time.Time{})
	seedCoupon(t, s, "PYB-3333-4444", time.Time{})

	slugs := make([]string, 0, 2)
	for _, code := range []string{"PYB-1111-2222", "PYB-3333-4444"} {
		resp := redeem(t, ts, map[string]string{
			"code": code, "name": "Jane Doe",
			"email": "jane+" + code + "@example.com", "title": "Same Title",
		})
		if resp.StatusCode != 200 {
			t.Fatalf("redeem %s: expected 200, got %d", code, resp.StatusCode)
		}
		var body map[string]any
		decodeJSON(t, resp, &body)
		slugs = append(slugs, body["client_slug"].(string)+"/"+body["project_slug"].(string))
	}
	if slugs[0] == slugs[1] {
		t.Fatalf("both redemptions landed on %q; slugs must be unique-ified", slugs[0])
	}
	if slugs[0] != "jane-doe/same-title" || slugs[1] != "jane-doe-2/same-title" {
		t.Errorf("unexpected slugs %v, want [jane-doe/same-title jane-doe-2/same-title]", slugs)
	}
}

// ─── gating ───

// TestFactoryEndpointsRejectAnonymousCallers: a passed project is the
// customer's, not the public's. Without the client cookie (or admin header),
// upload and convert are refused and nothing is created.
func TestFactoryEndpointsRejectAnonymousCallers(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	pass, password, clientSlug, _ := grantedPass(t, s, ts, "Ida Tester", "Gated Book")

	// Upload with the right project_id but no credentials.
	resp := uploadBook(t, ts, itoa(pass.ProjectID), nil, false)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("anonymous upload: expected 401, got %d", resp.StatusCode)
	}
	resp.Body.Close()
	if n := countRows(t, s, `SELECT COUNT(*) FROM books`); n != 0 {
		t.Fatalf("anonymous upload created %d books, want 0", n)
	}

	// Upload as the customer, so there is a book to try to convert.
	cookie := clientCookie(t, ts, clientSlug, password)
	resp = uploadBook(t, ts, itoa(pass.ProjectID), cookie, false)
	if resp.StatusCode != 201 {
		t.Fatalf("customer upload: expected 201, got %d", resp.StatusCode)
	}
	var created map[string]any
	decodeJSON(t, resp, &created)
	bookID := itoa(int64(created["id"].(float64)))

	// Convert without credentials.
	resp = apiRequest(t, ts, "POST", "/api/books/"+bookID+"/convert", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("anonymous convert: expected 401, got %d", resp.StatusCode)
	}
	resp.Body.Close()
	if n := countLedger(t, s, pass.ID, "build"); n != 0 {
		t.Fatalf("refused convert wrote %d debit rows, want 0", n)
	}

	// Preflight without credentials.
	resp = apiRequest(t, ts, "POST", "/api/projects/"+itoa(pass.ProjectID)+"/preflight",
		map[string]any{"book_id": 1})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("anonymous preflight: expected 401, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

// TestUploadRequiresProjectIDForNonAdmin: without a project_id there is no
// pass to authorize against, so a non-admin upload is refused.
func TestUploadRequiresProjectIDForNonAdmin(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	resp := uploadBook(t, ts, "", nil, false)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("upload with no project_id and no admin header: expected 401, got %d", resp.StatusCode)
	}
	resp.Body.Close()
	if n := countRows(t, s, `SELECT COUNT(*) FROM books`); n != 0 {
		t.Fatalf("refused upload created %d books, want 0", n)
	}

	// The admin path is unchanged: unlinked uploads still work.
	resp = uploadBook(t, ts, "", nil, true)
	if resp.StatusCode != 201 {
		t.Fatalf("admin upload with no project_id: expected 201, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

// TestExpiredPassBlocksUploadAndConvert: past expires_at the project goes
// read-only for the customer — 403, no debit.
func TestExpiredPassBlocksUploadAndConvert(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	pass, password, clientSlug, _ := grantedPass(t, s, ts, "Rip Winkle", "Slept Too Long")
	cookie := clientCookie(t, ts, clientSlug, password)

	// A book uploaded while the pass was live.
	resp := uploadBook(t, ts, itoa(pass.ProjectID), cookie, false)
	if resp.StatusCode != 201 {
		t.Fatalf("upload while live: expected 201, got %d", resp.StatusCode)
	}
	var created map[string]any
	decodeJSON(t, resp, &created)
	bookID := itoa(int64(created["id"].(float64)))

	if _, err := s.DB.Exec(`UPDATE passes SET expires_at = datetime(CURRENT_TIMESTAMP, '-1 day') WHERE id = ?`, pass.ID); err != nil {
		t.Fatalf("expire pass: %v", err)
	}

	resp = uploadBook(t, ts, itoa(pass.ProjectID), cookie, false)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("upload on expired pass: expected 403, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	req, err := http.NewRequest("POST", ts.URL+"/api/books/"+bookID+"/convert", nil)
	if err != nil {
		t.Fatalf("create convert request: %v", err)
	}
	req.AddCookie(cookie)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do convert: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("convert on expired pass: expected 403, got %d", resp.StatusCode)
	}
	resp.Body.Close()
	if n := countLedger(t, s, pass.ID, "build"); n != 0 {
		t.Fatalf("expired-pass convert debited %d credits, want 0", n)
	}
}

// TestConvertWithoutCreditsReturns402: credits are the scope fence. With none
// left the customer gets 402 with credits_remaining, and no debit is written.
func TestConvertWithoutCreditsReturns402(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	pass, password, clientSlug, _ := grantedPass(t, s, ts, "Spent Budget", "All Used Up")
	cookie := clientCookie(t, ts, clientSlug, password)

	resp := uploadBook(t, ts, itoa(pass.ProjectID), cookie, false)
	if resp.StatusCode != 201 {
		t.Fatalf("upload: expected 201, got %d", resp.StatusCode)
	}
	var created map[string]any
	decodeJSON(t, resp, &created)
	bookID := itoa(int64(created["id"].(float64)))

	// Burn every included build.
	if _, err := s.DB.Exec(`UPDATE passes SET builds_used = builds_included WHERE id = ?`, pass.ID); err != nil {
		t.Fatalf("exhaust credits: %v", err)
	}

	req, err := http.NewRequest("POST", ts.URL+"/api/books/"+bookID+"/convert", nil)
	if err != nil {
		t.Fatalf("create convert request: %v", err)
	}
	req.AddCookie(cookie)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do convert: %v", err)
	}
	if resp.StatusCode != http.StatusPaymentRequired {
		t.Fatalf("convert with no credits: expected 402, got %d", resp.StatusCode)
	}
	var body map[string]any
	decodeJSON(t, resp, &body)
	if body["error"] != "no builds remaining" {
		t.Errorf("402 error = %v, want \"no builds remaining\"", body["error"])
	}
	if body["credits_remaining"] != float64(0) {
		t.Errorf("402 credits_remaining = %v, want 0", body["credits_remaining"])
	}
	if n := countLedger(t, s, pass.ID, "build"); n != 0 {
		t.Fatalf("402 wrote %d debit rows, want 0", n)
	}
	var status string
	if err := s.DB.QueryRow(`SELECT status FROM books WHERE id = ?`, bookID).Scan(&status); err != nil {
		t.Fatalf("read book status: %v", err)
	}
	if status == "converting" {
		t.Errorf("402 must not start a build (book status %q)", status)
	}
}

// TestConvertOneInFlightPerProject: builds are minutes of CPU and the status
// field is per-book, so a second concurrent build for the same project is a
// 409 and costs no credit.
func TestConvertOneInFlightPerProject(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	pass, password, clientSlug, _ := grantedPass(t, s, ts, "Busy Author", "Queued Book")
	cookie := clientCookie(t, ts, clientSlug, password)

	resp := uploadBook(t, ts, itoa(pass.ProjectID), cookie, false)
	if resp.StatusCode != 201 {
		t.Fatalf("upload: expected 201, got %d", resp.StatusCode)
	}
	var created map[string]any
	decodeJSON(t, resp, &created)
	bookID := itoa(int64(created["id"].(float64)))

	// Pretend a sibling build is already running for this project.
	q := dbgen.New(s.DB)
	sibling, err := q.CreateBook(t.Context(), dbgen.CreateBookParams{
		Title: "Sibling", Author: "Busy Author", SourceFilename: "s.docx",
		SourceData: []byte("x"),
		ProjectID:  sql.NullInt64{Int64: pass.ProjectID, Valid: true},
	})
	if err != nil {
		t.Fatalf("create sibling book: %v", err)
	}
	if err := q.UpdateBookStatus(t.Context(), dbgen.UpdateBookStatusParams{
		Status: "converting", ID: sibling.ID,
	}); err != nil {
		t.Fatalf("mark sibling converting: %v", err)
	}

	req, err := http.NewRequest("POST", ts.URL+"/api/books/"+bookID+"/convert", nil)
	if err != nil {
		t.Fatalf("create convert request: %v", err)
	}
	req.AddCookie(cookie)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do convert: %v", err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("second in-flight build: expected 409, got %d", resp.StatusCode)
	}
	resp.Body.Close()
	if n := countLedger(t, s, pass.ID, "build"); n != 0 {
		t.Fatalf("409 wrote %d debit rows, want 0", n)
	}
}

// TestConvertDebitsThenRefundsOnFailure: the credit is spent synchronously,
// before the pipeline runs (so a crashed build can't be free), and handed back
// when the build fails. The fake .docx guarantees pandoc fails here — which is
// exactly the "failed builds don't count" promise.
func TestConvertDebitsThenRefundsOnFailure(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	pass, password, clientSlug, _ := grantedPass(t, s, ts, "Unlucky Author", "Broken Docx")
	cookie := clientCookie(t, ts, clientSlug, password)

	resp := uploadBook(t, ts, itoa(pass.ProjectID), cookie, false)
	if resp.StatusCode != 201 {
		t.Fatalf("upload: expected 201, got %d", resp.StatusCode)
	}
	var created map[string]any
	decodeJSON(t, resp, &created)
	bookID := int64(created["id"].(float64))

	req, err := http.NewRequest("POST", ts.URL+"/api/books/"+itoa(bookID)+"/convert", nil)
	if err != nil {
		t.Fatalf("create convert request: %v", err)
	}
	req.AddCookie(cookie)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do convert: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("convert: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// The debit is written by the handler, so it is already visible.
	if n := countLedger(t, s, pass.ID, "build"); n != 1 {
		t.Fatalf("debit rows after convert = %d, want 1 (debit must happen before the pipeline starts)", n)
	}

	// The refund lands when the background build fails.
	waitForLedger(t, s, pass.ID, "build_failed_refund")

	q := dbgen.New(s.DB)
	reloaded, err := q.GetPassByProject(t.Context(), pass.ProjectID)
	if err != nil {
		t.Fatalf("reload pass: %v", err)
	}
	if reloaded.BuildsUsed != 0 {
		t.Errorf("builds_used = %d after a failed build, want 0 (refunded)", reloaded.BuildsUsed)
	}
	if got := passCreditsRemaining(reloaded); got != passBuildsIncluded {
		t.Errorf("credits_remaining = %d after a failed build, want %d", got, passBuildsIncluded)
	}

	// The ledger keeps both movements — the refund is a new row, not an
	// erasure, so the dogfood data shows how many builds were attempted.
	if n := countLedger(t, s, pass.ID, "build"); n != 1 {
		t.Errorf("debit row disappeared after refund (%d rows)", n)
	}
}

// TestAdminConvertOnPassedProjectStillDebits: the workshop instructor runs a
// build from the admin UI. Gating is skipped, but the ledger still records it,
// so credits mean the same thing regardless of who pressed the button.
func TestAdminConvertOnPassedProjectStillDebits(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	pass, _, _, _ := grantedPass(t, s, ts, "Workshop Attendee", "Instructor Built")

	resp := uploadBook(t, ts, itoa(pass.ProjectID), nil, true)
	if resp.StatusCode != 201 {
		t.Fatalf("admin upload: expected 201, got %d", resp.StatusCode)
	}
	var created map[string]any
	decodeJSON(t, resp, &created)
	bookID := itoa(int64(created["id"].(float64)))

	// Exhaust the credits: an admin build is still allowed (no 402)…
	if _, err := s.DB.Exec(`UPDATE passes SET builds_used = builds_included WHERE id = ?`, pass.ID); err != nil {
		t.Fatalf("exhaust credits: %v", err)
	}
	resp = apiRequestAdmin(t, ts, "POST", "/api/books/"+bookID+"/convert", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("admin convert with no credits left: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// …and is still debited.
	if n := countLedger(t, s, pass.ID, "build"); n != 1 {
		t.Fatalf("admin convert wrote %d debit rows, want 1", n)
	}
	waitForLedger(t, s, pass.ID, "build_failed_refund")
}

// TestDebitAndRefundBuildCredit unit-tests the ledger pair directly: a debit
// spends exactly one credit and records it; the refund puts it back.
func TestDebitAndRefundBuildCredit(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	pass, _, _, _ := grantedPass(t, s, ts, "Ledger Test", "Ledger Book")

	if err := s.debitBuildCredit(t.Context(), pass.ID, 0); err != nil {
		t.Fatalf("debit: %v", err)
	}
	q := dbgen.New(s.DB)
	after, err := q.GetPass(t.Context(), pass.ID)
	if err != nil {
		t.Fatalf("reload pass: %v", err)
	}
	if after.BuildsUsed != 1 || passCreditsRemaining(after) != passBuildsIncluded-1 {
		t.Fatalf("after debit: builds_used=%d credits=%d, want 1 and %d",
			after.BuildsUsed, passCreditsRemaining(after), passBuildsIncluded-1)
	}
	if n := countLedger(t, s, pass.ID, "build"); n != 1 {
		t.Fatalf("debit ledger rows = %d, want 1", n)
	}

	if err := s.refundBuildCredit(t.Context(), pass.ID, 0); err != nil {
		t.Fatalf("refund: %v", err)
	}
	after, err = q.GetPass(t.Context(), pass.ID)
	if err != nil {
		t.Fatalf("reload pass: %v", err)
	}
	if after.BuildsUsed != 0 || passCreditsRemaining(after) != passBuildsIncluded {
		t.Fatalf("after refund: builds_used=%d credits=%d, want 0 and %d",
			after.BuildsUsed, passCreditsRemaining(after), passBuildsIncluded)
	}
	if n := countLedger(t, s, pass.ID, "build_failed_refund"); n != 1 {
		t.Fatalf("refund ledger rows = %d, want 1", n)
	}
}

// ─── customer status endpoints ───

// TestGetProjectPassReportsCredits: the badge the customer page renders.
func TestGetProjectPassReportsCredits(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	pass, password, clientSlug, _ := grantedPass(t, s, ts, "Status Reader", "Status Book")
	cookie := clientCookie(t, ts, clientSlug, password)

	if err := s.debitBuildCredit(t.Context(), pass.ID, 0); err != nil {
		t.Fatalf("debit: %v", err)
	}

	req, err := http.NewRequest("GET", ts.URL+"/api/projects/"+itoa(pass.ProjectID)+"/pass", nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.AddCookie(cookie)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("get pass: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("get pass: expected 200, got %d", resp.StatusCode)
	}
	var body map[string]any
	decodeJSON(t, resp, &body)
	if body["exists"] != true || body["live"] != true {
		t.Fatalf("expected an existing live pass, got %+v", body)
	}
	if body["builds_used"] != float64(1) {
		t.Errorf("builds_used = %v, want 1", body["builds_used"])
	}
	if body["credits_remaining"] != float64(passBuildsIncluded-1) {
		t.Errorf("credits_remaining = %v, want %d", body["credits_remaining"], passBuildsIncluded-1)
	}
	if body["expires_at"] == "" {
		t.Errorf("expires_at should be set, got %v", body["expires_at"])
	}
}

// TestGetProjectPassExistsFalseForPasslessProject: admin-only projects behave
// as before — the endpoint answers, it just says there is no pass.
func TestGetProjectPassExistsFalseForPasslessProject(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	pid := createProjectT(t, ts, "No Pass Here", "npc", "one")
	resp := apiRequestAdmin(t, ts, "GET", "/api/projects/"+itoa(pid)+"/pass", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("get pass: expected 200, got %d", resp.StatusCode)
	}
	var body map[string]any
	decodeJSON(t, resp, &body)
	if body["exists"] != false {
		t.Fatalf("expected exists=false for a project with no pass, got %+v", body)
	}
}

// TestListProjectBooksScopedToProject: the customer's book list shows only
// their project's books, never another project's.
func TestListProjectBooksScopedToProject(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	pass, password, clientSlug, _ := grantedPass(t, s, ts, "Mine Only", "My Book")
	cookie := clientCookie(t, ts, clientSlug, password)

	resp := uploadBook(t, ts, itoa(pass.ProjectID), cookie, false)
	if resp.StatusCode != 201 {
		t.Fatalf("upload: expected 201, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Someone else's book, in someone else's project.
	otherPID := createProjectT(t, ts, "Other Project", "other", "book")
	q := dbgen.New(s.DB)
	if _, err := q.CreateBook(t.Context(), dbgen.CreateBookParams{
		Title: "Not Yours", Author: "Someone", SourceFilename: "x.docx",
		SourceData: []byte("x"), ProjectID: sql.NullInt64{Int64: otherPID, Valid: true},
	}); err != nil {
		t.Fatalf("create other book: %v", err)
	}

	req, err := http.NewRequest("GET", ts.URL+"/api/projects/"+itoa(pass.ProjectID)+"/books", nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.AddCookie(cookie)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("list books: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("list books: expected 200, got %d", resp.StatusCode)
	}
	var books []map[string]any
	decodeJSON(t, resp, &books)
	if len(books) != 1 {
		t.Fatalf("expected exactly the project's 1 book, got %d: %+v", len(books), books)
	}
	if books[0]["Title"] != "Test Manuscript" {
		t.Errorf("unexpected book in the customer's list: %+v", books[0])
	}
}

// ─── admin surfaces ───

// TestAdminCouponIssueAndList: issuing a code for a registration binds it to
// that registrant and shows up in the coupon list as unredeemed.
func TestAdminCouponIssueAndList(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	resp := apiRequest(t, ts, "POST", "/api/public/register", map[string]any{
		"name": "Cohort Member", "email": "member@example.com", "material": "a zine",
	})
	resp.Body.Close()
	var regID int64
	if err := s.DB.QueryRow(`SELECT id FROM event_registrations WHERE email = ?`, "member@example.com").Scan(&regID); err != nil {
		t.Fatalf("registration not seeded: %v", err)
	}

	resp = apiRequestAdmin(t, ts, "POST", "/api/admin/coupons", map[string]any{
		"registration_id": regID, "note": "cohort code",
	})
	if resp.StatusCode != 201 {
		t.Fatalf("issue coupon: expected 201, got %d", resp.StatusCode)
	}
	var issued map[string]any
	decodeJSON(t, resp, &issued)
	code, _ := issued["code"].(string)
	if !strings.HasPrefix(code, couponPrefix+"-") || len(code) != 13 {
		t.Fatalf("code %q does not match PYB-XXXX-XXXX", code)
	}
	if strings.ContainsAny(code[4:], "01OI") {
		t.Errorf("code %q must avoid the ambiguous characters 0/O/1/I", code)
	}
	if issued["issued_to_email"] != "member@example.com" {
		t.Errorf("issued_to_email = %v, want the registrant's email", issued["issued_to_email"])
	}

	resp = apiRequestAdmin(t, ts, "GET", "/api/admin/coupons", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("list coupons: expected 200, got %d", resp.StatusCode)
	}
	var coupons []map[string]any
	decodeJSON(t, resp, &coupons)
	if len(coupons) != 1 {
		t.Fatalf("expected 1 coupon, got %d", len(coupons))
	}
	if coupons[0]["redeemed"] != false || coupons[0]["redeemed_at"] != "" {
		t.Errorf("a fresh coupon should read as unredeemed, got %+v", coupons[0])
	}

	// After redemption the list links the project it created.
	resp = redeem(t, ts, map[string]string{
		"code": code, "name": "Cohort Member",
		"email": "member@example.com", "title": "The Zine",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("redeem issued code: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	resp = apiRequestAdmin(t, ts, "GET", "/api/admin/coupons", nil)
	decodeJSON(t, resp, &coupons)
	if coupons[0]["redeemed"] != true {
		t.Errorf("coupon should read as redeemed, got %+v", coupons[0])
	}
	if path, _ := coupons[0]["project_path"].(string); !strings.HasSuffix(path, "/factory/") {
		t.Errorf("project_path = %q, want the customer's factory page", path)
	}
}

// TestAdminCouponRoutesRequireAdmin: the coupon and pass surfaces are
// admin-only — an anonymous caller can neither mint codes nor read them.
func TestAdminPassRoutesRequireAdmin(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	for _, tc := range []struct {
		method, path string
		body         any
	}{
		{"POST", "/api/admin/coupons", map[string]any{"note": "sneaky"}},
		{"GET", "/api/admin/coupons", nil},
		{"GET", "/api/admin/passes", nil},
		{"POST", "/api/admin/passes", map[string]any{"name": "X", "email": "x@example.com", "title": "Y"}},
		{"POST", "/api/admin/passes/1/grant", map[string]any{"builds": 3}},
	} {
		resp := apiRequest(t, ts, tc.method, tc.path, tc.body)
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("%s %s without admin header: expected 401, got %d", tc.method, tc.path, resp.StatusCode)
		}
		resp.Body.Close()
	}
}

// TestAdminGrantBuildsAddsCreditsAndLedgerRow: a bought pack raises
// builds_extra and is recorded, so credits_remaining goes up by the grant.
func TestAdminGrantBuildsAddsCreditsAndLedgerRow(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	pass, _, _, _ := grantedPass(t, s, ts, "Pack Buyer", "Needs More Builds")

	resp := apiRequestAdmin(t, ts, "POST", "/api/admin/passes/"+itoa(pass.ID)+"/grant",
		map[string]any{"builds": 3, "reason": "pack"})
	if resp.StatusCode != 200 {
		t.Fatalf("grant builds: expected 200, got %d", resp.StatusCode)
	}
	var body map[string]any
	decodeJSON(t, resp, &body)
	if body["credits_remaining"] != float64(passBuildsIncluded+3) {
		t.Errorf("credits_remaining = %v, want %d", body["credits_remaining"], passBuildsIncluded+3)
	}
	if n := countLedger(t, s, pass.ID, "pack"); n != 1 {
		t.Errorf("pack ledger rows = %d, want 1", n)
	}

	// Bad grants are rejected.
	resp = apiRequestAdmin(t, ts, "POST", "/api/admin/passes/"+itoa(pass.ID)+"/grant",
		map[string]any{"builds": 0})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("grant of 0 builds: expected 400, got %d", resp.StatusCode)
	}
	resp.Body.Close()
	resp = apiRequestAdmin(t, ts, "POST", "/api/admin/passes/9999/grant",
		map[string]any{"builds": 1})
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("grant on missing pass: expected 404, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

// TestAdminListPasses: the operator view — project, client, credits, expiry.
func TestAdminListPasses(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	pass, _, clientSlug, projectSlug := grantedPass(t, s, ts, "Listed Author", "Listed Book")

	resp := apiRequestAdmin(t, ts, "GET", "/api/admin/passes", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("list passes: expected 200, got %d", resp.StatusCode)
	}
	var rows []map[string]any
	decodeJSON(t, resp, &rows)
	if len(rows) != 1 {
		t.Fatalf("expected 1 pass, got %d", len(rows))
	}
	row := rows[0]
	if row["id"] != float64(pass.ID) || row["source"] != "admin" {
		t.Errorf("unexpected pass row: %+v", row)
	}
	if row["project_path"] != "/"+clientSlug+"/"+projectSlug+"/factory/" {
		t.Errorf("project_path = %v, want the factory page path", row["project_path"])
	}
	if row["credits_remaining"] != float64(passBuildsIncluded) || row["live"] != true {
		t.Errorf("expected %d credits on a live pass, got %+v", passBuildsIncluded, row)
	}
}

// ─── page route ───

// TestFactoryPageRoute: /{client}/{project}/factory/ serves the customer page,
// the un-slashed form redirects, and deeper paths fall through to static
// assets (same shape as /transmittal/).
func TestFactoryPageRoute(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	createProjectT(t, ts, "Factory Page", "fpc", "book")

	resp, err := http.Get(ts.URL + "/fpc/book/factory/")
	if err != nil {
		t.Fatalf("get factory page: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("factory page: expected 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("factory page content type = %q, want text/html", ct)
	}

	noRedirect := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	resp, err = noRedirect.Get(ts.URL + "/fpc/book/factory")
	if err != nil {
		t.Fatalf("get factory page without trailing slash: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMovedPermanently {
		t.Errorf("factory page without trailing slash: expected 301, got %d", resp.StatusCode)
	}

	resp, err = http.Get(ts.URL + "/fpc/book/factory/theme.css")
	if err != nil {
		t.Fatalf("get factory asset: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("asset under the factory page: expected 200, got %d", resp.StatusCode)
	}
}

// ─── code + password generation ───

func TestGenerateCouponCodeShape(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		code, err := generateCouponCode()
		if err != nil {
			t.Fatalf("generate code: %v", err)
		}
		if len(code) != 13 || !strings.HasPrefix(code, couponPrefix+"-") || code[8] != '-' {
			t.Fatalf("code %q is not PYB-XXXX-XXXX", code)
		}
		if code != strings.ToUpper(code) {
			t.Fatalf("code %q must be uppercase", code)
		}
		for _, r := range code[4:] {
			if r == '-' {
				continue
			}
			if !strings.ContainsRune(codeAlphabet, r) {
				t.Fatalf("code %q contains %q, which is outside the unambiguous alphabet", code, r)
			}
		}
		seen[code] = true
	}
	if len(seen) < 45 {
		t.Errorf("only %d distinct codes out of 50 — generation looks weak", len(seen))
	}
}

func TestGenerateClientPasswordShape(t *testing.T) {
	pw, err := generateClientPassword()
	if err != nil {
		t.Fatalf("generate password: %v", err)
	}
	if len(pw) != 12 {
		t.Fatalf("password %q has length %d, want 12", pw, len(pw))
	}
	for _, r := range pw {
		if !strings.ContainsRune(passwordAlphabet, r) {
			t.Fatalf("password %q contains %q, which is outside the unambiguous alphabet", pw, r)
		}
	}
}

// TestPassEmailBodiesCarryTheEssentials: the fulfillment mail must contain the
// three things the customer cannot get anywhere else (portal, sign-in name,
// password) plus the support edges; the build mail must carry the links.
func TestPassEmailBodiesCarryTheEssentials(t *testing.T) {
	res := fulfillPassResult{
		Pass: dbgen.Pass{
			CustomerName: "Ada Lovelace", CustomerEmail: "ada@example.com",
			BuildsIncluded: passBuildsIncluded, ProjectID: 7,
			ExpiresAt: time.Date(2027, 3, 3, 0, 0, 0, 0, time.UTC),
		},
		Title:       "Notes",
		ClientSlug:  "ada-lovelace",
		ProjectSlug: "notes",
		Password:    "Xy7mQ2rKp4Ln",
		PortalURL:   "https://example.test/ada-lovelace/notes/factory/",
	}
	for _, body := range []string{passFulfillmentText(res), passFulfillmentHTML(res)} {
		for _, want := range []string{res.PortalURL, res.ClientSlug, res.Password, "3 March 2027", "100/hr"} {
			if !strings.Contains(body, want) {
				t.Errorf("fulfillment email is missing %q:\n%s", want, body)
			}
		}
	}

	s := &Server{BaseURL: "https://example.test"}
	// Email is unconfigured, so this must log-and-skip rather than panic.
	s.sendBuildDeliveredEmail(res.Pass, dbgen.Book{ID: 11, Title: "Notes"})
	s.sendPassFulfillmentEmail(res)
}

// ─── build = PDF + EPUB ───

// TestFinalizeBuildRunsEPUBBeforeReady: a build is both deliverables, so the
// EPUB stage runs while the book is still "converting" and the status only
// flips to ready afterwards. Asserted by having the fake EPUB runner observe
// the status it was called under.
func TestFinalizeBuildRunsEPUBBeforeReady(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	pass, _, _, _ := grantedPass(t, s, ts, "Both Formats", "Two Deliverables")
	q := dbgen.New(s.DB)
	book, err := q.CreateBook(t.Context(), dbgen.CreateBookParams{
		Title: "Two Deliverables", Author: "Both Formats", SourceFilename: "b.docx",
		SourceData: []byte("x"), ProjectID: sql.NullInt64{Int64: pass.ProjectID, Valid: true},
	})
	if err != nil {
		t.Fatalf("create book: %v", err)
	}
	if err := q.UpdateBookStatus(t.Context(), dbgen.UpdateBookStatusParams{
		Status: "converting", ErrorMsg: "stale error from a previous attempt", ID: book.ID,
	}); err != nil {
		t.Fatalf("mark converting: %v", err)
	}

	called := 0
	statusDuringEPUB := ""
	s.epubRunner = func(bid int64, b dbgen.Book) error {
		called++
		if bid != book.ID {
			t.Errorf("epub runner got book %d, want %d", bid, book.ID)
		}
		_ = s.DB.QueryRow(`SELECT status FROM books WHERE id = ?`, bid).Scan(&statusDuringEPUB)
		return nil
	}

	if err := s.finalizeBuild(t.Context(), book.ID, book); err != nil {
		t.Fatalf("finalizeBuild: %v", err)
	}
	if called != 1 {
		t.Fatalf("epub stage ran %d times, want exactly 1 (a build is PDF + EPUB)", called)
	}
	if statusDuringEPUB != "converting" {
		t.Errorf("status during the EPUB stage = %q, want \"converting\" (the UI polls on this)", statusDuringEPUB)
	}

	var status, errMsg string
	if err := s.DB.QueryRow(`SELECT status, error_msg FROM books WHERE id = ?`, book.ID).Scan(&status, &errMsg); err != nil {
		t.Fatalf("read book: %v", err)
	}
	if status != "ready" {
		t.Errorf("status after a complete build = %q, want \"ready\"", status)
	}
	if errMsg != "" {
		t.Errorf("a clean build must clear error_msg, got %q", errMsg)
	}
}

// TestFinalizeBuildKeepsPDFOnlyBuildAndDoesNotRefund: the customer got a
// correctly typeset PDF, so a failed EPUB is a note, not a failed build —
// status ready, reason in error_msg, credit stays spent.
func TestFinalizeBuildKeepsPDFOnlyBuildAndDoesNotRefund(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	pass, _, _, _ := grantedPass(t, s, ts, "Pdf Only", "Epub Broke")
	q := dbgen.New(s.DB)
	book, err := q.CreateBook(t.Context(), dbgen.CreateBookParams{
		Title: "Epub Broke", Author: "Pdf Only", SourceFilename: "b.docx",
		SourceData: []byte("x"), ProjectID: sql.NullInt64{Int64: pass.ProjectID, Valid: true},
	})
	if err != nil {
		t.Fatalf("create book: %v", err)
	}
	// The build is in flight and its credit is already spent.
	if err := s.debitBuildCredit(t.Context(), pass.ID, book.ID); err != nil {
		t.Fatalf("debit: %v", err)
	}

	s.epubRunner = func(int64, dbgen.Book) error {
		return errors.New("pandoc epub: exit status 43")
	}
	if err := s.finalizeBuild(t.Context(), book.ID, book); err != nil {
		t.Fatalf("finalizeBuild must not fail the build when only the EPUB failed: %v", err)
	}

	var status, errMsg string
	if err := s.DB.QueryRow(`SELECT status, error_msg FROM books WHERE id = ?`, book.ID).Scan(&status, &errMsg); err != nil {
		t.Fatalf("read book: %v", err)
	}
	if status != "ready" {
		t.Errorf("status = %q, want \"ready\" — the PDF is downloadable", status)
	}
	if !strings.HasPrefix(errMsg, "EPUB failed:") {
		t.Errorf("error_msg = %q, want a non-fatal \"EPUB failed: …\" note", errMsg)
	}

	reloaded, err := q.GetPass(t.Context(), pass.ID)
	if err != nil {
		t.Fatalf("reload pass: %v", err)
	}
	if reloaded.BuildsUsed != 1 {
		t.Errorf("builds_used = %d, want 1 — a delivered PDF is not refunded", reloaded.BuildsUsed)
	}
	if n := countLedger(t, s, pass.ID, "build_failed_refund"); n != 0 {
		t.Errorf("a PDF-only build wrote %d refund rows, want 0", n)
	}
}

// TestGenerateEPUBStaysAdminOnly: re-running the EPUB on its own is an
// operator tool, not a customer button (a customer's re-run would be a
// second, unmetered build of the same manuscript).
func TestGenerateEPUBStaysAdminOnly(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	pass, password, clientSlug, _ := grantedPass(t, s, ts, "Curious Customer", "Epub Rerun")
	cookie := clientCookie(t, ts, clientSlug, password)

	resp := uploadBook(t, ts, itoa(pass.ProjectID), cookie, false)
	if resp.StatusCode != 201 {
		t.Fatalf("upload: expected 201, got %d", resp.StatusCode)
	}
	var created map[string]any
	decodeJSON(t, resp, &created)
	bookID := itoa(int64(created["id"].(float64)))

	req, err := http.NewRequest("POST", ts.URL+"/api/books/"+bookID+"/generate-epub", nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.AddCookie(cookie)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do generate-epub: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("customer generate-epub: expected 401, got %d", resp.StatusCode)
	}
}

// TestPublicFactoryOfferPageRoute: GET /factory serves the offer page from
// the on-disk public-docs directory (same as /workshop), so the storefront
// copy can be edited without a rebuild.
func TestPublicFactoryOfferPageRoute(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "factory.html"),
		[]byte("<html><body>Factory Pass offer</body></html>"), 0o600); err != nil {
		t.Fatalf("write public doc: %v", err)
	}
	t.Setenv("PRODCAL_PUBLIC_DOCS", dir)

	_, ts, cleanup := testServer(t)
	defer cleanup()

	resp, err := http.Get(ts.URL + "/factory")
	if err != nil {
		t.Fatalf("get /factory: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("/factory: expected 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("/factory content type = %q, want text/html", ct)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if !strings.Contains(string(body), "Factory Pass offer") {
		t.Errorf("/factory should serve the on-disk page, got %q", string(body))
	}
}
