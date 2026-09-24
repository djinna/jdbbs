package srv

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"log/slog"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"srv.exe.dev/db/dbgen"
)

// ─── Factory Pass ───
//
// A Factory Pass is one manuscript's trip through the book factory: a project,
// unlimited preflights, a fixed number of included builds, and six months of
// storage. Tables live in migration 023; the shared contract with the customer
// page (static/factory.html), the admin tracker, and the public offer page is
// docs/specs/FACTORY-PASS-API-2026-09-03.md.
//
// Fulfillment is a single function (fulfillPass) with a `source`, so the coupon
// path, an admin grant, and a later Stripe webhook all land in the same place.

const (
	passSKU = "factory-pass"

	// passBuildsIncluded is N in "N included builds" — a build is one
	// successful convert. Failed builds are refunded (see failConversion).
	passBuildsIncluded = 3

	// passStorageMonths documents the six-month liveness window. The window
	// itself is computed in SQL (CreatePass) so expires_at always matches
	// fulfilled_at.
	passStorageMonths = 6

	// couponDefaultExpiry is the default coupon expiry for the workshop cohort
	// (codes should be redeemed before session 1 on Sep 21).
	couponDefaultExpiry = "2026-09-30"

	// couponPrefix labels Protocolize-Your-Book codes.
	couponPrefix = "PYB"
)

// codeAlphabet excludes 0/O/1/I so a code survives being read aloud off a
// slide or retyped from a printout.
const codeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// passwordAlphabet is the generated client-password alphabet: same
// unambiguous principle, mixed case for a little more entropy per character.
const passwordAlphabet = "abcdefghjkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// ─── small helpers ───

// randomFrom returns n characters drawn uniformly from alphabet using
// crypto/rand.
func randomFrom(alphabet string, n int) (string, error) {
	var b strings.Builder
	limit := big.NewInt(int64(len(alphabet)))
	for i := 0; i < n; i++ {
		idx, err := rand.Int(rand.Reader, limit)
		if err != nil {
			return "", err
		}
		b.WriteByte(alphabet[idx.Int64()])
	}
	return b.String(), nil
}

// generateCouponCode returns a code shaped PYB-XXXX-XXXX (uppercase, no 0/O/1/I).
func generateCouponCode() (string, error) {
	a, err := randomFrom(codeAlphabet, 4)
	if err != nil {
		return "", err
	}
	b, err := randomFrom(codeAlphabet, 4)
	if err != nil {
		return "", err
	}
	return couponPrefix + "-" + a + "-" + b, nil
}

// generateClientPassword returns the 12-character password mailed to the
// customer on fulfillment (never stored in plaintext).
func generateClientPassword() (string, error) {
	return randomFrom(passwordAlphabet, 12)
}

// normalizeCouponCode makes redemption forgiving about case and stray spaces.
func normalizeCouponCode(s string) string {
	return strings.ToUpper(strings.TrimSpace(s))
}

// passCreditsRemaining implements the contract's
// credits_remaining = builds_included + builds_extra - builds_used.
func passCreditsRemaining(p dbgen.Pass) int64 {
	n := p.BuildsIncluded + p.BuildsExtra - p.BuildsUsed
	if n < 0 {
		return 0
	}
	return n
}

// passLive reports whether a pass may still be used: active and unexpired.
func passLive(p dbgen.Pass) bool {
	return p.Status == "active" && p.ExpiresAt.After(time.Now())
}

// passForProject returns the project's pass, or nil when it has none.
// A missing pass is normal (admin-only projects predate the storefront), so
// this never surfaces an error to the caller — it logs the unexpected ones.
func (s *Server) passForProject(ctx context.Context, projectID int64) *dbgen.Pass {
	if projectID <= 0 {
		return nil
	}
	q := dbgen.New(s.DB)
	p, err := q.GetPassByProject(ctx, projectID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		slog.Error("pass lookup failed", "project_id", projectID, "err", err)
		return nil
	}
	return &p
}

// ─── gating ───

// requirePassAccess is the entitlement gate for the factory endpoints that
// used to be admin-only (upload, detect-chapters, preflight, convert):
//
//	admin header → ok regardless of pass (isAdmin = true)
//	otherwise    → requireAuth(projectID) AND the project's pass is live
//
// The returned pass is nil only when the project has none, which (given the
// rule above) means the caller is an admin. Callers that meter usage must
// debit whenever pass != nil, admin included, so instructor-run builds still
// show up in the ledger.
//
// On failure the response has already been written.
func (s *Server) requirePassAccess(w http.ResponseWriter, r *http.Request, projectID int64) (*dbgen.Pass, bool, bool) {
	if s.isAdmin(r) {
		return s.passForProject(r.Context(), projectID), true, true
	}
	if projectID <= 0 {
		jsonErr(w, "project_id required", http.StatusBadRequest)
		return nil, false, false
	}
	if !s.requireAuth(w, r, projectID) {
		return nil, false, false
	}
	pass := s.passForProject(r.Context(), projectID)
	if pass == nil {
		jsonErr(w, "this project has no factory pass", http.StatusForbidden)
		return nil, false, false
	}
	if !passLive(*pass) {
		jsonErr(w, "this factory pass has expired", http.StatusForbidden)
		return nil, false, false
	}
	return pass, false, true
}

// debitBuildCredit spends one build credit: bumps builds_used and appends a
// `-1 build` ledger row. Called before a conversion starts, so a crash
// mid-build can never hand out a free build; failConversion refunds it.
func (s *Server) debitBuildCredit(ctx context.Context, passID, bookID int64) error {
	q := dbgen.New(s.DB)
	if err := q.IncrementPassBuildsUsed(ctx, passID); err != nil {
		return err
	}
	return q.CreatePassLedgerEntry(ctx, dbgen.CreatePassLedgerEntryParams{
		PassID: passID,
		BookID: sql.NullInt64{Int64: bookID, Valid: bookID > 0},
		Delta:  -1,
		Reason: "build",
	})
}

// refundBuildCredit hands a spent credit back (failed pandoc/typst run).
func (s *Server) refundBuildCredit(ctx context.Context, passID, bookID int64) error {
	q := dbgen.New(s.DB)
	if err := q.DecrementPassBuildsUsed(ctx, passID); err != nil {
		return err
	}
	return q.CreatePassLedgerEntry(ctx, dbgen.CreatePassLedgerEntryParams{
		PassID: passID,
		BookID: sql.NullInt64{Int64: bookID, Valid: bookID > 0},
		Delta:  1,
		Reason: "build_failed_refund",
	})
}

// ─── fulfillment ───

// passInputError is a customer-visible validation failure (bad code, missing
// field). Handlers turn it into a 400; anything else is a 500.
type passInputError struct{ msg string }

func (e passInputError) Error() string { return e.msg }

func passBadInput(format string, args ...any) error {
	return passInputError{msg: fmt.Sprintf(format, args...)}
}

type fulfillPassInput struct {
	Code   string // coupon code (source = "coupon")
	Name   string // customer name → client name + slug
	Email  string
	Title  string // manuscript title → project name + slug
	Author string
	Note   string

	// Paid purchases (source = "stripe") — stamped on the pass inside the
	// same transaction so a crash can never leave a paid pass unlinked from
	// its Checkout Session (which is what makes fulfilment idempotent).
	StripeSessionID string
	AmountPaid      int64 // cents, after discount
	PromoCode       string
	BuildsExtra     int64 // add-ons bought in the same cart
	ExtraMonths     int64
	IndexIncluded   bool // back-of-book index add-on in the same cart
}

type fulfillPassResult struct {
	Pass        dbgen.Pass
	Title       string
	ClientSlug  string
	ProjectSlug string
	Password    string // plaintext, emailed once and never stored
	PortalURL   string
}

// portalURL is the customer's factory page for a fulfilled pass.
func (s *Server) portalURL(clientSlug, projectSlug string) string {
	return fmt.Sprintf("%s/%s/%s/factory/", strings.TrimRight(s.BaseURL, "/"), clientSlug, projectSlug)
}

