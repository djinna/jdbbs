package srv

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"log/slog"
	"strings"
	"sync"
	"time"
)

// transmittalNotifier sends a throttled email when a client updates a transmittal.
// At most one notification per project per 30 minutes.
type transmittalNotifier struct {
	mu       sync.Mutex
	lastSent map[int64]time.Time // projectID -> last notification time
}

var txNotifier = &transmittalNotifier{
	lastSent: make(map[int64]time.Time),
}

const txNotifyThrottle = 30 * time.Minute
const txNotifyRecipient = "j@djinna.com"

// maybeNotify checks throttle and sends a notification email in the background.
func (n *transmittalNotifier) maybeNotify(s *Server, projectID int64) {
	if s.Email == nil {
		return
	}

	n.mu.Lock()
	last, ok := n.lastSent[projectID]
	if ok && time.Since(last) < txNotifyThrottle {
		n.mu.Unlock()
		return
	}
	n.lastSent[projectID] = time.Now()
	n.mu.Unlock()

	go n.send(s, projectID)
}

func (n *transmittalNotifier) send(s *Server, projectID int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Load project info
	var projName, clientSlug, projectSlug string
	err := s.DB.QueryRowContext(ctx,
		`SELECT name, client_slug, project_slug FROM projects WHERE id = ?`, projectID,
	).Scan(&projName, &clientSlug, &projectSlug)
	if err != nil {
		slog.Error("transmittal notify: load project", "error", err, "project_id", projectID)
		return
	}

	// Load transmittal data
	var status, dataStr string
	err = s.DB.QueryRowContext(ctx,
		`SELECT status, data FROM transmittals WHERE project_id = ?`, projectID,
	).Scan(&status, &dataStr)
	if err != nil {
		slog.Error("transmittal notify: load transmittal", "error", err, "project_id", projectID)
		return
	}

	var txData struct {
		Book struct {
			Author   string `json:"author"`
			Title    string `json:"title"`
			Subtitle string `json:"subtitle"`
		} `json:"book"`
		Production struct {
			TransmittalDate string `json:"transmittal_date"`
		} `json:"production"`
		Editing struct {
			CopyeditingLevel string `json:"copyediting_level"`
			Instructions     string `json:"instructions"`
		} `json:"editing"`
	}
	_ = json.Unmarshal([]byte(dataStr), &txData)

	projectURL := fmt.Sprintf("%s/%s/%s/transmittal/", s.BaseURL, clientSlug, projectSlug)

	bookTitle := txData.Book.Title
	if bookTitle == "" {
		bookTitle = projName
	}

	subject := fmt.Sprintf("📋 Transmittal Updated: %s (%s)", bookTitle, clientSlug)

	textBody := buildTxNotifyText(projName, clientSlug, bookTitle, txData.Book.Author, status, projectURL)
	htmlBody := buildTxNotifyHTML(projName, clientSlug, bookTitle, txData.Book.Author, status, projectURL)

	if err := s.Email.sendEmail([]string{txNotifyRecipient}, nil, subject, textBody, htmlBody); err != nil {
		slog.Error("transmittal notify: send email", "error", err, "project_id", projectID)
		// Clear throttle so it retries next time
		n.mu.Lock()
		delete(n.lastSent, projectID)
		n.mu.Unlock()
		return
	}

	slog.Info("transmittal update notification sent",
		"to", txNotifyRecipient,
		"project", projName,
		"client", clientSlug,
		"book", bookTitle,
	)
}

func buildTxNotifyText(projName, clientSlug, bookTitle, author, status, url string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("TRANSMITTAL UPDATED: %s\n", strings.ToUpper(bookTitle)))
	b.WriteString(strings.Repeat("=", 40) + "\n\n")
	b.WriteString(fmt.Sprintf("Project:  %s\n", projName))
	b.WriteString(fmt.Sprintf("Client:   %s\n", clientSlug))
	if author != "" {
		b.WriteString(fmt.Sprintf("Author:   %s\n", author))
	}
	b.WriteString(fmt.Sprintf("Status:   %s\n", strings.ToUpper(status)))
	b.WriteString(fmt.Sprintf("Time:     %s\n\n", time.Now().Format("2006-01-02 15:04 MST")))
	b.WriteString(fmt.Sprintf("View: %s\n", url))
	return b.String()
}

