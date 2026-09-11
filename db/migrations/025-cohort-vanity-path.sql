-- Roster moves to its attendee-facing vanity URL; /cohort/{slug} 301s to it.
UPDATE site_pages SET route = '/2026-pi-symposium',
  note = 'attendee-facing; no emails/notes/status. /cohort/protocolize-your-book-2026-09 301s here',
  updated_at = CURRENT_TIMESTAMP
WHERE route = '/cohort/protocolize-your-book-2026-09';
