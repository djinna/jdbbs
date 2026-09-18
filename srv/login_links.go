package srv

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// Magic-link client sign-in (punch list 5.7, EMAIL_SYSTEM.md pathway 8).
//
// A customer types the email address their Factory Pass (or client record)
// is tied to; we mail a one-shot link that sets the same client-level cookie
// a password login sets. The token is 32 random bytes, base64url in the URL,
// stored only as its sha256 hex. 30-minute expiry, single use, and at most
// three live links per client per 15 minutes. The request endpoint always
// answers {ok:true} so nobody can probe which addresses exist.

const (
	loginLinkTTL        = 30 * time.Minute
	loginLinkRateWin    = 15 * time.Minute
	loginLinkRateMax    = 3
	mailKindLoginLink   = "login_link"
	loginLinkTimeLayout = "2006-01-02 15:04:05"
)

// loginLinkEmailMatches reports whether email (already lowercased/trimmed)
// is an address we recognise for the client: the client's own contact
// address, any Factory Pass on one of its projects, a store order for such
// a pass, or the coupon/registration the pass was redeemed from.
func (s *Server) loginLinkEmailMatches(ctx context.Context, slug, email string) (bool, error) {
	if slug == "" || email == "" || !strings.Contains(email, "@") {
		return false, nil
	}
	var one int
	err := s.DB.QueryRowContext(ctx, `
		SELECT 1 FROM clients c WHERE c.slug = ?1 AND lower(trim(c.email)) = ?2
		UNION ALL
		SELECT 1 FROM passes p JOIN projects pr ON pr.id = p.project_id
			WHERE pr.client_slug = ?1 AND lower(trim(p.customer_email)) = ?2
		UNION ALL
		SELECT 1 FROM store_orders so JOIN passes p ON p.id = so.pass_id
			JOIN projects pr ON pr.id = p.project_id
			WHERE pr.client_slug = ?1 AND lower(trim(so.customer_email)) = ?2
		UNION ALL
		SELECT 1 FROM coupons cp JOIN passes p ON p.coupon_id = cp.id
			JOIN projects pr ON pr.id = p.project_id
			LEFT JOIN event_registrations er ON er.id = cp.registration_id
			WHERE pr.client_slug = ?1
			  AND (lower(trim(cp.issued_to_email)) = ?2 OR lower(trim(er.email)) = ?2)
		LIMIT 1`, slug, email).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

// newLoginLinkToken returns the URL token and its stored hash.
func newLoginLinkToken() (token, hash string, err error) {
	raw := make([]byte, 32)
	if _, err = rand.Read(raw); err != nil {
		return "", "", err
	}
	token = base64.RawURLEncoding.EncodeToString(raw)
	return token, loginLinkHash(token), nil
}

func loginLinkHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// issueLoginLink creates a row and returns the plaintext token (to be mailed,
// never stored). ok=false with nil err means the per-client rate limit hit.
func (s *Server) issueLoginLink(ctx context.Context, slug, email, ip string) (token string, ok bool, err error) {
	now := time.Now().UTC()
	var live int
	if err = s.DB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM login_links
		WHERE client_slug = ? AND used_at IS NULL AND created_at > ?`,
		slug, now.Add(-loginLinkRateWin).Format(loginLinkTimeLayout)).Scan(&live); err != nil {
		return "", false, err
	}
	if live >= loginLinkRateMax {
		return "", false, nil
	}
	token, hash, err := newLoginLinkToken()
	if err != nil {
		return "", false, err
	}
	_, err = s.DB.ExecContext(ctx, `
		INSERT INTO login_links (client_slug, token_hash, email, created_at, expires_at, ip)
		VALUES (?, ?, ?, ?, ?, ?)`,
		slug, hash, email, now.Format(loginLinkTimeLayout), now.Add(loginLinkTTL).Format(loginLinkTimeLayout), ip)
	if err != nil {
		return "", false, err
	}
	return token, true, nil
}

// loginLinkURL is the absolute redeem URL, built like the fulfilment mail's
// portal link (s.BaseURL, trailing slash trimmed).
func (s *Server) loginLinkURL(token string) string {
	return fmt.Sprintf("%s/auth/link?t=%s", strings.TrimRight(s.BaseURL, "/"), token)
}

// handlePublicLoginLink — POST /api/public/login-link {client, email}.
// Always 200 {ok:true}. Sends only when the address is one we know for that
// client and the rate limit allows.
func (s *Server) handlePublicLoginLink(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Client string `json:"client"`
		Email  string `json:"email"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&body); err != nil {
		jsonErr(w, "bad request", 400)
		return
	}
	slug := normalizeProjectSlug(strings.TrimSpace(body.Client))
	email := strings.ToLower(strings.TrimSpace(body.Email))
	// Answer first in spirit: whatever happens below, the caller sees ok.
	defer jsonOK(w, map[string]any{"ok": true})

	var name string
	if err := s.DB.QueryRowContext(r.Context(), `SELECT name FROM clients WHERE slug = ?`, slug).Scan(&name); err != nil {
		slog.Info("login link: unknown client", "client", slug)
		return
	}
	match, err := s.loginLinkEmailMatches(r.Context(), slug, email)
	if err != nil {
		slog.Error("login link: match query", "err", err, "client", slug)
		return
	}
	if !match {
		slog.Info("login link: email not recognised for client", "client", slug)
		s.factoryEvent(0, slug, "login.link_denied", "anon", "sign-in link requested with an unrecognised address")
		return
	}
	token, ok, err := s.issueLoginLink(r.Context(), slug, email, clientIP(r))
	if err != nil {
		slog.Error("login link: issue", "err", err, "client", slug)
		return
	}
	if !ok {
		slog.Warn("login link: rate limited", "client", slug)
		return
	}
	slog.Info("login link issued", "client", slug, "ip", clientIP(r))
	s.factoryEvent(0, slug, "login.link_sent", "anon", "sign-in link emailed for /"+slug+"/")
	if name == "" {
		name = slug
	}
	s.sendLoginLinkEmail(slug, name, email, s.loginLinkURL(token))
}

