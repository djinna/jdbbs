-- Factory Pass: coupons, passes, and the credit ledger (migration 023).
-- See docs/specs/FACTORY-PASS-API-2026-09-03.md.

-- name: CreateCoupon :one
INSERT INTO coupons (code, sku, max_redemptions, expires_at, registration_id, issued_to_email, note)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetCouponByCode :one
SELECT * FROM coupons WHERE code = ?;

-- name: IncrementCouponRedeemed :exec
UPDATE coupons SET redeemed_count = redeemed_count + 1 WHERE id = ?;

-- name: SetCouponRegistration :exec
UPDATE coupons SET registration_id = ? WHERE id = ? AND registration_id IS NULL;

-- name: ListCoupons :many
SELECT
    c.id, c.code, c.sku, c.max_redemptions, c.redeemed_count, c.expires_at,
    c.registration_id, c.issued_to_email, c.note, c.created_at,
    p.id AS pass_id, p.fulfilled_at AS redeemed_at,
    COALESCE(pr.client_slug, '') AS client_slug,
    COALESCE(pr.project_slug, '') AS project_slug,
    COALESCE(pr.name, '') AS project_name
FROM coupons c
LEFT JOIN passes p ON p.coupon_id = c.id
LEFT JOIN projects pr ON pr.id = p.project_id
ORDER BY c.created_at DESC, c.id DESC;

-- name: CreatePass :one
-- expires_at is computed in SQL so the six-month window always matches
-- fulfilled_at and uses SQLite's timestamp format.
INSERT INTO passes (
    project_id, sku, source, coupon_id, customer_email, customer_name,
    builds_included, expires_at, note
)
VALUES (?, ?, ?, ?, ?, ?, ?, datetime(CURRENT_TIMESTAMP, '+6 months'), ?)
RETURNING *;

-- name: GetPass :one
SELECT * FROM passes WHERE id = ?;

-- name: GetPassByProject :one
SELECT * FROM passes WHERE project_id = ?;

-- name: ListPasses :many
SELECT
    p.id, p.project_id, p.sku, p.source, p.coupon_id, p.customer_email,
    p.customer_name, p.builds_included, p.builds_used, p.builds_extra,
    p.fulfilled_at, p.expires_at, p.status, p.note,
    p.stripe_session_id, p.amount_paid, p.promo_code, p.index_included,
    pr.name AS project_name, pr.client_slug, pr.project_slug,
    COALESCE(c.code, '') AS coupon_code
FROM passes p
JOIN projects pr ON pr.id = p.project_id
LEFT JOIN coupons c ON c.id = p.coupon_id
ORDER BY p.fulfilled_at DESC, p.id DESC;

-- name: IncrementPassBuildsUsed :exec
UPDATE passes SET builds_used = builds_used + 1 WHERE id = ?;

-- name: DecrementPassBuildsUsed :exec
UPDATE passes SET builds_used = CASE WHEN builds_used > 0 THEN builds_used - 1 ELSE 0 END
WHERE id = ?;

-- name: AddPassBuildsExtra :exec
UPDATE passes SET builds_extra = builds_extra + ? WHERE id = ?;

-- name: UpdatePassStatus :exec
UPDATE passes SET status = ? WHERE id = ?;

-- name: CreatePassLedgerEntry :exec
INSERT INTO pass_ledger (pass_id, book_id, delta, reason) VALUES (?, ?, ?, ?);

-- name: ListPassLedger :many
SELECT * FROM pass_ledger WHERE pass_id = ? ORDER BY id DESC;

-- name: FindRegistrationIDByEmail :one
SELECT id FROM event_registrations WHERE email = ? ORDER BY id DESC LIMIT 1;

-- name: CountConvertingBooksByProject :one
-- Backs the one-in-flight-build-per-project guard on POST /api/books/{id}/convert.
SELECT COUNT(*) FROM books WHERE project_id = ? AND status = 'converting';

-- name: SetPassPurchase :exec
-- Stamps a pass with the Checkout Session that bought it (source = stripe).
UPDATE passes SET stripe_session_id = ?, amount_paid = ?, promo_code = ? WHERE id = ?;

-- name: AddPassExtras :exec
-- Add-on fulfilment: more build credits and/or a longer storage window.
-- months is a SQLite modifier string such as '+6 months' ('+0 months' = none).
UPDATE passes
SET builds_extra = builds_extra + ?,
    expires_at   = datetime(expires_at, ?)
WHERE id = ?;

-- name: GetStoreOrderBySession :one
SELECT * FROM store_orders WHERE stripe_session_id = ?;

-- name: CreateStoreOrder :one
INSERT INTO store_orders (
    stripe_session_id, kind, pass_id, customer_email, customer_name,
    amount_total, currency, promo_code, items, payment_intent_id, note
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: ListStoreOrders :many
SELECT o.*, COALESCE(pr.name, '') AS project_name,
       COALESCE(pr.client_slug, '') AS client_slug, COALESCE(pr.project_slug, '') AS project_slug
FROM store_orders o
LEFT JOIN passes p ON p.id = o.pass_id
LEFT JOIN projects pr ON pr.id = p.project_id
ORDER BY o.fulfilled_at DESC, o.id DESC
LIMIT ?;

-- name: ListStoreOrdersForPass :many
SELECT * FROM store_orders WHERE pass_id = ? ORDER BY fulfilled_at DESC, id DESC;

-- name: SetPassIndexIncluded :exec
-- Back-of-book index add-on fulfilment (store "index" item or admin grant).
UPDATE passes SET index_included = 1 WHERE id = ?;
