package srv

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEmailNotConfigured(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	// Create project
	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name": "Email Test", "client_slug": "em", "project_slug": "test", "start_date": "2025-01-01",
	})
	var project map[string]any
	decodeJSON(t, resp, &project)
	pid := itoa(int64(project["ID"].(float64)))

	// All email endpoints should return 503 when email not configured
	endpoints := []string{
		"/api/projects/" + pid + "/transmittal/email",
		"/api/projects/" + pid + "/snapshot/email",
		"/api/projects/" + pid + "/activity/email",
	}

	for _, ep := range endpoints {
		resp = apiRequestAdmin(t, ts, "POST", ep, map[string]any{
			"recipients": []string{"test@example.com"},
		})
		if resp.StatusCode != 503 {
			t.Errorf("%s: expected 503 (email not configured), got %d", ep, resp.StatusCode)
		}
		resp.Body.Close()
	}
}

func TestEmailRecipientValidation(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()

	// Configure a mock email (we won't actually send)
	s.Email = &EmailConfig{APIKey: "fake", InboxID: "fake"}

	// Create project with transmittal
	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name": "Validation Test", "client_slug": "val", "project_slug": "test", "start_date": "2025-01-01",
	})
	var project map[string]any
	decodeJSON(t, resp, &project)
	pid := itoa(int64(project["ID"].(float64)))

	// No recipients
	resp = apiRequestAdmin(t, ts, "POST", "/api/projects/"+pid+"/snapshot/email", map[string]any{
		"recipients": []string{},
	})
	if resp.StatusCode != 400 {
		t.Errorf("expected 400 for no recipients, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Invalid email (no @)
	resp = apiRequestAdmin(t, ts, "POST", "/api/projects/"+pid+"/snapshot/email", map[string]any{
		"recipients": []string{"notanemail"},
	})
	if resp.StatusCode != 400 {
		t.Errorf("expected 400 for invalid email, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Invalid email (no .)
	resp = apiRequestAdmin(t, ts, "POST", "/api/projects/"+pid+"/snapshot/email", map[string]any{
		"recipients": []string{"test@localhost"},
	})
	if resp.StatusCode != 400 {
		t.Errorf("expected 400 for email without dot, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

// --- Email content builder tests ---

func TestBuildSnapshotHTML_WithFileLog(t *testing.T) {
	p := snapshotParams{
		ProjectName: "Test Project",
		ProjectURL:  "https://example.com/project",
		Generated:   "January 1, 2025",
		FileLog: []fileLogEntry{
			{Direction: "inbound", Filename: "manuscript.docx", FileType: "Word", SentBy: "Author", ReceivedBy: "Editor", TransferDate: "2025-01-15"},
			{Direction: "outbound", Filename: "edits.docx", FileType: "Word", SentBy: "Editor", ReceivedBy: "Author", TransferDate: "2025-01-20"},
		},
		Journal: []journalEntry{
			{EntryType: "call", Content: "Discussed timeline", CreatedAt: "2025-01-16 10:00:00"},
			{EntryType: "decision", Content: "Moved deadline to March", CreatedAt: "2025-01-17 14:30:00"},
		},
	}

	html := buildSnapshotHTML(p)

	// Check file log section exists
	if !strings.Contains(html, "Recent Files") {
		t.Error("expected 'Recent Files' section in HTML")
	}
	if !strings.Contains(html, "manuscript.docx") {
		t.Error("expected filename 'manuscript.docx' in HTML")
	}
	if !strings.Contains(html, "↓ In") {
		t.Error("expected inbound indicator in HTML")
	}
	if !strings.Contains(html, "↑ Out") {
		t.Error("expected outbound indicator in HTML")
	}

	// Check journal section exists
	if !strings.Contains(html, "Recent Journal") {
		t.Error("expected 'Recent Journal' section in HTML")
	}
	if !strings.Contains(html, "Discussed timeline") {
		t.Error("expected journal content in HTML")
	}
	if !strings.Contains(html, "📞") {
		t.Error("expected call emoji in HTML")
	}
	if !strings.Contains(html, "⚖️") {
		t.Error("expected decision emoji in HTML")
	}
}

func TestBuildSnapshotHTML_EmptyFileLogAndJournal(t *testing.T) {
	p := snapshotParams{
		ProjectName: "Empty Project",
		ProjectURL:  "https://example.com",
		Generated:   "January 1, 2025",
		FileLog:     nil,
		Journal:     nil,
	}

	html := buildSnapshotHTML(p)

	// Should NOT contain these sections when empty
	if strings.Contains(html, "Recent Files") {
		t.Error("should not have 'Recent Files' section when empty")
	}
	if strings.Contains(html, "Recent Journal") {
		t.Error("should not have 'Recent Journal' section when empty")
	}
}

func TestBuildSnapshotText_WithFileLog(t *testing.T) {
	p := snapshotParams{
		ProjectName: "Test Project",
		Generated:   "2025-01-01 10:00 UTC",
		FileLog: []fileLogEntry{
			{Direction: "inbound", Filename: "test.pdf", TransferDate: "2025-01-15"},
		},
		Journal: []journalEntry{
			{EntryType: "note", Content: "A note", CreatedAt: "2025-01-16 10:00:00"},
		},
	}

	text := buildSnapshotText(p)

	if !strings.Contains(text, "RECENT FILES") {
		t.Error("expected 'RECENT FILES' section in text")
	}
	if !strings.Contains(text, "test.pdf") {
		t.Error("expected filename in text")
	}
	if !strings.Contains(text, "RECENT JOURNAL") {
		t.Error("expected 'RECENT JOURNAL' section in text")
	}
	if !strings.Contains(text, "A note") {
		t.Error("expected journal content in text")
	}
}

func TestBuildActivityHTML_NoActivity(t *testing.T) {
	p := activityParams{
		ProjectName: "Empty Activity",
		Days:        7,
		FileLog:     nil,
		Journal:     nil,
	}

	html := buildActivityHTML(p)

	if !strings.Contains(html, "No activity in the last 7 days") {
		t.Error("expected 'no activity' message in HTML")
	}
}

func TestBuildActivityHTML_WithActivity(t *testing.T) {
	p := activityParams{
		ProjectName: "Active Project",
		Days:        7,
		FileLog: []fileLogEntry{
			{Filename: "chapter1.docx", Direction: "inbound"},
		},
		Journal: []journalEntry{
			{EntryType: "approval", Content: "Cover approved"},
		},
	}

	html := buildActivityHTML(p)

	if strings.Contains(html, "No activity") {
		t.Error("should not show 'no activity' when there is activity")
	}
	if !strings.Contains(html, "chapter1.docx") {
		t.Error("expected filename in HTML")
	}
	if !strings.Contains(html, "Cover approved") {
		t.Error("expected journal content in HTML")
	}
	if !strings.Contains(html, "✅") {
		t.Error("expected approval emoji in HTML")
	}
}

func TestBuildActivityText_NoActivity(t *testing.T) {
	p := activityParams{
		ProjectName: "Empty",
		Days:        14,
		FileLog:     nil,
		Journal:     nil,
	}

	text := buildActivityText(p)

	if !strings.Contains(text, "No activity in the last 14 days") {
		t.Error("expected 'no activity' message in text")
	}
}

func TestActivityJournalEmoji(t *testing.T) {
	tests := []struct {
		entryType string
		want      string
	}{
		{"call", "📞"},
		{"decision", "⚖️"},
		{"approval", "✅"},
		{"note", "📝"},
		{"unknown", "📝"}, // default
	}

	for _, tt := range tests {
		got := activityJournalEmoji(tt.entryType)
		if got != tt.want {
			t.Errorf("activityJournalEmoji(%q) = %q, want %q", tt.entryType, got, tt.want)
		}
	}
}

// --- Transmittal builder parity (text vs HTML) ---

func TestTransmittalTextHTMLParity(t *testing.T) {
	var data transmittalEmailData
	if err := json.Unmarshal([]byte(`{
		"book": {"title": "The <Great> Book", "subtitle": "A Tale", "author": "A. Author",
			"publisher": "Pub House", "editor": "E. Editor", "isbn_paper": "978-1", "isbn_cloth": "978-2"},
		"production": {"transmittal_date": "2026-01-15", "mechs_delivery": "2026-03-01",
			"weeks_in_production": "10", "bound_book_date": "2026-06-01", "print_run": "5000"},
		"checklist_stats": {"parts": "3", "chapters": "12", "words_chars": "90,000", "ms_pp": "310", "est_book_pp": "288"},
		"checklist": [{"component": "Front matter", "here_now": true, "to_come_when": ""}],
		"backmatter": [{"component": "Index", "here_now": false, "to_come_when": "with pages"}],
		"design": {"trim": "6 x 9", "est_pages": "288", "complexity": "simple"},
		"other_instructions": "Handle <b>with care</b>"
	}`), &data); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	text := buildTransmittalTextSummary("final", &data)
	htmlOut := buildTransmittalHTMLSummary("final", &data, "https://example.com/c/p/transmittal/")

	// Every section heading in the text version must have an HTML counterpart.
	sections := [][2]string{
		{"BOOK INFORMATION", "Book Information"},
		{"PRODUCTION", "Production"},
		{"MANUSCRIPT", "Manuscript"},
		{"COMPONENT CHECKLIST", "Component Checklist"},
		{"DESIGN", "Design"},
		{"OTHER INSTRUCTIONS", "Other Instructions"},
	}
	for _, s := range sections {
		if !strings.Contains(text, s[0]) {
			t.Errorf("text summary missing section %q", s[0])
		}
		if !strings.Contains(htmlOut, ">"+s[1]+"</h2>") {
			t.Errorf("HTML summary missing section heading %q", s[1])
		}
	}

	// Gmail strips <style> blocks — styling must be inline only.
	if strings.Contains(htmlOut, "<style") {
		t.Error("HTML summary contains a <style> block; all styles must be inline")
	}

	// Interpolated values must stay escaped.
	if strings.Contains(htmlOut, "<Great>") || !strings.Contains(htmlOut, "&lt;Great&gt;") {
		t.Error("book title not HTML-escaped")
	}
	if strings.Contains(htmlOut, "<b>with care</b>") {
		t.Error("other instructions not HTML-escaped")
	}
}

// --- snapshotFormatDate fallback ---

func TestSnapshotFormatDateFallback(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"2026-03-05", "Mar 5, 2026"},
		{"2026-03-05T14:30:00", "Mar 5, 2026"},
		{"", "—"},
		{"not-a-date", "not-a-date"},
		// Malformed input must not carry markup into HTML call sites.
		{`<img src=x onerror=alert(1)>`, "&lt;img src=x onerror=alert(1)&gt;"},
	}
	for _, tt := range tests {
		got := snapshotFormatDate(tt.in)
		if got != tt.want {
			t.Errorf("snapshotFormatDate(%q) = %q, want %q", tt.in, got, tt.want)
		}
		if strings.ContainsAny(got, "<>") {
			t.Errorf("snapshotFormatDate(%q) = %q contains raw angle brackets", tt.in, got)
		}
	}
}
