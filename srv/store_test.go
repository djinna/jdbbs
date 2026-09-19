package srv

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"srv.exe.dev/db/dbgen"
)

// fakeStoreStripe is the small portion of Stripe used by the store tests.
type fakeStoreStripe struct {
	t        *testing.T
	sessions map[string]any
	mu       sync.Mutex
	form     url.Values
	server   *httptest.Server
}

func newFakeStoreStripe(t *testing.T) *fakeStoreStripe {
	t.Helper()
	f := &fakeStoreStripe{t: t, sessions: map[string]any{}}
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/prices", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method", http.StatusMethodNotAllowed)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{
			{"id": "price_pass", "lookup_key": "factory-pass", "unit_amount": 54900, "currency": "usd", "active": true, "product": "prod_pass"},
			{"id": "price_b3", "lookup_key": "builds-3", "unit_amount": 9900, "currency": "usd", "active": true, "product": "prod_b3"},
			{"id": "price_s6", "lookup_key": "storage-6mo", "unit_amount": 2900, "currency": "usd", "active": true, "product": "prod_s6"},
		}, "has_more": false})
	})
	mux.HandleFunc("/v1/products/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"id": strings.TrimPrefix(r.URL.Path, "/v1/products/"), "name": "x"})
	})
	mux.HandleFunc("/v1/coupons/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"id": strings.TrimPrefix(r.URL.Path, "/v1/coupons/")})
	})
	mux.HandleFunc("/v1/promotion_codes", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"id": "promo_" + code, "code": code, "active": true}}, "has_more": false})
	})
	mux.HandleFunc("/v1/checkout/sessions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse checkout form: %v", err)
		}
		f.mu.Lock()
		f.form = r.PostForm
		f.mu.Unlock()
		json.NewEncoder(w).Encode(map[string]any{"id": "cs_test_1", "url": "https://checkout.stripe.com/c/pay/cs_test_1"})
	})
	mux.HandleFunc("/v1/checkout/sessions/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/v1/checkout/sessions/")
		f.mu.Lock()
		session, ok := f.sessions[id]
		f.mu.Unlock()
		if !ok {
			http.Error(w, `{"error":{"code":"resource_missing","message":"not found"}}`, http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(session)
	})
	f.server = httptest.NewServer(mux)
	t.Cleanup(f.server.Close)
	return f
}

func (f *fakeStoreStripe) checkoutForm() url.Values {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.form
}

func setupStoreTest(t *testing.T) (*Server, *httptest.Server, *fakeStoreStripe, func()) {
	t.Helper()
	fake := newFakeStoreStripe(t)
	t.Setenv("PRODCAL_STRIPE_URL", fake.server.URL)
	t.Setenv("PRODCAL_STORE", "off") // testServer must not start the real poller.
	s, ts, cleanup := testServer(t)
	s.BaseURL = ts.URL
	s.Store = &store{stripe: newStripeClient(), prices: map[string]stripePrice{
		storePassKey:  {ID: "price_pass", LookupKey: storePassKey, UnitAmount: 34900, Product: "prod_pass"},
		"builds-3":    {ID: "price_b3", LookupKey: "builds-3", UnitAmount: 4900, Product: "prod_b3"},
		"storage-6mo": {ID: "price_s6", LookupKey: "storage-6mo", UnitAmount: 2900, Product: "prod_s6"},
	}}
	return s, ts, fake, cleanup
}

func storePassSession(id string) map[string]any {
	return map[string]any{
		"id": id, "status": "complete", "payment_status": "paid", "amount_total": 14900, "currency": "usd", "payment_intent": "pi_1",
		"metadata":         map[string]string{"kind": "pass"},
		"customer_details": map[string]string{"email": "BUYER@EXAMPLE.COM", "name": "Buyer Person"},
		"custom_fields": []map[string]any{
			{"key": "title", "text": map[string]string{"value": "My Book"}},
			{"key": "author", "text": map[string]string{"value": "B. Person"}},
		},
		"discounts": []map[string]any{{"promotion_code": map[string]string{"id": "promo_1", "code": "PYB149"}}},
		"line_items": map[string]any{"data": []map[string]any{
			{"quantity": 1, "price": map[string]string{"id": "price_pass", "lookup_key": "factory-pass"}},
			{"quantity": 2, "price": map[string]string{"id": "price_b3", "lookup_key": "builds-3"}},
		}},
	}
}