func buildTxNotifyHTML(projName, clientSlug, bookTitle, author, status, url string) string {
	var b strings.Builder

	badgeBg := "#fef3c7"
	badgeColor := "#92400e"
	if strings.ToLower(status) == "final" {
		badgeBg = "#d1fae5"
		badgeColor = "#065f46"
	}

	b.WriteString(`<!DOCTYPE html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width"></head>`)
	b.WriteString(`<body style="margin:0;padding:0;background:#f4f3f9;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;color:#333;">`)
	b.WriteString(`<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background:#f4f3f9;"><tr><td align="center" style="padding:24px 12px;">`)
	b.WriteString(`<table role="presentation" width="560" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:12px;overflow:hidden;box-shadow:0 2px 12px rgba(108,99,255,0.08);">`)

	// Header
	b.WriteString(`<tr><td style="background:linear-gradient(135deg,#6c63ff 0%,#8b83ff 100%);padding:28px 32px;">`)
	b.WriteString(`<p style="margin:0 0 4px;font-size:13px;color:rgba(255,255,255,0.7);">📋 Transmittal Update Notification</p>`)
	b.WriteString(fmt.Sprintf(`<h1 style="margin:0;font-size:22px;font-weight:700;color:#ffffff;">%s</h1>`, html.EscapeString(bookTitle)))
	b.WriteString(`</td></tr>`)

	// Body
	b.WriteString(`<tr><td style="padding:28px 32px;">`)

	// Info table
	b.WriteString(`<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="margin-bottom:20px;">`)
	b.WriteString(fmt.Sprintf(`<tr><td style="padding:6px 0;font-size:13px;color:#888;width:100px;">Project</td><td style="padding:6px 0;font-size:14px;font-weight:500;color:#333;">%s</td></tr>`, html.EscapeString(projName)))
	b.WriteString(fmt.Sprintf(`<tr><td style="padding:6px 0;font-size:13px;color:#888;">Client</td><td style="padding:6px 0;font-size:14px;color:#333;">%s</td></tr>`, html.EscapeString(clientSlug)))
	if author != "" {
		b.WriteString(fmt.Sprintf(`<tr><td style="padding:6px 0;font-size:13px;color:#888;">Author</td><td style="padding:6px 0;font-size:14px;color:#333;">%s</td></tr>`, html.EscapeString(author)))
	}
	b.WriteString(fmt.Sprintf(`<tr><td style="padding:6px 0;font-size:13px;color:#888;">Status</td><td style="padding:6px 0;"><span style="display:inline-block;padding:2px 10px;border-radius:10px;font-size:12px;font-weight:600;background:%s;color:%s;">%s</span></td></tr>`,
		badgeBg, badgeColor, strings.ToUpper(status)))
	b.WriteString(fmt.Sprintf(`<tr><td style="padding:6px 0;font-size:13px;color:#888;">Updated</td><td style="padding:6px 0;font-size:14px;color:#333;">%s</td></tr>`, time.Now().Format("January 2, 2006 at 3:04 PM MST")))
	b.WriteString(`</table>`)

	// CTA button
	b.WriteString(fmt.Sprintf(`<a href="%s" style="display:inline-block;padding:10px 24px;background:#6c63ff;color:#ffffff;text-decoration:none;border-radius:6px;font-size:14px;font-weight:600;">View Transmittal →</a>`, url))

	b.WriteString(`</td></tr>`)

	// Footer
	b.WriteString(`<tr><td style="padding:16px 32px;border-top:1px solid #eee;">`)
	b.WriteString(`<p style="margin:0;font-size:12px;color:#aaa;">This is an automated notification from ProdCal. You receive this when a client updates a manuscript transmittal form.</p>`)
	b.WriteString(`</td></tr>`)

	b.WriteString(`</table></td></tr></table>`)
	b.WriteString(`</body></html>`)

	return b.String()
}
