-- Migration 028: store the exact message that went out (2026-09-11).
-- The admin Mail report shows to/subject/status; this lets it show the body.
-- Rows that predate this column read as "body not recorded".
ALTER TABLE outbound_email ADD COLUMN text_body TEXT NOT NULL DEFAULT '';
ALTER TABLE outbound_email ADD COLUMN html_body TEXT NOT NULL DEFAULT '';
