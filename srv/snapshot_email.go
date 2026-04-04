package srv

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

func (s *Server) handleSendProjectSnapshot(w http.ResponseWriter, r *http.Request) {
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

	// ── Load project ──
	var projID int64
	var projName, startDate, clientSlug, projectSlug string
	err = s.DB.QueryRowContext(r.Context(),
		`SELECT id, name, start_date, client_slug, project_slug FROM projects WHERE id = ?`, pid,
	).Scan(&projID, &projName, &startDate, &clientSlug, &projectSlug)
	if err != nil {
		jsonErr(w, "project not found", 404)
		return
	}
	projectURL := fmt.Sprintf("%s/%s/%s/", s.BaseURL, clientSlug, projectSlug)

	// ── Load tasks ──
	rows, err := s.DB.QueryContext(r.Context(),
		`SELECT sort_order, assignee, title, is_milestone, status, curr_due, actual_done,
		        orig_budget, curr_budget, actual_budget
		 FROM tasks WHERE project_id = ? ORDER BY sort_order`, pid,
	)
	if err != nil {
		jsonErr(w, "query tasks: "+err.Error(), 500)
		return
	}
	defer rows.Close()

	var tasks []snapshotTask
	for rows.Next() {
		var t snapshotTask
		if err := rows.Scan(&t.SortOrder, &t.Assignee, &t.Title, &t.IsMilestone,
			&t.Status, &t.CurrDue, &t.ActualDone,
			&t.OrigBudget, &t.CurrBudget, &t.ActualBudget); err != nil {
			jsonErr(w, "scan tasks: "+err.Error(), 500)
			return
		}
		tasks = append(tasks, t)
	}

	// ── Load transmittal (optional) ──
	var txStatus, txDataStr string
	var hasTx bool
	err = s.DB.QueryRowContext(r.Context(),
		`SELECT status, data FROM transmittals WHERE project_id = ?`, pid,
	).Scan(&txStatus, &txDataStr)
	if err == nil {
		hasTx = true
	} else if err != sql.ErrNoRows {
		slog.Warn("snapshot: load transmittal", "error", err)
	}

	var txData transmittalEmailData
	if hasTx {
		_ = json.Unmarshal([]byte(txDataStr), &txData)
	}

	// ── Load recent file log (last 10) ──
	fileRows, err := s.DB.QueryContext(r.Context(),
		`SELECT id, project_id, direction, filename, file_type, sent_by, received_by, notes, transfer_date, created_at
		 FROM file_log WHERE project_id = ? ORDER BY transfer_date DESC, created_at DESC LIMIT 10`, pid,
	)
	if err != nil {
		jsonErr(w, "query file log: "+err.Error(), 500)
		return
	}
	defer fileRows.Close()
	var fileLogEntries []fileLogEntry
	for fileRows.Next() {
		var e fileLogEntry
		if err := fileRows.Scan(&e.ID, &e.ProjectID, &e.Direction, &e.Filename, &e.FileType, &e.SentBy, &e.ReceivedBy, &e.Notes, &e.TransferDate, &e.CreatedAt); err != nil {
			jsonErr(w, "scan file log: "+err.Error(), 500)
			return
		}
		fileLogEntries = append(fileLogEntries, e)
	}

	// ── Load recent journal (last 10) ──
	journalRows, err := s.DB.QueryContext(r.Context(),
		`SELECT id, project_id, entry_type, content, created_at
		 FROM journal WHERE project_id = ? ORDER BY created_at DESC LIMIT 10`, pid,
	)
	if err != nil {
		jsonErr(w, "query journal: "+err.Error(), 500)
		return
	}
	defer journalRows.Close()
	var journalEntries []journalEntry
	for journalRows.Next() {
		var e journalEntry
		if err := journalRows.Scan(&e.ID, &e.ProjectID, &e.EntryType, &e.Content, &e.CreatedAt); err != nil {
			jsonErr(w, "scan journal: "+err.Error(), 500)
			return
		}
		journalEntries = append(journalEntries, e)
	}

	// ── Compute stats ──
	var doneCount, activeCount, pendingCount int
	var totalOrig, totalCurr, totalActual float64
	now := time.Now()
	today := now.Format("2006-01-02")

	for _, t := range tasks {
		switch t.Status {
		case "done":
			doneCount++
		case "active", "in_progress":
			activeCount++
		default:
			pendingCount++
		}
		totalOrig += t.OrigBudget
		totalCurr += t.CurrBudget
		totalActual += t.ActualBudget
	}
	totalTasks := len(tasks)
	pctComplete := 0
	if totalTasks > 0 {
		pctComplete = doneCount * 100 / totalTasks
	}

	// Checklist completion count
	var checkTotal, checkDone int
	for _, c := range txData.Checklist {
		checkTotal++
		if c.HereNow {
			checkDone++
		}
	}
	for _, c := range txData.Backmatter {
		checkTotal++
		if c.HereNow {
			checkDone++
		}
	}

	generated := now.Format("January 2, 2006 at 3:04 PM MST")
	generatedShort := now.Format("2006-01-02 15:04 MST")

	// ── Build HTML ──
	htmlBody := buildSnapshotHTML(snapshotParams{
		ProjectName:  projName,
		ProjectURL:   projectURL,
		Generated:    generated,
		Tasks:        tasks,
		DoneCount:    doneCount,
		ActiveCount:  activeCount,
		PendingCount: pendingCount,
		TotalTasks:   totalTasks,
		PctComplete:  pctComplete,
		TotalOrig:    totalOrig,
		TotalCurr:    totalCurr,
		TotalActual:  totalActual,
		Today:        today,
		HasTx:        hasTx,
		TxStatus:     txStatus,
		TxData:       txData,
		CheckTotal:   checkTotal,
		CheckDone:    checkDone,
		FileLog:      fileLogEntries,
		Journal:      journalEntries,
	})

	// ── Build plain text ──
	textBody := buildSnapshotText(snapshotParams{
		ProjectName:  projName,
		ProjectURL:   projectURL,
		Generated:    generatedShort,
		Tasks:        tasks,
		DoneCount:    doneCount,
		ActiveCount:  activeCount,
		PendingCount: pendingCount,
		TotalTasks:   totalTasks,
		PctComplete:  pctComplete,
		TotalOrig:    totalOrig,
		TotalCurr:    totalCurr,
		TotalActual:  totalActual,
		Today:        today,
		HasTx:        hasTx,
		TxStatus:     txStatus,
		TxData:       txData,
		CheckTotal:   checkTotal,
		CheckDone:    checkDone,
		FileLog:      fileLogEntries,
		Journal:      journalEntries,
	})

	subject := fmt.Sprintf("Project Snapshot: %s — %s", projName, now.Format("Jan 2, 2006"))

	to := []string{body.Recipients[0]}
	var cc []string
	if len(body.Recipients) > 1 {
		cc = body.Recipients[1:]
	}

	if err := s.Email.sendEmail(to, cc, subject, textBody, htmlBody); err != nil {
		slog.Error("send snapshot email", "error", err)
		jsonErr(w, "failed to send email: "+err.Error(), 500)
		return
	}

	jsonOK(w, map[string]any{
		"ok":      true,
		"sent_to": body.Recipients,
		"subject": subject,
	})
}

