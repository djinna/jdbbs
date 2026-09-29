-- Help system (punch list 8.5): /help/ renders docs/help/*.md in the shell;
-- llms.txt + per-page .md twins for agents; FTS5 search. Pages start as
-- drafts (admin-only) until Jenna approves them.
INSERT OR IGNORE INTO site_pages (route, title, owner, source, visibility, listed, page_type, status, note) VALUES
 ('/help/', 'Help', 'prodcal', 'docs/help/*.md (srv/help.go)', 'public', 'unlisted', 'prose', 'review', 'Studio help. Public tier; pages with visibility: admin or status: draft show only to the admin. Groups derive from the site_pages tier of each page''s routes. Not in any nav until the first pages are approved (8.6 adds "?" links).'),
 ('/help/llms.txt', 'Help — llms.txt (agent index)', 'generated', 'srv/help.go', 'public', 'unlisted', 'artifact', 'live', 'llms.txt index of live help pages; each page also at /help/<slug>.md.');
