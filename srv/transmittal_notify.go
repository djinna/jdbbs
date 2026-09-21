package srv

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"log/slog"
	"net/url"
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

	// The recipient is always the studio admin, so the link goes through the
	// exe.dev proxy's login redirect (adminLoginURL): it lands on the factory
	// page with the admin header set instead of the Factory Pass sign-in gate
	// (Jenna, 2026-09-21 — opened on a phone that wasn't signed in).
	projectURL := adminLoginURL(s.BaseURL, fmt.Sprintf("/%s/%s/factory/#transmittal", clientSlug, projectSlug))

	bookTitle := txData.Book.Title
	if bookTitle == "" {
		bookTitle = projName
	}

	subject := fmt.Sprintf("Transmittal Updated: %s (%s)", bookTitle, clientSlug)

	textBody := buildTxNotifyText(projName, clientSlug, bookTitle, txData.Book.Author, status, projectURL)
	htmlBody := buildTxNotifyHTML(projName, clientSlug, bookTitle, txData.Book.Author, status, projectURL)

	if err := s.mail(mailMeta{Kind: mailKindTransmittalUpdate, RefType: "project", RefID: mailRef(projectID), TriggeredBy: "system"}, []string{txNotifyRecipient}, nil, subject, textBody, htmlBody); err != nil {
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
	rows := [][2]string{
		{"Project", html.EscapeString(projName)},
		{"Client", html.EscapeString(clientSlug)},
	}
	if author != "" {
		rows = append(rows, [2]string{"Author", html.EscapeString(author)})
	}
	rows = append(rows,
		[2]string{"Status", emailStatus(status)},
		[2]string{"Updated", html.EscapeString(time.Now().Format("January 2, 2006 at 3:04 PM MST"))},
	)

	body := emailKV(rows) + emailButton(url, "View transmittal")

	return emailShell(body, emailShellOpts{
		Kicker: "Transmittal updated",
		Title:  bookTitle,
		Footer: "This is an automated notification from jdbb studio. You receive this when a client updates a manuscript transmittal form.",
	})
}

// adminLoginURL wraps a site path in the exe.dev proxy's login redirect
// (https://exe.dev/docs/login-with-exe): the visitor authenticates with
// exe.dev — a near-instant redirect when already signed in on the device —
// and arrives with X-ExeDev-UserID set, so the app's admin bypass applies
// and no client password gate appears. Only for links mailed to the studio
// admin; customer-facing mail keeps plain URLs.
func adminLoginURL(base, path string) string {
	return strings.TrimRight(base, "/") + "/__exe.dev/login?redirect=" + url.QueryEscape(path)
}
