package srv

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ─── Public workshop registration ───
//
// A single public, unauthenticated endpoint (POST /api/public/register) that
// captures a lightweight cohort application for the "Protocolize Your Book"
// workshop (Protocol Symposium 2026). Rows land in event_registrations; the
// admin reviews/curates via /admin/registrations. On submit we notify the
// organizer and send the applicant a "request received" auto-reply (the actual
// seat confirmation is manual, since the cohort is soft-capped and curated).
//
// Anti-abuse: a hidden honeypot field ("company") plus a small per-IP rate
// limiter. No captcha — this is a niche form with a human in the loop.

const workshopSlug = "protocolize-your-book"

// workshopSoftCap is the target cohort size; submissions past it are accepted
// but flagged waitlist-ish in the admin view. Not enforced at the DB level.
const workshopSoftCap = 8

// ── tiny in-memory per-IP rate limiter (no external deps) ──

type regRateLimiter struct {
	mu   sync.Mutex
	hits map[string][]time.Time
}

func newRegRateLimiter() *regRateLimiter {
	return &regRateLimiter{hits: make(map[string][]time.Time)}
}

// allow reports whether ip may submit now: max `limit` submissions per `window`.
func (l *regRateLimiter) allow(ip string, limit int, window time.Duration) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-window)
	kept := l.hits[ip][:0]
	for _, t := range l.hits[ip] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	// Opportunistically prune unrelated stale entries so the map can't grow
	// unbounded over a long-lived process.
	for k, v := range l.hits {
		if k == ip {
			continue
		}
		if len(v) == 0 || v[len(v)-1].Before(cutoff) {
			delete(l.hits, k)
		}
	}
	if len(kept) >= limit {
		l.hits[ip] = kept
		return false
	}
	l.hits[ip] = append(kept, now)
	return true
}

