-- Migration 042: magic-link client sign-in (punch list 5.7).
-- A client asks for a link by email; we store only the sha256 of the token,
-- never the plaintext. Links are single-use and expire 30 minutes after issue.
CREATE TABLE IF NOT EXISTS login_links (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    client_slug TEXT      NOT NULL REFERENCES clients (slug) ON DELETE CASCADE,
    token_hash  TEXT      NOT NULL UNIQUE,          -- sha256 hex of the URL token
    email       TEXT      NOT NULL,                 -- address the link was mailed to (lowercased)
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at  TIMESTAMP NOT NULL,                 -- created_at + 30 min
    used_at     TIMESTAMP,                          -- NULL until redeemed
    ip          TEXT      NOT NULL DEFAULT ''       -- requester IP (issue)
);
CREATE INDEX IF NOT EXISTS idx_login_links_client ON login_links (client_slug, created_at DESC);

-- Optional contact address on the client itself, so a client that has no
-- Factory Pass (legacy production clients) can still be sent a sign-in link.
ALTER TABLE clients ADD COLUMN email TEXT NOT NULL DEFAULT '';

INSERT OR IGNORE INTO site_pages (route, title, owner, source, visibility, listed, page_type, status, note) VALUES
 ('/auth/link', 'Sign-in link redeem', 'prodcal', 'srv/login_links.go', 'public', 'unlisted', 'tool', 'live', 'GET ?t=TOKEN from the "Your sign-in link" email: sets the client cookie and 302s to /{client}/. Expired/used → themed page with a button back to the portal to request a fresh one.');

INSERT OR IGNORE INTO migrations (migration_number, migration_name) VALUES (042, '042-login-links');