type snapshotTask struct {
	SortOrder    int64
	Assignee     string
	Title        string
	IsMilestone  int64
	Status       string
	CurrDue      string
	ActualDone   string
	OrigBudget   float64
	CurrBudget   float64
	ActualBudget float64
}

type snapshotParams struct {
	ProjectName  string
	ProjectURL   string
	Generated    string
	Tasks        []snapshotTask
	DoneCount    int
	ActiveCount  int
	PendingCount int
	TotalTasks   int
	PctComplete  int
	TotalOrig    float64
	TotalCurr    float64
	TotalActual  float64
	Today        string
	HasTx        bool
	TxStatus     string
	TxData       transmittalEmailData
	CheckTotal   int
	CheckDone    int
	FileLog      []fileLogEntry
	Journal      []journalEntry
}

func snapshotStatusLabel(status string) string {
	switch status {
	case "done":
		return "Done"
	case "active", "in_progress":
		return "Active"
	default:
		return "Pending"
	}
}

func snapshotFormatMoney(v float64) string {
	if v == 0 {
		return "—"
	}
	// Format with commas (Go doesn't have %,f)
	neg := v < 0
	if neg {
		v = -v
	}
	whole := int64(v)
	frac := v - float64(whole)

	// Add commas to the whole part
	s := fmt.Sprintf("%d", whole)
	if len(s) > 3 {
		var parts []string
		for len(s) > 3 {
			parts = append([]string{s[len(s)-3:]}, parts...)
			s = s[:len(s)-3]
		}
		parts = append([]string{s}, parts...)
		s = strings.Join(parts, ",")
	}
	result := fmt.Sprintf("$%s.%02d", s, int64(frac*100+0.5))
	if neg {
		result = "-" + result
	}
	return result
}

