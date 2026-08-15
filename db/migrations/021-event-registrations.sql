-- Migration 021: Public event/workshop registration form.
-- Backs the "Protocolize Your Book" workshop reg form (Protocol Symposium
-- 2026): a public, unauthenticated POST captures a lightweight cohort
-- application; admin reviews/exports. Kept generic (event_slug) so the same
-- table can serve future events.

CREATE TABLE IF NOT EXISTS event_registrations (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    event_slug     TEXT    NOT NULL DEFAULT 'protocolize-your-book',
    name           TEXT    NOT NULL DEFAULT '',
    email          TEXT    NOT NULL DEFAULT '',
    region         TEXT    NOT NULL DEFAULT '',   -- EU | US | Asia | Other
    all_sessions   INTEGER NOT NULL DEFAULT 0,    -- can attend all four (bool)
    material        TEXT   NOT NULL DEFAULT '',   -- what they'll bring (free text)
    material_type   TEXT   NOT NULL DEFAULT '',   -- book|essays|archive|zine|poetry|other
    background      TEXT   NOT NULL DEFAULT '',   -- publishing|design|writing|other
    new_to_protocol INTEGER NOT NULL DEFAULT 0,   -- tiebreaker flag (bool)
    goals          TEXT    NOT NULL DEFAULT '',   -- optional free text
    consent_email  INTEGER NOT NULL DEFAULT 0,    -- ok to email (bool)
    status         TEXT    NOT NULL DEFAULT 'requested', -- requested|confirmed|waitlist|declined
    notes          TEXT    NOT NULL DEFAULT '',   -- admin-only curation notes
    user_agent     TEXT    NOT NULL DEFAULT '',
    ip             TEXT    NOT NULL DEFAULT '',
    created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_event_registrations_event
    ON event_registrations (event_slug, created_at DESC);

-- One active (non-declined) request per email per event; a re-submit updates.
CREATE UNIQUE INDEX IF NOT EXISTS idx_event_registrations_unique
    ON event_registrations (event_slug, email);
