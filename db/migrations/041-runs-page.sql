-- Read-only archive of punch lists / run logs rendered from docs/runs/*.md.
INSERT OR IGNORE INTO site_pages (route, title, owner, source, visibility, listed, page_type, status, note) VALUES
 ('/admin/runs/', 'Runs (punch-list archive)', 'prodcal', 'srv/runs.go', 'admin', 'nav', 'reference', 'live', 'Index of docs/runs/*.md (exported by scripts/punchlist-export.py: checklist + every note thread). /admin/runs/{name} renders one file via pandoc gfm. Read-only; the live list stays on the runpage server :8766.');