// limiter returns the Server's own rate limiter, lazily creating one so both
// New() and hand-constructed test Servers get an isolated limiter (avoids
// cross-test bleed on the shared loopback IP).
func (s *Server) limiter() *regRateLimiter {
	s.regLimiterOnce.Do(func() {
		if s.regLimiter == nil {
			s.regLimiter = newRegRateLimiter()
		}
	})
	return s.regLimiter
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.IndexByte(xff, ','); i >= 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

type registrationInput struct {
	Name          string `json:"name"`
	Email         string `json:"email"`
	Region        string `json:"region"`
	AllSessions   bool   `json:"all_sessions"`
	Material      string `json:"material"`
	MaterialType  string `json:"material_type"`
	Background    string `json:"background"`
	NewToProtocol bool   `json:"new_to_protocol"`
	Goals         string `json:"goals"`
	ConsentEmail  bool   `json:"consent_email"`
	Company       string `json:"company"` // honeypot — must be empty
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

func clip(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) > n {
		return s[:n]
	}
	return s
}

// looksLikeEmail is a cheap sanity check (not RFC-perfect on purpose).
func looksLikeEmail(s string) bool {
	at := strings.IndexByte(s, '@')
	if at <= 0 || at == len(s)-1 {
		return false
	}
	return strings.IndexByte(s[at+1:], '.') >= 0 && !strings.ContainsAny(s, " \t\r\n")
}

func (s *Server) handlePublicRegister(w http.ResponseWriter, r *http.Request) {
	ip := clientIP(r)
	if !s.limiter().allow(ip, 5, 10*time.Minute) {
		jsonErr(w, "Too many submissions. Please try again in a few minutes.", http.StatusTooManyRequests)
		return
	}

	var in registrationInput
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64*1024)).Decode(&in); err != nil {
		jsonErr(w, "Sorry, we couldn't read that submission.", http.StatusBadRequest)
		return
	}

	// Honeypot: real users never fill this hidden field. Pretend success so
	// bots don't learn they were caught.
	if strings.TrimSpace(in.Company) != "" {
		slog.Info("registration honeypot tripped", "ip", ip)
		jsonOK(w, map[string]any{"ok": true})
		return
	}

	in.Name = clip(in.Name, 200)
	in.Email = clip(strings.ToLower(in.Email), 320)
	in.Region = clip(in.Region, 40)
	in.Material = clip(in.Material, 2000)
	in.MaterialType = clip(in.MaterialType, 40)
	in.Background = clip(in.Background, 40)
	in.Goals = clip(in.Goals, 2000)

	if in.Name == "" || in.Email == "" {
		jsonErr(w, "Please give your name and email.", http.StatusBadRequest)
		return
	}
	if !looksLikeEmail(in.Email) {
		jsonErr(w, "That email address doesn't look right.", http.StatusBadRequest)
		return
	}
	if in.Material == "" {
		jsonErr(w, "Please tell us what material you'll bring — it's the one hard requirement.", http.StatusBadRequest)
		return
	}

	// Upsert on (event_slug, email): a re-submission updates the existing row
	// rather than erroring on the unique index.
	res, err := s.DB.ExecContext(r.Context(), `
		INSERT INTO event_registrations
			(event_slug, name, email, region, all_sessions, material, material_type,
			 background, new_to_protocol, goals, consent_email, user_agent, ip)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(event_slug, email) DO UPDATE SET
			name=excluded.name, region=excluded.region, all_sessions=excluded.all_sessions,
			material=excluded.material, material_type=excluded.material_type,
			background=excluded.background, new_to_protocol=excluded.new_to_protocol,
			goals=excluded.goals, consent_email=excluded.consent_email,
			user_agent=excluded.user_agent, ip=excluded.ip,
			updated_at=CURRENT_TIMESTAMP
	`,
		workshopSlug, in.Name, in.Email, in.Region, b2i(in.AllSessions), in.Material,
		in.MaterialType, in.Background, b2i(in.NewToProtocol), in.Goals,
		b2i(in.ConsentEmail), clip(r.UserAgent(), 500), ip,
	)
	if err != nil {
		slog.Error("registration insert failed", "err", err, "email", in.Email)
		jsonErr(w, "Something went wrong saving your registration. Please email us instead.", http.StatusInternalServerError)
		return
	}
	rowID, _ := res.LastInsertId()

	// Fire-and-forget emails: never block or fail the user's submission on the
	// mailer. Copy values into the closure (r/ctx won't outlive the request).
	if s.Email != nil {
		go s.sendRegistrationEmails(in, rowID)
	} else {
		slog.Warn("registration received but email not configured", "email", in.Email)
	}

	jsonOK(w, map[string]any{"ok": true})
}

