-- Migration 027: outbound email log (2026-09-11).
--
-- Every AgentMail send attempt — success or failure — is recorded here so the
-- admin can audit what went to whom. Before this, sends only existed in the
-- systemd journal, which rotates. Read via GET /api/admin/email and the
-- admin "Mail" tab / registrations tracker "Sent mail" panel.
CREATE TABLE IF NOT EXISTS outbound_email (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    sent_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    to_addrs     TEXT NOT NULL,                  -- comma-separated
    cc_addrs     TEXT NOT NULL DEFAULT '',
    subject      TEXT NOT NULL,
    kind         TEXT NOT NULL,                  -- template: registration_confirm, announcement, factory_pass, ...
    ref_type     TEXT NOT NULL DEFAULT '',       -- registration | project | client | pass | book
    ref_id       TEXT NOT NULL DEFAULT '',
    triggered_by TEXT NOT NULL DEFAULT 'system', -- admin email | client | public | system
    status_code  INTEGER NOT NULL DEFAULT 0,     -- HTTP status from AgentMail; 0 = transport error
    error        TEXT NOT NULL DEFAULT '',
    note         TEXT NOT NULL DEFAULT ''        -- e.g. "backfilled from doc"
);
CREATE INDEX IF NOT EXISTS idx_outbound_email_sent ON outbound_email (sent_at DESC);
CREATE INDEX IF NOT EXISTS idx_outbound_email_kind ON outbound_email (kind, sent_at DESC);
CREATE INDEX IF NOT EXISTS idx_outbound_email_ref  ON outbound_email (ref_type, ref_id);

-- Backfill: six Factory Pass code announcements sent 2026-09-04. The
-- event_announcements table has one row per send (subject + exact timestamp)
-- but no recipient; the coupons issued in the same second tell us who. Joined
-- by rowid order: registrations 3..8 in id order ↔ announcements 1..6.
-- Andrea Leiter (consent_email=false) was skipped; Toby registered later.
INSERT INTO outbound_email (sent_at, to_addrs, subject, kind, ref_type, ref_id, triggered_by, status_code, note)
SELECT a.created_at, r.email, a.subject,
       'announcement', 'registration', CAST(r.id AS TEXT), 'j@djinna.com', 200,
       'backfilled from event_announcements + FACTORY-PASS-SESSION-A-DRY-RUN-2026-09-04.md'
FROM (SELECT r.*, ROW_NUMBER() OVER (ORDER BY r.id) rn
      FROM event_registrations r JOIN coupons c ON c.registration_id = r.id
      WHERE r.consent_email = 1 AND DATE(c.created_at) = '2026-09-04') r
JOIN (SELECT a.*, ROW_NUMBER() OVER (ORDER BY a.id) rn FROM event_announcements a WHERE DATE(a.created_at) = '2026-09-04') a
  ON a.rn = r.rn
WHERE NOT EXISTS (SELECT 1 FROM outbound_email o WHERE o.kind = 'announcement' AND o.ref_id = CAST(r.id AS TEXT) AND o.note LIKE 'backfilled%');

-- Registration confirmations + organizer alerts that predate the log (one per
-- registration; timestamps = registration time). Mike Check (#10) is a fixture.
INSERT INTO outbound_email (sent_at, to_addrs, subject, kind, ref_type, ref_id, triggered_by, status_code, note)
SELECT r.created_at, r.email, 'We got your Protocolize Your Book registration',
       'registration_confirm', 'registration', CAST(r.id AS TEXT), 'public', 200, 'backfilled from event_registrations'
FROM event_registrations r
WHERE NOT EXISTS (SELECT 1 FROM outbound_email o WHERE o.kind = 'registration_confirm' AND o.ref_id = CAST(r.id AS TEXT));
