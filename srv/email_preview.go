package srv

import (
	"fmt"
	"html"
	"net/http"
	"strings"
	"time"

	"srv.exe.dev/db/dbgen"
)

// ─── Admin email preview: /admin/email-preview/{kind} ───
//
// Renders every outbound template with fixture data so the studio look can be
// checked without sending. Admin-gated; nothing is mailed or logged. The index
// (no kind) lists all templates; ?part=text shows the plain-text part.

type emailPreview struct {
	Kind    string
	Subject string
	Text    string
	HTML    string
}

func emailPreviewFixtures(base string) []emailPreview {
	now := time.Now().UTC()
	fileLog := []fileLogEntry{
		{Direction: "inbound", Filename: "manuscript-v3.docx", FileType: "docx", SentBy: "Author", ReceivedBy: "Jenna", TransferDate: now.AddDate(0, 0, -3).Format("2006-01-02"), Notes: "Chapters 1–9, notes pending"},
		{Direction: "outbound", Filename: "first-pages.pdf", FileType: "pdf", SentBy: "Jenna", ReceivedBy: "Author", TransferDate: now.AddDate(0, 0, -1).Format("2006-01-02")},
	}
	journal := []journalEntry{
		{EntryType: "call", Content: "Walked through the transmittal. Author wants running heads by part, not chapter.", CreatedAt: now.AddDate(0, 0, -4).Format("2006-01-02 15:04")},
		{EntryType: "decision", Content: "Trim confirmed at 5.5 × 8.5. Endnotes, not footnotes.", CreatedAt: now.AddDate(0, 0, -2).Format("2006-01-02 15:04")},
		{EntryType: "approval", Content: "First pages approved.", CreatedAt: now.AddDate(0, 0, -1).Format("2006-01-02 15:04")},
	}
	var tx transmittalEmailData
	tx.Book.Title = "Building in the Wrong Market"
	tx.Book.Subtitle = "Field notes from a decade of near-misses"
	tx.Book.Author = "Mike Casey"
	tx.Book.Publisher = "Portico Advisers"
	tx.Production.TransmittalDate = now.Format("2006-01-02")
	tx.Production.PrintRun = "1,500"
	tx.ChecklistStats.Chapters = "12"
	tx.ChecklistStats.WordsChars = "61,000"
	tx.Editing.CopyeditingLevel = "light"

	pass := dbgen.Pass{ID: 3, ProjectID: 18, CustomerEmail: "mike@example.com", CustomerName: "Mike Casey", BuildsIncluded: 3, BuildsUsed: 1, ExpiresAt: now.AddDate(0, 6, 0)}
	book := dbgen.Book{ID: 41, Title: "Building in the Wrong Market"}
	res := fulfillPassResult{Pass: pass, Title: book.Title, ClientSlug: "mike-casey", ProjectSlug: "casey-001", Password: "otter-lantern-42", PortalURL: base + "/mike-casey/casey-001/factory/"}

	snap := snapshotParams{
		ProjectName: "Building in the Wrong Market", ProjectURL: base + "/mike-casey/casey-001/", Generated: now.Format("January 2, 2006 at 3:04 PM UTC"),
		Tasks: []snapshotTask{
			{Title: "Transmittal", Assignee: "Author", Status: "done", CurrDue: now.AddDate(0, 0, -10).Format("2006-01-02"), ActualDone: now.AddDate(0, 0, -9).Format("2006-01-02"), OrigBudget: 0, CurrBudget: 0},
			{Title: "Copyedit", Assignee: "Jenna", Status: "active", CurrDue: now.AddDate(0, 0, 5).Format("2006-01-02"), OrigBudget: 1200, CurrBudget: 1200, ActualBudget: 600},
			{Title: "First pages", Assignee: "Jenna", Status: "pending", IsMilestone: 1, CurrDue: now.AddDate(0, 0, 21).Format("2006-01-02"), OrigBudget: 900, CurrBudget: 950},
		},
		DoneCount: 1, ActiveCount: 1, PendingCount: 1, TotalTasks: 3, PctComplete: 33,
		TotalOrig: 2100, TotalCurr: 2150, TotalActual: 600, Today: now.Format("2006-01-02"),
		HasTx: true, TxStatus: "draft", TxData: tx, CheckTotal: 14, CheckDone: 9, FileLog: fileLog, Journal: journal,
	}
	act := activityParams{ProjectName: snap.ProjectName, ProjectURL: snap.ProjectURL, Days: 7, Since: now.AddDate(0, 0, -7).Format("Jan 2"), Generated: snap.Generated, FileLog: fileLog, Journal: journal}
	actEmpty := act
	actEmpty.FileLog, actEmpty.Journal = nil, nil
	digest := clientDigestParams{ClientName: "Mike Casey", ClientSlug: "mike-casey", BaseURL: base, Days: 7, Since: act.Since, Generated: snap.Generated,
		Projects: []clientDigestProject{{Name: snap.ProjectName, FileLog: fileLog, Journal: journal}, {Name: "Second Book (working title)"}}, TotalFiles: 2, TotalJournal: 3}

	credits := passCreditsRemaining(pass)
	total := pass.BuildsIncluded + pass.BuildsExtra
	pdfURL, epubURL, reportURL := base+"/api/books/41/download/pdf", base+"/api/books/41/download/epub", base+"/api/projects/18/preflight/report?book_id=41"

	announceBody := "Session 1 is Monday 15:00 UTC on Discord. Before then, redeem your Factory Pass code and fill in the transmittal — even a rough one.\n\nYour code: PYB-XXXX-XXXX"

	return []emailPreview{
		{Kind: "registration_confirm", Subject: "We got your Protocolize Your Book registration", Text: applicantAutoReplyText("Mike Casey"), HTML: applicantAutoReplyHTML("Mike Casey")},
		{Kind: "announcement", Subject: "Protocolize Your Book — redeem your Factory Pass before session 1", Text: announcementText("Mike Casey", announceBody), HTML: announcementHTML("Mike Casey", announceBody)},
		{Kind: "factory_pass", Subject: "Your Factory Pass: " + res.Title, Text: passFulfillmentText(res), HTML: passFulfillmentHTML(res)},
		{Kind: mailKindLoginLink, Subject: "Your sign-in link for Mike Casey", Text: loginLinkText("Mike Casey", base+"/auth/link?t=EXAMPLE-TOKEN-not-valid"), HTML: loginLinkHTML("Mike Casey", base+"/auth/link?t=EXAMPLE-TOKEN-not-valid")},
		{Kind: "template_ready", Subject: "Your template is ready: " + book.Title, Text: templateReadyText(pass, book.Title, res.PortalURL, base+"/api/projects/18/word-template"), HTML: templateReadyHTML(pass, book.Title, res.PortalURL, base+"/api/projects/18/word-template")},
		{Kind: "build_delivered", Subject: "Build ready: " + book.Title, Text: buildDeliveredText(pass, book, "both", pdfURL, epubURL, reportURL, credits, total), HTML: buildDeliveredHTML(pass, book, "both", pdfURL, epubURL, reportURL, credits, total)},
		{Kind: "transmittal", Subject: "Transmittal [DRAFT]: " + tx.Book.Title, Text: buildTransmittalTextSummary("draft", &tx), HTML: buildTransmittalHTMLSummary("draft", &tx, snap.ProjectURL+"transmittal/")},
		{Kind: "transmittal_update", Subject: "Transmittal Updated: " + tx.Book.Title + " (mike-casey)", Text: buildTxNotifyText(snap.ProjectName, "mike-casey", tx.Book.Title, tx.Book.Author, "draft", snap.ProjectURL+"transmittal/"), HTML: buildTxNotifyHTML(snap.ProjectName, "mike-casey", tx.Book.Title, tx.Book.Author, "draft", snap.ProjectURL+"transmittal/")},
		{Kind: "snapshot", Subject: "Project Snapshot: " + snap.ProjectName + " — " + now.Format("Jan 2, 2006"), Text: buildSnapshotText(snap), HTML: buildSnapshotHTML(snap)},
		{Kind: "activity", Subject: "Activity Update: " + act.ProjectName, Text: buildActivityText(act), HTML: buildActivityHTML(act)},
		{Kind: "activity_empty", Subject: "Activity Update: " + act.ProjectName + " (no activity)", Text: buildActivityText(actEmpty), HTML: buildActivityHTML(actEmpty)},
		{Kind: "client_digest", Subject: "Client digest: Mike Casey", Text: buildClientDigestText(digest), HTML: buildClientDigestHTML(digest)},
	}
}