func TestStoreCheckoutPassSession(t *testing.T) {
	_, ts, fake, cleanup := setupStoreTest(t)
	defer cleanup()

	resp := apiRequest(t, ts, http.MethodPost, "/api/public/store/checkout", map[string]string{"kind": "pass"})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("checkout status = %d, want 200", resp.StatusCode)
	}
	var body map[string]string
	decodeJSON(t, resp, &body)
	if body["url"] != "https://checkout.stripe.com/c/pay/cs_test_1" {
		t.Errorf("url = %q", body["url"])
	}
	form := fake.checkoutForm()
	for key, want := range map[string]string{
		"mode":                     "payment",
		"line_items[0][price]":     "price_pass",
		"custom_fields[0][key]":    "title",
		"custom_fields[1][key]":    "author",
		"optional_items[0][price]": "price_b3",
		"optional_items[1][price]": "price_s6",
		"allow_promotion_codes":    "true",
		"metadata[kind]":           "pass",
	} {
		if got := form.Get(key); got != want {
			t.Errorf("Stripe form %s = %q, want %q", key, got, want)
		}
	}
	if !strings.Contains(form.Get("success_url"), "/factory/thanks?session_id={CHECKOUT_SESSION_ID}") {
		t.Errorf("success_url = %q", form.Get("success_url"))
	}
}

func TestStoreCheckout404WhenOff(t *testing.T) {
	t.Setenv("PRODCAL_STORE", "off")
	s, ts, cleanup := testServer(t)
	defer cleanup()
	if s.Store != nil {
		t.Fatal("store unexpectedly enabled")
	}

	resp := apiRequest(t, ts, http.MethodPost, "/api/public/store/checkout", map[string]string{"kind": "pass"})
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("checkout status = %d, want 404", resp.StatusCode)
	}
	resp.Body.Close()
	resp = apiRequest(t, ts, http.MethodGet, "/api/public/store/config", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("config status = %d, want 200", resp.StatusCode)
	}
	var config struct {
		Enabled bool `json:"enabled"`
	}
	decodeJSON(t, resp, &config)
	if config.Enabled {
		t.Error("config enabled = true, want false")
	}
}

func TestStoreFulfilPass(t *testing.T) {
	s, _, fake, cleanup := setupStoreTest(t)
	defer cleanup()
	fake.sessions["cs_pass_1"] = storePassSession("cs_pass_1")

	res, err := s.fulfillStoreSession(context.Background(), "cs_pass_1")
	if err != nil {
		t.Fatalf("fulfil pass: %v", err)
	}
	if !res.Fresh || res.Pass == nil {
		t.Fatalf("result = %#v, want fresh pass", res)
	}
	pass := *res.Pass
	if pass.Source != "stripe" || pass.StripeSessionID != "cs_pass_1" || pass.AmountPaid != 14900 || pass.PromoCode != "PYB149" || pass.BuildsExtra != 6 || pass.CustomerEmail != "buyer@example.com" {
		t.Errorf("pass = %#v", pass)
	}
	if res.Order.Kind != "pass" {
		t.Errorf("order kind = %q, want pass", res.Order.Kind)
	}
	project, err := dbgen.New(s.DB).GetProject(context.Background(), pass.ProjectID)
	if err != nil {
		t.Fatalf("get project: %v", err)
	}
	if project.Name != "My Book" {
		t.Errorf("project name = %q, want My Book", project.Name)
	}

	again, err := s.fulfillStoreSession(context.Background(), "cs_pass_1")
	if err != nil {
		t.Fatalf("fulfil duplicate pass: %v", err)
	}
	if again.Fresh {
		t.Error("duplicate fulfilment was fresh")
	}
	if got := countRows(t, s, "SELECT COUNT(*) FROM passes"); got != 1 {
		t.Errorf("pass count = %d, want 1", got)
	}
	if got := countRows(t, s, "SELECT COUNT(*) FROM store_orders"); got != 1 {
		t.Errorf("store order count = %d, want 1", got)
	}
}

