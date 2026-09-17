package srv

// The Factory Pass store: Stripe Checkout in front of fulfillPass.
//
// Shape
//   - Catalog lives in Stripe, keyed by lookup_key (factory-pass, builds-3,
//     storage-6mo). ensureStoreCatalog creates what's missing on start-up,
//     so a fresh sandbox or the live account bootstraps itself from code.
//   - POST /api/public/store/checkout creates a hosted Checkout Session and
//     returns its URL. A pass checkout collects the manuscript title/author
//     as Stripe custom fields; an add-on checkout is tied to a pass id.
//   - Fulfilment is one function, fulfillStoreSession, idempotent by session
//     id (store_orders.stripe_session_id UNIQUE + an in-process mutex). Two
//     callers: the thanks page (GET /api/public/store/session) and the event
//     poller (checkout.session.completed every 60 s) for buyers who closed
//     the tab. exe.dev's proxy recommends polling over webhooks.
//   - PRODCAL_STORE=on enables all of it. Off (default): routes 404, no
//     poller, no catalog calls — so the freeze build is inert.
//
// Discounts are Stripe promotion codes entered on the Checkout page. The
// first, PYB149, is $200 off the pass only (never the add-ons).

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"srv.exe.dev/db/dbgen"
)

// ---- catalog ----

type storeItem struct {
	LookupKey   string
	Name        string
	Description string
	Amount      int64 // cents
	Builds      int64 // credits granted per unit
	Months      int64 // storage months granted per unit
}

var storeCatalog = []storeItem{
	{LookupKey: "factory-pass", Name: "Factory Pass", Description: "One manuscript through the jdbb studio book factory: transmittal, generated Word template, unlimited preflight, unlimited EPUB builds, three print PDF builds, six months of storage.", Amount: 34900},
	{LookupKey: "builds-3", Name: "+3 print builds", Description: "Three more print PDF builds on the same pass. EPUB builds are always unlimited.", Amount: 4900, Builds: 3},
	{LookupKey: "storage-6mo", Name: "+6 months storage", Description: "Keeps the project rebuildable and downloadable for six more months.", Amount: 2900, Months: 6},
}

const (
	storePassKey     = "factory-pass"
	storePromoCode   = "PYB149"
	storePromoOff    = 20000 // cents
	storePromoCoupon = "pyb149-200-off"
)

func storeItemByKey(key string) *storeItem {
	for i := range storeCatalog {
		if storeCatalog[i].LookupKey == key {
			return &storeCatalog[i]
		}
	}
	return nil
}

// storeEnabled: PRODCAL_STORE=on|1|true.
func storeEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("PRODCAL_STORE")))
	return v == "on" || v == "1" || v == "true"
}

// store holds the Stripe client, the resolved price ids, and the fulfilment
// mutex. Nil on the Server when the store is off.
type store struct {
	stripe *stripeClient
	mu     sync.Mutex // serialises fulfilment (thanks page vs poller)

	pricesMu sync.RWMutex
	prices   map[string]stripePrice // lookup_key → price
	product  string                 // product id of the pass (promo restriction)
}

func (s *Server) initStore() {
	if !storeEnabled() {
		return
	}
	s.Store = &store{stripe: newStripeClient(), prices: map[string]stripePrice{}}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		if err := s.Store.ensureCatalog(ctx); err != nil {
			slog.Error("store: catalog bootstrap failed (will retry on first checkout)", "err", err)
		} else {
			slog.Info("store: catalog ready", "prices", len(s.Store.prices))
		}
	}()
	go s.storePoller()
}