func (s *Server) handleAdminEmailPreview(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdmin(w, r) {
		return
	}
	kind := r.PathValue("kind")
	previews := emailPreviewFixtures(strings.TrimRight(s.BaseURL, "/"))
	if kind == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		var b strings.Builder
		b.WriteString(`<!doctype html><html lang="en"><head><meta charset="utf-8"><title>Email previews · jdbb studio</title><link rel="stylesheet" href="/static/theme.css"><style>.grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(360px,1fr));gap:18px}.card{border-top:1px solid var(--border-strong);padding-top:10px}.card h3{margin:0 0 4px;font-size:14px}.card .k{font:10px var(--mono);letter-spacing:.08em;text-transform:uppercase;color:var(--text-muted)}.card iframe{width:100%;height:420px;border:1px solid var(--border);background:#fff;margin-top:10px}.card .links{font:12px var(--mono);margin-top:6px}.card .links a{margin-right:12px}</style></head><body><div class="jdbb-shell"><header class="jdbb-masthead"><a class="jdbb-wordmark" href="/"><span class="bracket">[</span><span class="kj">j</span>dbb<span class="bracket">]</span><span class="studio">studio</span></a><nav data-admin-nav><a href="/admin/">Admin</a><a href="/admin/#mail">Mail</a><div id="theme-bar"></div></nav></header><main><section class="jdbb-prose"><h1>Email previews</h1><p>Every outbound template rendered with fixture data. Nothing here is sent or logged. Text parts are the primary content; HTML is the wrap.</p></section><div class="grid">`)
		for _, p := range previews {
			fmt.Fprintf(&b, `<div class="card"><div class="k">%s</div><h3>%s</h3><div class="links"><a href="/admin/email-preview/%s" target="_blank">HTML</a><a href="/admin/email-preview/%s?part=text" target="_blank">Text</a></div><iframe sandbox="" src="/admin/email-preview/%s"></iframe></div>`, html.EscapeString(p.Kind), html.EscapeString(p.Subject), p.Kind, p.Kind, p.Kind)
		}
		b.WriteString(`</div></main><footer class="jdbb-footer"><a class="jdbb-wordmark" href="/"><span class="bracket">[</span><span class="kj">j</span>dbb<span class="bracket">]</span></a><nav aria-label="Footer"><a href="/admin/">Admin</a><a href="/">Home</a></nav><span class="copy">&copy; 2026 Jenna Dixon</span></footer></div><script src="/static/theme.js"></script></body></html>`)
		_, _ = w.Write([]byte(b.String()))
		return
	}
	var p *emailPreview
	for i := range previews {
		if previews[i].Kind == kind {
			p = &previews[i]
		}
	}
	if p == nil {
		http.NotFound(w, r)
		return
	}
	if r.URL.Query().Get("part") == "text" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintf(w, "Subject: %s\n\n%s", p.Subject, p.Text)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(p.HTML))
}
