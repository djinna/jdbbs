-- Report a nit (punch list 8.7): anyone, no login. One pipe for help errors
-- and app bugs; each also lands in punch-list Section 0, tagged.
CREATE TABLE IF NOT EXISTS nits (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    kind        TEXT NOT NULL DEFAULT 'app',     -- help | app
    page        TEXT NOT NULL DEFAULT '',        -- path (+ #hash) the reader was on
    selection   TEXT NOT NULL DEFAULT '',        -- text they had selected, if any
    comment     TEXT NOT NULL DEFAULT '',
    email       TEXT NOT NULL DEFAULT '',        -- optional; only for a "fixed, thanks" reply (8.x, consent guard)
    ip_hash     TEXT NOT NULL DEFAULT '',        -- salted hash, for abuse review only
    user_agent  TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'new',     -- new | triaged | fixed | wontfix | spam
    inbox_ok    INTEGER NOT NULL DEFAULT 0       -- 1 once pushed to the punch-list inbox
);
CREATE INDEX IF NOT EXISTS idx_nits_created ON nits(created_at DESC);
