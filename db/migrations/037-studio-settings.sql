-- Studio settings: small editable strings the admin can change without a deploy.
-- First use: the email sign-off and the default email footer line, edited from
-- /admin/docs/. Keys are fixed in Go (see srv/settings.go); values are plain text.
CREATE TABLE IF NOT EXISTS studio_settings (
  key        TEXT PRIMARY KEY,
  value      TEXT NOT NULL,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT OR IGNORE INTO site_pages (route, title, owner, source, visibility, listed, page_type, status, note) VALUES
 ('/admin/docs/', 'Doc editor (jdbbs-public pages + email sign-off)', 'prodcal', 'srv/static/docs-editor.html', 'admin', 'nav', 'tool', 'live', 'Edits the on-disk jdbbs-public files listed in this registry (owner = jdbbs-public) and the studio_settings strings (email signature, footer). Saves publish instantly; commit to git from the same page.');
