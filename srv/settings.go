package srv

import (
	"context"
	"database/sql"
	"strings"
	"sync"
)

// ─── Studio settings: admin-editable strings, no deploy needed ───
//
// A handful of strings that used to be Go constants (the email sign-off, the
// default footer line) now live in studio_settings and are edited from
// /admin/docs/. The set of keys is fixed here so a typo can't create a new
// one; defaults keep every email identical to before the table existed.

type settingDef struct {
	Key     string `json:"key"`
	Label   string `json:"label"`
	Help    string `json:"help"`
	Default string `json:"default"`
}

var settingDefs = []settingDef{
	{
		Key:     "email_signature",
		Label:   "Email sign-off",
		Help:    "Closes every studio email. First line is the name; the lines after it are set in the secondary colour in HTML mail. Plain text — no HTML.",
		Default: "— Jenna\n[jdbb] studio",
	},
	{
		Key:     "email_footer",
		Label:   "Email footer line",
		Help:    "Small print under the closing hairline when a template does not supply its own (announcements and digests do). Plain text; a bare URL becomes a link.",
		Default: "https://jdbbs.exe.xyz/ · Reply to this email to reach Jenna.",
	},
}

var settingsCache = struct {
	sync.RWMutex
	v map[string]string
}{v: map[string]string{}}

func settingDefault(key string) string {
	for _, d := range settingDefs {
		if d.Key == key {
			return d.Default
		}
	}
	return ""
}

func isSettingKey(key string) bool {
	for _, d := range settingDefs {
		if d.Key == key {
			return true
		}
	}
	return false
}

// setting returns the current value for key: the DB value loaded at startup
// (or on save), else the compiled default.
func setting(key string) string {
	settingsCache.RLock()
	v, ok := settingsCache.v[key]
	settingsCache.RUnlock()
	if ok && strings.TrimSpace(v) != "" {
		return v
	}
	return settingDefault(key)
}

// loadSettings fills the cache from studio_settings. Missing table or rows
// are not errors: the defaults stand.
func (s *Server) loadSettings(ctx context.Context) error {
	rows, err := s.DB.QueryContext(ctx, `SELECT key, value FROM studio_settings`)
	if err != nil {
		return err
	}
	defer rows.Close()
	m := map[string]string{}
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return err
		}
		m[k] = v
	}
	settingsCache.Lock()
	settingsCache.v = m
	settingsCache.Unlock()
	return rows.Err()
}

// saveSetting writes one value (empty = revert to default) and refreshes the cache.
func (s *Server) saveSetting(ctx context.Context, key, value string) error {
	value = strings.ReplaceAll(strings.TrimRight(value, " \t\r\n"), "\r\n", "\n")
	var err error
	if strings.TrimSpace(value) == "" {
		_, err = s.DB.ExecContext(ctx, `DELETE FROM studio_settings WHERE key = ?`, key)
	} else {
		_, err = s.DB.ExecContext(ctx, `INSERT INTO studio_settings (key, value) VALUES (?, ?)
			ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP`, key, value)
	}
	if err != nil {
		return err
	}
	return s.loadSettings(ctx)
}

// settingMeta is what the editor shows: definition + current value + whether
// the stored value differs from the default.
type settingMeta struct {
	settingDef
	Value     string `json:"value"`
	UpdatedAt string `json:"updated_at"`
	Custom    bool   `json:"custom"`
}

func (s *Server) listSettings(ctx context.Context) ([]settingMeta, error) {
	out := make([]settingMeta, 0, len(settingDefs))
	for _, d := range settingDefs {
		m := settingMeta{settingDef: d, Value: setting(d.Key)}
		var at sql.NullString
		err := s.DB.QueryRowContext(ctx, `SELECT updated_at FROM studio_settings WHERE key = ?`, d.Key).Scan(&at)
		if err == nil && at.Valid {
			m.UpdatedAt = at.String
			m.Custom = m.Value != d.Default
		} else if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
		out = append(out, m)
	}
	return out, nil
}
