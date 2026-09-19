-- 0.17 (C), 2026-09-19: the transmittal is section 1 of the factory page.
-- /{client}/{project}/transmittal/ now 302s to /{client}/{project}/factory/#transmittal;
-- srv/static/transmittal.html is gone (transmittal.js mounts inside factory.html).
UPDATE site_pages
   SET status = 'retire', listed = 'retired', source = '',
       note = '302 → /{client}/{project}/factory/#transmittal since 2026-09-19 (0.17 C)'
 WHERE route = '/{client}/{project}/transmittal/';
UPDATE site_pages
   SET title = 'Factory (transmittal + build)', source = 'srv/static/factory.html',
       note = 'one page, five steps: 1 transmittal (transmittal.js embedded) · upload · inspect · build · download'
 WHERE route = '/{client}/{project}/factory/';