// sendLoginLinkEmail mails the link. Fire-and-forget like the fulfilment
// mail: the request already answered ok, and s.Email is nil in dev/test.
func (s *Server) sendLoginLinkEmail(slug, clientName, email, link string) {
	if s.Email == nil {
		slog.Warn("login link created but email not configured", "client", slug)
		return
	}
	send := func() {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("login link email panic", "recover", rec)
			}
		}()
		subject := fmt.Sprintf("Your sign-in link for %s", clientName)
		err := s.mail(mailMeta{Kind: mailKindLoginLink, RefType: "client", RefID: slug, TriggeredBy: "public"},
			[]string{email}, nil, subject, loginLinkText(clientName, link), loginLinkHTML(clientName, link))
		if err != nil {
			slog.Error("login link email failed", "err", err, "client", slug)
		}
	}
	go send()
}

func loginLinkText(clientName, link string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Hi,\n\nHere is your sign-in link for %s:\n\n%s\n\n", clientName, link)
	b.WriteString("It works once and for 30 minutes. If it has expired, go back to your portal and enter your email again for a fresh one.\n\n")
	b.WriteString("If you didn't ask for this, you can ignore it — nothing changes until the link is opened.\n\n")
	fmt.Fprintf(&b, "%s\n", emailSignoffText())
	return b.String()
}

func loginLinkHTML(clientName, link string) string {
	var b strings.Builder
	b.WriteString(emailP("Hi,"))
	b.WriteString(emailP(fmt.Sprintf("Here is your sign-in link for <b>%s</b>:", html.EscapeString(clientName))))
	b.WriteString(emailButton(link, "Sign in"))
	b.WriteString(emailSmall("Or paste this into your browser: " + emailLink(link, link)))
	b.WriteString(emailP("It works once and for 30 minutes. If it has expired, go back to your portal and enter your email again for a fresh one."))
	b.WriteString(emailSmall("If you didn&rsquo;t ask for this, you can ignore it &mdash; nothing changes until the link is opened."))
	b.WriteString(emailSignoff())
	return emailShell(b.String(), emailShellOpts{Kicker: "Sign-in link", Title: clientName})
}

