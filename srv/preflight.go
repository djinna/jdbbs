package srv

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"srv.exe.dev/db/dbgen"
)

type preflightRequest struct {
	BookID int64 `json:"book_id"`
}

type preflightSummary struct {
	Total  int            `json:"total"`
	High   int            `json:"high"`
	Medium int            `json:"medium"`
	Low    int            `json:"low"`
	ByType map[string]int `json:"by_type"`
}

type preflightHistoryEntry struct {
	ID        int64  `json:"id"`
	Status    string `json:"status,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
	ReportURL string `json:"report_url,omitempty"`
	Latest    bool   `json:"latest,omitempty"`
}

type preflightResponse struct {
	Exists         bool                    `json:"exists"`
	ProjectID      int64                   `json:"project_id,omitempty"`
	BookID         int64                   `json:"book_id,omitempty"`
	Status         string                  `json:"status,omitempty"`
	SourceFilename string                  `json:"source_filename,omitempty"`
	UpdatedAt      string                  `json:"updated_at,omitempty"`
	Summary        *preflightSummary       `json:"summary,omitempty"`
	Images         []map[string]any        `json:"images,omitempty"`
	ReportURL      string                  `json:"report_url,omitempty"`
	History        []preflightHistoryEntry `json:"history,omitempty"`
	Error          string                  `json:"error,omitempty"`
}

func declaredCustomStylesList(specData string) ([]map[string]any, error) {
	if strings.TrimSpace(specData) == "" {
		return []map[string]any{}, nil
	}
	var spec map[string]any
	if err := json.Unmarshal([]byte(specData), &spec); err != nil {
		return nil, err
	}
	styles, ok := spec["custom_styles"].([]any)
	if !ok {
		return []map[string]any{}, nil
	}
	out := make([]map[string]any, 0, len(styles))
	for _, item := range styles {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		out = append(out, m)
	}
	return out, nil
}

func writeDeclaredStylesFile(tmpDir, specData string) (string, error) {
	styles, err := declaredCustomStylesList(specData)
	if err != nil {
		return "", err
	}
	path := filepath.Join(tmpDir, "declared-styles.json")
	body, err := json.Marshal(styles)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, body, 0644); err != nil {
		return "", err
	}
	return path, nil
}

func (s *Server) defaultPreflightRunner(docxPath string, declaredStylesPath string) ([]byte, []byte, error) {
	tmpDir, err := os.MkdirTemp("", "prodcal-preflight-*")
	if err != nil {
		return nil, nil, err
	}
	defer os.RemoveAll(tmpDir)

	htmlPath := filepath.Join(tmpDir, "report.html")
	jsonPath := filepath.Join(tmpDir, "report.json")
	args := []string{filepath.Join(typesettingRoot(), "scripts", "detect-edge-cases.py"), docxPath, "-o", htmlPath, "--json"}
	if declaredStylesPath != "" {
		args = append(args, "--declared-styles", declaredStylesPath)
	}
	cmd := exec.Command("python3", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, nil, fmt.Errorf("python detector failed: %w\n%s", err, string(out))
	}
	htmlBytes, err := os.ReadFile(htmlPath)
	if err != nil {
		return nil, nil, err
	}
	jsonBytes, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, nil, err
	}
	return htmlBytes, jsonBytes, nil
}

func (s *Server) getPreflightRunner() preflightRunnerFunc {
	if s.preflightRunner != nil {
		return s.preflightRunner
	}
	return s.defaultPreflightRunner
}

func buildPreflightSummary(raw []byte) (*preflightSummary, error) {
	var findings []map[string]any
	if len(raw) == 0 {
		return &preflightSummary{ByType: map[string]int{}}, nil
	}
	if err := json.Unmarshal(raw, &findings); err != nil {
		return nil, err
	}
	summary := &preflightSummary{ByType: map[string]int{}}
	for _, item := range findings {
		summary.Total++
		if typ, _ := item["type"].(string); typ != "" {
			summary.ByType[typ]++
		}
		switch sev, _ := item["severity"].(string); sev {
		case "high":
			summary.High++
		case "medium":
			summary.Medium++
		case "low":
			summary.Low++
		}
	}
	return summary, nil
}

func parseStoredSummary(raw string) (*preflightSummary, error) {
	if raw == "" {
		return &preflightSummary{ByType: map[string]int{}}, nil
	}
	var summary preflightSummary
	if err := json.Unmarshal([]byte(raw), &summary); err != nil {
		return nil, err
	}
	if summary.ByType == nil {
		summary.ByType = map[string]int{}
	}
	return &summary, nil
}

func parseStoredImages(raw string) []map[string]any {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var findings []map[string]any
	if err := json.Unmarshal([]byte(raw), &findings); err != nil {
		return nil
	}
	images := make([]map[string]any, 0)
	for _, item := range findings {
		if typ, _ := item["type"].(string); typ == "image_inventory" {
			images = append(images, item)
		}
	}
	return images
}

func normalizeStyleName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func declaredCustomStyles(specData string) (map[string]bool, error) {
	declared := map[string]bool{}
	styles, err := declaredCustomStylesList(specData)
	if err != nil {
		return nil, err
	}
	for _, m := range styles {
		if wordStyle, _ := m["word_style"].(string); normalizeStyleName(wordStyle) != "" {
			declared[normalizeStyleName(wordStyle)] = true
		}
		if name, _ := m["name"].(string); normalizeStyleName(name) != "" {
			declared[normalizeStyleName(name)] = true
		}
	}
	return declared, nil
}

func appendUndeclaredStyleWarnings(report []byte, specData string) ([]byte, error) {
	declared, err := declaredCustomStyles(specData)
	if err != nil {
		return nil, err
	}
	var findings []map[string]any
	if len(report) > 0 {
		if err := json.Unmarshal(report, &findings); err != nil {
			return nil, err
		}
	}
	seenWarnings := map[string]bool{}
	seenDeclared := map[string]bool{}
	for _, item := range findings {
		if typ, _ := item["type"].(string); typ == "undeclared_custom_style" {
			if name, _ := item["style_name"].(string); name != "" {
				seenWarnings[normalizeStyleName(name)] = true
			}
		}
		if typ, _ := item["type"].(string); typ == "declared_custom_style_used" {
			if name, _ := item["style_name"].(string); name != "" {
				seenDeclared[normalizeStyleName(name)] = true
			}
		}
	}
	for _, item := range findings {
		if typ, _ := item["type"].(string); typ != "observed_style" {
			continue
		}
		styleName, _ := item["style_name"].(string)
		styleKind, _ := item["style_kind"].(string)
		normalized := normalizeStyleName(styleName)
		if normalized == "" {
			continue
		}
		if normalized == "normal" || normalized == "default paragraph font" {
			continue
		}
		if declared[normalized] {
			if !seenDeclared[normalized] {
				findings = append(findings, map[string]any{
					"type":       "declared_custom_style_used",
					"style_name": styleName,
					"style_kind": styleKind,
					"location":   item["location"],
					"text":       item["text"],
					"severity":   "low",
					"suggestion": "Declared custom style is present in the manuscript and ready for intentional EPUB/Typst handling.",
				})
				seenDeclared[normalized] = true
			}
			continue
		}
		if seenWarnings[normalized] {
			continue
		}
		findings = append(findings, map[string]any{
			"type":       "undeclared_custom_style",
			"style_name": styleName,
			"style_kind": styleKind,
			"location":   item["location"],
			"text":       item["text"],
			"severity":   "high",
			"suggestion": "Custom style is used in the manuscript but not declared in the project spec/transmittal. Add it before production so EPUB and Typst can treat it intentionally.",
		})
		seenWarnings[normalized] = true
	}
	sort.SliceStable(findings, func(i, j int) bool {
		return fmt.Sprint(findings[i]["type"]) < fmt.Sprint(findings[j]["type"])
	})
	return json.Marshal(findings)
}

func preflightReportURL(projectID int64, bookID int64, preflightID int64) string {
	base := fmt.Sprintf("/api/projects/%d/preflight/report?book_id=%d", projectID, bookID)
	if preflightID > 0 {
		return fmt.Sprintf("%s&preflight_id=%d", base, preflightID)
	}
	return base
}

func preflightHistoryEntries(projectID int64, rows []dbgen.ManuscriptPreflight) []preflightHistoryEntry {
	if len(rows) == 0 {
		return nil
	}
	entries := make([]preflightHistoryEntry, 0, len(rows))
	for i, row := range rows {
		entries = append(entries, preflightHistoryEntry{
			ID:        row.ID,
			Status:    row.Status,
			UpdatedAt: row.UpdatedAt.Format(time.RFC3339),
			ReportURL: preflightReportURL(projectID, row.BookID, row.ID),
			Latest:    i == 0,
		})
	}
	return entries
}

func (s *Server) preflightResponseFromRow(projectID int64, row dbgen.ManuscriptPreflight, history []dbgen.ManuscriptPreflight) preflightResponse {
	resp := preflightResponse{
		Exists:         true,
		ProjectID:      projectID,
		BookID:         row.BookID,
		Status:         row.Status,
		SourceFilename: row.SourceFilename,
		UpdatedAt:      row.UpdatedAt.Format(time.RFC3339),
		ReportURL:      preflightReportURL(projectID, row.BookID, row.ID),
		History:        preflightHistoryEntries(projectID, history),
	}
	if summary, err := parseStoredSummary(row.SummaryJson); err == nil {
		resp.Summary = summary
	}
	resp.Images = parseStoredImages(row.ReportJson)
	if row.Status == "error" && row.ErrorMsg != "" {
		resp.Error = row.ErrorMsg
	}
	return resp
}

func (s *Server) handleGetManuscriptPreflight(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	bookID, err := strconv.ParseInt(r.URL.Query().Get("book_id"), 10, 64)
	if err != nil || bookID <= 0 {
		jsonErr(w, "bad book_id", 400)
		return
	}
	q := dbgen.New(s.DB)
	history, err := q.ListManuscriptPreflights(r.Context(), dbgen.ListManuscriptPreflightsParams{ProjectID: pid, BookID: bookID})
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	if len(history) == 0 {
		jsonOK(w, preflightResponse{Exists: false})
		return
	}
	jsonOK(w, s.preflightResponseFromRow(pid, history[0], history))
}

func (s *Server) handleGetManuscriptPreflightReport(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	bookID, err := strconv.ParseInt(r.URL.Query().Get("book_id"), 10, 64)
	if err != nil || bookID <= 0 {
		jsonErr(w, "bad book_id", 400)
		return
	}
	preflightIDStr := strings.TrimSpace(r.URL.Query().Get("preflight_id"))
	q := dbgen.New(s.DB)
	var row dbgen.ManuscriptPreflight
	if preflightIDStr == "" {
		row, err = q.GetLatestManuscriptPreflight(r.Context(), dbgen.GetLatestManuscriptPreflightParams{ProjectID: pid, BookID: bookID})
		if errors.Is(err, sql.ErrNoRows) {
			jsonErr(w, "not found", 404)
			return
		}
		if err != nil {
			jsonErr(w, err.Error(), 500)
			return
		}
	} else {
		preflightID, idErr := strconv.ParseInt(preflightIDStr, 10, 64)
		if idErr != nil || preflightID <= 0 {
			jsonErr(w, "bad preflight_id", 400)
			return
		}
		history, listErr := q.ListManuscriptPreflights(r.Context(), dbgen.ListManuscriptPreflightsParams{ProjectID: pid, BookID: bookID})
		if listErr != nil {
			jsonErr(w, listErr.Error(), 500)
			return
		}
		found := false
		for _, candidate := range history {
			if candidate.ID == preflightID {
				row = candidate
				found = true
				break
			}
		}
		if !found {
			jsonErr(w, "not found", 404)
			return
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(row.ReportHtml))
}

func (s *Server) handleRunManuscriptPreflight(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	pid, err := s.projectIDFromPath(r)
	if err != nil {
		jsonErr(w, "bad id", 400)
		return
	}
	var body preflightRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, "bad request", 400)
		return
	}
	if body.BookID <= 0 {
		jsonErr(w, "book_id required", 400)
		return
	}
	q := dbgen.New(s.DB)
	book, err := q.GetBook(r.Context(), body.BookID)
	if err != nil {
		jsonErr(w, "book not found", 404)
		return
	}
	specData := ""
	spec, err := q.GetBookSpec(r.Context(), pid)
	if errors.Is(err, sql.ErrNoRows) {
		full, uErr := q.UpsertBookSpec(r.Context(), dbgen.UpsertBookSpecParams{ProjectID: pid, Data: defaultSpecData()})
		if uErr != nil {
			jsonErr(w, uErr.Error(), 500)
			return
		}
		specData = full.Data
	} else if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	} else {
		specData = spec.Data
	}
	if book.SourceData == nil || len(book.SourceData) == 0 {
		jsonErr(w, "no source file", 400)
		return
	}
	baseName := strings.TrimSpace(book.SourceFilename)
	if baseName == "" {
		baseName = fmt.Sprintf("book-%d.docx", book.ID)
	}
	baseName = filepath.Base(baseName)
	tmpDir, err := os.MkdirTemp("", "prodcal-preflight-docx-*")
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	defer os.RemoveAll(tmpDir)
	tmpPath := filepath.Join(tmpDir, baseName)
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	if _, err := tmpFile.Write(book.SourceData); err != nil {
		_ = tmpFile.Close()
		jsonErr(w, err.Error(), 500)
		return
	}
	_ = tmpFile.Close()

	runnerTmpDir, err := os.MkdirTemp("", "prodcal-preflight-declared-*")
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	defer os.RemoveAll(runnerTmpDir)
	declaredStylesPath, err := writeDeclaredStylesFile(runnerTmpDir, specData)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}

	htmlBytes, jsonBytes, runErr := s.getPreflightRunner()(tmpPath, declaredStylesPath)
	status := "ready"
	errorMsg := ""
	if runErr != nil {
		status = "error"
		errorMsg = runErr.Error()
		htmlBytes = []byte("<html><body><h1>Preflight Error</h1><pre>" + errorMsg + "</pre></body></html>")
		jsonBytes = []byte("[]")
	}
	jsonBytes, err = appendUndeclaredStyleWarnings(jsonBytes, specData)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	summary, err := buildPreflightSummary(jsonBytes)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	summaryBytes, err := json.Marshal(summary)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	row, err := q.CreateManuscriptPreflight(r.Context(), dbgen.CreateManuscriptPreflightParams{
		ProjectID:      pid,
		BookID:         body.BookID,
		Status:         status,
		SummaryJson:    string(summaryBytes),
		ReportJson:     string(jsonBytes),
		ReportHtml:     string(htmlBytes),
		ErrorMsg:       errorMsg,
		SourceFilename: book.SourceFilename,
	})
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	history, histErr := q.ListManuscriptPreflights(r.Context(), dbgen.ListManuscriptPreflightsParams{ProjectID: pid, BookID: body.BookID})
	if histErr != nil {
		jsonErr(w, histErr.Error(), 500)
		return
	}
	jsonOK(w, s.preflightResponseFromRow(pid, row, history))
}
