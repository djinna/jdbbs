-- Migration 024: client cohort membership + the admin Pages registry.
--
-- clients.cohort_slug — the "client-visible + cohort flag" tier
-- (docs/PAGE-DESIGN-HOSTING-VISIBILITY-2026-09-11.md §7). Set when a Factory
-- Pass coupon that is bound to a workshop registration is redeemed; admin can
-- also set it directly. A page gated on a cohort accepts any authenticated
-- client whose cohort_slug matches.
ALTER TABLE clients ADD COLUMN cohort_slug TEXT NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_clients_cohort ON clients (cohort_slug);

-- site_pages — inventory of record for every route we spin up (public docs,
-- admin tools, client surfaces, retired redirects). Authorization is decided
-- by the route's Go handler; this table only records what exists, who owns
-- it, how visible/discoverable it is, and whether we mean to keep it.
CREATE TABLE IF NOT EXISTS site_pages (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    route       TEXT NOT NULL UNIQUE,                 -- '/field-notes'
    title       TEXT NOT NULL DEFAULT '',
    owner       TEXT NOT NULL DEFAULT 'prodcal',      -- prodcal | pi-public | generated
    source      TEXT NOT NULL DEFAULT '',             -- file path or generator
    visibility  TEXT NOT NULL DEFAULT 'public',       -- public | client | cohort | admin
    listed      TEXT NOT NULL DEFAULT 'unlisted',     -- listed | nav | unlisted | retired
    page_type   TEXT NOT NULL DEFAULT 'prose',        -- marketing|prose|ledger|admin|client|artifact|deck|roster
    status      TEXT NOT NULL DEFAULT 'live',         -- live | keep | review | retire
    note        TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT OR IGNORE INTO site_pages (route, title, owner, source, visibility, listed, page_type, status, note) VALUES
 ('/',                            'Homepage',                        'prodcal',   'srv/static/landing.html',            'public', 'listed',   'marketing', 'live', 'reference shell'),
 ('/lg',                          'Landing +20% (retired)',          'prodcal',   '',                                   'public', 'retired',  'marketing', 'retire', '301 → / since 2026-09-11'),
 ('/admin/',                      'Admin dashboard',                 'prodcal',   'srv/static/admin.html',              'admin',  'nav',      'admin',     'live', ''),
 ('/admin/registrations',         'Cohort tracker',                  'prodcal',   'srv/static/registrations.html',      'admin',  'nav',      'admin',     'live', 'two print modes; consent-aware email'),
 ('/cohort/protocolize-your-book-2026-09', 'Workshop cohort roster', 'prodcal',   'srv/static/cohort.html',             'cohort', 'nav',      'roster',    'live', 'attendee-facing; no emails/notes/status'),
 ('/{client}/',                   'Client portal',                   'prodcal',   'srv/static/client.html',             'client', 'unlisted', 'client',    'live', ''),
 ('/{client}/{project}/',         'Project app (SPA)',               'prodcal',   'srv/static/index.html',              'client', 'unlisted', 'client',    'live', ''),
 ('/{client}/{project}/factory/', 'Factory (customer build)',        'prodcal',   'srv/static/factory.html',            'client', 'unlisted', 'client',    'live', ''),
 ('/{client}/{project}/transmittal/', 'Transmittal',                 'prodcal',   'srv/static/transmittal.html',        'client', 'unlisted', 'client',    'live', 'print view'),
 ('/stylesheet/',                 'Editorial stylesheet',            'prodcal',   'srv/static/stylesheet/index.html',   'public', 'listed',   'ledger',    'review', 'split into /stylesheet-pi (tool) + anonymized public version pending'),
 ('/stylesheet/authors',          'Stylesheet — author queries',     'prodcal',   'srv/static/stylesheet/authors.html', 'public', 'nav',      'ledger',    'review', ''),
 ('/factory',                     'Factory Pass offer + redeem',     'pi-public', 'factory.html',                       'public', 'unlisted', 'prose',     'live', 'coupon redeem form'),
 ('/exedeck',                     'exe.dev talk deck (SIGPfB)',      'pi-public', 'exedeck.html',                       'public', 'unlisted', 'deck',      'live', 'documented exception: own 960px stage'),
 ('/litmags',                     'Lit-mag tool stack',              'pi-public', 'litmags.html',                       'public', 'listed',   'ledger',    'live', ''),
 ('/field-notes',                 'Field notes: language as protocol','pi-public','field-notes.html',                   'public', 'listed',   'prose',     'live', ''),
 ('/workshop',                    'Workshop registration',           'pi-public', 'workshop.html',                      'public', 'listed',   'prose',     'live', 'live form → /api/public/register'),
 ('/field-guide',                 'Field guide (anonymized client)', 'generated', 'pi-public/client-raw/anonymize.sh',  'public', 'unlisted', 'artifact',  'live', 'never hand-edit'),
 ('/work-notes-standard',         'Work-notes standard (anonymized)','generated', 'pi-public/client-raw/anonymize.sh',  'public', 'unlisted', 'artifact',  'live', 'never hand-edit'),
 ('/architecture-plan',           'Architecture plan (anonymized)',  'generated', 'pi-public/client-raw/anonymize.sh',  'public', 'unlisted', 'artifact',  'live', 'never hand-edit');
