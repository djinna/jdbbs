-- Register the factory map (who acts at each stage, and where) in the pages registry.
-- Unlisted until the facilitator decides to share it with the cohort.
INSERT OR IGNORE INTO site_pages (route, title, owner, source, visibility, listed, page_type, status, note) VALUES
 ('/2026-pi-symposium/map', 'Factory map (actors × stages × routes)', 'jdbbs-public', '2026-pi-symposium/map.html', 'public', 'unlisted', 'prose', 'live', 'Mermaid flow of S1–S4: author / rules / human / deterministic / LLM at each stage, with client + admin routes. Answers "where does the LLM sit" (nowhere in the pipeline). Linked from the session guide footer nav.');
