package srv

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Factory activity feed (monitoring L2). Every point that already writes a
// slog line for a pass holder's action also drops a row here, so during a
// workshop session /admin/factory/ answers "who did what in the last hour"
// without a terminal. Writes are best-effort and never fail the request.

// factoryEvent records one row. projectID 0 means client-level (sign-in).
// actor is requestActor(r) for requests, "factory" for background work.
func (s *Server) factoryEvent(projectID int64, clientSlug, kind, actor, detail string) {
	if actor == "" {
		actor = "anon"
	}
	var pid any
	if projectID > 0 {
		pid = projectID
	}
	if clientSlug == "" && projectID > 0 {
		_ = s.DB.QueryRowContext(context.Background(),
			`SELECT client_slug FROM projects WHERE id = ?`, projectID).Scan(&clientSlug)
	}
	_, err := s.DB.ExecContext(context.Background(),
		`INSERT INTO factory_events (project_id, client_slug, kind, actor, detail) VALUES (?, ?, ?, ?, ?)`,
		pid, clientSlug, kind, actor, clip(detail, 400))
	if err != nil {
		slog.Warn("factory event not recorded", "kind", kind, "project_id", projectID, "err", err)
	}
}

// factoryEventR is the request-scoped form: actor from the request.
func (s *Server) factoryEventR(r *http.Request, projectID int64, kind, detail string) {
	s.factoryEvent(projectID, "", kind, requestActor(r), detail)
}

type factoryEventRow struct {
	ID          int64  `json:"id"`
	At          string `json:"at"`
	ProjectID   int64  `json:"project_id,omitempty"`
	ClientSlug  string `json:"client_slug"`
	ProjectSlug string `json:"project_slug,omitempty"`
	ProjectName string `json:"project_name,omitempty"`
	Kind        string `json:"kind"`
	Actor       string `json:"actor"`
	Detail      string `json:"detail"`
}

// handleAdminFactoryEvents — GET /api/admin/factory/events?after=ID&limit=N&project=ID&since=RFC3339
// Newest first. `after` returns only rows with id > after (for polling).
func (s *Server) handleAdminFactoryEvents(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 1000 {
		limit = 300
	}
	after, _ := strconv.ParseInt(r.URL.Query().Get("after"), 10, 64)
	project, _ := strconv.ParseInt(r.URL.Query().Get("project"), 10, 64)
	where := []string{"1=1"}
	args := []any{}
	if after > 0 {
		where = append(where, "e.id > ?")
		args = append(args, after)
	}
	if project > 0 {
		where = append(where, "e.project_id = ?")
		args = append(args, project)
	}
	if since := r.URL.Query().Get("since"); since != "" {
		if t, err := time.Parse(time.RFC3339, since); err == nil {
			where = append(where, "e.created_at >= ?")
			args = append(args, t.UTC().Format("2006-01-02 15:04:05"))
		}
	}
	args = append(args, limit)
	rows, err := s.DB.QueryContext(r.Context(), `
		SELECT e.id, e.created_at, COALESCE(e.project_id, 0), e.client_slug,
		       COALESCE(p.project_slug, ''), COALESCE(p.name, ''), e.kind, e.actor, e.detail
		FROM factory_events e
		LEFT JOIN projects p ON p.id = e.project_id
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY e.id DESC LIMIT ?`, args...)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	defer rows.Close()
	out := []factoryEventRow{}
	for rows.Next() {
		var e factoryEventRow
		var at time.Time
		if err := rows.Scan(&e.ID, &at, &e.ProjectID, &e.ClientSlug, &e.ProjectSlug, &e.ProjectName, &e.Kind, &e.Actor, &e.Detail); err != nil {
			jsonErr(w, err.Error(), 500)
			return
		}
		e.At = at.UTC().Format(time.RFC3339)
		out = append(out, e)
	}
	jsonOK(w, out)
}

// factoryBoardRow is the per-project state /admin/factory/ shows next to
// each pass (the pass fields themselves come from /api/admin/passes).
type factoryBoardRow struct {
	ProjectID         int64  `json:"project_id"`
	TransmittalStatus string `json:"transmittal_status"` // "", draft, final
	TransmittalAt     string `json:"transmittal_at,omitempty"`
	BookID            int64  `json:"book_id,omitempty"`
	BookFile          string `json:"book_file,omitempty"`
	BookStatus        string `json:"book_status,omitempty"` // uploaded, converting, ready, error
	BookError         string `json:"book_error,omitempty"`
	BookAt            string `json:"book_at,omitempty"`
	InspectStatus     string `json:"inspect_status,omitempty"`
	InspectHigh       int    `json:"inspect_high"`
	InspectMedium     int    `json:"inspect_medium"`
	InspectLow        int    `json:"inspect_low"`
	InspectAt         string `json:"inspect_at,omitempty"`
	HasCover          bool   `json:"has_cover"`
	LastEventKind     string `json:"last_event_kind,omitempty"`
	LastEventAt       string `json:"last_event_at,omitempty"`
	LastActor         string `json:"last_actor,omitempty"`
}