// fulfillPass creates everything a customer needs to walk into the factory: a
// client (generated password), a project, and the pass itself. source is
// "coupon" (redemption), "admin" (hand-granted), or later "stripe".
//
// It all happens in one transaction: a half-fulfilled pass — a client with no
// project, or a burned coupon with no pass — is the worst possible state.
// Email is sent by the caller afterwards; mail is never part of the tx.
func (s *Server) fulfillPass(ctx context.Context, source string, in fulfillPassInput) (*fulfillPassResult, error) {
	in.Name = clip(in.Name, 200)
	in.Email = clip(strings.ToLower(in.Email), 320)
	in.Title = clip(in.Title, 300)
	in.Author = clip(in.Author, 200)
	in.Note = clip(in.Note, 500)

	if in.Name == "" {
		return nil, passBadInput("Please give your name.")
	}
	if !looksLikeEmail(in.Email) {
		return nil, passBadInput("That email address doesn't look right.")
	}
	if in.Title == "" {
		return nil, passBadInput("Please give the title of the manuscript.")
	}

	password, err := generateClientPassword()
	if err != nil {
		return nil, fmt.Errorf("generate password: %w", err)
	}
	passwordHash, err := hashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()
	q := dbgen.New(tx)

	// 1) Coupon (coupon source only): must exist, be unexpired, and have a
	//    redemption left. Never silently re-fulfil an already-used code.
	var couponID sql.NullInt64
	var coupon dbgen.Coupon
	if source == "coupon" {
		code := normalizeCouponCode(in.Code)
		if code == "" {
			return nil, passBadInput("Please enter your access code.")
		}
		coupon, err = q.GetCouponByCode(ctx, code)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, passBadInput("That code isn't valid. Check it for typos, or email us.")
		}
		if err != nil {
			return nil, fmt.Errorf("look up coupon: %w", err)
		}
		if coupon.ExpiresAt.Valid && !coupon.ExpiresAt.Time.After(time.Now()) {
			return nil, passBadInput("That code has expired. Email us and we'll sort it out.")
		}
		if coupon.RedeemedCount >= coupon.MaxRedemptions {
			return nil, passBadInput("That code has already been redeemed.")
		}
		couponID = sql.NullInt64{Int64: coupon.ID, Valid: true}
	}

	// 2) Client: slug from the author's name when the buyer gave one (the
	// portal is the author's address, not the payer's — C1), else the
	// customer's; unique-ified, random password.
	slugName := strings.TrimSpace(in.Author)
	if slugName == "" {
		slugName = in.Name
	}
	clientSlug, err := uniqueClientSlug(ctx, tx, slugName)
	if err != nil {
		return nil, err
	}
	// A coupon issued to a workshop registration puts the new client in that
	// cohort (clients.cohort_slug), which is what gates the cohort roster.
	cohortSlug := ""
	if couponID.Valid && coupon.RegistrationID.Valid {
		_ = tx.QueryRowContext(ctx, `SELECT event_slug FROM event_registrations WHERE id = ?`,
			coupon.RegistrationID.Int64).Scan(&cohortSlug)
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO clients (slug, name, password_hash, cohort_slug) VALUES (?, ?, ?, ?)`,
		clientSlug, in.Name, passwordHash, cohortSlug); err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	// 3) Project: next book-NNN for this client; the title lives in name.
	projectSlug, err := nextProjectSlug(ctx, tx, clientSlug)
	if err != nil {
		return nil, err
	}
	project, err := q.CreateProject(ctx, dbgen.CreateProjectParams{
		Name:        in.Title,
		StartDate:   time.Now().UTC().Format("2006-01-02"),
		ClientSlug:  clientSlug,
		ProjectSlug: projectSlug,
	})
	if err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}

	// 4) The pass. expires_at = fulfilled_at + 6 months (computed in SQL).
	pass, err := q.CreatePass(ctx, dbgen.CreatePassParams{
		ProjectID:      project.ID,
		Sku:            passSKU,
		Source:         source,
		CouponID:       couponID,
		CustomerEmail:  in.Email,
		CustomerName:   in.Name,
		BuildsIncluded: passBuildsIncluded,
		Note:           in.Note,
	})
	if err != nil {
		return nil, fmt.Errorf("create pass: %w", err)
	}

	// 4b) Seed the transmittal draft with what the form already asked for,
	// so the author does not meet an empty title/author on first open.
	seed, err := seededTransmittalData(in.Title, in.Author)
	if err != nil {
		return nil, fmt.Errorf("seed transmittal: %w", err)
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO transmittals (project_id, status, data) VALUES (?, 'draft', ?)`,
		project.ID, seed); err != nil {
		return nil, fmt.Errorf("seed transmittal: %w", err)
	}
	if in.StripeSessionID != "" {
		if err := q.SetPassPurchase(ctx, dbgen.SetPassPurchaseParams{
			StripeSessionID: in.StripeSessionID, AmountPaid: in.AmountPaid, PromoCode: in.PromoCode, ID: pass.ID,
		}); err != nil {
			return nil, fmt.Errorf("stamp purchase: %w", err)
		}
		pass.StripeSessionID, pass.AmountPaid, pass.PromoCode = in.StripeSessionID, in.AmountPaid, in.PromoCode
	}
	if in.BuildsExtra > 0 || in.ExtraMonths > 0 {
		if err := q.AddPassExtras(ctx, dbgen.AddPassExtrasParams{
			BuildsExtra: in.BuildsExtra, Datetime: fmt.Sprintf("+%d months", in.ExtraMonths), ID: pass.ID,
		}); err != nil {
			return nil, fmt.Errorf("add extras: %w", err)
		}
		pass, err = q.GetPass(ctx, pass.ID)
		if err != nil {
			return nil, fmt.Errorf("reload pass: %w", err)
		}
	}
	if in.IndexIncluded {
		if err := s.grantPassIndex(ctx, q, pass.ID, "purchase"); err != nil {
			return nil, fmt.Errorf("index add-on: %w", err)
		}
		pass.IndexIncluded = 1
	}

	// 5) Burn the redemption and, when the redeeming email matches a workshop
	//    registration, attribute the code to it (if not already bound).
	if couponID.Valid {
		if err := q.IncrementCouponRedeemed(ctx, coupon.ID); err != nil {
			return nil, fmt.Errorf("mark coupon redeemed: %w", err)
		}
		if !coupon.RegistrationID.Valid {
			regID, rErr := q.FindRegistrationIDByEmail(ctx, in.Email)
			switch {
			case rErr == nil:
				if err := q.SetCouponRegistration(ctx, dbgen.SetCouponRegistrationParams{
					RegistrationID: sql.NullInt64{Int64: regID, Valid: true},
					ID:             coupon.ID,
				}); err != nil {
					return nil, fmt.Errorf("link coupon registration: %w", err)
				}
				// Late attribution: the client was created before we knew the
				// registration, so set its cohort now.
				if _, err := tx.ExecContext(ctx, `UPDATE clients SET cohort_slug =
					(SELECT event_slug FROM event_registrations WHERE id = ?) WHERE slug = ?`,
					regID, clientSlug); err != nil {
					return nil, fmt.Errorf("set client cohort: %w", err)
				}
			case errors.Is(rErr, sql.ErrNoRows):
				// Redeemed with an email we don't recognize — fine, just
				// unattributed.
			default:
				return nil, fmt.Errorf("find registration: %w", rErr)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	slog.Info("factory pass fulfilled", "source", source, "pass_id", pass.ID,
		"project", clientSlug+"/"+projectSlug, "email", in.Email)
	s.factoryEvent(pass.ProjectID, clientSlug, "pass.fulfilled", source,
		fmt.Sprintf("%s <%s> → /%s/%s/", in.Name, in.Email, clientSlug, projectSlug))

	return &fulfillPassResult{
		Pass:        pass,
		Title:       in.Title,
		ClientSlug:  clientSlug,
		ProjectSlug: projectSlug,
		Password:    password,
		PortalURL:   s.portalURL(clientSlug, projectSlug),
	}, nil
}

// maxSlugLen keeps generated slugs short enough to read in a URL.
const maxSlugLen = 24

// truncateSlug trims a normalized slug to maxSlugLen without leaving a
// trailing hyphen.
func truncateSlug(slug string) string {
	if len(slug) <= maxSlugLen {
		return slug
	}
	return strings.Trim(slug[:maxSlugLen], "-")
}

// slugQuerier is the subset of *sql.Tx / *sql.DB the slug helpers need, so
// they can run inside fulfillPass's transaction (the pool is capped at one
// connection — a query on s.DB inside a tx would deadlock).
type slugQuerier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// ─── public: redemption ───

type redeemInput struct {
	Code    string `json:"code"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Title   string `json:"title"`
	Author  string `json:"author"`
	Company string `json:"company"` // honeypot — must be empty
}

// handlePublicRedeem is the public "I have a code" form: POST /api/public/redeem.
// Same anti-abuse shape as the workshop registration form (hidden honeypot
// field + per-IP limiter), no captcha.
func (s *Server) handlePublicRedeem(w http.ResponseWriter, r *http.Request) {
	ip := clientIP(r)
	if !s.limiter().allow(ip, 5, 10*time.Minute) {
		jsonErr(w, "Too many attempts. Please try again in a few minutes.", http.StatusTooManyRequests)
		return
	}

	var in redeemInput
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64*1024)).Decode(&in); err != nil {
		jsonErr(w, "Sorry, we couldn't read that submission.", http.StatusBadRequest)
		return
	}

	// Honeypot: real users never fill this hidden field. Pretend success so
	// bots don't learn they were caught, and do nothing at all.
	if strings.TrimSpace(in.Company) != "" {
		slog.Info("redeem honeypot tripped", "ip", ip)
		jsonOK(w, map[string]any{"ok": true})
		return
	}

	res, err := s.fulfillPass(r.Context(), "coupon", fulfillPassInput{
		Code:   in.Code,
		Name:   in.Name,
		Email:  in.Email,
		Title:  in.Title,
		Author: in.Author,
	})
	if err != nil {
		var inputErr passInputError
		if errors.As(err, &inputErr) {
			jsonErr(w, inputErr.msg, http.StatusBadRequest)
			return
		}
		slog.Error("pass redemption failed", "err", err, "ip", ip)
		jsonErr(w, "Something went wrong redeeming that code. Please email us instead.", http.StatusInternalServerError)
		return
	}

	s.sendPassFulfillmentEmail(*res, "public")

	jsonOK(w, map[string]any{
		"ok":           true,
		"portal_url":   res.PortalURL,
		"client_slug":  res.ClientSlug,
		"project_slug": res.ProjectSlug,
	})
}

// ─── customer: pass status + book list ───

type passStatusResponse struct {
	Exists           bool   `json:"exists"`
	Status           string `json:"status,omitempty"`
	Live             bool   `json:"live,omitempty"`
	BuildsIncluded   int64  `json:"builds_included,omitempty"`
	BuildsExtra      int64  `json:"builds_extra"`
	BuildsUsed       int64  `json:"builds_used"`
	CreditsRemaining int64  `json:"credits_remaining"`
	ExpiresAt        string `json:"expires_at,omitempty"`
	CustomerName     string `json:"customer_name,omitempty"`
	IndexIncluded    bool   `json:"index_included"`        // back-of-book index add-on
	FinalsGate       string `json:"finals_gate,omitempty"` // "off" while the studio waives the no-finals refusal
}

// handleGetProjectPass: GET /api/projects/{id}/pass — the credits/expiry badge
// for the customer page. Projects with no pass return {"exists":false}.
func (s *Server) handleGetProjectPass(w http.ResponseWriter, r *http.Request) {
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", http.StatusBadRequest)
		return
	}
	if !s.requireAuth(w, r, pid) {
		return
	}
	pass := s.passForProject(r.Context(), pid)
	if pass == nil {
		jsonOK(w, passStatusResponse{Exists: false})
		return
	}
	jsonOK(w, passStatusResponse{
		Exists:           true,
		Status:           pass.Status,
		Live:             passLive(*pass),
		BuildsIncluded:   pass.BuildsIncluded,
		BuildsExtra:      pass.BuildsExtra,
		BuildsUsed:       pass.BuildsUsed,
		CreditsRemaining: passCreditsRemaining(*pass),
		ExpiresAt:        pass.ExpiresAt.UTC().Format(time.RFC3339),
		CustomerName:     pass.CustomerName,
		IndexIncluded:    passIndexIncluded(pass),
		FinalsGate:       map[bool]string{true: "off", false: ""}[finalsGateOff()],
	})
}

// handleListProjectBooks: GET /api/projects/{id}/books — that project's books
// only (the global GET /api/books stays admin-only).
func (s *Server) handleListProjectBooks(w http.ResponseWriter, r *http.Request) {
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", http.StatusBadRequest)
		return
	}
	if !s.requireAuth(w, r, pid) {
		return
	}
	q := dbgen.New(s.DB)
	books, err := q.GetBooksByProject(r.Context(), sql.NullInt64{Int64: pid, Valid: true})
	if err != nil {
		jsonErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if books == nil {
		books = []dbgen.GetBooksByProjectRow{}
	}
	jsonOK(w, books)
}

// ─── admin: coupons ───

type couponInput struct {
	RegistrationID int64  `json:"registration_id"`
	IssuedToEmail  string `json:"issued_to_email"`
	Note           string `json:"note"`
	ExpiresAt      string `json:"expires_at"` // YYYY-MM-DD; default couponDefaultExpiry
}

// handleAdminCreateCoupon issues one free-access code, optionally bound to a
// cohort registration (the tracker's "issue code" action).
func (s *Server) handleAdminCreateCoupon(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	var in couponInput
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 32*1024)).Decode(&in); err != nil {
		jsonErr(w, "invalid coupon request", http.StatusBadRequest)
		return
	}
	in.IssuedToEmail = clip(strings.ToLower(in.IssuedToEmail), 320)
	in.Note = clip(in.Note, 500)

	expiry := strings.TrimSpace(in.ExpiresAt)
	if expiry == "" {
		expiry = couponDefaultExpiry
	}
	expiresAt, err := time.Parse("2006-01-02", expiry)
	if err != nil {
		jsonErr(w, "expires_at must be YYYY-MM-DD", http.StatusBadRequest)
		return
	}
	// End of the given day, so a code dated Sep 30 works all of Sep 30.
	expiresAt = expiresAt.Add(24*time.Hour - time.Second)

	// Bind to a registration when asked, and borrow its email if none given.
	var registrationID sql.NullInt64
	if in.RegistrationID > 0 {
		var email string
		err := s.DB.QueryRowContext(r.Context(),
			`SELECT email FROM event_registrations WHERE id = ?`, in.RegistrationID).Scan(&email)
		if errors.Is(err, sql.ErrNoRows) {
			jsonErr(w, "registration not found", http.StatusNotFound)
			return
		}
		if err != nil {
			jsonErr(w, "could not load registration", http.StatusInternalServerError)
			return
		}
		registrationID = sql.NullInt64{Int64: in.RegistrationID, Valid: true}
		if in.IssuedToEmail == "" {
			in.IssuedToEmail = email
		}
	}

	q := dbgen.New(s.DB)
	// Retry on the (vanishingly unlikely) code collision.
	for attempt := 0; attempt < 5; attempt++ {
		code, err := generateCouponCode()
		if err != nil {
			jsonErr(w, "could not generate a code", http.StatusInternalServerError)
			return
		}
		coupon, err := q.CreateCoupon(r.Context(), dbgen.CreateCouponParams{
			Code:           code,
			Sku:            passSKU,
			MaxRedemptions: 1,
			ExpiresAt:      sql.NullTime{Time: expiresAt, Valid: true},
			RegistrationID: registrationID,
			IssuedToEmail:  in.IssuedToEmail,
			Note:           in.Note,
		})
		if err != nil {
			if isUniqueErr(err) {
				continue
			}
			jsonErr(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		jsonOK(w, map[string]any{
			"id":              coupon.ID,
			"code":            coupon.Code,
			"expires_at":      coupon.ExpiresAt.Time.UTC().Format(time.RFC3339),
			"issued_to_email": coupon.IssuedToEmail,
			"registration_id": coupon.RegistrationID.Int64,
		})
		return
	}
	jsonErr(w, "could not generate a unique code", http.StatusInternalServerError)
}

type couponRow struct {
	ID             int64  `json:"id"`
	Code           string `json:"code"`
	Sku            string `json:"sku"`
	MaxRedemptions int64  `json:"max_redemptions"`
	RedeemedCount  int64  `json:"redeemed_count"`
	Redeemed       bool   `json:"redeemed"`
	RedeemedAt     string `json:"redeemed_at"`
	ExpiresAt      string `json:"expires_at"`
	RegistrationID int64  `json:"registration_id"`
	IssuedToEmail  string `json:"issued_to_email"`
	Note           string `json:"note"`
	CreatedAt      string `json:"created_at"`
	PassID         int64  `json:"pass_id"`
	ProjectName    string `json:"project_name"`
	ProjectPath    string `json:"project_path"`
}

func (s *Server) handleAdminListCoupons(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	q := dbgen.New(s.DB)
	rows, err := q.ListCoupons(r.Context())
	if err != nil {
		jsonErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	out := make([]couponRow, 0, len(rows))
	for _, c := range rows {
		item := couponRow{
			ID:             c.ID,
			Code:           c.Code,
			Sku:            c.Sku,
			MaxRedemptions: c.MaxRedemptions,
			RedeemedCount:  c.RedeemedCount,
			Redeemed:       c.RedeemedCount >= c.MaxRedemptions,
			RegistrationID: c.RegistrationID.Int64,
			IssuedToEmail:  c.IssuedToEmail,
			Note:           c.Note,
			CreatedAt:      c.CreatedAt.UTC().Format(time.RFC3339),
			PassID:         c.PassID.Int64,
			ProjectName:    c.ProjectName,
		}
		if c.ExpiresAt.Valid {
			item.ExpiresAt = c.ExpiresAt.Time.UTC().Format(time.RFC3339)
		}
		if c.RedeemedAt.Valid {
			item.RedeemedAt = c.RedeemedAt.Time.UTC().Format(time.RFC3339)
		}
		if c.ClientSlug != "" && c.ProjectSlug != "" {
			item.ProjectPath = "/" + c.ClientSlug + "/" + c.ProjectSlug + "/factory/"
		}
		out = append(out, item)
	}
	jsonOK(w, out)
}

// ─── admin: passes ───

type passRow struct {
	ID               int64  `json:"id"`
	ProjectID        int64  `json:"project_id"`
	ProjectName      string `json:"project_name"`
	ProjectPath      string `json:"project_path"`
	ClientSlug       string `json:"client_slug"`
	Sku              string `json:"sku"`
	Source           string `json:"source"`
	CouponCode       string `json:"coupon_code"`
	CustomerName     string `json:"customer_name"`
	CustomerEmail    string `json:"customer_email"`
	BuildsIncluded   int64  `json:"builds_included"`
	BuildsExtra      int64  `json:"builds_extra"`
	BuildsUsed       int64  `json:"builds_used"`
	CreditsRemaining int64  `json:"credits_remaining"`
	FulfilledAt      string `json:"fulfilled_at"`
	ExpiresAt        string `json:"expires_at"`
	Status           string `json:"status"`
	Live             bool   `json:"live"`
	Note             string `json:"note"`
	StripeSessionID  string `json:"stripe_session_id"`
	AmountPaid       int64  `json:"amount_paid"` // cents, after discount; 0 for coupon/admin passes
	PromoCode        string `json:"promo_code"`
	IndexIncluded    bool   `json:"index_included"` // back-of-book index add-on
	HasToken         bool   `json:"has_token"`      // project has an API token (factory.py / bearer)
}

func (s *Server) handleAdminListPasses(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	q := dbgen.New(s.DB)
	rows, err := q.ListPasses(r.Context())
	if err != nil {
		jsonErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tokened := s.projectsWithTokens(r.Context())
	out := make([]passRow, 0, len(rows))
	for _, p := range rows {
		// Reuse the credits/live helpers by rehydrating the fields they read.
		as := dbgen.Pass{
			BuildsIncluded: p.BuildsIncluded,
			BuildsExtra:    p.BuildsExtra,
			BuildsUsed:     p.BuildsUsed,
			ExpiresAt:      p.ExpiresAt,
			Status:         p.Status,
		}
		out = append(out, passRow{
			HasToken:         tokened[p.ProjectID],
			ID:               p.ID,
			ProjectID:        p.ProjectID,
			ProjectName:      p.ProjectName,
			ProjectPath:      "/" + p.ClientSlug + "/" + p.ProjectSlug + "/factory/",
			ClientSlug:       p.ClientSlug,
			Sku:              p.Sku,
			Source:           p.Source,
			CouponCode:       p.CouponCode,
			CustomerName:     p.CustomerName,
			CustomerEmail:    p.CustomerEmail,
			BuildsIncluded:   p.BuildsIncluded,
			BuildsExtra:      p.BuildsExtra,
			BuildsUsed:       p.BuildsUsed,
			CreditsRemaining: passCreditsRemaining(as),
			FulfilledAt:      p.FulfilledAt.UTC().Format(time.RFC3339),
			ExpiresAt:        p.ExpiresAt.UTC().Format(time.RFC3339),
			Status:           p.Status,
			Live:             passLive(as),
			Note:             p.Note,
			StripeSessionID:  p.StripeSessionID,
			AmountPaid:       p.AmountPaid,
			PromoCode:        p.PromoCode,
			IndexIncluded:    p.IndexIncluded != 0,
		})
	}
	jsonOK(w, out)
}

// handleAdminCreatePass hand-grants a pass (source = "admin") without a
// coupon — the concierge path for an attendee who lost their code or a
// customer who paid outside Stripe. Same fulfillment function as redemption.
func (s *Server) handleAdminCreatePass(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	var in struct {
		Name      string `json:"name"`
		Email     string `json:"email"`
		Title     string `json:"title"`
		Author    string `json:"author"`
		Note      string `json:"note"`
		SendEmail *bool  `json:"send_email"`
		ProjectID int64  `json:"project_id"` // attach to an existing project (second book for an existing client)
		APIToken  bool   `json:"api_token"`  // also mint a project token (factory.py / bearer); returned once as "token"
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 32*1024)).Decode(&in); err != nil {
		jsonErr(w, "invalid pass request", http.StatusBadRequest)
		return
	}
	if in.ProjectID > 0 {
		s.attachPassToProject(w, r, in.ProjectID, in.Email, in.Name, in.Note, in.APIToken)
		return
	}
	res, err := s.fulfillPass(r.Context(), "admin", fulfillPassInput{
		Name: in.Name, Email: in.Email, Title: in.Title, Author: in.Author, Note: in.Note,
	})
	if err != nil {
		var inputErr passInputError
		if errors.As(err, &inputErr) {
			jsonErr(w, inputErr.msg, http.StatusBadRequest)
			return
		}
		jsonErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if in.SendEmail == nil || *in.SendEmail {
		s.sendPassFulfillmentEmail(*res, triggeredBy(r, "admin"))
	}
	token := ""
	if in.APIToken {
		if token, err = s.mintProjectToken(r.Context(), res.Pass.ProjectID, triggeredBy(r, "admin")); err != nil {
			slog.Error("mint token after pass", "project", res.Pass.ProjectID, "err", err)
		}
	}
	w.WriteHeader(http.StatusCreated)
	jsonOK(w, map[string]any{
		"ok":           true,
		"pass_id":      res.Pass.ID,
		"project_id":   res.Pass.ProjectID,
		"portal_url":   res.PortalURL,
		"client_slug":  res.ClientSlug,
		"project_slug": res.ProjectSlug,
		"password":     res.Password,
		"token":        token,
	})
}

// attachPassToProject is the second-book path: a client who already holds a
// pass created another project from the portal (POST /api/clients/{c}/projects),
// which has no pass and so is read-only in the factory. The admin attaches a
// fresh pass here; customer name/email default to the client's existing pass
// (or the client row) so the build-ready / template-ready emails keep flowing.
// No new client, no new password, no fulfillment email.
func (s *Server) attachPassToProject(w http.ResponseWriter, r *http.Request, projectID int64, email, name, note string, apiToken bool) {
	ctx := r.Context()
	q := dbgen.New(s.DB)
	project, err := q.GetProject(ctx, projectID)
	if errors.Is(err, sql.ErrNoRows) {
		jsonErr(w, "project not found", http.StatusNotFound)
		return
	}
	if err != nil {
		jsonErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := q.GetPassByProject(ctx, projectID); err == nil {
		jsonErr(w, "this project already has a factory pass", http.StatusConflict)
		return
	}
	email = clip(strings.ToLower(strings.TrimSpace(email)), 320)
	name = clip(strings.TrimSpace(name), 200)
	if email == "" || name == "" {
		// Sibling pass on the same client is the best source of truth.
		var sibEmail, sibName string
		_ = s.DB.QueryRowContext(ctx, `SELECT p.customer_email, p.customer_name FROM passes p
			JOIN projects pr ON pr.id = p.project_id
			WHERE pr.client_slug = ? ORDER BY p.id DESC LIMIT 1`, project.ClientSlug).Scan(&sibEmail, &sibName)
		if email == "" {
			email = sibEmail
		}
		if name == "" {
			name = sibName
		}
		if name == "" {
			_ = s.DB.QueryRowContext(ctx, `SELECT name FROM clients WHERE slug = ?`, project.ClientSlug).Scan(&name)
		}
	}
	if email != "" && !looksLikeEmail(email) {
		jsonErr(w, "That email address doesn't look right.", http.StatusBadRequest)
		return
	}
	pass, err := q.CreatePass(ctx, dbgen.CreatePassParams{
		ProjectID:      projectID,
		Sku:            passSKU,
		Source:         "admin",
		CustomerEmail:  email,
		CustomerName:   name,
		BuildsIncluded: passBuildsIncluded,
		Note:           clip(note, 500),
	})
	if err != nil {
		jsonErr(w, "create pass: "+err.Error(), http.StatusInternalServerError)
		return
	}
	slog.Info("factory pass attached to existing project", "pass_id", pass.ID,
		"project", project.ClientSlug+"/"+project.ProjectSlug, "email", email, "by", triggeredBy(r, "admin"))
	token := ""
	if apiToken {
		if token, err = s.mintProjectToken(r.Context(), projectID, triggeredBy(r, "admin")); err != nil {
			slog.Error("mint token after pass", "project", projectID, "err", err)
		}
	}
	w.WriteHeader(http.StatusCreated)
	jsonOK(w, map[string]any{
		"ok":           true,
		"pass_id":      pass.ID,
		"project_id":   projectID,
		"portal_url":   s.portalURL(project.ClientSlug, project.ProjectSlug),
		"client_slug":  project.ClientSlug,
		"project_slug": project.ProjectSlug,
		"attached":     true,
		"token":        token,
	})
}

// ─── API access: project tokens minted by the studio ───
//
// A project token is what factory.py / a bearer caller sends. It is stored
// hashed like any password, so it can be shown exactly once — at mint time —
// and afterwards only rolled. Minting replaces every existing token on the
// project (there is one customer per project; "roll" == "mint").

func (s *Server) projectsWithTokens(ctx context.Context) map[int64]bool {
	m := map[int64]bool{}
	rows, err := s.DB.QueryContext(ctx, `SELECT DISTINCT project_id FROM auth_tokens`)
	if err != nil {
		return m
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		if rows.Scan(&id) == nil {
			m[id] = true
		}
	}
	return m
}

func (s *Server) mintProjectToken(ctx context.Context, projectID int64, by string) (string, error) {
	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := "fk_" + base64.RawURLEncoding.EncodeToString(raw)
	hash, err := hashPassword(token)
	if err != nil {
		return "", err
	}
	q := dbgen.New(s.DB)
	if err := q.DeleteAuthTokensByProject(ctx, projectID); err != nil {
		return "", err
	}
	if err := q.CreateAuthToken(ctx, dbgen.CreateAuthTokenParams{ProjectID: projectID, TokenHash: hash, Label: "api"}); err != nil {
		return "", err
	}
	slog.Info("project token minted", "project", projectID, "by", by)
	return token, nil
}

// handleAdminMintToken — POST /api/admin/projects/{id}/token. Mints (or rolls)
// the project's API token and returns it once, with the paste-ready lines the
// studio hands the customer.
func (s *Server) handleAdminMintToken(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", http.StatusBadRequest)
		return
	}
	q := dbgen.New(s.DB)
	project, err := q.GetProject(r.Context(), pid)
	if err != nil {
		jsonErr(w, "project not found", http.StatusNotFound)
		return
	}
	token, err := s.mintProjectToken(r.Context(), pid, triggeredBy(r, "admin"))
	if err != nil {
		jsonErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]any{
		"ok":           true,
		"project_id":   pid,
		"token":        token,
		"client_slug":  project.ClientSlug,
		"project_slug": project.ProjectSlug,
		"factory_url":  s.portalURL(project.ClientSlug, project.ProjectSlug),
	})
}

// handleAdminResetClientPassword rotates a Factory Pass client's password and
// re-sends the fulfillment-style email. The plaintext password is returned
// once to the authenticated admin so a customer can still be recovered when
// mail is unconfigured or the provider rejects the send.
func (s *Server) handleAdminResetClientPassword(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	clientSlug := strings.TrimSpace(r.PathValue("slug"))
	if clientSlug == "" {
		jsonErr(w, "client slug required", http.StatusBadRequest)
		return
	}

	var res fulfillPassResult
	res.ClientSlug = clientSlug
	err := s.DB.QueryRowContext(r.Context(), `
		SELECT p.id, p.project_id, p.customer_email, p.customer_name,
		       p.builds_included, p.expires_at, pr.name, pr.project_slug
		FROM clients c
		JOIN projects pr ON pr.client_slug = c.slug
		JOIN passes p ON p.project_id = pr.id
		WHERE c.slug = ?
		ORDER BY p.fulfilled_at DESC, p.id DESC
		LIMIT 1
	`, clientSlug).Scan(
		&res.Pass.ID, &res.Pass.ProjectID, &res.Pass.CustomerEmail,
		&res.Pass.CustomerName, &res.Pass.BuildsIncluded, &res.Pass.ExpiresAt,
		&res.Title, &res.ProjectSlug,
	)
	if errors.Is(err, sql.ErrNoRows) {
		jsonErr(w, "factory pass client not found", http.StatusNotFound)
		return
	}
	if err != nil {
		slog.Error("factory pass password reset lookup failed", "client", clientSlug, "err", err)
		jsonErr(w, "could not load factory pass client", http.StatusInternalServerError)
		return
	}

	password, err := generateClientPassword()
	if err != nil {
		jsonErr(w, "could not generate a password", http.StatusInternalServerError)
		return
	}
	passwordHash, err := hashPassword(password)
	if err != nil {
		jsonErr(w, "could not secure the password", http.StatusInternalServerError)
		return
	}
	result, err := s.DB.ExecContext(r.Context(),
		`UPDATE clients SET password_hash = ? WHERE slug = ?`, passwordHash, clientSlug)
	if err != nil {
		slog.Error("factory pass password reset update failed", "client", clientSlug, "err", err)
		jsonErr(w, "could not reset the password", http.StatusInternalServerError)
		return
	}
	if n, _ := result.RowsAffected(); n != 1 {
		jsonErr(w, "factory pass client not found", http.StatusNotFound)
		return
	}
	// A reset is usually defensive: burn any unredeemed sign-in links too,
	// or a link obtained before the reset would mint a fresh cookie for the
	// new password (2026-09-24 review).
	if _, err := s.DB.ExecContext(r.Context(),
		`UPDATE login_links SET used_at = ? WHERE client_slug = ? AND used_at IS NULL`,
		time.Now().UTC().Format(loginLinkTimeLayout), clientSlug); err != nil {
		slog.Error("password reset: invalidate login links", "client", clientSlug, "err", err)
	}

	res.Password = password
	res.PortalURL = s.portalURL(res.ClientSlug, res.ProjectSlug)
	emailSent := false
	emailMessage := "Email is not configured; copy the password now."
	if s.Email != nil {
		if err := s.deliverPassFulfillmentEmail(res, mailMeta{Kind: mailKindClientPassword, TriggeredBy: triggeredBy(r, "admin")}); err != nil {
			slog.Error("factory pass password reset email failed", "client", clientSlug,
				"pass_id", res.Pass.ID, "err", err)
			emailMessage = "The password was reset, but the email could not be sent; copy the password now."
		} else {
			emailSent = true
			emailMessage = "The new password was emailed to the Factory Pass customer."
		}
	}

	slog.Info("factory pass client password reset", "client", clientSlug,
		"pass_id", res.Pass.ID, "email_sent", emailSent)
	jsonOK(w, map[string]any{
		"ok":            true,
		"client_slug":   clientSlug,
		"portal_url":    res.PortalURL,
		"password":      password,
		"email_sent":    emailSent,
		"email_message": emailMessage,
	})
}

var validGrantReasons = map[string]bool{"pack": true, "grant": true}

// handleAdminGrantPassBuilds adds build credits to a pass (a bought pack, or a
// goodwill grant) and records the movement in the ledger.
func (s *Server) handleAdminGrantPassBuilds(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		jsonErr(w, "invalid pass id", http.StatusBadRequest)
		return
	}
	var in struct {
		Builds int64  `json:"builds"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8*1024)).Decode(&in); err != nil {
		jsonErr(w, "invalid grant", http.StatusBadRequest)
		return
	}
	if in.Builds <= 0 || in.Builds > 50 {
		jsonErr(w, "builds must be between 1 and 50", http.StatusBadRequest)
		return
	}
	if in.Reason == "" {
		in.Reason = "pack"
	}
	if !validGrantReasons[in.Reason] {
		jsonErr(w, "reason must be pack or grant", http.StatusBadRequest)
		return
	}

	q := dbgen.New(s.DB)
	pass, err := q.GetPass(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		jsonErr(w, "pass not found", http.StatusNotFound)
		return
	}
	if err != nil {
		jsonErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := q.AddPassBuildsExtra(r.Context(), dbgen.AddPassBuildsExtraParams{
		BuildsExtra: in.Builds, ID: pass.ID,
	}); err != nil {
		jsonErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := q.CreatePassLedgerEntry(r.Context(), dbgen.CreatePassLedgerEntryParams{
		PassID: pass.ID,
		Delta:  in.Builds,
		Reason: in.Reason,
	}); err != nil {
		jsonErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	updated, err := q.GetPass(r.Context(), id)
	if err != nil {
		jsonErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]any{
		"ok":                true,
		"builds_extra":      updated.BuildsExtra,
		"credits_remaining": passCreditsRemaining(updated),
	})
}

// ─── emails (EMAIL_SYSTEM.md pathway #6) ───

// workshopEndsAt is the switchover instant for the Sep 21–22 workshop copy:
// end of day Tuesday in Hong Kong, the same reading WORKSHOP49 uses
// (store.go). After this instant the workshop sentences drop out of the pass
// emails, the registration auto-reply, and the factory page automatically —
// nothing to hand-edit on Tuesday night.
var workshopEndsAt = time.Date(2026, 9, 22, 23, 59, 59, 0, hkt)

func workshopLive() bool { return time.Now().Before(workshopEndsAt) }

// passSupportEdges is the shared copy for what is and isn't included. The
// page (factory.html, gated the same way in factory.js) and both emails say
// the same thing on purpose. workshop=true is the Sep 21–22 window.
func passSupportEdges(workshop bool) []string {
	if workshop {
		return []string{
			"Live help during the workshop sessions (Sep 21–22) is included.",
			"Outside that, email support is not included — the preflight report and the docs are the self-serve path.",
			"Live help is available at USD 100/hr, booked in advance, one-hour minimum.",
		}
	}
	return []string{
		"Email support is not included — the preflight report and the docs are the self-serve path.",
		"Live help is available at USD 100/hr, booked in advance, one-hour minimum.",
	}
}

// deliverPassFulfillmentEmail performs the actual send (Resend, or AgentMail fallback). Redemption
// calls it in a goroutine; the admin password-reset path calls it synchronously
// so the tracker can report whether the replacement password was really sent.
func (s *Server) deliverPassFulfillmentEmail(res fulfillPassResult, meta mailMeta) error {
	if s.Email == nil {
		return errors.New("email not configured")
	}
	if meta.Kind == "" {
		meta.Kind = mailKindFactoryPass
	}
	meta.RefType, meta.RefID = "pass", mailRef(res.Pass.ID)
	subject := fmt.Sprintf("Your Factory Pass: %s", res.Title)
	workshop := workshopLive()
	return s.mail(meta, []string{res.Pass.CustomerEmail}, nil, subject,
		passFulfillmentText(res, workshop), passFulfillmentHTML(res, workshop))
}

// sendPassFulfillmentEmail mails the customer their portal URL, client
// password, and what the pass includes. Fire-and-forget: mail must never fail
// a fulfillment, and s.Email is nil in local/dev/test runs.
func (s *Server) sendPassFulfillmentEmail(res fulfillPassResult, by string) {
	if s.Email == nil {
		slog.Warn("factory pass fulfilled but email not configured",
			"pass_id", res.Pass.ID, "email", res.Pass.CustomerEmail)
		return
	}
	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("pass fulfillment email panic", "recover", rec)
			}
		}()
		if err := s.deliverPassFulfillmentEmail(res, mailMeta{Kind: mailKindFactoryPass, TriggeredBy: by}); err != nil {
			slog.Error("pass fulfillment email failed", "err", err, "pass_id", res.Pass.ID)
		}
	}()
}

