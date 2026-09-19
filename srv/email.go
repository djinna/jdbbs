package srv

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Outbound mail config — set via environment variables. Two transports:
//
// Resend (preferred; From is our own domain, so SPF/DKIM/DMARC are ours):
//
//	PRODCAL_MAIL_FROM       — sending address, e.g. factory@mail.jdbb.studio (turns Resend on)
//	PRODCAL_RESEND_URL      — API base (default https://resend.int.exe.xyz, the exe.dev
//	                          proxy that injects the key; https://api.resend.com locally)
//	RESEND_API_KEY          — bearer, only needed when not going through the proxy
//
// AgentMail (fallback when PRODCAL_MAIL_FROM is unset; also the archive inbox):
//
//	AGENTMAIL_API_KEY       — Bearer token
//	AGENTMAIL_INBOX_ID      — inbox ID for jdbb@agentmail.to (also the sending address)
//
// Common:
//
//	PRODCAL_MAIL_FROM_NAME  — sender display name (default "jdbb studio"; set empty to disable)
//	PRODCAL_MAIL_REPLY_TO   — Reply-To address (default "j@djinna.com"; set empty to disable)
//	PRODCAL_MAIL_BCC        — audit copy BCC'd on every send, batch or single
//	                          (default "j@djinna.com"; set empty to disable).
//	                          Skipped when that address is already in To/Cc.
type EmailConfig struct {
	Provider string // "resend" | "agentmail" (empty = agentmail, for old tests)
	APIKey   string
	InboxID  string // AgentMail inbox; for Resend, the From address
	FromName string
	ReplyTo  string
	BCC      string
	APIBase  string // tests may override; empty uses the provider's default
}

// From is the sending address as it appears to recipients.
func (cfg *EmailConfig) From() string { return cfg.InboxID }

func LoadEmailConfig() *EmailConfig {
	var provider, key, inbox string
	if from := strings.TrimSpace(os.Getenv("PRODCAL_MAIL_FROM")); from != "" {
		provider, key, inbox = "resend", os.Getenv("RESEND_API_KEY"), from
	} else {
		provider, key, inbox = "agentmail", os.Getenv("AGENTMAIL_API_KEY"), os.Getenv("AGENTMAIL_INBOX_ID")
		if key == "" || inbox == "" {
			return nil
		}
	}
	fromName := "jdbb studio"
	if v, ok := os.LookupEnv("PRODCAL_MAIL_FROM_NAME"); ok {
		fromName = v
	}
	replyTo := "j@djinna.com"
	if v, ok := os.LookupEnv("PRODCAL_MAIL_REPLY_TO"); ok {
		replyTo = v
	}
	bcc := "j@djinna.com"
	if v, ok := os.LookupEnv("PRODCAL_MAIL_BCC"); ok {
		bcc = v
	}
	return &EmailConfig{Provider: provider, APIKey: key, InboxID: inbox, FromName: fromName, ReplyTo: replyTo, BCC: bcc, APIBase: os.Getenv("PRODCAL_RESEND_URL")}
}

// bccFor returns the audit BCC list for a send: the configured address unless
// it is already a visible recipient (e.g. the organizer alert, or a snapshot
// the admin mailed to herself).
func (cfg *EmailConfig) bccFor(to, cc []string) []string {
	b := strings.ToLower(strings.TrimSpace(cfg.BCC))
	if b == "" {
		return nil
	}
	for _, a := range append(append([]string{}, to...), cc...) {
		if strings.ToLower(strings.TrimSpace(a)) == b {
			return nil
		}
	}
	return []string{cfg.BCC}
}

// sendEmail sends via AgentMail API.
// POST https://api.agentmail.to/v0/inboxes/:inbox_id/messages/send
func (cfg *EmailConfig) sendEmail(to []string, cc []string, subject, textBody, htmlBody string) error {
	return cfg.sendEmailWithHeaders(to, cc, subject, textBody, htmlBody, nil)
}

// sendEmailWithHeaders is sendEmail plus extra MIME headers via the AgentMail
// generic "headers" map (e.g. List-Unsubscribe for the recurring digest).
func (cfg *EmailConfig) sendEmailWithHeaders(to []string, cc []string, subject, textBody, htmlBody string, extraHeaders map[string]string) error {
	_, err := cfg.send(to, cc, subject, textBody, htmlBody, extraHeaders)
	return err
}

