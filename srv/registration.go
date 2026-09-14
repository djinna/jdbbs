package srv

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html"
	"log/slog"
	"net"
	"net/http"
	"regexp"
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

const workshopSlug = "protocolize-your-book-2026-09"

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

	htmlBody := emailShell(
		fmt.Sprintf(`<pre style="margin:0;font:13px/1.5 %s;color:%s;white-space:pre-wrap">%s</pre>`, emailMono, emailText, html.EscapeString(t.String())),
		emailShellOpts{Kicker: "New registration", Title: subj, Footer: "Automated notice from jdbb studio registration."},
	)
	if err := s.mail(mailMeta{Kind: mailKindRegistrationAlert, RefType: "registration", RefID: mailRef(rowID), TriggeredBy: "public"}, []string{organizer}, nil, subj, t.String(), htmlBody); err != nil {
		slog.Error("registration organizer email failed", "err", err, "email", in.Email)
	}

	// 2) Applicant auto-reply.
	applicantSubj := "We got your Protocolize Your Book registration"
	txt := applicantAutoReplyText(in.Name)
	htmlReply := applicantAutoReplyHTML(in.Name)
	if err := s.mail(mailMeta{Kind: mailKindRegistrationConfirm, RefType: "registration", RefID: mailRef(rowID), TriggeredBy: "public"}, []string{in.Email}, nil, applicantSubj, txt, htmlReply); err != nil {
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
with the Discord invite, calendar invites for all four sessions, and a short note on
prepping your manuscript. If the cohort fills, we'll offer you the session
recordings and a spot in the next round.

The four sessions (all times UTC, cumulative — please plan to attend all four):
  1. Handshake         Mon Sept 21   15:00–16:30 UTC
  2. Preflight         Mon Sept 21   16:30–18:00 UTC
  3. Build + delegate  Tue Sept 22   15:00–16:30 UTC
  4. Show-and-tell     Tue Sept 22   16:30–18:00 UTC
  (US Eastern 11:00 / 12:30 · US Pacific 08:00 / 09:30 · Central Europe 17:00 / 18:30)

One reminder: the workshop runs on YOUR material, so have a real manuscript or
text collection ready to bring — any size, rough is welcome.

See you in the factory,
Jenna Dixon · jdbb studio`, first)
}

func applicantAutoReplyHTML(name string) string {
	first := html.EscapeString(firstName(name))
	var b strings.Builder
	b.WriteString(emailP(fmt.Sprintf("Hi %s,", first)))
	b.WriteString(emailP("Thanks &mdash; your request for the <b>Protocolize Your Book</b> workshop at Protocol Symposium 2026 is in."))
	b.WriteString(emailP("This is a small hands-on lab (<b>8 seats</b>), so we curate for a mix of source material and backgrounds. We&rsquo;ll email within a few days to confirm your spot, with the Discord invite, calendar invites for all four sessions, and a short note on prepping your manuscript. If the cohort fills, we&rsquo;ll offer you the session recordings and a spot in the next round."))
	b.WriteString(emailH2("The four sessions"))
	b.WriteString(emailSmall("All times UTC, cumulative &mdash; please plan to attend all four."))
	b.WriteString(emailTable([]string{"#", "Session", "Day", "Time (UTC)"}, [][]string{
		{"1", "Handshake", "Mon Sept 21", "15:00&ndash;16:30"},
		{"2", "Preflight", "Mon Sept 21", "16:30&ndash;18:00"},
		{"3", "Build + delegate", "Tue Sept 22", "15:00&ndash;16:30"},
		{"4", "Show-and-tell", "Tue Sept 22", "16:30&ndash;18:00"},
	}, nil))
	b.WriteString(emailSmall("US Eastern 11:00 / 12:30 &middot; US Pacific 08:00 / 09:30 &middot; Central Europe 17:00 / 18:30"))
	b.WriteString(emailP("One reminder: the workshop runs on <b>your</b> material, so have a real manuscript or text collection ready to bring &mdash; any size, rough is welcome."))
	b.WriteString(emailP("See you in the factory,"))
	b.WriteString(emailSignoff())
	return emailShell(b.String(), emailShellOpts{Kicker: "Protocolize Your Book", Title: "Your registration is in"})
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
	ID               int64  `json:"id"`
	Name             string `json:"name"`
	Email            string `json:"email"`
	Region           string `json:"region"`
	AllSessions      bool   `json:"all_sessions"`
	Material         string `json:"material"`
	MaterialType     string `json:"material_type"`
	Background       string `json:"background"`
	NewToProtocol    bool   `json:"new_to_protocol"`
	Goals            string `json:"goals"`
	ConsentEmail     bool   `json:"consent_email"`
	Status           string `json:"status"`
	PrepStatus       string `json:"prep_status"`
	AttendedSessions int    `json:"attended_sessions"`
	Notes            string `json:"notes"`
	LastEmailedAt    string `json:"last_emailed_at"`
	CreatedAt        string `json:"created_at"`
	// Factory Pass tracking: the code issued to this registrant (if any) and
	// when it was redeemed, so the tracker can show issued/redeemed at a glance.
	CouponCode       string `json:"coupon_code"`
	CouponRedeemedAt string `json:"coupon_redeemed_at"`
	ClientSlug       string `json:"client_slug"`
	ProjectPath      string `json:"project_path"`
}

func (s *Server) queryRegistrations(r *http.Request) ([]registrationRow, error) {
	// The coupons/passes LEFT JOIN adds Factory Pass state per registrant.
	// A registrant has at most one code in practice; MAX() keeps the row
	// count stable if one ever gets a second.
	rows, err := s.DB.QueryContext(r.Context(), `
		SELECT e.id, e.name, e.email, e.region, e.all_sessions, e.material, e.material_type,
		       e.background, e.new_to_protocol, e.goals, e.consent_email, e.status,
		       e.prep_status, e.attended_sessions, e.notes, e.last_emailed_at, e.created_at,
		       COALESCE(MAX(c.code), '') AS coupon_code,
		       MAX(p.fulfilled_at)       AS coupon_redeemed_at,
		       COALESCE(MAX(pr.client_slug), '') AS client_slug,
		       COALESCE(MAX(pr.project_slug), '') AS project_slug
		FROM event_registrations e
		LEFT JOIN coupons c ON c.registration_id = e.id
		LEFT JOIN passes  p ON p.coupon_id = c.id
		LEFT JOIN projects pr ON pr.id = p.project_id
		WHERE e.event_slug = ?
		GROUP BY e.id
		ORDER BY e.created_at ASC
	`, workshopSlug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []registrationRow{}
	for rows.Next() {
		var e registrationRow
		var allS, newP, consent int
		var notes, lastEmailed, couponRedeemed sql.NullString
		var projectSlug string
		if err := rows.Scan(&e.ID, &e.Name, &e.Email, &e.Region, &allS, &e.Material,
			&e.MaterialType, &e.Background, &newP, &e.Goals, &consent, &e.Status,
			&e.PrepStatus, &e.AttendedSessions, &notes, &lastEmailed, &e.CreatedAt,
			&e.CouponCode, &couponRedeemed, &e.ClientSlug, &projectSlug); err != nil {
			return nil, err
		}
		e.AllSessions = allS == 1
		e.NewToProtocol = newP == 1
		e.ConsentEmail = consent == 1
		e.Notes = notes.String
		e.LastEmailedAt = lastEmailed.String
		e.CouponRedeemedAt = couponRedeemed.String
		if e.ClientSlug != "" && projectSlug != "" {
			e.ProjectPath = "/" + e.ClientSlug + "/" + projectSlug + "/factory/"
		}
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
	announcements, err := s.queryAnnouncements(r)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	mail, err := s.queryRegistrationMail(r)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	jsonOK(w, map[string]any{
		"registrations": rows,
		"announcements": announcements,
		"mail":          mail,
		"count":         len(rows),
		"soft_cap":      workshopSoftCap,
		"event_slug":    workshopSlug,
		"email_ready":   s.Email != nil,
	})
}

type registrationUpdate struct {
	Name             *string `json:"name"`
	Email            *string `json:"email"`
	Region           *string `json:"region"`
	AllSessions      *bool   `json:"all_sessions"`
	Material         *string `json:"material"`
	MaterialType     *string `json:"material_type"`
	Background       *string `json:"background"`
	NewToProtocol    *bool   `json:"new_to_protocol"`
	Goals            *string `json:"goals"`
	ConsentEmail     *bool   `json:"consent_email"`
	Status           *string `json:"status"`
	PrepStatus       *string `json:"prep_status"`
	AttendedSessions *int    `json:"attended_sessions"`
	Notes            *string `json:"notes"`
}

var validRegistrationStatuses = map[string]bool{
	"requested": true, "confirmed": true, "waitlist": true, "declined": true,
}

var validPrepStatuses = map[string]bool{
	"not-started": true, "contacted": true, "ready": true, "needs-help": true,
}

func (s *Server) handleAdminUpdateRegistration(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		jsonErr(w, "invalid registration id", http.StatusBadRequest)
		return
	}
	var in registrationUpdate
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 32*1024)).Decode(&in); err != nil {
		jsonErr(w, "invalid update", http.StatusBadRequest)
		return
	}
	var current registrationRow
	var allS, newP, consent int
	var notes sql.NullString
	err = s.DB.QueryRowContext(r.Context(), `
		SELECT name, email, region, all_sessions, material, material_type, background,
		       new_to_protocol, goals, consent_email, status, prep_status,
		       attended_sessions, notes
		FROM event_registrations WHERE id=? AND event_slug=?
	`, id, workshopSlug).Scan(&current.Name, &current.Email, &current.Region, &allS,
		&current.Material, &current.MaterialType, &current.Background, &newP,
		&current.Goals, &consent, &current.Status, &current.PrepStatus,
		&current.AttendedSessions, &notes)
	if err == sql.ErrNoRows {
		jsonErr(w, "registration not found", http.StatusNotFound)
		return
	}
	if err != nil {
		jsonErr(w, "could not load registration", http.StatusInternalServerError)
		return
	}
	current.AllSessions = allS == 1
	current.NewToProtocol = newP == 1
	current.ConsentEmail = consent == 1
	current.Notes = notes.String

	if in.Name != nil {
		current.Name = clip(*in.Name, 200)
	}
	if in.Email != nil {
		current.Email = clip(strings.ToLower(*in.Email), 320)
	}
	if in.Region != nil {
		current.Region = clip(*in.Region, 40)
	}
	if in.AllSessions != nil {
		current.AllSessions = *in.AllSessions
	}
	if in.Material != nil {
		current.Material = clip(*in.Material, 2000)
	}
	if in.MaterialType != nil {
		current.MaterialType = clip(*in.MaterialType, 40)
	}
	if in.Background != nil {
		current.Background = clip(*in.Background, 40)
	}
	if in.NewToProtocol != nil {
		current.NewToProtocol = *in.NewToProtocol
	}
	if in.Goals != nil {
		current.Goals = clip(*in.Goals, 2000)
	}
	if in.ConsentEmail != nil {
		current.ConsentEmail = *in.ConsentEmail
	}
	if in.Status != nil {
		current.Status = *in.Status
	}
	if in.PrepStatus != nil {
		current.PrepStatus = *in.PrepStatus
	}
	if in.AttendedSessions != nil {
		current.AttendedSessions = *in.AttendedSessions
	}
	if in.Notes != nil {
		current.Notes = clip(*in.Notes, 4000)
	}

	if current.Name == "" || !looksLikeEmail(current.Email) || current.Material == "" {
		jsonErr(w, "name, valid email, and material are required", http.StatusBadRequest)
		return
	}
	if !validRegistrationStatuses[current.Status] {
		jsonErr(w, "invalid status", http.StatusBadRequest)
		return
	}
	if !validPrepStatuses[current.PrepStatus] {
		jsonErr(w, "invalid prep status", http.StatusBadRequest)
		return
	}
	if current.AttendedSessions < 0 || current.AttendedSessions > 15 {
		jsonErr(w, "attended_sessions must be a four-bit mask", http.StatusBadRequest)
		return
	}
	res, err := s.DB.ExecContext(r.Context(), `
		UPDATE event_registrations
		SET name=?, email=?, region=?, all_sessions=?, material=?, material_type=?,
		    background=?, new_to_protocol=?, goals=?, consent_email=?, status=?,
		    prep_status=?, attended_sessions=?, notes=?, updated_at=CURRENT_TIMESTAMP
		WHERE id=? AND event_slug=?
	`, current.Name, current.Email, current.Region, b2i(current.AllSessions),
		current.Material, current.MaterialType, current.Background,
		b2i(current.NewToProtocol), current.Goals, b2i(current.ConsentEmail),
		current.Status, current.PrepStatus, current.AttendedSessions, current.Notes,
		id, workshopSlug)
	if err != nil {
		jsonErr(w, "could not save registration", http.StatusInternalServerError)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		jsonErr(w, "registration not found", http.StatusNotFound)
		return
	}
	jsonOK(w, map[string]any{"ok": true})
}

type announcementRow struct {
	ID             int64  `json:"id"`
	Subject        string `json:"subject"`
	RecipientCount int    `json:"recipient_count"`
	SentCount      int    `json:"sent_count"`
	FailedCount    int    `json:"failed_count"`
	CreatedAt      string `json:"created_at"`
}

// registrationMailRow is one outbound_email entry tied to a registration —
// the recipient-level view the announcement history lacks.
type registrationMailRow struct {
	ID             int64  `json:"id"`
	RegistrationID int64  `json:"registration_id"`
	SentAt         string `json:"sent_at"`
	To             string `json:"to"`
	Subject        string `json:"subject"`
	Kind           string `json:"kind"`
	StatusCode     int    `json:"status_code"`
	Error          string `json:"error"`
	Note           string `json:"note"`
}

// queryRegistrationMail returns every logged send addressed to a workshop
// registrant (by ref, or by recipient address for pass/build mail), newest first.
func (s *Server) queryRegistrationMail(r *http.Request) ([]registrationMailRow, error) {
	rows, err := s.DB.QueryContext(r.Context(), `
		SELECT o.id, e.id, o.sent_at, o.to_addrs, o.subject, o.kind, o.status_code, o.error, o.note
		FROM outbound_email o
		JOIN event_registrations e
		  ON (o.ref_type = 'registration' AND o.ref_id = CAST(e.id AS TEXT))
		  OR (o.ref_type <> 'registration' AND o.kind <> 'registration_admin_alert' AND LOWER(o.to_addrs) = LOWER(e.email))
		WHERE e.event_slug = ?
		ORDER BY o.sent_at DESC, o.id DESC LIMIT 500
	`, workshopSlug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []registrationMailRow{}
	for rows.Next() {
		var m registrationMailRow
		if err := rows.Scan(&m.ID, &m.RegistrationID, &m.SentAt, &m.To, &m.Subject, &m.Kind, &m.StatusCode, &m.Error, &m.Note); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Server) queryAnnouncements(r *http.Request) ([]announcementRow, error) {
	rows, err := s.DB.QueryContext(r.Context(), `
		SELECT id, subject, recipient_count, sent_count, failed_count, created_at
		FROM event_announcements WHERE event_slug=?
		ORDER BY created_at DESC LIMIT 20
	`, workshopSlug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []announcementRow{}
	for rows.Next() {
		var a announcementRow
		if err := rows.Scan(&a.ID, &a.Subject, &a.RecipientCount, &a.SentCount, &a.FailedCount, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

type announcementInput struct {
	RegistrationIDs []int64 `json:"registration_ids"`
	Subject         string  `json:"subject"`
	Body            string  `json:"body"`
	// Preview merges the body for each recipient and returns the letters
	// without sending anything or logging a batch.
	Preview bool `json:"preview"`
}

type announcementRecipient struct {
	ID    int64
	Name  string
	Email string
	// Factory Pass state, for merge fields. Code is "" when none is issued;
	// FactoryURL is "" until the code has been redeemed.
	Code       string
	FactoryURL string
}

// announcementMergeFields lists what a message body may contain. Kept in one
// place so the compose UI hint and the merge stay in step.
const announcementMergeFields = "{{first}} {{name}} {{code}} {{factory_url}} {{#code}}…{{/code}} {{#redeemed}}…{{/redeemed}} {{#nocode}}…{{/nocode}}"

var announcementBlockRe = regexp.MustCompile(`(?s)\{\{#(code|redeemed|nocode)\}\}(.*?)\{\{/(code|redeemed|nocode)\}\}`)

// mergeAnnouncement personalizes one message body for one recipient.
// Blocks: {{#code}} shows when an unredeemed code exists, {{#redeemed}} when the
// pass has been redeemed, {{#nocode}} when no code has been issued. Blank lines
// left by removed blocks are collapsed so the letter still reads cleanly.
func mergeAnnouncement(body string, rec announcementRecipient) string {
	hasCode := rec.Code != ""
	redeemed := rec.FactoryURL != ""
	out := announcementBlockRe.ReplaceAllStringFunc(body, func(m string) string {
		sub := announcementBlockRe.FindStringSubmatch(m)
		if sub[1] != sub[3] {
			return m
		}
		keep := false
		switch sub[1] {
		case "code":
			keep = hasCode && !redeemed
		case "redeemed":
			keep = redeemed
		case "nocode":
			keep = !hasCode
		}
		if keep {
			return strings.TrimSpace(sub[2])
		}
		return ""
	})
	// {{code}} / {{factory_url}} are only substituted when the recipient has
	// one; otherwise the tag is left in place for announcementUnmergedRe to
	// catch, so no letter goes out with a blank where the code should be.
	pairs := []string{"{{first}}", firstName(rec.Name), "{{name}}", rec.Name}
	if rec.Code != "" {
		pairs = append(pairs, "{{code}}", rec.Code)
	}
	if rec.FactoryURL != "" {
		pairs = append(pairs, "{{factory_url}}", rec.FactoryURL)
	}
	out = strings.NewReplacer(pairs...).Replace(out)
	out = regexp.MustCompile(`\n{3,}`).ReplaceAllString(out, "\n\n")
	return strings.TrimSpace(out)
}

// announcementUnmergedRe catches a body that still has a bare {{code}} or
// {{factory_url}} after merging (recipient without one) — better to refuse the
// batch than send a letter with a hole in it.
var announcementUnmergedRe = regexp.MustCompile(`\{\{[#/]?[a-z_]+\}\}`)

func (s *Server) handleAdminSendAnnouncement(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	var in announcementInput
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 128*1024)).Decode(&in); err != nil {
		jsonErr(w, "invalid announcement", http.StatusBadRequest)
		return
	}
	if s.Email == nil && !in.Preview {
		jsonErr(w, "email is not configured", http.StatusServiceUnavailable)
		return
	}
	in.Subject = clip(in.Subject, 200)
	in.Body = clip(in.Body, 20000)
	if in.Subject == "" || in.Body == "" {
		jsonErr(w, "subject and message are required", http.StatusBadRequest)
		return
	}
	if len(in.RegistrationIDs) == 0 || len(in.RegistrationIDs) > 50 {
		jsonErr(w, "select between 1 and 50 recipients", http.StatusBadRequest)
		return
	}
	ids := s.withSmokeRegistration(r.Context(), in.RegistrationIDs)

	seen := make(map[int64]bool)
	recipients := make([]announcementRecipient, 0, len(ids))
	for _, id := range ids {
		if id <= 0 || seen[id] {
			continue
		}
		seen[id] = true
		var rec announcementRecipient
		var clientSlug, projectSlug string
		err := s.DB.QueryRowContext(r.Context(), `
			SELECT e.id, e.name, e.email,
			       COALESCE(MAX(c.code), ''), COALESCE(MAX(pr.client_slug), ''), COALESCE(MAX(pr.project_slug), '')
			FROM event_registrations e
			LEFT JOIN coupons c ON c.registration_id = e.id
			LEFT JOIN passes  p ON p.coupon_id = c.id
			LEFT JOIN projects pr ON pr.id = p.project_id
			WHERE e.id=? AND e.event_slug=? AND e.consent_email=1 AND e.status != 'declined'
			GROUP BY e.id
		`, id, workshopSlug).Scan(&rec.ID, &rec.Name, &rec.Email, &rec.Code, &clientSlug, &projectSlug)
		if err == nil {
			if clientSlug != "" && projectSlug != "" {
				rec.FactoryURL = s.portalURL(clientSlug, projectSlug)
			}
			recipients = append(recipients, rec)
		} else if err != sql.ErrNoRows {
			jsonErr(w, "could not load recipients", http.StatusInternalServerError)
			return
		}
	}
	if len(recipients) == 0 {
		jsonErr(w, "none of the selected registrations are eligible for announcement email", http.StatusBadRequest)
		return
	}

	// Merge per recipient up front so a hole in any one letter stops the
	// whole batch before anything is sent.
	merged := make([]string, len(recipients))
	for i, rec := range recipients {
		merged[i] = mergeAnnouncement(in.Body, rec)
		if m := announcementUnmergedRe.FindString(merged[i]); m != "" {
			jsonErr(w, fmt.Sprintf("%s has no value for %s — issue a code first, or wrap that part in {{#code}}…{{/code}}", rec.Name, m), http.StatusBadRequest)
			return
		}
	}

	if in.Preview {
		type letter struct {
			ID    int64  `json:"id"`
			Name  string `json:"name"`
			Email string `json:"email"`
			Text  string `json:"text"`
		}
		letters := make([]letter, len(recipients))
		for i, rec := range recipients {
			letters[i] = letter{rec.ID, rec.Name, rec.Email, announcementText(rec.Name, merged[i])}
		}
		jsonOK(w, map[string]any{"ok": true, "selected": len(recipients), "subject": in.Subject, "letters": letters})
		return
	}

	sent := 0
	failures := []string{}
	sentIDs := []int64{}
	for i, rec := range recipients {
		textBody := announcementText(rec.Name, merged[i])
		htmlBody := announcementHTML(rec.Name, merged[i])
		if err := s.mail(mailMeta{Kind: mailKindAnnouncement, RefType: "registration", RefID: mailRef(rec.ID), TriggeredBy: triggeredBy(r, "admin")}, []string{rec.Email}, nil, in.Subject, textBody, htmlBody); err != nil {
			slog.Error("workshop announcement failed", "err", err, "registration_id", rec.ID)
			failures = append(failures, rec.Name)
			continue
		}
		sent++
		sentIDs = append(sentIDs, rec.ID)
		_, _ = s.DB.ExecContext(r.Context(), `
			UPDATE event_registrations SET last_emailed_at=CURRENT_TIMESTAMP, updated_at=CURRENT_TIMESTAMP
			WHERE id=?
		`, rec.ID)
	}

	idsJSON, _ := json.Marshal(sentIDs)
	_, logErr := s.DB.ExecContext(r.Context(), `
		INSERT INTO event_announcements
			(event_slug, subject, body, recipient_ids, recipient_count, sent_count, failed_count)
		VALUES (?,?,?,?,?,?,?)
	`, workshopSlug, in.Subject, in.Body, string(idsJSON), len(recipients), sent, len(failures))
	if logErr != nil {
		slog.Error("workshop announcement log failed", "err", logErr)
	}
	jsonOK(w, map[string]any{
		"ok":       len(failures) == 0,
		"selected": len(recipients),
		"sent":     sent,
		"failed":   failures,
	})
}

// smokeRegistrationEmail is the permanent smoke-test persona ("Mike Check").
// Every batch send includes this registration so the admin sees exactly what
// registrants received, in a real inbox, without depending on BCC handling.
const smokeRegistrationEmail = "bookiq@gmail.com"

// withSmokeRegistration appends the smoke persona's registration id to a
// batch recipient list when it is not already selected. No-op if the
// persona is not registered (tests, fresh databases).
func (s *Server) withSmokeRegistration(ctx context.Context, ids []int64) []int64 {
	var smokeID int64
	err := s.DB.QueryRowContext(ctx, `
		SELECT id FROM event_registrations
		WHERE event_slug=? AND lower(email)=? AND consent_email=1 AND status != 'declined'
		ORDER BY id LIMIT 1`, workshopSlug, smokeRegistrationEmail).Scan(&smokeID)
	if err != nil {
		return ids
	}
	for _, id := range ids {
		if id == smokeID {
			return ids
		}
	}
	return append(append([]int64{}, ids...), smokeID)
}

func announcementText(name, body string) string {
	return fmt.Sprintf("Hi %s,\n\n%s\n\n— Jenna\njdbb studio\n\nYou’re receiving this workshop announcement because you opted in when registering for Protocolize Your Book. Reply to this email if you’d rather not receive further announcements.", firstName(name), strings.TrimSpace(body))
}

// announcementURLRe matches a bare URL in already-escaped body text so the
// HTML part can link it. Trailing punctuation is left outside the link.
var announcementURLRe = regexp.MustCompile(`https?://[^\s<]+[^\s<.,;:!?)]`)

func announcementHTML(name, body string) string {
	safeBody := html.EscapeString(strings.TrimSpace(body))
	safeBody = strings.ReplaceAll(safeBody, "\r\n", "\n")
	safeBody = strings.ReplaceAll(safeBody, "\n", "<br>")
	safeBody = announcementURLRe.ReplaceAllStringFunc(safeBody, func(u string) string {
		return fmt.Sprintf(`<a href="%s" style="color:%s">%s</a>`, u, emailAccent, u)
	})
	b := emailP(fmt.Sprintf("Hi %s,", html.EscapeString(firstName(name)))) + emailP(safeBody) + emailSignoff()
	return emailShell(b, emailShellOpts{
		Kicker: "Protocolize Your Book",
		Footer: "You&rsquo;re receiving this workshop announcement because you opted in when registering for Protocolize Your Book. Reply if you&rsquo;d rather not receive further announcements.",
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
	b.WriteString("id,created_at,name,email,region,all_sessions,material_type,material,background,new_to_protocol,goals,consent_email,status,prep_status,attended_sessions,last_emailed_at,notes,coupon_code,coupon_redeemed_at\n")
	for _, e := range rows {
		cells := []string{
			strconv.FormatInt(e.ID, 10), e.CreatedAt, e.Name, e.Email, e.Region,
			strconv.FormatBool(e.AllSessions), e.MaterialType, e.Material, e.Background,
			strconv.FormatBool(e.NewToProtocol), e.Goals, strconv.FormatBool(e.ConsentEmail),
			e.Status, e.PrepStatus, strconv.Itoa(e.AttendedSessions), e.LastEmailedAt, e.Notes,
			e.CouponCode, e.CouponRedeemedAt,
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