func snapshotFormatDate(d string) string {
	if d == "" {
		return "—"
	}
	t, err := time.Parse("2006-01-02", d)
	if err != nil {
		// Try with time component
		t, err = time.Parse("2006-01-02T15:04:05", d)
		if err != nil {
			return d
		}
	}
	return t.Format("Jan 2, 2006")
}

func isOverdue(currDue, today, status string) bool {
	if status == "done" || currDue == "" {
		return false
	}
	return currDue < today
}

// ─── HTML builder ───

func buildSnapshotHTML(p snapshotParams) string {
	var b strings.Builder

	b.WriteString(`<!DOCTYPE html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width"></head>`)
	b.WriteString(`<body style="margin:0;padding:0;background:#f4f3f9;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;color:#333;">`)

	// Outer wrapper table for email clients
	b.WriteString(`<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background:#f4f3f9;"><tr><td align="center" style="padding:24px 12px;">`)
	b.WriteString(`<table role="presentation" width="640" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:12px;overflow:hidden;box-shadow:0 2px 12px rgba(108,99,255,0.08);">`)

	// ── Header ──
	b.WriteString(`<tr><td style="background:linear-gradient(135deg,#6c63ff 0%,#8b83ff 100%);padding:32px 36px;">`)
	b.WriteString(fmt.Sprintf(`<h1 style="margin:0 0 4px;font-size:24px;font-weight:700;color:#ffffff;">%s</h1>`, html.EscapeString(p.ProjectName)))
	b.WriteString(fmt.Sprintf(`<p style="margin:0;font-size:13px;color:rgba(255,255,255,0.8);">Project Snapshot · %s</p>`, html.EscapeString(p.Generated)))
	if p.ProjectURL != "" {
		b.WriteString(fmt.Sprintf(`<p style="margin:8px 0 0;font-size:13px;"><a href="%s" style="color:#d4d0ff;text-decoration:underline;">View project online →</a></p>`, p.ProjectURL))
	}
	b.WriteString(`</td></tr>`)

	// ── Schedule Overview ──
	b.WriteString(`<tr><td style="padding:28px 36px 0;">`)
	b.WriteString(`<h2 style="margin:0 0 16px;font-size:16px;font-weight:700;color:#6c63ff;text-transform:uppercase;letter-spacing:0.5px;">Schedule Overview</h2>`)

	// Stat cards row
	b.WriteString(`<table role="presentation" width="100%" cellpadding="0" cellspacing="0"><tr>`)

	// Pct complete card
	b.WriteString(`<td width="25%" style="padding:0 6px 0 0;">`)
	b.WriteString(fmt.Sprintf(`<div style="background:#f0eeff;border-radius:8px;padding:14px 16px;text-align:center;">`+
		`<div style="font-size:28px;font-weight:700;color:#6c63ff;">%d%%</div>`+
		`<div style="font-size:11px;color:#888;margin-top:2px;">Complete</div></div>`, p.PctComplete))
	b.WriteString(`</td>`)

	// Done card
	b.WriteString(`<td width="25%" style="padding:0 6px;">`)
	b.WriteString(fmt.Sprintf(`<div style="background:#d1fae5;border-radius:8px;padding:14px 16px;text-align:center;">`+
		`<div style="font-size:28px;font-weight:700;color:#065f46;">%d</div>`+
		`<div style="font-size:11px;color:#888;margin-top:2px;">Done</div></div>`, p.DoneCount))
	b.WriteString(`</td>`)

	// Active card
	b.WriteString(`<td width="25%" style="padding:0 6px;">`)
	b.WriteString(fmt.Sprintf(`<div style="background:#dbeafe;border-radius:8px;padding:14px 16px;text-align:center;">`+
		`<div style="font-size:28px;font-weight:700;color:#1e40af;">%d</div>`+
		`<div style="font-size:11px;color:#888;margin-top:2px;">Active</div></div>`, p.ActiveCount))
	b.WriteString(`</td>`)

	// Pending card
	b.WriteString(`<td width="25%" style="padding:0 0 0 6px;">`)
	b.WriteString(fmt.Sprintf(`<div style="background:#fef3c7;border-radius:8px;padding:14px 16px;text-align:center;">`+
		`<div style="font-size:28px;font-weight:700;color:#92400e;">%d</div>`+
		`<div style="font-size:11px;color:#888;margin-top:2px;">Pending</div></div>`, p.PendingCount))
	b.WriteString(`</td>`)

	b.WriteString(`</tr></table>`)
	b.WriteString(`</td></tr>`)

	// ── Task Schedule Table ──
	if len(p.Tasks) > 0 {
		b.WriteString(`<tr><td style="padding:28px 36px 0;">`)
		b.WriteString(`<h2 style="margin:0 0 12px;font-size:16px;font-weight:700;color:#6c63ff;text-transform:uppercase;letter-spacing:0.5px;">Task Schedule</h2>`)
		b.WriteString(`<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="border:1px solid #e5e5e5;border-radius:8px;overflow:hidden;">`)

		// Table header
		b.WriteString(`<tr style="background:#6c63ff;">`)
		b.WriteString(`<td style="padding:10px 12px;font-size:11px;font-weight:700;color:#fff;text-transform:uppercase;letter-spacing:0.5px;">Task</td>`)
		b.WriteString(`<td style="padding:10px 12px;font-size:11px;font-weight:700;color:#fff;text-transform:uppercase;letter-spacing:0.5px;width:90px;">Assignee</td>`)
		b.WriteString(`<td style="padding:10px 12px;font-size:11px;font-weight:700;color:#fff;text-transform:uppercase;letter-spacing:0.5px;width:72px;">Status</td>`)
		b.WriteString(`<td style="padding:10px 12px;font-size:11px;font-weight:700;color:#fff;text-transform:uppercase;letter-spacing:0.5px;width:100px;">Due Date</td>`)
		b.WriteString(`</tr>`)

		for i, t := range p.Tasks {
			rowBg := "#ffffff"
			if i%2 == 1 {
				rowBg = "#faf9ff"
			}
			overdue := isOverdue(t.CurrDue, p.Today, t.Status)
			if overdue {
				rowBg = "#fef2f2"
			}

			// Status badge colors
			var badgeBg, badgeColor string
			switch t.Status {
			case "done":
				badgeBg = "#d1fae5"
				badgeColor = "#065f46"
			case "active", "in_progress":
				badgeBg = "#dbeafe"
				badgeColor = "#1e40af"
			default:
				badgeBg = "#f3f4f6"
				badgeColor = "#6b7280"
			}

			titleStyle := "font-size:13px;font-weight:500;color:#333;"
			if t.IsMilestone != 0 {
				titleStyle = "font-size:13px;font-weight:700;color:#6c63ff;"
			}

			dueDateDisplay := snapshotFormatDate(t.CurrDue)
			dueDateStyle := "font-size:13px;color:#555;"
			if overdue {
				dueDateStyle = "font-size:13px;color:#dc2626;font-weight:600;"
				dueDateDisplay += " \u26a0"
			}

			milestoneIcon := ""
			if t.IsMilestone != 0 {
				milestoneIcon = `<span style="margin-right:4px;">\u25c6</span>`
			}

			b.WriteString(fmt.Sprintf(`<tr style="background:%s;">`, rowBg))
			b.WriteString(fmt.Sprintf(`<td style="padding:9px 12px;border-top:1px solid #eee;%s">%s%s</td>`,
				titleStyle, milestoneIcon, html.EscapeString(t.Title)))
			b.WriteString(fmt.Sprintf(`<td style="padding:9px 12px;border-top:1px solid #eee;font-size:13px;color:#555;">%s</td>`,
				html.EscapeString(t.Assignee)))
			b.WriteString(fmt.Sprintf(`<td style="padding:9px 12px;border-top:1px solid #eee;"><span style="display:inline-block;padding:2px 8px;border-radius:10px;font-size:11px;font-weight:600;background:%s;color:%s;">%s</span></td>`,
				badgeBg, badgeColor, snapshotStatusLabel(t.Status)))
			b.WriteString(fmt.Sprintf(`<td style="padding:9px 12px;border-top:1px solid #eee;%s">%s</td>`,
				dueDateStyle, dueDateDisplay))
			b.WriteString(`</tr>`)
		}

		b.WriteString(`</table>`)
		b.WriteString(`</td></tr>`)
	}

	// ── Budget Summary ──
	b.WriteString(`<tr><td style="padding:28px 36px 0;">`)
	b.WriteString(`<h2 style="margin:0 0 12px;font-size:16px;font-weight:700;color:#6c63ff;text-transform:uppercase;letter-spacing:0.5px;">Budget Summary</h2>`)
	b.WriteString(`<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="border:1px solid #e5e5e5;border-radius:8px;overflow:hidden;">`)

	// Budget header
	b.WriteString(`<tr style="background:#6c63ff;">`)
	b.WriteString(`<td style="padding:10px 12px;font-size:11px;font-weight:700;color:#fff;text-transform:uppercase;letter-spacing:0.5px;"></td>`)
	b.WriteString(`<td style="padding:10px 12px;font-size:11px;font-weight:700;color:#fff;text-transform:uppercase;letter-spacing:0.5px;text-align:right;width:120px;">Original</td>`)
	b.WriteString(`<td style="padding:10px 12px;font-size:11px;font-weight:700;color:#fff;text-transform:uppercase;letter-spacing:0.5px;text-align:right;width:120px;">Current</td>`)
	b.WriteString(`<td style="padding:10px 12px;font-size:11px;font-weight:700;color:#fff;text-transform:uppercase;letter-spacing:0.5px;text-align:right;width:120px;">Actual</td>`)
	b.WriteString(`</tr>`)

	// Totals row
	b.WriteString(`<tr style="background:#ffffff;">`)
	b.WriteString(`<td style="padding:12px;font-size:14px;font-weight:600;color:#333;">Total Budget</td>`)
	b.WriteString(fmt.Sprintf(`<td style="padding:12px;font-size:14px;text-align:right;color:#555;">%s</td>`, snapshotFormatMoney(p.TotalOrig)))
	b.WriteString(fmt.Sprintf(`<td style="padding:12px;font-size:14px;text-align:right;color:#555;">%s</td>`, snapshotFormatMoney(p.TotalCurr)))
	b.WriteString(fmt.Sprintf(`<td style="padding:12px;font-size:14px;text-align:right;color:#555;">%s</td>`, snapshotFormatMoney(p.TotalActual)))
	b.WriteString(`</tr>`)

	// Variance row
	variance := p.TotalCurr - p.TotalActual
	varianceColor := "#065f46"
	varianceLabel := "Under budget"
	if variance < 0 {
		varianceColor = "#dc2626"
		varianceLabel = "Over budget"
		variance = -variance
	} else if variance == 0 {
		varianceColor = "#555"
		varianceLabel = "On budget"
	}

	b.WriteString(fmt.Sprintf(`<tr style="background:#faf9ff;border-top:1px solid #eee;">`+
		`<td colspan="3" style="padding:12px;font-size:13px;font-weight:600;color:%s;">%s</td>`+
		`<td style="padding:12px;font-size:14px;font-weight:700;text-align:right;color:%s;">%s</td>`+
		`</tr>`, varianceColor, varianceLabel, varianceColor, snapshotFormatMoney(variance)))

	b.WriteString(`</table>`)
	b.WriteString(`</td></tr>`)

	// ── Transmittal Status ──
	if p.HasTx {
		b.WriteString(`<tr><td style="padding:28px 36px 0;">`)
		b.WriteString(`<h2 style="margin:0 0 12px;font-size:16px;font-weight:700;color:#6c63ff;text-transform:uppercase;letter-spacing:0.5px;">Transmittal Status</h2>`)

		// Transmittal card
		b.WriteString(`<div style="border:1px solid #e5e5e5;border-radius:8px;overflow:hidden;">`)

		// Title bar
		txBadgeBg := "#fef3c7"
		txBadgeColor := "#92400e"
		if strings.ToLower(p.TxStatus) == "final" {
			txBadgeBg = "#d1fae5"
			txBadgeColor = "#065f46"
		}
		bookTitle := p.TxData.Book.Title
		if bookTitle == "" {
			bookTitle = "Untitled"
		}
		b.WriteString(fmt.Sprintf(`<div style="padding:14px 16px;background:#faf9ff;border-bottom:1px solid #eee;display:flex;align-items:center;">`+
			`<span style="font-size:14px;font-weight:600;color:#333;">%s</span>`+
			`<span style="display:inline-block;margin-left:10px;padding:2px 10px;border-radius:10px;font-size:11px;font-weight:600;background:%s;color:%s;">%s</span>`+
			`</div>`, html.EscapeString(bookTitle), txBadgeBg, txBadgeColor, strings.ToUpper(p.TxStatus)))

		// Details table
		b.WriteString(`<table role="presentation" width="100%" cellpadding="0" cellspacing="0">`)

		if p.TxData.Book.Author != "" {
			b.WriteString(fmt.Sprintf(`<tr><td style="padding:8px 16px;font-size:12px;color:#888;width:130px;">Author</td>`+
				`<td style="padding:8px 16px;font-size:13px;color:#333;">%s</td></tr>`, html.EscapeString(p.TxData.Book.Author)))
		}
		if p.TxData.Book.Publisher != "" {
			b.WriteString(fmt.Sprintf(`<tr><td style="padding:8px 16px;font-size:12px;color:#888;">Publisher</td>`+
				`<td style="padding:8px 16px;font-size:13px;color:#333;">%s</td></tr>`, html.EscapeString(p.TxData.Book.Publisher)))
		}
		if p.TxData.Production.TransmittalDate != "" {
			b.WriteString(fmt.Sprintf(`<tr><td style="padding:8px 16px;font-size:12px;color:#888;">Transmittal date</td>`+
				`<td style="padding:8px 16px;font-size:13px;color:#333;">%s</td></tr>`, html.EscapeString(p.TxData.Production.TransmittalDate)))
		}
		if p.TxData.Production.BoundBookDate != "" {
			b.WriteString(fmt.Sprintf(`<tr><td style="padding:8px 16px;font-size:12px;color:#888;">Bound book date</td>`+
				`<td style="padding:8px 16px;font-size:13px;color:#333;">%s</td></tr>`, html.EscapeString(p.TxData.Production.BoundBookDate)))
		}
		if p.TxData.Production.MechsDelivery != "" {
			b.WriteString(fmt.Sprintf(`<tr><td style="padding:8px 16px;font-size:12px;color:#888;">Mechs delivery</td>`+
				`<td style="padding:8px 16px;font-size:13px;color:#333;">%s</td></tr>`, html.EscapeString(p.TxData.Production.MechsDelivery)))
		}
		if p.TxData.Production.WeeksInProd != "" {
			b.WriteString(fmt.Sprintf(`<tr><td style="padding:8px 16px;font-size:12px;color:#888;">Weeks in prod</td>`+
				`<td style="padding:8px 16px;font-size:13px;color:#333;">%s</td></tr>`, html.EscapeString(p.TxData.Production.WeeksInProd)))
		}

		// Checklist progress
		b.WriteString(fmt.Sprintf(`<tr><td style="padding:8px 16px;font-size:12px;color:#888;">Checklist</td>`+
			`<td style="padding:8px 16px;font-size:13px;color:#333;">%d / %d items received</td></tr>`, p.CheckDone, p.CheckTotal))

		b.WriteString(`</table>`)
		b.WriteString(`</div>`) // end card
		b.WriteString(`</td></tr>`)
	}

	// ── Recent Files ──
	if len(p.FileLog) > 0 {
		b.WriteString(`<tr><td style="padding:28px 36px 0;">`)
		b.WriteString(`<h2 style="margin:0 0 12px;font-size:16px;font-weight:700;color:#6c63ff;text-transform:uppercase;letter-spacing:0.5px;">Recent Files</h2>`)
		b.WriteString(`<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="border:1px solid #e5e5e5;border-radius:8px;overflow:hidden;">`)

		// Table header
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
		b.WriteString(fmt.Sprintf(`<p style="margin:8px 0 0;font-size:12px;color:#aaa;">Showing %d most recent file transfers</p>`, len(p.FileLog)))
		b.WriteString(`</td></tr>`)
	}

	// ── Recent Journal ──
	if len(p.Journal) > 0 {
		b.WriteString(`<tr><td style="padding:28px 36px 0;">`)
		b.WriteString(`<h2 style="margin:0 0 12px;font-size:16px;font-weight:700;color:#6c63ff;text-transform:uppercase;letter-spacing:0.5px;">Recent Journal</h2>`)

		for i, e := range p.Journal {
			rowBg := "#ffffff"
			if i%2 == 1 {
				rowBg = "#faf9ff"
			}
			emoji := "📝"
			switch e.EntryType {
			case "call":
				emoji = "📞"
			case "decision":
				emoji = "⚖️"
			case "approval":
				emoji = "✅"
			}
			dateStr := snapshotFormatDate(e.CreatedAt)
			// Try parsing as datetime for a more detailed display
			if t, err := time.Parse("2006-01-02T15:04:05", e.CreatedAt); err == nil {
				dateStr = t.Format("Jan 2, 2006 3:04 PM")
			} else if t, err := time.Parse("2006-01-02 15:04:05", e.CreatedAt); err == nil {
				dateStr = t.Format("Jan 2, 2006 3:04 PM")
			}
			borderRadius := ""
			if i == 0 {
				borderRadius = "border-radius:8px 8px 0 0;"
			}
			if i == len(p.Journal)-1 {
				if i == 0 {
					borderRadius = "border-radius:8px;"
				} else {
					borderRadius = "border-radius:0 0 8px 8px;"
				}
			}
			borderBottom := "border-bottom:0;"
			if i == len(p.Journal)-1 {
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

	// ── Footer ──
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

func buildSnapshotText(p snapshotParams) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("PROJECT SNAPSHOT: %s\n", strings.ToUpper(p.ProjectName)))
	b.WriteString(strings.Repeat("=", 50) + "\n")
	b.WriteString(fmt.Sprintf("Generated: %s\n", p.Generated))
	if p.ProjectURL != "" {
		b.WriteString(fmt.Sprintf("Online:    %s\n", p.ProjectURL))
	}
	b.WriteString("\n")

	// Schedule overview
	b.WriteString("SCHEDULE OVERVIEW\n")
	b.WriteString(strings.Repeat("\u2500", 35) + "\n")
	b.WriteString(fmt.Sprintf("  Complete:  %d%% (%d/%d tasks)\n", p.PctComplete, p.DoneCount, p.TotalTasks))
	b.WriteString(fmt.Sprintf("  Done:      %d\n", p.DoneCount))
	b.WriteString(fmt.Sprintf("  Active:    %d\n", p.ActiveCount))
	b.WriteString(fmt.Sprintf("  Pending:   %d\n", p.PendingCount))
	b.WriteString("\n")

	// Task list
	if len(p.Tasks) > 0 {
		b.WriteString("TASK SCHEDULE\n")
		b.WriteString(strings.Repeat("\u2500", 35) + "\n")
		for _, t := range p.Tasks {
			overdue := isOverdue(t.CurrDue, p.Today, t.Status)
			marker := "  "
			if overdue {
				marker = "\u26a0 "
			}
			icon := " "
			if t.IsMilestone != 0 {
				icon = "\u25c6"
			}
			due := t.CurrDue
			if due == "" {
				due = "\u2014"
			}
			b.WriteString(fmt.Sprintf("%s%s %-30s  %-8s  %-8s  %s\n",
				marker, icon, t.Title, t.Assignee, snapshotStatusLabel(t.Status), due))
		}
		b.WriteString("\n")
	}

	// Budget
	b.WriteString("BUDGET SUMMARY\n")
	b.WriteString(strings.Repeat("\u2500", 35) + "\n")
	b.WriteString(fmt.Sprintf("  Original:  %s\n", snapshotFormatMoney(p.TotalOrig)))
	b.WriteString(fmt.Sprintf("  Current:   %s\n", snapshotFormatMoney(p.TotalCurr)))
	b.WriteString(fmt.Sprintf("  Actual:    %s\n", snapshotFormatMoney(p.TotalActual)))
	variance := p.TotalCurr - p.TotalActual
	if variance >= 0 {
		b.WriteString(fmt.Sprintf("  Variance:  %s under budget\n", snapshotFormatMoney(variance)))
	} else {
		b.WriteString(fmt.Sprintf("  Variance:  %s OVER budget\n", snapshotFormatMoney(-variance)))
	}
	b.WriteString("\n")

	// Transmittal
	if p.HasTx {
		b.WriteString("TRANSMITTAL STATUS\n")
		b.WriteString(strings.Repeat("\u2500", 35) + "\n")
		bookTitle := p.TxData.Book.Title
		if bookTitle == "" {
			bookTitle = "Untitled"
		}
		b.WriteString(fmt.Sprintf("  Book:       %s\n", bookTitle))
		b.WriteString(fmt.Sprintf("  Status:     %s\n", strings.ToUpper(p.TxStatus)))
		if p.TxData.Book.Author != "" {
			b.WriteString(fmt.Sprintf("  Author:     %s\n", p.TxData.Book.Author))
		}
		if p.TxData.Production.TransmittalDate != "" {
			b.WriteString(fmt.Sprintf("  TX Date:    %s\n", p.TxData.Production.TransmittalDate))
		}
		if p.TxData.Production.BoundBookDate != "" {
			b.WriteString(fmt.Sprintf("  Bound Book: %s\n", p.TxData.Production.BoundBookDate))
		}
		b.WriteString(fmt.Sprintf("  Checklist:  %d / %d items received\n", p.CheckDone, p.CheckTotal))
		b.WriteString("\n")
	}

	// Recent files
	if len(p.FileLog) > 0 {
		b.WriteString("RECENT FILES\n")
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

	// Recent journal
	if len(p.Journal) > 0 {
		b.WriteString("RECENT JOURNAL\n")
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

	b.WriteString(strings.Repeat("\u2500", 50) + "\n")
	b.WriteString(fmt.Sprintf("Sent: %s\n", p.Generated))
	if p.ProjectURL != "" {
		b.WriteString(fmt.Sprintf("View online: %s\n", p.ProjectURL))
	}

	return b.String()
}