// ensureCatalog resolves every catalog item to an active Stripe price,
// creating price+product when missing, then makes sure the promo exists.
// Safe to call repeatedly; it is keyed entirely by lookup_key / fixed ids.
func (st *store) ensureCatalog(ctx context.Context) error {
	var list stripeList[stripePrice]
	form := stripeForm{"active": "true", "limit": "100"}
	for i, it := range storeCatalog {
		form[fmt.Sprintf("lookup_keys[%d]", i)] = it.LookupKey
	}
	if err := st.stripe.do(ctx, http.MethodGet, "/v1/prices", form, &list); err != nil {
		return fmt.Errorf("list prices: %w", err)
	}
	found := map[string]stripePrice{}
	for _, p := range list.Data {
		found[p.LookupKey] = p
	}
	for _, it := range storeCatalog {
		p, ok := found[it.LookupKey]
		if ok && p.UnitAmount == it.Amount {
			continue
		}
		// Missing, or the amount in code changed: create a new price and move
		// the lookup_key to it (transfer_lookup_key). Old price stays for
		// old sessions; Stripe prices are immutable by design.
		var created stripePrice
		f := stripeForm{
			"currency":                           "usd",
			"unit_amount":                        strconv.FormatInt(it.Amount, 10),
			"lookup_key":                         it.LookupKey,
			"transfer_lookup_key":                "true",
			"product_data[name]":                 it.Name,
			"product_data[metadata][sku]":        it.LookupKey,
			"metadata[sku]":                      it.LookupKey,
			"product_data[statement_descriptor]": "JDBB STUDIO",
		}
		if ok {
			// Reuse the product; only the price changes.
			delete(f, "product_data[name]")
			delete(f, "product_data[metadata][sku]")
			delete(f, "product_data[statement_descriptor]")
			f["product"] = p.Product
		}
		if err := st.stripe.do(ctx, http.MethodPost, "/v1/prices", f, &created); err != nil {
			return fmt.Errorf("create price %s: %w", it.LookupKey, err)
		}
		if it.Description != "" {
			_ = st.stripe.do(ctx, http.MethodPost, "/v1/products/"+created.Product,
				stripeForm{"description": it.Description}, nil)
		}
		slog.Info("store: created price", "lookup_key", it.LookupKey, "price", created.ID, "amount", it.Amount)
		found[it.LookupKey] = created
	}
	st.pricesMu.Lock()
	st.prices = found
	st.product = found[storePassKey].Product
	st.pricesMu.Unlock()
	return st.ensurePromo(ctx)
}

// ensurePromo: a $200-off coupon restricted to the pass product, exposed as
// promotion code PYB149 (one use per customer). Idempotent by fixed ids.
func (st *store) ensurePromo(ctx context.Context) error {
	var c stripeCoupon
	err := st.stripe.do(ctx, http.MethodGet, "/v1/coupons/"+storePromoCoupon, nil, &c)
	if isStripeNotFound(err) {
		err = st.stripe.do(ctx, http.MethodPost, "/v1/coupons", stripeForm{
			"id":                      storePromoCoupon,
			"name":                    "Workshop alumni: $200 off a Factory Pass",
			"amount_off":              strconv.FormatInt(storePromoOff, 10),
			"currency":                "usd",
			"duration":                "once",
			"applies_to[products][0]": st.product,
		}, &c)
		if err == nil {
			slog.Info("store: created coupon", "id", c.ID)
		}
	}
	if err != nil {
		return fmt.Errorf("coupon: %w", err)
	}
	var codes stripeList[stripePromotionCode]
	if err := st.stripe.do(ctx, http.MethodGet, "/v1/promotion_codes", stripeForm{"code": storePromoCode, "limit": "1"}, &codes); err != nil {
		return fmt.Errorf("list promo codes: %w", err)
	}
	if len(codes.Data) > 0 {
		return nil
	}
	expires := time.Date(2026, 10, 31, 23, 59, 59, 0, time.UTC).Unix()
	var pc stripePromotionCode
	if err := st.stripe.do(ctx, http.MethodPost, "/v1/promotion_codes", stripeForm{
		"coupon":                                storePromoCoupon,
		"code":                                  storePromoCode,
		"expires_at":                            strconv.FormatInt(expires, 10),
		"restrictions[minimum_amount]":          strconv.FormatInt(storePromoOff+100, 10),
		"restrictions[minimum_amount_currency]": "usd",
	}, &pc); err != nil {
		return fmt.Errorf("create promo code: %w", err)
	}
	slog.Info("store: created promotion code", "code", pc.Code)
	return nil
}