// sendRegistrationEmails sends the organizer notification and the applicant
// auto-reply. Runs in its own goroutine; logs but never surfaces errors.
func (s *Server) sendRegistrationEmails(in registrationInput, rowID int64) {
	defer func() {
		if rec := recover(); rec != nil {
			slog.Error("registration email panic", "recover", rec)
		}
	}()

	// Count current requests to note cohort pressure in the organizer email.
	var total int
	_ = s.DB.QueryRow(`SELECT COUNT(*) FROM event_registrations WHERE event_slug=?`, workshopSlug).Scan(&total)

	// 1) Organizer notification (to the Reply-To / contact address).
	organizer := strings.TrimSpace(s.Email.ReplyTo)
	if organizer == "" {
		organizer = "j@djinna.com"
	}
	yn := func(b bool) string {
		if b {
			return "yes"
		}
		return "no"
	}
	capNote := ""
	if total > workshopSoftCap {
		capNote = fmt.Sprintf(" — OVER soft cap of %d", workshopSoftCap)
	}
	subj := fmt.Sprintf("New workshop registration: %s (#%d of %d%s)", in.Name, total, workshopSoftCap, capNote)
	var t strings.Builder
	fmt.Fprintf(&t, "New registration for Protocolize Your Book.\n\n")
	fmt.Fprintf(&t, "Name:            %s\n", in.Name)
	fmt.Fprintf(&t, "Email:           %s\n", in.Email)
	fmt.Fprintf(&t, "Region:          %s\n", firstNonEmpty(in.Region, "—"))
	fmt.Fprintf(&t, "Can attend all4: %s\n", yn(in.AllSessions))
	fmt.Fprintf(&t, "Material:        %s\n", in.Material)
	fmt.Fprintf(&t, "Material type:   %s\n", firstNonEmpty(in.MaterialType, "—"))
	fmt.Fprintf(&t, "Background:      %s\n", firstNonEmpty(in.Background, "—"))
	fmt.Fprintf(&t, "New to protocol: %s\n", yn(in.NewToProtocol))
	fmt.Fprintf(&t, "Goals:           %s\n", firstNonEmpty(in.Goals, "—"))
	fmt.Fprintf(&t, "OK to email:     %s\n", yn(in.ConsentEmail))
	fmt.Fprintf(&t, "\nRequests so far: %d (soft cap %d)\n", total, workshopSoftCap)
	fmt.Fprintf(&t, "Review: %s/admin/registrations\n", strings.TrimRight(s.BaseURL, "/"))

	htmlBody := "<pre style=\"font:14px/1.5 ui-monospace,Menlo,Consolas,monospace\">" + html.EscapeString(t.String()) + "</pre>"
	if err := s.Email.sendEmail([]string{organizer}, nil, subj, t.String(), htmlBody); err != nil {
		slog.Error("registration organizer email failed", "err", err, "email", in.Email)
	}

	// 2) Applicant auto-reply.
	applicantSubj := "We got your Protocolize Your Book registration"
	txt := applicantAutoReplyText(in.Name)
	htmlReply := applicantAutoReplyHTML(in.Name)
	if err := s.Email.sendEmail([]string{in.Email}, nil, applicantSubj, txt, htmlReply); err != nil {
		slog.Error("registration applicant email failed", "err", err, "email", in.Email)
	}
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

func applicantAutoReplyText(name string) string {
	first := firstName(name)
	return fmt.Sprintf(`Hi %s,

Thanks — your request for the "Protocolize Your Book" workshop at Protocol
Symposium 2026 is in.

This is a small hands-on lab (8 seats), so we curate for a mix of source
material and backgrounds. We'll email within a few days to confirm your spot,
with the Zoom link, calendar invites for all four sessions, and a short note on
prepping your manuscript. If the cohort fills, we'll offer you the session
recordings and a spot in the next round.

The four sessions (all times UTC, cumulative — please plan to attend all four):
  1. Handshake         Mon Sept 21   15:00–16:30 UTC
  2. Preflight         Mon Sept 21   17:00–18:30 UTC
  3. Build + delegate  Tue Sept 22   15:00–16:30 UTC
  4. Show-and-tell     Tue Sept 22   17:00–18:30 UTC
  (US Eastern 11:00 / 13:00 · US Pacific 08:00 / 10:00 · Central Europe 17:00 / 19:00)

One reminder: the workshop runs on YOUR material, so have a real manuscript or
text collection ready to bring — any size, rough is welcome.

See you in the factory,
Jenna Dixon · jdbb studio`, first)
}

func applicantAutoReplyHTML(name string) string {
	first := html.EscapeString(firstName(name))
	return fmt.Sprintf(`<div style="font:15px/1.6 -apple-system,Segoe UI,Helvetica,Arial,sans-serif;color:#0E1116;max-width:560px">
<p>Hi %s,</p>
<p>Thanks — your request for the <b>Protocolize Your Book</b> workshop at Protocol Symposium 2026 is in.</p>
<p>This is a small hands-on lab (<b>8 seats</b>), so we curate for a mix of source material and backgrounds. We&rsquo;ll email within a few days to confirm your spot, with the Zoom link, calendar invites for all four sessions, and a short note on prepping your manuscript. If the cohort fills, we&rsquo;ll offer you the session recordings and a spot in the next round.</p>
<p style="margin:0 0 6px"><b>The four sessions</b> (all times UTC, cumulative &mdash; please plan to attend all four):</p>
<table style="border-collapse:collapse;font:13px/1.5 ui-monospace,Menlo,Consolas,monospace">
<tr><td style="padding:2px 14px 2px 0">1. Handshake</td><td style="padding:2px 14px 2px 0">Mon Sept 21</td><td>15:00&ndash;16:30 UTC</td></tr>
<tr><td style="padding:2px 14px 2px 0">2. Preflight</td><td style="padding:2px 14px 2px 0">Mon Sept 21</td><td>17:00&ndash;18:30 UTC</td></tr>
<tr><td style="padding:2px 14px 2px 0">3. Build + delegate</td><td style="padding:2px 14px 2px 0">Tue Sept 22</td><td>15:00&ndash;16:30 UTC</td></tr>
<tr><td style="padding:2px 14px 2px 0">4. Show-and-tell</td><td style="padding:2px 14px 2px 0">Tue Sept 22</td><td>17:00&ndash;18:30 UTC</td></tr>
</table>
<p style="color:#5D6B76;font-size:13px">US Eastern 11:00 / 13:00 &middot; US Pacific 08:00 / 10:00 &middot; Central Europe 17:00 / 19:00</p>
<p>One reminder: the workshop runs on <b>your</b> material, so have a real manuscript or text collection ready to bring &mdash; any size, rough is welcome.</p>
<p>See you in the factory,<br>Jenna Dixon &middot; jdbb studio</p>
</div>`, first)
}

func firstName(full string) string {
	full = strings.TrimSpace(full)
	if full == "" {
		return "there"
	}
	if i := strings.IndexByte(full, ' '); i > 0 {
		return full[:i]
	}
	return full
}

// ─── Admin: list / export registrations ───

type registrationRow struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	Region        string `json:"region"`
	AllSessions   bool   `json:"all_sessions"`
	Material      string `json:"material"`
	MaterialType  string `json:"material_type"`
	Background    string `json:"background"`
	NewToProtocol bool   `json:"new_to_protocol"`
	Goals         string `json:"goals"`
	ConsentEmail  bool   `json:"consent_email"`
	Status        string `json:"status"`
	Notes         string `json:"notes"`
	CreatedAt     string `json:"created_at"`
}

