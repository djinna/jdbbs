package srv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// AgentMail config — set via environment variables:
//
//	AGENTMAIL_API_KEY       — Bearer token
//	AGENTMAIL_INBOX_ID      — inbox ID for jdbb@agentmail.to (also the sending address)
//	PRODCAL_MAIL_FROM_NAME  — sender display name (default "ProdCal"; set empty to disable)
//	PRODCAL_MAIL_REPLY_TO   — Reply-To address (default "j@djinna.com"; set empty to disable)
type EmailConfig struct {
	APIKey   string
	InboxID  string
	FromName string
	ReplyTo  string
}

func LoadEmailConfig() *EmailConfig {
	key := os.Getenv("AGENTMAIL_API_KEY")
	inbox := os.Getenv("AGENTMAIL_INBOX_ID")
	if key == "" || inbox == "" {
		return nil
	}
	fromName := "ProdCal"
	if v, ok := os.LookupEnv("PRODCAL_MAIL_FROM_NAME"); ok {
		fromName = v
	}
	replyTo := "j@djinna.com"
	if v, ok := os.LookupEnv("PRODCAL_MAIL_REPLY_TO"); ok {
		replyTo = v
	}
	return &EmailConfig{APIKey: key, InboxID: inbox, FromName: fromName, ReplyTo: replyTo}
}

// sendEmail sends via AgentMail API.
// POST https://api.agentmail.to/v0/inboxes/:inbox_id/messages/send
func (cfg *EmailConfig) sendEmail(to []string, cc []string, subject, textBody, htmlBody string) error {
	return cfg.sendEmailWithHeaders(to, cc, subject, textBody, htmlBody, nil)
}

// sendEmailWithHeaders is sendEmail plus extra MIME headers via the AgentMail
// generic "headers" map (e.g. List-Unsubscribe for the recurring digest).
func (cfg *EmailConfig) sendEmailWithHeaders(to []string, cc []string, subject, textBody, htmlBody string, extraHeaders map[string]string) error {
	url := fmt.Sprintf("https://api.agentmail.to/v0/inboxes/%s/messages/send", cfg.InboxID)

	body := map[string]any{
		"to":      to,
		"subject": subject,
	}
	if len(cc) > 0 {
		body["cc"] = cc
	}
	if textBody != "" {
		body["text"] = textBody
	}
	if htmlBody != "" {
		body["html"] = htmlBody
	}
	if cfg.ReplyTo != "" {
		body["reply_to"] = cfg.ReplyTo // native SendMessageRequest field
	}
	headers := map[string]string{}
	if cfg.FromName != "" {
		// AgentMail has no per-message from-name field; the sender is always
		// the inbox. Pass a From header (same inbox address, plus display
		// name) through the generic headers map instead.
		headers["From"] = fmt.Sprintf("%q <%s>", cfg.FromName, cfg.InboxID)
	}
	for k, v := range extraHeaders {
		headers[k] = v
	}
	if len(headers) > 0 {
		body["headers"] = headers
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal email body: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		// Log the full response body server-side only. Handlers surface this
		// error's message to end users, so keep the body out of it.
		slog.Error("agentmail API error", "status", resp.StatusCode, "body", string(respBody), "to", to, "subject", subject)
		return fmt.Errorf("agentmail API error: status %d", resp.StatusCode)
	}

	slog.Info("email sent", "to", to, "cc", cc, "subject", subject, "status", resp.StatusCode)
	return nil
}

// ─── Transmittal email summary ───

type transmittalEmailData struct {
	Book struct {
		Author    string `json:"author"`
		Title     string `json:"title"`
		Subtitle  string `json:"subtitle"`
		Publisher string `json:"publisher"`
		Editor    string `json:"editor"`
		ISBNPaper string `json:"isbn_paper"`
		ISBNCloth string `json:"isbn_cloth"`
	} `json:"book"`
	Production struct {
		TransmittalDate string `json:"transmittal_date"`
		MechsDelivery   string `json:"mechs_delivery"`
		WeeksInProd     string `json:"weeks_in_production"`
		BoundBookDate   string `json:"bound_book_date"`
		PrintRun        string `json:"print_run"`
	} `json:"production"`
	ChecklistStats struct {
		Parts      string `json:"parts"`
		Chapters   string `json:"chapters"`
		WordsChars string `json:"words_chars"`
		MSPP       string `json:"ms_pp"`
		EstBookPP  string `json:"est_book_pp"`
	} `json:"checklist_stats"`
	Checklist []struct {
		Component  string `json:"component"`
		HereNow    bool   `json:"here_now"`
		ToComeWhen string `json:"to_come_when"`
	} `json:"checklist"`
	Backmatter []struct {
		Component  string `json:"component"`
		HereNow    bool   `json:"here_now"`
		ToComeWhen string `json:"to_come_when"`
	} `json:"backmatter"`
	Editing struct {
		CopyeditingLevel  string `json:"copyediting_level"`
		SpecialCharacters string `json:"special_characters"`
		Instructions      string `json:"instructions"`
	} `json:"editing"`
	Design struct {
		Trim       string `json:"trim"`
		EstPages   string `json:"est_pages"`
		Complexity string `json:"complexity"`
	} `json:"design"`
	OtherInstructions string `json:"other_instructions"`
}

