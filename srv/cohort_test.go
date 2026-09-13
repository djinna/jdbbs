package srv

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestCohortRosterGate(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	seedClient(t, s, "member", "Mike Check", "pw1")
	seedClient(t, s, "outsider", "Some Client", "pw2")
	if _, err := s.DB.Exec(`UPDATE clients SET cohort_slug=? WHERE slug='member'`, workshopSlug); err != nil {
		t.Fatal(err)
	}
	if _, err := s.DB.Exec(`INSERT INTO event_registrations (event_slug,name,email,region,material,goals,status,notes)
		VALUES (?,'Real Member','member@example.com','US','a manuscript','ship it','requested','PRIVATE NOTE'),
		       (?,'Declined Person','no@example.com','EU','x','y','declined',''),
		       (?,'Mike Check','bookiq@gmail.com','US','smoke test','verify','requested','')`, workshopSlug, workshopSlug, workshopSlug); err != nil {
		t.Fatal(err)
	}
	path := "/api/cohort/" + workshopSlug + "/roster"

	// 1) anonymous → 401
	if resp := apiRequest(t, ts, "GET", path, nil); resp.StatusCode != 401 {
		t.Fatalf("anonymous: want 401, got %d", resp.StatusCode)
	}

	// 2) authenticated client outside the cohort → 401
	req, _ := http.NewRequest("GET", ts.URL+path, nil)
	req.AddCookie(clientCookie(t, ts, "outsider", "pw2"))
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != 401 {
		t.Fatalf("outsider: want 401, got %d", resp.StatusCode)
	}

	// 3) cohort member → 200, no private fields, declined omitted
	req, _ = http.NewRequest("GET", ts.URL+path, nil)
	req.AddCookie(clientCookie(t, ts, "member", "pw1"))
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != 200 {
		t.Fatalf("member: want 200, got %d", resp.StatusCode)
	}
	var body struct {
		Count   int                      `json:"count"`
		Members []map[string]interface{} `json:"members"`
	}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Count != 1 || len(body.Members) != 1 {
		t.Fatalf("want 1 member (declined and smoke persona omitted), got %d", body.Count)
	}
	m := body.Members[0]
	for _, k := range []string{"email", "notes", "status", "prep_status", "coupon_code", "ip"} {
		if _, ok := m[k]; ok {
			t.Errorf("roster leaks %q", k)
		}
	}
}

func TestCohortVanityPath(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	res, err := http.Get(ts.URL + "/2026-pi-symposium")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != 200 || !strings.Contains(string(body), "'"+workshopSlug+"'") {
		t.Fatalf("vanity page: status %d, slug injected=%v", res.StatusCode, strings.Contains(string(body), workshopSlug))
	}

	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	res, err = client.Get(ts.URL + "/cohort/" + workshopSlug)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != 301 || res.Header.Get("Location") != "/2026-pi-symposium" {
		t.Fatalf("legacy path: %d -> %q", res.StatusCode, res.Header.Get("Location"))
	}
	res, _ = client.Get(ts.URL + "/cohort/nope")
	res.Body.Close()
	if res.StatusCode != 404 {
		t.Fatalf("unknown cohort: %d", res.StatusCode)
	}
}
