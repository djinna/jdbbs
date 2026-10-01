-- Site-wide policies (punch list 10.1/10.2): /privacy (Stripe requires a
-- privacy policy: what is collected, its use, who it is disclosed to and how,
-- and the security practices) and /terms (Terms of Use). Public documents in
-- jdbbs-public, plain English, same voice as /factory/terms. Drafts until
-- Jenna approves; then linked from every footer, /factory and checkout.
INSERT OR IGNORE INTO site_pages (route, title, owner, source, visibility, listed, page_type, status, note) VALUES
 ('/privacy', 'Privacy policy', 'jdbbs-public', 'privacy.html', 'public', 'unlisted', 'prose', 'draft', 'What we collect, why, who it goes to and how, how long we keep it, security, your choices. Built from a code inventory (1 Oct 2026). Draft until Jenna approves (10.4); then footer + /factory + checkout links and the Stripe public-details URL.'),
 ('/terms', 'Terms of Use', 'jdbbs-public', 'terms.html', 'public', 'unlisted', 'prose', 'draft', 'Site-wide terms: account, your work stays yours, fair use, buying a pass (refunds, codes, nothing to cancel), availability, liability, changes, law. Defers to /factory/terms for pass details. Draft until Jenna approves (10.4).');