func buildTransmittalTextSummary(status string, data *transmittalEmailData) string {
	var b strings.Builder
	b.WriteString("MANUSCRIPT TRANSMITTAL — FINAL\n")
	b.WriteString(strings.Repeat("=", 40) + "\n\n")

	// Book info
	b.WriteString("BOOK INFORMATION\n")
	b.WriteString(strings.Repeat("─", 30) + "\n")
	if data.Book.Title != "" {
		b.WriteString(fmt.Sprintf("Title:     %s\n", data.Book.Title))
	}
	if data.Book.Subtitle != "" {
		b.WriteString(fmt.Sprintf("Subtitle:  %s\n", data.Book.Subtitle))
	}
	if data.Book.Author != "" {
		b.WriteString(fmt.Sprintf("Author:    %s\n", data.Book.Author))
	}
	if data.Book.Publisher != "" {
		b.WriteString(fmt.Sprintf("Publisher: %s\n", data.Book.Publisher))
	}
	if data.Book.Editor != "" {
		b.WriteString(fmt.Sprintf("Editor:    %s\n", data.Book.Editor))
	}
	if data.Book.ISBNPaper != "" {
		b.WriteString(fmt.Sprintf("ISBN (pb):  %s\n", data.Book.ISBNPaper))
	}
	if data.Book.ISBNCloth != "" {
		b.WriteString(fmt.Sprintf("ISBN (hc):  %s\n", data.Book.ISBNCloth))
	}
	b.WriteString("\n")

	// Production
	b.WriteString("PRODUCTION\n")
	b.WriteString(strings.Repeat("─", 30) + "\n")
	if data.Production.TransmittalDate != "" {
		b.WriteString(fmt.Sprintf("Transmittal date: %s\n", data.Production.TransmittalDate))
	}
	if data.Production.MechsDelivery != "" {
		b.WriteString(fmt.Sprintf("Mechs delivery:   %s\n", data.Production.MechsDelivery))
	}
	if data.Production.WeeksInProd != "" {
		b.WriteString(fmt.Sprintf("Weeks in prod:    %s\n", data.Production.WeeksInProd))
	}
	if data.Production.BoundBookDate != "" {
		b.WriteString(fmt.Sprintf("Bound book date:  %s\n", data.Production.BoundBookDate))
	}
	if data.Production.PrintRun != "" {
		b.WriteString(fmt.Sprintf("Print run:        %s\n", data.Production.PrintRun))
	}
	b.WriteString("\n")

	// Manuscript stats
	if data.ChecklistStats.Parts != "" || data.ChecklistStats.Chapters != "" {
		b.WriteString("MANUSCRIPT\n")
		b.WriteString(strings.Repeat("─", 30) + "\n")
		if data.ChecklistStats.Parts != "" {
			b.WriteString(fmt.Sprintf("Parts:     %s\n", data.ChecklistStats.Parts))
		}
		if data.ChecklistStats.Chapters != "" {
			b.WriteString(fmt.Sprintf("Chapters:  %s\n", data.ChecklistStats.Chapters))
		}
		if data.ChecklistStats.WordsChars != "" {
			b.WriteString(fmt.Sprintf("Words:     %s\n", data.ChecklistStats.WordsChars))
		}
		if data.ChecklistStats.MSPP != "" {
			b.WriteString(fmt.Sprintf("MS pages:  %s\n", data.ChecklistStats.MSPP))
		}
		if data.ChecklistStats.EstBookPP != "" {
			b.WriteString(fmt.Sprintf("Est pages: %s\n", data.ChecklistStats.EstBookPP))
		}
		b.WriteString("\n")
	}

	// Checklist
	b.WriteString("COMPONENT CHECKLIST\n")
	b.WriteString(strings.Repeat("─", 30) + "\n")
	for _, item := range data.Checklist {
		check := "☐"
		if item.HereNow {
			check = "☑"
		}
		line := fmt.Sprintf("%s %s", check, item.Component)
		if item.ToComeWhen != "" {
			line += fmt.Sprintf(" (to come: %s)", item.ToComeWhen)
		}
		b.WriteString(line + "\n")
	}
	b.WriteString("\n")

	// Backmatter
	for _, item := range data.Backmatter {
		check := "☐"
		if item.HereNow {
			check = "☑"
		}
		line := fmt.Sprintf("%s %s", check, item.Component)
		if item.ToComeWhen != "" {
			line += fmt.Sprintf(" (to come: %s)", item.ToComeWhen)
		}
		b.WriteString(line + "\n")
	}
	b.WriteString("\n")

	// Design
	if data.Design.Trim != "" || data.Design.EstPages != "" {
		b.WriteString("DESIGN\n")
		b.WriteString(strings.Repeat("─", 30) + "\n")
		if data.Design.Trim != "" {
			b.WriteString(fmt.Sprintf("Trim:       %s\n", data.Design.Trim))
		}
		if data.Design.EstPages != "" {
			b.WriteString(fmt.Sprintf("Est pages:  %s\n", data.Design.EstPages))
		}
		if data.Design.Complexity != "" {
			b.WriteString(fmt.Sprintf("Complexity: %s\n", data.Design.Complexity))
		}
		b.WriteString("\n")
	}

	// Other instructions
	if data.OtherInstructions != "" {
		b.WriteString("OTHER INSTRUCTIONS\n")
		b.WriteString(strings.Repeat("─", 30) + "\n")
		b.WriteString(data.OtherInstructions + "\n\n")
	}

	b.WriteString(strings.Repeat("─", 40) + "\n")
	b.WriteString(fmt.Sprintf("Status: %s\n", strings.ToUpper(status)))
	b.WriteString(fmt.Sprintf("Sent: %s\n", time.Now().Format("2006-01-02 15:04 MST")))

	return b.String()
}