// passIncludedSummary is what a pass actually contains, add-ons folded in,
// so the fulfilment email never quotes the base pass when the customer
// bought more (caught 2026-09-17: "3 builds … (6 months)" on a 6-build,
// 12-month pass).
type passIncludedSummary struct {
	Builds int64
	Months int
	AddOns []string
}

func passIncluded(p dbgen.Pass) passIncludedSummary {
	inc := passIncludedSummary{Builds: p.BuildsIncluded + p.BuildsExtra}
	// Whole months between fulfilment and expiry, rounded to the nearest.
	days := p.ExpiresAt.Sub(p.FulfilledAt).Hours() / 24
	inc.Months = int(days/30.44 + 0.5)
	if inc.Months < 1 {
		inc.Months = passStorageMonths
	}
	if p.BuildsExtra > 0 {
		inc.AddOns = append(inc.AddOns, fmt.Sprintf("+%d builds", p.BuildsExtra))
	}
	if extra := inc.Months - passStorageMonths; extra > 0 {
		inc.AddOns = append(inc.AddOns, fmt.Sprintf("+%d months storage", extra))
	}
	return inc
}

func passFulfillmentText(res fulfillPassResult, workshop bool) string {
	expires := res.Pass.ExpiresAt.UTC().Format("2 January 2006")
	var b strings.Builder
	fmt.Fprintf(&b, "Hi %s,\n\n", firstName(res.Pass.CustomerName))
	if workshop {
		fmt.Fprintf(&b, "Your Factory Pass is live: one manuscript, all the way through the protocol. If you redeemed a workshop code, this is the account you'll use in the sessions; Discord and calendar invites arrive separately.\n\n")
	} else {
		fmt.Fprintf(&b, "Your Factory Pass is live: one manuscript, all the way through the protocol.\n\n")
	}
	fmt.Fprintf(&b, "Manuscript:     %s\n", res.Title)
	fmt.Fprintf(&b, "Your factory:   %s\n", res.PortalURL)
	fmt.Fprintf(&b, "Sign-in name:   %s\n\n", res.ClientSlug)
	fmt.Fprintf(&b, "Easiest way in: open your portal and enter this email address (%s) — we'll send you a sign-in link. No password needed.\n\n", res.Pass.CustomerEmail)
	fmt.Fprintf(&b, "Password (if you'd rather): %s\n\n", res.Password)
	inc := passIncluded(res.Pass)
	fmt.Fprintf(&b, "What's included\n")
	fmt.Fprintf(&b, "  - %d builds: each one makes the EPUB and the print PDF together\n", inc.Builds)
	fmt.Fprintf(&b, "  - Unlimited preflights: the report tells you what to fix\n")
	fmt.Fprintf(&b, "  - Your project stays live and rebuildable until %s (%d months)\n", expires, inc.Months)
	for _, a := range inc.AddOns {
		fmt.Fprintf(&b, "  - Add-on: %s\n", a)
	}
	b.WriteString("\n")
	fmt.Fprintf(&b, "First steps\n")
	step1, step2 := "", ""
	if workshop {
		step1 = " Workshop attendees fill it live in session 1 (Mon Sep 21)."
		step2 = " Workshop attendees: be ready to do this in session 2 (Mon Sep 21)."
	}
	fmt.Fprintf(&b, "  1. Fill the transmittal: it is the spec your book is built from. Mark it final; if you are starting from a blank page, download the authoring template generated from it.%s\n", step1)
	fmt.Fprintf(&b, "  2. Upload your manuscript as .docx — from any editor, using Heading 1/2 and [[quote]]-style markers, or written in that template.%s\n", step2)
	fmt.Fprintf(&b, "  3. Run a preflight (free, as often as you like) and fix what it flags.\n")
	fmt.Fprintf(&b, "  4. Build. EPUBs are free and unlimited; print PDFs count. Failed builds don't count against your %d.\n\n", inc.Builds)
	fmt.Fprintf(&b, "Support\n")
	for _, edge := range passSupportEdges(workshop) {
		fmt.Fprintf(&b, "  - %s\n", edge)
	}
	fmt.Fprintf(&b, "\nSee you in the factory,\nJenna Dixon · [jdbb] studio\n")
	return b.String()
}