// price returns the resolved price for a lookup key, bootstrapping the
// catalog on demand if start-up failed (e.g. the proxy wasn't attached yet).
func (st *store) price(ctx context.Context, key string) (stripePrice, error) {
	st.pricesMu.RLock()
	p, ok := st.prices[key]
	st.pricesMu.RUnlock()
	if ok {
		return p, nil
	}
	if err := st.ensureCatalog(ctx); err != nil {
		return stripePrice{}, err
	}
	st.pricesMu.RLock()
	defer st.pricesMu.RUnlock()
	p, ok = st.prices[key]
	if !ok {
		return stripePrice{}, fmt.Errorf("no price for %q", key)
	}
	return p, nil
}

// ---- checkout ----

type checkoutInput struct {
	Kind      string `json:"kind"`                 // pass | addon
	ProjectID int64  `json:"project_id,omitempty"` // addon: the project whose pass gets the extras
	Items     []struct {
		Key string `json:"key"`
		Qty int64  `json:"qty"`
	} `json:"items,omitempty"` // add-on keys; ignored for kind=pass (optional_items instead)
}

// handleStoreCheckout — POST /api/public/store/checkout → {url}.
// kind=pass: anyone. kind=addon: the pass's logged-in customer (or admin).
func (s *Server) handleStoreCheckout(w http.ResponseWriter, r *http.Request) {
	if s.Store == nil {
		http.NotFound(w, r)
		return
	}
	ip := clientIP(r)
	if !s.limiter().allow("checkout:"+ip, 10, 10*time.Minute) {
		jsonErr(w, "Too many attempts. Please try again in a few minutes.", http.StatusTooManyRequests)
		return
	}
	var in checkoutInput
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16*1024)).Decode(&in); err != nil {
		jsonErr(w, "Sorry, we couldn't read that request.", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	f := stripeForm{
		"mode":                         "payment",
		"success_url":                  s.BaseURL + "/factory/thanks?session_id={CHECKOUT_SESSION_ID}",
		"cancel_url":                   s.BaseURL + "/factory#price",
		"allow_promotion_codes":        "true",
		"billing_address_collection":   "auto",
		"custom_text[submit][message]": "Full refund any time before your first build. Reply to any email from us to reach Jenna.",
	}
	switch in.Kind {
	case "pass":
		pass, err := s.Store.price(ctx, storePassKey)
		if err != nil {
			slog.Error("store: price lookup", "err", err)
			jsonErr(w, "The store isn't reachable right now. Please try again in a minute.", http.StatusBadGateway)
			return
		}
		f["line_items[0][price]"] = pass.ID
		f["line_items[0][quantity]"] = "1"
		f["metadata[kind]"] = "pass"
		// Add-ons offered on the Checkout page itself, off by default.
		i := 0
		for _, key := range []string{"builds-3", "storage-6mo"} {
			p, err := s.Store.price(ctx, key)
			if err != nil {
				continue
			}
			f[fmt.Sprintf("optional_items[%d][price]", i)] = p.ID
			f[fmt.Sprintf("optional_items[%d][quantity]", i)] = "1"
			f[fmt.Sprintf("optional_items[%d][adjustable_quantity][enabled]", i)] = "true"
			f[fmt.Sprintf("optional_items[%d][adjustable_quantity][minimum]", i)] = "1"
			f[fmt.Sprintf("optional_items[%d][adjustable_quantity][maximum]", i)] = "5"
			i++
		}
		f["custom_fields[0][key]"] = "title"
		f["custom_fields[0][label][type]"] = "custom"
		f["custom_fields[0][label][custom]"] = "Manuscript title"
		f["custom_fields[0][type]"] = "text"
		f["custom_fields[0][text][maximum_length]"] = "200"
		f["custom_fields[1][key]"] = "author"
		f["custom_fields[1][label][type]"] = "custom"
		f["custom_fields[1][label][custom]"] = "Author name as it should appear in the book"
		f["custom_fields[1][type]"] = "text"
		f["custom_fields[1][optional]"] = "true"
		f["custom_fields[1][text][maximum_length]"] = "120"
	case "addon":
		// The customer must be signed in to this project (or be admin), and
		// the pass must be live — requirePassAccess writes the error itself.
		passPtr, _, ok := s.requirePassAccess(w, r, in.ProjectID)
		if !ok {
			return
		}
		if passPtr == nil {
			jsonErr(w, "This project has no Factory Pass to add to.", http.StatusNotFound)
			return
		}
		pass := *passPtr
		if len(in.Items) == 0 {
			jsonErr(w, "Choose at least one add-on.", http.StatusBadRequest)
			return
		}
		n := 0
		for _, item := range in.Items {
			it := storeItemByKey(item.Key)
			if it == nil || it.LookupKey == storePassKey || item.Qty < 1 || item.Qty > 5 {
				jsonErr(w, "That add-on isn't available.", http.StatusBadRequest)
				return
			}
			p, err := s.Store.price(ctx, it.LookupKey)
			if err != nil {
				jsonErr(w, "The store isn't reachable right now. Please try again in a minute.", http.StatusBadGateway)
				return
			}
			f[fmt.Sprintf("line_items[%d][price]", n)] = p.ID
			f[fmt.Sprintf("line_items[%d][quantity]", n)] = strconv.FormatInt(item.Qty, 10)
			n++
		}
		f["metadata[kind]"] = "addon"
		f["metadata[pass_id]"] = strconv.FormatInt(pass.ID, 10)
		if pass.CustomerEmail != "" {
			f["customer_email"] = pass.CustomerEmail
		}
		// Back to the customer's factory page afterwards.
		if pr, err := dbgen.New(s.DB).GetProject(ctx, pass.ProjectID); err == nil {
			back := s.BaseURL + "/vgr/aog/factory/?client=" + pr.ClientSlug + "&project=" + pr.ProjectSlug
			f["success_url"] = back + "&order={CHECKOUT_SESSION_ID}"
			f["cancel_url"] = back
		}
	default:
		jsonErr(w, "kind must be pass or addon", http.StatusBadRequest)
		return
	}
	f["metadata[base_url]"] = s.BaseURL

	var sess stripeCheckoutSession
	if err := s.Store.stripe.do(ctx, http.MethodPost, "/v1/checkout/sessions", f, &sess); err != nil {
		slog.Error("store: create checkout session", "err", err, "kind", in.Kind)
		jsonErr(w, "Couldn't start checkout. Please try again, or email us.", http.StatusBadGateway)
		return
	}
	slog.Info("store: checkout session created", "id", sess.ID, "kind", in.Kind, "ip", ip)
	jsonOK(w, map[string]any{"url": sess.URL, "id": sess.ID})
}

// ---- fulfilment ----

type storeFulfilResult struct {
	Order     dbgen.StoreOrder
	Pass      *dbgen.Pass
	PortalURL string
	Title     string
	Fresh     bool // true if this call did the work (send email); false if already fulfilled
}

// fetchSession retrieves a session with the expansions fulfilment needs.
func (st *store) fetchSession(ctx context.Context, id string) (*stripeCheckoutSession, error) {
	var sess stripeCheckoutSession
	err := st.stripe.do(ctx, http.MethodGet, "/v1/checkout/sessions/"+id, stripeForm{
		"expand[0]": "line_items",
		"expand[1]": "discounts.promotion_code",
	}, &sess)
	if err != nil {
		return nil, err
	}
	return &sess, nil
}

var errStoreUnpaid = errors.New("session not paid")

// fulfillStoreSession turns a paid Checkout Session into a pass (or add-on
// bump) exactly once. Returns errStoreUnpaid while the session is still
// open. Safe to call from any number of places.
func (s *Server) fulfillStoreSession(ctx context.Context, sessionID string) (*storeFulfilResult, error) {
	if s.Store == nil {
		return nil, errors.New("store off")
	}
	if !strings.HasPrefix(sessionID, "cs_") || len(sessionID) > 200 {
		return nil, errors.New("bad session id")
	}
	s.Store.mu.Lock()
	defer s.Store.mu.Unlock()

	// Already done? (Cheap path first; no Stripe call.)
	if o, err := dbgen.New(s.DB).GetStoreOrderBySession(ctx, sessionID); err == nil {
		return s.storeResultFor(ctx, o, false), nil
	}
	sess, err := s.Store.fetchSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if sess.Status != "complete" || sess.PaymentStatus != "paid" {
		return nil, errStoreUnpaid
	}

	// Line items → what was bought.
	var builds, months int64
	var items []map[string]any
	hasPass := false
	if sess.LineItems != nil {
		for _, li := range sess.LineItems.Data {
			it := storeItemByKey(li.Price.LookupKey)
			if it == nil {
				slog.Warn("store: unknown line item", "price", li.Price.ID, "lookup_key", li.Price.LookupKey)
				continue
			}
			items = append(items, map[string]any{"lookup_key": it.LookupKey, "quantity": li.Quantity})
			if it.LookupKey == storePassKey {
				hasPass = true
			}
			builds += it.Builds * li.Quantity
			months += it.Months * li.Quantity
		}
	}
	itemsJSON, _ := json.Marshal(items)
	email, name := "", ""
	if sess.CustomerDetails != nil {
		email, name = sess.CustomerDetails.Email, sess.CustomerDetails.Name
	}
	promo := sess.promoCode()
	kind := sess.Metadata["kind"]

	var pass *dbgen.Pass
	var portal, title string
	switch {
	case kind == "pass" || (kind == "" && hasPass):
		kind = "pass"
		title = sess.customField("title")
		if title == "" {
			title = "Untitled manuscript"
		}
		if name == "" {
			name = strings.SplitN(email, "@", 2)[0]
		}
		res, err := s.fulfillPass(ctx, "stripe", fulfillPassInput{
			Name: name, Email: email, Title: title, Author: sess.customField("author"),
			Note:            "Stripe " + sessionID,
			StripeSessionID: sessionID, AmountPaid: sess.AmountTotal, PromoCode: promo,
			BuildsExtra: builds, ExtraMonths: months,
		})
		if err != nil {
			return nil, fmt.Errorf("fulfil pass: %w", err)
		}
		pass, portal = &res.Pass, res.PortalURL
		s.sendPassFulfillmentEmail(*res, "stripe")
	case kind == "addon":
		// Extras and the ledger row commit together, so a failed insert can
		// never leave credits applied without the row that stops a re-apply.
		id, _ := strconv.ParseInt(sess.Metadata["pass_id"], 10, 64)
		tx, err := s.DB.BeginTx(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("begin tx: %w", err)
		}
		defer tx.Rollback()
		q := dbgen.New(tx)
		p, err := q.GetPass(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("addon for unknown pass %d", id)
		}
		if err := q.AddPassExtras(ctx, dbgen.AddPassExtrasParams{
			BuildsExtra: builds, Datetime: fmt.Sprintf("+%d months", months), ID: p.ID,
		}); err != nil {
			return nil, fmt.Errorf("add extras: %w", err)
		}
		p, _ = q.GetPass(ctx, p.ID)
		pass = &p
		if pr, err := q.GetProject(ctx, p.ProjectID); err == nil {
			title = pr.Name
			portal = s.portalURL(pr.ClientSlug, pr.ProjectSlug)
		}
		order, err := q.CreateStoreOrder(ctx, storeOrderParams(sessionID, kind, pass, email, name, promo, string(itemsJSON), sess))
		if err != nil {
			return nil, fmt.Errorf("order row: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit: %w", err)
		}
		slog.Info("store: fulfilled", "session", sessionID, "kind", kind, "amount", sess.AmountTotal, "promo", promo, "email", email)
		s.sendAddonEmail(*pass, order, title, portal)
		return &storeFulfilResult{Order: order, Pass: pass, PortalURL: portal, Title: title, Fresh: true}, nil
	default:
		return nil, fmt.Errorf("session %s: nothing recognisable was bought", sessionID)
	}

	// Pass path: the pass row already carries stripe_session_id (unique), so
	// even if this insert fails a retry cannot mint a second pass.
	order, err := dbgen.New(s.DB).CreateStoreOrder(ctx, storeOrderParams(sessionID, kind, pass, email, name, promo, string(itemsJSON), sess))
	if err != nil {
		slog.Error("store: order row failed after pass fulfilment", "err", err, "session", sessionID)
		return nil, err
	}
	slog.Info("store: fulfilled", "session", sessionID, "kind", kind, "amount", sess.AmountTotal, "promo", promo, "email", email)
	return &storeFulfilResult{Order: order, Pass: pass, PortalURL: portal, Title: title, Fresh: true}, nil
}

func storeOrderParams(sessionID, kind string, pass *dbgen.Pass, email, name, promo, items string, sess *stripeCheckoutSession) dbgen.CreateStoreOrderParams {
	var passID sql.NullInt64
	if pass != nil {
		passID = sql.NullInt64{Int64: pass.ID, Valid: true}
	}
	return dbgen.CreateStoreOrderParams{
		StripeSessionID: sessionID, Kind: kind, PassID: passID,
		CustomerEmail: email, CustomerName: name,
		AmountTotal: sess.AmountTotal, Currency: sess.Currency, PromoCode: promo,
		Items: items, PaymentIntentID: sess.paymentIntentID(),
	}
}

func (s *Server) storeResultFor(ctx context.Context, o dbgen.StoreOrder, fresh bool) *storeFulfilResult {
	res := &storeFulfilResult{Order: o, Fresh: fresh}
	if o.PassID.Valid {
		if p, err := dbgen.New(s.DB).GetPass(ctx, o.PassID.Int64); err == nil {
			res.Pass = &p
			if pr, err := dbgen.New(s.DB).GetProject(ctx, p.ProjectID); err == nil {
				res.Title = pr.Name
				res.PortalURL = s.portalURL(pr.ClientSlug, pr.ProjectSlug)
			}
		}
	}
	return res
}

// handleStoreSession — GET /api/public/store/session?session_id=cs_…
// The thanks page polls this. Fulfils if paid; reports pending otherwise.
func (s *Server) handleStoreSession(w http.ResponseWriter, r *http.Request) {
	if s.Store == nil {
		http.NotFound(w, r)
		return
	}
	id := strings.TrimSpace(r.URL.Query().Get("session_id"))
	if !s.limiter().allow("session:"+clientIP(r), 30, 10*time.Minute) {
		jsonErr(w, "Too many requests.", http.StatusTooManyRequests)
		return
	}
	res, err := s.fulfillStoreSession(r.Context(), id)
	switch {
	case errors.Is(err, errStoreUnpaid):
		jsonOK(w, map[string]any{"status": "pending"})
		return
	case err != nil:
		slog.Error("store: session fulfilment", "err", err, "session", id)
		jsonErr(w, "We couldn't confirm that payment yet. If you were charged, you'll get an email within a few minutes — or write to us.", http.StatusBadGateway)
		return
	}
	out := map[string]any{
		"status":     "paid",
		"kind":       res.Order.Kind,
		"title":      res.Title,
		"portal_url": res.PortalURL,
		"email":      res.Order.CustomerEmail,
		"amount":     res.Order.AmountTotal,
		"promo_code": res.Order.PromoCode,
	}
	if res.Pass != nil {
		out["credits_remaining"] = passCreditsRemaining(*res.Pass)
		out["expires_at"] = res.Pass.ExpiresAt
	}
	jsonOK(w, out)
}

// ---- poller ----

// storePoller catches sessions whose buyer never came back to the thanks
// page. Every 60 s: list checkout.session.completed events newer than the
// last one seen and fulfil each. Stripe keeps events 30 days, so on a cold
// start we look back 3 days — fulfilment is idempotent so re-seeing is free.
func (s *Server) storePoller() {
	since := time.Now().Add(-72 * time.Hour).Unix()
	time.Sleep(20 * time.Second) // let the catalog bootstrap go first
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		var events stripeList[stripeEvent]
		err := s.Store.stripe.do(ctx, http.MethodGet, "/v1/events", stripeForm{
			"type":         "checkout.session.completed",
			"created[gte]": strconv.FormatInt(since, 10),
			"limit":        "100",
		}, &events)
		if err != nil {
			slog.Warn("store: poll events", "err", err)
		} else {
			for i := len(events.Data) - 1; i >= 0; i-- { // oldest first
				ev := events.Data[i]
				if ev.Created > since {
					since = ev.Created
				}
				sess := ev.Data.Object
				if sess.PaymentStatus != "paid" {
					continue
				}
				if _, err := dbgen.New(s.DB).GetStoreOrderBySession(ctx, sess.ID); err == nil {
					continue
				}
				if res, err := s.fulfillStoreSession(ctx, sess.ID); err != nil && !errors.Is(err, errStoreUnpaid) {
					slog.Error("store: poller fulfilment", "err", err, "session", sess.ID)
				} else if err == nil && res.Fresh {
					slog.Info("store: poller fulfilled a session the buyer never returned for", "session", sess.ID)
				}
			}
		}
		cancel()
		time.Sleep(60 * time.Second)
	}
}