func buildTransmittalHTMLSummary(status string, data *transmittalEmailData, projectURL string) string {
	var b strings.Builder

	// Gmail strips <style> blocks, so all styling is inline (see EMAIL_SYSTEM.md).
	const h2Style = `color:#555;font-size:15px;margin:24px 0 8px;`
	const tableStyle = `width:100%;border-collapse:collapse;margin:8px 0;`
	const labelStyle = `padding:4px 8px;font-size:14px;vertical-align:top;color:#888;width:140px;white-space:nowrap;`
	const valueStyle = `padding:4px 8px;font-size:14px;vertical-align:top;`

	// writeSection emits an <h2> plus a label/value table, skipping empty rows.
	writeSection := func(heading string, rows [][2]string) {
		b.WriteString(fmt.Sprintf(`<h2 style="%s">%s</h2><table style="%s">`, h2Style, heading, tableStyle))
		for _, r := range rows {
			if r[1] != "" {
				b.WriteString(fmt.Sprintf(`<tr><td style="%s">%s</td><td style="%s"><strong>%s</strong></td></tr>`,
					labelStyle, r[0], valueStyle, html.EscapeString(r[1])))
			}
		}
		b.WriteString(`</table>`)
	}

	b.WriteString(`<!DOCTYPE html><html><head><meta charset="utf-8"></head>`)
	b.WriteString(`<body style="margin:0;padding:20px;font-family:-apple-system,Helvetica,Arial,sans-serif;color:#333;">`)
	b.WriteString(`<div style="max-width:600px;margin:0 auto;">`)

	badgeStyle := `display:inline-block;padding:3px 10px;border-radius:12px;font-size:12px;font-weight:600;background:#fef3c7;color:#92400e;`
	if strings.ToLower(status) == "final" {
		badgeStyle = `display:inline-block;padding:3px 10px;border-radius:12px;font-size:12px;font-weight:600;background:#d1fae5;color:#065f46;`
	}

	title := data.Book.Title
	if title == "" {
		title = "Untitled"
	}
	b.WriteString(fmt.Sprintf(`<h1 style="color:#6c63ff;font-size:20px;border-bottom:2px solid #6c63ff;padding-bottom:8px;margin:0 0 12px;">📋 Manuscript Transmittal: %s</h1>`, html.EscapeString(title)))
	b.WriteString(fmt.Sprintf(`<span style="%s">%s</span>`, badgeStyle, html.EscapeString(strings.ToUpper(status))))

	// Book info table
	writeSection("Book Information", [][2]string{
		{"Title", data.Book.Title},
		{"Subtitle", data.Book.Subtitle},
		{"Author", data.Book.Author},
		{"Publisher", data.Book.Publisher},
		{"Editor", data.Book.Editor},
		{"ISBN (paper)", data.Book.ISBNPaper},
		{"ISBN (cloth)", data.Book.ISBNCloth},
	})

	// Production table
	writeSection("Production", [][2]string{
		{"Transmittal date", data.Production.TransmittalDate},
		{"Mechs delivery", data.Production.MechsDelivery},
		{"Weeks in prod", data.Production.WeeksInProd},
		{"Bound book date", data.Production.BoundBookDate},
		{"Print run", data.Production.PrintRun},
	})

	// Manuscript stats (same gate as the text version)
	if data.ChecklistStats.Parts != "" || data.ChecklistStats.Chapters != "" {
		writeSection("Manuscript", [][2]string{
			{"Parts", data.ChecklistStats.Parts},
			{"Chapters", data.ChecklistStats.Chapters},
			{"Words", data.ChecklistStats.WordsChars},
			{"MS pages", data.ChecklistStats.MSPP},
			{"Est pages", data.ChecklistStats.EstBookPP},
		})
	}

	// Checklist
	b.WriteString(fmt.Sprintf(`<h2 style="%s">Component Checklist</h2>`, h2Style))
	for _, item := range data.Checklist {
		color := "#ccc"
		sym := "☐"
		if item.HereNow {
			color = "#10b981"
			sym = "☑"
		}
		extra := ""
		if item.ToComeWhen != "" {
			extra = fmt.Sprintf(` <span style="color:#888">(to come: %s)</span>`, html.EscapeString(item.ToComeWhen))
		}
		b.WriteString(fmt.Sprintf(`<div style="font-size:14px;padding:3px 0;color:%s;">%s %s%s</div>`, color, sym, html.EscapeString(item.Component), extra))
	}
	for _, item := range data.Backmatter {
		color := "#ccc"
		sym := "☐"
		if item.HereNow {
			color = "#10b981"
			sym = "☑"
		}
		extra := ""
		if item.ToComeWhen != "" {
			extra = fmt.Sprintf(` <span style="color:#888">(to come: %s)</span>`, html.EscapeString(item.ToComeWhen))
		}
		b.WriteString(fmt.Sprintf(`<div style="font-size:14px;padding:3px 0;color:%s;">%s %s%s</div>`, color, sym, html.EscapeString(item.Component), extra))
	}

	// Design (same gate as the text version)
	if data.Design.Trim != "" || data.Design.EstPages != "" {
		writeSection("Design", [][2]string{
			{"Trim", data.Design.Trim},
			{"Est pages", data.Design.EstPages},
			{"Complexity", data.Design.Complexity},
		})
	}

	// Other instructions (same gate as the text version)
	if data.OtherInstructions != "" {
		b.WriteString(fmt.Sprintf(`<h2 style="%s">Other Instructions</h2>`, h2Style))
		b.WriteString(fmt.Sprintf(`<p style="font-size:14px;margin:8px 0;">%s</p>`,
			strings.ReplaceAll(html.EscapeString(data.OtherInstructions), "\n", "<br>")))
	}

	// Footer
	b.WriteString(`<div style="margin-top:24px;padding-top:12px;border-top:1px solid #ddd;font-size:12px;color:#999;">`)
	b.WriteString(fmt.Sprintf(`Sent %s`, time.Now().Format("January 2, 2006 at 3:04 PM MST")))
	if projectURL != "" {
		b.WriteString(fmt.Sprintf(` · <a href="%s" style="color:#6c63ff;">View transmittal online</a>`, projectURL))
	}
	b.WriteString(`</div></div></body></html>`)

	return b.String()
}

