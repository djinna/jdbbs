-- pi-public was renamed jdbbs-public (2026-09-13).
UPDATE site_pages SET owner = 'jdbbs-public' WHERE owner = 'pi-public';
UPDATE site_pages SET source = REPLACE(source, 'pi-public/', 'jdbbs-public/') WHERE source LIKE 'pi-public/%';
