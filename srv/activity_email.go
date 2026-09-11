package srv

import (
	"encoding/json"
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (s *Server) handleSendActivityEmail(w http.ResponseWriter, r *http.Request) {
	if s.Email == nil {
		jsonErr(w, "email not configured (set AGENTMAIL_API_KEY and AGENTMAIL_INBOX_ID)", 503)
		return
	}

	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	if !s.requireAuth(w, r, pid) {
		return
	}

	var body struct {
		Recipients []string `json:"recipients"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, "bad request", 400)
		return
	}
	if len(body.Recipients) == 0 {
		jsonErr(w, "at least one recipient required", 400)
		return
	}
	for _, addr := range body.Recipients {
		if !strings.Contains(addr, "@") || !strings.Contains(addr, ".") {
			jsonErr(w, fmt.Sprintf("invalid email: %s", addr), 400)
			return
		}
	}

	if !enforceEmailSendLimits(w, fmt.Sprintf("project:%d", pid), body.Recipients) {
		return
	}

	// Days parameter
	days := 7
	if d, err := strconv.Atoi(r.URL.Query().Get("days")); err == nil && d > 0 && d <= 90 {
		days = d
	}

	// Load project
	var projName, clientSlug, projectSlug string
	err = s.DB.QueryRowContext(r.Context(),
		`SELECT name, client_slug, project_slug FROM projects WHERE id = ?`, pid,
	).Scan(&projName, &clientSlug, &projectSlug)
	if err != nil {
		jsonErr(w, "project not found", 404)
		return
	}
	projectURL := fmt.Sprintf("%s/%s/%s/", s.BaseURL, clientSlug, projectSlug)

	now := time.Now()
	since := now.AddDate(0, 0, -days).Format("2006-01-02")

	// Load file log entries since cutoff
	fileRows, err := s.DB.QueryContext(r.Context(),
		`SELECT id, project_id, direction, filename, file_type, sent_by, received_by, notes, transfer_date, created_at
		 FROM file_log WHERE project_id = ? AND transfer_date >= ? ORDER BY transfer_date DESC, created_at DESC`, pid, since,
	)
	if err != nil {
		jsonErr(w, "query file log: "+err.Error(), 500)
		return
	}
	defer fileRows.Close()
	var files []fileLogEntry
	for fileRows.Next() {
		var e fileLogEntry
		if err := fileRows.Scan(&e.ID, &e.ProjectID, &e.Direction, &e.Filename, &e.FileType, &e.SentBy, &e.ReceivedBy, &e.Notes, &e.TransferDate, &e.CreatedAt); err != nil {
			jsonErr(w, "scan file log: "+err.Error(), 500)
			return
		}
		files = append(files, e)
	}

	// Load journal entries since cutoff
	journalRows, err := s.DB.QueryContext(r.Context(),
		`SELECT id, project_id, entry_type, content, created_at
		 FROM journal WHERE project_id = ? AND created_at >= ? ORDER BY created_at DESC`, pid, since,
	)
	if err != nil {
		jsonErr(w, "query journal: "+err.Error(), 500)
		return
	}
	defer journalRows.Close()
	var journal []journalEntry
	for journalRows.Next() {
		var e journalEntry
		if err := journalRows.Scan(&e.ID, &e.ProjectID, &e.EntryType, &e.Content, &e.CreatedAt); err != nil {
			jsonErr(w, "scan journal: "+err.Error(), 500)
			return
		}
		journal = append(journal, e)
	}

	ap := activityParams{
		ProjectName: projName,
		ProjectURL:  projectURL,
		Days:        days,
		Since:       since,
		Generated:   now.Format("January 2, 2006 at 3:04 PM MST"),
		FileLog:     files,
		Journal:     journal,
	}

	dateRange := fmt.Sprintf("%s – %s", time.Now().AddDate(0, 0, -days).Format("Jan 2"), now.Format("Jan 2, 2006"))
	subject := fmt.Sprintf("Activity Update: %s — %s", projName, dateRange)

	htmlBody := buildActivityHTML(ap)
	textBody := buildActivityText(ap)

	to := []string{body.Recipients[0]}
	var cc []string
	if len(body.Recipients) > 1 {
		cc = body.Recipients[1:]
	}

	if err := s.mail(mailMeta{Kind: mailKindActivity, RefType: "project", RefID: mailRef(pid), TriggeredBy: triggeredBy(r, "client")}, to, cc, subject, textBody, htmlBody); err != nil {
		slog.Error("send activity email", "error", err)
		jsonErr(w, "failed to send email: "+err.Error(), 500)
		return
	}

	jsonOK(w, map[string]any{
		"ok":      true,
		"sent_to": body.Recipients,
		"subject": subject,
	})
}

type activityParams struct {
	ProjectName string
	ProjectURL  string
	Days        int
	Since       string
	Generated   string
	FileLog     []fileLogEntry
	Journal     []journalEntry
}

// activityJournalLabel returns the journal entry type as a plain word for
// display (mono uppercase via emailStatus). Empty types read as "note".
func activityJournalLabel(entryType string) string {
	if strings.TrimSpace(entryType) == "" {
		return "note"
	}
	return entryType
}

// ─── HTML builder ───

func buildActivityHTML(p activityParams) string {
	var b strings.Builder

	dateRange := fmt.Sprintf("%s \u2013 %s",
		time.Now().AddDate(0, 0, -p.Days).Format("Jan 2"),
		time.Now().Format("Jan 2, 2006"))

	fileCount := len(p.FileLog)
	journalCount := len(p.Journal)
	if fileCount == 0 && journalCount == 0 {
		b.WriteString(emailP(fmt.Sprintf("No activity in the last %d days.", p.Days)))
	} else {
		b.WriteString(emailStats([][2]string{
			{"File transfers", fmt.Sprintf("%d", fileCount)},
			{"Journal entries", fmt.Sprintf("%d", journalCount)},
		}))

		if fileCount > 0 {
			b.WriteString(emailH2("File transfers"))
			b.WriteString(emailTable([]string{"Date", "Dir", "File", "Type", "From \u2192 To"}, fileLogTableRows(p.FileLog, true), nil))
		}

		if journalCount > 0 {
			b.WriteString(emailH2("Journal entries"))
			b.WriteString(emailTable([]string{"When", "Type", "Entry"}, journalTableRows(p.Journal), nil))
		}
	}

	if p.ProjectURL != "" {
		b.WriteString(emailButton(p.ProjectURL, "View project online"))
	}

	footer := "Reply to this email to reach Jenna."
	if p.Generated != "" {
		footer = fmt.Sprintf("Generated %s &middot; %s", html.EscapeString(p.Generated), footer)
	}

	return emailShell(b.String(), emailShellOpts{
		Kicker: "Activity update \u00b7 " + dateRange,
		Title:  p.ProjectName,
		Footer: footer,
	})
}

// ─── Plain text builder ───

func buildActivityText(p activityParams) string {
	var b strings.Builder

	dateRange := fmt.Sprintf("%s – %s",
		time.Now().AddDate(0, 0, -p.Days).Format("Jan 2"),
		time.Now().Format("Jan 2, 2006"))

	b.WriteString(fmt.Sprintf("ACTIVITY UPDATE: %s\n", strings.ToUpper(p.ProjectName)))
	b.WriteString(strings.Repeat("=", 50) + "\n")
	b.WriteString(fmt.Sprintf("Period: %s (%d days)\n", dateRange, p.Days))
	if p.ProjectURL != "" {
		b.WriteString(fmt.Sprintf("Online: %s\n", p.ProjectURL))
	}
	b.WriteString("\n")

	if len(p.FileLog) == 0 && len(p.Journal) == 0 {
		b.WriteString(fmt.Sprintf("No activity in the last %d days.\n\n", p.Days))
	} else {
		b.WriteString(fmt.Sprintf("Summary: %d file transfers, %d journal entries\n\n", len(p.FileLog), len(p.Journal)))

		if len(p.FileLog) > 0 {
			b.WriteString("FILE TRANSFERS\n")
			b.WriteString(strings.Repeat("─", 35) + "\n")
			for _, e := range p.FileLog {
				dir := "↓ In "
				if e.Direction == "outbound" {
					dir = "↑ Out"
				}
				date := e.TransferDate
				if date == "" {
					date = "—"
				}
				b.WriteString(fmt.Sprintf("  %s  %s  %-28s  %-10s  %s → %s\n",
					date, dir, e.Filename, e.FileType, e.SentBy, e.ReceivedBy))
			}
			b.WriteString("\n")
		}

		if len(p.Journal) > 0 {
			b.WriteString("JOURNAL ENTRIES\n")
			b.WriteString(strings.Repeat("─", 35) + "\n")
			for _, e := range p.Journal {
				entryType := strings.ToUpper(e.EntryType)
				dateStr := e.CreatedAt
				if len(dateStr) > 19 {
					dateStr = dateStr[:19]
				}
				b.WriteString(fmt.Sprintf("  [%s] %s  %s\n", entryType, dateStr, e.Content))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString(strings.Repeat("─", 50) + "\n")
	b.WriteString(fmt.Sprintf("Sent: %s\n", p.Generated))
	if p.ProjectURL != "" {
		b.WriteString(fmt.Sprintf("View online: %s\n", p.ProjectURL))
	}

	return b.String()
}
