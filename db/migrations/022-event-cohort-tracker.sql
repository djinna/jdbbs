-- Migration 022: Turn workshop registrations into a reusable cohort tracker.
-- The dated slug keeps repeat runs separate; operational fields support prep,
-- attendance, notes, and consent-aware announcement history.

UPDATE event_registrations
SET event_slug = 'protocolize-your-book-2026-09'
WHERE event_slug = 'protocolize-your-book';

ALTER TABLE event_registrations ADD COLUMN prep_status TEXT NOT NULL DEFAULT 'not-started';
ALTER TABLE event_registrations ADD COLUMN attended_sessions INTEGER NOT NULL DEFAULT 0;
ALTER TABLE event_registrations ADD COLUMN last_emailed_at TIMESTAMP;

CREATE TABLE IF NOT EXISTS event_announcements (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    event_slug      TEXT NOT NULL,
    subject         TEXT NOT NULL,
    body            TEXT NOT NULL,
    recipient_ids   TEXT NOT NULL DEFAULT '[]',
    recipient_count INTEGER NOT NULL DEFAULT 0,
    sent_count      INTEGER NOT NULL DEFAULT 0,
    failed_count    INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_event_announcements_event
    ON event_announcements (event_slug, created_at DESC);