func nullTimeRFC(t sql.NullTime) string {
	if !t.Valid {
		return ""
	}
	return t.Time.UTC().Format(time.RFC3339)
}

// handleAdminFactoryBoard — GET /api/admin/factory/board → one row per
// project that holds a pass (any status), keyed by project_id.
// One query per aspect, each fully drained before the next: the pool is
// MaxOpenConns(1), so nested queries deadlock.
func (s *Server) handleAdminFactoryBoard(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	ctx := r.Context()
	const inPasses = `IN (SELECT DISTINCT project_id FROM passes)`
	ids := []int64{}
	rows, err := s.DB.QueryContext(ctx, `SELECT DISTINCT project_id FROM passes ORDER BY project_id`)
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	for rows.Next() {
		var id int64
		if rows.Scan(&id) == nil {
			ids = append(ids, id)
		}
	}
	rows.Close()
	board := map[int64]*factoryBoardRow{}
	for _, id := range ids {
		board[id] = &factoryBoardRow{ProjectID: id}
	}

	if rows, err := s.DB.QueryContext(ctx, `SELECT project_id, status, updated_at FROM transmittals WHERE project_id `+inPasses); err == nil {
		for rows.Next() {
			var pid int64
			var st string
			var at sql.NullTime
			if rows.Scan(&pid, &st, &at) == nil {
				if b := board[pid]; b != nil {
					b.TransmittalStatus, b.TransmittalAt = st, nullTimeRFC(at)
				}
			}
		}
		rows.Close()
	}
	if rows, err := s.DB.QueryContext(ctx, `
		SELECT b.project_id, b.id, b.source_filename, b.status, b.error_msg, b.updated_at FROM books b
		WHERE b.project_id `+inPasses+` AND b.id = (SELECT MAX(id) FROM books b2 WHERE b2.project_id = b.project_id)`); err == nil {
		for rows.Next() {
			var pid, bid int64
			var file, st, errMsg string
			var at sql.NullTime
			if rows.Scan(&pid, &bid, &file, &st, &errMsg, &at) == nil {
				if b := board[pid]; b != nil {
					b.BookID, b.BookFile, b.BookStatus, b.BookError, b.BookAt = bid, file, st, clip(errMsg, 200), nullTimeRFC(at)
				}
			}
		}
		rows.Close()
	}
	if rows, err := s.DB.QueryContext(ctx, `
		SELECT m.project_id, m.status, m.summary_json, m.created_at FROM manuscript_preflights m
		WHERE m.project_id `+inPasses+` AND m.id = (SELECT MAX(id) FROM manuscript_preflights m2 WHERE m2.project_id = m.project_id)`); err == nil {
		for rows.Next() {
			var pid int64
			var st, summary string
			var at sql.NullTime
			if rows.Scan(&pid, &st, &summary, &at) == nil {
				if b := board[pid]; b != nil {
					b.InspectStatus, b.InspectAt = st, nullTimeRFC(at)
					var sm struct {
						High   int `json:"high"`
						Medium int `json:"medium"`
						Low    int `json:"low"`
					}
					if json.Unmarshal([]byte(summary), &sm) == nil {
						b.InspectHigh, b.InspectMedium, b.InspectLow = sm.High, sm.Medium, sm.Low
					}
				}
			}
		}
		rows.Close()
	}
	if rows, err := s.DB.QueryContext(ctx, `SELECT project_id FROM book_specs WHERE project_id `+inPasses+` AND cover_data IS NOT NULL AND length(cover_data) > 0`); err == nil {
		for rows.Next() {
			var pid int64
			if rows.Scan(&pid) == nil {
				if b := board[pid]; b != nil {
					b.HasCover = true
				}
			}
		}
		rows.Close()
	}
	if rows, err := s.DB.QueryContext(ctx, `
		SELECT e.project_id, e.kind, e.actor, e.created_at FROM factory_events e
		WHERE e.project_id `+inPasses+` AND e.id = (SELECT MAX(id) FROM factory_events e2 WHERE e2.project_id = e.project_id)`); err == nil {
		for rows.Next() {
			var pid int64
			var kind, actor string
			var at sql.NullTime
			if rows.Scan(&pid, &kind, &actor, &at) == nil {
				if b := board[pid]; b != nil {
					b.LastEventKind, b.LastActor, b.LastEventAt = kind, actor, nullTimeRFC(at)
				}
			}
		}
		rows.Close()
	}
	out := make([]factoryBoardRow, 0, len(ids))
	for _, id := range ids {
		out = append(out, *board[id])
	}
	jsonOK(w, out)
}
