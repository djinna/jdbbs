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
		jsonErr(w, "email not configured (set PRODCAL_MAIL_FROM for Resend, or AGENTMAIL_API_KEY + AGENTMAIL_INBOX_ID)", 503)
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

	if !enforceEmailSendLimits(w, "client:"+clientSlug, body.Recipients) {
		return
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

	// The digest is the recurring pathway: advertise an unsubscribe route
	// (RFC 2369) via the provider's generic headers map (Resend and AgentMail both take one).
	unsubAddr := s.Email.ReplyTo
	if unsubAddr == "" {
		unsubAddr = s.Email.InboxID
	}
	unsubHeaders := map[string]string{
		"List-Unsubscribe": fmt.Sprintf("<mailto:%s?subject=unsubscribe>", unsubAddr),
	}

	if err := s.mail(mailMeta{Kind: mailKindClientDigest, RefType: "client", RefID: clientSlug, TriggeredBy: triggeredBy(r, "client"), Headers: unsubHeaders}, to, cc, subject, textBody, htmlBody); err != nil {
		slog.Error("send client digest email", "error", err)
		jsonErr(w, "email send failed", 500)
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

	projCount := len(p.Projects)
	if projCount == 0 {
		b.WriteString(emailP(fmt.Sprintf("No activity across any projects in the last %d days.", p.Days)))
	} else {
		b.WriteString(emailStats([][2]string{
			{"Active projects", fmt.Sprintf("%d", projCount)},
			{"File transfers", fmt.Sprintf("%d", p.TotalFiles)},
			{"Journal entries", fmt.Sprintf("%d", p.TotalJournal)},
		}))

		for _, proj := range p.Projects {
			b.WriteString(emailH2(proj.Name))

			if len(proj.FileLog) > 0 {
				b.WriteString(emailSmall(fmt.Sprintf("File transfers (%d)", len(proj.FileLog))))
				b.WriteString(emailTable([]string{"Date", "Dir", "File", "Type"}, fileLogTableRows(proj.FileLog, false), nil))
			}

			if len(proj.Journal) > 0 {
				b.WriteString(emailSmall(fmt.Sprintf("Journal entries (%d)", len(proj.Journal))))
				b.WriteString(emailTable([]string{"When", "Type", "Entry"}, journalTableRows(proj.Journal), nil))
			}
		}
	}

	clientURL := fmt.Sprintf("%s/%s/", p.BaseURL, p.ClientSlug)
	b.WriteString(emailButton(clientURL, "View client portal"))

	return emailShell(b.String(), emailShellOpts{
		Kicker: "Client digest \u00b7 " + dateRange,
		Title:  p.ClientName,
		Footer: fmt.Sprintf("Generated %s &middot; Reply to this email to stop receiving digests.", html.EscapeString(p.Generated)),
	})
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
	b.WriteString("Reply to this email to stop receiving digests.\n")

	return b.String()
}