func passFulfillmentHTML(res fulfillPassResult, workshop bool) string {
	expires := res.Pass.ExpiresAt.UTC().Format("2 January 2006")
	var edges []string
	for _, edge := range passSupportEdges(workshop) {
		edges = append(edges, html.EscapeString(edge))
	}
	var b strings.Builder
	b.WriteString(emailP(fmt.Sprintf("Hi %s,", html.EscapeString(firstName(res.Pass.CustomerName)))))
	if workshop {
		b.WriteString(emailP("Your <b>Factory Pass</b> is live &mdash; one manuscript, all the way through the protocol. If you redeemed a workshop code, this is the account you&rsquo;ll use in the sessions; Discord and calendar invites arrive separately."))
	} else {
		b.WriteString(emailP("Your <b>Factory Pass</b> is live &mdash; one manuscript, all the way through the protocol."))
	}
	b.WriteString(emailKV([][2]string{
		{"Manuscript", "<b>" + html.EscapeString(res.Title) + "</b>"},
		{"Your factory", fmt.Sprintf(`<a href="%s" style="color:%s;text-decoration:underline">%s</a>`, html.EscapeString(res.PortalURL), emailAccent, emailCode(res.PortalURL))},
		{"Sign-in name", emailCode(res.ClientSlug)},
	}))
	b.WriteString(emailP(fmt.Sprintf("<b>Easiest way in:</b> open your portal and enter this email address (%s) &mdash; we&rsquo;ll send you a sign-in link. No password needed.", emailCode(res.Pass.CustomerEmail))))
	b.WriteString(emailKV([][2]string{
		{"Password (if you'd rather)", emailCode(res.Password)},
	}))
	inc := passIncluded(res.Pass)
	included := []string{
		fmt.Sprintf("%d builds &mdash; each one makes the EPUB and the print PDF together", inc.Builds),
		"Unlimited preflights &mdash; the report tells you what to fix",
		fmt.Sprintf("Your project stays live and rebuildable until <b>%s</b> (%d months)", html.EscapeString(expires), inc.Months),
	}
	for _, a := range inc.AddOns {
		included = append(included, "Add-on: "+html.EscapeString(a))
	}
	b.WriteString(emailH2("What's included"))
	b.WriteString(emailList(included, false))
	step1, step2 := "", ""
	if workshop {
		step1 = " Workshop attendees fill it live in session 1 (Mon Sep 21)."
		step2 = " Workshop attendees: be ready to do this in session 2 (Mon Sep 21)."
	}
	b.WriteString(emailH2("First steps"))
	b.WriteString(emailList([]string{
		"Fill the <b>transmittal</b> &mdash; it is the spec your book is built from. Mark it final; if you are starting from a blank page, download the authoring template generated from it." + step1,
		"Upload your manuscript as .docx — from any editor, using Heading 1/2 and [[quote]]-style markers, or written in that template." + step2,
		"Run a <b>preflight</b> (free, as often as you like) and fix what it flags.",
		fmt.Sprintf("<b>Build.</b> EPUBs are free and unlimited; print PDFs count. Failed builds don&rsquo;t count against your %d.", inc.Builds),
	}, true))
	// Support edges: same sentences as the page and the text part, in small type.
	b.WriteString(emailH2("Support"))
	fmt.Fprintf(&b, `<ul style="margin:0 0 14px;padding-left:22px;font-size:13px;color:%s">`, emailSecondary)
	for _, e := range edges {
		fmt.Fprintf(&b, `<li style="margin:0 0 5px">%s</li>`, e)
	}
	b.WriteString(`</ul>`)
	b.WriteString(emailP("See you in the factory,"))
	b.WriteString(emailSignoff())
	return emailShell(b.String(), emailShellOpts{Kicker: "Factory Pass", Title: res.Title})
}

