package srv

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	dbgen "srv.exe.dev/db/dbgen"
)

// tokenRequest is a machine caller: no browser, no cookie, just the project
// token in X-Auth-Token.
func tokenRequest(t *testing.T, ts *httptest.Server, token, method, path string, body any) *http.Response {
	t.Helper()
	var rd *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewBuffer(b)
	} else {
		rd = &bytes.Buffer{}
	}
	req, err := http.NewRequest(method, ts.URL+path, rd)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-Token", token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

// TestMachineFactoryFiveCalls is the "your factory calls my factory" path
// from MACHINE-FACTORY-SPEC-2026-09-18.md, with no browser in the loop:
// token → PUT final transmittal → upload → convert (with callback_url) →
// GET /api/books/{id} until ready. The spec must reflect the transmittal at
// build time without the caller ever downloading the Word template, and the
// callback must receive the same status JSON the GET returns.
func TestMachineFactoryFiveCalls(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()
	s.allowLocalCallbacks = true

	pass, _, _, _ := grantedPass(t, s, ts, "Robot Author", "Machine Book")
	pid := itoa(pass.ProjectID)

	// 0. The studio mints a project token (admin, once).
	resp := apiRequestAdmin(t, ts, "POST", "/api/projects/"+pid+"/auth", map[string]string{"password": "machine-secret"})
	if resp.StatusCode != 200 {
		t.Fatalf("mint token: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()
	const token = "machine-secret"

	// 1. Transmittal: read the default shape, fill it, mark final.
	resp = tokenRequest(t, ts, token, "GET", "/api/projects/"+pid+"/transmittal", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("get transmittal with token: expected 200, got %d", resp.StatusCode)
	}
	var tx map[string]any
	decodeJSON(t, resp, &tx)
	data := tx["data"].(map[string]any)
	data["design"].(map[string]any)["trim"] = "6 x 9"
	resp = tokenRequest(t, ts, token, "PUT", "/api/projects/"+pid+"/transmittal", map[string]any{"status": "final", "data": data})
	if resp.StatusCode != 200 {
		t.Fatalf("put transmittal: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 2. Manuscript upload, multipart with the token header.
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("title", "Machine Book")
	_ = mw.WriteField("author", "Robot Author")
	_ = mw.WriteField("project_id", pid)
	fw, _ := mw.CreateFormFile("file", "machine.docx")
	_, _ = fw.Write([]byte("not-a-real-docx"))
	_ = mw.Close()
	req, _ := http.NewRequest("POST", ts.URL+"/api/books/upload", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("X-Auth-Token", token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 201 {
		t.Fatalf("upload with token: expected 201, got %d", resp.StatusCode)
	}
	var created map[string]any
	decodeJSON(t, resp, &created)
	bookID := itoa(int64(created["id"].(float64)))

	// A receiver for the completion callback.
	var mu sync.Mutex
	var got []map[string]any
	cb := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		mu.Lock()
		got = append(got, body)
		mu.Unlock()
		w.WriteHeader(204)
	}))
	defer cb.Close()

	// Fake EPUB stage (the PDF stage needs pandoc/typst).
	q := dbgen.New(s.DB)
	s.epubRunner = func(bid int64, book dbgen.Book) error {
		_, err := q.CreateBookOutput(t.Context(), dbgen.CreateBookOutputParams{
			BookID: bid, OutputFormat: "epub", OutputData: []byte("PK"), SourceFilename: book.SourceFilename,
		})
		return err
	}

	// 3. Build. The spec has never been synced (no template download).
	resp = tokenRequest(t, ts, token, "POST", "/api/books/"+bookID+"/convert",
		map[string]string{"format": "epub", "callback_url": cb.URL + "/done?secret=x"})
	if resp.StatusCode != 200 {
		t.Fatalf("convert: expected 200, got %d", resp.StatusCode)
	}
	var started map[string]any
	decodeJSON(t, resp, &started)
	if started["status_url"] != "/api/books/"+bookID {
		t.Errorf("convert should return status_url, got %v", started)
	}

	// The convert call synced the final transmittal into the spec.
	spec, err := q.GetBookSpec(t.Context(), pass.ProjectID)
	if err != nil {
		t.Fatalf("spec after convert: %v", err)
	}
	if !strings.Contains(spec.Data, `"trim":"6 x 9"`) && !strings.Contains(spec.Data, `"trim": "6 x 9"`) {
		t.Errorf("spec was not refreshed from the final transmittal at build time: %s", spec.Data)
	}

	// 4. Completion: poll GET /api/books/{id}.
	var st map[string]any
	deadline := time.Now().Add(5 * time.Second)
	for {
		resp = tokenRequest(t, ts, token, "GET", "/api/books/"+bookID, nil)
		if resp.StatusCode != 200 {
			t.Fatalf("get book: expected 200, got %d", resp.StatusCode)
		}
		decodeJSON(t, resp, &st)
		if st["status"] == "ready" || st["status"] == "error" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("build never finished: %v", st)
		}
		time.Sleep(20 * time.Millisecond)
	}
	if st["status"] != "ready" {
		t.Fatalf("build status %v, error %v", st["status"], st["error"])
	}
	outs := st["outputs"].([]any)
	if len(outs) != 1 {
		t.Fatalf("expected one output, got %v", outs)
	}
	o := outs[0].(map[string]any)
	if o["format"] != "epub" || !strings.HasPrefix(o["download_url"].(string), "/api/books/"+bookID+"/outputs/") {
		t.Errorf("unexpected output row: %v", o)
	}
	// 5. …and the download works with the token.
	resp = tokenRequest(t, ts, token, "GET", o["download_url"].(string), nil)
	if resp.StatusCode != 200 {
		t.Errorf("download with token: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// The callback arrived once, with the same shape.
	deadline = time.Now().Add(3 * time.Second)
	for {
		mu.Lock()
		n := len(got)
		mu.Unlock()
		if n > 0 || time.Now().After(deadline) {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(got) != 1 {
		t.Fatalf("callback called %d times, want 1", len(got))
	}
	if got[0]["status"] != "ready" || got[0]["book_id"] != st["book_id"] {
		t.Errorf("callback body %v does not match GET %v", got[0], st)
	}

	// No token, no entry.
	resp = apiRequest(t, ts, "GET", "/api/books/"+bookID, nil)
	if resp.StatusCode != 401 && resp.StatusCode != 403 {
		t.Errorf("anonymous GET /api/books/{id}: expected 401/403, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

// TestCallbackURLRejectsLocalTargets: the callback is an outbound POST from
// the server, so it must not be aimable at the server itself or the LAN.
func TestCallbackURLRejectsLocalTargets(t *testing.T) {
	bad := []string{
		"http://localhost:8000/x", "http://127.0.0.1/x", "http://10.0.0.5/hook",
		"http://192.168.1.1/", "http://[::1]/", "ftp://example.com/x",
		"http://user:pw@example.com/x", "http://169.254.169.254/latest/meta-data",
	}
	for _, u := range bad {
		if err := validateCallbackURL(u, false); err == nil {
			t.Errorf("validateCallbackURL(%q) accepted", u)
		}
	}
}

// TestBearerTokenAuth: "Authorization: Bearer <token>" is a third spelling
// of the project token, equivalent to the cookie and X-Auth-Token — so a
// caller's stock HTTP tooling (which usually has a bearer option and no
// custom-header option) works unchanged. Wrong scheme or wrong token stays out.
func TestBearerTokenAuth(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	pass, _, _, _ := grantedPass(t, s, ts, "Bearer Author", "Bearer Book")
	pid := itoa(pass.ProjectID)
	resp := apiRequestAdmin(t, ts, "POST", "/api/projects/"+pid+"/auth", map[string]string{"password": "bearer-secret"})
	if resp.StatusCode != 200 {
		t.Fatalf("mint token: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	get := func(auth string) int {
		req, _ := http.NewRequest("GET", ts.URL+"/api/projects/"+pid+"/transmittal", nil)
		if auth != "" {
			req.Header.Set("Authorization", auth)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		return resp.StatusCode
	}
	if got := get("Bearer bearer-secret"); got != 200 {
		t.Errorf("Bearer with the project token: expected 200, got %d", got)
	}
	if got := get("bearer bearer-secret"); got != 200 {
		t.Errorf("scheme is case-insensitive: expected 200, got %d", got)
	}
	if got := get("Bearer wrong"); got != 401 {
		t.Errorf("Bearer with a wrong token: expected 401, got %d", got)
	}
	if got := get("Basic bearer-secret"); got != 401 {
		t.Errorf("non-Bearer scheme must not be read as a token: expected 401, got %d", got)
	}
	if got := get(""); got != 401 {
		t.Errorf("no auth: expected 401, got %d", got)
	}
}

// TestCallbackDialRefusesRebind: a callback host that resolved to a public
// address when the build was submitted, but to loopback when the build
// finishes (DNS rebinding), must be refused at dial time — the connection is
// never made. Same check, applied to the addresses actually dialled.
func TestCallbackDialRefusesRebind(t *testing.T) {
	hits := 0
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { hits++ }))
	defer target.Close()
	_, port, _ := net.SplitHostPort(strings.TrimPrefix(target.URL, "http://"))

	answer := []net.IPAddr{{IP: net.ParseIP("93.184.216.34")}} // public at validation time
	orig := callbackLookup
	callbackLookup = func(ctx context.Context, host string) ([]net.IPAddr, error) {
		if host == "hook.example.test" {
			return answer, nil
		}
		return orig(ctx, host)
	}
	defer func() { callbackLookup = orig }()

	u := "http://hook.example.test:" + port + "/done"
	if err := validateCallbackURL(u, false); err != nil {
		t.Fatalf("validation with public answer should pass: %v", err)
	}

	// Rebind: the name now points at the loopback listener.
	answer = []net.IPAddr{{IP: net.ParseIP("127.0.0.1")}}
	client := &http.Client{Timeout: 5 * time.Second, Transport: callbackTransport(false)}
	_, err := client.Post(u, "application/json", strings.NewReader("{}"))
	if err == nil || !strings.Contains(err.Error(), "private or local") {
		t.Fatalf("dial should refuse rebound loopback target, got err=%v", err)
	}
	if hits != 0 {
		t.Fatalf("loopback listener was reached %d times", hits)
	}

	// Literal IPs are checked at dial time too, regardless of validation.
	if _, err := client.Post(target.URL+"/done", "application/json", strings.NewReader("{}")); err == nil || hits != 0 {
		t.Fatalf("literal loopback should be refused at dial, err=%v hits=%d", err, hits)
	}
	// And with allowLocal (tests only) the same listener is reachable.
	local := &http.Client{Timeout: 5 * time.Second, Transport: callbackTransport(true)}
	if resp, err := local.Post(target.URL+"/done", "application/json", strings.NewReader("{}")); err != nil || hits != 1 {
		t.Fatalf("allowLocal dial failed: err=%v hits=%d", err, hits)
	} else {
		resp.Body.Close()
	}
}

// TestBlockedCallbackIP covers the address classes the factory must never
// POST to, including wrapped forms that sidestep a naive IPv4 check.
func TestBlockedCallbackIP(t *testing.T) {
	blocked := []string{"127.0.0.1", "10.1.2.3", "172.16.0.1", "192.168.0.1", "169.254.169.254",
		"0.0.0.0", "0.1.2.3", "100.64.0.1", "224.0.0.1", "::1", "fc00::1", "fe80::1", "::",
		"::ffff:127.0.0.1", "::ffff:10.0.0.1", "2002:7f00:0001::1", "2002:0a00:0001::1"}
	for _, s := range blocked {
		if !blockedCallbackIP(net.ParseIP(s)) {
			t.Errorf("%s should be blocked", s)
		}
	}
	allowed := []string{"93.184.216.34", "8.8.8.8", "2606:4700::1111", "2002:5db8:d822::1"}
	for _, s := range allowed {
		if blockedCallbackIP(net.ParseIP(s)) {
			t.Errorf("%s should be allowed", s)
		}
	}
}
