-- Factory activity feed (L2 monitoring): one row per thing a pass holder or
-- the factory did — sign-in, upload, Inspect, build start/done/failed,
-- transmittal final, template and file downloads, cover changes, pass
-- fulfilled. Read by /admin/factory/. Written best-effort; never blocks a
-- request.
CREATE TABLE IF NOT EXISTS factory_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    project_id INTEGER,                    -- NULL for client-level events (sign-in)
    client_slug TEXT NOT NULL DEFAULT '',
    kind TEXT NOT NULL,                    -- dotted: build.started, inspect, login.failed …
    actor TEXT NOT NULL DEFAULT '',        -- requestActor(): admin:email | client:slug | factory | anon
    detail TEXT NOT NULL DEFAULT ''        -- one human line
);
CREATE INDEX IF NOT EXISTS idx_factory_events_id_desc ON factory_events(id DESC);
CREATE INDEX IF NOT EXISTS idx_factory_events_project ON factory_events(project_id, id DESC);

INSERT OR IGNORE INTO site_pages (route, title, owner, source, visibility, listed, page_type, status, note) VALUES
 ('/admin/factory/', 'Factory floor (live activity during sessions)', 'prodcal', 'srv/static/factory-admin.html', 'admin', 'nav', 'tool', 'live', 'One screen for the workshop: a board of every live pass (transmittal, manuscript, last Inspect, builds left) and a live feed of sign-ins, uploads, Inspects, builds, downloads. Polls every 10 s.');
