-- Migration 002: Update users table for auth integration
-- Run: psql -d cms_ai -f server/migrations/002_auth_update.sql

ALTER TABLE users ADD COLUMN name TEXT;
ALTER TABLE users ADD COLUMN updated_at TIMESTAMPTZ DEFAULT NOW();

-- Rename memberships to user_orgs for consistency with code
ALTER TABLE memberships RENAME TO user_orgs;