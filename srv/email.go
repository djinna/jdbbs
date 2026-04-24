package srv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

// AgentMail config — set via environment variables:
//
//	AGENTMAIL_API_KEY  — Bearer token
//	AGENTMAIL_INBOX_ID — inbox ID for jdbb@agentmail.to
type EmailConfig struct {
	APIKey  string
	InboxID string
}

func LoadEmailConfig() *EmailConfig {
	key := os.Getenv("AGENTMAIL_API_KEY")
	inbox := os.Getenv("AGENTMAIL_INBOX_ID")
	if key == "" || inbox == "" {
		return nil
	}
	return &EmailConfig{APIKey: key, InboxID: inbox}
}

// sendEmail sends via AgentMail API.
// POST https://api.agentmail.to/v0/inboxes/:inbox_id/messages/send
func (cfg *EmailConfig) sendEmail(to []string, cc []string, subject, textBody, htmlBody string) error {
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
		return fmt.Errorf("agentmail API error %d: %s", resp.StatusCode, string(respBody))
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
		TransmittalDate  string `json:"transmittal_date"`
		MechsDelivery    string `json:"mechs_delivery"`
		WeeksInProd      string `json:"weeks_in_production"`
		BoundBookDate    string `json:"bound_book_date"`
		PrintRun         string `json:"print_run"`
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
	b.WriteString(`<!DOCTYPE html><html><head><style>
		body { font-family: -apple-system, Helvetica, Arial, sans-serif; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
		h1 { color: #6c63ff; font-size: 20px; border-bottom: 2px solid #6c63ff; padding-bottom: 8px; }
		h2 { color: #555; font-size: 15px; margin-top: 24px; margin-bottom: 8px; }
		table { width: 100%; border-collapse: collapse; margin: 8px 0; }
		td { padding: 4px 8px; font-size: 14px; vertical-align: top; }
		td:first-child { color: #888; width: 140px; white-space: nowrap; }
		.check { font-size: 14px; padding: 3px 0; }
		.check-yes { color: #10b981; }
		.check-no { color: #ccc; }
		.footer { margin-top: 24px; padding-top: 12px; border-top: 1px solid #ddd; font-size: 12px; color: #999; }
		.badge { display: inline-block; padding: 3px 10px; border-radius: 12px; font-size: 12px; font-weight: 600; }
		.badge-final { background: #d1fae5; color: #065f46; }
		.badge-draft { background: #fef3c7; color: #92400e; }
		a { color: #6c63ff; }
	</style></head><body>`)

	badgeClass := "badge-draft"
	if strings.ToLower(status) == "final" {
		badgeClass = "badge-final"
	}

	title := data.Book.Title
	if title == "" {
		title = "Untitled"
	}
	b.WriteString(fmt.Sprintf(`<h1>📋 Manuscript Transmittal: %s</h1>`, title))
	b.WriteString(fmt.Sprintf(`<span class="badge %s">%s</span>`, badgeClass, strings.ToUpper(status)))

	// Book info table
	b.WriteString(`<h2>Book Information</h2><table>`)
	rows := [][2]string{
		{"Title", data.Book.Title},
		{"Subtitle", data.Book.Subtitle},
		{"Author", data.Book.Author},
		{"Publisher", data.Book.Publisher},
		{"Editor", data.Book.Editor},
		{"ISBN (paper)", data.Book.ISBNPaper},
		{"ISBN (cloth)", data.Book.ISBNCloth},
	}
	for _, r := range rows {
		if r[1] != "" {
			b.WriteString(fmt.Sprintf(`<tr><td>%s</td><td><strong>%s</strong></td></tr>`, r[0], r[1]))
		}
	}
	b.WriteString(`</table>`)

	// Production table
	b.WriteString(`<h2>Production</h2><table>`)
	prodRows := [][2]string{
		{"Transmittal date", data.Production.TransmittalDate},
		{"Mechs delivery", data.Production.MechsDelivery},
		{"Weeks in prod", data.Production.WeeksInProd},
		{"Bound book date", data.Production.BoundBookDate},
		{"Print run", data.Production.PrintRun},
	}
	for _, r := range prodRows {
		if r[1] != "" {
			b.WriteString(fmt.Sprintf(`<tr><td>%s</td><td><strong>%s</strong></td></tr>`, r[0], r[1]))
		}
	}
	b.WriteString(`</table>`)

	// Checklist
	b.WriteString(`<h2>Component Checklist</h2>`)
	for _, item := range data.Checklist {
		cls := "check-no"
		sym := "☐"
		if item.HereNow {
			cls = "check-yes"
			sym = "☑"
		}
		extra := ""
		if item.ToComeWhen != "" {
			extra = fmt.Sprintf(` <span style="color:#888">(to come: %s)</span>`, item.ToComeWhen)
		}
		b.WriteString(fmt.Sprintf(`<div class="check %s">%s %s%s</div>`, cls, sym, item.Component, extra))
	}
	for _, item := range data.Backmatter {
		cls := "check-no"
		sym := "☐"
		if item.HereNow {
			cls = "check-yes"
			sym = "☑"
		}
		extra := ""
		if item.ToComeWhen != "" {
			extra = fmt.Sprintf(` <span style="color:#888">(to come: %s)</span>`, item.ToComeWhen)
		}
		b.WriteString(fmt.Sprintf(`<div class="check %s">%s %s%s</div>`, cls, sym, item.Component, extra))
	}

	// Footer
	b.WriteString(`<div class="footer">`)
	b.WriteString(fmt.Sprintf(`Sent %s`, time.Now().Format("January 2, 2006 at 3:04 PM MST")))
	if projectURL != "" {
		b.WriteString(fmt.Sprintf(` · <a href="%s">View transmittal online</a>`, projectURL))
	}
	b.WriteString(`</div></body></html>`)

	return b.String()
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
		jsonErr(w, "failed to send email: "+err.Error(), 500)
		return
	}

	jsonOK(w, map[string]any{
		"ok":         true,
		"sent_to":    body.Recipients,
		"subject":    subject,
	})
}

// handleEmailStatus reports whether email is configured.
func (s *Server) handleEmailStatus(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, map[string]any{
		"configured": s.Email != nil,
	})
}