// sendBuildDeliveredEmail is the receipt for a successful build: links to the
// PDF, the EPUB, and the preflight report, plus credits remaining. Called from
// runConversion once the artifact is stored. Runs on the conversion goroutine,
// which has already outlived the request.
//
// format is "pdf", "epub" or "both" — the receipt names what was made.
//
// Only finals get a receipt (punch list 0.28). Proofs are free and unlimited,
// the factory page is live while one runs, and a mail per proof would be
// noise — so a proof build sends nothing and is only logged as a factory
// event. The final's links carry ?kind=final so they can never resolve to a
// newer proof.
func (s *Server) sendBuildDeliveredEmail(pass dbgen.Pass, book dbgen.Book, format string) {
	if pass.CustomerEmail == "" || book.BuildKind == buildKindProof {
		return
	}
	if s.Email == nil {
		slog.Warn("build delivered but email not configured", "pass_id", pass.ID, "book_id", book.ID)
		return
	}
	defer func() {
		if rec := recover(); rec != nil {
			slog.Error("build delivered email panic", "recover", rec)
		}
	}()

	base := strings.TrimRight(s.BaseURL, "/")
	pdfURL := fmt.Sprintf("%s/api/books/%d/download/pdf?kind=final", base, book.ID)
	epubURL := fmt.Sprintf("%s/api/books/%d/download/epub?kind=final", base, book.ID)
	reportURL := fmt.Sprintf("%s/api/projects/%d/preflight/report?book_id=%d", base, pass.ProjectID, book.ID)
	credits := passCreditsRemaining(pass)
	total := pass.BuildsIncluded + pass.BuildsExtra

	subject := fmt.Sprintf("%s: %s", buildDeliveredNoun(format), book.Title)
	textBody := buildDeliveredText(pass, book, format, pdfURL, epubURL, reportURL, credits, total)
	htmlBody := buildDeliveredHTML(pass, book, format, pdfURL, epubURL, reportURL, credits, total)

	if err := s.mail(mailMeta{Kind: mailKindBuildDelivered, RefType: "pass", RefID: mailRef(pass.ID), TriggeredBy: "system"}, []string{pass.CustomerEmail}, nil, subject, textBody, htmlBody); err != nil {
		slog.Error("build delivered email failed", "err", err, "pass_id", pass.ID, "book_id", book.ID)
	}
}

