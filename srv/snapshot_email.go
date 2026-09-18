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
		jsonErr(w, "email not configured (set PRODCAL_MAIL_FROM for Resend, or AGENTMAIL_API_KEY + AGENTMAIL_INBOX_ID)", 503)
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

	if err := s.mail(mailMeta{Kind: mailKindSnapshot, RefType: "project", RefID: mailRef(pid), TriggeredBy: triggeredBy(r, "client")}, to, cc, subject, textBody, htmlBody); err != nil {
		slog.Error("send snapshot email", "error", err)
		jsonErr(w, "email send failed", 500)
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
			// Unparseable input flows into HTML builders unescaped (here, in
			// activity_email.go and client_digest_email.go), so return an
			// HTML-escaped form to keep malformed values from carrying markup.
			return html.EscapeString(d)
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

// snapshotJournalWhen formats a journal timestamp for HTML tables; falls
// back to the date-only formatter (which escapes unparseable input).
func snapshotJournalWhen(createdAt string) string {
	if t, err := time.Parse("2006-01-02T15:04:05", createdAt); err == nil {
		return t.Format("Jan 2, 2006 3:04 PM")
	}
	if t, err := time.Parse("2006-01-02 15:04:05", createdAt); err == nil {
		return t.Format("Jan 2, 2006 3:04 PM")
	}
	return snapshotFormatDate(createdAt)
}

// fileDirLabel renders a file-log direction as a mono "in"/"out" label.
func fileDirLabel(direction string) string {
	if direction == "outbound" {
		return emailStatus("out")
	}
	return emailStatus("in")
}

// fileLogTableRows builds the shared Date / Dir / File / Type / From → To
// rows used by the snapshot, activity and digest emails. Cells are escaped.
func fileLogTableRows(entries []fileLogEntry, withParties bool) [][]string {
	rows := make([][]string, 0, len(entries))
	for _, e := range entries {
		row := []string{
			snapshotFormatDate(e.TransferDate),
			fileDirLabel(e.Direction),
			html.EscapeString(e.Filename),
			html.EscapeString(e.FileType),
		}
		if withParties {
			row = append(row, html.EscapeString(e.SentBy)+" &rarr; "+html.EscapeString(e.ReceivedBy))
		}
		rows = append(rows, row)
	}
	return rows
}

// journalTableRows builds the shared When / Type / Entry rows. Cells are escaped.
func journalTableRows(entries []journalEntry) [][]string {
	rows := make([][]string, 0, len(entries))
	for _, e := range entries {
		rows = append(rows, []string{
			snapshotJournalWhen(e.CreatedAt),
			emailStatus(activityJournalLabel(e.EntryType)),
			html.EscapeString(e.Content),
		})
	}
	return rows
}

// ─── HTML builder ───

func buildSnapshotHTML(p snapshotParams) string {
	var b strings.Builder

	// ── Schedule overview ──
	b.WriteString(emailH2("Schedule overview"))
	b.WriteString(emailStats([][2]string{
		{"Complete", fmt.Sprintf("%d%%", p.PctComplete)},
		{"Done", fmt.Sprintf("%d", p.DoneCount)},
		{"Active", fmt.Sprintf("%d", p.ActiveCount)},
		{"Pending", fmt.Sprintf("%d", p.PendingCount)},
	}))

	// ── Task schedule ──
	if len(p.Tasks) > 0 {
		b.WriteString(emailH2("Task schedule"))
		rows := make([][]string, 0, len(p.Tasks))
		for _, t := range p.Tasks {
			overdue := isOverdue(t.CurrDue, p.Today, t.Status)
			title := html.EscapeString(t.Title)
			if t.IsMilestone != 0 {
				title = fmt.Sprintf(`<span style="color:%s">&#9670;</span> <strong>%s</strong>`, emailAccent, title)
			}
			due := snapshotFormatDate(t.CurrDue)
			if overdue {
				due = fmt.Sprintf(`<span style="color:%s;font-weight:600">%s</span> %s`, emailRed, due, emailStatus("overdue"))
			}
			rows = append(rows, []string{
				title,
				html.EscapeString(t.Assignee),
				emailStatus(snapshotStatusLabel(t.Status)),
				due,
			})
		}
		b.WriteString(emailTable([]string{"Task", "Assignee", "Status", "Due"}, rows, nil))
	}

	// ── Budget summary ──
	b.WriteString(emailH2("Budget summary"))
	variance := p.TotalCurr - p.TotalActual
	varianceColor := emailGreen
	varianceLabel := "Under budget"
	if variance < 0 {
		varianceColor = emailRed
		varianceLabel = "Over budget"
		variance = -variance
	} else if variance == 0 {
		varianceColor = emailSecondary
		varianceLabel = "On budget"
	}
	b.WriteString(emailTable(
		[]string{"", "Original", "Current", "Actual"},
		[][]string{
			{"<strong>Total budget</strong>", snapshotFormatMoney(p.TotalOrig), snapshotFormatMoney(p.TotalCurr), snapshotFormatMoney(p.TotalActual)},
			{fmt.Sprintf(`<span style="color:%s;font-weight:600">%s</span>`, varianceColor, varianceLabel), "", "",
				fmt.Sprintf(`<span style="color:%s;font-weight:600">%s</span>`, varianceColor, snapshotFormatMoney(variance))},
		},
		[]string{"l", "r", "r", "r"},
	))

	// ── Transmittal status ──
	if p.HasTx {
		b.WriteString(emailH2("Transmittal status"))
		bookTitle := p.TxData.Book.Title
		if bookTitle == "" {
			bookTitle = "Untitled"
		}
		kv := [][2]string{
			{"Book", html.EscapeString(bookTitle) + " &nbsp;" + emailStatus(p.TxStatus)},
		}
		if p.TxData.Book.Author != "" {
			kv = append(kv, [2]string{"Author", html.EscapeString(p.TxData.Book.Author)})
		}
		if p.TxData.Book.Publisher != "" {
			kv = append(kv, [2]string{"Publisher", html.EscapeString(p.TxData.Book.Publisher)})
		}
		if p.TxData.Production.TransmittalDate != "" {
			kv = append(kv, [2]string{"Transmittal date", html.EscapeString(p.TxData.Production.TransmittalDate)})
		}
		if p.TxData.Production.BoundBookDate != "" {
			kv = append(kv, [2]string{"Bound book date", html.EscapeString(p.TxData.Production.BoundBookDate)})
		}
		if p.TxData.Production.MechsDelivery != "" {
			kv = append(kv, [2]string{"Mechs delivery", html.EscapeString(p.TxData.Production.MechsDelivery)})
		}
		if p.TxData.Production.WeeksInProd != "" {
			kv = append(kv, [2]string{"Weeks in prod", html.EscapeString(p.TxData.Production.WeeksInProd)})
		}
		kv = append(kv, [2]string{"Checklist", fmt.Sprintf("%d / %d items received", p.CheckDone, p.CheckTotal)})
		b.WriteString(emailKV(kv))
	}

	// ── Recent files ──
	if len(p.FileLog) > 0 {
		b.WriteString(emailH2("Recent Files"))
		b.WriteString(emailTable([]string{"Date", "Dir", "File", "Type", "From \u2192 To"}, fileLogTableRows(p.FileLog, true), nil))
		b.WriteString(emailSmall(fmt.Sprintf("Showing %d most recent file transfers", len(p.FileLog))))
	}

	// ── Recent journal ──
	if len(p.Journal) > 0 {
		b.WriteString(emailH2("Recent Journal"))
		b.WriteString(emailTable([]string{"When", "Type", "Entry"}, journalTableRows(p.Journal), nil))
	}

	if p.ProjectURL != "" {
		b.WriteString(emailButton(p.ProjectURL, "View project online"))
	}

	return emailShell(b.String(), emailShellOpts{
		Kicker: "Project snapshot \u00b7 " + p.Generated,
		Title:  p.ProjectName,
		Footer: fmt.Sprintf(`Generated %s &middot; Reply to this email to reach Jenna.`, html.EscapeString(p.Generated)),
	})
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
