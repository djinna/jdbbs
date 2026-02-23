-- Migration 005: Add clients table for client-level authentication
-- Allows a client (e.g. "VGR") to access all their projects with one password,
-- complementing the existing per-project auth (auth_tokens table).

CREATE TABLE IF NOT EXISTS clients (
    slug TEXT PRIMARY KEY,          -- e.g. 'vgr', 'sw'
    name TEXT NOT NULL DEFAULT '',  -- e.g. 'Venkatesh Rao', 'Sarah Williams'
    password_hash TEXT NOT NULL DEFAULT '',  -- SHA-256 hash, empty = no auth
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT OR IGNORE INTO migrations (migration_number, migration_name) VALUES (005, '005-clients');
