-- 038: Factory Pass store (Stripe Checkout).
--
-- A paid pass records which Checkout Session bought it, so fulfilment is
-- idempotent (the thanks page and the event poller can both try). Add-ons
-- (+3 builds, +6 months storage) bought later against an existing pass are
-- their own rows in store_orders; the pass itself is bumped in place
-- (builds_extra / expires_at), so a pass's history is the order list.
--
-- store_orders is the ledger of every Checkout Session we fulfilled, pass or
-- add-on, with the Stripe ids needed to reconcile against the dashboard.

ALTER TABLE passes ADD COLUMN stripe_session_id TEXT NOT NULL DEFAULT '';
ALTER TABLE passes ADD COLUMN amount_paid INTEGER NOT NULL DEFAULT 0;   -- cents, after discount
ALTER TABLE passes ADD COLUMN promo_code TEXT NOT NULL DEFAULT '';      -- Stripe promotion code used, if any

CREATE UNIQUE INDEX IF NOT EXISTS idx_passes_stripe_session
    ON passes (stripe_session_id) WHERE stripe_session_id <> '';

CREATE TABLE IF NOT EXISTS store_orders (
    id                 INTEGER PRIMARY KEY AUTOINCREMENT,
    stripe_session_id  TEXT    NOT NULL UNIQUE,
    kind               TEXT    NOT NULL,                 -- pass | addon
    pass_id            INTEGER REFERENCES passes (id) ON DELETE SET NULL,
    customer_email     TEXT    NOT NULL DEFAULT '',
    customer_name      TEXT    NOT NULL DEFAULT '',
    amount_total       INTEGER NOT NULL DEFAULT 0,       -- cents, after discount
    currency           TEXT    NOT NULL DEFAULT 'usd',
    promo_code         TEXT    NOT NULL DEFAULT '',
    items              TEXT    NOT NULL DEFAULT '[]',    -- JSON [{lookup_key, quantity}]
    payment_intent_id  TEXT    NOT NULL DEFAULT '',
    fulfilled_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    note               TEXT    NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_store_orders_pass ON store_orders (pass_id);
