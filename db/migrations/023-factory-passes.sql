-- Migration 023: Factory Pass — the sellable unit of the book factory.
-- One manuscript's trip through the protocol: a project, unlimited preflights,
-- N included builds (a build = one successful convert), 6 months of storage.
--
-- Three tables:
--   coupons     — redeemable codes (workshop free-access codes today, promo later)
--   passes      — one row per project; the entitlement itself
--   pass_ledger — append-only credit movements (debit per build, refund on failure)
--
-- credits_remaining = builds_included + builds_extra - builds_used
-- live              = status = 'active' AND expires_at > now
--
-- See docs/specs/FACTORY-PASS-API-2026-09-03.md for the shared contract.

CREATE TABLE IF NOT EXISTS coupons (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    code            TEXT    NOT NULL UNIQUE,             -- PYB-XXXX-XXXX
    sku             TEXT    NOT NULL DEFAULT 'factory-pass',
    max_redemptions INTEGER NOT NULL DEFAULT 1,
    redeemed_count  INTEGER NOT NULL DEFAULT 0,
    expires_at      TIMESTAMP,                            -- NULL = never expires
    registration_id INTEGER REFERENCES event_registrations (id) ON DELETE SET NULL,
    issued_to_email TEXT    NOT NULL DEFAULT '',
    note            TEXT    NOT NULL DEFAULT '',
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_coupons_registration
    ON coupons (registration_id);

CREATE TABLE IF NOT EXISTS passes (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id      INTEGER NOT NULL UNIQUE REFERENCES projects (id) ON DELETE CASCADE,
    sku             TEXT    NOT NULL DEFAULT 'factory-pass',
    source          TEXT    NOT NULL DEFAULT 'coupon',    -- coupon|admin|stripe
    coupon_id       INTEGER REFERENCES coupons (id) ON DELETE SET NULL,
    customer_email  TEXT    NOT NULL DEFAULT '',
    customer_name   TEXT    NOT NULL DEFAULT '',
    builds_included INTEGER NOT NULL DEFAULT 3,
    builds_used     INTEGER NOT NULL DEFAULT 0,
    builds_extra    INTEGER NOT NULL DEFAULT 0,           -- bought/granted packs
    fulfilled_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at      TIMESTAMP NOT NULL,                   -- fulfilled_at + 6 months
    status          TEXT    NOT NULL DEFAULT 'active',    -- active|expired|purged|revoked
    note            TEXT    NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_passes_coupon ON passes (coupon_id);
CREATE INDEX IF NOT EXISTS idx_passes_status ON passes (status, expires_at);

CREATE TABLE IF NOT EXISTS pass_ledger (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    pass_id    INTEGER NOT NULL REFERENCES passes (id) ON DELETE CASCADE,
    book_id    INTEGER REFERENCES books (id) ON DELETE SET NULL,
    delta      INTEGER NOT NULL,                          -- -1 debit, +1 refund, +N pack
    reason     TEXT    NOT NULL,                          -- build|build_failed_refund|pack|grant|revoke
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_pass_ledger_pass
    ON pass_ledger (pass_id, created_at DESC);

INSERT OR IGNORE INTO migrations (migration_number, migration_name) VALUES (023, '023-factory-passes');
