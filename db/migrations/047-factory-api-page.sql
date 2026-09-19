-- Factory API recipe page (punch list 0.9): the six calls, curl + stdlib Python.
INSERT OR IGNORE INTO site_pages (route, title, owner, source, visibility, listed, page_type, status, note) VALUES
 ('/factory/api', 'Factory API', 'pi-public', 'factory-api.html', 'public', 'listed', 'prose', 'live', 'HTML rendering of docs/API-CLI-RECIPE-2026-09-19.md. Linked from /factory ("Have your own factory?"). /factory/api/factory-cli.py serves scripts/factory-cli.py via a symlink in jdbbs-public.');
