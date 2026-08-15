package srv

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
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
