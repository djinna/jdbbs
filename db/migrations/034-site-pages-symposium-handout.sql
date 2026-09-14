-- Register the participant session guide (workshop handout) in the pages registry.
INSERT OR IGNORE INTO site_pages (route, title, owner, source, visibility, listed, page_type, status, note) VALUES
 ('/2026-pi-symposium/workshop', 'Session guide (participant handout)', 'jdbbs-public', '2026-pi-symposium/workshop.html', 'public', 'listed', 'prose', 'live', 'Four sessions, homework, what to bring, Discord links. Linked from the prep email and cohort roster.');
