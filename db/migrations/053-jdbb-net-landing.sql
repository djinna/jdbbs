-- jdbb.net landing page (punch list 8.3): replaces the Blot stub at the apex.
-- Served by landingHost for Host jdbb.net once DNS moves; preview at /jdbb-net.
INSERT OR IGNORE INTO site_pages (route, title, owner, source, visibility, listed, page_type, status, note) VALUES
 ('/jdbb-net', 'jdbb.net — Jenna Dixon, bookbuilder (landing)', 'jdbbs-public', 'jdbb-net.html', 'public', 'unlisted', 'marketing', 'draft', 'Apex landing page for jdbb.net (Host switch in srv/landing_host.go; www → apex; studio paths 301 to studio.jdbb.net; old Blot URLs 302 to /). Preview on the studio host at /jdbb-net until the DNS cutover.');
