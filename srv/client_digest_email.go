package srv

import (
	"encoding/json"
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type clientDigestProject struct {
	Name    string
	FileLog []fileLogEntry
	Journal []journalEntry
}

type clientDigestParams struct {
	ClientName   string
	ClientSlug   string
	BaseURL      string
	Days         int
	Since        string
	Generated    string
	Projects     []clientDigestProject
	TotalFiles   int
	TotalJournal int
}

func (s *Server) handleSendClientDigest(w http.ResponseWriter, r *http.Request) {
	if s.Email == nil {
		jsonErr(w, "email not configured (set AGENTMAIL_API_KEY and AGENTMAIL_INBOX_ID)", 503)
		return
	}

	clientSlug := r.PathValue("client")
	if !s.checkClientAuthOrProjectAuth(w, r, clientSlug) {
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

	now := time.Now()
	days := 7
	since := now.AddDate(0, 0, -days).Format("2006-01-02")

	// Get client name
	var clientName string
	err := s.DB.QueryRowContext(r.Context(),
		`SELECT COALESCE(name, slug) FROM clients WHERE slug = ?`, clientSlug,
	).Scan(&clientName)
	if err != nil {
		clientName = clientSlug
	}

	// Load all projects for this client
	projRows, err := s.DB.QueryContext(r.Context(),
		`SELECT id, name FROM projects WHERE client_slug = ? ORDER BY name`, clientSlug,
	)
	if err != nil {
		jsonErr(w, "query projects: "+err.Error(), 500)
		return
	}
	defer projRows.Close()

	type projInfo struct {
		ID   int64
		Name string
	}
	var allProjs []projInfo
	for projRows.Next() {
		var p projInfo
		if err := projRows.Scan(&p.ID, &p.Name); err != nil {
			jsonErr(w, "scan projects: "+err.Error(), 500)
			return
		}
		allProjs = append(allProjs, p)
	}

	var projects []clientDigestProject
	totalFiles := 0
	totalJournal := 0

	for _, proj := range allProjs {
		// File log for this project since cutoff
		fileRows, err := s.DB.QueryContext(r.Context(),
			`SELECT id, project_id, direction, filename, file_type, sent_by, received_by, notes, transfer_date, created_at
			 FROM file_log WHERE project_id = ? AND transfer_date >= ? ORDER BY transfer_date DESC, created_at DESC`, proj.ID, since,
		)
		if err != nil {
			continue
		}
		var files []fileLogEntry
		for fileRows.Next() {
			var e fileLogEntry
			if err := fileRows.Scan(&e.ID, &e.ProjectID, &e.Direction, &e.Filename, &e.FileType, &e.SentBy, &e.ReceivedBy, &e.Notes, &e.TransferDate, &e.CreatedAt); err != nil {
				continue
			}
			files = append(files, e)
		}
		fileRows.Close()

		// Journal for this project since cutoff
		journalRows, err := s.DB.QueryContext(r.Context(),
			`SELECT id, project_id, entry_type, content, created_at
			 FROM journal WHERE project_id = ? AND created_at >= ? ORDER BY created_at DESC`, proj.ID, since,
		)
		if err != nil {
			continue
		}
		var journal []journalEntry
		for journalRows.Next() {
			var e journalEntry
			if err := journalRows.Scan(&e.ID, &e.ProjectID, &e.EntryType, &e.Content, &e.CreatedAt); err != nil {
				continue
			}
			journal = append(journal, e)
		}
		journalRows.Close()

		// Only include projects with activity
		if len(files) > 0 || len(journal) > 0 {
			projects = append(projects, clientDigestProject{
				Name:    proj.Name,
				FileLog: files,
				Journal: journal,
			})
			totalFiles += len(files)
			totalJournal += len(journal)
		}
	}

	dp := clientDigestParams{
		ClientName:   clientName,
		ClientSlug:   clientSlug,
		BaseURL:      s.BaseURL,
		Days:         days,
		Since:        since,
		Generated:    now.Format("January 2, 2006 at 3:04 PM MST"),
		Projects:     projects,
		TotalFiles:   totalFiles,
		TotalJournal: totalJournal,
	}

	dateRange := fmt.Sprintf("%s \u2013 %s", now.AddDate(0, 0, -days).Format("Jan 2"), now.Format("Jan 2, 2006"))
	subject := fmt.Sprintf("Weekly Digest: %s \u2014 %s", clientName, dateRange)

	htmlBody := buildClientDigestHTML(dp)
	textBody := buildClientDigestText(dp)

	to := []string{body.Recipients[0]}
	var cc []string
	if len(body.Recipients) > 1 {
		cc = body.Recipients[1:]
	}

	if err := s.Email.sendEmail(to, cc, subject, textBody, htmlBody); err != nil {
		slog.Error("send client digest email", "error", err)
		jsonErr(w, "failed to send email: "+err.Error(), 500)
		return
	}

	jsonOK(w, map[string]any{
		"ok":      true,
		"sent_to": body.Recipients,
		"subject": subject,
	})
}

func buildClientDigestHTML(p clientDigestParams) string {
	var b strings.Builder

	dateRange := fmt.Sprintf("%s \u2013 %s",
		time.Now().AddDate(0, 0, -p.Days).Format("Jan 2"),
		time.Now().Format("Jan 2, 2006"))

	b.WriteString(`<!DOCTYPE html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width"></head>`)
	b.WriteString(`<body style="margin:0;padding:0;background:#f4f3f9;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;color:#333;">`)

	b.WriteString(`<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background:#f4f3f9;"><tr><td align="center" style="padding:24px 12px;">`)
	b.WriteString(`<table role="presentation" width="640" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:12px;overflow:hidden;box-shadow:0 2px 12px rgba(108,99,255,0.08);">`)

	// Header
	b.WriteString(`<tr><td style="background:linear-gradient(135deg,#6c63ff 0%,#8b83ff 100%);padding:32px 36px;">`)
	b.WriteString(fmt.Sprintf(`<h1 style="margin:0 0 4px;font-size:24px;font-weight:700;color:#ffffff;">%s</h1>`, html.EscapeString(p.ClientName)))
	b.WriteString(fmt.Sprintf(`<p style="margin:0;font-size:13px;color:rgba(255,255,255,0.8);">Weekly Digest \u00b7 %s</p>`, html.EscapeString(dateRange)))
	b.WriteString(`</td></tr>`)

	// Summary bar
	projCount := len(p.Projects)
	if projCount == 0 {
		b.WriteString(`<tr><td style="padding:36px;text-align:center;">`)
		b.WriteString(fmt.Sprintf(`<p style="font-size:15px;color:#888;">No activity across any projects in the last %d days.</p>`, p.Days))
		b.WriteString(`</td></tr>`)
	} else {
		// Stat cards
		b.WriteString(`<tr><td style="padding:28px 36px 0;">`)
		b.WriteString(`<table role="presentation" width="100%" cellpadding="0" cellspacing="0"><tr>`)

		b.WriteString(`<td width="33%" style="padding:0 6px 0 0;">`)
		b.WriteString(fmt.Sprintf(`<div style="background:#f0eeff;border-radius:8px;padding:14px 16px;text-align:center;">`+
			`<div style="font-size:28px;font-weight:700;color:#6c63ff;">%d</div>`+
			`<div style="font-size:11px;color:#888;margin-top:2px;">Active Projects</div></div>`, projCount))
		b.WriteString(`</td>`)

		b.WriteString(`<td width="33%" style="padding:0 6px;">`)
		b.WriteString(fmt.Sprintf(`<div style="background:#f0eeff;border-radius:8px;padding:14px 16px;text-align:center;">`+
			`<div style="font-size:28px;font-weight:700;color:#6c63ff;">%d</div>`+
			`<div style="font-size:11px;color:#888;margin-top:2px;">File Transfers</div></div>`, p.TotalFiles))
		b.WriteString(`</td>`)

		b.WriteString(`<td width="33%" style="padding:0 0 0 6px;">`)
		b.WriteString(fmt.Sprintf(`<div style="background:#f0eeff;border-radius:8px;padding:14px 16px;text-align:center;">`+
			`<div style="font-size:28px;font-weight:700;color:#6c63ff;">%d</div>`+
			`<div style="font-size:11px;color:#888;margin-top:2px;">Journal Entries</div></div>`, p.TotalJournal))
		b.WriteString(`</td>`)

		b.WriteString(`</tr></table>`)
		b.WriteString(`</td></tr>`)

		// Per-project sections
		for _, proj := range p.Projects {
			b.WriteString(`<tr><td style="padding:28px 36px 0;">`)
			b.WriteString(fmt.Sprintf(`<h2 style="margin:0 0 12px;font-size:16px;font-weight:700;color:#6c63ff;letter-spacing:0.5px;">%s</h2>`, html.EscapeString(proj.Name)))

			// File transfers for this project
			if len(proj.FileLog) > 0 {
				b.WriteString(fmt.Sprintf(`<p style="margin:0 0 8px;font-size:12px;font-weight:600;color:#888;text-transform:uppercase;letter-spacing:0.5px;">File Transfers (%d)</p>`, len(proj.FileLog)))
				b.WriteString(`<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="border:1px solid #e5e5e5;border-radius:8px;overflow:hidden;margin-bottom:12px;">`)

				b.WriteString(`<tr style="background:#6c63ff;">`)
				b.WriteString(`<td style="padding:8px 10px;font-size:11px;font-weight:700;color:#fff;text-transform:uppercase;width:80px;">Date</td>`)
				b.WriteString(`<td style="padding:8px 10px;font-size:11px;font-weight:700;color:#fff;text-transform:uppercase;width:40px;">Dir</td>`)
				b.WriteString(`<td style="padding:8px 10px;font-size:11px;font-weight:700;color:#fff;text-transform:uppercase;">File</td>`)
				b.WriteString(`<td style="padding:8px 10px;font-size:11px;font-weight:700;color:#fff;text-transform:uppercase;width:60px;">Type</td>`)
				b.WriteString(`</tr>`)

				for i, e := range proj.FileLog {
					rowBg := "#ffffff"
					if i%2 == 1 {
						rowBg = "#faf9ff"
					}
					dirArrow := "\u2193 In"
					if e.Direction == "outbound" {
						dirArrow = "\u2191 Out"
					}
					b.WriteString(fmt.Sprintf(`<tr style="background:%s;">`, rowBg))
					b.WriteString(fmt.Sprintf(`<td style="padding:7px 10px;border-top:1px solid #eee;font-size:12px;color:#555;">%s</td>`, snapshotFormatDate(e.TransferDate)))
					b.WriteString(fmt.Sprintf(`<td style="padding:7px 10px;border-top:1px solid #eee;font-size:12px;color:#555;">%s</td>`, dirArrow))
					b.WriteString(fmt.Sprintf(`<td style="padding:7px 10px;border-top:1px solid #eee;font-size:12px;font-weight:500;color:#333;">%s</td>`, html.EscapeString(e.Filename)))
					b.WriteString(fmt.Sprintf(`<td style="padding:7px 10px;border-top:1px solid #eee;font-size:12px;color:#555;">%s</td>`, html.EscapeString(e.FileType)))
					b.WriteString(`</tr>`)
				}

				b.WriteString(`</table>`)
			}

			// Journal entries for this project
			if len(proj.Journal) > 0 {
				b.WriteString(fmt.Sprintf(`<p style="margin:0 0 8px;font-size:12px;font-weight:600;color:#888;text-transform:uppercase;letter-spacing:0.5px;">Journal Entries (%d)</p>`, len(proj.Journal)))

				for i, e := range proj.Journal {
					rowBg := "#ffffff"
					if i%2 == 1 {
						rowBg = "#faf9ff"
					}
					emoji := activityJournalEmoji(e.EntryType)
					dateStr := snapshotFormatDate(e.CreatedAt)
					if t, err := time.Parse("2006-01-02T15:04:05", e.CreatedAt); err == nil {
						dateStr = t.Format("Jan 2, 3:04 PM")
					} else if t, err := time.Parse("2006-01-02 15:04:05", e.CreatedAt); err == nil {
						dateStr = t.Format("Jan 2, 3:04 PM")
					}

					cnt := len(proj.Journal)
					borderRadius := ""
					if i == 0 && i == cnt-1 {
						borderRadius = "border-radius:8px;"
					} else if i == 0 {
						borderRadius = "border-radius:8px 8px 0 0;"
					} else if i == cnt-1 {
						borderRadius = "border-radius:0 0 8px 8px;"
					}
					borderBottom := "border-bottom:0;"
					if i == cnt-1 {
						borderBottom = ""
					}

					b.WriteString(fmt.Sprintf(`<div style="background:%s;padding:10px 14px;border:1px solid #e5e5e5;%s%s">`, rowBg, borderRadius, borderBottom))
					b.WriteString(fmt.Sprintf(`<div style="font-size:11px;color:#888;margin-bottom:3px;">%s %s \u00b7 %s</div>`,
						emoji, html.EscapeString(strings.ToUpper(e.EntryType)), dateStr))
					b.WriteString(fmt.Sprintf(`<div style="font-size:13px;color:#333;">%s</div>`, html.EscapeString(e.Content)))
					b.WriteString(`</div>`)
				}
			}

			b.WriteString(`</td></tr>`)
		}
	}

	// Footer
	b.WriteString(`<tr><td style="padding:28px 36px;">`)
	b.WriteString(`<div style="border-top:2px solid #f0eeff;padding-top:16px;text-align:center;">`)
	b.WriteString(fmt.Sprintf(`<p style="margin:0 0 6px;font-size:12px;color:#aaa;">Generated %s</p>`, html.EscapeString(p.Generated)))
	clientURL := fmt.Sprintf("%s/%s/", p.BaseURL, p.ClientSlug)
	b.WriteString(fmt.Sprintf(`<p style="margin:0;"><a href="%s" style="font-size:13px;color:#6c63ff;text-decoration:none;font-weight:600;">View Client Portal \u2192</a></p>`, clientURL))
	b.WriteString(`</div>`)
	b.WriteString(`</td></tr>`)

	b.WriteString(`</table></td></tr></table>`)
	b.WriteString(`</body></html>`)

	return b.String()
}

func buildClientDigestText(p clientDigestParams) string {
	var b strings.Builder

	dateRange := fmt.Sprintf("%s \u2013 %s",
		time.Now().AddDate(0, 0, -p.Days).Format("Jan 2"),
		time.Now().Format("Jan 2, 2006"))

	b.WriteString(fmt.Sprintf("WEEKLY DIGEST: %s\n", strings.ToUpper(p.ClientName)))
	b.WriteString(strings.Repeat("=", 50) + "\n")
	b.WriteString(fmt.Sprintf("Period: %s\n", dateRange))
	b.WriteString(fmt.Sprintf("Summary: %d file transfers, %d journal entries across %d projects\n\n",
		p.TotalFiles, p.TotalJournal, len(p.Projects)))

	if len(p.Projects) == 0 {
		b.WriteString(fmt.Sprintf("No activity in the last %d days.\n\n", p.Days))
	} else {
		for _, proj := range p.Projects {
			b.WriteString(fmt.Sprintf("%s\n", strings.ToUpper(proj.Name)))
			b.WriteString(strings.Repeat("\u2500", 35) + "\n")

			if len(proj.FileLog) > 0 {
				b.WriteString(fmt.Sprintf("  File Transfers (%d):\n", len(proj.FileLog)))
				for _, e := range proj.FileLog {
					dir := "\u2193 In "
					if e.Direction == "outbound" {
						dir = "\u2191 Out"
					}
					date := e.TransferDate
					if date == "" {
						date = "\u2014"
					}
					b.WriteString(fmt.Sprintf("    %s  %s  %s  %s\n", date, dir, e.Filename, e.FileType))
				}
			}

			if len(proj.Journal) > 0 {
				b.WriteString(fmt.Sprintf("  Journal Entries (%d):\n", len(proj.Journal)))
				for _, e := range proj.Journal {
					entryType := strings.ToUpper(e.EntryType)
					dateStr := e.CreatedAt
					if len(dateStr) > 19 {
						dateStr = dateStr[:19]
					}
					b.WriteString(fmt.Sprintf("    [%s] %s  %s\n", entryType, dateStr, e.Content))
				}
			}

			b.WriteString("\n")
		}
	}

	b.WriteString(strings.Repeat("\u2500", 50) + "\n")
	b.WriteString(fmt.Sprintf("Sent: %s\n", p.Generated))
	clientURL := fmt.Sprintf("%s/%s/", p.BaseURL, p.ClientSlug)
	b.WriteString(fmt.Sprintf("View portal: %s\n", clientURL))

	return b.String()
}