// ─── Manual-send guardrails (H8) ───

// maxEmailRecipients caps how many addresses a single manual send may target,
// preventing the endpoints from being abused as a spam relay.
const maxEmailRecipients = 10

// Manual sends are rate limited to emailSendLimit per key per emailSendWindow.
const (
	emailSendLimit  = 5
	emailSendWindow = time.Hour
)

// emailRateLimiter throttles the manual email-send handlers, keyed by project
// id or client slug. Modeled on transmittalNotifier.
type emailRateLimiter struct {
	mu   sync.Mutex
	hits map[string][]time.Time // key -> recent send timestamps
}

var emailLimiter = &emailRateLimiter{hits: make(map[string][]time.Time)}

// allow records a send attempt for key and reports whether it stays within the
// rate limit. A false return means the caller MUST reject with 429.
func (l *emailRateLimiter) allow(key string) bool {
	now := time.Now()
	cutoff := now.Add(-emailSendWindow)

	l.mu.Lock()
	defer l.mu.Unlock()

	// Keep only timestamps inside the window (in-place filter).
	recent := l.hits[key][:0]
	for _, t := range l.hits[key] {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}
	if len(recent) >= emailSendLimit {
		l.hits[key] = recent
		return false
	}
	l.hits[key] = append(recent, now)
	return true
}

