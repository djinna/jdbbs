package srv

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	_ "embed"
)

// House stylesheet — the public, anonymized, read-only sibling of the PI
// review tool in stylesheet.go. Served at /stylesheet/ (HTML) with the same
// data at /stylesheet/index.md and /api/stylesheet/house. Filter by
// ?kind=fiction|nonfiction (default: all). Every item is "accepted" by
// definition — there is no decision workflow here; edits happen via the seed
// file or direct SQL until the per-client instance idea (docs/IDEAS.md) lands.

//go:embed house_style_seed.json
var houseStyleSeedJSON []byte

type hsItem struct {
	ID         int64  `json:"id"`
	SectionOrd int    `json:"section_ord"`
	Section    string `json:"section"`
	ItemOrd    int    `json:"item_ord"`
	Kind       string `json:"kind"`
	Col1       string `json:"col1"`
	Col2       string `json:"col2"`
	Col3       string `json:"col3"`
	Body       string `json:"body"`
	BookKind   string `json:"book_kind"`
}

type hsSection struct {
	Ord   int      `json:"ord"`
	Title string   `json:"title"`
	Items []hsItem `json:"items"`
}

func (s *Server) registerHouseStyleRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /stylesheet/{$}", s.handleHouseStylePage)
	mux.HandleFunc("GET /stylesheet/index.md", s.handleHouseStyleMD)
	mux.HandleFunc("GET /api/stylesheet/house", s.handleHouseStyleJSON)
}

func (s *Server) hsEnsureSeeded(ctx context.Context) error {
	var n int
	if err := s.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM house_style").Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	var rows []hsItem
	if err := json.Unmarshal(houseStyleSeedJSON, &rows); err != nil {
		return fmt.Errorf("parse house style seed: %w", err)
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, r := range rows {
		if r.BookKind == "" {
			r.BookKind = "both"
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO house_style (section_ord, section, item_ord, kind, col1, col2, col3, body, book_kind)
			 VALUES (?,?,?,?,?,?,?,?,?)`,
			r.SectionOrd, r.Section, r.ItemOrd, r.Kind, r.Col1, r.Col2, r.Col3, r.Body, r.BookKind); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// hsKindParam validates ?kind=. "" means all.
func hsKindParam(r *http.Request) string {
	k := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("kind")))
	if k == "fiction" || k == "nonfiction" {
		return k
	}
	return ""
}

func (s *Server) hsLoad(ctx context.Context, kind string) ([]hsSection, error) {
	if err := s.hsEnsureSeeded(ctx); err != nil {
		return nil, err
	}
	rows, err := s.DB.QueryContext(ctx,
		`SELECT id, section_ord, section, item_ord, kind, col1, col2, col3, body, book_kind
		 FROM house_style ORDER BY section_ord, item_ord, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var secs []hsSection
	for rows.Next() {
		var it hsItem
		if err := rows.Scan(&it.ID, &it.SectionOrd, &it.Section, &it.ItemOrd, &it.Kind,
			&it.Col1, &it.Col2, &it.Col3, &it.Body, &it.BookKind); err != nil {
			return nil, err
		}
		if kind != "" && it.BookKind != "both" && it.BookKind != kind {
			continue
		}
		if len(secs) == 0 || secs[len(secs)-1].Ord != it.SectionOrd {
			secs = append(secs, hsSection{Ord: it.SectionOrd, Title: it.Section})
		}
		secs[len(secs)-1].Items = append(secs[len(secs)-1].Items, it)
	}
	return secs, rows.Err()
}

func (s *Server) handleHouseStyleJSON(w http.ResponseWriter, r *http.Request) {
	secs, err := s.hsLoad(r.Context(), hsKindParam(r))
	if err != nil {
		jsonErr(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"kind": hsKindParam(r), "sections": secs})
}

func (s *Server) handleHouseStyleMD(w http.ResponseWriter, r *http.Request) {
	kind := hsKindParam(r)
	secs, err := s.hsLoad(r.Context(), kind)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", hsTitle(kind))
	fmt.Fprintf(&b, "_[jdbb] studio house editorial stylesheet · generated %s · https://jdbbs.exe.xyz/stylesheet/_\n\n", time.Now().UTC().Format("2006-01-02"))
	for _, sec := range secs {
		fmt.Fprintf(&b, "## %d. %s\n\n", sec.Ord, sec.Title)
		var rules, prose []hsItem
		for _, it := range sec.Items {
			if it.Kind == "rule" {
				rules = append(rules, it)
			} else {
				prose = append(prose, it)
			}
		}
		if len(rules) > 0 {
			b.WriteString("| Item | Rule | Example / note |\n|---|---|---|\n")
			for _, it := range rules {
				fmt.Fprintf(&b, "| %s | %s | %s |\n", mdCell(it.Col1), mdCell(it.Col2), mdCell(it.Col3))
			}
			b.WriteString("\n")
		}
		for _, it := range prose {
			b.WriteString(strings.TrimSpace(it.Body) + "\n\n")
		}
	}
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	_, _ = w.Write([]byte(b.String()))
}

func hsTitle(kind string) string {
	switch kind {
	case "fiction":
		return "House Editorial Stylesheet — Fiction"
	case "nonfiction":
		return "House Editorial Stylesheet — Nonfiction"
	}
	return "House Editorial Stylesheet"
}

// handleHouseStylePage serves the static shell; the browser fetches JSON and
// renders. Kept as a shell (not server-rendered) so the fiction/nonfiction
// tabs switch without a reload and the page shares the studio theme boot.
func (s *Server) handleHouseStylePage(w http.ResponseWriter, r *http.Request) {
	data, err := staticFS.ReadFile("static/housestyle.html")
	if err != nil {
		http.Error(w, "not found", 404)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(data)
}