// handleStoreConfig — GET /api/public/store/config: is the store on, and the
// catalog prices — so factory.html can show the Buy button and the numbers
// from one source of truth. Always 200 so the page's JS is unconditional.
func (s *Server) handleStoreConfig(w http.ResponseWriter, r *http.Request) {
	out := map[string]any{"enabled": s.Store != nil}
	items := map[string]any{}
	for _, it := range storeCatalog {
		items[it.LookupKey] = map[string]any{"name": it.Name, "amount": it.Amount, "display": fmtUSD(it.Amount)}
	}
	out["items"] = items
	out["promo_hint"] = "Promotion codes are entered on the checkout page."
	jsonOK(w, out)
}

// ---- admin ----

// handleAdminStoreOrders — GET /api/admin/store/orders: the ledger, newest first.
func (s *Server) handleAdminStoreOrders(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	rows, err := dbgen.New(s.DB).ListStoreOrders(r.Context(), 500)
	if err != nil {
		jsonErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	type row struct {
		ID              int64  `json:"id"`
		StripeSessionID string `json:"stripe_session_id"`
		PaymentIntentID string `json:"payment_intent_id"`
		Kind            string `json:"kind"`
		PassID          int64  `json:"pass_id"`
		ProjectName     string `json:"project_name"`
		ProjectPath     string `json:"project_path"`
		CustomerEmail   string `json:"customer_email"`
		CustomerName    string `json:"customer_name"`
		AmountTotal     int64  `json:"amount_total"`
		Currency        string `json:"currency"`
		PromoCode       string `json:"promo_code"`
		Items           string `json:"items"`
		FulfilledAt     string `json:"fulfilled_at"`
		Note            string `json:"note"`
	}
	out := make([]row, 0, len(rows))
	for _, o := range rows {
		path := ""
		if o.ClientSlug != "" {
			path = "/" + o.ClientSlug + "/" + o.ProjectSlug + "/factory/"
		}
		out = append(out, row{
			ID: o.ID, StripeSessionID: o.StripeSessionID, PaymentIntentID: o.PaymentIntentID, Kind: o.Kind,
			PassID: o.PassID.Int64, ProjectName: o.ProjectName, ProjectPath: path,
			CustomerEmail: o.CustomerEmail, CustomerName: o.CustomerName,
			AmountTotal: o.AmountTotal, Currency: o.Currency, PromoCode: o.PromoCode,
			Items: storeItemsSummary(o.Items), FulfilledAt: o.FulfilledAt.UTC().Format(time.RFC3339), Note: o.Note,
		})
	}
	jsonOK(w, map[string]any{"enabled": s.Store != nil, "orders": out})
}

// handleAdminPassStatus — POST /api/admin/passes/{id}/status {"status":"revoked"|"active"}.
// Revoking is the refund companion: money back in the Stripe dashboard, pass
// off here. Reversible (active) for the inevitable mis-click.
func (s *Server) handleAdminPassStatus(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonErr(w, "bad id", http.StatusBadRequest)
		return
	}
	var in struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&in); err != nil {
		jsonErr(w, "bad json", http.StatusBadRequest)
		return
	}
	if in.Status != "revoked" && in.Status != "active" {
		jsonErr(w, "status must be revoked or active", http.StatusBadRequest)
		return
	}
	q := dbgen.New(s.DB)
	p, err := q.GetPass(r.Context(), id)
	if err != nil {
		jsonErr(w, "no such pass", http.StatusNotFound)
		return
	}
	if err := q.UpdatePassStatus(r.Context(), dbgen.UpdatePassStatusParams{Status: in.Status, ID: id}); err != nil {
		jsonErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	slog.Info("admin: pass status", "pass_id", id, "from", p.Status, "to", in.Status)
	jsonOK(w, map[string]any{"ok": true, "id": id, "status": in.Status})
}
