package srv

import (
	"database/sql"
	"net/http"
	"strings"
)

// Cohort roster — the attendee-facing sibling of the admin cohort tracker.
//
// Visibility tier: client-visible + cohort flag
// (docs/PAGE-DESIGN-HOSTING-VISIBILITY-2026-09-11.md §7). A visitor sees the
// roster only if they hold a valid client cookie for a client whose
// clients.cohort_slug equals the roster's cohort. exe.dev admin bypasses.
//
// What the roster deliberately does NOT expose: emails, admin notes,
// status/prep fields, IP/user-agent, Factory Pass codes, CSV, editing, and
// any email functionality. Full names are shown (decision 2026-09-11).

// cohortForClient returns the cohort_slug for a client, "" if none.
func (s *Server) cohortForClient(r *http.Request, clientSlug string) string {
	var cohort string
	_ = s.DB.QueryRowContext(r.Context(),
		`SELECT cohort_slug FROM clients WHERE slug = ?`, clientSlug).Scan(&cohort)
	return cohort
}

// cohortClientFromRequest finds the first client cookie on the request that
// (a) validates and (b) belongs to the given cohort. Returns the slug or "".
func (s *Server) cohortClientFromRequest(r *http.Request, cohort string) string {
	for _, c := range r.Cookies() {
		if !strings.HasPrefix(c.Name, "prodcal_client_") {
			continue
		}
		slug := strings.TrimPrefix(c.Name, "prodcal_client_")
		if s.checkClientAuth(r, slug) && s.cohortForClient(r, slug) == cohort {
			return slug
		}
	}
	return ""
}

// requireCohort gates a request on cohort membership. Admin bypasses.
func (s *Server) requireCohort(w http.ResponseWriter, r *http.Request, cohort string) bool {
	if r.Header.Get("X-ExeDev-UserID") != "" {
		return true
	}
	if s.cohortClientFromRequest(r, cohort) != "" {
		return true
	}
	jsonErr(w, "unauthorized", http.StatusUnauthorized)
	return false
}

// handleCohortPage serves the roster shell. The page itself is gated in the
// browser by the JSON endpoint (an unauthenticated visitor sees a sign-in
// prompt pointing at the client portal), so the HTML carries no roster data.
func (s *Server) handleCohortPage(w http.ResponseWriter, r *http.Request) {
	if r.PathValue("cohort") != workshopSlug {
		http.NotFound(w, r)
		return
	}
	s.serveStaticHTML(w, "static/cohort.html")
}

type cohortMember struct {
	Name             string `json:"name"`
	Region           string `json:"region"`
	AllSessions      bool   `json:"all_sessions"`
	Material         string `json:"material"`
	MaterialType     string `json:"material_type"`
	Background       string `json:"background"`
	Goals            string `json:"goals"`
	AttendedSessions int    `json:"attended_sessions"`
	ClientSlug       string `json:"client_slug"`
	IsYou            bool   `json:"is_you"`
}

// handleCohortRoster returns the confirmed/requested participants of the
// cohort, minus everything private. Declined registrants are omitted.
func (s *Server) handleCohortRoster(w http.ResponseWriter, r *http.Request) {
	cohort := r.PathValue("cohort")
	if cohort != workshopSlug {
		http.NotFound(w, r)
		return
	}
	if !s.requireCohort(w, r, cohort) {
		return
	}
	you := s.cohortClientFromRequest(r, cohort)
	rows, err := s.DB.QueryContext(r.Context(), `
		SELECT e.name, e.region, e.all_sessions, e.material, e.material_type,
		       e.background, e.goals, e.attended_sessions,
		       COALESCE(MAX(pr.client_slug), '') AS client_slug
		FROM event_registrations e
		LEFT JOIN coupons c ON c.registration_id = e.id
		LEFT JOIN passes  p ON p.coupon_id = c.id
		LEFT JOIN projects pr ON pr.id = p.project_id
		WHERE e.event_slug = ? AND e.status IN ('requested','confirmed')
		GROUP BY e.id
		ORDER BY e.name COLLATE NOCASE ASC
	`, cohort)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	defer rows.Close()
	out := []cohortMember{}
	for rows.Next() {
		var m cohortMember
		var allS int
		var goals sql.NullString
		if err := rows.Scan(&m.Name, &m.Region, &allS, &m.Material, &m.MaterialType,
			&m.Background, &goals, &m.AttendedSessions, &m.ClientSlug); err != nil {
			jsonErr(w, err.Error(), 500)
			return
		}
		m.AllSessions = allS == 1
		m.Goals = goals.String
		m.IsYou = you != "" && m.ClientSlug == you
		out = append(out, m)
	}
	jsonOK(w, map[string]any{
		"cohort":  cohort,
		"title":   "Protocolize Your Book",
		"dates":   "September 21–22, 2026",
		"members": out,
		"count":   len(out),
		"you":     you,
	})
}
