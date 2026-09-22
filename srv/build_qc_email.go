package srv

import (
	"context"
	"fmt"
	"html"
	"log/slog"
	"os"
	"strings"
	"time"

	"srv.exe.dev/db/dbgen"
)

// Build QC mail (EMAIL_SYSTEM.md pathway #9) — an internal before/after
// notice to Jenna for every factory build, proof or final, succeeded or
// failed. Workshop day 2 (2026-09-22): two attendees had emailed her their
// uploaded .docx and the proof so we could compare; this makes the pair
// arrive on its own. Links, not attachments: the source .docx
// (/download/source), the newest PDF and EPUB, the Inspect report, the
// Floor and the customer's factory page.
//
// Switched on by PRODCAL_BUILD_QC_EMAIL=<recipient> in the environment
// (.env); unset = off. Meant to be on for the workshop week only.

const mailKindBuildQC = "build_qc"

func buildQCRecipient() string {
	return strings.TrimSpace(os.Getenv("PRODCAL_BUILD_QC_EMAIL"))
}

// sendBuildQCEmail runs on the conversion goroutine after the build's
// factory event is recorded. failure is "" for a green build, else the
// customer-facing message.
func (s *Server) sendBuildQCEmail(book dbgen.Book, format string, elapsed time.Duration, failure string) {
	to := buildQCRecipient()
	if to == "" || !book.ProjectID.Valid {
		return
	}
	if s.Email == nil {
		slog.Warn("build QC mail: email not configured", "book_id", book.ID)
		return
	}
	defer func() {
		if rec := recover(); rec != nil {
			slog.Error("build QC email panic", "recover", rec)
		}
	}()

	ctx := context.Background()
	q := dbgen.New(s.DB)
	project, err := q.GetProject(ctx, book.ProjectID.Int64)
	if err != nil {
		slog.Warn("build QC mail: project lookup failed", "err", err, "book_id", book.ID)
		return
	}
	base := strings.TrimRight(s.BaseURL, "/")
	kind := buildKindOf(book)
	links := [][2]string{
		{"Before", fmt.Sprintf("%s/api/books/%d/download/source", base, book.ID)},
		{"After (PDF)", fmt.Sprintf("%s/api/books/%d/download/pdf?kind=%s", base, book.ID, kind)},
		{"After (EPUB)", fmt.Sprintf("%s/api/books/%d/download/epub?kind=%s", base, book.ID, kind)},
		{"Inspect report", fmt.Sprintf("%s/api/projects/%d/preflight/report?book_id=%d", base, project.ID, book.ID)},
		{"Factory page", s.portalURL(project.ClientSlug, project.ProjectSlug)},
		{"Floor", base + "/admin/factory/"},
	}
	if format == "epub" {
		links = append(links[:1], links[2:]...)
	} else if format == "pdf" {
		links = append(links[:2], links[3:]...)
	}

	outcome := fmt.Sprintf("%s %s built in %s", kind, format, elapsed.Round(time.Second))
	subject := fmt.Sprintf("QC · %s · %s (%s %s, book %d)", project.ClientSlug, book.Title, kind, format, book.ID)
	if failure != "" {
		outcome = "build FAILED: " + failure
		subject = fmt.Sprintf("QC · %s · %s (%s %s FAILED, book %d)", project.ClientSlug, book.Title, kind, format, book.ID)
	}

	var tb strings.Builder
	fmt.Fprintf(&tb, "%s\n%s — /%s/%s/\n\nUploaded: %s\n%s\n\n", book.Title, project.Name, project.ClientSlug, project.ProjectSlug, book.SourceFilename, outcome)
	for _, l := range links {
		fmt.Fprintf(&tb, "%s: %s\n", l[0], l[1])
	}
	tb.WriteString("\nInternal QC notice — every factory build this week. Switch off by unsetting PRODCAL_BUILD_QC_EMAIL.\n")

	rows := [][2]string{
		{"Project", fmt.Sprintf("%s · <span style=\"font-family:monospace\">/%s/%s/</span>", html.EscapeString(project.Name), html.EscapeString(project.ClientSlug), html.EscapeString(project.ProjectSlug))},
		{"Uploaded", html.EscapeString(book.SourceFilename)},
		{"Outcome", html.EscapeString(outcome)},
	}
	for _, l := range links {
		rows = append(rows, [2]string{l[0], emailLink(l[1], l[1])})
	}
	var hb strings.Builder
	hb.WriteString(emailKV(rows))
	hb.WriteString(emailSmall("Internal QC notice — every factory build this week (proof and final, green or failed). Switch off by unsetting PRODCAL_BUILD_QC_EMAIL."))
	htmlBody := emailShell(hb.String(), emailShellOpts{Kicker: "Build QC · before / after", Title: book.Title})

	if err := s.mail(mailMeta{Kind: mailKindBuildQC, RefType: "book", RefID: mailRef(book.ID), TriggeredBy: "system"}, []string{to}, nil, subject, tb.String(), htmlBody); err != nil {
		slog.Error("build QC email failed", "err", err, "book_id", book.ID)
	}
}
