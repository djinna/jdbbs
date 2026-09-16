package srv

// Minimal Stripe REST client. No SDK: the key never reaches this process —
// requests go to the exe.dev integration proxy (https://stripe.int.exe.xyz),
// which injects Authorization at the network edge. We need six endpoints
// (prices, coupons, promotion codes, checkout sessions, events), all
// form-encoded POST / JSON GET, so a hundred lines of net/http beats a
// dependency that fights custom base URLs.
//
//	PRODCAL_STRIPE_URL  — base URL (default https://stripe.int.exe.xyz)
//	PRODCAL_STRIPE_KEY  — optional: send our own bearer (local Mac, no proxy)

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const stripeDefaultURL = "https://stripe.int.exe.xyz"

type stripeClient struct {
	base string
	key  string
	http *http.Client
}

func newStripeClient() *stripeClient {
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("PRODCAL_STRIPE_URL")), "/")
	if base == "" {
		base = stripeDefaultURL
	}
	return &stripeClient{
		base: base,
		key:  strings.TrimSpace(os.Getenv("PRODCAL_STRIPE_KEY")),
		http: &http.Client{Timeout: 30 * time.Second},
	}
}

// stripeError is a non-2xx response, carrying Stripe's own message so the
// log says "No such price: 'price_x'" rather than "400".
type stripeError struct {
	Status  int
	Type    string `json:"type"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Param   string `json:"param"`
}

func (e *stripeError) Error() string {
	return fmt.Sprintf("stripe %d %s: %s", e.Status, e.Code, e.Message)
}

// stripeForm builds Stripe's bracketed form encoding: {"a[b]": "x"} etc.
// Callers write keys literally ("line_items[0][price]") — it reads like the
// docs and there's no nesting to get wrong.
type stripeForm map[string]string

func (f stripeForm) values() url.Values {
	v := url.Values{}
	for k, val := range f {
		v.Set(k, val)
	}
	return v
}

// do performs one request. form is sent as the body for POST and as the
// query string for GET/DELETE. out, if non-nil, receives the JSON body.
func (c *stripeClient) do(ctx context.Context, method, path string, form stripeForm, out any) error {
	u := c.base + path
	var body io.Reader
	if form != nil {
		enc := form.values().Encode()
		if method == http.MethodPost {
			body = strings.NewReader(enc)
		} else if enc != "" {
			sep := "?"
			if strings.Contains(u, "?") {
				sep = "&"
			}
			u += sep + enc
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return err
	}
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if c.key != "" {
		req.Header.Set("Authorization", "Bearer "+c.key)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("stripe %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return fmt.Errorf("stripe read: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var wrap struct {
			Error stripeError `json:"error"`
		}
		_ = json.Unmarshal(raw, &wrap)
		wrap.Error.Status = resp.StatusCode
		if wrap.Error.Message == "" {
			wrap.Error.Message = strings.TrimSpace(string(raw[:min(len(raw), 300)]))
		}
		return &wrap.Error
	}
	if out != nil {
		if err := json.Unmarshal(raw, out); err != nil {
			return fmt.Errorf("stripe decode %s: %w", path, err)
		}
	}
	return nil
}

// isStripeNotFound reports a 404 / resource_missing from Stripe.
func isStripeNotFound(err error) bool {
	var se *stripeError
	return errors.As(err, &se) && (se.Status == 404 || se.Code == "resource_missing")
}

// ---- Object shapes (only the fields we read) ----

type stripePrice struct {
	ID         string `json:"id"`
	LookupKey  string `json:"lookup_key"`
	UnitAmount int64  `json:"unit_amount"`
	Currency   string `json:"currency"`
	Active     bool   `json:"active"`
	Product    string `json:"product"`
}

type stripeList[T any] struct {
	Data    []T  `json:"data"`
	HasMore bool `json:"has_more"`
}

type stripeCoupon struct {
	ID        string `json:"id"`
	AmountOff int64  `json:"amount_off"`
	Valid     bool   `json:"valid"`
}

type stripePromotionCode struct {
	ID     string `json:"id"`
	Code   string `json:"code"`
	Active bool   `json:"active"`
	Coupon struct {
		ID string `json:"id"`
	} `json:"coupon"`
}

type stripeCheckoutSession struct {
	ID              string            `json:"id"`
	URL             string            `json:"url"`
	Status          string            `json:"status"`         // open | complete | expired
	PaymentStatus   string            `json:"payment_status"` // paid | unpaid | no_payment_required
	AmountTotal     int64             `json:"amount_total"`
	Currency        string            `json:"currency"`
	PaymentIntent   json.RawMessage   `json:"payment_intent"` // id string, or object if expanded
	Metadata        map[string]string `json:"metadata"`
	Created         int64             `json:"created"`
	CustomerDetails *struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	} `json:"customer_details"`
	CustomFields []struct {
		Key  string `json:"key"`
		Text *struct {
			Value string `json:"value"`
		} `json:"text"`
	} `json:"custom_fields"`
	Discounts []struct {
		PromotionCode json.RawMessage `json:"promotion_code"` // id, or object if expanded
	} `json:"discounts"`
	LineItems *stripeList[struct {
		Quantity int64 `json:"quantity"`
		Price    struct {
			ID        string `json:"id"`
			LookupKey string `json:"lookup_key"`
		} `json:"price"`
	}] `json:"line_items"`
}

// customField returns the text value of a custom field by key.
func (cs *stripeCheckoutSession) customField(key string) string {
	for _, f := range cs.CustomFields {
		if f.Key == key && f.Text != nil {
			return strings.TrimSpace(f.Text.Value)
		}
	}
	return ""
}

// promoCode returns the human code of the first promotion code applied,
// which requires the session to have been fetched with
// expand[]=discounts.promotion_code. Falls back to the raw id.
func (cs *stripeCheckoutSession) promoCode() string {
	for _, d := range cs.Discounts {
		if len(d.PromotionCode) == 0 || string(d.PromotionCode) == "null" {
			continue
		}
		var obj struct {
			Code string `json:"code"`
		}
		if err := json.Unmarshal(d.PromotionCode, &obj); err == nil && obj.Code != "" {
			return obj.Code
		}
		var id string
		if err := json.Unmarshal(d.PromotionCode, &id); err == nil {
			return id
		}
	}
	return ""
}

// paymentIntentID handles both the id-string and expanded-object forms.
func (cs *stripeCheckoutSession) paymentIntentID() string {
	if len(cs.PaymentIntent) == 0 {
		return ""
	}
	var id string
	if err := json.Unmarshal(cs.PaymentIntent, &id); err == nil {
		return id
	}
	var obj struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(cs.PaymentIntent, &obj)
	return obj.ID
}

type stripeEvent struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Created int64  `json:"created"`
	Data    struct {
		Object stripeCheckoutSession `json:"object"`
	} `json:"data"`
}