// handleAuthLink — GET /auth/link?t=TOKEN. Verifies, burns, sets the client
// cookie, and 302s to the portal. Anything else gets the expired page.
func (s *Server) handleAuthLink(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.URL.Query().Get("t"))
	if token == "" || len(token) > 128 {
		s.serveLoginLinkExpired(w, "", http.StatusUnauthorized)
		return
	}
	hash := loginLinkHash(token)
	var (
		id               int64
		slug, storedHash string
		expires          time.Time // TIMESTAMP columns come back as time.Time from the driver
		usedAt           sql.NullTime
	)
	err := s.DB.QueryRowContext(r.Context(), `
		SELECT id, client_slug, token_hash, expires_at, used_at FROM login_links WHERE token_hash = ?`, hash).
		Scan(&id, &slug, &storedHash, &expires, &usedAt)
	if err != nil {
		slog.Warn("login link: no such token", "ip", clientIP(r))
		s.serveLoginLinkExpired(w, "", http.StatusUnauthorized)
		return
	}
	// The lookup already matched; keep the compare constant-time anyway.
	if subtle.ConstantTimeCompare([]byte(hash), []byte(storedHash)) != 1 {
		s.serveLoginLinkExpired(w, slug, http.StatusUnauthorized)
		return
	}
	expired := time.Now().UTC().After(expires)
	if usedAt.Valid || expired {
		slog.Info("login link rejected", "client", slug, "used", usedAt.Valid, "expired", expired)
		s.serveLoginLinkExpired(w, slug, http.StatusUnauthorized)
		return
	}
	// Burn it. The WHERE guard makes two racing redeems resolve to one winner.
	res, err := s.DB.ExecContext(r.Context(), `UPDATE login_links SET used_at = ? WHERE id = ? AND used_at IS NULL`,
		time.Now().UTC().Format(loginLinkTimeLayout), id)
	if err != nil {
		slog.Error("login link: mark used", "err", err, "client", slug)
		s.serveLoginLinkExpired(w, slug, http.StatusInternalServerError)
		return
	}
	if n, _ := res.RowsAffected(); n != 1 {
		s.serveLoginLinkExpired(w, slug, http.StatusUnauthorized)
		return
	}
	var passwordHash string
	if err := s.DB.QueryRowContext(r.Context(), `SELECT password_hash FROM clients WHERE slug = ?`, slug).Scan(&passwordHash); err != nil {
		s.serveLoginLinkExpired(w, slug, http.StatusUnauthorized)
		return
	}
	if passwordHash != "" {
		// Same cookie as a password login (client.go). A client with no
		// password has an open portal, so there is nothing to set.
		s.setClientAuthCookie(w, slug, passwordHash)
	}
	slog.Info("login link redeemed", "client", slug, "ip", clientIP(r))
	s.factoryEvent(0, slug, "login", "client:"+slug, "signed in to /"+slug+"/ via emailed link")
	http.Redirect(w, r, "/"+slug+"/", http.StatusFound)
}

// serveLoginLinkExpired is the small themed page for a dead link. slug may be
// empty (token unknown) — then the button goes to the homepage.
func (s *Server) serveLoginLinkExpired(w http.ResponseWriter, slug string, code int) {
	back := "/"
	backLabel := "Back to start"
	if slug != "" {
		back = "/" + slug + "/"
		backLabel = "Request a new link"
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Sign-in link expired — jdbb studio</title>
<meta name="robots" content="noindex">
<link rel="icon" href="/static/favicon.svg?v=2" type="image/svg+xml">
<link rel="stylesheet" href="/static/theme.css">
<style>
*{box-sizing:border-box} body{margin:0;background:var(--bg);color:var(--text);font:16px/1.6 var(--body)}
.auth-box{max-width:420px;margin:96px auto 0;padding:40px 4px 44px;border-top:1px solid var(--border-strong);border-bottom:1px solid var(--border-strong)}
.auth-box h2{font-family:var(--mono);font-size:20px;font-weight:600;letter-spacing:-.01em;margin:14px 0 8px}
.auth-box .auth-sub{color:var(--text-secondary);font-size:14px;margin-bottom:24px}
.auth-box .btn-fill{width:100%%;display:block;text-align:center;text-decoration:none}
.back-link{margin-top:12px}
</style>
</head>
<body>
<div class="jdbb-shell">
<header class="jdbb-masthead"><a class="jdbb-wordmark" href="/"><span class="bracket">[</span><span class="kj">j</span>dbb<span class="bracket">]</span><span class="studio">studio</span></a><nav data-public-nav><div id="theme-bar"></div></nav></header>
<main>
<div class="auth-box">
  <div class="kicker">Client portal</div>
  <h2>That sign-in link has expired</h2>
  <p class="auth-sub">Sign-in links work once and for 30 minutes. Go back to your portal, enter your email address, and we&rsquo;ll send you a fresh one.</p>
  <a class="btn-fill" href="%s">%s</a>
  <div class="back-link"><a href="/" class="link-action">&larr; Home</a></div>
</div>
</main>
<footer class="jdbb-footer"><a class="jdbb-wordmark" href="/"><span class="bracket">[</span><span class="kj">j</span>dbb<span class="bracket">]</span></a><nav aria-label="Footer"><a href="/">Home</a><a href="/field-notes">Field notes</a><a href="/workshop">Workshop</a></nav><span class="copy">&copy; 2026 Jenna Dixon</span></footer>
</div>
<script src="/static/theme.js"></script>
</body>
</html>
`, html.EscapeString(back), html.EscapeString(backLabel))
}
