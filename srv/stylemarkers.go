package srv

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"srv.exe.dev/db/dbgen"
)

// [[style]] markers — the convention for authors without custom paragraph
// styles (Google Docs). `[[style:computer text]]` (or `[[code block]]`) at the
// start of a paragraph restyles it; a `[[/code block]]` / `[[/]]` / `[[end]]`
// at the end of a later paragraph extends the run. The pre-pass
// typesetting/scripts/apply-style-markers.py rewrites the docx in place before
// pandoc, so the print PDF and the EPUB see the same styled paragraphs.
// Inspect (detect-edge-cases.py) reports the same markers as 'style_marker'.

// styleMarkersScriptPath resolves to typesetting/scripts/apply-style-markers.py.
func styleMarkersScriptPath() string {
	return filepath.Join(typesettingRoot(), "scripts", "apply-style-markers.py")
}

// styleMarkerReport is one marker occurrence as the script reports it
// (1-based paragraph numbers, matching Inspect).
type styleMarkerReport struct {
	Marker     string `json:"marker"`
	Resolved   string `json:"resolved"`
	Via        string `json:"via"`
	Start      int    `json:"start"`
	End        int    `json:"end"`
	Paragraphs int    `json:"paragraphs"`
}

// applyStyleMarkers runs the pre-pass on docxPath in place. declaredPath may
// be "" (no transmittal). The returned report lists every marker seen;
// resolved ones have Resolved != "". On error the docx is left as it was
// (the script writes to a temp file and replaces atomically), so callers
// treat failure as non-fatal: log and continue with the original.
func applyStyleMarkers(docxPath, declaredPath string) ([]styleMarkerReport, error) {
	script := styleMarkersScriptPath()
	if _, err := os.Stat(script); err != nil {
		return nil, fmt.Errorf("apply-style-markers.py not found: %w", err)
	}
	reportPath := filepath.Join(filepath.Dir(docxPath), "style-markers.json")
	args := []string{script, docxPath, docxPath, "--report", reportPath}
	if declaredPath != "" {
		args = append(args, "--declared-styles", declaredPath)
	}
	cmd, cancel := toolCommand(context.Background(), toolTimeoutPython, "python3", args...)
	defer cancel()
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("apply-style-markers.py: %w\n%s", toolErr(cmd, err), string(out))
	}
	raw, err := os.ReadFile(reportPath)
	if err != nil {
		return nil, fmt.Errorf("apply-style-markers.py: read report: %w", err)
	}
	var report []styleMarkerReport
	if err := json.Unmarshal(raw, &report); err != nil {
		return nil, fmt.Errorf("apply-style-markers.py: parse report: %w", err)
	}
	return report, nil
}

// summarizeStyleMarkers renders the report as the one-line summary the build
// log carries, e.g. "[[computer text]]→Code Block ¶118; unresolved: [[caption]] ¶179".
func summarizeStyleMarkers(report []styleMarkerReport) (resolved, unresolved int, summary string) {
	var res, unres []string
	for _, r := range report {
		loc := fmt.Sprintf("¶%d", r.Start)
		if r.End != r.Start {
			loc = fmt.Sprintf("¶%d–%d", r.Start, r.End)
		}
		if r.Resolved != "" {
			resolved++
			res = append(res, fmt.Sprintf("[[%s]]→%s %s", r.Marker, r.Resolved, loc))
		} else {
			unresolved++
			unres = append(unres, fmt.Sprintf("[[%s]] %s", r.Marker, loc))
		}
	}
	parts := []string{}
	if len(res) > 0 {
		parts = append(parts, strings.Join(res, ", "))
	}
	if len(unres) > 0 {
		parts = append(parts, "unresolved: "+strings.Join(unres, ", "))
	}
	return resolved, unresolved, strings.Join(parts, "; ")
}

// applyStyleMarkersForBuild is the shared build step: write the declared
// styles from the linked project's spec (if any), run the pre-pass on the
// working docx in place and log the outcome. Never fails the build.
func (s *Server) applyStyleMarkersForBuild(stage string, bid int64, book dbgen.Book, tmpDir, docxPath string) {
	declaredPath := ""
	if book.ProjectID.Valid {
		if spec, err := dbgen.New(s.DB).GetBookSpec(context.Background(), book.ProjectID.Int64); err == nil {
			if p, err := writeDeclaredStylesFile(tmpDir, spec.Data); err == nil {
				declaredPath = p
			} else {
				slog.Warn("style markers: declared styles unavailable", "stage", stage, "book_id", bid, "err", err)
			}
		}
	}
	report, err := applyStyleMarkers(docxPath, declaredPath)
	if err != nil {
		slog.Warn("style markers: pre-pass failed; building from the original docx", "stage", stage, "book_id", bid, "err", err)
		return
	}
	if len(report) == 0 {
		return
	}
	resolved, unresolved, summary := summarizeStyleMarkers(report)
	slog.Info("style markers applied", "stage", stage, "book_id", bid,
		"resolved", resolved, "unresolved", unresolved, "markers", summary)
}