// enforceEmailSendLimits applies the recipient cap and per-key rate limit to a
// manual email send. It writes a JSON error and returns false when the request
// must be rejected (400 for too many recipients, 429 when rate limited); a true
// return consumes one send slot for key.
func enforceEmailSendLimits(w http.ResponseWriter, key string, recipients []string) bool {
	if len(recipients) > maxEmailRecipients {
		jsonErr(w, fmt.Sprintf("too many recipients (max %d)", maxEmailRecipients), 400)
		return false
	}
	if !emailLimiter.allow(key) {
		jsonErr(w, fmt.Sprintf("rate limit exceeded: at most %d sends per hour", emailSendLimit), 429)
		return false
	}
	return true
}

// ─── HTTP handler ───

func (s *Server) handleSendTransmittalEmail(w http.ResponseWriter, r *http.Request) {
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

	// Parse request: who to send to
	var body struct {
		Recipients []string `json:"recipients"` // email addresses
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, "bad request", 400)
		return
	}
	if len(body.Recipients) == 0 {
		jsonErr(w, "at least one recipient required", 400)
		return
	}

	// Validate email addresses (basic)
	for _, addr := range body.Recipients {
		if !strings.Contains(addr, "@") || !strings.Contains(addr, ".") {
			jsonErr(w, fmt.Sprintf("invalid email: %s", addr), 400)
			return
		}
	}

	if !enforceEmailSendLimits(w, fmt.Sprintf("project:%d", pid), body.Recipients) {
		return
	}

	// Load transmittal data
	var status, dataStr string
	err = s.DB.QueryRowContext(r.Context(),
		`SELECT status, data FROM transmittals WHERE project_id = ?`, pid,
	).Scan(&status, &dataStr)
	if err != nil {
		jsonErr(w, "transmittal not found", 404)
		return
	}

	var txData transmittalEmailData
	if err := json.Unmarshal([]byte(dataStr), &txData); err != nil {
		jsonErr(w, "parse transmittal data: "+err.Error(), 500)
		return
	}

	// Build project URL for the link in email
	var clientSlug, projectSlug string
	_ = s.DB.QueryRowContext(r.Context(),
		`SELECT client_slug, project_slug FROM projects WHERE id = ?`, pid,
	).Scan(&clientSlug, &projectSlug)
	projectURL := fmt.Sprintf("%s/%s/%s/transmittal/", s.BaseURL, clientSlug, projectSlug)

	title := txData.Book.Title
	if title == "" {
		title = "Untitled"
	}
	subject := fmt.Sprintf("Transmittal [%s]: %s", strings.ToUpper(status), title)

	textBody := buildTransmittalTextSummary(status, &txData)
	htmlBody := buildTransmittalHTMLSummary(status, &txData, projectURL)

	// First recipient is the "to", rest are "cc"
	to := []string{body.Recipients[0]}
	var cc []string
	if len(body.Recipients) > 1 {
		cc = body.Recipients[1:]
	}

	if err := s.Email.sendEmail(to, cc, subject, textBody, htmlBody); err != nil {
		slog.Error("send transmittal email", "error", err)
		jsonErr(w, "email send failed", 500)
		return
	}

	jsonOK(w, map[string]any{
		"ok":      true,
		"sent_to": body.Recipients,
		"subject": subject,
	})
}

// handleEmailStatus reports whether email is configured.
func (s *Server) handleEmailStatus(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, map[string]any{
		"configured": s.Email != nil,
	})
}
