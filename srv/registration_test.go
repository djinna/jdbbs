package srv

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func decodeMap(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	defer resp.Body.Close()
	var m map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return m
}

func TestRegistrationValidSubmission(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	resp := apiRequest(t, ts, "POST", "/api/public/register", map[string]any{
		"name":          "Ada Lovelace",
		"email":         "Ada@Example.com",
		"region":        "EU",
		"all_sessions":  true,
		"material":      "half-finished essays",
		"consent_email": true,
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}
	if ok, _ := decodeMap(t, resp)["ok"].(bool); !ok {
		t.Fatalf("expected ok:true")
	}

	// Admin list should show exactly one row with a lowercased email.
	lresp := apiRequestAdmin(t, ts, "GET", "/api/admin/registrations", nil)
	m := decodeMap(t, lresp)
	if c, _ := m["count"].(float64); c != 1 {
		t.Fatalf("want count 1, got %v", m["count"])
	}
	regs, _ := m["registrations"].([]any)
	first, _ := regs[0].(map[string]any)
	if first["email"] != "ada@example.com" {
		t.Fatalf("email should be lowercased, got %v", first["email"])
	}
}

func TestRegistrationRequiresMaterial(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()
	resp := apiRequest(t, ts, "POST", "/api/public/register", map[string]any{
		"name": "Bob", "email": "bob@example.com",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("want 400 without material, got %d", resp.StatusCode)
	}
}

func TestRegistrationRejectsBadEmail(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()
	resp := apiRequest(t, ts, "POST", "/api/public/register", map[string]any{
		"name": "Bob", "email": "not-an-email", "material": "x",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("want 400 for bad email, got %d", resp.StatusCode)
	}
}

func TestRegistrationHoneypotSilentlyDropped(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()
	// Honeypot filled → fake success, no row stored.
	resp := apiRequest(t, ts, "POST", "/api/public/register", map[string]any{
		"name": "Spammer", "email": "spam@bot.com", "material": "junk", "company": "AcmeBot",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("honeypot should fake 200, got %d", resp.StatusCode)
	}
	lresp := apiRequestAdmin(t, ts, "GET", "/api/admin/registrations", nil)
	if c, _ := decodeMap(t, lresp)["count"].(float64); c != 0 {
		t.Fatalf("honeypot row should not be stored, count=%v", c)
	}
}

func TestRegistrationUpsertsOnEmail(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()
	for _, mat := range []string{"first material", "revised material"} {
		resp := apiRequest(t, ts, "POST", "/api/public/register", map[string]any{
			"name": "Ada", "email": "ada@example.com", "material": mat,
		})
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("submit %q: want 200, got %d", mat, resp.StatusCode)
		}
	}
	lresp := apiRequestAdmin(t, ts, "GET", "/api/admin/registrations", nil)
	m := decodeMap(t, lresp)
	if c, _ := m["count"].(float64); c != 1 {
		t.Fatalf("re-submit should upsert, want count 1, got %v", c)
	}
	regs, _ := m["registrations"].([]any)
	first, _ := regs[0].(map[string]any)
	if first["material"] != "revised material" {
		t.Fatalf("upsert should keep latest material, got %v", first["material"])
	}
}

func TestRegistrationAdminRequiresAuth(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()
	resp := apiRequest(t, ts, "GET", "/api/admin/registrations", nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("want 401 without admin header, got %d", resp.StatusCode)
	}
}

func TestRegistrationTrackerUpdate(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()
	apiRequest(t, ts, "POST", "/api/public/register", map[string]any{
		"name": "Ada", "email": "ada@example.com", "material": "essays",
	}).Body.Close()

	resp := apiRequestAdmin(t, ts, "PUT", "/api/admin/registrations/1", map[string]any{
		"status": "confirmed", "prep_status": "ready", "attended_sessions": 5, "notes": "Great fit",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("update: want 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	m := decodeMap(t, apiRequestAdmin(t, ts, "GET", "/api/admin/registrations", nil))
	regs := m["registrations"].([]any)
	got := regs[0].(map[string]any)
	if got["status"] != "confirmed" || got["prep_status"] != "ready" || got["attended_sessions"] != float64(5) || got["notes"] != "Great fit" {
		t.Fatalf("tracker update did not round-trip: %#v", got)
	}
}

func TestAnnouncementHonorsConsentAndLogsSend(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()
	for _, person := range []map[string]any{
		{"name": "Ada Lovelace", "email": "ada@example.com", "material": "essays", "consent_email": true},
		{"name": "No Mail", "email": "no@example.com", "material": "notes", "consent_email": false},
	} {
		apiRequest(t, ts, "POST", "/api/public/register", person).Body.Close()
	}

	var sends atomic.Int32
	var sentTo atomic.Value
	mailAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		_ = json.NewDecoder(r.Body).Decode(&payload)
		to, _ := payload["to"].([]any)
		if len(to) > 0 {
			sentTo.Store(to[0].(string))
		}
		sends.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer mailAPI.Close()
	s.Email = &EmailConfig{APIKey: "test", InboxID: "test@example.com", APIBase: mailAPI.URL}

	resp := apiRequestAdmin(t, ts, "POST", "/api/admin/registrations/announce", map[string]any{
		"registration_ids": []int{1, 2},
		"subject":          "Workshop update",
		"body":             "Bring your manuscript.",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("announce: want 200, got %d", resp.StatusCode)
	}
	out := decodeMap(t, resp)
	if out["selected"] != float64(1) || out["sent"] != float64(1) || sends.Load() != 1 {
		t.Fatalf("only opted-in recipient should be sent: response=%#v sends=%d", out, sends.Load())
	}
	if got, _ := sentTo.Load().(string); got != "ada@example.com" {
		t.Fatalf("sent to %q, want opted-in address", got)
	}

	m := decodeMap(t, apiRequestAdmin(t, ts, "GET", "/api/admin/registrations", nil))
	regs := m["registrations"].([]any)
	byEmail := map[string]map[string]any{}
	for _, raw := range regs {
		row := raw.(map[string]any)
		byEmail[row["email"].(string)] = row
	}
	if byEmail["ada@example.com"]["last_emailed_at"] == "" {
		t.Fatal("successful recipient should get last_emailed_at")
	}
	if byEmail["no@example.com"]["last_emailed_at"] != "" {
		t.Fatal("opted-out recipient must not get last_emailed_at")
	}
	ann := m["announcements"].([]any)
	if len(ann) != 1 || ann[0].(map[string]any)["sent_count"] != float64(1) {
		t.Fatalf("announcement history not logged: %#v", ann)
	}
}

func TestRegistrationCSVExport(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()
	apiRequest(t, ts, "POST", "/api/public/register", map[string]any{
		"name": "Grace, Hopper", "email": "grace@example.com", "material": "memoir\nwith newline",
	}).Body.Close()

	resp := apiRequestAdmin(t, ts, "GET", "/api/admin/registrations.csv", nil)
	defer resp.Body.Close()
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/csv") {
		t.Fatalf("want text/csv, got %q", ct)
	}
	raw, _ := io.ReadAll(resp.Body)
	body := string(raw)
	// Field with comma must be quoted; newline preserved inside quotes.
	if !strings.Contains(body, `"Grace, Hopper"`) {
		t.Fatalf("comma field should be quoted, got:\n%s", body)
	}
	if !strings.Contains(body, `"memoir`+"\n"+`with newline"`) {
		t.Fatalf("newline field should be quoted, got:\n%s", body)
	}
}

func TestMergeAnnouncement(t *testing.T) {
	body := "Hi {{first}}.\n\n{{#code}}Your code is {{code}}. Redeem it at /factory#redeem.{{/code}}\n\n{{#redeemed}}You're already in: {{factory_url}}{{/redeemed}}\n\n{{#nocode}}Your code follows separately.{{/nocode}}\n\nSee you Monday."
	cases := []struct {
		rec  announcementRecipient
		want string
	}{
		{announcementRecipient{Name: "Toby Shorin", Code: "PYB-AAAA-BBBB"},
			"Hi Toby.\n\nYour code is PYB-AAAA-BBBB. Redeem it at /factory#redeem.\n\nSee you Monday."},
		{announcementRecipient{Name: "Mike Casey", Code: "PYB-CCCC-DDDD", FactoryURL: "https://x/mike-casey/casey-001/factory/"},
			"Hi Mike.\n\nYou're already in: https://x/mike-casey/casey-001/factory/\n\nSee you Monday."},
		{announcementRecipient{Name: "No Code"},
			"Hi No.\n\nYour code follows separately.\n\nSee you Monday."},
	}
	for _, c := range cases {
		if got := mergeAnnouncement(body, c.rec); got != c.want {
			t.Errorf("%s:\n got %q\nwant %q", c.rec.Name, got, c.want)
		}
	}
	if m := announcementUnmergedRe.FindString(mergeAnnouncement("code {{code}}", announcementRecipient{Name: "X"})); m != "{{code}}" {
		t.Errorf("bare {{code}} with no code should be left for the guard, got %q", m)
	}
	if m := announcementUnmergedRe.FindString(mergeAnnouncement("{{#code}}oops", announcementRecipient{})); m == "" {
		t.Errorf("unclosed block should be caught")
	}
	h := announcementHTML("A B", "see https://jdbbs.exe.xyz/factory#redeem.")
	if !strings.Contains(h, `<a href="https://jdbbs.exe.xyz/factory#redeem"`) {
		t.Errorf("url not linked: %s", h)
	}
}
