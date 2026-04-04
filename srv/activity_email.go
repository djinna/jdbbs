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

	if err := s.Email.sendEmail(to, cc, subject, textBody, htmlBody); err != nil {
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

func activityJournalEmoji(entryType string) string {
	switch entryType {
	case "call":
		return "📞"
	case "decision":
		return "⚖️"
	case "approval":
		return "✅"
	default:
		return "📝"
	}
}

// ─── HTML builder ───

func buildActivityHTML(p activityParams) string {
	var b strings.Builder

	dateRange := fmt.Sprintf("%s – %s",
		time.Now().AddDate(0, 0, -p.Days).Format("Jan 2"),
		time.Now().Format("Jan 2, 2006"))

	b.WriteString(`<!DOCTYPE html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width"></head>`)
	b.WriteString(`<body style="margin:0;padding:0;background:#f4f3f9;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;color:#333;">`)

	// Outer wrapper
	b.WriteString(`<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background:#f4f3f9;"><tr><td align="center" style="padding:24px 12px;">`)
	b.WriteString(`<table role="presentation" width="640" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:12px;overflow:hidden;box-shadow:0 2px 12px rgba(108,99,255,0.08);">`)

	// Header
	b.WriteString(`<tr><td style="background:linear-gradient(135deg,#6c63ff 0%,#8b83ff 100%);padding:32px 36px;">`)
	b.WriteString(fmt.Sprintf(`<h1 style="margin:0 0 4px;font-size:24px;font-weight:700;color:#ffffff;">%s</h1>`, html.EscapeString(p.ProjectName)))
	b.WriteString(fmt.Sprintf(`<p style="margin:0;font-size:13px;color:rgba(255,255,255,0.8);">Activity Update · %s</p>`, html.EscapeString(dateRange)))
	if p.ProjectURL != "" {
		b.WriteString(fmt.Sprintf(`<p style="margin:8px 0 0;font-size:13px;"><a href="%s" style="color:#d4d0ff;text-decoration:underline;">View project online →</a></p>`, p.ProjectURL))
	}
	b.WriteString(`</td></tr>`)

	// Summary bar
	fileCount := len(p.FileLog)
	journalCount := len(p.Journal)
	if fileCount == 0 && journalCount == 0 {
		b.WriteString(`<tr><td style="padding:36px;text-align:center;">`)
		b.WriteString(fmt.Sprintf(`<p style="font-size:15px;color:#888;">No activity in the last %d days.</p>`, p.Days))
		b.WriteString(`</td></tr>`)
	} else {
		// Stat cards
		b.WriteString(`<tr><td style="padding:28px 36px 0;">`)
		b.WriteString(`<table role="presentation" width="100%" cellpadding="0" cellspacing="0"><tr>`)
		b.WriteString(`<td width="50%" style="padding:0 6px 0 0;">`)
		b.WriteString(fmt.Sprintf(`<div style="background:#f0eeff;border-radius:8px;padding:14px 16px;text-align:center;">`+
			`<div style="font-size:28px;font-weight:700;color:#6c63ff;">%d</div>`+
			`<div style="font-size:11px;color:#888;margin-top:2px;">File Transfers</div></div>`, fileCount))
		b.WriteString(`</td>`)
		b.WriteString(`<td width="50%" style="padding:0 0 0 6px;">`)
		b.WriteString(fmt.Sprintf(`<div style="background:#f0eeff;border-radius:8px;padding:14px 16px;text-align:center;">`+
			`<div style="font-size:28px;font-weight:700;color:#6c63ff;">%d</div>`+
			`<div style="font-size:11px;color:#888;margin-top:2px;">Journal Entries</div></div>`, journalCount))
		b.WriteString(`</td>`)
		b.WriteString(`</tr></table>`)
		b.WriteString(`</td></tr>`)

		// File transfers section
		if fileCount > 0 {
			b.WriteString(`<tr><td style="padding:28px 36px 0;">`)
			b.WriteString(`<h2 style="margin:0 0 12px;font-size:16px;font-weight:700;color:#6c63ff;text-transform:uppercase;letter-spacing:0.5px;">File Transfers</h2>`)
			b.WriteString(`<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="border:1px solid #e5e5e5;border-radius:8px;overflow:hidden;">`)

			b.WriteString(`<tr style="background:#6c63ff;">`)
			b.WriteString(`<td style="padding:10px 12px;font-size:11px;font-weight:700;color:#fff;text-transform:uppercase;letter-spacing:0.5px;width:90px;">Date</td>`)
			b.WriteString(`<td style="padding:10px 12px;font-size:11px;font-weight:700;color:#fff;text-transform:uppercase;letter-spacing:0.5px;width:50px;">Dir</td>`)
			b.WriteString(`<td style="padding:10px 12px;font-size:11px;font-weight:700;color:#fff;text-transform:uppercase;letter-spacing:0.5px;">File</td>`)
			b.WriteString(`<td style="padding:10px 12px;font-size:11px;font-weight:700;color:#fff;text-transform:uppercase;letter-spacing:0.5px;width:72px;">Type</td>`)
			b.WriteString(`<td style="padding:10px 12px;font-size:11px;font-weight:700;color:#fff;text-transform:uppercase;letter-spacing:0.5px;width:130px;">From → To</td>`)
			b.WriteString(`</tr>`)

			for i, e := range p.FileLog {
				rowBg := "#ffffff"
				if i%2 == 1 {
					rowBg = "#faf9ff"
				}
				dirArrow := "↓ In"
				if e.Direction == "outbound" {
					dirArrow = "↑ Out"
				}
				fromTo := html.EscapeString(e.SentBy) + " → " + html.EscapeString(e.ReceivedBy)
				b.WriteString(fmt.Sprintf(`<tr style="background:%s;">`, rowBg))
				b.WriteString(fmt.Sprintf(`<td style="padding:9px 12px;border-top:1px solid #eee;font-size:13px;color:#555;">%s</td>`, snapshotFormatDate(e.TransferDate)))
				b.WriteString(fmt.Sprintf(`<td style="padding:9px 12px;border-top:1px solid #eee;font-size:13px;color:#555;">%s</td>`, dirArrow))
				b.WriteString(fmt.Sprintf(`<td style="padding:9px 12px;border-top:1px solid #eee;font-size:13px;font-weight:500;color:#333;">%s</td>`, html.EscapeString(e.Filename)))
				b.WriteString(fmt.Sprintf(`<td style="padding:9px 12px;border-top:1px solid #eee;font-size:13px;color:#555;">%s</td>`, html.EscapeString(e.FileType)))
				b.WriteString(fmt.Sprintf(`<td style="padding:9px 12px;border-top:1px solid #eee;font-size:13px;color:#555;">%s</td>`, fromTo))
				b.WriteString(`</tr>`)
			}

			b.WriteString(`</table>`)
			b.WriteString(`</td></tr>`)
		}

		// Journal entries section
		if journalCount > 0 {
			b.WriteString(`<tr><td style="padding:28px 36px 0;">`)
			b.WriteString(`<h2 style="margin:0 0 12px;font-size:16px;font-weight:700;color:#6c63ff;text-transform:uppercase;letter-spacing:0.5px;">Journal Entries</h2>`)

			for i, e := range p.Journal {
				rowBg := "#ffffff"
				if i%2 == 1 {
					rowBg = "#faf9ff"
				}
				emoji := activityJournalEmoji(e.EntryType)
				dateStr := snapshotFormatDate(e.CreatedAt)
				if t, err := time.Parse("2006-01-02T15:04:05", e.CreatedAt); err == nil {
					dateStr = t.Format("Jan 2, 2006 3:04 PM")
				} else if t, err := time.Parse("2006-01-02 15:04:05", e.CreatedAt); err == nil {
					dateStr = t.Format("Jan 2, 2006 3:04 PM")
				}
				borderRadius := ""
				if i == 0 && i == journalCount-1 {
					borderRadius = "border-radius:8px;"
				} else if i == 0 {
					borderRadius = "border-radius:8px 8px 0 0;"
				} else if i == journalCount-1 {
					borderRadius = "border-radius:0 0 8px 8px;"
				}
				borderBottom := "border-bottom:0;"
				if i == journalCount-1 {
					borderBottom = ""
				}
				b.WriteString(fmt.Sprintf(`<div style="background:%s;padding:12px 16px;border:1px solid #e5e5e5;%s%s">`, rowBg, borderRadius, borderBottom))
				b.WriteString(fmt.Sprintf(`<div style="font-size:12px;color:#888;margin-bottom:4px;">%s %s · %s</div>`,
					emoji, html.EscapeString(strings.ToUpper(e.EntryType)), dateStr))
				b.WriteString(fmt.Sprintf(`<div style="font-size:13px;color:#333;">%s</div>`, html.EscapeString(e.Content)))
				b.WriteString(`</div>`)
			}

			b.WriteString(`</td></tr>`)
		}
	}

	// Footer
	b.WriteString(`<tr><td style="padding:28px 36px;">`)
	b.WriteString(`<div style="border-top:2px solid #f0eeff;padding-top:16px;text-align:center;">`)
	b.WriteString(fmt.Sprintf(`<p style="margin:0 0 6px;font-size:12px;color:#aaa;">Generated %s</p>`, html.EscapeString(p.Generated)))
	if p.ProjectURL != "" {
		b.WriteString(fmt.Sprintf(`<p style="margin:0;"><a href="%s" style="font-size:13px;color:#6c63ff;text-decoration:none;font-weight:600;">View Project Online →</a></p>`, p.ProjectURL))
	}
	b.WriteString(`</div>`)
	b.WriteString(`</td></tr>`)

	b.WriteString(`</table></td></tr></table>`)
	b.WriteString(`</body></html>`)

	return b.String()
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
