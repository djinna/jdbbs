package srv

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode/utf8"
)

// Report a nit (punch list 8.7). Anyone, no login: a reader on any page —
// help or app — can flag a typo, a wrong sentence or a bug, with the text
// they had selected. Spam defence is a honeypot field plus per-IP and global
// rate limits. Each nit is stored (table nits, migration 055) and pushed to
// the punch-list inbox (Section 0) tagged [nit·help] or [nit·app].
//
// The optional email is stored only so a later "fixed, thanks" reply can go
// through the email-consent guard (plan item 13); nothing is sent here.

const (
	nitMaxComment   = 2000
	nitMaxSelection = 500
	nitPerIP        = 5 // per nitIPWindow
	nitIPWindow     = 10 * time.Minute
	nitGlobalPerDay = 200
)

type nitInput struct {
	Page      string `json:"page"`
	Selection string `json:"selection"`
	Comment   string `json:"comment"`
	Email     string `json:"email"`
	Kind      string `json:"kind"`
	Website   string `json:"website"` // honeypot: humans never see it
}

// nitInboxURL is the punch-list page's add endpoint. PRODCAL_NIT_INBOX_URL
// overrides; "off" disables. The default is the runpage on this VM; when it
// isn't running the push fails quietly and the nit stays in the DB.
func nitInboxURL() string {
	v := os.Getenv("PRODCAL_NIT_INBOX_URL")
	if v == "off" {
		return ""
	}
	if v != "" {
		return v
	}
	return "http://127.0.0.1:8766/add"
}

func clipRunes(s string, n int) string {
	s = strings.TrimSpace(s)
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	r := []rune(s)
	return string(r[:n]) + "…"
}

// oneLine flattens text for a checklist row: no newlines, no markdown
// checkbox or heading syntax surviving at the start.
func oneLine(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	return strings.NewReplacer("[", "⟦", "]", "⟧").Replace(s)
}

// POST /api/nits
func (s *Server) handleNitCreate(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var in nitInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, `{"error":"bad request"}`, http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	// Honeypot: answer as if accepted so bots learn nothing.
	if strings.TrimSpace(in.Website) != "" {
		slog.Info("nit: honeypot", "ip", clientIP(r))
		fmt.Fprint(w, `{"ok":true}`)
		return
	}
	in.Comment = clipRunes(in.Comment, nitMaxComment)
	in.Selection = clipRunes(in.Selection, nitMaxSelection)
	in.Page = clipRunes(in.Page, 300)
	in.Email = clipRunes(in.Email, 200)
	if in.Comment == "" && in.Selection == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error":"Say what's wrong, or select the text first."}`)
		return
	}
	if in.Email != "" && !looksLikeEmail(in.Email) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error":"That email address doesn't look right. Leave it blank if you like."}`)
		return
	}
	if !strings.HasPrefix(in.Page, "/") {
		in.Page = "/" + strings.TrimLeft(in.Page, "/")
	}
	if in.Kind != "help" {
		in.Kind = "app"
	}
	if strings.HasPrefix(in.Page, "/help") {
		in.Kind = "help"
	}
	ip := clientIP(r)
	if !s.limiter().allow("nit:"+ip, nitPerIP, nitIPWindow) || !s.limiter().allow("nit:*", nitGlobalPerDay, 24*time.Hour) {
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprint(w, `{"error":"Thanks — that's a lot of nits at once. Try again in a few minutes."}`)
		return
	}
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte("nit-ip:" + ip))
	ipHash := hex.EncodeToString(mac.Sum(nil))[:16]
	res, err := s.DB.Exec(`INSERT INTO nits (kind, page, selection, comment, email, ip_hash, user_agent) VALUES (?,?,?,?,?,?,?)`,
		in.Kind, in.Page, in.Selection, in.Comment, in.Email, ipHash, clipRunes(r.UserAgent(), 200))
	if err != nil {
		slog.Error("nit: insert", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"error":"Couldn't save that. Please try again."}`)
		return
	}
	id, _ := res.LastInsertId()
	go s.pushNitToInbox(id, in)
	fmt.Fprintf(w, `{"ok":true,"id":%d}`, id)
}

// pushNitToInbox adds a Section 0 row on the punch list and marks the nit.
func (s *Server) pushNitToInbox(id int64, in nitInput) {
	url := nitInboxURL()
	if url == "" {
		return
	}
	text := fmt.Sprintf("[nit·%s #%d] `%s`", in.Kind, id, oneLine(in.Page))
	if in.Selection != "" {
		text += " — “" + oneLine(clipRunes(in.Selection, 160)) + "”"
	}
	if in.Comment != "" {
		text += " — " + oneLine(clipRunes(in.Comment, 400))
	}
	if in.Email != "" {
		text += " (reply ok: " + oneLine(in.Email) + ")"
	}
	body, _ := json.Marshal(map[string]string{"text": text})
	c := &http.Client{Timeout: 3 * time.Second}
	resp, err := c.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		slog.Info("nit: inbox push skipped", "id", id, "err", err)
		return
	}
	resp.Body.Close()
	if resp.StatusCode/100 == 2 && s.DB != nil {
		s.DB.Exec(`UPDATE nits SET inbox_ok = 1 WHERE id = ?`, id)
	}
}

// GET /api/admin/nits — newest first (the /admin/help/ queue reads this, 8.9).
func (s *Server) handleAdminNits(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	rows, err := s.DB.Query(`SELECT id, created_at, kind, page, selection, comment, email, status, inbox_ok FROM nits ORDER BY id DESC LIMIT 200`)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()
	type nit struct {
		ID        int64  `json:"id"`
		CreatedAt string `json:"created_at"`
		Kind      string `json:"kind"`
		Page      string `json:"page"`
		Selection string `json:"selection"`
		Comment   string `json:"comment"`
		Email     string `json:"email"`
		Status    string `json:"status"`
		InboxOK   bool   `json:"inbox_ok"`
	}
	out := []nit{}
	for rows.Next() {
		var n nit
		if rows.Scan(&n.ID, &n.CreatedAt, &n.Kind, &n.Page, &n.Selection, &n.Comment, &n.Email, &n.Status, &n.InboxOK) == nil {
			out = append(out, n)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}