// sendTemplateReadyEmail tells the customer their Word template exists, the
// moment they mark the transmittal final (EMAIL_SYSTEM.md pathway #7). Fired
// once per draft→final transition from handleUpdateTransmittal; the download
// itself is the self-serve GET /api/projects/{id}/word-template, which needs
// the client sign-in, so the mail points at the factory page where the button
// lives and includes the direct link for a signed-in browser.
func (s *Server) sendTemplateReadyEmail(pass dbgen.Pass, title string) {
	if pass.CustomerEmail == "" {
		return
	}
	if s.Email == nil {
		slog.Warn("template ready but email not configured", "pass_id", pass.ID)
		return
	}
	defer func() {
		if rec := recover(); rec != nil {
			slog.Error("template ready email panic", "recover", rec)
		}
	}()
	var clientSlug, projectSlug string
	if err := s.DB.QueryRow(`SELECT client_slug, project_slug FROM projects WHERE id = ?`, pass.ProjectID).Scan(&clientSlug, &projectSlug); err != nil {
		slog.Error("template ready email: project lookup", "err", err, "project_id", pass.ProjectID)
		return
	}
	factoryURL := s.portalURL(clientSlug, projectSlug)
	templateURL := fmt.Sprintf("%s/api/projects/%d/word-template", strings.TrimRight(s.BaseURL, "/"), pass.ProjectID)
	if title == "" {
		title = "your book"
	}
	subject := fmt.Sprintf("Your template is ready: %s", title)
	textBody := templateReadyText(pass, title, factoryURL, templateURL)
	htmlBody := templateReadyHTML(pass, title, factoryURL, templateURL)
	if err := s.mail(mailMeta{Kind: mailKindTemplateReady, RefType: "pass", RefID: mailRef(pass.ID), TriggeredBy: "client"}, []string{pass.CustomerEmail}, nil, subject, textBody, htmlBody); err != nil {
		slog.Error("template ready email failed", "err", err, "pass_id", pass.ID)
	}
}

