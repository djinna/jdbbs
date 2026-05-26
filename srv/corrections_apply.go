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

// correctionsScriptPath resolves to typesetting/scripts/apply-corrections-docx.py.
// Matches the typesetting-root convention used in books.go.
func correctionsScriptPath() string {
	return filepath.Join(typesettingRoot(), "scripts", "apply-corrections-docx.py")
}

// materializePendingCorrectionsYAML reads pending corrections for projectID and
// writes them as a YAML ledger to a file inside dir. Returns ("", "", nil) when
// the project has no pending corrections — callers should treat that as "skip
// the patch step." snapshot is the exact YAML written, suitable for persisting
// alongside the generated artifact for lineage.
func (s *Server) materializePendingCorrectionsYAML(ctx context.Context, projectID int64, dir string) (path, snapshot string, err error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT find_text, replace_text, COALESCE(chapter, ''), COALESCE(note, '')
		FROM corrections
		WHERE project_id = ? AND status = 'pending'
		ORDER BY created_at ASC, id ASC
	`, projectID)
	if err != nil {
		return "", "", err
	}
	defer rows.Close()

	var sb strings.Builder
	sb.WriteString("# Auto-materialized pending corrections for project ")
	sb.WriteString(fmt.Sprintf("%d\n", projectID))
	sb.WriteString("corrections:\n")

	count := 0
	for rows.Next() {
		var find, replace, chapter, note string
		if err := rows.Scan(&find, &replace, &chapter, &note); err != nil {
			return "", "", err
		}
		sb.WriteString(fmt.Sprintf("  - find: %q\n", find))
		sb.WriteString(fmt.Sprintf("    replace: %q\n", replace))
		if chapter != "" {
			sb.WriteString(fmt.Sprintf("    chapter: %q\n", chapter))
		}
		if note != "" {
			sb.WriteString(fmt.Sprintf("    note: %q\n", note))
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return "", "", err
	}
	if count == 0 {
		return "", "", nil
	}

	snapshot = sb.String()
	path = filepath.Join(dir, "corrections.yaml")
	if err := os.WriteFile(path, []byte(snapshot), 0644); err != nil {
		return "", "", err
	}
	return path, snapshot, nil
}

// applyCorrectionsToDocx shells out to apply-corrections-docx.py, writing the
// patched docx to outPath. The script preserves Word run formatting and is the
// same tool the manual ledger workflow uses, so the result is identical to a
// hand-applied patch.
func applyCorrectionsToDocx(yamlPath, docxPath, outPath string) error {
	cmd := exec.Command("python3", correctionsScriptPath(), yamlPath, docxPath, "-o", outPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("apply-corrections-docx.py: %w\n%s", err, string(out))
	}
	return nil
}

// applyCorrectionsIfAny applies pending corrections for a project-linked book
// to the docx at docxPath in-place. Returns the snapshot YAML (empty when no
// corrections were applied). Failure to apply corrections is logged but does
// NOT fail the compile — a typo patcher should not block a build.
func (s *Server) applyCorrectionsIfAny(ctx context.Context, bid, projectID int64, tmpDir, docxPath string) string {
	yamlPath, snapshot, err := s.materializePendingCorrectionsYAML(ctx, projectID, tmpDir)
	if err != nil {
		slog.Warn("corrections: materialize failed", "id", bid, "project_id", projectID, "err", err)
		return ""
	}
	if yamlPath == "" {
		return ""
	}
	outPath := filepath.Join(tmpDir, "input.corrected.docx")
	if err := applyCorrectionsToDocx(yamlPath, docxPath, outPath); err != nil {
		slog.Warn("corrections: patcher failed; compiling from unpatched source", "id", bid, "project_id", projectID, "err", err)
		return ""
	}
	// Replace the source path's contents with the corrected docx so downstream
	// (pandoc) sees the patched bytes without needing to know about this step.
	if err := os.Rename(outPath, docxPath); err != nil {
		slog.Warn("corrections: rename corrected docx failed", "id", bid, "err", err)
		return ""
	}
	slog.Info("corrections applied", "id", bid, "project_id", projectID, "bytes_yaml", len(snapshot))
	return snapshot
}
