package srv

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type backupStatusResponse struct {
	OK               bool     `json:"ok"`
	Status           string   `json:"status"`
	GeneratedAt      string   `json:"generated_at"`
	LocalLastSuccess string   `json:"local_last_success"`
	R2LastSuccess    string   `json:"r2_last_success"`
	HealthLastOK     string   `json:"health_last_ok"`
	NewestLocal      string   `json:"newest_local"`
	NewestLocalAgeH  int      `json:"newest_local_age_hours"`
	LocalCount       int      `json:"local_count"`
	LocalBytes       int64    `json:"local_bytes"`
	Problems         []string `json:"problems"`
}

func (s *Server) collectBackupStatus() backupStatusResponse {
	backupDir := os.Getenv("BACKUP_DIR")
	if strings.TrimSpace(backupDir) == "" {
		backupDir = filepath.Join(os.Getenv("HOME"), "backups")
		if strings.TrimSpace(os.Getenv("HOME")) == "" {
			backupDir = "/home/exedev/backups"
		}
	}
	now := time.Now().UTC()
	res := backupStatusResponse{
		OK:          true,
		Status:      "OK",
		GeneratedAt: now.Format(time.RFC3339),
		Problems:    []string{},
	}

	readSentinel := func(name string) map[string]string {
		out := map[string]string{}
		b, err := os.ReadFile(filepath.Join(backupDir, name))
		if err != nil {
			return out
		}
		for _, line := range strings.Split(string(b), "\n") {
			k, v, ok := strings.Cut(line, ":")
			if ok {
				out[strings.TrimSpace(k)] = strings.TrimSpace(v)
			}
		}
		return out
	}

	if s := readSentinel(".LAST-SUCCESS"); s["time"] != "" {
		res.LocalLastSuccess = s["time"]
	} else {
		res.Problems = append(res.Problems, "local backup success sentinel missing")
	}
	if s := readSentinel(".LAST-R2-SUCCESS"); s["time"] != "" {
		res.R2LastSuccess = s["time"]
	} else {
		res.Problems = append(res.Problems, "R2 backup success sentinel missing")
	}
	if s := readSentinel(".HEALTH-OK"); s["time"] != "" {
		res.HealthLastOK = s["time"]
	}

	for _, fail := range []string{".LAST-FAILURE", ".LAST-R2-FAILURE", ".HEALTH-FAIL"} {
		if _, err := os.Stat(filepath.Join(backupDir, fail)); err == nil {
			res.Problems = append(res.Problems, fail+" is present")
		}
	}

	entries, err := filepath.Glob(filepath.Join(backupDir, "prodcal-*.sqlite3.gz"))
	if err == nil {
		var newestPath string
		var newestMod time.Time
		for _, p := range entries {
			st, err := os.Stat(p)
			if err != nil || st.IsDir() {
				continue
			}
			res.LocalCount++
			res.LocalBytes += st.Size()
			if newestPath == "" || st.ModTime().After(newestMod) {
				newestPath = p
				newestMod = st.ModTime()
			}
		}
		if newestPath != "" {
			res.NewestLocal = filepath.Base(newestPath)
			res.NewestLocalAgeH = int(now.Sub(newestMod).Hours())
			if now.Sub(newestMod) > 30*time.Hour {
				res.Problems = append(res.Problems, "newest local backup is older than 30h")
			}
		} else {
			res.Problems = append(res.Problems, "no local backup files found")
		}
	}

	for _, ts := range []struct{ label, value string }{
		{"local backup", res.LocalLastSuccess},
		{"R2 backup", res.R2LastSuccess},
	} {
		if ts.value == "" {
			continue
		}
		if t, err := time.Parse(time.RFC3339, ts.value); err == nil && now.Sub(t) > 30*time.Hour {
			res.Problems = append(res.Problems, ts.label+" is older than 30h")
		}
	}

	if len(res.Problems) > 0 {
		res.OK = false
		res.Status = "ACTION NEEDED"
	}
	return res
}

func (s *Server) handleAdminBackupStatus(w http.ResponseWriter, r *http.Request) {
	if !s.requireExeDevAdminAPI(w, r) {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(s.collectBackupStatus())
}

func formatBytesIEC(n int64) string {
	const unit = 1024
	if n < unit {
		return strconv.FormatInt(n, 10) + " B"
	}
	div, exp := int64(unit), 0
	for v := n / unit; v >= unit; v /= unit {
		div *= unit
		exp++
	}
	return strconv.FormatFloat(float64(n)/float64(div), 'f', 1, 64) + " " + string("KMGTPE"[exp]) + "iB"
}
