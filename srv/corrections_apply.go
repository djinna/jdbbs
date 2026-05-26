package srv

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// correctionPair is the find/replace subset of a correction row — enough to
// patch a docx (via the python script) or a metadata string (in-process).
type correctionPair struct {
	Find    string
	Replace string
	Chapter string
	Note    string
}

// correctionsScriptPath resolves to typesetting/scripts/apply-corrections-docx.py.
func correctionsScriptPath() string {
	return filepath.Join(typesettingRoot(), "scripts", "apply-corrections-docx.py")
}

// loadPendingCorrectionPairs reads pending corrections for projectID in
// creation order. Returns an empty slice (not nil error) when the project has
// none.
func (s *Server) loadPendingCorrectionPairs(ctx context.Context, projectID int64) ([]correctionPair, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT find_text, replace_text, COALESCE(chapter, ''), COALESCE(note, '')
		FROM corrections
		WHERE project_id = ? AND status = 'pending'
		ORDER BY created_at ASC, id ASC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []correctionPair
	for rows.Next() {
		var p correctionPair
		if err := rows.Scan(&p.Find, &p.Replace, &p.Chapter, &p.Note); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// renderCorrectionsYAML emits the python-script-compatible ledger format for
// the given pairs. The snapshot is suitable for persisting to
// book_outputs.corrections_snapshot.
func renderCorrectionsYAML(projectID int64, pairs []correctionPair) string {
	var sb strings.Builder
	sb.WriteString("# Auto-materialized pending corrections for project ")
	sb.WriteString(fmt.Sprintf("%d\n", projectID))
	sb.WriteString("corrections:\n")
	for _, p := range pairs {
		sb.WriteString(fmt.Sprintf("  - find: %q\n", p.Find))
		sb.WriteString(fmt.Sprintf("    replace: %q\n", p.Replace))
		if p.Chapter != "" {
			sb.WriteString(fmt.Sprintf("    chapter: %q\n", p.Chapter))
		}
		if p.Note != "" {
			sb.WriteString(fmt.Sprintf("    note: %q\n", p.Note))
		}
	}
	return sb.String()
}

// applyPairsToString runs the find/replace pairs against an arbitrary string,
// in load order, ignoring the chapter filter (metadata strings have no chapter
// context). Used to patch book title/author/etc. before they reach pandoc as
// metadata flags — without this, the EPUB title page renders the unpatched
// values from the books table.
func applyPairsToString(s string, pairs []correctionPair) string {
	for _, p := range pairs {
		if p.Chapter != "" {
			// Chapter-scoped corrections only make sense inside body content.
			continue
		}
		s = strings.ReplaceAll(s, p.Find, p.Replace)
	}
	return s
}

// applyCorrectionsToDocx shells out to apply-corrections-docx.py, writing the
// patched docx to outPath. The script preserves Word run formatting and walks
// body, tables, headers/footers, footnotes, and endnotes.
func applyCorrectionsToDocx(yamlPath, docxPath, outPath string) error {
	cmd := exec.Command("python3", correctionsScriptPath(), yamlPath, docxPath, "-o", outPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("apply-corrections-docx.py: %w\n%s", err, string(out))
	}
	return nil
}

// applyCorrectionsIfAny patches the docx at docxPath in-place with the
// project's pending corrections and returns the YAML snapshot (empty when
// nothing applied) plus the pair list so callers can also patch in-memory
// metadata strings. Failure to patch is logged but does NOT fail the compile.
func (s *Server) applyCorrectionsIfAny(ctx context.Context, bid, projectID int64, tmpDir, docxPath string) (snapshot string, pairs []correctionPair) {
	pairs, err := s.loadPendingCorrectionPairs(ctx, projectID)
	if err != nil {
		slog.Warn("corrections: load failed", "id", bid, "project_id", projectID, "err", err)
		return "", nil
	}
	if len(pairs) == 0 {
		return "", nil
	}

	snapshot = renderCorrectionsYAML(projectID, pairs)
	yamlPath := filepath.Join(tmpDir, "corrections.yaml")
	if err := os.WriteFile(yamlPath, []byte(snapshot), 0644); err != nil {
		slog.Warn("corrections: write yaml failed", "id", bid, "err", err)
		return "", pairs
	}
	outPath := filepath.Join(tmpDir, "input.corrected.docx")
	if err := applyCorrectionsToDocx(yamlPath, docxPath, outPath); err != nil {
		slog.Warn("corrections: patcher failed; compiling from unpatched source", "id", bid, "project_id", projectID, "err", err)
		return snapshot, pairs
	}
	// Pandoc reads from docxPath; renaming on top is the cheapest way to make
	// the patched bytes the canonical source for downstream steps.
	if err := os.Rename(outPath, docxPath); err != nil {
		slog.Warn("corrections: rename corrected docx failed", "id", bid, "err", err)
		return snapshot, pairs
	}
	slog.Info("corrections applied", "id", bid, "project_id", projectID, "count", len(pairs))
	return snapshot, pairs
}