func TestStoreFulfilUnpaid(t *testing.T) {
	s, ts, fake, cleanup := setupStoreTest(t)
	defer cleanup()
	fake.sessions["cs_unpaid_1"] = map[string]any{"id": "cs_unpaid_1", "status": "open", "payment_status": "unpaid"}

	if _, err := s.fulfillStoreSession(context.Background(), "cs_unpaid_1"); err != errStoreUnpaid {
		t.Fatalf("fulfil unpaid error = %v, want errStoreUnpaid", err)
	}
	resp := apiRequest(t, ts, http.MethodGet, "/api/public/store/session?session_id=cs_unpaid_1", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("session status = %d, want 200", resp.StatusCode)
	}
	var body map[string]string
	decodeJSON(t, resp, &body)
	if body["status"] != "pending" {
		t.Errorf("status = %q, want pending", body["status"])
	}
}

func TestStoreFulfilAddon(t *testing.T) {
	s, _, fake, cleanup := setupStoreTest(t)
	defer cleanup()
	seed, err := s.fulfillPass(context.Background(), "admin", fulfillPassInput{Name: "Existing Buyer", Email: "buyer@example.com", Title: "Existing Book"})
	if err != nil {
		t.Fatalf("seed pass: %v", err)
	}
	before := seed.Pass
	fake.sessions["cs_addon_1"] = map[string]any{
		"id": "cs_addon_1", "status": "complete", "payment_status": "paid", "amount_total": 7800, "currency": "usd", "payment_intent": "pi_addon",
		"metadata":         map[string]string{"kind": "addon", "pass_id": itoa(before.ID)},
		"customer_details": map[string]string{"email": "buyer@example.com", "name": "Existing Buyer"},
		"line_items": map[string]any{"data": []map[string]any{
			{"quantity": 1, "price": map[string]string{"id": "price_b3", "lookup_key": "builds-3"}},
			{"quantity": 1, "price": map[string]string{"id": "price_s6", "lookup_key": "storage-6mo"}},
		}},
	}

	res, err := s.fulfillStoreSession(context.Background(), "cs_addon_1")
	if err != nil {
		t.Fatalf("fulfil addon: %v", err)
	}
	if !res.Fresh || res.Pass == nil || res.Order.Kind != "addon" || !res.Order.PassID.Valid || res.Order.PassID.Int64 != before.ID {
		t.Fatalf("addon result = %#v", res)
	}
	after, err := dbgen.New(s.DB).GetPass(context.Background(), before.ID)
	if err != nil {
		t.Fatalf("reload pass: %v", err)
	}
	if after.BuildsExtra != before.BuildsExtra+3 {
		t.Errorf("builds_extra = %d, want %d", after.BuildsExtra, before.BuildsExtra+3)
	}
	if !after.ExpiresAt.After(before.ExpiresAt.AddDate(0, 5, 20)) {
		t.Errorf("expires_at moved from %s to %s, want about six months", before.ExpiresAt, after.ExpiresAt)
	}
}

func TestStoreSessionEndpointBadID(t *testing.T) {
	_, ts, _, cleanup := setupStoreTest(t)
	defer cleanup()
	resp := apiRequest(t, ts, http.MethodGet, "/api/public/store/session?session_id=nonsense", nil)
	if resp.StatusCode == http.StatusOK {
		t.Error("bad session id returned 200")
	}
	var body map[string]string
	decodeJSON(t, resp, &body)
	if body["error"] == "" {
		t.Errorf("error response = %#v, want JSON error", body)
	}
}