func (s *Server) queryRegistrations(r *http.Request) ([]registrationRow, error) {
	rows, err := s.DB.QueryContext(r.Context(), `
		SELECT id, name, email, region, all_sessions, material, material_type,
		       background, new_to_protocol, goals, consent_email, status, notes, created_at
		FROM event_registrations
		WHERE event_slug = ?
		ORDER BY created_at ASC
	`, workshopSlug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []registrationRow{}
	for rows.Next() {
		var e registrationRow
		var allS, newP, consent int
		var notes sql.NullString
		if err := rows.Scan(&e.ID, &e.Name, &e.Email, &e.Region, &allS, &e.Material,
			&e.MaterialType, &e.Background, &newP, &e.Goals, &consent, &e.Status,
			&notes, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.AllSessions = allS == 1
		e.NewToProtocol = newP == 1
		e.ConsentEmail = consent == 1
		e.Notes = notes.String
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Server) handleAdminRegistrationsPage(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdmin(w, r) {
		return
	}
	s.serveStaticHTML(w, "static/registrations.html")
}

func (s *Server) handleAdminListRegistrations(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	rows, err := s.queryRegistrations(r)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	jsonOK(w, map[string]any{
		"registrations": rows,
		"count":         len(rows),
		"soft_cap":      workshopSoftCap,
	})
}

func csvQuote(s string) string {
	if strings.ContainsAny(s, ",\"\n\r") {
		return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
	}
	return s
}

func (s *Server) handleAdminExportRegistrations(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	rows, err := s.queryRegistrations(r)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="protocolize-registrations.csv"`)
	var b strings.Builder
	b.WriteString("id,created_at,name,email,region,all_sessions,material_type,material,background,new_to_protocol,goals,consent_email,status,notes\n")
	for _, e := range rows {
		cells := []string{
			strconv.FormatInt(e.ID, 10), e.CreatedAt, e.Name, e.Email, e.Region,
			strconv.FormatBool(e.AllSessions), e.MaterialType, e.Material, e.Background,
			strconv.FormatBool(e.NewToProtocol), e.Goals, strconv.FormatBool(e.ConsentEmail),
			e.Status, e.Notes,
		}
		for i, c := range cells {
			if i > 0 {
				b.WriteByte(',')
			}
			b.WriteString(csvQuote(c))
		}
		b.WriteByte('\n')
	}
	_, _ = w.Write([]byte(b.String()))
}