func templateReadyText(pass dbgen.Pass, title, factoryURL, templateURL string) string {
	var t strings.Builder
	fmt.Fprintf(&t, "Hi %s,\n\nYou marked the transmittal for %q final, so your authoring template (.docx) has been generated from it.\n\n", firstName(pass.CustomerName), title)
	fmt.Fprintf(&t, "Download it from step 1 of your factory:\n%s\n\n", factoryURL)
	fmt.Fprintf(&t, "Direct link (works once you're signed in):\n%s\n\n", templateURL)
	fmt.Fprintf(&t, "You do not have to use it. If you already have a draft, keep it in your own editor: Heading 1 for chapter titles, Heading 2 for sub-heads, and a marker such as [[quote]] or [[verse]] at the start of any special paragraph, then export .docx and run Inspect (free). Inside a paragraph, italic, bold and Word's own footnotes need no marker — just use the buttons; they carry through as they are. The template is for starting from a blank page: it carries every paragraph style your transmittal asked for, including your custom styles, and nothing else (also available as .odt for LibreOffice Writer). Either way, upload the finished .docx to the factory and run Inspect.\n\n")
	fmt.Fprintf(&t, "If you change the transmittal later, mark it final again and a fresh template is generated.\n\n")
	fmt.Fprintf(&t, "%s\n", emailSignoffText())
	return t.String()
}