// send is the raw transport. It returns the HTTP status code (0 on
// transport failure) so the outbound_email log can record it. Server.mail is
// the logging wrapper every handler should call.
func (cfg *EmailConfig) send(to []string, cc []string, subject, textBody, htmlBody string, extraHeaders map[string]string) (int, error) {
	if cfg.Provider == "resend" {
		return cfg.sendResend(to, cc, subject, textBody, htmlBody, extraHeaders)
	}
	base := strings.TrimRight(cfg.APIBase, "/")
	if base == "" {
		base = "https://api.agentmail.to"
	}
	url := fmt.Sprintf("%s/v0/inboxes/%s/messages/send", base, cfg.InboxID)

	body := map[string]any{
		"to":      to,
		"subject": subject,
	}
	if len(cc) > 0 {
		body["cc"] = cc
	}
	if bcc := cfg.bccFor(to, cc); len(bcc) > 0 {
		body["bcc"] = bcc
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

	return cfg.post(url, body, to, cc, subject)
}

// sendResend is the Resend transport: POST {base}/emails. Same shape as
// AgentMail's except From is explicit (our domain) and Reply-To is a list.
// Through the exe.dev proxy no bearer is needed; APIKey covers direct use.
func (cfg *EmailConfig) sendResend(to []string, cc []string, subject, textBody, htmlBody string, extraHeaders map[string]string) (int, error) {
	base := strings.TrimRight(cfg.APIBase, "/")
	if base == "" {
		base = "https://resend.int.exe.xyz"
	}
	from := cfg.InboxID
	if cfg.FromName != "" {
		// Resend wants the plain `Name <addr>` form; a Go-quoted name came
		// through to Gmail as a bare address.
		from = fmt.Sprintf("%s <%s>", cfg.FromName, cfg.InboxID)
	}
	body := map[string]any{
		"from":    from,
		"to":      to,
		"subject": subject,
	}
	if len(cc) > 0 {
		body["cc"] = cc
	}
	if bcc := cfg.bccFor(to, cc); len(bcc) > 0 {
		body["bcc"] = bcc
	}
	if textBody != "" {
		body["text"] = textBody
	}
	if htmlBody != "" {
		body["html"] = htmlBody
	}
	if cfg.ReplyTo != "" {
		body["reply_to"] = []string{cfg.ReplyTo}
	}
	if len(extraHeaders) > 0 {
		body["headers"] = extraHeaders
	}
	return cfg.post(base+"/emails", body, to, cc, subject)
}

func (cfg *EmailConfig) post(url string, body map[string]any, to, cc []string, subject string) (int, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return 0, fmt.Errorf("marshal email body: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		// Log the full response body server-side only. Handlers surface this
		// error's message to end users, so keep the body out of it.
		slog.Error("mail API error", "provider", cfg.Provider, "status", resp.StatusCode, "body", string(respBody), "to", to, "subject", subject)
		return resp.StatusCode, fmt.Errorf("mail API error: status %d", resp.StatusCode)
	}

	slog.Info("email sent", "to", to, "cc", cc, "subject", subject, "status", resp.StatusCode)
	return resp.StatusCode, nil
}

// ─── Outbound email log ───

// Email kinds (template names) recorded in outbound_email.kind.
const (
	mailKindRegistrationConfirm = "registration_confirm"
	mailKindRegistrationAlert   = "registration_admin_alert"
	mailKindAnnouncement        = "announcement"
	mailKindFactoryPass         = "factory_pass"
	mailKindClientPassword      = "client_password"
	mailKindBuildDelivered      = "build_delivered"
	mailKindTemplateReady       = "template_ready"
	mailKindTransmittalUpdate   = "transmittal_update"
	mailKindTransmittal         = "transmittal"
	mailKindSnapshot            = "snapshot"
	mailKindActivity            = "activity"
	mailKindClientDigest        = "client_digest"
)

// mailMeta describes a send for the audit log: which template, what record it
// concerns, and who caused it (admin email header, "client", "public", or
// "system").
type mailMeta struct {
	Kind        string
	RefType     string
	RefID       string
	TriggeredBy string
	Headers     map[string]string
}

// mailRef formats an int64 id for mailMeta.RefID.
func mailRef(id int64) string { return strconv.FormatInt(id, 10) }

// triggeredBy derives the actor label from the request: the exe.dev admin
// email when present, otherwise the supplied fallback ("client" or "public").
func triggeredBy(r *http.Request, fallback string) string {
	if r == nil {
		return fallback
	}
	if e := strings.ToLower(strings.TrimSpace(r.Header.Get("X-ExeDev-Email"))); e != "" {
		return e
	}
	if r.Header.Get("X-ExeDev-UserID") != "" {
		return "admin"
	}
	return fallback
}

// mail sends via AgentMail and records the attempt — success or failure — in
// outbound_email. It is the single path every handler should use so the admin
// Mail report is complete. Returns the transport error unchanged.
func (s *Server) mail(meta mailMeta, to []string, cc []string, subject, textBody, htmlBody string) error {
	if s.Email == nil {
		return errors.New("email not configured")
	}
	status, err := s.Email.send(to, cc, subject, textBody, htmlBody, meta.Headers)
	s.logOutboundEmail(meta, to, cc, s.Email.bccFor(to, cc), subject, textBody, htmlBody, status, err)
	return err
}

func (s *Server) logOutboundEmail(meta mailMeta, to, cc, bcc []string, subject, textBody, htmlBody string, status int, sendErr error) {
	if s.DB == nil {
		return
	}
	errText := ""
	if sendErr != nil {
		errText = sendErr.Error()
	}
	if meta.TriggeredBy == "" {
		meta.TriggeredBy = "system"
	}
	_, err := s.DB.Exec(`
		INSERT INTO outbound_email (to_addrs, cc_addrs, bcc_addrs, subject, kind, ref_type, ref_id, triggered_by, status_code, error, text_body, html_body)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		strings.Join(to, ", "), strings.Join(cc, ", "), strings.Join(bcc, ", "), subject,
		meta.Kind, meta.RefType, meta.RefID, meta.TriggeredBy, status, errText, textBody, htmlBody)
	if err != nil {
		slog.Error("outbound_email log insert failed", "err", err, "kind", meta.Kind, "to", to)
	}
}

type outboundEmailRow struct {
	ID          int64  `json:"id"`
	SentAt      string `json:"sent_at"`
	To          string `json:"to"`
	CC          string `json:"cc"`
	BCC         string `json:"bcc"`
	Subject     string `json:"subject"`
	Kind        string `json:"kind"`
	RefType     string `json:"ref_type"`
	RefID       string `json:"ref_id"`
	TriggeredBy string `json:"triggered_by"`
	StatusCode  int    `json:"status_code"`
	Error       string `json:"error"`
	Note        string `json:"note"`
	HasBody     bool   `json:"has_body"`
}

// handleAdminGetOutboundEmail — GET /api/admin/email/{id}: one log row with
// the exact text and HTML bodies that were sent (empty if predating the log).
func (s *Server) handleAdminGetOutboundEmail(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	var m outboundEmailRow
	var text, htmlBody string
	err = s.DB.QueryRowContext(r.Context(), `
		SELECT id, sent_at, to_addrs, cc_addrs, bcc_addrs, subject, kind, ref_type, ref_id, triggered_by, status_code, error, note, text_body, html_body
		FROM outbound_email WHERE id = ?`, id).Scan(
		&m.ID, &m.SentAt, &m.To, &m.CC, &m.BCC, &m.Subject, &m.Kind, &m.RefType, &m.RefID, &m.TriggeredBy, &m.StatusCode, &m.Error, &m.Note, &text, &htmlBody)
	if err != nil {
		jsonErr(w, "not found", 404)
		return
	}
	m.HasBody = text != "" || htmlBody != ""
	jsonOK(w, map[string]any{"email": m, "text_body": text, "html_body": htmlBody})
}

// handleAdminListOutboundEmail — GET /api/admin/email?kind=&to=&ref_type=&ref_id=&since=&failed=1&limit=&offset=
// Read-only audit of every send attempt, newest first, paginated.
func (s *Server) handleAdminListOutboundEmail(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	q := r.URL.Query()
	where := []string{"1=1"}
	args := []any{}
	if v := strings.TrimSpace(q.Get("kind")); v != "" {
		where = append(where, "kind = ?")
		args = append(args, v)
	}
	if v := strings.TrimSpace(q.Get("to")); v != "" {
		where = append(where, "(to_addrs LIKE ? OR cc_addrs LIKE ?)")
		args = append(args, "%"+v+"%", "%"+v+"%")
	}
	if v := strings.TrimSpace(q.Get("ref_type")); v != "" {
		where = append(where, "ref_type = ?")
		args = append(args, v)
	}
	if v := strings.TrimSpace(q.Get("ref_id")); v != "" {
		where = append(where, "ref_id = ?")
		args = append(args, v)
	}
	if v := strings.TrimSpace(q.Get("since")); v != "" {
		where = append(where, "sent_at >= ?")
		args = append(args, v)
	}
	if q.Get("failed") == "1" {
		where = append(where, "(status_code = 0 OR status_code >= 300)")
	}
	limit := 50
	if v, err := strconv.Atoi(q.Get("limit")); err == nil && v > 0 && v <= 500 {
		limit = v
	}
	offset := 0
	if v, err := strconv.Atoi(q.Get("offset")); err == nil && v > 0 {
		offset = v
	}
	whereSQL := strings.Join(where, " AND ")
	var total int
	if err := s.DB.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM outbound_email WHERE `+whereSQL, args...).Scan(&total); err != nil {
		jsonErr(w, "count failed", 500)
		return
	}
	args = append(args, limit, offset)
	rows, err := s.DB.QueryContext(r.Context(), `
		SELECT id, sent_at, to_addrs, cc_addrs, bcc_addrs, subject, kind, ref_type, ref_id, triggered_by, status_code, error, note,
		       (text_body <> '' OR html_body <> '')
		FROM outbound_email WHERE `+whereSQL+`
		ORDER BY sent_at DESC, id DESC LIMIT ? OFFSET ?`, args...)
	if err != nil {
		jsonErr(w, "query failed", 500)
		return
	}
	defer rows.Close()
	out := []outboundEmailRow{}
	for rows.Next() {
		var m outboundEmailRow
		if err := rows.Scan(&m.ID, &m.SentAt, &m.To, &m.CC, &m.BCC, &m.Subject, &m.Kind, &m.RefType, &m.RefID, &m.TriggeredBy, &m.StatusCode, &m.Error, &m.Note, &m.HasBody); err != nil {
			jsonErr(w, "scan failed", 500)
			return
		}
		out = append(out, m)
	}
	// Kind counts for the filter dropdown.
	kinds := map[string]int{}
	krows, err := s.DB.QueryContext(r.Context(), `SELECT kind, COUNT(*) FROM outbound_email GROUP BY kind`)
	if err == nil {
		defer krows.Close()
		for krows.Next() {
			var k string
			var n int
			if krows.Scan(&k, &n) == nil {
				kinds[k] = n
			}
		}
	}
	jsonOK(w, map[string]any{"emails": out, "kinds": kinds, "total": total, "limit": limit, "offset": offset})
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
		TargetDate      string `json:"target_date"` // C6: the one schedule field the factory asks for
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
	Typography struct {
		Pairing          string `json:"pairing"`
		Size             string `json:"size"`
		SectionBreak     string `json:"section_break"`
		SectionBreakText string `json:"section_break_text"`
		Paragraphs       string `json:"paragraphs"`
	} `json:"typography"`
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
	if data.Production.TargetDate != "" {
		b.WriteString(fmt.Sprintf("Target date:      %s\n", data.Production.TargetDate))
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

	// Typography choices (always present once the form has been saved)
	if data.Typography.Pairing != "" {
		b.WriteString("TYPOGRAPHY\n")
		b.WriteString(strings.Repeat("─", 30) + "\n")
		for _, kv := range typographyChoiceLines(data) {
			b.WriteString(fmt.Sprintf("%-14s %s\n", kv[0]+":", kv[1]))
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
	// kv emits a section label plus a ledger of label/value rows, skipping empty rows.
	kv := func(heading string, rows [][2]string) {
		var keep [][2]string
		for _, r := range rows {
			if r[1] != "" {
				keep = append(keep, [2]string{r[0], html.EscapeString(r[1])})
			}
		}
		b.WriteString(emailH2(heading))
		if len(keep) > 0 {
			b.WriteString(emailKV(keep))
		}
	}

	title := data.Book.Title
	if title == "" {
		title = "Untitled"
	}

	// Book info table
	kv("Book Information", [][2]string{
		{"Title", data.Book.Title},
		{"Subtitle", data.Book.Subtitle},
		{"Author", data.Book.Author},
		{"Publisher", data.Book.Publisher},
		{"Editor", data.Book.Editor},
		{"ISBN (paper)", data.Book.ISBNPaper},
		{"ISBN (cloth)", data.Book.ISBNCloth},
	})

	// Production table (status row is pre-rendered HTML, so it bypasses kv's escaping)
	b.WriteString(emailH2("Production"))
	prodRows := [][2]string{{"Status", emailStatus(status)}}
	for _, r := range [][2]string{
		{"Transmittal date", data.Production.TransmittalDate},
		{"Mechs delivery", data.Production.MechsDelivery},
		{"Weeks in prod", data.Production.WeeksInProd},
		{"Bound book date", data.Production.BoundBookDate},
		{"Target date", data.Production.TargetDate},
		{"Print run", data.Production.PrintRun},
	} {
		if r[1] != "" {
			prodRows = append(prodRows, [2]string{r[0], html.EscapeString(r[1])})
		}
	}
	b.WriteString(emailKV(prodRows))

	// Manuscript stats (same gate as the text version)
	if data.ChecklistStats.Parts != "" || data.ChecklistStats.Chapters != "" {
		kv("Manuscript", [][2]string{
			{"Parts", data.ChecklistStats.Parts},
			{"Chapters", data.ChecklistStats.Chapters},
			{"Words", data.ChecklistStats.WordsChars},
			{"MS pages", data.ChecklistStats.MSPP},
			{"Est pages", data.ChecklistStats.EstBookPP},
		})
	}

	// Checklist: one ledger row per component, "here" / "to come" in mono.
	b.WriteString(emailH2("Component Checklist"))
	var rows [][]string
	addRow := func(component string, hereNow bool, toComeWhen string) {
		state := emailStatus("to come")
		if hereNow {
			state = fmt.Sprintf(`<span style="color:%s;font:12px/1.4 %s;letter-spacing:.04em;text-transform:uppercase">here</span>`, emailGreen, emailMono)
		}
		rows = append(rows, []string{html.EscapeString(component), state, html.EscapeString(toComeWhen)})
	}
	for _, item := range data.Checklist {
		addRow(item.Component, item.HereNow, item.ToComeWhen)
	}
	for _, item := range data.Backmatter {
		addRow(item.Component, item.HereNow, item.ToComeWhen)
	}
	if len(rows) > 0 {
		b.WriteString(emailTable([]string{"Component", "Status", "To come"}, rows, nil))
	}

	// Design (same gate as the text version)
	if data.Design.Trim != "" || data.Design.EstPages != "" {
		kv("Design", [][2]string{
			{"Trim", data.Design.Trim},
			{"Est pages", data.Design.EstPages},
			{"Complexity", data.Design.Complexity},
		})
	}
	if data.Typography.Pairing != "" {
		kv("Typography", typographyChoiceLines(data))
	}

	// Other instructions (same gate as the text version)
	if data.OtherInstructions != "" {
		b.WriteString(emailH2("Other Instructions"))
		b.WriteString(emailP(strings.ReplaceAll(html.EscapeString(data.OtherInstructions), "\n", "<br>")))
	}

	if projectURL != "" {
		b.WriteString(emailButton(projectURL, "View transmittal online"))
	}
	b.WriteString(emailSmall(fmt.Sprintf("Sent %s", time.Now().Format("January 2, 2006 at 3:04 PM MST"))))
	b.WriteString(emailSignoff())

	return emailShell(b.String(), emailShellOpts{
		Kicker: "Transmittal · " + strings.ToUpper(status),
		Title:  title,
	})
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
	projectURL := fmt.Sprintf("%s/%s/%s/factory/#transmittal", s.BaseURL, clientSlug, projectSlug)

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

	if err := s.mail(mailMeta{Kind: mailKindTransmittal, RefType: "project", RefID: mailRef(pid), TriggeredBy: triggeredBy(r, "client")}, to, cc, subject, textBody, htmlBody); err != nil {
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

// typographyChoiceLines renders the transmittal's four typographic choices
// as label/value pairs for the final-transmittal email.
func typographyChoiceLines(data *transmittalEmailData) [][2]string {
	t := data.Typography
	pairing := "Studio's choice"
	if t.Pairing != "" && t.Pairing != "studio" {
		p := resolvePairing(t.Pairing)
		pairing = fmt.Sprintf("%s (%s / %s)", p.Name, p.Body, p.Heading)
	}
	or := func(v, def string) string {
		if v == "" {
			return def
		}
		return v
	}
	return [][2]string{
		{"Typeface", pairing},
		{"Text size", or(t.Size, "standard")},
		{"Section break", sectionBreakSummary(t.SectionBreak, t.SectionBreakText)},
		{"Paragraphs", or(t.Paragraphs, "indented")},
	}
}

// sectionBreakSummary names the transmittal's section-break choice for the
// email summary; a custom mark is shown in quotes.
func sectionBreakSummary(choice, text string) string {
	if choice == "custom" && strings.TrimSpace(text) != "" {
		return "custom " + strconv.Quote(strings.TrimSpace(text))
	}
	if choice == "" {
		return "space"
	}
	return choice
}
