-- Register the store admin page in the Pages registry.
INSERT OR IGNORE INTO site_pages (route, title, owner, source, visibility, listed, page_type, status, note) VALUES
 ('/admin/store/', 'Store (Factory Pass sales, orders, revoke)', 'prodcal', 'srv/static/store-admin.html', 'admin', 'nav', 'tool', 'live', 'Passes with source/amount/promo and Revoke/Reinstate/+3 builds; Stripe order ledger; catalog prices. Refunds are done in the Stripe dashboard, then revoke here.'),
 ('/factory/thanks', 'Factory Pass — thanks / confirmation page', 'prodcal', 'srv/static/store-thanks.html', 'public', 'unlisted', 'tool', 'live', 'Stripe Checkout return URL (?session_id=). Polls /api/public/store/session, which fulfils the pass. 404 while PRODCAL_STORE is off.');