func templateReadyHTML(pass dbgen.Pass, title, factoryURL, templateURL string) string {
	var hb strings.Builder
	hb.WriteString(emailP(fmt.Sprintf("Hi %s,", html.EscapeString(firstName(pass.CustomerName)))))
	hb.WriteString(emailP(fmt.Sprintf("You marked the transmittal for <b>%s</b> final, so your authoring template (.docx) has been generated from it.", html.EscapeString(title))))
	hb.WriteString(emailList([]string{
		emailLink(factoryURL, "Your factory") + " &mdash; the download is in step 1",
		emailLink(templateURL, "Direct download") + " &mdash; works once you&rsquo;re signed in",
	}, false))
	hb.WriteString(emailP("You do not have to use it. If you already have a draft, keep it in your own editor: <b>Heading 1</b> for chapter titles, <b>Heading 2</b> for sub-heads, and a marker such as <code>[[quote]]</code> or <code>[[verse]]</code> at the start of any special paragraph &mdash; then export .docx and run Inspect (free). Inside a paragraph, <i>italic</i>, <b>bold</b> and Word&rsquo;s own footnotes need no marker &mdash; just use the buttons; they carry through as they are. The template is for starting from a blank page: it carries every paragraph style your transmittal asked for, including your custom styles, and nothing else (also available as .odt for LibreOffice Writer). Either way, upload the finished .docx to the factory and run Inspect."))
	hb.WriteString(emailSmall("If you change the transmittal later, mark it final again and a fresh template is generated."))
	hb.WriteString(emailSignoff())
	return emailShell(hb.String(), emailShellOpts{Kicker: "Template ready", Title: title})
}

// buildDeliveredText is the plain-text part of the build-ready receipt.
// buildDeliveredNoun is the subject/kicker for a build receipt.
func buildDeliveredNoun(format string) string {
	switch format {
	case "epub":
		return "EPUB ready"
	case "pdf":
		return "Final print PDF ready"
	}
	return "Final files ready"
}

func buildDeliveredWhat(format string) string {
	switch format {
	case "epub":
		return "EPUB"
	case "pdf":
		return "print PDF"
	}
	return "print PDF and EPUB"
}

func buildDeliveredText(pass dbgen.Pass, book dbgen.Book, format, pdfURL, epubURL, reportURL string, credits, total int64) string {
	var t strings.Builder
	fmt.Fprintf(&t, "Hi %s,\n\nYour %s of %q is done.\n\n", firstName(pass.CustomerName), buildDeliveredWhat(format), book.Title)
	if format != "epub" {
		fmt.Fprintf(&t, "Print PDF:        %s\n", pdfURL)
	}
	if format != "pdf" {
		fmt.Fprintf(&t, "EPUB:             %s\n", epubURL)
	}
	fmt.Fprintf(&t, "Preflight report: %s\n\n", reportURL)
	if format == "both" {
		fmt.Fprintf(&t, "Read the EPUB first: it's the quickest way to see how the machine\nunderstood your file. Then check the print PDF.\n\n")
	}
	fmt.Fprintf(&t, "This was a final: a clean print PDF, no proof footer.\nFinals remaining: %d of %d. Proofs are free and unlimited.\n\n", credits, total)
	fmt.Fprintf(&t, "Sign in to your factory with the client password from your welcome email.\n\n")
	fmt.Fprintf(&t, "The deliverable is a correctly typeset %s of the manuscript as it\n", buildDeliveredWhat(format))
	fmt.Fprintf(&t, "conforms to your transmittal. Preflight tells you what doesn't conform.\n\n")
	fmt.Fprintf(&t, "%s\n", emailSignoffText())
	return t.String()
}

// buildDeliveredHTML is the HTML part of the build-ready receipt.
func buildDeliveredHTML(pass dbgen.Pass, book dbgen.Book, format, pdfURL, epubURL, reportURL string, credits, total int64) string {
	var hb strings.Builder
	hb.WriteString(emailP(fmt.Sprintf("Hi %s,", html.EscapeString(firstName(pass.CustomerName)))))
	hb.WriteString(emailP(fmt.Sprintf("Your %s of <b>%s</b> is done.", buildDeliveredWhat(format), html.EscapeString(book.Title))))
	var links []string
	if format != "epub" {
		links = append(links, emailLink(pdfURL, "Print PDF"))
	}
	if format != "pdf" {
		links = append(links, emailLink(epubURL, "EPUB"))
	}
	links = append(links, emailLink(reportURL, "Preflight report"))
	hb.WriteString(emailList(links, false))
	if format == "both" {
		hb.WriteString(emailP("Read the EPUB first: it&rsquo;s the quickest way to see how the machine understood your file. Then check the print PDF."))
	}
	hb.WriteString(emailP(fmt.Sprintf("This was a final: a clean print PDF, no proof footer. <b>Finals remaining: %d of %d.</b> Proofs are free and unlimited.", credits, total)))
	hb.WriteString(emailSmall("Sign in to your factory with the client password from your welcome email."))
	hb.WriteString(emailSmall(fmt.Sprintf("The deliverable is a correctly typeset %s of the manuscript as it conforms to your transmittal. Preflight tells you what doesn&rsquo;t conform.", buildDeliveredWhat(format))))
	hb.WriteString(emailSignoff())
	return emailShell(hb.String(), emailShellOpts{Kicker: buildDeliveredNoun(format), Title: book.Title})
}
