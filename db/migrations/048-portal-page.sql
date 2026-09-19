-- Client portal sign-in page (punch list 0.24): the client-code form moved off
-- the landing page to its own route so / reads as one narrative.
INSERT OR IGNORE INTO site_pages (route, title, owner, source, visibility, listed, page_type, status, note) VALUES
 ('/portal', 'Client portal (sign-in)', 'prodcal', 'srv/static/portal.html', 'public', 'listed', 'tool', 'live', 'One field: client code → /{client}/. Linked from every public nav (PUBLIC_NAV in theme.js) and the footer. Was the #portal section on / until 2026-09-19.');
